package podcast

import (
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/ac1982/haul/internal/format"
)

// xiaoyuzhouRef is an episode (xiaoyuzhoufm.com/episode/<id>) or a show (/podcast/<id>).
type xiaoyuzhouRef struct {
	id      string
	episode bool
}

func (r xiaoyuzhouRef) url() string {
	if r.id == "" {
		return ""
	}
	kind := "podcast"
	if r.episode {
		kind = "episode"
	}
	return "https://www.xiaoyuzhoufm.com/" + kind + "/" + r.id
}

var xiaoyuzhouID = regexp.MustCompile(`^[A-Za-z0-9]{8,}$`)

func parseXiaoyuzhou(link string) (xiaoyuzhouRef, bool) {
	_, host, path := components(link)
	if host != "xiaoyuzhoufm.com" && !strings.HasSuffix(host, ".xiaoyuzhoufm.com") ||
		len(path) < 2 || (path[0] != "episode" && path[0] != "podcast") || !xiaoyuzhouID.MatchString(path[1]) {
		return xiaoyuzhouRef{}, false
	}
	return xiaoyuzhouRef{id: path[1], episode: path[0] == "episode"}, true
}

// appleRef is a show, or one of its episodes when the link has ?i=<episode>. The country picks the store the
// lookup searches; an episode can exist in one and not another.
type appleRef struct {
	show    string
	episode string
	country string
}

func (r appleRef) url() string {
	if r.show == "" {
		return ""
	}
	u := "https://podcasts.apple.com/" + r.country + "/podcast/id" + r.show
	if r.episode != "" {
		u += "?i=" + r.episode
	}
	return u
}

var (
	appleShowID = regexp.MustCompile(`^id\d+$`)
	number      = regexp.MustCompile(`^\d+$`)
	country     = regexp.MustCompile(`^[A-Za-z]{2}$`)
)

// parseApple reads podcasts.apple.com/<cc>/podcast/<slug>/id<show>[?i=<episode>]; without a country it is "us".
func parseApple(link string) (appleRef, bool) {
	u, host, path := components(link)
	if host != "podcasts.apple.com" && host != "itunes.apple.com" || !slices.Contains(path, "podcast") {
		return appleRef{}, false
	}
	ref := appleRef{country: "us"}
	for j := len(path) - 1; j >= 0; j-- {
		if appleShowID.MatchString(path[j]) {
			ref.show = path[j][2:]
			break
		}
	}
	if ref.show == "" {
		return appleRef{}, false
	}
	if ep := format.QueryValue("i", u.String()); number.MatchString(ep) {
		ref.episode = ep
	}
	if country.MatchString(path[0]) {
		ref.country = strings.ToLower(path[0])
	}
	return ref, true
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
