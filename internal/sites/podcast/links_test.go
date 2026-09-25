package podcast

import "testing"

const (
	eid = "6650a1b2c3d4e5f6a7b8c9d0"
	pid = "6021f949a789fca4eff4492c"
)

func TestXiaoyuzhouLinks(t *testing.T) {
	t.Parallel()
	ep, ok := parseXiaoyuzhou("https://www.xiaoyuzhoufm.com/episode/" + eid + "?s=eyJ1")
	if !ok || ep.id != eid || !ep.episode {
		t.Errorf("episode = %+v, %v", ep, ok)
	}
	show, ok := parseXiaoyuzhou("xiaoyuzhoufm.com/podcast/" + pid)
	if !ok || show.id != pid || show.episode {
		t.Errorf("show = %+v, %v", show, ok)
	}
	for _, link := range []string{
		"https://www.xiaoyuzhoufm.com/",
		"https://www.xiaoyuzhoufm.com/about",
		"https://www.xiaoyuzhoufm.com/episode/short",
		"https://notxiaoyuzhoufm.com/episode/" + eid,
	} {
		if _, ok := parseXiaoyuzhou(link); ok {
			t.Errorf("%s matched", link)
		}
	}
}

func TestAppleLinks(t *testing.T) {
	t.Parallel()
	ep, ok := parseApple("https://podcasts.apple.com/cn/podcast/%E5%A3%B0%E4%B8%9C/id1487143507?i=1000655000001&l=en")
	if !ok || ep != (appleRef{show: "1487143507", episode: "1000655000001", country: "cn"}) {
		t.Errorf("episode = %+v", ep)
	}
	show, ok := parseApple("podcasts.apple.com/US/podcast/the-daily/id1200361736")
	if !ok || show != (appleRef{show: "1200361736", country: "us"}) {
		t.Errorf("show = %+v", show)
	}
	// Without a country the US store is asked.
	bare, ok := parseApple("https://podcasts.apple.com/podcast/id1200361736?i=abc")
	if !ok || bare != (appleRef{show: "1200361736", country: "us"}) {
		t.Errorf("bare = %+v", bare)
	}
	for _, link := range []string{
		"https://podcasts.apple.com/us/browse",
		"https://podcasts.apple.com/us/podcast/the-daily",
		"https://music.apple.com/us/album/x/id123",
	} {
		if _, ok := parseApple(link); ok {
			t.Errorf("%s matched", link)
		}
	}
}

func TestMatchNormalizes(t *testing.T) {
	t.Parallel()
	x, a := NewXiaoyuzhou(nil), NewApple(nil)
	cases := []struct {
		ex         *Extractor
		link, want string
		ok         bool
	}{
		{x, "https://www.xiaoyuzhoufm.com/episode/" + eid + "?s=abc", "https://www.xiaoyuzhoufm.com/episode/" + eid, true},
		{x, "xiaoyuzhoufm.com/podcast/" + pid, "https://www.xiaoyuzhoufm.com/podcast/" + pid, true},
		{x, "https://podcasts.apple.com/us/podcast/id1", "", false},
		{a, "https://podcasts.apple.com/cn/podcast/slug/id1487143507?i=1000655000001", "https://podcasts.apple.com/cn/podcast/id1487143507?i=1000655000001", true},
		{a, "https://podcasts.apple.com/podcast/the-daily/id1200361736", "https://podcasts.apple.com/us/podcast/id1200361736", true},
		{a, "https://www.xiaoyuzhoufm.com/episode/" + eid, "", false},
	}
	for _, c := range cases {
		got, ok := c.ex.Match(c.link)
		if got != c.want || ok != c.ok {
			t.Errorf("%s Match(%q) = %q, %v; want %q, %v", c.ex.Info().Name, c.link, got, ok, c.want, c.ok)
		}
	}
}

func TestInfo(t *testing.T) {
	t.Parallel()
	x, a := NewXiaoyuzhou(nil).Info(), NewApple(nil).Info()
	if x.Site != "xiaoyuzhou" || x.Name != "Xiaoyuzhou" || x.Unit != "episode" || x.OwnerLabel != "host" || len(x.Requires) != 0 {
		t.Errorf("Xiaoyuzhou info = %+v", x)
	}
	if a.Site != "applePodcasts" || a.Name != "Apple Podcasts" || a.Unit != "episode" || a.OwnerLabel != "host" || len(a.Requires) != 0 {
		t.Errorf("Apple info = %+v", a)
	}
}
