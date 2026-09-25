package testkit

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/shell"
)

// FFmpeg and FFprobe are the tools on PATH, or "".
var (
	FFmpeg  = shell.FindExecutable("ffmpeg")
	FFprobe = shell.FindExecutable("ffprobe")
)

// RequireFFmpeg skips the test when ffmpeg or ffprobe is missing.
func RequireFFmpeg(t testing.TB) {
	t.Helper()
	if FFmpeg == "" || FFprobe == "" {
		t.Skip("needs ffmpeg and ffprobe")
	}
}

// Media is small real media: H.264 only, AAC only, both in one file, a JPEG, and an MP3 carrying its own picture.
type Media struct {
	Video, Audio, Combined, Cover, MP3 []byte
}

var (
	mediaOnce sync.Once
	mediaVal  Media
	mediaErr  error
)

// MakeMedia builds the files once per test binary.
func MakeMedia(t testing.TB) Media {
	t.Helper()
	RequireFFmpeg(t)
	mediaOnce.Do(func() {
		dir, err := os.MkdirTemp("", "haul-media-")
		if err != nil {
			mediaErr = err
			return
		}
		run := func(args ...string) {
			if mediaErr != nil {
				return
			}
			res, err := shell.Run(context.Background(), FFmpeg, append([]string{"-v", "error", "-y"}, args...), shell.Options{Capture: true})
			if err != nil || res.Status != 0 {
				mediaErr = &ffmpegError{args: args, out: res.Errors, err: err}
			}
		}
		v := []string{"-f", "lavfi", "-i", "testsrc=size=64x64:rate=10", "-t", "1"}
		a := []string{"-f", "lavfi", "-i", "sine=frequency=440:duration=1"}
		run(append(v, "-c:v", "libx264", "-pix_fmt", "yuv420p", filepath.Join(dir, "v.mp4"))...)
		run(append(a, "-c:a", "aac", filepath.Join(dir, "a.m4a"))...)
		run(append(append(v, a...), "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", filepath.Join(dir, "av.mp4"))...)
		run("-f", "lavfi", "-i", "color=red:size=32x32", "-frames:v", "1", filepath.Join(dir, "c.jpg"))
		run("-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-f", "lavfi", "-i", "color=blue:size=16x16", "-frames:v", "1",
			"-map", "0", "-map", "1", "-c:a", "libmp3lame", "-c:v", "png", "-disposition:v:0", "attached_pic", filepath.Join(dir, "a.mp3"))
		read := func(name string) []byte {
			b, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil && mediaErr == nil {
				mediaErr = err
			}
			return b
		}
		mediaVal = Media{Video: read("v.mp4"), Audio: read("a.m4a"), Combined: read("av.mp4"), Cover: read("c.jpg"), MP3: read("a.mp3")}
	})
	if mediaErr != nil {
		t.Fatalf("making test media: %v", mediaErr)
	}
	return mediaVal
}

type ffmpegError struct {
	args []string
	out  string
	err  error
}

func (e *ffmpegError) Error() string { return "ffmpeg " + e.out }

// Probe is what ffprobe sees in a file.
type Probe struct {
	// Streams are "video:h264", "audio:aac"…, without the data track ffmpeg adds for chapters.
	Streams  []string
	Chapters []string
	Tags     map[string]string
	// Languages are the language tags of the subtitle streams.
	Languages []string
}

// ProbeFile runs ffprobe.
func ProbeFile(t testing.TB, path string) Probe {
	t.Helper()
	RequireFFmpeg(t)
	res, err := shell.Run(context.Background(), FFprobe, []string{"-v", "error", "-show_entries",
		"stream=codec_type,codec_name:stream_tags=language:format_tags", "-show_chapters", "-of", "json", path}, shell.Options{Capture: true})
	if err != nil || res.Status != 0 {
		t.Fatalf("ffprobe %s: %v %s", path, err, res.Errors)
	}
	j, err := jsonv.ParseString(res.Output)
	if err != nil {
		t.Fatal(err)
	}
	var p Probe
	for _, s := range j.Get("streams").Array() {
		kind := s.Get("codec_type").String()
		if kind == "data" {
			continue
		}
		p.Streams = append(p.Streams, kind+":"+s.Get("codec_name").String())
		if kind == "subtitle" {
			p.Languages = append(p.Languages, s.Path("tags", "language").String())
		}
	}
	for _, c := range j.Get("chapters").Array() {
		p.Chapters = append(p.Chapters, c.Path("tags", "title").String())
	}
	p.Tags = map[string]string{}
	for k, v := range j.Path("format", "tags").Object() {
		p.Tags[k] = v.String()
	}
	return p
}
