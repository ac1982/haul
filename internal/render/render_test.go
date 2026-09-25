package render

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/media"
)

var plain = console.PlainStyle

func video(rank int, q, codec string, w, h int, fps float64, kbps int64) media.VideoFormat {
	return media.VideoFormat{ID: q + codec, Rank: rank, Quality: q, Width: w, Height: h, FPS: fps, Codec: codec, Bitrate: kbps,
		Source: media.Resource{URL: "https://cdn/" + q + "-" + codec}}
}

func sampleFormats() ([]media.VideoFormat, []media.AudioFormat) {
	return []media.VideoFormat{
		video(120, "4K", "HEVC", 3840, 2160, 60, 4493), video(80, "1080P", "HEVC", 1920, 1080, 30, 1000),
		video(120, "4K", "AVC", 3840, 2160, 60, 11450), video(80, "1080P", "AV1", 1920, 1080, 30, 1000),
	}, []media.AudioFormat{
		{ID: "a", Codec: "M4A", Bitrate: 172, Source: media.Resource{URL: "https://cdn/a"}},
		{ID: "b", Codec: "M4A", Bitrate: 66, Source: media.Resource{URL: "https://cdn/b"}},
	}
}

func TestFormatting(t *testing.T) {
	cases := [][2]string{
		{FPS(60), "60fps"}, {FPS(29.412), "29.4fps"}, {FPS(0), ""}, {Resolution(1920, 1080), "1920×1080"}, {Resolution(0, 0), ""},
		{ETA(7), "7s"}, {ETA(65), "1m05s"}, {ETA(3700), "1h01m"}, {ETA(math.Inf(1)), "--"},
	}
	for _, c := range cases {
		if c[0] != c[1] {
			t.Errorf("got %q, want %q", c[0], c[1])
		}
	}
}

func TestStreamTableAlignsAndMarksTheChoice(t *testing.T) {
	v, a := sampleFormats()
	lines := Streams(v, a, false, 0, 1, 100*time.Second, false, false, plain)
	var rows []string
	for _, l := range lines {
		if strings.Contains(l, "×") {
			rows = append(rows, l)
		}
	}
	if len(rows) != 4 {
		t.Fatalf("rows: %q", lines)
	}
	for _, r := range rows[1:] {
		if console.DisplayWidth(r) != console.DisplayWidth(rows[0]) {
			t.Errorf("rows differ in width:\n%s\n%s", rows[0], r)
		}
	}
	if !strings.HasPrefix(rows[0], "  ▶ 0  4K") || !strings.HasPrefix(rows[1], "    1  1080P") {
		t.Errorf("rows: %q", rows)
	}
	if !strings.Contains(rows[0], "3840×2160") || !strings.Contains(rows[0], "60fps") || !strings.Contains(rows[0], "4493 kbps") {
		t.Errorf("row 0: %q", rows[0])
	}
	var audio []string
	for _, l := range lines {
		if strings.Contains(l, "M4A") {
			audio = append(audio, l)
		}
	}
	if !strings.HasPrefix(audio[1], "  ▶ 1  M4A") || strings.Contains(strings.Join(lines, ""), "\x1b") {
		t.Errorf("audio: %q", audio)
	}
}

func TestStreamTableCollapsesOtherCodecs(t *testing.T) {
	v, a := sampleFormats()
	lines := Streams(v, a, false, 0, 0, 100*time.Second, true, false, plain)
	n := 0
	for _, l := range lines {
		if strings.Contains(l, "×") {
			n++
		}
	}
	if n != 2 || !strings.Contains(strings.Join(lines, "\n"), "2 more AV1 / AVC streams") {
		t.Errorf("collapsed: %q", lines)
	}
	urls := Streams(v, a, false, 0, 0, 0, false, true, plain)
	if !strings.Contains(strings.Join(urls, "\n"), "    https://cdn/4K-HEVC") {
		t.Errorf("urls: %q", urls)
	}
	withAudioInside := Streams(v[:1], nil, true, 0, -1, 0, false, false, plain)
	if !strings.Contains(strings.Join(withAudioInside, "\n"), "Audio  inside the video file") {
		t.Errorf("audio inside: %q", withAudioInside)
	}
}

