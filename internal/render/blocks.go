// Package render is what people see: the item header, the stream table, progress lines, summaries, and the
// arrow-key stream chooser. The blocks are pure functions of their input and a Style, so they are testable and
// degrade to plain text; Terminal wires them to the engine as its Observer and Chooser.
package render

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
)

const indent = console.Indent

// Header is the item's title and a line of facts: who, when, how long, how many, logged in.
func Header(item *media.Item, site extract.Info, s console.Style) []string {
	var meta []string
	owner := item.Uploader.Name
	for _, e := range item.Entries {
		if owner == "" {
			owner = e.Uploader.Name
		}
	}
	if owner != "" {
		meta = append(meta, cmpOr(site.OwnerLabel, "by")+" "+owner)
	}
	if !item.Published.IsZero() {
		meta = append(meta, item.Published.Local().Format("2006-01-02"))
	}
	if len(item.Entries) == 1 && item.Entries[0].Duration > 0 {
		meta = append(meta, format.Duration(int(item.Entries[0].Duration.Seconds()), true))
	} else if len(item.Entries) > 1 {
		var total time.Duration
		known := true
		for _, e := range item.Entries {
			total += e.Duration
			known = known && e.Duration > 0
		}
		if known {
			meta = append(meta, format.Duration(int(total.Seconds()), true)+" total")
		}
		meta = append(meta, fmt.Sprintf("%d %ss", len(item.Entries), cmpOr(site.Unit, "page")))
	}
	if item.LoggedIn != nil {
		if *item.LoggedIn {
			meta = append(meta, s.Green("logged in"))
		} else {
			meta = append(meta, s.Yellow("logged out (lower qualities only)"))
		}
	}
	return []string{indent + s.Bold(item.Title), indent + s.Dim(strings.Join(meta, "  ·  "))}
}

// EntryList lists the entries of an item with several: the first five, or all.
func EntryList(entries []*media.Entry, all bool, s console.Style) []string {
	if len(entries) < 2 {
		return nil
	}
	width := len(strconv.Itoa(len(entries)))
	shown := entries
	if !all && len(shown) > 5 {
		shown = shown[:5]
	}
	var lines []string
	for _, e := range shown {
		d := ""
		if e.Duration > 0 {
			d = "  " + s.Dim(format.Duration(int(e.Duration.Seconds()), true))
		}
		lines = append(lines, indent+s.Dim(fmt.Sprintf("P%0*d", width, e.Index))+"  "+e.Title+d)
	}
	if len(shown) < len(entries) {
		lines = append(lines, indent+s.Dim(fmt.Sprintf("…  %d more (--show-all lists them)", len(entries)-len(shown))))
	}
	return lines
}

// EntryHeader is [03/12] Title.
func EntryHeader(n, count int, title string, s console.Style) string {
	return indent + s.Bold(fmt.Sprintf("[%0*d/%d] %s", len(strconv.Itoa(count)), n, count, title))
}

// FPS is 60fps / 29.4fps.
func FPS(v float64) string {
	if v <= 0 {
		return ""
	}
	if v == math.Round(v) {
		return fmt.Sprintf("%dfps", int(v))
	}
	return fmt.Sprintf("%.1ffps", v)
}

// Resolution is 1920×1080.
func Resolution(w, h int) string {
	if w == 0 || h == 0 {
		return ""
	}
	return fmt.Sprintf("%d×%d", w, h)
}

// Streams is the video and audio tables with ▶ on the chosen rows. collapse hides video codecs other than the
// chosen one behind a one-line summary.
func Streams(video []media.VideoFormat, audio []media.AudioFormat, videoHasAudio bool, chosenVideo, chosenAudio int,
	duration time.Duration, collapse, urls bool, s console.Style) []string {
	var out []string
	if len(video) > 0 {
		out = append(out, indent+s.Dim("Video"))
		keep := ""
		if collapse && chosenVideo >= 0 && chosenVideo < len(video) {
			keep = video[chosenVideo].Codec
		}
		var rows []row
		var hidden []string
		hiddenCount := 0
		for i, v := range video {
			if keep != "" && v.Codec != keep {
				hiddenCount++
				if !slices.Contains(hidden, v.Codec) {
					hidden = append(hidden, v.Codec)
				}
				continue
			}
			rows = append(rows, row{index: i, url: sourceURL(v.Source, v.Parts), cells: []string{v.Quality, Resolution(v.Width, v.Height), v.Codec,
				FPS(v.FPS), kbps(v.Bitrate), size(media.EstimatedSize(v.Size, v.Bitrate, duration))}})
		}
		out = append(out, table(rows, chosenVideo, map[int]bool{4: true, 5: true}, urls, s)...)
		if hiddenCount > 0 {
			slices.Sort(hidden)
			out = append(out, indent+"  "+s.Dim(fmt.Sprintf("…  %d more %s streams (-i or --show-all lists them)", hiddenCount, strings.Join(hidden, " / "))))
		}
	}
	if videoHasAudio {
		out = append(out, indent+s.Dim("Audio  inside the video file"))
	}
	if len(audio) > 0 {
		out = append(out, indent+s.Dim("Audio"))
		var rows []row
		for i, a := range audio {
			rows = append(rows, row{index: i, url: a.Source.URL, cells: []string{a.Codec, kbps(a.Bitrate), size(media.EstimatedSize(a.Size, a.Bitrate, duration))}})
		}
		out = append(out, table(rows, chosenAudio, map[int]bool{1: true, 2: true}, urls, s)...)
	}
	return out
}

