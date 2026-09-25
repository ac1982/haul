package ytdlp

import (
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/shell"
)

// googlevideoRange is the largest range googlevideo serves: it answers only closed byte ranges, and plain or
// open-ended requests get 403.
const googlevideoRange = 10 * 1024 * 1024

// profile is what differs between the sites yt-dlp reads for haul.
type profile struct {
	info  extract.Info
	match func(link string) (string, bool)
	// trimEllipsis: X titles are the post text, which yt-dlp shortens with a trailing "..." a file name is better
	// off without. Other sites' titles are the uploader's own and stay as they are.
	trimEllipsis bool
	// needsDeno: yt-dlp solves YouTube's player challenges in deno; without it most formats are missing.
	needsDeno bool
	policy    media.RangePolicy
	maxRange  int64
}

var youtubeProfile = profile{
	info: extract.Info{
		Site:       "youtube",
		Name:       "YouTube",
		Links:      "youtube.com/watch?v=…, youtu.be/…, /shorts/…, /embed/…, /live/…",
		Unit:       "video",
		OwnerLabel: "channel",
		Requires:   []extract.Tool{{Name: "yt-dlp", Install: shell.InstallHint("yt-dlp deno")}},
	},
	match: func(link string) (string, bool) {
		id := youtubeID(link)
		if id == "" {
			return "", false
		}
		return "https://www.youtube.com/watch?v=" + id, true
	},
	needsDeno: true,
	policy:    media.Sequential,
	maxRange:  googlevideoRange,
}

var xProfile = profile{
	info: extract.Info{
		Site:       "x",
		Name:       "X",
		Links:      "x.com/<user>/status/<id>, twitter.com/…; /video/<n> picks one",
		Unit:       "video",
		OwnerLabel: "by",
		Requires:   []extract.Tool{{Name: "yt-dlp", Install: shell.InstallHint("yt-dlp")}},
	},
	match: func(link string) (string, bool) {
		if tweetID(link) == "" {
			return "", false
		}
		// Kept as given: `/video/2` picks one video of a multi-video post.
		link = strings.TrimSpace(link)
		if !strings.HasPrefix(link, "http") {
			link = "https://" + link
		}
		return link, true
	},
	trimEllipsis: true,
	policy:       media.Parallel,
}

var youtubeIDShape = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// youtubeID is the 11-character video id of a YouTube video link, or "" for anything else (playlists, channels).
func youtubeID(link string) string {
	u, host, path := components(link)
	if u == nil {
		return ""
	}
	var id string
	switch {
	case host == "youtu.be":
		if len(path) > 0 {
			id = path[0]
		}
	case host == "youtube.com" || strings.HasSuffix(host, ".youtube.com") || strings.HasSuffix(host, "youtube-nocookie.com"):
		if len(path) > 0 && path[0] == "watch" {
			id = format.QueryValue("v", u.String())
		} else if len(path) >= 2 && slices.Contains([]string{"shorts", "embed", "live", "v", "e"}, path[0]) {
			id = path[1]
		}
	}
	if !youtubeIDShape.MatchString(id) {
		return ""
	}
	return id
}

var digits = regexp.MustCompile(`^\d+$`)

// tweetID is the post id of an X (Twitter) status link: x.com/<user>/status/<id>, twitter.com/i/web/status/<id>…
func tweetID(link string) string {
	u, host, path := components(link)
	if u == nil {
		return ""
	}
	host = strings.TrimPrefix(strings.TrimPrefix(host, "www."), "mobile.")
	if host != "x.com" && host != "twitter.com" {
		return ""
	}
	i := slices.Index(path, "status")
	if i < 0 || i+1 >= len(path) || !digits.MatchString(path[i+1]) {
		return ""
	}
	return path[i+1]
}

// components parses a link given with or without its scheme: the URL, its lower-cased host and its path segments.
func components(link string) (*url.URL, string, []string) {
	text := strings.TrimSpace(link)
	if !strings.HasPrefix(text, "http") {
		text = "https://" + text
	}
	u, err := url.Parse(text)
	if err != nil || u.Hostname() == "" {
		return nil, "", nil
	}
	var path []string
	for _, p := range strings.Split(u.Path, "/") {
		if p != "" {
			path = append(path, p)
		}
	}
	return u, strings.ToLower(u.Hostname()), path
}
