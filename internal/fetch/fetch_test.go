package fetch

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

// bigMedia is n MiB and a bit, built from a 1 MiB pattern so a misplaced byte still shows.
func bigMedia(n int) []byte {
	block := testkit.Pattern(1 << 20)
	var b bytes.Buffer
	for i := range n {
		block[len(block)-1] = byte(i)
		b.Write(block)
	}
	b.Write(testkit.Pattern(12345))
	return b.Bytes()
}

func setup(t *testing.T) (*testkit.Stub, *HTTP, string) {
	t.Helper()
	t.Parallel()
	stub := testkit.NewStub()
	return stub, NewHTTP(stub.Client()), t.TempDir()
}

func res(u string, policy media.RangePolicy) media.Resource {
	return media.Resource{URL: u, Policy: policy}
}

func mustEqualFile(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s: %d bytes differ from the %d expected", path, len(got), len(want))
	}
}

func TestWholeStream(t *testing.T) {
	stub, h, dir := setup(t)
	data := testkit.Pattern(300_000)
	stub.On("dl.test/single", func(r *http.Request) testkit.Response { return testkit.Ranged(data, r) })
	out := filepath.Join(dir, "a.m4a")
	if err := h.Fetch(context.Background(), res("https://dl.test/single", media.Whole), out, nil); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
	if got := stub.Ranges("dl.test/single"); !reflect.DeepEqual(got, []string{"bytes=0-"}) {
		t.Errorf("ranges = %q", got)
	}
}

