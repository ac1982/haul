// Package mux puts downloaded tracks, subtitles, chapters, a cover and tags into one file with ffmpeg (or MP4Box).
// Arguments go through argv, so titles with quotes or backslashes need no escaping.
package mux

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/fetch"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/shell"
)

// Container is what a file is written as.
type Container string

const (
	MP4 Container = "mp4"
	M4A Container = "m4a"
	MP3 Container = "mp3"
)

// AudioContainer is the container audio alone goes into: MP3 stays MP3, everything else becomes M4A.
func AudioContainer(codec string) Container {
	if strings.EqualFold(codec, "MP3") {
		return MP3
	}
	return M4A
}

// Track is an extra audio track and its tags.
type Track struct {
	Path   string
	Title  string
	Artist string
}

// Subtitle is a subtitle file and how its stream is tagged.
type Subtitle struct {
	Path string
	Code string // ISO 639-2
	Name string
}

// Tags are the metadata written into the file; a zero Tags writes none.
type Tags struct {
	Title       string
	Album       string
	Artist      string
	Comment     string
	Description string
	Date        time.Time
}

// Job is one output file.
type Job struct {
	// Video may carry its own audio (X, FLV).
	Video string
	Audio string
	// AudioOnly takes only the audio of the first input, so the cover stays the only picture (a video with its
	// audio inside, or a podcast MP3 that carries its own picture).
	AudioOnly bool
	// VideoOnly drops every audio stream.
	VideoOnly     bool
	ExtraAudio    []Track
	Subtitles     []Subtitle
	Cover         string
	Chapters      []media.Chapter
	AudioLanguage string
	Tags          Tags
	// HEVC is tagged hvc1, which QuickTime needs.
	HEVC      bool
	Container Container
	Output    string
}

// Muxer writes a Job.
type Muxer interface {
	Mux(ctx context.Context, job Job) error
	Name() string
}

// FFmpeg muxes with ffmpeg.
type FFmpeg struct{ Path string }

// Name is "ffmpeg".
func (FFmpeg) Name() string { return "ffmpeg" }

// Mux runs ffmpeg for the job.
func (f FFmpeg) Mux(ctx context.Context, job Job) error {
	chapters := ""
	if len(job.Chapters) > 0 {
		tmp, err := os.CreateTemp("", "haul-chapters-*.txt")
		if err != nil {
			return err
		}
		chapters = tmp.Name()
		defer os.Remove(chapters)
		_, err = tmp.WriteString(FFmetadata(job.Chapters))
		tmp.Close()
		if err != nil {
			return err
		}
	}
	args := Arguments(job, chapters, console.Debug())
	console.Debugf("ffmpeg %s", strings.Join(args, " "))
	res, err := shell.Run(ctx, f.Path, args, shell.Options{Echo: true})
	if err != nil {
		return err
	}
	if res.Status != 0 {
		os.Remove(job.Output)
		return errs.New("Muxing failed (ffmpeg exit code %d); --debug shows its output", res.Status)
	}
	return nil
}

