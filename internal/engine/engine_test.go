package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/fetch"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

// fakeSite is an extractor over canned items: links are "test:<name>".
type fakeSite struct {
	items   map[string]*media.Item
	lazy    map[string]*media.Formats // formats loaded on demand, by entry id
	loads   atomic.Int32
	failFor int32 // the first failFor loads fail
}

func (f *fakeSite) Info() extract.Info {
	return extract.Info{Site: "test", Name: "Test", Unit: "video", OwnerLabel: "by"}
}

func (f *fakeSite) Match(link string) (string, bool) {
	return strings.TrimPrefix(link, "test:"), strings.HasPrefix(link, "test:")
}

func (f *fakeSite) Resolve(_ context.Context, link string) (*media.Item, error) {
	if it, ok := f.items[link]; ok {
		return it, nil
	}
	return nil, errs.New("no such item %s", link)
}

func (f *fakeSite) Formats(_ context.Context, _ *media.Item, e *media.Entry) (*media.Formats, error) {
	if n := f.loads.Add(1); n <= f.failFor {
		return nil, errs.New("temporary failure %d", n)
	}
	return f.lazy[e.ID], nil
}

type fixture struct {
	t    *testing.T
	stub *testkit.Stub
	site *fakeSite
	dir  string
	home string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HAUL_HOME", home)
	return &fixture{t: t, stub: testkit.NewStub(), site: &fakeSite{items: map[string]*media.Item{}, lazy: map[string]*media.Formats{}},
		dir: t.TempDir(), home: home}
}

func (f *fixture) engine(configure func(*Options)) *Engine {
	o := DefaultOptions()
	o.Dir = f.dir
	if configure != nil {
		configure(&o)
	}
	return &Engine{Router: extract.NewRouter(f.site), Client: f.stub.Client(), Fetcher: fetch.NewHTTP(f.stub.Client()),
		FFmpeg: testkit.FFmpeg, Options: o}
}

func (f *fixture) run(link string, configure func(*Options)) (*Result, error) {
	return f.engine(configure).Run(context.Background(), link)
}

