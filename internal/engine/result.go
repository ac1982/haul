package engine

import (
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/fetch"
	"github.com/ac1982/haul/internal/media"
)

// Status is what happened to an entry.
type Status string

const (
	// Listed: streams listed (info), nothing downloaded.
	Listed Status = "listed"
	// Downloaded: the output was written.
	Downloaded Status = "downloaded"
	// Skipped: the output already existed, or the archive had the entry.
	Skipped Status = "skipped"
)

// Result is what a run found and did.
type Result struct {
	Item *media.Item
	Site extract.Info
	// Entries follow Item.Entries, one each.
	Entries []*EntryResult
}

// EntryResult is one entry's outcome.
type EntryResult struct {
	Entry *media.Entry
	// Selected: the run took this entry (-p, or the entry the link points at).
	Selected bool
	Status   Status
	// Reason says why an entry was skipped (exists, archive) or left unmuxed (no-mux).
	Reason string
	// File is the main output, absolute.
	File string
	Size int64
	// ExtraFiles are written next to it: subtitles, sidecars, a cover, unmuxed streams. Absolute.
	ExtraFiles []string
	// Video and Audio are the streams in the order haul chooses from; nil until loaded.
	Video []media.VideoFormat
	Audio []media.AudioFormat
	// ChosenVideo and ChosenAudio index them; -1 when none.
	ChosenVideo int
	ChosenAudio int
	// Subtitles are the tracks the entry offers after --sub-lang and auto filtering.
	Subtitles []media.Subtitle
	// VideoHasAudio: the video streams carry the audio (X); there are no audio streams.
	VideoHasAudio bool
}

// Files are the outputs of the entries the run took, whether written now or already there.
func (r *Result) Files() []string {
	var files []string
	for _, e := range r.Entries {
		if e.File != "" && (e.Status == Downloaded || e.Status == Skipped) {
			files = append(files, e.File)
		}
		files = append(files, e.ExtraFiles...)
	}
	return files
}

// Observer is how a run is shown to people. Calls come from the run's goroutine, except Transfer's progress.
type Observer interface {
	// Resolved: the item is known, and which entries the run takes.
	Resolved(item *media.Item, site extract.Info, selected []*media.Entry)
	// EntryStarted: n of count entries is about to be worked on.
	EntryStarted(n, count int, entry *media.Entry)
	// Streams: the entry's streams in choice order and the chosen ones (-1 none).
	Streams(entry *media.Entry, r *EntryResult)
	// Transfer starts a download's progress display; label is "video", "audio"…
	Transfer(label string) fetch.Progress
	// EntryFinished: the entry is done.
	EntryFinished(entry *media.Entry, r *EntryResult)
	// Finished: the run is done.
	Finished(r *Result)
}

// Chooser picks streams interactively; it returns indexes into the lists it is given.
type Chooser interface {
	Choose(entry *media.Entry, r *EntryResult) (video, audio int, err error)
}

// NopObserver shows nothing.
type NopObserver struct{}

func (NopObserver) Resolved(*media.Item, extract.Info, []*media.Entry) {}
func (NopObserver) EntryStarted(int, int, *media.Entry)                {}
func (NopObserver) Streams(*media.Entry, *EntryResult)                 {}
func (NopObserver) Transfer(string) fetch.Progress                     { return nil }
func (NopObserver) EntryFinished(*media.Entry, *EntryResult)           {}
func (NopObserver) Finished(*Result)                                   {}