func TestParallelAcrossClips(t *testing.T) {
	stub, h, dir := setup(t)
	data := bigMedia(41)
	stub.On("dl.test/segmented", func(r *http.Request) testkit.Response { return testkit.Ranged(data, r) })
	out := filepath.Join(dir, "v.mp4")
	var last atomic.Int64
	p := progressFunc(func(done, total int64) { last.Store(done) })
	if err := h.Fetch(context.Background(), res("https://dl.test/segmented", media.Parallel), out, p); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
	ranges := map[string]bool{}
	for _, r := range stub.Ranges("dl.test/segmented") {
		ranges[r] = true
	}
	for _, want := range []string{"bytes=0-20971520", "bytes=20971521-41943041", "bytes=41943042-"} {
		if !ranges[want] {
			t.Errorf("missing range %s in %v", want, ranges)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 1 {
		t.Errorf("clip files left behind: %v", entries)
	}
	if last.Load() != int64(len(data)) {
		t.Errorf("progress ended at %d", last.Load())
	}
}

func TestServerIgnoringRangesFallsBackToOneStream(t *testing.T) {
	stub, h, dir := setup(t)
	data := bigMedia(41)
	stub.On("dl.test/norange", func(*http.Request) testkit.Response { return testkit.Response{Body: data} })
	out := filepath.Join(dir, "v.mp4")
	if err := h.Fetch(context.Background(), res("https://dl.test/norange", media.Parallel), out, nil); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
}

// rangeOnly is googlevideo: plain and open-ended requests get 403, and so does a range over the limit.
func rangeOnly(data []byte, limit int64) testkit.Handler {
	return func(r *http.Request) testkit.Response {
		from, to, ok := testkit.Range(r)
		if !ok || to < 0 || to-from >= limit {
			return testkit.Status(403)
		}
		return testkit.Ranged(data, r)
	}
}

func TestSequentialGetsClosedRangesOneAtATime(t *testing.T) {
	stub, h, dir := setup(t)
	data := testkit.Pattern(2_500_000)
	stub.On("dl.test/rangeonly", rangeOnly(data, 1_000_000))
	out := filepath.Join(dir, "v.mp4")
	r := media.Resource{URL: "https://dl.test/rangeonly", Policy: media.Sequential, MaxRange: 1_000_000}
	if err := h.Fetch(context.Background(), r, out, nil); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
	want := []string{"bytes=0-0", "bytes=0-999999", "bytes=1000000-1999999", "bytes=2000000-2499999"}
	if got := stub.Ranges("dl.test/rangeonly"); !reflect.DeepEqual(got, want) {
		t.Errorf("ranges = %q, want %q", got, want)
	}
}

func TestSequentialWithoutASizeFails(t *testing.T) {
	stub, h, dir := setup(t)
	stub.On("dl.test/nosize", func(r *http.Request) testkit.Response {
		if _, to, _ := testkit.Range(r); to < 0 {
			return testkit.Status(403)
		}
		return testkit.Response{Status: 206, Header: http.Header{"Content-Range": {"bytes 0-0/*"}}, Body: []byte{0}}
	})
	out := filepath.Join(dir, "v.mp4")
	err := h.Fetch(context.Background(), media.Resource{URL: "https://dl.test/nosize", Policy: media.Sequential, MaxRange: 1_000_000}, out, nil)
	if !errs.Is(err, errs.Failed) {
		t.Errorf("err = %v", err)
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("no file should be written")
	}
}

func TestUnknownSizeFallsBackToOneStream(t *testing.T) {
	stub, h, dir := setup(t)
	data := testkit.Pattern(50_000)
	stub.On("dl.test/nolength", func(r *http.Request) testkit.Response {
		resp := testkit.Ranged(data, r)
		if _, _, ok := testkit.Range(r); !ok {
			resp.NoLength = true
		}
		return resp
	})
	out := filepath.Join(dir, "v.mp4")
	if err := h.Fetch(context.Background(), res("https://dl.test/nolength", media.Parallel), out, nil); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
}

func TestKnownSizeSkipsTheProbe(t *testing.T) {
	stub, h, dir := setup(t)
	data := testkit.Pattern(1_500_000)
	stub.On("dl.test/knownsize", rangeOnly(data, 1_000_000))
	out := filepath.Join(dir, "a.m4a")
	r := media.Resource{URL: "https://dl.test/knownsize", Policy: media.Sequential, MaxRange: 1_000_000, Size: int64(len(data))}
	if err := h.Fetch(context.Background(), r, out, nil); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
	if got := stub.Ranges("dl.test/knownsize"); !reflect.DeepEqual(got, []string{"bytes=0-999999", "bytes=1000000-1499999"}) {
		t.Errorf("ranges = %q", got)
	}
}

func TestDroppedConnectionResumes(t *testing.T) {
	stub, h, dir := setup(t)
	data := testkit.Pattern(400_000)
	var calls atomic.Int32
	stub.On("dl.test/dropped", func(r *http.Request) testkit.Response {
		resp := testkit.Ranged(data, r)
		if calls.Add(1) == 1 {
			resp.TruncateAt = 150_000 // promise everything, deliver part of it
		}
		return resp
	})
	out := filepath.Join(dir, "a.m4a")
	if err := h.Fetch(context.Background(), res("https://dl.test/dropped", media.Whole), out, nil); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
	if got := stub.Ranges("dl.test/dropped"); !reflect.DeepEqual(got, []string{"bytes=0-", "bytes=150000-"}) {
		t.Errorf("ranges = %q", got)
	}
}

func TestPartialFileIsResumed(t *testing.T) {
	stub, h, dir := setup(t)
	data := testkit.Pattern(200_000)
	stub.On("dl.test/partial", func(r *http.Request) testkit.Response { return testkit.Ranged(data, r) })
	out := filepath.Join(dir, "a.m4a")
	if err := os.WriteFile(out+".part", data[:1000], 0o644); err != nil {
		t.Fatal(err)
	}
	if err := h.Fetch(context.Background(), res("https://dl.test/partial", media.Whole), out, nil); err != nil {
		t.Fatal(err)
	}
	mustEqualFile(t, out, data)
	if got := stub.Ranges("dl.test/partial"); !reflect.DeepEqual(got, []string{"bytes=1000-"}) {
		t.Errorf("ranges = %q", got)
	}
}

func TestHTTPErrorsFailWithoutRetryingClientErrors(t *testing.T) {
	stub, h, dir := setup(t)
	stub.On("dl.test/missing", func(*http.Request) testkit.Response { return testkit.Status(404) })
	out := filepath.Join(dir, "a.m4a")
	err := h.Fetch(context.Background(), res("https://dl.test/missing", media.Parallel), out, nil)
	var se *httpx.StatusError
	if !errorsAs(err, &se) || se.Status != 404 {
		t.Errorf("err = %v", err)
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("no file should be written")
	}
	if n := len(stub.Requests("dl.test/missing")); n != 1 {
		t.Errorf("a 404 is not retried, got %d requests", n)
	}
}

func TestResourceHeadersAreSent(t *testing.T) {
	stub, h, dir := setup(t)
	data := testkit.Pattern(1000)
	stub.On("dl.test/headers", func(r *http.Request) testkit.Response { return testkit.Ranged(data, r) })
	r := media.Resource{URL: "https://dl.test/headers", Policy: media.Whole, Header: httpx.Header("User-Agent", "UA-X", "Cookie", "a=b")}
	if err := h.Fetch(context.Background(), r, filepath.Join(dir, "1.bin"), nil); err != nil {
		t.Fatal(err)
	}
	req := stub.Requests("dl.test/headers")[0]
	if req.Header.Get("User-Agent") != "UA-X" || req.Header.Get("Cookie") != "a=b" || req.Header.Get("Referer") != "" {
		t.Errorf("headers = %v", req.Header)
	}
}

func TestClipsTileEveryFileSize(t *testing.T) {
	for chunk := int64(1); chunk <= 12; chunk++ {
		for size := int64(1); size <= 200; size++ {
			for _, closed := range []bool{false, true} {
				clips := Clips(size, chunk, closed)
				next := int64(0)
				for i, c := range clips {
					end := c.To
					if end < 0 {
						end = size - 1
					}
					if c.Index != i || c.From != next || end < c.From || end-c.From > chunk {
						t.Fatalf("size %d chunk %d: clip %+v", size, chunk, c)
					}
					if !closed && i < len(clips)-1 && c.To < 0 {
						t.Fatalf("only the last clip is open: %+v", c)
					}
					next = end + 1
				}
				last := clips[len(clips)-1]
				if next != size || (closed && last.To != size-1) || (!closed && last.To != -1) {
					t.Fatalf("size %d chunk %d closed %v: %+v", size, chunk, closed, clips)
				}
			}
		}
	}
	if len(Clips(0, 10, false)) != 0 || len(Clips(10, 20<<20, false)) != 1 {
		t.Error("edge sizes")
	}
	got := Clips(25, 9, true)
	if !reflect.DeepEqual(got, []Clip{{0, 0, 9}, {1, 10, 19}, {2, 20, 24}}) {
		t.Errorf("closed clips = %+v", got)
	}
}

func TestAria2cArguments(t *testing.T) {
	a := &Aria2c{Path: "aria2c", Args: SplitArgs(`--all-proxy="http://127.0.0.1:7890"`)}
	r := media.Resource{URL: "https://video.test/a.mp4", Header: httpx.Header("User-Agent", "UA-X", "Accept", "*/*")}
	got := a.Arguments(r, filepath.Join("d", "v.mp4"))
	var headers []string
	for _, arg := range got {
		if len(arg) > 9 && arg[:9] == "--header=" {
			headers = append(headers, arg)
		}
	}
	if !reflect.DeepEqual(headers, []string{"--header=Accept: */*", "--header=User-Agent: UA-X"}) {
		t.Errorf("headers = %q", headers)
	}
	tail := got[len(got)-6:]
	if !reflect.DeepEqual(tail, []string{"--all-proxy=http://127.0.0.1:7890", "https://video.test/a.mp4", "-d", "d", "-o", "v.mp4"}) {
		t.Errorf("tail = %q", tail)
	}
}

func TestSplitArgs(t *testing.T) {
	cases := map[string][]string{
		`--all-proxy="http://127.0.0.1:7890" -x 4 'a b'`: {"--all-proxy=http://127.0.0.1:7890", "-x", "4", "a b"},
		``:          nil,
		`a\ b c`:    {"a b", "c"},
		`'' "x y"z`: {"", "x yz"},
	}
	for in, want := range cases {
		if got := SplitArgs(in); !reflect.DeepEqual(got, want) {
			t.Errorf("SplitArgs(%q) = %q, want %q", in, got, want)
		}
	}
}

type progressFunc func(done, total int64)

func (f progressFunc) Update(done, total int64) { f(done, total) }
func (f progressFunc) Finish(bool)              {}

func errorsAs(err error, target any) bool { return errors.As(err, target) }