func (f *fixture) files() []string {
	var out []string
	filepath.Walk(f.dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(f.dir, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func eq[T any](t *testing.T, what string, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %#v, want %#v", what, got, want)
	}
}

// clip is a YouTube-like item: DASH video and audio behind a range-only server, a json3 subtitle, a cover, chapters.
func (f *fixture) clip(host string, m testkit.Media) *media.Item {
	f.stub.On(host+"/v", rangeOnly(m.Video))
	f.stub.On(host+"/a", rangeOnly(m.Audio))
	f.stub.On(host+"/sub", func(*http.Request) testkit.Response {
		return testkit.JSON(`{"events":[{"tStartMs":0,"dDurationMs":900,"segs":[{"utf8":"Hello"}]}]}`)
	})
	f.stub.On(host+"/cover.jpg", func(*http.Request) testkit.Response { return testkit.Response{Body: m.Cover} })
	res := func(path string) media.Resource {
		return media.Resource{URL: "https://" + host + path, Header: httpx.Header("User-Agent", "UA-YT"), Policy: media.Sequential, MaxRange: 10 << 20}
	}
	entry := &media.Entry{Index: 1, ID: "ytTestVid01", Title: "Clip", Duration: time.Second, URL: "https://www.youtube.com/watch?v=ytTestVid01",
		Uploader: media.Person{Name: "Channel", ID: "UC1"}, Published: time.Unix(1790186492, 0),
		Chapters: []media.Chapter{{Title: "Intro", End: time.Second}},
		Formats: &media.Formats{
			Video: []media.VideoFormat{{ID: "134", Rank: 360010, Quality: "360p", Width: 64, Height: 64, FPS: 10, Codec: "AVC", Bitrate: 200, Source: res("/v")}},
			Audio: []media.AudioFormat{{ID: "140", Codec: "M4A", Bitrate: 130, Source: res("/a")}},
			Subtitles: []media.Subtitle{
				{Lang: "en", Format: media.YouTubeJSON3, Source: res("/sub")},
				{Lang: "en", Auto: true, Format: media.YouTubeJSON3, Source: res("/sub?auto")},
			},
		}}
	return &media.Item{Site: "test", ID: "ytTestVid01", Title: "YouTube Clip", Uploader: entry.Uploader, Thumbnail: "https://" + host + "/cover.jpg",
		URL: entry.URL, Published: entry.Published, Entries: []*media.Entry{entry}}
}

// rangeOnly is googlevideo: closed ranges only.
func rangeOnly(data []byte) testkit.Handler {
	return func(r *http.Request) testkit.Response {
		if _, to, ok := testkit.Range(r); !ok || to < 0 {
			return testkit.Status(403)
		}
		return testkit.Ranged(data, r)
	}
}

func TestClipEndToEnd(t *testing.T) {
	m := testkit.MakeMedia(t)
	f := newFixture(t)
	f.site.items["clip"] = f.clip("yt.test", m)
	res, err := f.run("test:clip", nil)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(f.dir, "YouTube Clip.mp4")
	p := testkit.ProbeFile(t, out)
	eq(t, "streams", p.Streams, []string{"video:h264", "audio:aac", "subtitle:mov_text", "video:mjpeg"})
	eq(t, "chapters", p.Chapters, []string{"Intro"})
	eq(t, "comment", p.Tags["comment"], "https://www.youtube.com/watch?v=ytTestVid01")
	eq(t, "artist", p.Tags["artist"], "Channel")
	eq(t, "languages", p.Languages, []string{"eng"})
	// Closed ranges only, with the resource's headers.
	for _, r := range append(f.stub.Requests("yt.test/v"), f.stub.Requests("yt.test/a")...) {
		if _, to, ok := testkit.Range(r); !ok || to < 0 || r.Header.Get("User-Agent") != "UA-YT" {
			t.Errorf("request %s %v", r.URL, r.Header)
		}
	}
	// The auto subtitle was not wanted.
	eq(t, "auto subtitle requests", len(f.stub.Requests("sub?auto")), 0)

	e := res.Entries[0]
	eq(t, "status", e.Status, Downloaded)
	eq(t, "file", e.File, absPath(out))
	eq(t, "files", res.Files(), []string{absPath(out)})
	eq(t, "subtitles", len(e.Subtitles), 1)
	if e.Size <= 0 || e.ChosenVideo != 0 || e.ChosenAudio != 0 {
		t.Errorf("result %+v", e)
	}
	eq(t, "work files left", f.files(), []string{"YouTube Clip.mp4"})

	// Again: skipped, the file still reported, nothing downloaded or left behind.
	before := len(f.stub.Requests("yt.test"))
	res, err = f.run("test:clip", nil)
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "status again", res.Entries[0].Status, Skipped)
	eq(t, "reason", res.Entries[0].Reason, "exists")
	eq(t, "files again", res.Files(), []string{absPath(out)})
	eq(t, "requests again", len(f.stub.Requests("yt.test")), before)
	eq(t, "files after skip", f.files(), []string{"YouTube Clip.mp4"})
}

func TestMissingCoverAndSubtitleCostOnlyThemselves(t *testing.T) {
	m := testkit.MakeMedia(t)
	f := newFixture(t)
	item := f.clip("flaky.test", m)
	f.stub.On("flaky.test/sub", func(*http.Request) testkit.Response { return testkit.Status(429) })
	f.stub.On("flaky.test/cover.jpg", func(*http.Request) testkit.Response { return testkit.Status(404) })
	f.site.items["clip"] = item
	if _, err := f.run("test:clip", nil); err != nil {
		t.Fatal(err)
	}
	eq(t, "streams", testkit.ProbeFile(t, filepath.Join(f.dir, "YouTube Clip.mp4")).Streams, []string{"video:h264", "audio:aac"})
}

func TestListingStreamsDownloadsNothing(t *testing.T) {
	f := newFixture(t)
	f.site.items["clip"] = f.clip("list.test", testkit.Media{})
	res, err := f.run("test:clip", func(o *Options) { o.List = true; o.IncludeURLs = true })
	if err != nil {
		t.Fatal(err)
	}
	e := res.Entries[0]
	eq(t, "status", e.Status, Listed)
	eq(t, "video", e.Video[0].Source.URL, "https://list.test/v")
	eq(t, "subtitles", e.Subtitles[0].Lang, "en")
	eq(t, "files", res.Files(), []string(nil))
	eq(t, "requests", len(f.stub.Requests("list.test")), 0)
	eq(t, "dir", f.files(), []string(nil))
}

// Listing shows every track, auto ones marked; downloads leave auto ones out unless asked.
func TestSubtitleLanguagesAndAutoTracks(t *testing.T) {
	f := newFixture(t)
	f.site.items["clip"] = f.clip("subs.test", testkit.Media{})
	for _, c := range []struct {
		langs []string
		auto  bool
		want  int
	}{{nil, false, 2}, {nil, true, 2}, {[]string{"fr", "de"}, true, 0}, {[]string{"EN"}, false, 2}} {
		res, err := f.run("test:clip", func(o *Options) { o.List = true; o.SubtitleLangs = c.langs; o.AutoSubtitles = c.auto })
		if err != nil {
			t.Fatal(err)
		}
		eq(t, "subtitles", len(res.Entries[0].Subtitles), c.want)
	}
}

// post is an X-like item: whole files with the audio inside, one entry per video.
func (f *fixture) post(n int, m testkit.Media) *media.Item {
	f.stub.On("x.test/vid/", func(r *http.Request) testkit.Response { return testkit.Ranged(m.Combined, r) })
	f.stub.On("x.test/thumb/", func(*http.Request) testkit.Response { return testkit.Response{Body: m.Cover} })
	item := &media.Item{Site: "test", ID: "1790637656616943991", Title: "Someone - A post", Uploader: media.Person{Name: "Someone"},
		URL: "https://x.com/someone/status/1790637656616943991", Published: time.Unix(1715756260, 0)}
	for i := 1; i <= n; i++ {
		title := "Post"
		if n > 1 {
			title = "Video " + string(rune('0'+i))
		}
		item.Entries = append(item.Entries, &media.Entry{Index: i, ID: "v" + string(rune('0'+i)), Title: title, Duration: time.Second,
			Thumbnail: "https://x.test/thumb/" + string(rune('0'+i)) + ".jpg",
			Formats: &media.Formats{Video: []media.VideoFormat{{ID: "http-300", Rank: 64000, Quality: "64p", Width: 64, Height: 64,
				Codec: "AVC", Bitrate: 300, HasAudio: true, Source: media.Resource{URL: "https://x.test/vid/" + string(rune('0'+i)) + ".mp4"}}}}})
	}
	return item
}

func TestVideoWithAudioInside(t *testing.T) {
	m := testkit.MakeMedia(t)
	f := newFixture(t)
	f.site.items["post"] = f.post(1, m)
	if _, err := f.run("test:post", nil); err != nil {
		t.Fatal(err)
	}
	eq(t, "video", testkit.ProbeFile(t, filepath.Join(f.dir, "Someone - A post.mp4")).Streams, []string{"video:h264", "audio:aac", "video:mjpeg"})

	// --audio-only next to the finished video is its own file, not "already downloaded".
	res, err := f.run("test:post", func(o *Options) { o.Content.Tracks = AudioOnly })
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "audio status", res.Entries[0].Status, Downloaded)
	eq(t, "audio", testkit.ProbeFile(t, filepath.Join(f.dir, "Someone - A post.m4a")).Streams, []string{"audio:aac", "video:mjpeg"})

	if _, err := f.run("test:post", func(o *Options) { o.Content.Tracks = VideoOnly; o.Template = "<title> silent" }); err != nil {
		t.Fatal(err)
	}
	for _, s := range testkit.ProbeFile(t, filepath.Join(f.dir, "Someone - A post silent.mp4")).Streams {
		if strings.HasPrefix(s, "audio") {
			t.Errorf("video-only kept %s", s)
		}
	}
}

