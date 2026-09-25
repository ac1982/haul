package bilibili

import (
	"context"
	"regexp"
	"strings"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/jsonv"
)

var (
	avInLink      = regexp.MustCompile(`av(\d+)`)
	bvInLink      = regexp.MustCompile(`[Bb][Vv]1(\w+)`)
	epInLink      = regexp.MustCompile(`/ep(\d+)`)
	ssInLink      = regexp.MustCompile(`/ss(\d+)`)
	mdInLink      = regexp.MustCompile(`md(\d+)`)
	spaceInLink   = regexp.MustCompile(`space\.bilibili\.com/(\d+)`)
	intlPlayLink  = regexp.MustCompile(`\.bilibili\.tv/\w+/play/\d+/(\d+)`)
	mediaPageLink = regexp.MustCompile(`bangumi/media/md(\d+)`)
	initialState  = regexp.MustCompile(`window\.__INITIAL_STATE__=([\s\S]*?);\(function\(\)`)
)

func unsupported(link string) error { return errs.NewInput("Unsupported bilibili link: %s", link) }

// resolveID turns a link or bare id into what it points at, and returns the link as finally read (a b23.tv link
// expanded), whose ?p= may pick a page.
func (e *Extractor) resolveID(ctx context.Context, s *session, input string) (mediaID, string, error) {
	link, ok := matchLink(input)
	if !ok {
		return mediaID{}, "", unsupported(input)
	}
	id, err := e.resolveLink(ctx, s, &link)
	if err != nil {
		return mediaID{}, "", err
	}
	if id, err = e.fixVideoID(ctx, s, id); err != nil {
		return mediaID{}, "", err
	}
	console.Debugf("Resolved to %s", id)
	return id, link, nil
}

func (e *Extractor) resolveLink(ctx context.Context, s *session, link *string) (mediaID, error) {
	input := *link
	if strings.HasPrefix(input, "http") {
		if strings.Contains(input, "b23.tv") {
			expanded, err := e.finalURL(ctx, s, input)
			if err != nil {
				return mediaID{}, err
			}
			if expanded == input {
				return mediaID{}, errs.New("b23.tv link redirects to itself")
			}
			*link = expanded
		}
		return e.resolveURL(ctx, s, *link)
	}
	lower := strings.ToLower(input)
	switch {
	case strings.HasPrefix(lower, "bv"):
		aid, err := bvDecode(input[3:])
		if err != nil {
			return mediaID{}, err
		}
		return mediaID{kind: video, id: itoa(aid)}, nil
	case strings.HasPrefix(lower, "av"):
		return mediaID{kind: video, id: input[2:]}, nil
	case strings.HasPrefix(input, "cheese/"):
		return e.resolveCheese(ctx, s, input)
	case strings.HasPrefix(lower, "ep"):
		return mediaID{kind: episode, id: input[2:]}, nil
	case strings.HasPrefix(lower, "ss"):
		ep, err := e.bangumiFirstEp(ctx, s, input[2:])
		return mediaID{kind: episode, id: ep}, err
	case strings.HasPrefix(lower, "md"):
		ep, err := e.bangumiNewEp(ctx, s, input[2:])
		return mediaID{kind: episode, id: ep}, err
	}
	return mediaID{}, unsupported(input)
}

func (e *Extractor) resolveCheese(ctx context.Context, s *session, input string) (mediaID, error) {
	if m := epInLink.FindStringSubmatch(input); m != nil {
		return mediaID{kind: cheese, id: m[1]}, nil
	}
	if m := ssInLink.FindStringSubmatch(input); m != nil {
		ep, err := e.cheeseFirstEp(ctx, s, m[1])
		return mediaID{kind: cheese, id: ep}, err
	}
	return mediaID{}, errs.NewInput("Unrecognised bilibili course link: %s", input)
}