// Arguments is the ffmpeg command line for a job, reading chapters from chaptersFile when there are any.
func Arguments(job Job, chaptersFile string, verbose bool) []string {
	var inputs, meta []string
	count := 0
	add := func(path string) {
		inputs = append(inputs, "-i", path)
		count++
	}
	for _, p := range []string{job.Video, job.Audio} {
		if p != "" {
			add(p)
		}
	}
	if len(job.ExtraAudio) > 0 {
		meta = append(meta, "-metadata:s:a:0", "title=Original audio")
		for i, t := range job.ExtraAudio {
			add(t.Path)
			if strings.TrimSpace(t.Title) != "" {
				meta = append(meta, fmt.Sprintf("-metadata:s:a:%d", i+1), "title="+t.Title)
			}
			if strings.TrimSpace(t.Artist) != "" {
				meta = append(meta, fmt.Sprintf("-metadata:s:a:%d", i+1), "artist="+t.Artist)
			}
		}
	}
	if job.Cover != "" {
		add(job.Cover)
	}
	for i, s := range job.Subtitles {
		add(s.Path)
		meta = append(meta, fmt.Sprintf("-metadata:s:s:%d", i), "title="+s.Name, fmt.Sprintf("-metadata:s:s:%d", i), "language="+s.Code)
	}
	if job.Cover != "" {
		picture := 1
		if job.AudioOnly || job.Video == "" {
			picture = 0
		}
		meta = append(meta, fmt.Sprintf("-disposition:v:%d", picture), "attached_pic")
		// ID3 pictures carry a type; players look for the front cover.
		if job.Container == MP3 {
			meta = append(meta, "-metadata:s:v:0", "comment=Cover (front)")
		}
	}
	if chaptersFile != "" {
		inputs = append(inputs, "-i", chaptersFile, "-map_chapters", strconv.Itoa(count))
	}
	for i := range count {
		if i == 0 && job.AudioOnly {
			inputs = append(inputs, "-map", "0:a")
		} else {
			inputs = append(inputs, "-map", strconv.Itoa(i))
		}
	}

	level := "warning"
	if verbose {
		level = "verbose"
	}
	args := append([]string{"-loglevel", level, "-y"}, inputs...)
	args = append(args, meta...)
	t := job.Tags
	tag := func(k, v string) {
		if strings.TrimSpace(v) != "" {
			args = append(args, "-metadata", k+"="+v)
		}
	}
	tag("title", t.Title)
	tag("album", t.Album)
	tag("artist", t.Artist)
	tag("comment", t.Comment)
	tag("description", t.Description)
	if !t.Date.IsZero() {
		tag("creation_time", format.FFmpegCreationTime(t.Date.Unix()))
		if job.Container == MP3 {
			tag("date", t.Date.Local().Format("2006-01-02"))
		}
	}
	if job.AudioLanguage != "" {
		args = append(args, "-metadata:s:a:0", "language="+job.AudioLanguage)
	}
	args = append(args, "-c:v", "copy", "-c:a", "copy")
	if job.Container == MP3 {
		return append(args, "-f", "mp3", "--", job.Output)
	}
	if job.VideoOnly {
		args = append(args, "-an")
	}
	if len(job.Subtitles) > 0 {
		args = append(args, "-c:s", "mov_text")
	}
	if job.HEVC {
		args = append(args, "-tag:v:0", "hvc1")
	}
	return append(args, "-movflags", "faststart", "-strict", "unofficial", "-strict", "-2", "-f", "mp4", "--", job.Output)
}

// FFmetadata is the chapter list in ffmpeg's metadata format.
func FFmetadata(chapters []media.Chapter) string {
	var b strings.Builder
	b.WriteString(";FFMETADATA\n")
	for _, c := range chapters {
		fmt.Fprintf(&b, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=%d\nEND=%d\ntitle=%s\n\n", c.Start.Milliseconds(), c.End.Milliseconds(), c.Title)
	}
	return b.String()
}

// ConcatFLV joins FLV segments into one MP4: each is remuxed to MPEG-TS, the TS files are concatenated.
func ConcatFLV(ctx context.Context, ffmpeg string, parts []string, dst string) error {
	if len(parts) == 1 {
		return os.Rename(parts[0], dst)
	}
	var ts []string
	for _, p := range parts {
		out := strings.TrimSuffix(p, ".flv") + ".ts"
		args := []string{"-loglevel", "warning", "-y", "-i", p, "-map", "0", "-c", "copy", "-f", "mpegts", "-bsf:v", "h264_mp4toannexb", out}
		console.Debugf("ffmpeg %s", strings.Join(args, " "))
		res, err := shell.Run(ctx, ffmpeg, args, shell.Options{Echo: true})
		if err != nil {
			return err
		}
		if res.Status != 0 {
			return errs.New("Joining FLV segments failed (ffmpeg exit code %d)", res.Status)
		}
		os.Remove(p)
		ts = append(ts, out)
	}
	if err := fetch.Join(ts, dst); err != nil {
		return err
	}
	for _, f := range ts {
		os.Remove(f)
	}
	return nil
}

var libavutil = regexp.MustCompile(`libavutil\s+(\d+)\.\s*(\d+)\.`)

// SupportsDolbyVision: Dolby Vision needs libavutil ≥ 57.17 (ffmpeg 5.0).
func SupportsDolbyVision(ctx context.Context, ffmpeg string) bool {
	res, err := shell.Run(ctx, ffmpeg, []string{"-version"}, shell.Options{Capture: true})
	if err != nil {
		return false
	}
	m := libavutil.FindStringSubmatch(res.Output)
	if m == nil {
		return false
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	return major > 57 || major == 57 && minor >= 17
}
