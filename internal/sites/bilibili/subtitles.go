package bilibili

import (
	"context"
	"strings"
	"unicode"

	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// found is a subtitle as an endpoint lists it: its language key (ai- prefixed when machine-made) and URL.
type found struct {
	key string
	url string
}

// subtitles finds an entry's subtitles. Which endpoint answers depends on the API and the login: the
// international ones for intl; logged out, only the app's danmaku service still lists them; logged in, the web
// player, then the video info, then the app. The first list without holes wins.
func (e *Extractor) subtitles(ctx context.Context, s *session, r *ref, player jsonv.Value) []media.Subtitle {
	var list []found
	var ok bool
	switch {
	case r.api == Intl:
		if list, ok = e.intlWebSubtitles(ctx, s, r); !ok {
			list, _ = e.intlAppSubtitles(ctx, s, r)
		}
	case s.cookie == "":
		list, _ = e.appSubtitles(ctx, r)
	default:
		if list, ok = listed(player.Path("data", "subtitle"), "subtitles", "lan", "subtitle_url"); !ok {
			if list, ok = e.viewSubtitles(ctx, s, r); !ok {
				list, _ = e.appSubtitles(ctx, r)
			}
		}
	}
	var out []media.Subtitle
	for _, f := range list {
		u := f.url
		if strings.HasPrefix(u, "//") {
			u = "https:" + u
		}
		lang, auto := subtitleLanguage(f.key)
		format := media.BilibiliJSON
		if r.api == Intl && !strings.Contains(u, ".json") {
			format = media.ASS
		}
		out = append(out, media.Subtitle{Lang: lang, Auto: auto, Format: format,
			Source: media.Resource{URL: u, Header: s.mediaHeader(u), Policy: media.Whole}})
	}
	return out
}

// listed reads node[array] as subtitles; ok is false when there is no list or an entry has no URL.
func listed(node jsonv.Value, array, keyField, urlField string) ([]found, bool) {
	items := node.Get(array)
	if !items.IsArray() {
		return nil, false
	}
	var out []found
	for _, it := range items.Array() {
		f := found{key: it.Get(keyField).String(), url: strings.ReplaceAll(it.Get(urlField).String(), `\/`, "/")}
		if f.url == "" {
			return nil, false
		}
		out = append(out, f)
	}
	return out, true
}

func (e *Extractor) viewSubtitles(ctx context.Context, s *session, r *ref) ([]found, bool) {
	j, err := e.getJSON(ctx, s, "https://api.bilibili.com/x/web-interface/view?aid="+r.aid+"&cid="+r.cid)
	if err != nil {
		return nil, false
	}
	return listed(j.Path("data", "subtitle"), "list", "lan", "subtitle_url")
}

// appSubtitles asks the app's danmaku service (DmViewReply.subtitle.subtitles: lan 3, subtitleUrl 5).
func (e *Extractor) appSubtitles(ctx context.Context, r *ref) ([]found, bool) {
	reply, err := e.dmView(ctx, r.aid, r.cid)
	if err != nil {
		return nil, false
	}
	var out []found
	for _, item := range reply.message(3).messages(3) {
		f := found{key: item.string(3), url: item.string(5)}
		if f.url == "" {
			return nil, false
		}
		out = append(out, f)
	}
	return out, true
}

func (e *Extractor) intlWebSubtitles(ctx context.Context, s *session, r *ref) ([]found, bool) {
	host := s.epHost
	if host == defaultHost {
		host = "api.biliintl.com"
	}
	j, err := e.getJSON(ctx, s, "https://"+host+"/intl/gateway/web/v2/subtitle?episode_id="+r.epid)
	if err != nil {
		return nil, false
	}
	return listed(j.Get("data"), "subtitles", "lang_key", "url")
}

func (e *Extractor) intlAppSubtitles(ctx context.Context, s *session, r *ref) ([]found, bool) {
	host := "api.bilibili.tv"
	if s.biliPlus() {
		host = s.host
	}
	api := "https://" + host + "/intl/gateway/v2/ogv/view/app/season?ep_id=" + r.epid + "&platform=android&s_locale=zh_SG"
	if s.token != "" {
		api += "&access_key=" + s.token
	}
	j, err := e.getJSON(ctx, s, api)
	if err != nil {
		return nil, false
	}
	ep := j.Path("result", "modules").At(0).Path("data", "episodes").At(r.position - 1)
	return listed(ep, "subtitles", "key", "url")
}

// subtitleLanguage turns bilibili's key into a BCP 47 tag, and whether it was machine-made: ai-zh → zh, auto;
// zh-hans → zh-Hans; en-us → en-US.
func subtitleLanguage(key string) (string, bool) {
	auto := len(key) > 3 && strings.EqualFold(key[:3], "ai-")
	if auto {
		key = key[3:]
	}
	parts := strings.Split(strings.ReplaceAll(key, "_", "-"), "-")
	for i, p := range parts {
		switch {
		case i == 0:
			parts[i] = strings.ToLower(p)
		case len(p) == 4 && isLetters(p):
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		case len(p) == 2 && isLetters(p):
			parts[i] = strings.ToUpper(p)
		default:
			parts[i] = strings.ToLower(p)
		}
	}
	return strings.Join(parts, "-"), auto
}

func isLetters(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return !unicode.IsLetter(r) }) < 0
}
