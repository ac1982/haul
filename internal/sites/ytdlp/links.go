package ytdlp

import (
	"context"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/httpx"
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
	// expand turns a short link into the site's own before yt-dlp sees it; nil leaves links as they are.
	expand   func(ctx context.Context, client *httpx.Client, link string) (string, error)
	policy   media.RangePolicy
	maxRange int64
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

var weiboProfile = profile{
	info: extract.Info{
		Site:       "weibo",
		Name:       "Weibo",
		Links:      "weibo.com/<uid>/<id>, /tv/show/…, video.weibo.com/…, m.weibo.cn/…, t.cn/…",
		Unit:       "video",
		OwnerLabel: "by",
		Requires:   []extract.Tool{{Name: "yt-dlp", Install: shell.InstallHint("yt-dlp")}},
	},
	match: func(link string) (string, bool) {
		if u, host, path := components(link); host == "t.cn" && len(path) == 1 {
			u.Scheme = "https"
			return u.String(), true
		}
		return weiboLink(link)
	},
	expand: expandWeibo,
	policy: media.Parallel,
}

var (
	weiboPostID  = regexp.MustCompile(`^[A-Za-z0-9]+$`)
	weiboVideoID = regexp.MustCompile(`^\d+:(?:[0-9a-f]{32}|\d{16,})$`)
)

// weiboLink recognises the Weibo links yt-dlp reads: a post (weibo.com/<uid>/<id>, m.weibo.cn/status/<id> or
// /detail/<id>) or a video page (weibo.com/tv/show/<fid>, video.weibo.com/show?fid=<fid>). It returns the link with
// its scheme.
func weiboLink(link string) (string, bool) {
	u, host, path := components(link)
	if u == nil {
		return "", false
	}
	host = strings.TrimPrefix(host, "www.")
	ok := false
	switch host {
	case "weibo.com":
		ok = len(path) == 3 && path[0] == "tv" && path[1] == "show" && weiboVideoID.MatchString(path[2]) ||
			len(path) == 2 && digits.MatchString(path[0]) && weiboPostID.MatchString(path[1])
	case "video.weibo.com":
		ok = len(path) == 1 && path[0] == "show" && weiboVideoID.MatchString(u.Query().Get("fid"))
	case "m.weibo.cn":
		ok = len(path) == 2 && (path[0] == "status" || path[0] == "detail") && weiboPostID.MatchString(path[1])
	}
	if !ok {
		return "", false
	}
	u.Scheme = "https"
	return u.String(), true
}

// expandWeibo follows a t.cn link one hop. Weibo's shortener points at the video page, which then sends anyone
// without a session to a visitor wall that yt-dlp cannot read, so the hop is taken here and the rest left to yt-dlp.
func expandWeibo(ctx context.Context, client *httpx.Client, link string) (string, error) {
	if _, host, _ := components(link); host != "t.cn" {
		return link, nil
	}
	target, err := client.Location(ctx, link, nil)
	if err != nil {
		return "", errs.New("Could not expand %s: %v", link, err)
	}
	expanded, ok := weiboLink(target)
	if !ok {
		return "", errs.NewInput("%s leads to %s, which is not a Weibo video", link, target)
	}
	return expanded, nil
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
