package engine

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
)

// Default file-name templates. The extension is added from the container.
const (
	DefaultTemplate     = "<title>"
	DefaultListTemplate = "<title>/[P<pageNumberWithZero>]<pageTitle>"
)

// TemplateVariables are the variables templates know, with what they mean. Sites add their own (bilibili: bvid,
// aid, cid, api), listed by `haul templates`.
var TemplateVariables = [][2]string{
	{"title", "title of the video, post, show or list"},
	{"pageNumber", "page / episode number"},
	{"pageNumberWithZero", "page number, zero-padded to the page count"},
	{"pageTitle", "page / episode title"},
	{"id", "the page's id on its site (YouTube id, BV…, episode id)"},
	{"site", "youtube, x, bilibili, xiaoyuzhou, applePodcasts"},
	{"uploader", "uploader / channel / host name"},
	{"uploaderId", "uploader id"},
	{"quality", "video quality label (1080p, 4K…)"},
	{"resolution", "video resolution"},
	{"fps", "video frame rate"},
	{"videoCodec", "video codec"},
	{"videoBitrate", "video bitrate, kbps"},
	{"audioCodec", "audio codec"},
	{"audioBitrate", "audio bitrate, kbps"},
	{"publishDate", "publish date of the item or list"},
	{"pageDate", "publish date of the page"},
	{"bvid", "bilibili BV id"},
	{"aid", "bilibili aid"},
	{"cid", "bilibili cid"},
	{"api", "bilibili API: WEB / TV / APP / INTL"},
}

const defaultDateFormat = "yyyy-MM-dd_HH-mm-ss"

var variable = regexp.MustCompile(`<([\w:\-.]+?)>`)

// templateData is what a template can name.
type templateData struct {
	item  *media.Item
	entry *media.Entry
	video *media.VideoFormat
	audio *media.AudioFormat
}

// Render fills a template and returns a relative path without extension. Every value is made safe for a file
// name; a slash in the template makes folders.
func (d templateData) Render(tpl string) string {
	tpl = strings.ReplaceAll(tpl, `\`, "/")
	out := variable.ReplaceAllStringFunc(tpl, func(m string) string {
		key := m[1 : len(m)-1]
		dateFormat := defaultDateFormat
		for _, prefix := range []string{"publishDate:", "pageDate:"} {
			if strings.HasPrefix(key, prefix) {
				dateFormat = key[len(prefix):]
				key = prefix[:len(prefix)-1]
			}
		}
		v, ok := d.value(key, dateFormat)
		if !ok {
			return m
		}
		return format.CleanName(v)
	})
	var parts []string
	for _, p := range strings.Split(out, "/") {
		p = strings.TrimSpace(p)
		if p == "" || p == "." || p == ".." {
			continue
		}
		if strings.HasPrefix(p, ".") {
			p = "_" + p[1:]
		}
		parts = append(parts, p)
	}
	rel := path.Join(parts...)
	for _, ext := range []string{".mp4", ".m4a", ".mp3"} {
		rel = strings.TrimSuffix(rel, ext)
	}
	if rel == "" {
		rel = "untitled"
	}
	return rel
}

func (d templateData) value(key, dateFormat string) (string, bool) {
	e, it := d.entry, d.item
	date := func(t time.Time) string {
		if t.IsZero() {
			return "null"
		}
		return format.FormatDate(t, dateFormat)
	}
	switch key {
	case "title":
		return it.Title, true
	case "pageNumber":
		return strconv.Itoa(e.Index), true
	case "pageNumberWithZero":
		return fmt.Sprintf("%0*d", len(strconv.Itoa(len(it.Entries))), e.Index), true
	case "pageTitle":
		return e.Title, true
	case "id":
		return e.ID, true
	case "site":
		return string(it.Site), true
	case "uploader":
		return cmpOr(e.Uploader.Name, it.Uploader.Name), true
	case "uploaderId":
		return cmpOr(e.Uploader.ID, it.Uploader.ID), true
	case "publishDate":
		return date(cmpOrTime(it.Published, e.Published)), true
	case "pageDate":
		return date(cmpOrTime(e.Published, it.Published)), true
	case "quality", "resolution", "fps", "videoCodec", "videoBitrate":
		v := d.video
		if v == nil {
			return "", true
		}
		switch key {
		case "quality":
			return v.Quality, true
		case "resolution":
			if v.Width == 0 {
				return "", true
			}
			return fmt.Sprintf("%dx%d", v.Width, v.Height), true
		case "fps":
			if v.FPS == 0 {
				return "", true
			}
			return strconv.FormatFloat(v.FPS, 'f', -1, 64), true
		case "videoCodec":
			return v.Codec, true
		default:
			return strconv.FormatInt(v.Bitrate, 10), true
		}
	case "audioCodec":
		if d.audio == nil {
			return "", true
		}
		return d.audio.Codec, true
	case "audioBitrate":
		if d.audio == nil {
			return "", true
		}
		return strconv.FormatInt(d.audio.Bitrate, 10), true
	}
	if v, ok := e.Fields[key]; ok {
		return v, true
	}
	for _, known := range TemplateVariables {
		if known[0] == key {
			return "", true
		}
	}
	return "", false
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func cmpOrTime(a, b time.Time) time.Time {
	if !a.IsZero() {
		return a
	}
	return b
}
