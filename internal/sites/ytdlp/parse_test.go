package ytdlp

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

const youtubeSample = `{
  "id": "DdCEmlAydcw", "title": "Lab tour", "description": "desc", "channel": "Anthropic", "channel_id": "UC1",
  "timestamp": 1790186492, "duration": 75, "language": "en-US",
  "webpage_url": "https://www.youtube.com/watch?v=DdCEmlAydcw",
  "chapters": [{"title": "Intro", "start_time": 0.0, "end_time": 12.5}],
  "thumbnails": [
    {"url": "https://i.ytimg.com/vi/DdCEmlAydcw/hqdefault.jpg?sqp=x", "preference": -7, "width": 336},
    {"url": "https://i.ytimg.com/vi/DdCEmlAydcw/maxresdefault.jpg", "preference": -1},
    {"url": "https://i.ytimg.com/vi_webp/DdCEmlAydcw/maxresdefault.webp", "preference": 0}
  ],
  "formats": [
    {"format_id": "sb0", "protocol": "mhtml", "url": "u", "vcodec": "none", "acodec": "none"},
    {"format_id": "620", "protocol": "m3u8_native", "url": "u", "vcodec": "vp09.00.50.08", "acodec": "none"},
    {"format_id": "140-0", "protocol": "https", "url": "a-dub", "vcodec": "none", "acodec": "mp4a.40.2", "abr": 129.5,
     "filesize": 1222496, "language_preference": -1},
    {"format_id": "140-19", "protocol": "https", "url": "a-orig", "vcodec": "none", "acodec": "mp4a.40.2", "abr": 129.6,
     "filesize": 1220994, "language_preference": 10, "http_headers": {"User-Agent": "UA"}},
    {"format_id": "251-drc", "protocol": "https", "url": "a-drc", "vcodec": "none", "acodec": "opus", "abr": 131, "language_preference": 10},
    {"format_id": "251", "protocol": "https", "url": "a-drm", "vcodec": "none", "acodec": "opus", "abr": 131, "language_preference": 10, "has_drm": true},
    {"format_id": "18", "protocol": "https", "url": "muxed", "vcodec": "avc1.42001E", "acodec": "mp4a.40.2"},
    {"format_id": "271", "protocol": "https", "url": "v", "vcodec": "vp9", "acodec": "none", "format_note": "1080p",
     "width": 2048, "height": 1240, "fps": 24, "vbr": 5826.7, "filesize": 54868477}
  ],
  "subtitles": {"en": [{"ext": "vtt", "url": "s-vtt"}, {"ext": "json3", "url": "s-en"}], "live_chat": [{"ext": "json3", "url": "chat"}]},
  "automatic_captions": {
    "en-orig": [{"ext": "json3", "url": "asr-en"}], "fr-FR-orig": [{"ext": "json3", "url": "asr-fr"}],
    "de": [{"ext": "json3", "url": "tr-de"}]
  }
}`

const xPost = `{
  "_type": "playlist", "id": "1600649710662213632", "title": "Jocelyn Laidlaw - How her diagnosis changed...",
  "description": "full text", "uploader": "Jocelyn Laidlaw", "timestamp": 1670459604,
  "webpage_url": "https://twitter.com/CTVJLaidlaw/status/1600649710662213632",
  "entries": [
    {"id": "1600649511827038209", "title": "Jocelyn Laidlaw - How her diagnosis changed...", "duration": 113.49,
     "uploader": "Jocelyn Laidlaw", "uploader_id": "JocelynVLaidlaw",
     "thumbnails": [{"url": "https://pbs.twimg.com/a.jpg?name=small", "width": 680}, {"url": "https://pbs.twimg.com/a.jpg?name=orig", "width": 886}],
     "formats": [
       {"format_id": "hls-audio-128000-Audio", "protocol": "m3u8_native", "url": "hls", "vcodec": "none", "acodec": null},
       {"format_id": "http-632", "protocol": "https", "url": "https://video.twimg.com/ext_tw_video/1/pu/vid/320x568/a.mp4",
        "width": 320, "height": 568, "tbr": 632, "filesize_approx": 8965710},
       {"format_id": "http-2176", "protocol": "https", "url": "https://video.twimg.com/amplify_video/1/vid/avc1/720x1280/b.mp4",
        "width": 720, "height": 1280, "tbr": 2176, "http_headers": {"User-Agent": "UA"}}
     ]},
    {"id": "1600649511827013632", "duration": 102.2,
     "formats": [{"format_id": "http-950", "protocol": "https", "url": "c.mp4", "width": 480, "height": 852, "tbr": 950}]}
  ]
}`