func sourceURL(r media.Resource, parts []media.Resource) string {
	if len(parts) > 0 {
		urls := make([]string, len(parts))
		for i, p := range parts {
			urls[i] = p.URL
		}
		return strings.Join(urls, "\n"+indent+"    ")
	}
	return r.URL
}

func kbps(v int64) string {
	if v <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d kbps", v)
}

func size(b int64) string {
	if b <= 0 {
		return "-"
	}
	return format.FileSize(float64(b), 1)
}

type row struct {
	index int
	cells []string
	url   string
}

func table(rows []row, selected int, right map[int]bool, urls bool, s console.Style) []string {
	if len(rows) == 0 {
		return nil
	}
	widths := make([]int, len(rows[0].cells))
	maxIndex := 0
	for _, r := range rows {
		maxIndex = max(maxIndex, r.index)
		for c, cell := range r.cells {
			widths[c] = max(widths[c], console.DisplayWidth(cell))
		}
	}
	iw := len(strconv.Itoa(maxIndex))
	var out []string
	for _, r := range rows {
		cells := make([]string, len(r.cells))
		for c, cell := range r.cells {
			cells[c] = console.Pad(cell, widths[c], right[c])
		}
		body := console.Pad(strconv.Itoa(r.index), iw, true) + "  " + strings.Join(cells, "  ")
		marker := " "
		if r.index == selected {
			marker, body = s.Cyan("▶"), s.BoldCyan(body)
		}
		out = append(out, strings.TrimRight(indent+marker+" "+body, " "))
		if urls && r.url != "" {
			out = append(out, indent+"    "+s.Dim(r.url))
		}
	}
	return out
}

// Bar is a progress bar of width cells.
func Bar(fraction float64, width int, s console.Style) string {
	f := min(1, max(0, fraction))
	filled := int(f * float64(width))
	return s.Cyan(strings.Repeat("█", filled)) + s.Dim(strings.Repeat("░", width-filled))
}

// ETA is 7s / 1m05s / 1h01m, or -- when unknown.
func ETA(seconds float64) string {
	if math.IsInf(seconds, 0) || math.IsNaN(seconds) || seconds < 0 {
		return "--"
	}
	n := int(math.Round(seconds))
	switch {
	case n < 60:
		return fmt.Sprintf("%ds", n)
	case n < 3600:
		return fmt.Sprintf("%dm%02ds", n/60, n%60)
	}
	return fmt.Sprintf("%dh%02dm", n/3600, n%3600/60)
}

// ProgressLine is one transfer in flight: bar, percentage, sizes, rate, time left. Without a total it is a spinner
// and the bytes so far.
func ProgressLine(label string, done, total int64, rate float64, spinner rune, width int, s console.Style) string {
	name := console.Pad(label, 5, false)
	speed := "--"
	if rate > 0 {
		speed = format.FileSize(rate, 1) + "/s"
	}
	if total <= 0 {
		return indent + name + "  " + s.Dim(string(spinner)) + "  " + format.FileSize(float64(done), 1) + "   " + s.Dim(speed)
	}
	fraction := float64(done) / float64(total)
	percent := console.Pad(fmt.Sprintf("%3.0f%%", fraction*100), 4, true)
	tail := "  " + percent + "   " + format.FileSize(float64(done), 1) + " / " + format.FileSize(float64(total), 1) + "   " + speed
	if rate > 0 {
		tail += "   " + ETA(float64(total-done)/rate) + " left"
	}
	bar := min(24, max(8, width-console.DisplayWidth(indent+name+"  "+tail)-2))
	return indent + name + "  " + Bar(fraction, bar, s) + tail
}

// ProgressDone is a finished transfer: ✓, size, average rate, time taken.
func ProgressDone(label string, total int64, seconds float64, s console.Style) string {
	speed := "--"
	if seconds > 0 {
		speed = format.FileSize(float64(total)/seconds, 1) + "/s"
	}
	return indent + console.Pad(label, 5, false) + "  " + s.Green("✓") + "  " + format.FileSize(float64(total), 1) +
		s.Dim("  ·  "+speed+"  ·  "+ETA(seconds))
}

// Summary is the ✓ Done line of a finished entry, with size, streams and time taken.
func Summary(path string, size int64, video *media.VideoFormat, audio *media.AudioFormat, seconds float64, s console.Style) []string {
	parts := []string{format.FileSize(float64(size), 1)}
	var streams []string
	if video != nil {
		streams = append(streams, strings.Join(nonEmpty(video.Quality, video.Codec, FPS(video.FPS)), " "))
	}
	if audio != nil {
		if audio.Bitrate > 0 {
			streams = append(streams, fmt.Sprintf("%s %d kbps", audio.Codec, audio.Bitrate))
		} else {
			streams = append(streams, audio.Codec)
		}
	}
	if len(streams) > 0 {
		parts = append(parts, strings.Join(streams, " + "))
	}
	parts = append(parts, ETA(seconds))
	return []string{s.Green("✓ ") + s.Bold("Done") + "  " + console.PrettyPath(path), indent + "  " + s.Dim(strings.Join(parts, "  ·  "))}
}

func nonEmpty(ss ...string) []string {
	var out []string
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func cmpOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
