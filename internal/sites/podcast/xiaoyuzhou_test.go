package podcast

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

func episodeHTML(audio string, withData bool) string {
	next := fmt.Sprintf(`{"props":{"pageProps":{"episode":{"type":"EPISODE","eid":"%[1]s","pid":"%[2]s","title":" E42 声音的故事 ",
"description":"节目简介","duration":2712,"enclosure":{"url":"%[3]s"},
"media":{"id":"m1","size":43392000,"mimeType":"audio/mp4","source":{"mode":"PUBLIC","url":"%[3]s"}},
"pubDate":"2024-05-24T22:00:00.000Z","image":{"picUrl":"https://image.xyz.test/ep.jpg"},
"podcast":{"type":"PODCAST","pid":"%[2]s","title":"声东击西","author":"声动活泼","image":{"picUrl":"https://image.xyz.test/show.jpg"}}}},
"__N_SSP":true},"page":"/episode/[id]","query":{"id":"%[1]s"}}`, eid, pid, audio)
	script := ""
	if withData {
		script = `<script id="__NEXT_DATA__" type="application/json">` + next + `</script>`
	}
	return `<!DOCTYPE html><html><head><meta charset="utf-8"/><meta property="og:title" content="E42 声音的故事 &amp; 更多"/>
<meta content="https://media.xyz.test/og.mp3" property="og:audio"/></head>
<body><div id="__next"></div>` + script + `</body></html>`
}

var episodeAudio = "https://media.xyz.test/" + pid + "/ep.m4a"

var showHTML = `<html><body><script id="__NEXT_DATA__" type="application/json">{"props":{"pageProps":{"podcast":{"type":"PODCAST","pid":"` + pid + `",
"title":"声东击西","author":"声动活泼","description":"关于节目","image":{"picUrl":"https://image.xyz.test/show.jpg"},
"episodes":[
  {"eid":"e3","pid":"` + pid + `","title":"第三期","duration":60,"enclosure":{"url":"https://media.xyz.test/3.mp3"},"media":{"mimeType":"audio/mpeg"},"pubDate":"2024-03-01T00:00:00.000Z"},
  {"eid":"e2","pid":"` + pid + `","title":"第二期","duration":60,"enclosure":{"url":"https://media.xyz.test/2.m4a"},"pubDate":"2024-02-01T00:00:00.000Z","image":{"picUrl":"https://image.xyz.test/e2.jpg"}},
  {"eid":"e1","pid":"` + pid + `","title":"第一期","duration":60,"enclosure":{"url":"https://media.xyz.test/1.m4a"},"pubDate":"2024-01-01T00:00:00.000Z"},
  {"eid":"other","pid":"someoneelse1","title":"推荐","enclosure":{"url":"https://media.xyz.test/x.mp3"}}
]}}}}</script></body></html>`

func TestXiaoyuzhouEpisode(t *testing.T) {
	t.Parallel()
	item, err := parseXiaoyuzhouPage(episodeHTML(episodeAudio, true), xiaoyuzhouRef{id: eid, episode: true})
	if err != nil {
		t.Fatal(err)
	}
	if item.Site != "xiaoyuzhou" || item.Title != "E42 声音的故事" || item.Description != "节目简介" ||
		item.Thumbnail != "https://image.xyz.test/ep.jpg" || item.Collection || item.ID != eid {
		t.Errorf("item = %+v", item)
	}
	if !item.Published.Equal(time.Unix(1716588000, 0)) {
		t.Errorf("published = %v", item.Published)
	}
	if len(item.Entries) != 1 {
		t.Fatalf("entries = %d", len(item.Entries))
	}
	e := item.Entries[0]
	if e.Index != 1 || e.ID != eid || e.Duration != 2712*time.Second || e.Uploader != (media.Person{Name: "声动活泼", ID: pid}) ||
		e.Album != "声东击西" || e.URL != "https://www.xiaoyuzhoufm.com/episode/"+eid || e.Thumbnail != "https://image.xyz.test/ep.jpg" {
		t.Errorf("entry = %+v", e)
	}
	f := e.Formats
	if !f.AudioOnly || len(f.Video) != 0 || len(f.Audio) != 1 {
		t.Fatalf("formats = %+v", f)
	}
	a := f.Audio[0]
	// The stated size gives a bitrate for the table, but never drives the downloader.
	if a.Codec != "M4A" || a.Bitrate != 125 || a.Size != 0 {
		t.Errorf("audio = %+v", a)
	}
	want := media.Resource{URL: episodeAudio, Header: xiaoyuzhouHeader(), Policy: media.Parallel}
	if !reflect.DeepEqual(a.Source, want) || a.Source.Header.Get("Referer") != "https://www.xiaoyuzhoufm.com/" {
		t.Errorf("source = %+v", a.Source)
	}
}