func (e *Extractor) resolveURL(ctx context.Context, s *session, input string) (mediaID, error) {
	if strings.Contains(input, "video/av") {
		if m := avInLink.FindStringSubmatch(input); m != nil {
			return mediaID{kind: video, id: m[1]}, nil
		}
	}
	if strings.Contains(strings.ToLower(input), "video/bv") {
		if m := bvInLink.FindStringSubmatch(input); m != nil {
			aid, err := bvDecode(m[1])
			if err != nil {
				return mediaID{}, err
			}
			return mediaID{kind: video, id: itoa(aid)}, nil
		}
	}
	if strings.Contains(input, "/cheese/") {
		return e.resolveCheese(ctx, s, input)
	}
	if m := epInLink.FindStringSubmatch(input); m != nil {
		return mediaID{kind: episode, id: m[1]}, nil
	}
	if m := ssInLink.FindStringSubmatch(input); m != nil {
		ep, err := e.bangumiFirstEp(ctx, s, m[1])
		return mediaID{kind: episode, id: ep}, err
	}
	if strings.Contains(input, "/medialist/") && strings.Contains(input, "business_id=") {
		biz := format.QueryValue("business_id", input)
		if strings.Contains(input, "business=space_collection") {
			return mediaID{kind: collection, id: biz}, nil
		}
		if strings.Contains(input, "business=space_series") {
			return mediaID{kind: series, id: biz}, nil
		}
	}
	if strings.Contains(input, "/channel/collectiondetail?sid=") {
		return mediaID{kind: collection, id: format.QueryValue("sid", input)}, nil
	}
	if strings.Contains(input, "/channel/seriesdetail?sid=") {
		return mediaID{kind: series, id: format.QueryValue("sid", input)}, nil
	}
	if strings.Contains(input, "/space.bilibili.com/") {
		if strings.Contains(input, "/lists/") {
			// New-style space lists: …/lists/<sid>?type=season|series
			path, _, _ := strings.Cut(input, "?")
			path, _, _ = strings.Cut(path, "#")
			sid := path[strings.LastIndexByte(strings.TrimSuffix(path, "/"), '/')+1:]
			sid = strings.TrimSuffix(sid, "/")
			if strings.EqualFold(format.QueryValue("type", input), "series") {
				return mediaID{kind: series, id: sid}, nil
			}
			return mediaID{kind: collection, id: sid}, nil
		}
		if m := spaceInLink.FindStringSubmatch(input); m != nil {
			if strings.Contains(input, "/favlist") {
				return mediaID{kind: favorites, id: format.QueryValue("fid", input), mid: m[1]}, nil
			}
			return mediaID{kind: space, id: m[1]}, nil
		}
	}
	if strings.Contains(input, "ep_id=") {
		return mediaID{kind: episode, id: format.QueryValue("ep_id", input)}, nil
	}
	if m := intlPlayLink.FindStringSubmatch(input); m != nil {
		return mediaID{kind: episode, id: m[1]}, nil
	}
	if m := mediaPageLink.FindStringSubmatch(input); m != nil {
		ep, err := e.bangumiNewEp(ctx, s, m[1])
		return mediaID{kind: episode, id: ep}, err
	}
	// Last resort: the page's initial state may list episodes.
	html, err := e.getText(ctx, s, input)
	if err != nil {
		return mediaID{}, err
	}
	m := initialState.FindStringSubmatch(html)
	if m == nil {
		return mediaID{}, unsupported(input)
	}
	state, err := jsonv.ParseString(m[1])
	if err != nil {
		return mediaID{}, err
	}
	ep, ok := state.Get("epList").At(0).Get("id").Int64()
	if !ok {
		return mediaID{}, errs.New("No episode found on the page")
	}
	return mediaID{kind: episode, id: itoa(ep)}, nil
}

// fixVideoID: a plain av id may be licensed content; the video page then redirects to its episode.
func (e *Extractor) fixVideoID(ctx context.Context, s *session, id mediaID) (mediaID, error) {
	if id.kind != video {
		return id, nil
	}
	if _, ok := parseInt(id.id); !ok {
		return id, nil
	}
	location, err := e.finalURL(ctx, s, "https://www.bilibili.com/video/av"+id.id+"/")
	if err != nil {
		return mediaID{}, err
	}
	if m := epInLink.FindStringSubmatch(location); m != nil {
		return mediaID{kind: episode, id: m[1]}, nil
	}
	return id, nil
}

func (e *Extractor) cheeseFirstEp(ctx context.Context, s *session, seasonID string) (string, error) {
	j, err := e.getJSON(ctx, s, "https://api.bilibili.com/pugv/view/web/season?season_id="+seasonID)
	if err != nil {
		return "", err
	}
	if ep, ok := j.Path("data", "episodes").At(0).Get("id").Int64(); ok {
		return itoa(ep), nil
	}
	return "", errs.New("Course ss%s has no episodes", seasonID)
}

func (e *Extractor) bangumiFirstEp(ctx context.Context, s *session, seasonID string) (string, error) {
	j, err := e.getJSON(ctx, s, "https://"+s.epHost+"/pgc/view/web/season?season_id="+seasonID)
	if err != nil {
		return "", err
	}
	if ep, ok := j.Path("result", "episodes").At(0).Get("id").Int64(); ok {
		return itoa(ep), nil
	}
	return "", errs.New("Season ss%s has no episodes", seasonID)
}

func (e *Extractor) bangumiNewEp(ctx context.Context, s *session, mediaIDText string) (string, error) {
	j, err := e.getJSON(ctx, s, "https://api.bilibili.com/pgc/review/user?media_id="+mediaIDText)
	if err != nil {
		return "", err
	}
	if ep, ok := j.Path("result", "media", "new_ep", "id").Int64(); ok {
		return itoa(ep), nil
	}
	return "", errs.New("md%s has no episodes", mediaIDText)
}
