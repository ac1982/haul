package bilibili

import (
	"context"
	"fmt"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
)

func pageCount(total, size int) int { return (total + size - 1) / size }

// addVideoPages adds the parts of a video read on its own as list items titled title, or "<title>_P<n>_<part>"
// when there are several.
func addVideoPages(in *info, sub *info, title, desc string, multi bool) {
	for _, p := range sub.pages {
		if multi {
			p.title = fmt.Sprintf("%s_P%d_%s", title, p.number, p.title)
		} else {
			p.title = title
		}
		p.cover, p.desc = sub.cover, desc
		in.add(p)
	}
}

// fetchFavorites reads a favorites list, or the user's first one when fid is "".
func (e *Extractor) fetchFavorites(ctx context.Context, s *session, fid, mid string) (*info, error) {
	if fid == "" {
		j, err := e.getJSON(ctx, s, "https://api.bilibili.com/x/v3/fav/folder/created/list-all?up_mid="+mid)
		if err != nil {
			return nil, err
		}
		data, err := field(j, "data")
		if err != nil {
			return nil, err
		}
		if fid = data.Get("list").At(0).Get("id").String(); fid == "" {
			return nil, errs.New("This user has no public favorites list")
		}
	}
	const size = 20
	listURL := func(pn int) string {
		return fmt.Sprintf("https://api.bilibili.com/x/v3/fav/resource/list?media_id=%s&pn=%d&ps=%d&order=mtime&type=2&tid=0&platform=web", fid, pn, size)
	}
	first, err := e.getJSON(ctx, s, listURL(1))
	if err != nil {
		return nil, err
	}
	data, err := field(first, "data")
	if err != nil {
		return nil, err
	}
	medias := data.Get("medias").Array()
	for pn := 2; pn <= pageCount(data.Path("info", "media_count").IntOr(0), size); pn++ {
		j, err := e.getJSON(ctx, s, listURL(pn))
		if err != nil {
			return nil, err
		}
		medias = append(medias, j.Path("data", "medias").Array()...)
	}

	meta := data.Get("info")
	in := &info{
		listID:    fid,
		title:     trim(meta.Get("title")),
		desc:      trim(meta.Get("intro")),
		published: meta.Get("ctime").Int64Or(0),
		owner:     media.Person{Name: meta.Path("upper", "name").String(), ID: meta.Path("upper", "mid").String()},
	}
	for _, m := range medias {
		if m.Get("attr").IntOr(0) != 0 { // unavailable
			continue
		}
		if m.Get("page").IntOr(1) > 1 {
			sub, err := e.fetchVideo(ctx, s, m.Get("id").String())
			if err != nil {
				return nil, err
			}
			addVideoPages(in, sub, m.Get("title").String(), m.Get("intro").String(), true)
			continue
		}
		in.add(page{number: len(in.pages) + 1, aid: m.Get("id").String(), cid: m.Path("ugc", "first_cid").String(),
			title: m.Get("title").String(), duration: seconds(m.Get("duration")), published: m.Get("pubtime").Int64Or(0),
			cover: m.Get("cover").String(), desc: m.Get("intro").String(),
			owner: media.Person{Name: m.Path("upper", "name").String(), ID: m.Path("upper", "mid").String()}})
	}
	return in, nil
}

