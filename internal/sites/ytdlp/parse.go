package ytdlp

import (
	"fmt"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// parse turns `yt-dlp -J` output into an item. A post with several videos (X) comes back as a playlist; each of
// its videos becomes an entry.
func parse(root jsonv.Value, p *profile) (*media.Item, error) {
	nodes := []jsonv.Value{root}
	if s, _ := root.Get("_type").Str(); s == "playlist" {
		nodes = nil
		for _, n := range root.Get("entries").Array() {
			if n.IsObject() {
				nodes = append(nodes, n)
			}
		}
	}
	if len(nodes) == 0 {
		return nil, errs.New("yt-dlp found no downloadable video")
	}
	entries := make([]*media.Entry, len(nodes))
	for i, n := range nodes {
		title := ""
		if len(nodes) > 1 {
			// The videos of a post share its text; numbering tells them apart.
			title = fmt.Sprintf("Video %d", i+1)
		}
		e, err := entry(n, i+1, title, root, p)
		if err != nil {
			return nil, err
		}
		entries[i] = e
	}
	first := entries[0]
	item := &media.Item{
		Site:        p.info.Site,
		ID:          root.Get("id").String(),
		URL:         firstString(root.Get("webpage_url"), jsonv.Of(first.URL)),
		Title:       first.Title,
		Description: root.Get("description").String(),
		Uploader:    owner(root, nodes[0]),
		Published:   first.Published,
		Entries:     entries,
	}
	if t, ok := root.Get("title").Str(); ok {
		item.Title = p.title(t)
	}
	// Several videos have several covers; each entry brings its own.
	if len(entries) == 1 {
		item.Thumbnail = first.Thumbnail
	}
	return item, nil
}

// entry is one video; title overrides yt-dlp's when the post has several. Fields a playlist entry leaves out come
// from the post (parent).
func entry(n jsonv.Value, index int, title string, parent jsonv.Value, p *profile) (*media.Entry, error) {
	id, err := n.Require("id")
	if err != nil {
		return nil, errs.New("yt-dlp answered without a video id")
	}
	formats, header, err := streams(n.Get("formats"), p)
	if err != nil {
		return nil, err
	}
	formats.Subtitles = subtitles(n, header)
	formats.Raw = n.Serialized()
	if title == "" {
		title = p.title(n.Get("title").String())
	}
	return &media.Entry{
		Index:       index,
		ID:          id.String(),
		Title:       title,
		Description: firstString(n.Get("description"), parent.Get("description")),
		Duration:    seconds(n.Get("duration").FloatOr(0)),
		Published:   published(n, parent),
		Thumbnail:   cover(n),
		Uploader:    owner(n, parent),
		URL:         firstString(n.Get("webpage_url"), parent.Get("webpage_url")),
		Chapters:    chapters(n.Get("chapters")),
		Formats:     formats,
	}, nil
}

// title applies the site's title rule.
func (p *profile) title(text string) string {
	if p.trimEllipsis && strings.HasSuffix(text, "...") {
		return strings.TrimSpace(strings.TrimSuffix(text, "..."))
	}
	return text
}

// owner is the channel (YouTube) or account (X), from the node or else its parent.
func owner(n, parent jsonv.Value) media.Person {
	return media.Person{
		Name: firstString(n.Get("channel"), n.Get("uploader"), parent.Get("channel"), parent.Get("uploader")),
		ID:   firstString(n.Get("channel_id"), n.Get("uploader_id"), parent.Get("channel_id"), parent.Get("uploader_id")),
	}
}

// published is the exact timestamp, or else the upload day at midnight UTC.
func published(n, parent jsonv.Value) time.Time {
	for _, v := range []jsonv.Value{n.Get("timestamp"), parent.Get("timestamp")} {
		if ts, ok := v.Int64(); ok && ts != 0 {
			return time.Unix(ts, 0).UTC()
		}
	}
	for _, v := range []jsonv.Value{n.Get("upload_date"), parent.Get("upload_date")} {
		if day, ok := v.Str(); ok {
			if t, err := time.Parse("20060102", day); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

func chapters(list jsonv.Value) []media.Chapter {
	var out []media.Chapter
	for _, c := range list.Array() {
		out = append(out, media.Chapter{
			Title: c.Get("title").String(),
			Start: seconds(c.Get("start_time").FloatOr(0)),
			End:   seconds(c.Get("end_time").FloatOr(0)),
		})
	}
	return out
}

// cover is the largest JPEG thumbnail, as yt-dlp ranks them; the muxers cannot embed WebP.
func cover(n jsonv.Value) string {
	var best jsonv.Value
	better := func(a, b jsonv.Value) bool {
		pa, pb := a.Get("preference").IntOr(-100), b.Get("preference").IntOr(-100)
		if pa != pb {
			return pa > pb
		}
		return a.Get("width").IntOr(0) > b.Get("width").IntOr(0)
	}
	for _, t := range n.Get("thumbnails").Array() {
		if strings.Contains(t.Get("url").String(), ".jpg") && (best.IsNull() || better(t, best)) {
			best = t
		}
	}
	if u := best.Get("url").String(); u != "" {
		return u
	}
	return n.Get("thumbnail").String()
}

// firstString is the first value that is a non-empty string.
func firstString(values ...jsonv.Value) string {
	for _, v := range values {
		if s, ok := v.Str(); ok && s != "" {
			return s
		}
	}
	return ""
}

func seconds(s float64) time.Duration { return time.Duration(s * float64(time.Second)) }
