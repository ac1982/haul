// Package engine turns a link into files: resolve the item, select entries, choose streams, plan the output,
// download, mux. It works on the media model only and never asks which site it is on. People see it through an
// Observer (and answer through a Chooser); scripts get the structured Result.
package engine

import "time"

// Tracks is which media streams an output gets.
type Tracks int

const (
	// VideoAndAudio is the default: one file with both.
	VideoAndAudio Tracks = iota
	AudioOnly
	VideoOnly
	// NoTracks: no media file at all (only subtitles, a cover, or sidecars).
	NoTracks
)

// Placement is where subtitles or the cover go.
type Placement int

const (
	// Embed into the media file.
	Embed Placement = iota
	// Files next to where the media file would be.
	Files
	// Skip them.
	Skip
)

// Content is what a run produces for each entry.
type Content struct {
	Tracks    Tracks
	Subtitles Placement
	Cover     Placement
	// Sidecars are the side-file kinds to write too, e.g. "danmaku".
	Sidecars []string
	// NoMux keeps the downloaded streams as they are instead of muxing them into one file.
	NoMux bool
}

// DefaultContent is video and audio with subtitles and cover embedded.
func DefaultContent() Content { return Content{Tracks: VideoAndAudio, Subtitles: Embed, Cover: Embed} }

// Options configure a run.
type Options struct {
	// Dir is where files go; "" is the current directory.
	Dir string
	// Pages selects entries: 8, 1,2, 3-5, ALL, LAST. "" is the entry the link points at, else all.
	Pages   string
	Content Content

	// Quality and Codec are priorities, best first; unlisted ones follow, best first.
	Quality []string
	Codec   []string
	// CodecFirst sorts by codec before quality when both are given.
	CodecFirst     bool
	VideoAscending bool
	AudioAscending bool
	// VideoIndex and AudioIndex pick a stream by its index in the sorted lists; -1 takes the first.
	VideoIndex  int
	AudioIndex  int
	Interactive bool

	// SubtitleLangs keeps only these languages (a prefix matches: en takes en-US); empty keeps all.
	SubtitleLangs []string
	// AutoSubtitles keeps machine-generated subtitles too.
	AutoSubtitles bool

	// Template names the file of an item with one entry, ListTemplate of an item with several; "" uses the defaults.
	Template      string
	ListTemplate  string
	AudioLanguage string
	NoTags        bool
	UseMP4Box     bool

	// Archive remembers downloaded entries and skips them next time.
	Archive bool
	Delay   time.Duration

	// List only lists streams (haul info); IncludeURLs adds their URLs.
	List        bool
	IncludeURLs bool
}

// DefaultOptions is a plain download.
func DefaultOptions() Options {
	return Options{Content: DefaultContent(), VideoIndex: -1, AudioIndex: -1}
}
