package bilibili

import (
	"bytes"
	"compress/flate"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ac1982/haul/internal/testkit"
)

const danmakuXML = `<?xml version="1.0" encoding="UTF-8"?><i><chatserver>chat.bilibili.com</chatserver>
<d p="1.5,1,25,16646914,1700000000,0,abc,1">第一条</d>
<d p="2.0,5,25,16777215,1700000000,0,abc,2">顶部</d>
<d p="3.0,4,25,16777215,1700000000,0,abc,3">底部 &amp; more</d>
<d p="bad">忽略</d>
</i>`

func TestDanmakuColourAndTime(t *testing.T) {
	t.Parallel()
	if assColor("FE0302") != "0203FE" || assColor("abc") != "abc" {
		t.Fatal("assColor")
	}
	for s, want := range map[float64]string{65.5: "0:01:05.50", 3725.0: "1:02:05.00", 0: "0:00:00.00", 59.999: "0:00:60.00"} {
		if got := assTime(s); got != want {
			t.Errorf("%v → %s, want %s", s, got, want)
		}
	}
}

func TestDanmakuParseAndLayout(t *testing.T) {
	t.Parallel()
	items, err := parseDanmaku([]byte(danmakuXML))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 || items[0].color != "FE0302" || items[0].mode != scroll || items[1].mode != top || items[2].mode != bottom || items[2].content != "底部 & more" {
		t.Fatalf("%+v", items)
	}
	ass := danmakuASS(items)
	for _, want := range []string{
		"PlayResX: 1920\nPlayResY: 1080\n",
		"Style: Danmaku, 黑体, 40, &H00FFFFFF",
		`Dialogue: 2,0:00:01.50,0:00:09.50,Danmaku,,0000,0000,0000,,{\move(1920, 0, -120, 0)\c&H0203FE&}第一条`,
		`Dialogue: 2,0:00:02.00,0:00:06.00,Danmaku,,0000,0000,0000,,{\an8\pos(960, 0)}顶部`,
		`{\an8\pos(960, 1040)}底部 & more`,
	} {
		if !strings.Contains(ass, want) {
			t.Errorf("missing %q in\n%s", want, ass)
		}
	}
	if _, err := parseDanmaku([]byte("<i><d p=")); err == nil {
		t.Fatal("broken XML parsed")
	}
}

func TestDanmakuRowsFillThenOverflow(t *testing.T) {
	t.Parallel()
	r := newRows()
	// 13 rows in the top half; a 14th static comment at the same moment has nowhere to go.
	for i := range 13 {
		if y := r.place(top, 0, 1); y != i*40 {
			t.Fatalf("row %d at %d", i, y)
		}
	}
	if y := r.place(top, 1, 1); y != -1 {
		t.Fatalf("overflow placed at %d", y)
	}
	// Rows free up once their comment is gone; other modes have rows of their own.
	if y := r.place(top, 4, 1); y != 0 {
		t.Fatalf("freed row at %d", y)
	}
	if y := r.place(scroll, 0, 10); y != 0 {
		t.Fatalf("scroll row at %d", y)
	}
}

func TestDanmakuSidecar(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("comment.bilibili.com/2.xml", func(*http.Request) testkit.Response { return testkit.Text(danmakuXML) })
	sidecar := e.danmaku(plainSession(), "2")
	if sidecar.Kind != "danmaku" {
		t.Fatal(sidecar.Kind)
	}
	base := filepath.Join(t.TempDir(), "Video")
	files, err := sidecar.Write(ctx, base)
	if err != nil || !slices.Equal(files, []string{base + ".ass", base + ".xml"}) {
		t.Fatalf("%v %v", files, err)
	}
	if xml, _ := os.ReadFile(base + ".xml"); string(xml) != danmakuXML {
		t.Fatal("XML changed")
	}
	if req := stub.Requests("comment.bilibili.com")[0]; req.Header.Get("Referer") != "https://www.bilibili.com" || req.URL.Scheme != "https" {
		t.Fatalf("%s %v", req.URL, req.Header)
	}

	// Only ASS asked for: the XML is not kept.
	e.opts.DanmakuFormats = []string{"ass"}
	only := filepath.Join(t.TempDir(), "Only")
	if files, _ := e.danmaku(plainSession(), "2").Write(ctx, only); !slices.Equal(files, []string{only + ".ass"}) {
		t.Fatalf("%v", files)
	}
	if _, err := os.Stat(only + ".xml"); err == nil {
		t.Fatal("XML kept")
	}
}

func TestDanmakuNoneOrBroken(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("comment.bilibili.com/3.xml", func(*http.Request) testkit.Response {
		return testkit.Text(`<?xml version="1.0" encoding="UTF-8"?><i><chatserver>chat.bilibili.com</chatserver></i>`)
	})
	stub.On("comment.bilibili.com/4.xml", func(*http.Request) testkit.Response { return testkit.Text("<i><d p=") })
	stub.On("comment.bilibili.com/5.xml", status(500))
	dir := t.TempDir()
	for _, cid := range []string{"3", "4"} {
		files, err := e.danmaku(plainSession(), cid).Write(ctx, filepath.Join(dir, cid))
		if err != nil || files != nil {
			t.Fatalf("%s: %v %v", cid, files, err)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("%d files left", len(entries))
	}
	if _, err := e.danmaku(plainSession(), "5").Write(ctx, filepath.Join(dir, "5")); err == nil {
		t.Fatal("a failed download is no error")
	}
}

func TestDeflatedDanmaku(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	var raw bytes.Buffer
	fw, _ := flate.NewWriter(&raw, flate.BestCompression)
	fw.Write([]byte(danmakuXML))
	fw.Close()
	stub.On("comment.bilibili.com/6.xml", func(*http.Request) testkit.Response {
		return testkit.Response{Header: http.Header{"Content-Encoding": {"deflate"}}, Body: raw.Bytes()}
	})
	data, err := e.fetchDanmaku(ctx, plainSession(), "https://comment.bilibili.com/6.xml")
	if err != nil || string(data) != danmakuXML {
		t.Fatalf("%q %v", data, err)
	}
}
