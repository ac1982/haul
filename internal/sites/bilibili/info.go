package bilibili

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// page is one downloadable unit as the fetchers find it: a part, an episode, a list item.
type page struct {
	// number is the site's own page number (what ?p= and "_P2_" titles refer to).
	number    int
	aid       string
	cid       string
	epid      string
	title     string
	duration  time.Duration
	published int64
	cover     string
	desc      string
	owner     media.Person
}

// info is what a link resolved to, before it becomes a media.Item.
type info struct {
	title     string
	desc      string
	cover     string
	published int64
	owner     media.Person
	pages     []page
	seen      map[string]bool
	// season: a bangumi season, finished or not.
	season      bool
	finished    bool
	interactive bool
	seasonID    string
	// listID is a favorites list's id when the link named none.
	listID string
	// focus is the 1-based position of the page the link pointed at, 0 for none.
	focus int
}

// add appends a page unless one with the same (aid, cid, epid) is already there.
func (in *info) add(p page) bool {
	if in.seen == nil {
		in.seen = map[string]bool{}
	}
	key := p.aid + "/" + p.cid + "/" + p.epid
	if in.seen[key] {
		return false
	}
	in.seen[key] = true
	in.pages = append(in.pages, p)
	return true
}

// apiError is an answer without the field it should have carried, with bilibili's own code and message.
type apiError struct {
	key     string
	code    string
	message string
}

func (e *apiError) Error() string {
	if e.code != "" && e.code != "0" {
		return "bilibili answered " + e.code + ": " + e.message
	}
	return "Missing JSON field: " + e.key
}

// field is j[key], which the answer must carry.
func field(j jsonv.Value, key string) (jsonv.Value, error) {
	if v := j.Get(key); v.Exists() {
		return v, nil
	}
	return jsonv.Null, &apiError{key: key, code: j.Get("code").String(), message: j.Get("message").String()}
}

// fetchInfo reads the metadata of an id. An episode id that is no bangumi may be a course.
func (e *Extractor) fetchInfo(ctx context.Context, s *session, id mediaID) (mediaID, *info, error) {
	in, err := e.fetch(ctx, s, id)
	var missing *apiError
	if id.kind == episode && errors.As(err, &missing) {
		console.Warn("No bangumi has this ep/ss id; trying it as a course")
		id = mediaID{kind: cheese, id: id.id}
		in, err = e.fetchCheese(ctx, s, id.id)
	}
	return id, in, err
}

func (e *Extractor) fetch(ctx context.Context, s *session, id mediaID) (*info, error) {
	switch id.kind {
	case video:
		return e.fetchVideo(ctx, s, id.id)
	case episode:
		if e.opts.API == Intl {
			return e.fetchIntl(ctx, s, id.id)
		}
		return e.fetchBangumi(ctx, s, id.id)
	case cheese:
		return e.fetchCheese(ctx, s, id.id)
	case space:
		return e.fetchSpace(ctx, s, id.id)
	case collection, series:
		return e.fetchMediaList(ctx, s, id.id, id.kind)
	case favorites:
		return e.fetchFavorites(ctx, s, id.id, id.mid)
	}
	return nil, errs.New("Cannot read %s", id)
}

func trim(v jsonv.Value) string { return strings.TrimSpace(v.String()) }

func seconds(v jsonv.Value) time.Duration { return time.Duration(v.Int64Or(0)) * time.Second }

// pubTime reads bilibili's "yyyy-MM-dd HH:mm:ss" publish time.
func pubTime(v jsonv.Value) int64 {
	t, _ := format.ParseDateTime(v.String())
	return t
}

// fetchVideo reads a video (x/web-interface/view): its parts, and the branches of an interactive video.
func (e *Extractor) fetchVideo(ctx context.Context, s *session, aid string) (*info, error) {
	j, err := e.getJSON(ctx, s, "https://api.bilibili.com/x/web-interface/view?aid="+aid)
	if err != nil {
		return nil, err
	}
	data, err := field(j, "data")
	if err != nil {
		return nil, err
	}
	owner := media.Person{Name: data.Path("owner", "name").String(), ID: data.Path("owner", "mid").String()}
	published := data.Get("pubdate").Int64Or(0)
	in := &info{
		title:       trim(data.Get("title")),
		desc:        trim(data.Get("desc")),
		cover:       data.Get("pic").String(),
		published:   published,
		owner:       owner,
		interactive: data.Path("rights", "is_stein_gate").IntOr(0) == 1,
	}
	parts, err := field(data, "pages")
	if err != nil {
		return nil, err
	}
	for i, p := range parts.Array() {
		in.add(page{number: p.Get("page").IntOr(i + 1), aid: aid, cid: p.Get("cid").String(), title: trim(p.Get("part")),
			duration: seconds(p.Get("duration")), published: published, owner: owner})
	}
	if in.interactive {
		if err := e.addBranches(ctx, s, in, aid, data.Get("bvid").String(), data.Get("cid").String()); err != nil {
			return nil, err
		}
	}
	// Licensed content normally has one page; then the episode id is what a playurl wants.
	if redirect := data.Get("redirect_url").String(); strings.Contains(redirect, "bangumi") && len(parts.Array()) == 1 {
		if m := epNumber.FindStringSubmatch(redirect); m != nil {
			for i := range in.pages {
				in.pages[i].epid = m[1]
			}
		}
	}
	return in, nil
}

var (
	epNumber    = regexp.MustCompile(`ep(\d+)`)
	interaction = regexp.MustCompile(`<interaction>([\s\S]*?)</interaction>`)
)

// addBranches: an interactive video hides its branches behind the player graph; each choice becomes a page,
// numbered on from 2.
func (e *Extractor) addBranches(ctx context.Context, s *session, in *info, aid, bvid, cid string) error {
	player, err := e.getText(ctx, s, "https://api.bilibili.com/x/player.so?bvid="+bvid+"&id=cid:"+cid)
	if err != nil {
		return err
	}
	unreadable := errs.New("Could not read the pages of this interactive video")
	m := interaction.FindStringSubmatch(player)
	if m == nil || m[1] == "" {
		return unreadable
	}
	graph, err := jsonv.ParseString(format.UnescapeEntities(m[1]))
	if err != nil {
		return unreadable
	}
	version, ok := graph.Get("graph_version").Int64()
	if !ok {
		return unreadable
	}
	edges, err := e.getJSON(ctx, s, "https://api.bilibili.com/x/stein/edgeinfo_v2?graph_version="+itoa(version)+"&bvid="+bvid)
	if err != nil {
		return err
	}
	number := 2
	for _, q := range edges.Path("data", "edges", "questions").Array() {
		for _, choice := range q.Get("choices").Array() {
			in.add(page{number: number, aid: aid, cid: choice.Get("cid").String(), title: trim(choice.Get("option")),
				published: in.published, owner: in.owner})
			number++
		}
	}
	return nil
}
