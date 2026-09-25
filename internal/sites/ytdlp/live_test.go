package ytdlp

import (
	"context"
	"io"
	"net/http"
	"os"
	"slices"
	"testing"

	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
)

// Against the real sites. Off by default (slow, and the answers change); run with HAUL_LIVE=1. Needs yt-dlp (and
// deno for YouTube) on PATH.
func requireLive(t *testing.T) {
	t.Helper()
	if os.Getenv("HAUL_LIVE") != "1" {
		t.Skip("set HAUL_LIVE=1 to run against the real sites")
	}
}

func TestLiveYouTube(t *testing.T) {
	requireLive(t)
	yt := NewYouTube(httpx.Default, Options{})
	link, ok := yt.Match("https://youtu.be/DdCEmlAydcw")
	if !ok {
		t.Fatal("no match")
	}
	item, err := yt.Resolve(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "Inside Anthropic's molecular biology lab" || len(item.Entries) != 1 {
		t.Errorf("item = %q, %d entries", item.Title, len(item.Entries))
	}
	f := item.Entries[0].Formats
	if len(f.Video) == 0 || len(f.Audio) == 0 || slices.ContainsFunc(f.Video, func(v media.VideoFormat) bool { return v.HasAudio }) {
		t.Errorf("want DASH video and audio: %d video, %d audio", len(f.Video), len(f.Audio))
	}
	if !slices.ContainsFunc(f.Subtitles, func(s media.Subtitle) bool { return s.Lang == "en" && !s.Auto }) {
		t.Errorf("no en subtitle: %+v", f.Subtitles)
	}
	// googlevideo serves closed ranges with the headers the format carries.
	a := f.Audio[0]
	if a.Source.Policy != media.Sequential || a.Source.Size == 0 {
		t.Errorf("audio source = %+v", a.Source)
	}
	h := a.Source.Header.Clone()
	h.Set("Range", "bytes=0-1023")
	resp, err := httpx.Default.Do(context.Background(), http.MethodGet, a.Source.URL, h, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusPartialContent || len(body) != 1024 {
		t.Errorf("range answer = %d, %d bytes", resp.StatusCode, len(body))
	}
}

func TestLiveX(t *testing.T) {
	requireLive(t)
	x := NewX(httpx.Default, Options{})
	link, ok := x.Match("https://x.com/historyinmemes/status/1790637656616943991")
	if !ok {
		t.Fatal("no match")
	}
	item, err := x.Resolve(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	if len(item.Entries) != 1 || item.Entries[0].Uploader.Name != "Historic Vids" {
		t.Fatalf("item = %+v", item)
	}
	f := item.Entries[0].Formats
	if len(f.Video) == 0 || !f.Video[0].HasAudio {
		t.Errorf("want whole files with audio: %+v", f.Video)
	}
}