func mustParse(t *testing.T, text string, p *profile) *media.Item {
	t.Helper()
	root, err := jsonv.ParseString(text)
	if err != nil {
		t.Fatal(err)
	}
	item, err := parse(root, p)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func TestParseYouTube(t *testing.T) {
	t.Parallel()
	item := mustParse(t, youtubeSample, &youtubeProfile)
	if item.Site != "youtube" || item.ID != "DdCEmlAydcw" || item.Title != "Lab tour" || item.Description != "desc" ||
		item.URL != "https://www.youtube.com/watch?v=DdCEmlAydcw" || item.Collection {
		t.Errorf("item = %+v", item)
	}
	if !item.Published.Equal(time.Unix(1790186492, 0)) {
		t.Errorf("published = %v", item.Published)
	}
	if item.Uploader != (media.Person{Name: "Anthropic", ID: "UC1"}) {
		t.Errorf("uploader = %+v", item.Uploader)
	}
	if item.Thumbnail != "https://i.ytimg.com/vi/DdCEmlAydcw/maxresdefault.jpg" {
		t.Errorf("thumbnail = %q", item.Thumbnail)
	}
	if len(item.Entries) != 1 {
		t.Fatalf("entries = %d", len(item.Entries))
	}
	e := item.Entries[0]
	if e.Index != 1 || e.ID != "DdCEmlAydcw" || e.Title != "Lab tour" || e.Duration != 75*time.Second ||
		e.Uploader.Name != "Anthropic" || e.URL != "https://www.youtube.com/watch?v=DdCEmlAydcw" || e.Description != "desc" {
		t.Errorf("entry = %+v", e)
	}
	if want := []media.Chapter{{Title: "Intro", Start: 0, End: 12500 * time.Millisecond}}; !reflect.DeepEqual(e.Chapters, want) {
		t.Errorf("chapters = %+v", e.Chapters)
	}

	f := e.Formats
	if f == nil || f.AudioOnly || f.Raw == "" {
		t.Fatalf("formats = %+v", f)
	}
	// DASH is there, so the muxed format 18 is left out.
	if len(f.Video) != 1 {
		t.Fatalf("video = %+v", f.Video)
	}
	v := f.Video[0]
	if v.ID != "271" || v.Rank != 1240024 || v.Quality != "1080p" || v.Width != 2048 || v.Height != 1240 || v.FPS != 24 ||
		v.Codec != "VP9" || v.Bitrate != 5827 || v.Size != 54868477 || v.HasAudio {
		t.Errorf("video = %+v", v)
	}
	ua := http.Header{"User-Agent": {"UA"}}
	wantSource := media.Resource{URL: "v", Header: ua, Size: 54868477, Policy: media.Sequential, MaxRange: 10 << 20}
	if !reflect.DeepEqual(v.Source, wantSource) {
		t.Errorf("video source = %+v", v.Source)
	}
	// Only the original-language audio: no dub, DRC or DRM variant.
	if len(f.Audio) != 1 {
		t.Fatalf("audio = %+v", f.Audio)
	}
	a := f.Audio[0]
	if a.ID != "140-19" || a.Codec != "M4A" || a.Bitrate != 130 || a.Size != 1220994 {
		t.Errorf("audio = %+v", a)
	}
	if !reflect.DeepEqual(a.Source, media.Resource{URL: "a-orig", Header: ua, Size: 1220994, Policy: media.Sequential, MaxRange: 10 << 20}) {
		t.Errorf("audio source = %+v", a.Source)
	}
	// The uploaded track plus the spoken-language auto track; no chat, translations or dub transcripts.
	want := []media.Subtitle{
		{Lang: "en", Format: media.YouTubeJSON3, Source: media.Resource{URL: "s-en", Header: ua}},
		{Lang: "en", Auto: true, Format: media.YouTubeJSON3, Source: media.Resource{URL: "asr-en", Header: ua}},
	}
	if !reflect.DeepEqual(f.Subtitles, want) {
		t.Errorf("subtitles = %+v", f.Subtitles)
	}
}

func TestParseXMultiVideoPost(t *testing.T) {
	t.Parallel()
	item := mustParse(t, xPost, &xProfile)
	if item.Site != "x" || item.Title != "Jocelyn Laidlaw - How her diagnosis changed" || item.Thumbnail != "" ||
		item.Collection || item.URL != "https://twitter.com/CTVJLaidlaw/status/1600649710662213632" {
		t.Errorf("item = %+v", item)
	}
	if !item.Published.Equal(time.Unix(1670459604, 0)) {
		t.Errorf("published = %v", item.Published)
	}
	if item.Uploader != (media.Person{Name: "Jocelyn Laidlaw", ID: "JocelynVLaidlaw"}) {
		t.Errorf("uploader = %+v", item.Uploader)
	}
	if len(item.Entries) != 2 {
		t.Fatalf("entries = %d", len(item.Entries))
	}
	for i, e := range item.Entries {
		if e.Index != i+1 || e.Title != []string{"Video 1", "Video 2"}[i] {
			t.Errorf("entry %d = %d %q", i, e.Index, e.Title)
		}
		// Fields the entries leave out come from the post.
		if !e.Published.Equal(time.Unix(1670459604, 0)) || e.URL != item.URL || e.Description != "full text" {
			t.Errorf("entry %d = %+v", i, e)
		}
	}
	first := item.Entries[0]
	if first.ID != "1600649511827038209" || item.Entries[1].ID != "1600649511827013632" {
		t.Errorf("ids = %q %q", first.ID, item.Entries[1].ID)
	}
	if first.Thumbnail != "https://pbs.twimg.com/a.jpg?name=orig" || first.Duration != 113490*time.Millisecond ||
		first.Uploader != (media.Person{Name: "Jocelyn Laidlaw", ID: "JocelynVLaidlaw"}) {
		t.Errorf("first = %+v", first)
	}
	// Whole files with the audio inside, no HLS; X's approximate sizes are not trusted.
	f := first.Formats
	if len(f.Audio) != 0 || len(f.Video) != 2 {
		t.Fatalf("formats = %+v", f)
	}
	for i, v := range f.Video {
		want := []struct{ id, quality, codec string }{{"http-632", "320p", "MP4"}, {"http-2176", "720p", "AVC"}}[i]
		if v.ID != want.id || v.Quality != want.quality || v.Codec != want.codec || !v.HasAudio || v.Size != 0 {
			t.Errorf("video %d = %+v", i, v)
		}
		if v.Source.Policy != media.Parallel || v.Source.MaxRange != 0 || v.Source.Size != 0 || v.Source.Header.Get("User-Agent") != "UA" {
			t.Errorf("video %d source = %+v", i, v.Source)
		}
	}
	if f.Video[0].Rank != 320000 || f.Video[1].Rank != 720000 {
		t.Errorf("ranks = %d %d", f.Video[0].Rank, f.Video[1].Rank)
	}
	if v := item.Entries[1].Formats.Video; len(v) != 1 || v[0].Bitrate != 950 || v[0].Codec != "MP4" {
		t.Errorf("second video = %+v", v)
	}
}

func TestOnlyXTitlesLoseTheEllipsis(t *testing.T) {
	t.Parallel()
	const post = `{"id": "DdCEmlAydcw", "title": "Wait for it...", "formats": [{"format_id": "18", "protocol": "https",
		"url": "u", "vcodec": "avc1", "acodec": "mp4a.40.2", "width": 640, "height": 360}]}`
	if got := mustParse(t, post, &youtubeProfile); got.Title != "Wait for it..." || got.Entries[0].Title != "Wait for it..." {
		t.Errorf("YouTube title = %q", got.Title)
	}
	if got := mustParse(t, post, &xProfile); got.Title != "Wait for it" || got.Entries[0].Title != "Wait for it" {
		t.Errorf("X title = %q", got.Title)
	}
}

func TestProgressiveOnlyWithoutDASH(t *testing.T) {
	t.Parallel()
	// Only a muxed format: it is taken, with its audio.
	const muxed = `{"id": "DdCEmlAydcw", "upload_date": "20260910", "formats": [{"format_id": "18", "protocol": "https",
		"url": "u", "vcodec": "avc1.42001E", "acodec": "mp4a.40.2", "width": 640, "height": 360, "fps": 29.97, "tbr": 500.4}]}`
	item := mustParse(t, muxed, &youtubeProfile)
	v := item.Entries[0].Formats.Video
	if len(v) != 1 || !v[0].HasAudio || v[0].Rank != 360030 || v[0].Quality != "360p" || v[0].Codec != "AVC" || v[0].Bitrate != 500 {
		t.Errorf("video = %+v", v)
	}
	// No timestamp: the upload day, midnight UTC.
	if want := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC); !item.Published.Equal(want) {
		t.Errorf("published = %v", item.Published)
	}
	// Audio alone is DASH enough: the muxed format is not used.
	const audioOnly = `{"id": "DdCEmlAydcw", "formats": [
		{"format_id": "18", "protocol": "https", "url": "u", "vcodec": "avc1", "acodec": "mp4a.40.2", "width": 640, "height": 360},
		{"format_id": "251", "protocol": "https", "url": "a", "vcodec": "none", "acodec": "opus", "tbr": 120.6}]}`
	f := mustParse(t, audioOnly, &youtubeProfile).Entries[0].Formats
	if len(f.Video) != 0 || len(f.Audio) != 1 || f.Audio[0].Codec != "OPUS" || f.Audio[0].Bitrate != 121 {
		t.Errorf("formats = %+v", f)
	}
}