func TestSeveralEntriesGoInAFolder(t *testing.T) {
	m := testkit.MakeMedia(t)
	f := newFixture(t)
	f.site.items["post"] = f.post(2, m)
	if _, err := f.run("test:post", nil); err != nil {
		t.Fatal(err)
	}
	eq(t, "all", f.files(), []string{"Someone - A post/[P1]Video 1.mp4", "Someone - A post/[P2]Video 2.mp4"})

	f2 := newFixture(t)
	f2.site.items["post"] = f2.post(2, m)
	res, err := f2.run("test:post", func(o *Options) { o.Pages = "2" })
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "second", f2.files(), []string{"Someone - A post/[P2]Video 2.mp4"})
	eq(t, "selected", []bool{res.Entries[0].Selected, res.Entries[1].Selected}, []bool{false, true})
	eq(t, "tags", testkit.ProbeFile(t, filepath.Join(f2.dir, "Someone - A post", "[P2]Video 2.mp4")).Tags["album"], "Someone - A post")
}

func TestListingAListShowsEntriesUntilOneIsChosen(t *testing.T) {
	f := newFixture(t)
	f.site.items["post"] = f.post(3, testkit.Media{})
	res, err := f.run("test:post", func(o *Options) { o.List = true })
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range res.Entries {
		if e.Selected || e.Video != nil || e.Status != "" {
			t.Errorf("entry %+v", e)
		}
	}
	res, err = f.run("test:post", func(o *Options) { o.List = true; o.Pages = "2" })
	if err != nil {
		t.Fatal(err)
	}
	if e := res.Entries[1]; !e.Selected || e.Status != Listed || len(e.Video) != 1 || res.Entries[0].Video != nil {
		t.Errorf("entries %+v", res.Entries)
	}
}

