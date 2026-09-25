package bilibili

import (
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

func videoIDs(f *media.Formats) (ids, codecs []string) {
	for _, v := range f.Video {
		ids = append(ids, v.ID)
		codecs = append(codecs, v.Codec)
	}
	return
}

func TestWebDashTwoPassesMergedAndSigned(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.ReplaceHost, o.ForceHTTP = false, false })
	stubCaptured(t, stub)
	s := plainSession()
	s.wbi = "ea1db124af3c7062474693fa704f4ff8"
	f, err := e.streams(ctx, s, &ref{kind: video, aid: "626497566", cid: "220355130", api: Web})
	if err != nil {
		t.Fatal(err)
	}
	// Two passes (qn=0, then the maximum) are merged without duplicates.
	ids, codecs := videoIDs(f)
	if !slices.Equal(ids, []string{"32", "32", "16", "16"}) || !slices.Equal(codecs, []string{"AVC", "HEVC", "HEVC", "AVC"}) {
		t.Fatalf("%v %v", ids, codecs)
	}
	v := f.Video[0]
	if v.Quality != "480P" || v.Rank != 32 || v.Width != 852 || v.Height != 480 || v.FPS != 29.412 || v.Bitrate != 756 || v.HasAudio || v.Size != 0 {
		t.Fatalf("%+v", v)
	}
	if !strings.HasPrefix(v.Source.URL, "https://upos-sz-estgoss.bilivideo.com/upgcxcode/30/51/220355130/220355130_nb2-1-30032.m4s?") {
		t.Fatalf("url %s", v.Source.URL)
	}
	var audio []string
	for _, a := range f.Audio {
		audio = append(audio, a.ID+"/"+a.Codec+"/"+itoa(a.Bitrate))
	}
	if !slices.Equal(audio, []string{"30216/M4A/67", "30232/M4A/132", "30280/M4A/319"}) {
		t.Fatalf("audio %v", audio)
	}
	if !strings.Contains(f.Raw, `"dash":{`) || len(f.Chapters) != 0 || len(f.ExtraAudio) != 0 {
		t.Fatalf("raw / chapters / extra audio")
	}
	calls := stub.Requests("avid=626497566&cid=220355130")
	var qns []string
	for _, r := range calls {
		u := r.URL.String()
		qns = append(qns, r.URL.Query().Get("qn"))
		query := u[strings.Index(u, "?")+1 : strings.Index(u, "&w_rid=")]
		if !strings.HasSuffix(u, "&w_rid="+md5Hex(query+s.wbi)) || !strings.Contains(u, "try_look=1") {
			t.Errorf("not WBI-signed: %s", u)
		}
		if r.Header.Get("Referer") != "https://www.bilibili.com/" || r.Header.Get("User-Agent") != "UA" {
			t.Errorf("headers %v", r.Header)
		}
	}
	if !slices.Equal(qns, []string{"0", "127"}) {
		t.Fatalf("qn %v", qns)
	}
}

func TestStreamsGoToTheMirrorOverHTTPWithBilibilisHeaders(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	s := plainSession()
	s.cookie = "SESSDATA=x"
	f, err := e.streams(ctx, s, &ref{kind: video, aid: "626497566", cid: "220355130", api: Web})
	if err != nil {
		t.Fatal(err)
	}
	var resources []media.Resource
	for _, v := range f.Video {
		resources = append(resources, v.Source)
	}
	for _, a := range f.Audio {
		resources = append(resources, a.Source)
	}
	for _, r := range resources {
		if !strings.HasPrefix(r.URL, "http://"+backupHost+"/upgcxcode/") || r.Size != 0 || r.Policy != media.Parallel {
			t.Errorf("%+v", r)
		}
		if r.Header.Get("Referer") != "https://www.bilibili.com" || r.Header.Get("User-Agent") != "Mozilla/5.0" || r.Header.Get("Cookie") != "SESSDATA=x" {
			t.Errorf("header %v", r.Header)
		}
	}
	// The cookie also went to the playurl, which then does not ask for a preview.
	if u := stub.Requests("x/player/wbi/playurl")[0]; strings.Contains(u.URL.String(), "try_look") || u.Header.Get("Cookie") != "SESSDATA=x" {
		t.Fatalf("%s %v", u.URL, u.Header)
	}
}

