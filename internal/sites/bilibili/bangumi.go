package bilibili

import (
	"context"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// trailerBadge marks trailers in an episode list; they are left out.
const trailerBadge = "预告"

// fetchBangumi reads the season an episode belongs to (pgc/view/web/season).
func (e *Extractor) fetchBangumi(ctx context.Context, s *session, epID string) (*info, error) {
	j, err := e.getJSON(ctx, s, "https://"+s.epHost+"/pgc/view/web/season?ep_id="+epID)
	if err != nil {
		return nil, err
	}
	result, err := field(j, "result")
	if err != nil {
		return nil, err
	}
	episodes, err := field(result, "episodes")
	if err != nil {
		return nil, err
	}
	title := result.Get("title").String()
	// The episode may live in a section (extras, PVs) rather than the main list.
	marker := "/ep" + epID
	if len(episodes.Array()) == 0 || !strings.Contains(episodes.String(), marker) {
		for _, section := range result.Get("section").Array() {
			if strings.Contains(section.String(), marker) {
				title += "[" + section.Get("title").String() + "]"
				episodes = section.Get("episodes")
				break
			}
		}
	}
	in := &info{
		title:     strings.TrimSpace(title),
		desc:      trim(result.Get("evaluate")),
		cover:     result.Get("cover").String(),
		published: pubTime(result.Path("publish", "pub_time")),
		season:    true,
		finished:  result.Path("publish", "is_finish").IntOr(0) == 1,
		seasonID:  result.Get("season_id").String(),
	}
	addEpisodes(in, episodes.Array(), epID)
	return in, nil
}

// addEpisodes adds a season's episodes, trailers left out, and focuses the one the link named.
func addEpisodes(in *info, episodes []jsonv.Value, epID string) {
	for _, ep := range episodes {
		if ep.Get("badge").String() == trailerBadge {
			continue
		}
		p := page{
			number:    len(in.pages) + 1,
			aid:       ep.Get("aid").String(),
			cid:       ep.Get("cid").String(),
			epid:      ep.Get("id").String(),
			title:     strings.TrimSpace(ep.Get("title").String() + " " + ep.Get("long_title").String()),
			duration:  time.Duration(ep.Get("duration").Int64Or(0)) * time.Millisecond,
			published: ep.Get("pub_time").Int64Or(0),
		}
		if in.add(p) && p.epid == epID {
			in.focus = len(in.pages)
		}
	}
}

// fetchCheese reads a course (pugv/view/web/season).
func (e *Extractor) fetchCheese(ctx context.Context, s *session, epID string) (*info, error) {
	j, err := e.getJSON(ctx, s, "https://api.bilibili.com/pugv/view/web/season?ep_id="+epID)
	if err != nil {
		return nil, err
	}
	data, err := field(j, "data")
	if err != nil {
		return nil, err
	}
	episodes, err := field(data, "episodes")
	if err != nil {
		return nil, err
	}
	owner := media.Person{Name: data.Path("up_info", "uname").String(), ID: data.Path("up_info", "mid").String()}
	in := &info{
		title:    trim(data.Get("title")),
		desc:     trim(data.Get("subtitle")),
		cover:    data.Get("cover").String(),
		owner:    owner,
		seasonID: data.Get("season_id").String(),
	}
	for i, ep := range episodes.Array() {
		p := page{
			number:    ep.Get("index").IntOr(i + 1),
			aid:       ep.Get("aid").String(),
			cid:       ep.Get("cid").String(),
			epid:      ep.Get("id").String(),
			title:     trim(ep.Get("title")),
			duration:  seconds(ep.Get("duration")),
			published: ep.Get("release_date").Int64Or(0),
			owner:     owner,
		}
		if in.add(p) && p.epid == epID {
			in.focus = len(in.pages)
		}
	}
	if len(in.pages) > 0 {
		in.published = in.pages[0].published
	}
	return in, nil
}

// fetchIntl reads a season of the international site (bilibili.tv), directly or through a proxy.
func (e *Extractor) fetchIntl(ctx context.Context, s *session, epID string) (*info, error) {
	host := "api.bilibili.tv"
	if s.biliPlus() {
		host = s.host
	}
	api := "https://" + host + "/intl/gateway/v2/ogv/view/app/season?ep_id=" + epID + "&platform=android&s_locale=zh_SG&mobi_app=bstar_a"
	if s.token != "" {
		api += "&access_key=" + s.token
	}
	j, err := e.getJSON(ctx, s, api)
	if err != nil {
		return nil, err
	}
	result, err := field(j, "result")
	if err != nil {
		return nil, err
	}
	in := &info{
		title:     result.Get("title").String(),
		desc:      result.Get("evaluate").String(),
		cover:     result.Get("cover").String(),
		published: pubTime(result.Path("publish", "pub_time")),
		season:    true,
		seasonID:  result.Get("season_id").String(),
	}
	if in.cover == "" {
		// Some seasons carry no poster in the app API; the mainland page's state has it.
		html, err := e.getText(ctx, s, "https://bangumi.bilibili.com/anime/"+in.seasonID)
		if err != nil {
			return nil, err
		}
		if m := initialState.FindStringSubmatch(html); m != nil {
			if state, err := jsonv.ParseString(m[1]); err == nil {
				mi := state.Get("mediaInfo")
				in.cover, in.title, in.desc = mi.Get("cover").String(), mi.Get("title").String(), mi.Get("evaluate").String()
			}
		}
	}
	in.title, in.desc = strings.TrimSpace(in.title), strings.TrimSpace(in.desc)
	episodes := result.Get("episodes").Array()
	for _, module := range result.Get("modules").Array() {
		if strings.Contains(module.String(), "/"+epID) {
			episodes = module.Path("data", "episodes").Array()
			break
		}
	}
	addEpisodes(in, episodes, epID)
	return in, nil
}