func TestProgressLines(t *testing.T) {
	line := ProgressLine("video", 5<<20, 10<<20, 1<<20, '|', 100, plain)
	for _, want := range []string{" 50%", "5.0 MB / 10.0 MB", "1.0 MB/s", "5s left", "████████████░░░░░░░░░░░░"} {
		if !strings.Contains(line, want) {
			t.Errorf("missing %q in %q", want, line)
		}
	}
	unknown := ProgressLine("cover", 2048, 0, 0, '⠋', 80, plain)
	if !strings.Contains(unknown, "⠋") || !strings.Contains(unknown, "2.0 KB") || strings.Contains(unknown, "%") {
		t.Errorf("unknown: %q", unknown)
	}
	narrow := ProgressLine("video", 1, 100, 1, '|', 40, plain)
	if strings.Count(narrow, "█")+strings.Count(narrow, "░") != 8 {
		t.Errorf("narrow terminals get the minimum bar: %q", narrow)
	}
	done := ProgressDone("video", 10<<20, 2, plain)
	for _, want := range []string{"✓", "10.0 MB", "5.0 MB/s", "2s"} {
		if !strings.Contains(done, want) {
			t.Errorf("missing %q in %q", want, done)
		}
	}
}

func TestSummaryAndHeader(t *testing.T) {
	v := video(120, "4K", "HEVC", 3840, 2160, 60, 4493)
	a := media.AudioFormat{Codec: "M4A", Bitrate: 172}
	lines := Summary("/tmp/out.mp4", 300<<20, &v, &a, 24, plain)
	if lines[0] != "✓ Done  "+console.PrettyPath("/tmp/out.mp4") || !strings.Contains(lines[1], "300.0 MB") || !strings.Contains(lines[1], "4K HEVC 60fps + M4A 172 kbps") ||
		!strings.Contains(lines[1], "24s") {
		t.Errorf("summary: %q", lines)
	}
	yes := true
	item := &media.Item{Title: "标题", Published: time.Unix(1_700_000_000, 0), LoggedIn: &yes,
		Entries: []*media.Entry{{Index: 1, Title: "p1", Duration: 518 * time.Second, Uploader: media.Person{Name: "某UP"}}}}
	site := extract.Info{Unit: "page", OwnerLabel: "uploader"}
	h := Header(item, site, plain)
	if h[0] != "  标题" || !strings.Contains(h[1], "uploader 某UP") || !strings.Contains(h[1], "00:08:38") || strings.Contains(h[1], "pages") ||
		!strings.Contains(h[1], "logged in") {
		t.Errorf("header: %q", h)
	}
	no := false
	item.LoggedIn = &no
	if h := Header(item, site, plain); !strings.Contains(h[1], "logged out") {
		t.Errorf("logged out: %q", h)
	}
	item.Entries = nil
	for i := 1; i <= 8; i++ {
		item.Entries = append(item.Entries, &media.Entry{Index: i, Title: "第" + string(rune('0'+i)) + "话", Duration: time.Minute})
	}
	list := EntryList(item.Entries, false, plain)
	if len(list) != 6 || !strings.HasPrefix(list[0], "  P1  第1话") || !strings.Contains(list[5], "3 more") || len(EntryList(item.Entries, true, plain)) != 8 {
		t.Errorf("list: %q", list)
	}
	if h := Header(item, site, plain); !strings.Contains(h[1], "00:08:00 total") || !strings.Contains(h[1], "8 pages") {
		t.Errorf("list header: %q", h)
	}
	if got := EntryHeader(3, 12, "第三话", plain); got != "  [03/12] 第三话" {
		t.Errorf("entry header: %q", got)
	}
}
