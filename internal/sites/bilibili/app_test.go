package bilibili

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

func TestProtobufWireFormat(t *testing.T) {
	t.Parallel()
	var w pbWriter
	w.int(1, 150).string(2, "testing").bool(3, true).int(4, -1).uint(5, 0)
	want := []byte{0x08, 0x96, 0x01, 0x12, 0x07, 't', 'e', 's', 't', 'i', 'n', 'g', 0x18, 0x01,
		0x20, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01, 0x28, 0x00}
	if !bytes.Equal(w.buf, want) {
		t.Fatalf("% x", w.buf)
	}
	m, err := decodeProto(w.buf)
	if err != nil {
		t.Fatal(err)
	}
	if m.int(1) != 150 || m.string(2) != "testing" || m.uint(3) != 1 || m.int(4) != -1 || !m.has(5) || m.has(6) || m.string(6) != "" {
		t.Fatalf("%+v", m)
	}
}

func TestProtobufRoundTrip(t *testing.T) {
	t.Parallel()
	var inner, outer pbWriter
	inner.string(1, "a").string(1, "b").int(2, 7)
	outer.message(3, &inner).message(3, &inner).int(4, 1<<40)
	// Fixed-width fields (floats in the danmaku player config) are read past.
	outer.tag(5, wireFixed32)
	outer.buf = append(outer.buf, 0, 0, 0x80, 0x3f)
	outer.tag(6, wireFixed64)
	outer.buf = append(outer.buf, 1, 2, 3, 4, 5, 6, 7, 8)
	outer.string(7, "after")
	m, err := decodeProto(outer.buf)
	if err != nil {
		t.Fatal(err)
	}
	subs := m.messages(3)
	if len(subs) != 2 || !slices.Equal(subs[1].strings(1), []string{"a", "b"}) || subs[0].string(1) != "b" || subs[0].int(2) != 7 {
		t.Fatalf("%+v", subs)
	}
	if m.int(4) != 1<<40 || m.uint(5) != 0x3f800000 || m.string(7) != "after" || m.message(9) != nil || m.message(4) != nil {
		t.Fatalf("%+v", m)
	}
	for _, bad := range [][]byte{{0x08}, {0x12, 0x05, 'a'}, {0x0d, 1, 2}, {0x09, 1}, {0x0b}, {0x00, 0x01}, {0x80}} {
		if _, err := decodeProto(bad); err == nil {
			t.Errorf("% x decoded", bad)
		}
	}
}