func TestBangumiPreviewIsFLVWithChapters(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stubCaptured(t, stub)
	f, err := e.streams(ctx, plainSession(), &ref{kind: episode, aid: "840009001", cid: "178175633", epid: "317690", api: Web})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Video) != 1 || len(f.Audio) != 0 {
		t.Fatalf("%d video, %d audio", len(f.Video), len(f.Audio))
	}
	v := f.Video[0]
	if v.ID != "32" || v.Codec != "AVC" || v.Size != 10228474 || !v.HasAudio || v.Source.URL != "" || len(v.Parts) != 1 || v.Bitrate != 10228474*8/60101 {
		t.Fatalf("%+v", v)
	}
	// Segments keep their host (only DASH streams move to the mirror) but get the scheme and headers.
	if p := v.Parts[0]; !strings.HasPrefix(p.URL, "http://upos-sz-estgcos.bilivideo.com/") || p.Header.Get("Referer") != "https://www.bilibili.com" {
		t.Fatalf("%+v", p)
	}
	// Opening / ending markers become chapters with the main part in between.
	want := []media.Chapter{
		{Title: "片头", Start: 0, End: 24 * time.Second},
		{Title: "Main", Start: 24 * time.Second, End: 145 * time.Second},
		{Title: "片尾", Start: 145 * time.Second, End: 169 * time.Second},
	}
	if !slices.Equal(f.Chapters, want) {
		t.Fatalf("%+v", f.Chapters)
	}
	// Episodes: the pgc endpoint, unsigned, with the episode id; FLV is asked again at the top quality.
	calls := stub.Requests("pgc/player/web/v2/playurl")
	if len(calls) != 2 || calls[1].URL.Query().Get("qn") != "127" {
		t.Fatalf("%d calls", len(calls))
	}
	for _, c := range calls {
		if u := c.URL.String(); strings.Contains(u, "w_rid") || !strings.Contains(u, "ep_id=317690&session=") || c.Header.Get("Cookie") != "" {
			t.Errorf("%s %v", u, c.Header)
		}
	}
}

func TestOpeningAndEndingChapters(t *testing.T) {
	t.Parallel()
	clips, _ := jsonv.ParseString(`[{"start":1300,"end":1390,"toastText":"即将跳过片尾"},{"start":90,"end":180,"toastText":"即将跳过片头"}]`)
	got := openingAndEnding(clips)
	sec := func(n int) time.Duration { return time.Duration(n) * time.Second }
	want := []media.Chapter{
		{Title: "Main", Start: 0, End: sec(90)},
		{Title: "片头", Start: sec(90), End: sec(180)},
		{Title: "Main", Start: sec(180), End: sec(1300)},
		{Title: "片尾", Start: sec(1300), End: sec(1390)},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("%+v", got)
	}
	if openingAndEnding(jsonv.Null) != nil {
		t.Fatal("chapters from nothing")
	}
}

func TestPCDNHostsAreAvoided(t *testing.T) {
	t.Parallel()
	node, _ := jsonv.ParseString(`{"base_url": "http://1.2.3.4:4480/v.m4s", "backup_url": ["https://upos-sz-mirror.bilivideo.com/v.m4s"]}`)
	if got := pickURL(node); got != "https://upos-sz-mirror.bilivideo.com/v.m4s" {
		t.Fatal(got)
	}
	only, _ := jsonv.ParseString(`{"base_url": "http://1.2.3.4:4480/v.m4s"}`)
	if got := pickURL(only); got != "http://1.2.3.4:4480/v.m4s" {
		t.Fatal(got)
	}
}

func TestRewritesCDNHosts(t *testing.T) {
	t.Parallel()
	rewrite := func(configure func(*Options), u, area string) string {
		o := DefaultOptions()
		o.ReplaceHost = false // on by default: then every stream goes to the mirror
		configure(&o)
		return rewriteHost(u, &o, area)
	}
	none := func(*Options) {}
	cases := []struct {
		configure func(*Options)
		url, area string
		want      string
	}{
		{none, "http://1.2.3.4:4480/upgcxcode/v.m4s", "", "http://" + backupHost + "/upgcxcode/v.m4s"},
		{func(o *Options) { o.AllowPCDN = true }, "http://1.2.3.4:4480/v.m4s", "", "http://1.2.3.4:4480/v.m4s"},
		{none, "https://upos-hz.akamaized.net/v.m4s", "hk", "https://" + backupHost + "/v.m4s"},
		{none, "https://upos-hz.akamaized.net/v.m4s", "", "https://upos-hz.akamaized.net/v.m4s"},
		{func(o *Options) { o.UposHost = "cn.test" }, "https://a.bilivideo.com/v.m4s", "", "https://cn.test/v.m4s"},
		{func(o *Options) { o.UposHost = "cn.test"; o.ReplaceHost = true }, "https://a.bilivideo.com/v.m4s", "", "https://cn.test/v.m4s"},
		{func(o *Options) { o.ReplaceHost = true }, "https://a.bilivideo.com/v.m4s", "", "https://" + backupHost + "/v.m4s"},
		{func(o *Options) { o.ReplaceHost = true; o.AllowPCDN = true }, "http://1.2.3.4:4480/v.m4s", "", "http://" + backupHost + "/v.m4s"},
	}
	for i, c := range cases {
		if got := rewrite(c.configure, c.url, c.area); got != c.want {
			t.Errorf("%d: %s → %s, want %s", i, c.url, got, c.want)
		}
	}
}

