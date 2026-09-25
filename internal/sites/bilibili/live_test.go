package bilibili

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/httpx"
)

// TestLive reads a real video from bilibili, logged out. Opt in with HAUL_LIVE=1.
func TestLive(t *testing.T) {
	if os.Getenv("HAUL_LIVE") != "1" {
		t.Skip("set HAUL_LIVE=1 to talk to bilibili")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	e := New(httpx.Default, DefaultOptions())
	item, err := e.Resolve(ctx, "https://www.bilibili.com/video/BV1qt4y1X7TW")
	if err != nil {
		t.Fatal(err)
	}
	if item.ID != "BV1qt4y1X7TW" || len(item.Entries) != 1 || item.LoggedIn == nil {
		t.Fatalf("%+v", item)
	}
	f, err := e.Formats(ctx, item, item.Entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Video) == 0 || len(f.Audio) == 0 || f.Video[0].Source.URL == "" || f.Video[0].Width == 0 {
		t.Fatalf("%d video, %d audio: %+v", len(f.Video), len(f.Audio), f.Video)
	}
	for _, v := range f.Video {
		t.Logf("video %s %s %s %dx%d %.3f fps %d kbps", v.ID, v.Quality, v.Codec, v.Width, v.Height, v.FPS, v.Bitrate)
	}
	for _, a := range f.Audio {
		t.Logf("audio %s %s %d kbps", a.ID, a.Codec, a.Bitrate)
	}
	t.Logf("%d subtitles, %d chapters", len(f.Subtitles), len(f.Chapters))
	files, err := f.Sidecars[0].Write(ctx, filepath.Join(t.TempDir(), "live"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("danmaku: %v", files)

	// The chosen stream answers byte ranges on the mirror, over plain HTTP, with the headers it carries.
	src := f.Video[0].Source
	h := src.Header.Clone()
	h.Set("Range", "bytes=0-1023")
	resp, err := httpx.Default.Do(ctx, "GET", src.URL, h, nil)
	if err != nil {
		t.Fatalf("%s: %v", src.URL, err)
	}
	resp.Body.Close()
	if resp.StatusCode != 206 {
		t.Fatalf("%s: HTTP %d", src.URL, resp.StatusCode)
	}

	// The same video through the APP (gRPC) and TV APIs.
	for _, api := range []API{App, TV} {
		o := DefaultOptions()
		o.API = api
		x := New(httpx.Default, o)
		item, err := x.Resolve(ctx, "BV1qt4y1X7TW")
		if err != nil {
			t.Fatalf("%s: %v", api, err)
		}
		f, err := x.Formats(ctx, item, item.Entries[0])
		if err != nil {
			t.Fatalf("%s: %v", api, err)
		}
		if len(f.Video) == 0 || len(f.Audio) == 0 {
			t.Fatalf("%s: %d video, %d audio", api, len(f.Video), len(f.Audio))
		}
		t.Logf("%s: %d video (best %s %s), %d audio", api, len(f.Video), f.Video[0].Quality, f.Video[0].Codec, len(f.Audio))
	}
}
