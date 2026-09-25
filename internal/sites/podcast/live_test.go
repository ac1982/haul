package podcast

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ac1982/haul/internal/httpx"
)

// Against the real site. Off by default (slow, and the answers change); run with HAUL_LIVE=1.
func TestLiveApplePodcasts(t *testing.T) {
	if os.Getenv("HAUL_LIVE") != "1" {
		t.Skip("set HAUL_LIVE=1 to run against the real sites")
	}
	a := NewApple(httpx.Default)
	link, ok := a.Match("https://podcasts.apple.com/us/podcast/the-daily/id1200361736")
	if !ok {
		t.Fatal("no match")
	}
	item, err := a.Resolve(context.Background(), link)
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "The Daily" || len(item.Entries) <= 10 {
		t.Fatalf("item = %q, %d entries", item.Title, len(item.Entries))
	}
	newest := item.Entries[len(item.Entries)-1]
	if newest.Album != "The Daily" || len(newest.Formats.Audio) != 1 || !strings.HasPrefix(newest.Formats.Audio[0].Source.URL, "http") {
		t.Errorf("newest = %+v", newest)
	}
	if !newest.Published.After(item.Entries[0].Published) {
		t.Errorf("not oldest first: %v … %v", item.Entries[0].Published, newest.Published)
	}
}
