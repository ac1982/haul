// Package media is haul's domain model, the same for every site: an Item (a video, a post, a season, a show, a list)
// made of Entries, and for each entry the Formats it can be downloaded in. Extractors produce it; the engine consumes
// it without knowing which site it came from.
package media

import (
	"context"
	"net/http"
	"time"
)

// Site identifies where an item comes from.
type Site string

// Person is an uploader, channel, author or host.
type Person struct {
	Name string
	ID   string
}

// Item is what a link resolves to.
type Item struct {
	Site        Site
	ID          string
	URL         string
	Title       string
	Description string
	Uploader    Person
	Published   time.Time
	Thumbnail   string
	Entries     []*Entry
	// Focus is the entry the link points at (an episode link, ?p=2), 1-based; 0 when it points at the whole item.
	Focus int
	// Collection keeps the folder layout even for one entry (an ongoing season). Items with several entries always
	// get a folder named after them.
	Collection bool
	// LoggedIn is whether the site treated us as logged in; nil where that does not apply.
	LoggedIn *bool
}

// Entry is one downloadable unit of an item: a part, an episode, a video of a post.
type Entry struct {
	// Index is 1-based and is what -p takes.
	Index       int
	ID          string
	Title       string
	Description string
	Duration    time.Duration
	Published   time.Time
	Thumbnail   string
	Uploader    Person
	URL         string
	// Album is the show a podcast episode belongs to: it becomes the album tag, with the episode as the title.
	Album    string
	Chapters []Chapter
	// Fields are extra file-name template variables the site offers (bilibili: bvid, aid, cid).
	Fields map[string]string
	// Formats, when the site's metadata already carried them; otherwise the extractor loads them on demand.
	Formats *Formats
	// Ref is extractor-private: whatever it needs to load the entry's formats later.
	Ref any
}

// Chapter is a named span of the entry.
type Chapter struct {
	Title string
	Start time.Duration
	End   time.Duration
}

// Formats is everything an entry can be downloaded as.
type Formats struct {
	Video []VideoFormat
	Audio []AudioFormat
	// Subtitles the entry offers.
	Subtitles []Subtitle
	// Chapters found with the formats (e.g. bangumi opening / ending markers), used when the entry has none.
	Chapters []Chapter
	// ExtraAudio is muxed in as additional tracks (bilibili dubbing: background and voice roles).
	ExtraAudio []ExtraAudio
	// Sidecars are side files a site can add next to the output, like bilibili's danmaku.
	Sidecars []Sidecar
	// AudioOnly: the entry is audio by nature (a podcast episode) and is saved as audio in its codec's container.
	AudioOnly bool
	// Raw is the site's answer, kept for --debug.
	Raw string
}

// VideoFormat is one video stream; with HasAudio it is a whole file with the audio inside.
type VideoFormat struct {
	ID string
	// Rank orders qualities, higher is better (bilibili quality ids; height × 1000 + fps elsewhere).
	Rank     int
	Quality  string
	Width    int
	Height   int
	FPS      float64
	Codec    string
	Bitrate  int64 // kbps
	Size     int64 // bytes as the site states or estimates them; 0 when unknown
	HasAudio bool
	Source   Resource
	// Parts, instead of Source: segments downloaded one after the other and joined (bilibili FLV).
	Parts []Resource
}

// AudioFormat is one audio stream.
type AudioFormat struct {
	ID      string
	Codec   string
	Bitrate int64 // kbps
	Size    int64
	Source  Resource
}

// ExtraAudio is an additional audio track: its stream and the tags it gets.
type ExtraAudio struct {
	Title  string
	Artist string
	Audio  AudioFormat
}

// SubtitleFormat is how a subtitle file is encoded; everything is converted to SRT.
type SubtitleFormat string

const (
	SRT          SubtitleFormat = "srt"
	YouTubeJSON3 SubtitleFormat = "json3"
	BilibiliJSON SubtitleFormat = "bilijson"
	ASS          SubtitleFormat = "ass"
)

// Subtitle is one subtitle track.
type Subtitle struct {
	// Lang is a BCP 47 tag (en, en-US, zh-Hans).
	Lang string
	// Auto marks machine-generated subtitles, skipped unless asked for.
	Auto   bool
	Format SubtitleFormat
	Source Resource
}

// Sidecar is a side file a site can write next to the output. Write gets the output path without extension and
// returns the files it wrote.
type Sidecar struct {
	Kind  string
	Write func(ctx context.Context, base string) ([]string, error)
}

// RangePolicy is how a resource may be fetched.
type RangePolicy int

const (
	// Parallel: byte ranges over several connections (the default for CDN media).
	Parallel RangePolicy = iota
	// Sequential: only closed ranges of at most MaxRange bytes, one at a time (googlevideo refuses anything else).
	Sequential
	// Whole: one plain request, resumed with an open range when it drops.
	Whole
)

// Resource is a file to fetch, fully prepared by its extractor: URL, headers, and how to ask for it.
type Resource struct {
	URL    string
	Header http.Header
	// Size is the exact size when the site is trusted to know it; it saves a probe. 0 when unknown.
	Size     int64
	Policy   RangePolicy
	MaxRange int64
}

// EstimatedSize is the stated size, or bitrate × duration.
func EstimatedSize(stated, kbps int64, d time.Duration) int64 {
	if stated > 0 {
		return stated
	}
	return int64(d.Seconds() * float64(kbps) * 1000 / 8)
}