func TestXiaoyuzhouFallsBackToOpenGraph(t *testing.T) {
	t.Parallel()
	ref := xiaoyuzhouRef{id: eid, episode: true}
	item, err := parseXiaoyuzhouPage(episodeHTML(episodeAudio, false), ref)
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "E42 声音的故事 & 更多" || item.Entries[0].URL != "https://www.xiaoyuzhoufm.com/episode/"+eid {
		t.Errorf("item = %+v", item)
	}
	if a := item.Entries[0].Formats.Audio[0]; a.Source.URL != "https://media.xyz.test/og.mp3" || a.Codec != "MP3" {
		t.Errorf("audio = %+v", a)
	}
	// Neither page data nor Open Graph: a paid episode, or a changed page.
	if _, err := parseXiaoyuzhouPage("<html></html>", ref); !errs.Is(err, errs.Failed) {
		t.Errorf("err = %v", err)
	}
}

func TestXiaoyuzhouShowOldestFirst(t *testing.T) {
	t.Parallel()
	item, err := parseXiaoyuzhouPage(showHTML, xiaoyuzhouRef{id: pid})
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "声东击西" || item.Description != "关于节目" || item.Thumbnail != "" || !item.Collection ||
		item.URL != "https://www.xiaoyuzhoufm.com/podcast/"+pid || item.Uploader != (media.Person{Name: "声动活泼", ID: pid}) {
		t.Errorf("item = %+v", item)
	}
	var titles, covers, codecs []string
	for i, e := range item.Entries {
		if e.Index != i+1 || e.Uploader.Name != "声动活泼" || e.Album != "声东击西" {
			t.Errorf("entry %d = %+v", i, e)
		}
		titles = append(titles, e.Title)
		covers = append(covers, e.Thumbnail)
		codecs = append(codecs, e.Formats.Audio[0].Codec)
	}
	// Another show's episode on the page (a recommendation) is not listed.
	if !reflect.DeepEqual(titles, []string{"第一期", "第二期", "第三期"}) {
		t.Errorf("titles = %v", titles)
	}
	if !reflect.DeepEqual(covers, []string{"https://image.xyz.test/show.jpg", "https://image.xyz.test/e2.jpg", "https://image.xyz.test/show.jpg"}) {
		t.Errorf("covers = %v", covers)
	}
	if !reflect.DeepEqual(codecs, []string{"M4A", "M4A", "MP3"}) {
		t.Errorf("codecs = %v", codecs)
	}
	if !item.Published.Equal(time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("published = %v", item.Published)
	}
	if _, err := parseXiaoyuzhouPage(showHTML, xiaoyuzhouRef{id: "nosuchshow1"}); err == nil {
		t.Error("a show page without the show's episodes is an error")
	}
}

func TestXiaoyuzhouOverTheNetwork(t *testing.T) {
	t.Parallel()
	stub := testkit.NewStub()
	stub.On("www.xiaoyuzhoufm.com/episode/"+eid, func(*http.Request) testkit.Response {
		return testkit.Text(episodeHTML(episodeAudio, true))
	})
	x := NewXiaoyuzhou(stub.Client())
	link, _ := x.Match("https://www.xiaoyuzhoufm.com/episode/" + eid + "?s=share")
	item, err := x.Resolve(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "E42 声音的故事" {
		t.Errorf("title = %q", item.Title)
	}
	reqs := stub.Requests("xiaoyuzhoufm.com/episode")
	if len(reqs) != 1 || reqs[0].Header.Get("User-Agent") != httpx.BrowserUserAgent || reqs[0].URL.RawQuery != "" {
		t.Errorf("requests = %v", reqs)
	}
	f, err := x.Formats(context.Background(), item, item.Entries[0])
	if err != nil || f != item.Entries[0].Formats {
		t.Errorf("Formats = %v, %v", f, err)
	}
	if _, err := x.Resolve(context.Background(), "https://example.com/"); !errs.Is(err, errs.Input) {
		t.Errorf("a foreign link: %v", err)
	}
}

func TestXiaoyuzhouNetworkErrors(t *testing.T) {
	t.Parallel()
	stub := testkit.NewStub()
	stub.On("www.xiaoyuzhoufm.com/episode/", func(*http.Request) testkit.Response { return testkit.Status(404) })
	if _, err := NewXiaoyuzhou(stub.Client()).Resolve(context.Background(), "https://www.xiaoyuzhoufm.com/episode/"+eid); err == nil {
		t.Error("a 404 is an error")
	}
}
