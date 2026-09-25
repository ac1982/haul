package podcast

import (
	"cmp"
	"math"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// episode is one episode, whichever site it came from.
type episode struct {
	id       string
	title    string
	mediaURL string
	// mediaType is a MIME type or a file extension, when the site states one.
	mediaType string
	// size is as stated by the site: it gives a bitrate to show, but is not trusted to drive ranged downloads.
	size        int64
	duration    time.Duration
	published   time.Time
	cover       string
	description string
	show        string
	author      string
	showID      string
	pageURL     string
	video       bool
}

// show is what a show link resolves to besides its episodes.
type show struct {
	id          string
	url         string
	title       string
	description string
	author      string
}

// showItem is a show: its episodes oldest first, so the last entry is the newest episode. Episodes published at
// the same moment keep the reverse of the site's (newest first) order.
func showItem(site media.Site, s show, episodes []episode, header http.Header) *media.Item {
	order := make([]int, len(episodes))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int {
		if c := episodes[a].published.Compare(episodes[b].published); c != 0 {
			return c
		}
		return cmp.Compare(b, a)
	})
	item := &media.Item{
		Site:        site,
		ID:          s.id,
		URL:         s.url,
		Title:       s.title,
		Description: s.description,
		Uploader:    media.Person{Name: s.author, ID: s.id},
		// No Thumbnail: each entry brings its own cover.
		Collection: true,
	}
	for i, idx := range order {
		item.Entries = append(item.Entries, entry(episodes[idx], i+1, header))
	}
	if n := len(item.Entries); n > 0 {
		item.Published = item.Entries[n-1].Published
	}
	return item
}

// episodeItem is a single episode.
func episodeItem(site media.Site, e episode, header http.Header) *media.Item {
	return &media.Item{
		Site:        site,
		ID:          e.id,
		URL:         e.pageURL,
		Title:       e.title,
		Description: e.description,
		Uploader:    media.Person{Name: e.author, ID: e.showID},
		Published:   e.published,
		Thumbnail:   e.cover,
		Entries:     []*media.Entry{entry(e, 1, header)},
	}
}

func entry(e episode, index int, header http.Header) *media.Entry {
	return &media.Entry{
		Index:       index,
		ID:          e.id,
		Title:       e.title,
		Description: e.description,
		Duration:    e.duration,
		Published:   e.published,
		Thumbnail:   e.cover,
		Uploader:    media.Person{Name: e.author, ID: e.showID},
		URL:         e.pageURL,
		Album:       e.show,
		Formats:     formats(e, header),
	}
}

// formats is the episode's one file. Stated sizes are not trusted (hosts redirect through trackers to files that
// differ), so the resource size is left for the downloader to find.
func formats(e episode, header http.Header) *media.Formats {
	var kbps int64
	if secs := e.duration.Seconds(); e.size > 0 && secs > 0 {
		kbps = int64(math.Round(float64(e.size) * 8 / 1024 / secs))
	}
	source := media.Resource{URL: e.mediaURL, Header: header.Clone(), Policy: media.Parallel}
	if e.video {
		// A video podcast is one file with the audio inside, like an X post.
		return &media.Formats{Video: []media.VideoFormat{{
			ID: "0", Quality: "original", Codec: "MP4", Bitrate: kbps, HasAudio: true, Source: source,
		}}}
	}
	return &media.Formats{
		AudioOnly: true,
		Audio:     []media.AudioFormat{{ID: "0", Codec: audioCodec(e.mediaType, e.mediaURL), Bitrate: kbps, Source: source}},
	}
}

// audioCodec is MP3 or M4A from a MIME type or the URL's extension. With no hint at all it is MP3, as most
// podcasts are; a format we do not know keeps its own name, and so goes into an M4A (which carries nearly anything).
func audioCodec(mediaType, mediaURL string) string {
	var hints []string
	if t := strings.ToLower(mediaType); t != "" {
		hints = append(hints, t)
	}
	if u, err := url.Parse(mediaURL); err == nil {
		if ext := strings.ToLower(strings.TrimPrefix(path.Ext(u.Path), ".")); ext != "" {
			hints = append(hints, ext)
		}
	}
	for _, h := range hints {
		if strings.Contains(h, "mpeg") || strings.Contains(h, "mp3") {
			return "MP3"
		}
		for _, m4a := range []string{"mp4", "m4a", "m4b", "aac"} {
			if strings.Contains(h, m4a) {
				return "M4A"
			}
		}
	}
	if len(hints) == 0 {
		return "MP3"
	}
	last := hints[len(hints)-1]
	return strings.ToUpper(last[strings.LastIndex(last, "/")+1:])
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

func seconds(s int64) time.Duration { return time.Duration(s) * time.Second }