func TestResourcePolicyAndHeaders(t *testing.T) {
	t.Parallel()
	e := New(nil, DefaultOptions())
	s := plainSession()
	// cmcc hosts break on ranges and on plain HTTP: one request, as given.
	e.opts.ReplaceHost = false
	if r := e.resource(s, "https://cn-gdfs-cmcc-bcache-01.bilivideo.com/v.m4s", true); r.Policy != media.Whole || !strings.HasPrefix(r.URL, "https:") {
		t.Fatalf("%+v", r)
	}
	// mcdn hosts only speak what they advertise.
	if r := e.resource(s, "https://xy1x2x3x4xy.mcdn.bilivideo.cn:4483/v.m4s", false); r.URL != "https://xy1x2x3x4xy.mcdn.bilivideo.cn:4483/v.m4s" || r.Policy != media.Parallel {
		t.Fatalf("%+v", r)
	}
	// The Android and TV clients' URLs are refused with a Referer; logged out, there is no cookie.
	r := e.resource(s, "https://upos.bilivideo.com/v.m4s?platform=android_tv_yst", true)
	if r.URL != "http://upos.bilivideo.com/v.m4s?platform=android_tv_yst" || r.Header.Get("Referer") != "" || r.Header.Get("Cookie") != "" || r.Header.Get("User-Agent") != "Mozilla/5.0" {
		t.Fatalf("%+v", r)
	}
	e.opts.ForceHTTP = false
	if r := e.resource(s, "https://upos.bilivideo.com/v.m4s", true); r.URL != "https://upos.bilivideo.com/v.m4s" {
		t.Fatalf("%+v", r)
	}
	if forceHTTP("http://a/b?u=https://c") != "http://a/b?u=https://c" {
		t.Fatal("rewrote inside the query")
	}
}

func TestTVAnswersHaveNoGeometryAndNoDolby(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = TV })
	stub.On(defaultTVHost+"/x/tv/playurl?", answer(`{"code":0,"data":{"timelength":1000,"dash":{"video":[
	  {"id":80,"base_url":"https://u.bilivideo.com/v?platform=android_tv_yst","codecid":7,"bandwidth":2000000,"width":1920,"height":1080,"frame_rate":"25"}],
	  "audio":[{"id":30280,"base_url":"https://u.bilivideo.com/a","codecs":"mp4a.40.2","bandwidth":320000}],
	  "dolby":{"audio":[{"id":30250,"base_url":"https://u.bilivideo.com/d","codecs":"ec-3","bandwidth":640000}]}}}}`))
	s := plainSession()
	s.token = "tok"
	f, err := e.streams(ctx, s, &ref{kind: video, aid: "1", cid: "2", api: TV})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Video) != 1 || f.Video[0].Width != 0 || f.Video[0].Quality != "1080P" || len(f.Audio) != 1 || f.Audio[0].Codec != "M4A" {
		t.Fatalf("%+v %+v", f.Video, f.Audio)
	}
	if f.Video[0].Source.Header.Get("Referer") != "" {
		t.Fatal("TV stream sent a Referer")
	}
	calls := stub.Requests("x/tv/playurl")
	if len(calls) != 2 || !strings.HasPrefix(calls[0].URL.RawQuery, "access_key=tok&appkey="+tvAppKey) {
		t.Fatalf("%d calls", len(calls))
	}
}