func TestNoDownloadableStreamsIsAnError(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		`{"id": "DdCEmlAydcw", "formats": []}`,
		`{"id": "DdCEmlAydcw", "formats": [{"format_id": "hls", "protocol": "m3u8_native", "url": "u", "vcodec": "avc1", "width": 1}]}`,
		`{"_type": "playlist", "id": "1", "entries": []}`,
	} {
		root, _ := jsonv.ParseString(text)
		if _, err := parse(root, &youtubeProfile); !errs.Is(err, errs.Failed) {
			t.Errorf("parse(%s) = %v", text, err)
		}
	}
}

func TestSubtitlesWithoutSpokenLanguage(t *testing.T) {
	t.Parallel()
	// With no spoken language every original-language auto track is offered.
	n, _ := jsonv.ParseString(`{"automatic_captions": {"fr-FR-orig": [{"ext": "json3", "url": "fr"}], "en-orig": [{"ext": "vtt", "url": "vtt"}]}}`)
	subs := subtitles(n, nil)
	if len(subs) != 1 || subs[0].Lang != "fr-FR" || !subs[0].Auto || subs[0].Source.URL != "fr" {
		t.Errorf("subtitles = %+v", subs)
	}
}

func TestCodecs(t *testing.T) {
	t.Parallel()
	video := map[string]string{"avc1.64001F": "AVC", "hev1.1.6": "HEVC", "hvc1": "HEVC", "av01.0.08M.08": "AV1",
		"vp9": "VP9", "vp09.00.50.08": "VP9", "theora": "THEORA"}
	for in, want := range video {
		if got := videoCodec(in); got != want {
			t.Errorf("videoCodec(%q) = %q", in, got)
		}
	}
	urls := map[string]string{"https://v.test/vid/avc1/1.mp4": "AVC", "https://v.test/vid/hevc/1.mp4": "HEVC", "c.mp4": "MP4"}
	for in, want := range urls {
		if got := urlCodec(in); got != want {
			t.Errorf("urlCodec(%q) = %q", in, got)
		}
	}
	audio := map[string]string{"mp4a.40.2": "M4A", "opus": "OPUS", "ec-3": "E-AC-3", "ac-3": "AC-3", "flac": "FLAC"}
	for in, want := range audio {
		if got := audioCodec(in); got != want {
			t.Errorf("audioCodec(%q) = %q", in, got)
		}
	}
}

func TestCoverFallsBackToThumbnail(t *testing.T) {
	t.Parallel()
	n, _ := jsonv.ParseString(`{"thumbnail": "https://i.test/t.webp", "thumbnails": [{"url": "https://i.test/a.webp"}]}`)
	if got := cover(n); got != "https://i.test/t.webp" {
		t.Errorf("cover = %q", got)
	}
}
