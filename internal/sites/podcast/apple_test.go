package podcast

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

const lookup = `{"resultCount":4,"results":[
 {"wrapperType":"track","kind":"podcast","collectionId":1200361736,"trackId":1200361736,"artistName":"The New York Times",
  "collectionName":"The Daily","feedUrl":"https://feeds.test/daily","artworkUrl600":"https://is1-ssl.mzstatic.test/image/thumb/show/600x600bb.jpg"},
 {"wrapperType":"podcastEpisode","kind":"podcast-episode","trackId":1000671234567,"trackName":"Newer","collectionId":1200361736,
  "collectionName":"The Daily","releaseDate":"2024-09-20T09:45:00Z","trackTimeMillis":1800000,
  "episodeUrl":"https://podcast.test/redirect.mp3/new.mp3?dest-id=1","episodeFileExtension":"mp3","episodeContentType":"audio",
  "description":"About the newer one","artworkUrl600":"https://is1-ssl.mzstatic.test/image/thumb/ep/600x600bb.jpg",
  "trackViewUrl":"https://podcasts.apple.com/us/podcast/newer/id1200361736?i=1000671234567&uo=4"},
 {"wrapperType":"podcastEpisode","kind":"podcast-episode","trackId":1000671000001,"trackName":"Older","collectionId":1200361736,
  "collectionName":"The Daily","releaseDate":"2024-09-19T09:45:00Z","trackTimeMillis":1500000,
  "episodeUrl":"https://podcast.test/old.m4a","episodeFileExtension":"m4a","episodeContentType":"audio"},
 {"wrapperType":"podcastEpisode","kind":"podcast-episode","trackId":1000671000002,"trackName":"On camera","collectionId":1200361736,
  "collectionName":"The Daily","releaseDate":"2024-09-18T09:45:00Z","trackTimeMillis":600000,
  "episodeUrl":"https://podcast.test/video.mp4","episodeFileExtension":"mp4","episodeContentType":"video"}
]}`

const rss = `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"><channel>
<title>The Daily</title>
<item><title>The Daily</title><enclosure url="https://podcast.test/trailer.mp3" length="1" type="audio/mpeg"/></item>
<item><title>Intro</title><enclosure url="https://podcast.test/intro.mp3" length="1" type="audio/mpeg"/></item>
<item>
  <title><![CDATA[Bits &amp; Pieces: The Long One]]></title><itunes:title>Long One</itunes:title>
  <guid isPermaLink="false">abc-123</guid><pubDate>Fri, 20 Sep 2024 09:45:00 +0000</pubDate>
  <itunes:duration>1:02:03</itunes:duration><itunes:author>Host</itunes:author>
  <itunes:image href="https://podcast.test/long.jpg"/>
  <description><![CDATA[<p>What the <b>long</b> one is about.</p>]]></description>
  <enclosure length="123" type="audio/x-m4a" url="https://podcast.test/long.m4a?a=1&amp;b=2"/>
</item>
</channel></rss>`