func TestRequestMessages(t *testing.T) {
	t.Parallel()
	req, _ := decodeProto(playViewRequest(840009001, 178175633, codeAV1))
	if req.int(1) != 840009001 || req.int(2) != 178175633 || req.int(3) != 127 || req.int(5) != 4048 || !req.has(6) || req.int(6) != 0 ||
		req.int(7) != 2 || req.uint(8) != 1 || req.string(9) != "main.ugc-video-detail.0.0" || req.string(10) != "main.my-history.0.0" || req.int(12) != codeAV1 || req.has(4) {
		t.Fatalf("%+v", req)
	}
	dm, _ := decodeProto(dmViewRequest(1, 2))
	if dm.int(1) != 1 || dm.int(2) != 2 || dm.int(3) != 1 || dm.string(4) != "main.ugc-video-detail.0.0" {
		t.Fatalf("%+v", dm)
	}
	if codeType("AVC") != code264 || codeType("AV1") != codeAV1 || codeType("HEVC") != code265 || codeType("") != code265 {
		t.Fatal("codeType")
	}

	h := appHeader("tok")
	if h.Get("Host") != "grpc.biliapi.net" || h.Get("Authorization") != "identify_v1 tok" || h.Get("Content-Type") != "application/grpc" ||
		!strings.Contains(h.Get("User-Agent"), "build/7320200 channel/xiaomi_cn_tv.danmaku.bili_zm20200902") || h.Get("Grpc-Timeout") != "17996161u" {
		t.Fatalf("%v", h)
	}
	if v, ok := h["X-Bili-Restriction-Bin"]; !ok || v[0] != "" {
		t.Fatal("empty restriction header dropped")
	}
	decode := func(name string) pbMessage {
		b, err := base64.StdEncoding.DecodeString(h.Get(name))
		if err != nil {
			t.Fatal(err)
		}
		m, err := decodeProto(b)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	if md := decode("X-Bili-Metadata-Bin"); md.string(1) != "tok" || md.int(4) != appBuild || md.string(2) != "android" || !md.has(6) {
		t.Fatalf("metadata %+v", md)
	}
	if d := decode("X-Bili-Device-Bin"); d.int(1) != 1 || d.string(8) != appBrand || d.string(10) != "11" {
		t.Fatalf("device %+v", d)
	}
	if l := decode("X-Bili-Locale-Bin").message(1); l.string(1) != "zh" || l.string(3) != "CN" {
		t.Fatalf("locale %+v", l)
	}
	if n := decode("X-Bili-Network-Bin"); n.int(1) != 1 || n.string(3) != "46007" {
		t.Fatalf("network %+v", n)
	}
	if f := decode("X-Bili-Fawkes-Req-Bin"); f.string(1) != "android64" || f.string(2) != "prod" || f.string(3) != "dedf8669" {
		t.Fatalf("fawkes %+v", f)
	}
}

func TestGzipFramesRoundTrip(t *testing.T) {
	t.Parallel()
	message := bytes.Repeat([]byte("bilibili "), 1000)
	frame := packFrame(message)
	if frame[0] != 1 || int(binary.BigEndian.Uint32(frame[1:5])) != len(frame)-5 || len(frame) > len(message)/4 {
		t.Fatalf("frame header % x, %d bytes", frame[:5], len(frame))
	}
	if frame[5] != 0x1f || frame[6] != 0x8b {
		t.Fatal("not gzip")
	}
	got, err := unpackFrame(frame)
	if err != nil || !bytes.Equal(got, message) {
		t.Fatalf("%v", err)
	}
}

func TestUncompressedFrameIgnoresTrailingBytes(t *testing.T) {
	t.Parallel()
	got, err := unpackFrame([]byte{0, 0, 0, 0, 3, 9, 8, 7, 0xff})
	if err != nil || !bytes.Equal(got, []byte{9, 8, 7}) {
		t.Fatalf("% x %v", got, err)
	}
	// A length beyond the data takes what is there.
	if got, _ := unpackFrame([]byte{0, 0, 0, 0, 9, 1}); !bytes.Equal(got, []byte{1}) {
		t.Fatalf("% x", got)
	}
	for _, bad := range [][]byte{nil, {1, 0, 0}, {1, 0, 0, 0, 2, 'x', 'y'}} {
		if _, err := unpackFrame(bad); err == nil {
			t.Errorf("% x unpacked", bad)
		}
	}
}

func dashItem(id uint64, url string, bandwidth uint64) *pbWriter {
	var w pbWriter
	w.uint(1, id).string(2, url).string(3, url+"-backup").uint(4, bandwidth).uint(7, 1000)
	return &w
}

// playViewReply builds a PlayViewReply the way the server sends it: two DASH streams (and one segment-only),
// audio with FLAC and Dolby, OP/ED markers and dubbing.
func playViewReply() []byte {
	var info, reply pbWriter
	info.uint(1, 80).uint(3, 100_000) // quality, timelength (ms)
	for _, s := range []struct {
		q, codec, size uint64
		url            string
	}{{80, 12, 1_250_000, "https://u.bilivideo.com/v80"}, {64, 7, 500_000, "https://u.bilivideo.com/v64"}} {
		var si, dv, item pbWriter
		si.uint(1, s.q)
		dv.string(1, s.url).string(2, s.url+"-b").uint(4, s.codec).uint(6, s.size)
		item.message(1, &si).message(2, &dv)
		info.message(5, &item)
	}
	var segOnly, segInfo pbWriter
	segInfo.uint(1, 16)
	segOnly.message(1, &segInfo).bytes(3, nil)
	info.message(5, &segOnly)
	info.message(6, dashItem(30280, "https://u.bilivideo.com/a", 320_000))
	var dolby, flac pbWriter
	dolby.uint(1, 1).message(2, dashItem(30250, "https://u.bilivideo.com/d", 640_000))
	flac.message(2, dashItem(30251, "https://u.bilivideo.com/f", 1_500_000))
	info.message(7, &dolby).message(9, &flac)
	reply.message(1, &info)

	var business pbWriter
	for _, c := range []struct {
		start, end int64
		toast      string
	}{{0, 90, "即将跳过片头"}, {1300, 1390, "即将跳过片尾"}} {
		var clip pbWriter
		clip.int(2, c.start).int(3, c.end).string(5, c.toast)
		business.message(6, &clip)
	}
	reply.message(3, &business)

	var background, role, material, material2, dubbing, ext pbWriter
	background.string(1, "bg").message(7, dashItem(1, "https://u.bilivideo.com/bg-lo", 64_000)).message(7, dashItem(2, "https://u.bilivideo.com/bg-hi", 192_000))
	material.string(1, "r1").string(2, "角色").string(5, "声优").message(7, dashItem(3, "https://u.bilivideo.com/r1", 128_000))
	material2.string(1, "r2").string(3, "版本").message(7, dashItem(4, "https://u.bilivideo.com/r2", 128_000))
	role.message(4, &material).message(4, &material2)
	dubbing.message(1, &background).message(2, &role)
	ext.message(1, &dubbing)
	reply.message(7, &ext)
	return reply.buf
}

func TestPlayViewReplyReadsLikeTheWebAPI(t *testing.T) {
	t.Parallel()
	reply, err := decodeProto(playViewReply())
	if err != nil {
		t.Fatal(err)
	}
	j := playViewJSON(reply)
	video := j.Path("data", "dash", "video").Array()
	if len(video) != 2 || video[0].Get("id").IntOr(0) != 80 || video[0].Get("bandwidth").IntOr(0) != 1_250_000*8/100 ||
		video[0].Get("codecid").IntOr(0) != 12 || video[1].Get("backup_url").At(0).String() != "https://u.bilivideo.com/v64-b" {
		t.Fatalf("%s", j.Path("data", "dash", "video").Serialized())
	}
	var codecs []string
	for _, a := range j.Path("data", "dash", "audio").Array() {
		codecs = append(codecs, a.Get("codecs").String())
	}
	if !slices.Equal(codecs, []string{"M4A", "FLAC", "E-AC-3"}) || j.Path("data", "timelength").IntOr(0) != 100_000 {
		t.Fatalf("%v", codecs)
	}
	if clips := j.Path("data", "clip_info_list").Array(); len(clips) != 2 || clips[1].Get("start").IntOr(0) != 1300 || clips[0].Get("toastText").String() != "即将跳过片头" {
		t.Fatalf("%s", j.Path("data", "clip_info_list").Serialized())
	}
	roles := j.Path("dubbing_info", "role_audio_list").Array()
	if len(roles) != 2 || roles[0].Get("title").String() != "角色" || roles[0].Get("person_name").String() != "声优" ||
		roles[1].Get("title").String() != "r2" || roles[1].Get("person_name").String() != "版本" || len(j.Path("dubbing_info", "background_audio").Array()) != 2 {
		t.Fatalf("%s", j.Get("dubbing_info").Serialized())
	}
	// Without a reply everything is empty rather than missing.
	empty := playViewJSON(nil)
	if !empty.Path("data", "dash", "video").IsArray() || empty.Path("data", "timelength").IntOr(-1) != 0 {
		t.Fatalf("%s", empty.Serialized())
	}
}

// grpcAnswer answers a gRPC call with a gzip frame of message, and checks the request is a frame too.
func grpcAnswer(t *testing.T, message []byte, got *[]byte) testkit.Handler {
	return func(r *http.Request) testkit.Response {
		body, _ := io.ReadAll(r.Body)
		req, err := unpackFrame(body)
		if err != nil {
			t.Errorf("request: %v", err)
		}
		if got != nil {
			*got = req
		}
		return testkit.Response{Header: http.Header{"Content-Type": {"application/grpc"}}, Body: packFrame(message)}
	}
}

func TestAppAPIBangumiWithDubbing(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = App; o.PreferredCodec = "AVC" })
	var sent []byte
	stub.On(appPGCEndpoint, grpcAnswer(t, playViewReply(), &sent))
	s := plainSession()
	s.token = "tok"
	f, err := e.streams(ctx, s, &ref{kind: episode, aid: "1", cid: "178175633", epid: "317690", api: App})
	if err != nil {
		t.Fatal(err)
	}
	// Bangumi is asked by episode, in HEVC whatever was preferred.
	if req, _ := decodeProto(sent); req.int(1) != 317690 || req.int(2) != 178175633 || req.int(12) != code265 {
		t.Fatalf("%+v", req)
	}
	calls := stub.Requests(appPGCEndpoint)
	if len(calls) != 1 || calls[0].Host != "grpc.biliapi.net" || calls[0].Header.Get("Authorization") != "identify_v1 tok" || calls[0].Method != http.MethodPost {
		t.Fatalf("%d calls", len(calls))
	}
	ids, codecs := videoIDs(f)
	if !slices.Equal(ids, []string{"80", "64"}) || !slices.Equal(codecs, []string{"HEVC", "AVC"}) || f.Video[0].Bitrate != 100 || f.Video[0].Width != 0 {
		t.Fatalf("%v %v %+v", ids, codecs, f.Video[0])
	}
	if !strings.HasPrefix(f.Video[0].Source.URL, "http://"+backupHost+"/v80") {
		t.Fatalf("%s", f.Video[0].Source.URL)
	}
	var audio []string
	for _, a := range f.Audio {
		audio = append(audio, a.Codec)
	}
	if !slices.Equal(audio, []string{"M4A", "FLAC", "E-AC-3"}) {
		t.Fatalf("%v", audio)
	}
	if len(f.Chapters) != 3 || f.Chapters[0].Title != "片头" || f.Chapters[1] != (media.Chapter{Title: "Main", Start: 90 * time.Second, End: 1300 * time.Second}) {
		t.Fatalf("%+v", f.Chapters)
	}
	// Dubbing: the background, then each role, each with its best stream.
	if len(f.ExtraAudio) != 3 {
		t.Fatalf("%+v", f.ExtraAudio)
	}
	bg, r1, r2 := f.ExtraAudio[0], f.ExtraAudio[1], f.ExtraAudio[2]
	if bg.Title != "Background audio" || bg.Audio.Bitrate != 192 || !strings.HasSuffix(bg.Audio.Source.URL, "/bg-hi") {
		t.Fatalf("%+v", bg)
	}
	if r1.Title != "角色" || r1.Artist != "声优" || r2.Title != "r2" || r2.Artist != "版本" || r1.Audio.Source.Header.Get("User-Agent") != "Mozilla/5.0" {
		t.Fatalf("%+v %+v", r1, r2)
	}
	if !strings.Contains(f.Raw, `"dubbing_info":{`) {
		t.Fatal("raw is not the reshaped JSON")
	}
}

func TestAppAPIVideoAsksByAidInThePreferredCodec(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = App; o.PreferredCodec = "AV1" })
	var sent []byte
	stub.On(appUGCEndpoint, grpcAnswer(t, playViewReply(), &sent))
	f, err := e.streams(ctx, plainSession(), &ref{kind: video, aid: "626497566", cid: "220355130", api: App})
	if err != nil {
		t.Fatal(err)
	}
	if req, _ := decodeProto(sent); req.int(1) != 626497566 || req.int(12) != codeAV1 {
		t.Fatalf("%+v", req)
	}
	// One pass; dubbing and markers belong to bangumi only.
	if len(stub.Requests(appUGCEndpoint)) != 1 || len(f.ExtraAudio) != 0 || len(f.Chapters) != 0 {
		t.Fatalf("%d calls, %+v", len(stub.Requests(appUGCEndpoint)), f.ExtraAudio)
	}
}

func TestAppAnswerThatIsNoFrame(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = App })
	stub.On(appUGCEndpoint, func(*http.Request) testkit.Response { return testkit.Text("oops") })
	if _, err := e.streams(ctx, plainSession(), &ref{kind: video, aid: "1", cid: "2", api: App}); err == nil {
		t.Fatal("no error")
	}
}