// fetchMediaList reads a collection (a "season" of a space) or a series, 20 items at a time. A collection id that
// reads as nothing is tried as a series before giving up.
func (e *Extractor) fetchMediaList(ctx context.Context, s *session, bizID string, k kind) (*info, error) {
	listType, name := 8, "collection"
	if k == series {
		listType, name = 5, "series"
	}
	j, err := e.getJSON(ctx, s, fmt.Sprintf("https://api.bilibili.com/x/v1/medialist/info?type=%d&biz_id=%s&tid=0", listType, bizID))
	if err != nil {
		return nil, err
	}
	data := j.Get("data")
	if !data.IsObject() {
		if k == collection {
			if in, err := e.fetchMediaList(ctx, s, bizID, series); err == nil {
				return in, nil
			}
		}
		return nil, errs.New("Could not read the %s (code %s): %s", name, j.Get("code").String(), j.Get("message").String())
	}
	in := &info{
		title:     trim(data.Get("title")),
		desc:      trim(data.Get("intro")),
		published: data.Get("ctime").Int64Or(0),
		owner:     media.Person{Name: data.Path("upper", "name").String(), ID: data.Path("upper", "mid").String()},
	}
	desc, extra := "false", ""
	if k == series {
		desc, extra = "true", "&bvid="
	}
	oid := ""
	for more := true; more; {
		list, err := e.getJSON(ctx, s, fmt.Sprintf("https://api.bilibili.com/x/v2/medialist/resource/list?type=%d&oid=%s&otype=2&biz_id=%s%s&with_current=true&mobi_app=web&ps=20&direction=false&sort_field=1&tid=0&desc=%s",
			listType, oid, bizID, extra, desc))
		if err != nil {
			return nil, err
		}
		ldata := list.Get("data")
		if !ldata.IsObject() {
			return nil, errs.New("Could not read the video list (code %s): %s", list.Get("code").String(), list.Get("message").String())
		}
		more, _ = ldata.Get("has_more").Bool()
		items := ldata.Get("media_list").Array()
		if len(items) == 0 {
			break // has_more without items would never end
		}
		for _, m := range items {
			oid = m.Get("id").String()
			if m.Has("attr") && m.Get("attr").IntOr(0) != 0 {
				continue
			}
			parts := m.Get("page").IntOr(1)
			for _, pg := range m.Get("pages").Array() {
				title := m.Get("title").String()
				if parts != 1 {
					title = fmt.Sprintf("%s_P%s_%s", title, pg.Get("page").String(), pg.Get("title").String())
				}
				in.add(page{number: len(in.pages) + 1, aid: m.Get("id").String(), cid: pg.Get("id").String(), title: title,
					duration: seconds(pg.Get("duration")), published: m.Get("pubtime").Int64Or(0),
					cover: m.Get("cover").String(), desc: m.Get("intro").String(),
					owner: media.Person{Name: m.Path("upper", "name").String(), ID: m.Path("upper", "mid").String()}})
			}
		}
	}
	return in, nil
}

// fetchSpace reads all uploads of a user. The upload list has no cids, so each video costs a request of its own;
// they are paced so bilibili does not start refusing.
func (e *Extractor) fetchSpace(ctx context.Context, s *session, mid string) (*info, error) {
	key, err := e.wbiKey(ctx, s)
	if err != nil {
		return nil, err
	}
	user, err := e.getJSON(ctx, s, "https://api.live.bilibili.com/live_user/v1/Master/info?uid="+mid)
	if err != nil {
		return nil, err
	}
	userName := user.Path("data", "info", "uname").String()

	const size = 50
	listURL := func(pn int) string {
		return "https://api.bilibili.com/x/space/wbi/arc/search?" +
			wbiSign(fmt.Sprintf("mid=%s&order=pubdate&pn=%d&ps=%d&tid=0&wts=%s", mid, pn, size, format.UnixSeconds()), key)
	}
	first, err := e.getJSON(ctx, s, listURL(1))
	if err != nil {
		return nil, err
	}
	data, err := field(first, "data")
	if err != nil {
		return nil, err
	}
	items := data.Path("list", "vlist").Array()
	for pn := 2; pn <= pageCount(data.Path("page", "count").IntOr(len(items)), size); pn++ {
		j, err := e.getJSON(ctx, s, listURL(pn))
		if err != nil {
			return nil, err
		}
		items = append(items, j.Path("data", "list", "vlist").Array()...)
	}
	console.Status(fmt.Sprintf("%d uploads; reading their pages", len(items)))

	in := &info{title: userName + "'s uploads", owner: media.Person{Name: userName, ID: mid}}
	for n, item := range items {
		if n > 0 {
			if err := sleep(ctx, 200*time.Millisecond); err != nil {
				return nil, err
			}
		}
		if (n+1)%20 == 0 || n+1 == len(items) {
			console.Status(fmt.Sprintf("Uploads read %d/%d", n+1, len(items)))
		}
		aid := item.Get("aid").String()
		sub, err := e.fetchVideo(ctx, s, aid)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			console.Warn("Skipping av" + aid + ": could not read it")
			continue
		}
		addVideoPages(in, sub, sub.title, sub.desc, len(sub.pages) > 1)
	}
	return in, nil
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