func TestPageSelectionOutOfRangeStillReportsTheItem(t *testing.T) {
	f := newFixture(t)
	f.site.items["post"] = f.post(2, testkit.Media{})
	res, err := f.run("test:post", func(o *Options) { o.Pages = "99" })
	if !errs.Is(err, errs.Input) || res == nil || len(res.Entries) != 2 {
		t.Errorf("err = %v, res = %v", err, res)
	}
	_, err = f.run("test:post", func(o *Options) { o.Pages = "x" })
	if !errs.Is(err, errs.Input) {
		t.Errorf("bad spec: %v", err)
	}
}

func TestStreamIndexesAreChecked(t *testing.T) {
	f := newFixture(t)
	f.site.items["clip"] = f.clip("index.test", testkit.Media{})
	_, err := f.run("test:clip", func(o *Options) { o.List = true; o.AudioIndex = 3 })
	if !errs.Is(err, errs.Input) || !strings.Contains(err.Error(), "--audio-stream 3") {
		t.Errorf("err = %v", err)
	}
	res, err := f.run("test:clip", func(o *Options) { o.List = true; o.AudioIndex = 0; o.VideoIndex = 0 })
	if err != nil || res.Entries[0].ChosenAudio != 0 {
		t.Errorf("err = %v", err)
	}
}

func TestUnsupportedLinksAreInputErrors(t *testing.T) {
	f := newFixture(t)
	_, err := f.run("https://example.com/video/1", nil)
	if !errs.Is(err, errs.Input) || !strings.Contains(err.Error(), "Supported sites: Test") {
		t.Errorf("err = %v", err)
	}
}

func TestMissingFFmpegIsADependencyError(t *testing.T) {
	f := newFixture(t)
	f.site.items["clip"] = f.clip("deps.test", testkit.Media{})
	e := f.engine(nil)
	e.FFmpeg = ""
	if _, err := e.Run(context.Background(), "test:clip"); !errs.Is(err, errs.Dependency) {
		t.Errorf("err = %v", err)
	}
	e.Options.List = true
	if _, err := e.Run(context.Background(), "test:clip"); err != nil {
		t.Errorf("listing needs no ffmpeg: %v", err)
	}
}

// episode is a podcast-like item: an audio entry with a show.
func (f *fixture) episode(host string, audio []byte, codec string, m testkit.Media) *media.Item {
	f.stub.On(host+"/ep", func(r *http.Request) testkit.Response { return testkit.Ranged(audio, r) })
	f.stub.On(host+"/art.jpg", func(*http.Request) testkit.Response { return testkit.Response{Body: m.Cover} })
	entry := &media.Entry{Index: 1, ID: "1000671234567", Title: "Newer", Album: "The Daily", Uploader: media.Person{Name: "The New York Times"},
		Thumbnail: "https://" + host + "/art.jpg", Published: time.Unix(1726825500, 0),
		Formats: &media.Formats{AudioOnly: true, Audio: []media.AudioFormat{{ID: "0", Codec: codec,
			Source: media.Resource{URL: "https://" + host + "/ep", Header: httpx.Header("User-Agent", httpx.BrowserUserAgent)}}}}}
	return &media.Item{Site: "test", Title: "Newer", Entries: []*media.Entry{entry}}
}