func TestDolbyAndFlacAudio(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("x/player/wbi/playurl?", answer(`{"code":0,"data":{"dash":{"video":[],
	  "audio":[{"id":30280,"base_url":"https://u.bilivideo.com/a","codecs":"mp4a.40.2","bandwidth":320000}],
	  "dolby":{"audio":[{"id":30250,"base_url":"https://u.bilivideo.com/d","codecs":"ec-3","bandwidth":640000}]},
	  "flac":{"audio":{"id":30251,"base_url":"https://u.bilivideo.com/f","codecs":"fLaC","bandwidth":1500000}}}}}`))
	f, err := e.streams(ctx, plainSession(), &ref{kind: video, aid: "1", cid: "2", api: Web})
	if err != nil {
		t.Fatal(err)
	}
	var codecs []string
	for _, a := range f.Audio {
		codecs = append(codecs, a.Codec)
	}
	if !slices.Equal(codecs, []string{"M4A", "E-AC-3", "FLAC"}) {
		t.Fatalf("%v", codecs)
	}
}

func TestIntlStreamListBothCodecs(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.API = Intl })
	stub.On("prefer_code_type=0", answer(`{"code":0,"data":{"video_info":{"timelength":60000,"stream_list":[
	  {"stream_info":{"quality":80},"dash_video":{"base_url":"https://u.akamaized.net/v80","codecid":7,"bandwidth":1000000,"size":99}},
	  {"stream_info":{"quality":112},"dash_video":{"base_url":""}}],
	  "dash_audio":[{"id":30280,"base_url":"https://u.akamaized.net/a","bandwidth":128000}]}}}`))
	stub.On("prefer_code_type=1", answer(`{"code":0,"data":{"video_info":{"timelength":60000,"stream_list":[
	  {"stream_info":{"quality":80},"dash_video":{"base_url":"https://u.akamaized.net/v80h","codecid":12,"bandwidth":800000}}],
	  "dash_audio":[{"id":30280,"base_url":"https://u.akamaized.net/a","bandwidth":128000}]}}}`))
	f, err := e.streams(ctx, plainSession(), &ref{kind: episode, aid: "1", cid: "2", epid: "3", api: Intl})
	if err != nil {
		t.Fatal(err)
	}
	ids, codecs := videoIDs(f)
	if !slices.Equal(ids, []string{"80", "80"}) || !slices.Equal(codecs, []string{"AVC", "HEVC"}) || f.Video[0].Size != 99 || len(f.Audio) != 1 || f.Audio[0].Codec != "M4A" {
		t.Fatalf("%v %v %+v", ids, codecs, f.Audio)
	}
}

func TestPremiumOnlyFallsBackToThePagesPlayinfo(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("pgc/player/web/v2/playurl?", answer(`{"code":-10403,"message":"大会员专享限制"}`))
	stub.On("www.bilibili.com/bangumi/play/ep5", func(*http.Request) testkit.Response {
		return testkit.Text(`<script>window.__playinfo__={"code":0,"data":{"dash":{"video":[{"id":80,"base_url":"https://u.bilivideo.com/v","codecid":12,"bandwidth":1000}],"audio":[]}}}</script>`)
	})
	f, err := e.streams(ctx, plainSession(), &ref{kind: episode, aid: "1", cid: "2", epid: "5", api: Web})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Video) != 1 || f.Video[0].Codec != "HEVC" {
		t.Fatalf("%+v", f.Video)
	}
}

func TestNoStreamsSaysWhy(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(nil)
	stub.On("x/player/wbi/playurl?", answer(`{"code":87008,"message":"当前视频仅限充电专属"}`))
	s := e.session(ctx) // the nav check fails on the stub: a warning, not an error
	if s.loggedIn != nil {
		t.Fatal("login state from a failed check")
	}
	entry := &media.Entry{Title: "T", Ref: &ref{kind: video, aid: "1", cid: "2", api: Web}}
	_, err := e.Formats(ctx, &media.Item{}, entry)
	if err == nil || errs.KindOf(err) != errs.Failed || !strings.Contains(err.Error(), `"T": bilibili answered 87008: 当前视频仅限充电专属`) {
		t.Fatalf("err = %v", err)
	}
	if _, err := e.Formats(ctx, &media.Item{}, &media.Entry{ID: "x"}); err == nil {
		t.Fatal("an entry of another site")
	}
	carried := &media.Formats{Raw: "kept"}
	if f, _ := e.Formats(ctx, &media.Item{}, &media.Entry{Formats: carried}); f != carried {
		t.Fatal("carried formats were reloaded")
	}
}