func mustLookup(t *testing.T) jsonv.Value {
	t.Helper()
	v, err := jsonv.ParseString(lookup)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

var dailyRef = appleRef{show: "1200361736", country: "us"}

func TestAppleEpisode(t *testing.T) {
	t.Parallel()
	ref := dailyRef
	ref.episode = "1000671234567"
	item, err := parseAppleLookup(mustLookup(t), ref)
	if err != nil || item == nil {
		t.Fatalf("item = %v, %v", item, err)
	}
	if item.Site != "applePodcasts" || item.Title != "Newer" || item.Description != "About the newer one" || item.Collection ||
		item.Thumbnail != "https://is1-ssl.mzstatic.test/image/thumb/ep/1400x1400bb.jpg" {
		t.Errorf("item = %+v", item)
	}
	if !item.Published.Equal(time.Unix(1726825500, 0)) {
		t.Errorf("published = %v", item.Published)
	}
	e := item.Entries[0]
	if e.ID != "1000671234567" || e.Duration != 30*time.Minute || e.Uploader != (media.Person{Name: "The New York Times", ID: "1200361736"}) ||
		e.Album != "The Daily" || e.URL != "https://podcasts.apple.com/us/podcast/newer/id1200361736?i=1000671234567&uo=4" {
		t.Errorf("entry = %+v", e)
	}
	a := e.Formats.Audio[0]
	if !e.Formats.AudioOnly || a.Codec != "MP3" || a.Source.URL != "https://podcast.test/redirect.mp3/new.mp3?dest-id=1" ||
		a.Source.Header.Get("User-Agent") != httpx.BrowserUserAgent || a.Source.Policy != media.Parallel || a.Size != 0 {
		t.Errorf("audio = %+v", a)
	}
}

func TestAppleShow(t *testing.T) {
	t.Parallel()
	lk := mustLookup(t)
	item, err := parseAppleLookup(lk, dailyRef)
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "The Daily" || item.Thumbnail != "" || !item.Collection || item.URL != "https://podcasts.apple.com/us/podcast/id1200361736" ||
		item.Uploader != (media.Person{Name: "The New York Times", ID: "1200361736"}) {
		t.Errorf("item = %+v", item)
	}
	var titles []string
	for _, e := range item.Entries {
		titles = append(titles, e.Title)
	}
	if !reflect.DeepEqual(titles, []string{"On camera", "Older", "Newer"}) {
		t.Errorf("titles = %v", titles)
	}
	// A video episode is one file with its audio inside.
	video := item.Entries[0].Formats
	if video.AudioOnly || len(video.Audio) != 0 || len(video.Video) != 1 {
		t.Fatalf("video formats = %+v", video)
	}
	if v := video.Video[0]; !v.HasAudio || v.Codec != "MP4" || v.Quality != "original" || v.Source.URL != "https://podcast.test/video.mp4" {
		t.Errorf("video = %+v", v)
	}
	// An episode without artwork takes the show's.
	if older := item.Entries[1]; older.Formats.Audio[0].Codec != "M4A" || older.Thumbnail != "https://is1-ssl.mzstatic.test/image/thumb/show/1400x1400bb.jpg" {
		t.Errorf("older = %+v", older)
	}
}

func TestAppleLookupEdges(t *testing.T) {
	t.Parallel()
	lk := mustLookup(t)
	// An episode older than the lookup returns is looked up elsewhere.
	ref := dailyRef
	ref.episode = "42"
	if item, err := parseAppleLookup(lk, ref); item != nil || err != nil {
		t.Errorf("unknown episode = %v, %v", item, err)
	}
	empty, _ := jsonv.ParseString(`{"resultCount":0,"results":[]}`)
	_, err := parseAppleLookup(empty, appleRef{show: "1", country: "us"})
	if !errs.Is(err, errs.Failed) || err.Error() != "Apple Podcasts has no show id1 (the country in the link may be wrong)" {
		t.Errorf("unknown show: %v", err)
	}
	showOnly, _ := jsonv.ParseString(`{"results":[{"wrapperType":"track","kind":"podcast","collectionName":"The Daily"}]}`)
	if _, err := parseAppleLookup(showOnly, dailyRef); err == nil || err.Error() != "Apple Podcasts returned no episodes for this show" {
		t.Errorf("show without episodes: %v", err)
	}
	// 199 episodes plus the show is the whole list; 200 episodes may not be.
	results := lk.Get("results").Array()
	list := func(n int) []jsonv.Value {
		out := []jsonv.Value{results[0]}
		for range n {
			out = append(out, results[1])
		}
		return out
	}
	if appleListIsCut(list(199)) || !appleListIsCut(list(200)) {
		t.Error("appleListIsCut")
	}
}

func TestOlderEpisodeFoundInTheFeed(t *testing.T) {
	t.Parallel()
	// Apple titles its pages with an invisible mark and the show's name around the episode title.
	page := "<html><head><title>‎Bits &amp; Pieces: The Long One - The Daily - Apple Podcasts</title></head></html>"
	titles := appleEpisodeTitles(page, "The Daily")
	if !reflect.DeepEqual(titles, []string{"bits & pieces: the long one"}) {
		t.Errorf("titles = %q", titles)
	}
	e, ok := rssEpisode(rss, titles)
	if !ok {
		t.Fatal("not found")
	}
	want := episode{
		id: "abc-123", title: "Bits & Pieces: The Long One", mediaURL: "https://podcast.test/long.m4a?a=1&b=2", mediaType: "audio/x-m4a",
		duration: 3723 * time.Second, published: time.Unix(1726825500, 0).UTC(), cover: "https://podcast.test/long.jpg",
		description: "What the long one is about.", author: "Host",
	}
	if !reflect.DeepEqual(e, want) {
		t.Errorf("episode = %+v", e)
	}
}

// An episode the feed no longer has is an error, never the trailer or another item whose title fits inside.
func TestAMissingEpisodeIsNotReplacedByANearMatch(t *testing.T) {
	t.Parallel()
	page := "<html><head><title>An Old Episode - The Daily - Apple Podcasts</title></head></html>"
	if _, ok := rssEpisode(rss, appleEpisodeTitles(page, "The Daily")); ok {
		t.Error("an old episode matched")
	}
	if _, ok := rssEpisode(rss, []string{"Intro to everything"}); ok {
		t.Error("a longer title matched")
	}
	if e, ok := rssEpisode(rss, []string{"  INTRO "}); !ok || e.mediaURL != "https://podcast.test/intro.mp3" || e.id != e.mediaURL {
		t.Errorf("intro = %+v, %v", e, ok)
	}
}

func appleStub(t *testing.T) *testkit.Stub {
	t.Helper()
	stub := testkit.NewStub()
	stub.On("itunes.apple.com/lookup?id=1200361736", func(*http.Request) testkit.Response { return testkit.JSON(lookup) })
	return stub
}

func TestAppleOverTheNetwork(t *testing.T) {
	t.Parallel()
	stub := appleStub(t)
	a := NewApple(stub.Client())
	link, _ := a.Match("https://podcasts.apple.com/us/podcast/the-daily/id1200361736?i=1000671000001")
	item, err := a.Resolve(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "Older" {
		t.Errorf("title = %q", item.Title)
	}
	reqs := stub.Requests("itunes.apple.com/lookup")
	if len(reqs) != 1 {
		t.Fatalf("requests = %d", len(reqs))
	}
	q := reqs[0].URL.Query()
	if q.Get("country") != "us" || q.Get("limit") != "200" || q.Get("entity") != "podcastEpisode" || q.Get("media") != "podcast" {
		t.Errorf("query = %v", q)
	}
	// The country in the link picks the store.
	stub.On("itunes.apple.com/lookup?id=1487143507", func(*http.Request) testkit.Response { return testkit.JSON(`{"results":[]}`) })
	if _, err := a.Resolve(context.Background(), "https://podcasts.apple.com/cn/podcast/id1487143507"); err == nil {
		t.Error("an unknown show is an error")
	}
	if r := stub.Requests("id=1487143507"); len(r) != 1 || r[0].URL.Query().Get("country") != "cn" {
		t.Errorf("cn requests = %v", r)
	}
}

func TestOlderAppleEpisodeComesFromTheFeed(t *testing.T) {
	t.Parallel()
	stub := appleStub(t)
	link := "https://podcasts.apple.com/us/podcast/id1200361736?i=1000000000009"
	stub.On("podcasts.apple.com/us/podcast/id1200361736?i=1000000000009", func(*http.Request) testkit.Response {
		return testkit.Text(`<html><head><meta property="og:title" content="Bits &amp; Pieces: The Long One"></head></html>`)
	})
	stub.On("feeds.test/daily", func(*http.Request) testkit.Response { return testkit.Text(rss) })
	item, err := NewApple(stub.Client()).Resolve(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	e := item.Entries[0]
	if item.Title != "Bits & Pieces: The Long One" || e.ID != "1000000000009" || e.Album != "The Daily" ||
		e.Uploader != (media.Person{Name: "Host", ID: "1200361736"}) || e.URL != link || e.Thumbnail != "https://podcast.test/long.jpg" {
		t.Errorf("entry = %+v", e)
	}
	if a := e.Formats.Audio[0]; a.Source.URL != "https://podcast.test/long.m4a?a=1&b=2" || a.Codec != "M4A" {
		t.Errorf("audio = %+v", a)
	}
	for _, r := range append(stub.Requests("feeds.test"), stub.Requests("podcasts.apple.com")...) {
		if r.Header.Get("User-Agent") != httpx.BrowserUserAgent {
			t.Errorf("%s without the browser user agent", r.URL)
		}
	}
}

func TestOlderAppleEpisodeErrors(t *testing.T) {
	t.Parallel()
	const link = "https://podcasts.apple.com/us/podcast/id1200361736?i=1000000000009"
	cases := []struct {
		name, page, feed string
		feedStatus       int
		want             string
	}{
		{"not in feed", `<title>Gone - The Daily - Apple Podcasts</title>`, rss, 0,
			`Episode "gone" is not in the show's RSS feed (the feed may keep only newer episodes)`},
		{"no title", `<html></html>`, rss, 0, "No episode title on the Apple Podcasts page"},
		{"feed down", `<title>Gone</title>`, "", 500, "HTTP 500: https://feeds.test/daily"},
	}
	for _, c := range cases {
		stub := appleStub(t)
		stub.On("podcasts.apple.com/us/podcast/id1200361736?i=", func(*http.Request) testkit.Response { return testkit.Text(c.page) })
		stub.On("feeds.test/daily", func(*http.Request) testkit.Response {
			if c.feedStatus != 0 {
				return testkit.Status(c.feedStatus)
			}
			return testkit.Text(c.feed)
		})
		_, err := NewApple(stub.Client()).Resolve(context.Background(), link)
		if err == nil || err.Error() != c.want {
			t.Errorf("%s: err = %v", c.name, err)
		}
	}
	// A show without a feed cannot be searched.
	stub := testkit.NewStub()
	stub.On("itunes.apple.com/lookup", func(*http.Request) testkit.Response {
		return testkit.JSON(strings.Replace(lookup, `"feedUrl":"https://feeds.test/daily",`, "", 1))
	})
	if _, err := NewApple(stub.Client()).Resolve(context.Background(), link); err == nil || err.Error() != "Apple Podcasts gave no RSS feed for this show" {
		t.Errorf("no feed: %v", err)
	}
}