func TestPodcastEpisodesKeepTheirFormat(t *testing.T) {
	m := testkit.MakeMedia(t)
	f := newFixture(t)
	f.site.items["mp3"] = f.episode("pod.test", m.MP3, "MP3", m)
	if _, err := f.run("test:mp3", nil); err != nil {
		t.Fatal(err)
	}
	p := testkit.ProbeFile(t, filepath.Join(f.dir, "Newer.mp3"))
	eq(t, "mp3 streams", p.Streams, []string{"audio:mp3", "video:mjpeg"})
	eq(t, "mp3 tags", []string{p.Tags["title"], p.Tags["album"], p.Tags["artist"]}, []string{"Newer", "The Daily", "The New York Times"})
	for _, r := range f.stub.Requests("pod.test/ep") {
		eq(t, "user agent", r.Header.Get("User-Agent"), httpx.BrowserUserAgent)
	}

	f2 := newFixture(t)
	f2.site.items["m4a"] = f2.episode("pod2.test", m.Audio, "M4A", m)
	if _, err := f2.run("test:m4a", nil); err != nil {
		t.Fatal(err)
	}
	eq(t, "m4a streams", testkit.ProbeFile(t, filepath.Join(f2.dir, "Newer.m4a")).Streams, []string{"audio:aac", "video:mjpeg"})
	if _, err := f2.run("test:m4a", func(o *Options) { o.Content.Tracks = VideoOnly }); !errs.Is(err, errs.Input) {
		t.Errorf("video-only of an audio episode: %v", err)
	}
}

func TestSkipMuxKeepsTheStreams(t *testing.T) {
	f := newFixture(t)
	audio := testkit.Pattern(300_000)
	f.site.items["ep"] = f.episode("keep.test", audio, "M4A", testkit.Media{Cover: []byte{0xFF, 0xD8, 0xFF}})
	e := f.engine(func(o *Options) { o.Content.NoMux = true; o.Content.Cover = Skip })
	e.FFmpeg = "" // not needed
	res, err := e.Run(context.Background(), "test:ep")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "files", f.files(), []string{"Newer.audio.m4a"})
	got, _ := os.ReadFile(filepath.Join(f.dir, "Newer.audio.m4a"))
	eq(t, "bytes", len(got), len(audio))
	eq(t, "reason", res.Entries[0].Reason, "no-mux")
	eq(t, "result files", res.Files(), []string{absPath(filepath.Join(f.dir, "Newer.audio.m4a"))})
}

func TestSubtitlesCoverAndSidecarsAlone(t *testing.T) {
	f := newFixture(t)
	item := f.clip("side.test", testkit.Media{Cover: []byte{0xFF, 0xD8, 0xFF, 0xE0}})
	var bases []string
	item.Entries[0].Formats.Sidecars = []media.Sidecar{{Kind: "danmaku", Write: func(_ context.Context, base string) ([]string, error) {
		bases = append(bases, base)
		path := base + ".xml"
		return []string{path}, os.WriteFile(path, []byte("<i/>"), 0o644)
	}}}
	f.site.items["clip"] = item
	e := f.engine(func(o *Options) {
		o.Content = Content{Tracks: NoTracks, Subtitles: Files, Cover: Files, Sidecars: []string{"danmaku", "chat"}}
	})
	e.FFmpeg = ""
	res, err := e.Run(context.Background(), "test:clip")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "files", f.files(), []string{"YouTube Clip.en.srt", "YouTube Clip.jpg", "YouTube Clip.xml"})
	srt, _ := os.ReadFile(filepath.Join(f.dir, "YouTube Clip.en.srt"))
	eq(t, "srt", string(srt), "1\n00:00:00,000 --> 00:00:00,900\nHello\n\n")
	eq(t, "extra files", len(res.Entries[0].ExtraFiles), 3)
	eq(t, "status", res.Entries[0].Status, Downloaded)
	eq(t, "no video downloaded", len(f.stub.Requests("side.test/v")), 0)
}

func TestNothingToSaveIsSkipped(t *testing.T) {
	f := newFixture(t)
	item := f.clip("empty.test", testkit.Media{})
	item.Entries[0].Formats.Subtitles = item.Entries[0].Formats.Subtitles[1:] // only the auto track
	f.site.items["clip"] = item
	e := f.engine(func(o *Options) { o.Content = Content{Tracks: NoTracks, Subtitles: Files, Cover: Skip} })
	res, err := e.Run(context.Background(), "test:clip")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "status", res.Entries[0].Status, Skipped)
	eq(t, "reason", res.Entries[0].Reason, "empty")
	eq(t, "files", f.files(), []string(nil))
}

