package podcast

import (
	"net/http"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/jsonv"
)

func TestAudioCodec(t *testing.T) {
	t.Parallel()
	cases := []struct{ mediaType, url, want string }{
		{"audio/mpeg", "https://a.test/x", "MP3"},
		{"", "https://a.test/x.m4a?t=1", "M4A"},
		{"audio/mp4", "https://a.test/x.mp3", "M4A"},
		{"", "https://a.test/redirect", "MP3"},
		// A format we do not know must not be muxed as MP3: it keeps its name and goes into an M4A.
		{"audio/ogg", "https://a.test/x", "OGG"},
		{"", "https://a.test/x.opus", "OPUS"},
		{"audio/x-m4b", "https://a.test/x", "M4A"},
		{"mp3", "https://a.test/redirect.mp3/x", "MP3"},
		{"audio/aac", "", "M4A"},
	}
	for _, c := range cases {
		if got := audioCodec(c.mediaType, c.url); got != c.want {
			t.Errorf("audioCodec(%q, %q) = %q, want %q", c.mediaType, c.url, got, c.want)
		}
	}
}

func TestDates(t *testing.T) {
	t.Parallel()
	want := time.Unix(1716588000, 0)
	for _, text := range []string{"2024-05-24T22:00:00.000Z", "2024-05-24T22:00:00Z", "2024-05-25T06:00:00+08:00"} {
		if got := parseISODate(text); !got.Equal(want) {
			t.Errorf("parseISODate(%q) = %v", text, got)
		}
	}
	for _, text := range []string{
		"Fri, 24 May 2024 22:00:00 GMT",
		"Fri, 24 May 2024 22:00:00 +0000",
		" Fri, 24 May 2024 15:00:00 PDT ",
		"Fri, 24 May 2024 18:00:00 EDT",
		"24 May 2024 22:00:00 +0000",
		"Fri, 24 May 2024 22:00 +0000",
		"Sat, 25 May 2024 07:00:00 +0900",
	} {
		if got := parseRFC822Date(text); !got.Equal(want) {
			t.Errorf("parseRFC822Date(%q) = %v", text, got)
		}
	}
	for _, bad := range []string{"", "yesterday", "2024-05-24"} {
		if !parseISODate(bad).IsZero() || !parseRFC822Date(bad).IsZero() {
			t.Errorf("%q parsed", bad)
		}
	}
	durations := map[string]time.Duration{"3600": time.Hour, "59:30": 3570 * time.Second, "1:02:03": 3723 * time.Second,
		" 12:05 ": 725 * time.Second, "": 0, "90.5": 90 * time.Second}
	for text, want := range durations {
		if got := parseDuration(text); got != want {
			t.Errorf("parseDuration(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestArtworkIsUpscaled(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		`{"artworkUrl600": "https://x.test/a/600x600bb.png"}`:           "https://x.test/a/1400x1400bb.jpg",
		`{"artworkUrl160": "https://x.test/a/160x160bb.webp"}`:          "https://x.test/a/1400x1400bb.jpg",
		`{"artworkUrl600": "https://x.test/a/cover.jpg"}`:               "https://x.test/a/cover.jpg",
		`{"artworkUrl600": "", "artworkUrl160": "https://x/1x1bb.jpg"}`: "https://x/1400x1400bb.jpg",
		`{}`: "",
	}
	for text, want := range cases {
		v, _ := jsonv.ParseString(text)
		if got := artwork(v); got != want {
			t.Errorf("artwork(%s) = %q, want %q", text, got, want)
		}
	}
}

func TestMetaAndAttributes(t *testing.T) {
	t.Parallel()
	html := `<meta name="description" content="d"><meta content='https://a.test/x?a=1&amp;b=2' property='og:audio'>` +
		`<meta property="og:title" content="">`
	if got := meta("og:audio", html); got != "https://a.test/x?a=1&b=2" {
		t.Errorf("og:audio = %q", got)
	}
	if meta("description", html) != "d" || meta("og:title", html) != "" || meta("og:image", html) != "" {
		t.Error("meta")
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()
	if got := normalize("‎  Bits &  Pieces‏\n"); got != "bits & pieces" {
		t.Errorf("normalize = %q", got)
	}
}

func TestWalkVisitsInKeyOrder(t *testing.T) {
	t.Parallel()
	v, _ := jsonv.ParseString(`{"b": {"id": 2, "c": [{"id": 3}]}, "a": {"id": 1}, "id": 0}`)
	var ids []string
	walk(v, func(n jsonv.Value) { ids = append(ids, n.Get("id").String()) })
	if got := len(ids); got != 4 || ids[0] != "0" || ids[1] != "1" || ids[2] != "2" || ids[3] != "3" {
		t.Errorf("visit order = %v", ids)
	}
}

func TestBitrateFromStatedSize(t *testing.T) {
	t.Parallel()
	e := episode{mediaURL: "https://a.test/x.mp3", size: 43392000, duration: 2712 * time.Second}
	h := http.Header{"User-Agent": {"UA"}}
	f := formats(e, h)
	if a := f.Audio[0]; a.Bitrate != 125 || a.Size != 0 || a.Source.Size != 0 || a.Source.Header.Get("User-Agent") != "UA" {
		t.Errorf("audio = %+v", a)
	}
	// Unknown size or duration: no bitrate.
	e.duration = 0
	if a := formats(e, h).Audio[0]; a.Bitrate != 0 {
		t.Errorf("bitrate = %d", a.Bitrate)
	}
}
