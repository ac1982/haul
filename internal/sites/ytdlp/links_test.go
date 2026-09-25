package ytdlp

import (
	"testing"

	"github.com/ac1982/haul/internal/shell"
)

func TestYouTubeIDs(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"https://youtu.be/DdCEmlAydcw":                           "DdCEmlAydcw",
		"https://youtu.be/DdCEmlAydcw?si=abc&t=10":               "DdCEmlAydcw",
		"https://www.youtube.com/watch?v=DdCEmlAydcw&list=PL1":   "DdCEmlAydcw",
		"m.youtube.com/watch?feature=share&v=-abc_DEF123":        "-abc_DEF123",
		"https://www.youtube.com/shorts/DdCEmlAydcw":             "DdCEmlAydcw",
		"https://www.youtube.com/live/DdCEmlAydcw?feature=share": "DdCEmlAydcw",
		"https://www.youtube-nocookie.com/embed/DdCEmlAydcw":     "DdCEmlAydcw",
		"  https://music.youtube.com/watch?v=DdCEmlAydcw  ":      "DdCEmlAydcw",
		"https://www.youtube.com/playlist?list=PL1":              "",
		"https://www.youtube.com/watch?v=short":                  "",
		"https://www.youtube.com/@Anthropic":                     "",
		"https://notyoutube.com/watch?v=DdCEmlAydcw":             "",
		"https://www.bilibili.com/video/BV1qt4y1X7TW":            "",
		"BV1qt4y1X7TW": "",
		"":             "",
	}
	for link, want := range cases {
		if got := youtubeID(link); got != want {
			t.Errorf("youtubeID(%q) = %q, want %q", link, got, want)
		}
	}
}

func TestTweetIDs(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"https://x.com/historyinmemes/status/1790637656616943991":            "1790637656616943991",
		"https://twitter.com/CTVJLaidlaw/status/1600649710662213632/video/2": "1600649710662213632",
		"mobile.twitter.com/i/web/status/910031516746514432?s=20":            "910031516746514432",
		"https://www.x.com/a/status/12":                                      "12",
		"https://x.com/historyinmemes":                                       "",
		"https://x.com/a/status/abc":                                         "",
		"https://x.com/a/status":                                             "",
		"https://notx.com/a/status/1":                                        "",
		"https://www.youtube.com/watch?v=DdCEmlAydcw":                        "",
	}
	for link, want := range cases {
		if got := tweetID(link); got != want {
			t.Errorf("tweetID(%q) = %q, want %q", link, got, want)
		}
	}
}

func TestMatchNormalizes(t *testing.T) {
	t.Parallel()
	yt, x := NewYouTube(nil, Options{}), NewX(nil, Options{})
	cases := []struct {
		ex         *Extractor
		link, want string
		ok         bool
	}{
		{yt, "https://youtu.be/DdCEmlAydcw?si=x", "https://www.youtube.com/watch?v=DdCEmlAydcw", true},
		{yt, "youtube.com/shorts/DdCEmlAydcw", "https://www.youtube.com/watch?v=DdCEmlAydcw", true},
		{yt, "https://x.com/a/status/12", "", false},
		// Kept as given, so /video/2 still picks one video of the post.
		{x, "x.com/a/status/12/video/2", "https://x.com/a/status/12/video/2", true},
		{x, " https://twitter.com/a/status/12 ", "https://twitter.com/a/status/12", true},
		{x, "https://youtu.be/DdCEmlAydcw", "", false},
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
	yt, x := NewYouTube(nil, Options{}).Info(), NewX(nil, Options{}).Info()
	if yt.Site != "youtube" || yt.Name != "YouTube" || yt.Unit != "video" || yt.OwnerLabel != "channel" ||
		len(yt.Requires) != 1 || yt.Requires[0].Name != "yt-dlp" || yt.Requires[0].Install != shell.InstallHint("yt-dlp deno") {
		t.Errorf("YouTube info = %+v", yt)
	}
	if x.Site != "x" || x.Name != "X" || x.Unit != "video" || x.OwnerLabel != "by" ||
		len(x.Requires) != 1 || x.Requires[0].Install != shell.InstallHint("yt-dlp") {
		t.Errorf("X info = %+v", x)
	}
}
