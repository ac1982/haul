package render

import (
	"fmt"
	"sync"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/engine"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/fetch"
	"github.com/ac1982/haul/internal/media"
)

// Terminal shows a run to people and asks them to choose streams. It is the engine's Observer and Chooser.
type Terminal struct {
	// ShowAll lists every entry and every stream; HideStreams prints no stream table; URLs adds stream URLs.
	ShowAll     bool
	HideStreams bool
	URLs        bool
	// List is haul info: tables are shown in full.
	List bool
	// Interactive: the chooser draws the table itself.
	Interactive bool

	item     *media.Item
	site     extract.Info
	count    int
	started  time.Time
	entryAt  time.Time
	finished int
}

var _ engine.Observer = (*Terminal)(nil)
var _ engine.Chooser = (*Terminal)(nil)

func (t *Terminal) style() console.Style { return console.CurrentStyle() }

// Resolved prints the header and the entry list, and which entries the run takes.
func (t *Terminal) Resolved(item *media.Item, site extract.Info, selected []*media.Entry) {
	t.item, t.site, t.count, t.started = item, site, len(selected), time.Now()
	console.Lines(Header(item, site, t.style()))
	console.Lines(EntryList(item.Entries, t.ShowAll, t.style()))
	console.Plain("")
	total := len(item.Entries)
	if total < 2 || selected == nil {
		return
	}
	verb := "Downloading"
	if t.List {
		verb = "Listing"
	}
	if len(selected) == total {
		console.Status(fmt.Sprintf("%s all %d %ss", verb, total, site.Unit))
		return
	}
	nums := ""
	for i, e := range selected {
		if i > 0 {
			nums += ", "
		}
		nums += fmt.Sprint(e.Index)
	}
	console.Status(fmt.Sprintf("%s %d of %d %ss: %s", verb, len(selected), total, site.Unit, nums))
}

// EntryStarted prints [n/count] when the run takes several entries.
func (t *Terminal) EntryStarted(n, count int, entry *media.Entry) {
	t.entryAt = time.Now()
	if count > 1 {
		if n > 1 {
			console.Plain("")
		}
		console.Plain(EntryHeader(n, count, entry.Title, t.style()))
	}
}

// Streams prints the stream table, focused on the chosen codec unless everything is asked for.
func (t *Terminal) Streams(entry *media.Entry, r *engine.EntryResult) {
	if t.HideStreams || t.Interactive && !t.List {
		return
	}
	collapse := !(t.ShowAll || t.List)
	console.Lines(Streams(r.Video, r.Audio, r.VideoHasAudio, r.ChosenVideo, r.ChosenAudio, entry.Duration, collapse, t.URLs, t.style()))
}

// Transfer is a progress line for one download.
func (t *Terminal) Transfer(label string) fetch.Progress { return NewProgress(label) }

// EntryFinished prints the summary of a written file.
func (t *Terminal) EntryFinished(entry *media.Entry, r *engine.EntryResult) {
	if r.Status != engine.Downloaded {
		return
	}
	t.finished++
	if r.File == "" {
		return
	}
	var v *media.VideoFormat
	var a *media.AudioFormat
	if r.ChosenVideo >= 0 {
		v = &r.Video[r.ChosenVideo]
	}
	if r.ChosenAudio >= 0 {
		a = &r.Audio[r.ChosenAudio]
	}
	console.Lines(Summary(r.File, r.Size, v, a, time.Since(t.entryAt).Seconds(), t.style()))
}

// Finished prints the total of a run over several entries.
func (t *Terminal) Finished(r *engine.Result) {
	if t.count > 1 && !t.List {
		console.Plain("")
		console.Success(fmt.Sprintf("All done  %d %ss  ", t.count, t.site.Unit) + t.style().Dim("·  "+ETA(time.Since(t.started).Seconds())))
	}
}

// Choose asks for the video stream, then the audio stream, redrawing one table with ▶ on the cursor.
func (t *Terminal) Choose(entry *media.Entry, r *engine.EntryResult) (int, int, error) {
	s := t.style()
	draw := func(v, a int) []string {
		return Streams(r.Video, r.Audio, r.VideoHasAudio, v, a, entry.Duration, false, false, s)
	}
	video, audio := r.ChosenVideo, r.ChosenAudio
	onScreen := 0
	var err error
	if len(r.Video) > 0 {
		if video, err = console.Choose(len(r.Video), max(video, 0), "Choose the video stream", 0, func(c int) []string { return draw(c, -1) }); err != nil {
			return 0, 0, err
		}
		onScreen = len(draw(video, -1))
	}
	if len(r.Audio) > 0 {
		// Draw over the video frame so both choices live in one table.
		if audio, err = console.Choose(len(r.Audio), max(audio, 0), "Choose the audio stream", onScreen, func(c int) []string { return draw(video, c) }); err != nil {
			return 0, 0, err
		}
	}
	return video, audio, nil
}

// Progress is one transfer's line: redrawn in place on a terminal; elsewhere only its final line is printed.
type Progress struct {
	label   string
	started time.Time
	mu      sync.Mutex
	done    int64
	total   int64
	samples []sample
	frame   int
	over    bool
	stop    chan struct{}
	ticking sync.WaitGroup
}

type sample struct {
	at    time.Time
	bytes int64
}

var spinner = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

// NewProgress starts a progress line.
func NewProgress(label string) *Progress {
	p := &Progress{label: label, started: time.Now(), stop: make(chan struct{})}
	if console.IsTTY() {
		p.ticking.Add(1)
		go p.tick()
	}
	return p
}

func (p *Progress) tick() {
	defer p.ticking.Done()
	t := time.NewTicker(125 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-t.C:
			p.mu.Lock()
			p.frame++
			line := ProgressLine(p.label, p.done, p.total, rate(p.samples), spinner[p.frame%len(spinner)], console.Width(), console.CurrentStyle())
			p.mu.Unlock()
			console.Write("\r" + line + "\x1b[K")
		}
	}
}

// Update records bytes so far and the total (≤ 0 unknown).
func (p *Progress) Update(done, total int64) {
	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.done = done
	if total > 0 {
		p.total = total
	}
	p.samples = append(p.samples, sample{now, done})
	// Keep a three-second window for the rate.
	for len(p.samples) > 2 && now.Sub(p.samples[1].at) > 3*time.Second {
		p.samples = p.samples[1:]
	}
}

// Finish ends the line: ✓ with the size and rate on success, nothing on failure.
func (p *Progress) Finish(ok bool) {
	p.mu.Lock()
	if p.over {
		p.mu.Unlock()
		return
	}
	p.over = true
	done := p.done
	p.mu.Unlock()
	close(p.stop)
	p.ticking.Wait()
	if console.IsTTY() {
		console.Write("\r\x1b[K")
	}
	if ok {
		console.Plain(ProgressDone(p.label, done, time.Since(p.started).Seconds(), console.CurrentStyle()))
	}
}

func rate(s []sample) float64 {
	if len(s) < 2 {
		return 0
	}
	first, last := s[0], s[len(s)-1]
	d := last.at.Sub(first.at).Seconds()
	if d <= 0 {
		return 0
	}
	return float64(last.bytes-first.bytes) / d
}
