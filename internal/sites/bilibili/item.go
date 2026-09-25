package bilibili

import (
	"strconv"
	"time"

	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
)

// ref is what Formats needs to ask for an entry's streams.
type ref struct {
	kind kind // video, episode or cheese: which playurl endpoints apply
	aid  string
	cid  string
	epid string
	api  API
	// position is the entry's 1-based place in its season (the intl subtitle list is indexed by it).
	position int
}

// bangumi: an episode proper, not a course.
func (r *ref) bangumi() bool { return r.kind == episode }

// buildItem turns fetched info into the item. link is the input, resolved the link as finally read.
func buildItem(link, resolved string, id mediaID, in *info, api API) *media.Item {
	item := &media.Item{
		Site:        Site,
		Title:       in.title,
		Description: in.desc,
		Uploader:    in.owner,
		Thumbnail:   in.cover,
		Collection:  in.season && !in.finished,
	}
	if in.published > 0 {
		item.Published = time.Unix(in.published, 0)
	}
	pageKind := video
	if id.kind.pgc() {
		pageKind = id.kind
	}
	for i, p := range in.pages {
		r := &ref{kind: pageKind, aid: p.aid, cid: p.cid, epid: p.epid, api: api, position: i + 1}
		bv := bvid(p.aid)
		entry := &media.Entry{
			Index:       i + 1,
			ID:          entryID(r, bv),
			Title:       p.title,
			Description: p.desc,
			Duration:    p.duration,
			Thumbnail:   p.cover,
			Uploader:    p.owner,
			URL:         entryURL(r, bv, in.seasonID),
			Fields:      map[string]string{"bvid": bv, "aid": p.aid, "cid": p.cid, "api": api.Label()},
			Ref:         r,
		}
		if entry.Description == "" {
			entry.Description = in.desc
		}
		if entry.Thumbnail == "" {
			entry.Thumbnail = in.cover
		}
		if entry.Uploader == (media.Person{}) {
			entry.Uploader = in.owner
		}
		if p.published > 0 {
			entry.Published = time.Unix(p.published, 0)
		}
		item.Entries = append(item.Entries, entry)
	}
	item.ID, item.URL = itemIdentity(id, in, api, item.Entries, resolved)
	item.Focus = in.focus
	if item.Focus == 0 {
		item.Focus = focusFromLink(in, link, resolved)
	}
	return item
}

// focusFromLink is the entry ?p=N names, by the site's page number.
func focusFromLink(in *info, links ...string) int {
	for _, l := range links {
		n, err := strconv.Atoi(format.QueryValue("p", l))
		if err != nil {
			continue
		}
		for i, p := range in.pages {
			if p.number == n {
				return i + 1
			}
		}
	}
	return 0
}

// entryID is the BV id of a video's page, ep<id> for an episode or lesson, av<aid> when there is nothing better.
func entryID(r *ref, bv string) string {
	switch {
	case r.kind.pgc() && r.epid != "":
		return "ep" + r.epid
	case bv != "":
		return bv
	}
	return "av" + r.aid
}

func entryURL(r *ref, bv, seasonID string) string {
	switch {
	case r.kind == cheese && r.epid != "":
		return "https://www.bilibili.com/cheese/play/ep" + r.epid
	case r.kind == episode && r.api == Intl && seasonID != "":
		return "https://www.bilibili.tv/play/" + seasonID + "/" + r.epid
	case r.kind == episode && r.epid != "":
		return "https://www.bilibili.com/bangumi/play/ep" + r.epid
	case bv != "":
		return "https://www.bilibili.com/video/" + bv + "/"
	}
	return "https://www.bilibili.com/video/av" + r.aid + "/"
}

// itemIdentity is the item's id and canonical link.
func itemIdentity(id mediaID, in *info, api API, entries []*media.Entry, resolved string) (string, string) {
	switch id.kind {
	case video:
		if bv := bvid(id.id); bv != "" {
			return bv, "https://www.bilibili.com/video/" + bv + "/"
		}
		return "av" + id.id, resolved
	case episode, cheese:
		prefix := "https://www.bilibili.com/bangumi/play/ss"
		if id.kind == cheese {
			prefix = "https://www.bilibili.com/cheese/play/ss"
		}
		if api == Intl && id.kind == episode {
			prefix = "https://www.bilibili.tv/play/"
		}
		if in.seasonID != "" && in.seasonID != "0" {
			return "ss" + in.seasonID, prefix + in.seasonID
		}
		for _, e := range entries {
			if e.ID == "ep"+id.id {
				return e.ID, e.URL
			}
		}
		return "ep" + id.id, resolved
	case space:
		return "space" + id.id, "https://space.bilibili.com/" + id.id + "/upload/video"
	case favorites:
		return "fav" + in.listID, "https://space.bilibili.com/" + id.mid + "/favlist?fid=" + in.listID
	}
	return id.String(), resolved
}