func TestArchiveSkipsWhatWasDownloaded(t *testing.T) {
	f := newFixture(t)
	f.site.items["ep"] = f.episode("arch.test", testkit.Pattern(1000), "M4A", testkit.Media{})
	e := f.engine(func(o *Options) { o.Archive = true; o.Content.NoMux = true; o.Content.Cover = Skip })
	e.FFmpeg = ""
	if _, err := e.Run(context.Background(), "test:ep"); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(f.dir)
	res, err := e.Run(context.Background(), "test:ep")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "status", res.Entries[0].Status, Skipped)
	eq(t, "reason", res.Entries[0].Reason, "archive")
	archive, _ := os.ReadFile(filepath.Join(f.home, "archives.txt"))
	eq(t, "archive", string(archive), "test:1000671234567\n")
}

func TestFailedEntriesAreTriedAgainWithFreshFormats(t *testing.T) {
	f := newFixture(t)
	data := testkit.Pattern(2000)
	f.stub.On("retry.test/a", func(r *http.Request) testkit.Response { return testkit.Ranged(data, r) })
	entry := &media.Entry{Index: 1, ID: "lazy", Title: "Lazy"}
	f.site.items["lazy"] = &media.Item{Site: "test", Title: "Lazy", Entries: []*media.Entry{entry}}
	f.site.lazy["lazy"] = &media.Formats{AudioOnly: true, Audio: []media.AudioFormat{{ID: "a", Codec: "M4A", Source: media.Resource{URL: "https://retry.test/a"}}}}
	f.site.failFor = 2
	e := f.engine(func(o *Options) { o.Content.NoMux = true })
	e.FFmpeg = ""
	res, err := e.Run(context.Background(), "test:lazy")
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "loads", f.site.loads.Load(), int32(3))
	eq(t, "status", res.Entries[0].Status, Downloaded)

	f.site.loads.Store(0)
	f.site.failFor = 5
	if _, err := e.Run(context.Background(), "test:lazy"); !errs.Is(err, errs.Failed) || f.site.loads.Load() != 3 {
		t.Errorf("err = %v after %d loads", err, f.site.loads.Load())
	}
}

func TestSegmentsAreJoined(t *testing.T) {
	testkit.RequireFFmpeg(t)
	f := newFixture(t)
	// Two FLV segments made from the test video.
	m := testkit.MakeMedia(t)
	src := filepath.Join(t.TempDir(), "v.mp4")
	os.WriteFile(src, m.Video, 0o644)
	var parts []media.Resource
	for i := range 2 {
		flv := filepath.Join(t.TempDir(), "p.flv")
		if out, err := runFFmpeg("-i", src, "-c", "copy", "-f", "flv", flv); err != nil {
			t.Fatalf("%v %s", err, out)
		}
		data, _ := os.ReadFile(flv)
		host := "flv" + string(rune('0'+i)) + ".test"
		f.stub.On(host, func(r *http.Request) testkit.Response { return testkit.Ranged(data, r) })
		parts = append(parts, media.Resource{URL: "https://" + host + "/p.flv"})
	}
	entry := &media.Entry{Index: 1, ID: "flv", Title: "Old", Formats: &media.Formats{
		Video: []media.VideoFormat{{ID: "32", Rank: 32, Quality: "480P", Codec: "AVC", Parts: parts, HasAudio: true}}}}
	f.site.items["flv"] = &media.Item{Site: "test", Title: "Old", Entries: []*media.Entry{entry}}
	if _, err := f.run("test:flv", nil); err != nil {
		t.Fatal(err)
	}
	eq(t, "streams", testkit.ProbeFile(t, filepath.Join(f.dir, "Old.mp4")).Streams, []string{"video:h264"})
}

func TestResultJSONShape(t *testing.T) {
	// The Result is what --json prints through the CLI; here only that it survives encoding.
	f := newFixture(t)
	f.site.items["clip"] = f.clip("json.test", testkit.Media{})
	res, _ := f.run("test:clip", func(o *Options) { o.List = true })
	if _, err := json.Marshal(res.Entries[0].Video); err != nil {
		t.Error(err)
	}
	if !slices.Equal(res.Files(), nil) {
		t.Error("listing writes no files")
	}
}
