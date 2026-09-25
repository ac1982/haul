package mux

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/testkit"
)

// values is the value after each occurrence of flag.
func values(args []string, flag string) []string {
	var out []string
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			out = append(out, args[i+1])
		}
	}
	return out
}

func baseJob(dir string) Job {
	return Job{
		Video: dir + "/v.mp4", Audio: dir + "/a.m4a", Output: dir + "/out/T.mp4", Container: MP4,
		Tags: Tags{Title: "T", Artist: "UP", Description: "D", Comment: "https://www.bilibili.com/video/BV1qt4y1X7TW/",
			Date: time.Unix(1595684326, 0)},
	}
}

func TestFullJob(t *testing.T) {
	j := baseJob("/w")
	j.Cover = "/w/c.jpg"
	j.Subtitles = []Subtitle{{Path: "/w/en.srt", Code: "eng", Name: "English"}}
	j.Chapters = []media.Chapter{{Title: "A", End: 10 * time.Second}}
	j.HEVC = true
	args := Arguments(j, "/w/chapters", false)
	check := func(flag string, want ...string) {
		t.Helper()
		if got := values(args, flag); !reflect.DeepEqual(got, want) {
			t.Errorf("%s = %q, want %q", flag, got, want)
		}
	}
	check("-i", "/w/v.mp4", "/w/a.m4a", "/w/c.jpg", "/w/en.srt", "/w/chapters")
	check("-map", "0", "1", "2", "3")
	check("-map_chapters", "4")
	check("-disposition:v:1", "attached_pic")
	check("-metadata:s:s:0", "title=English", "language=eng")
	check("-c:s", "mov_text")
	check("-tag:v:0", "hvc1")
	meta := values(args, "-metadata")
	for _, want := range []string{"title=T", "artist=UP", "description=D", "comment=https://www.bilibili.com/video/BV1qt4y1X7TW/",
		"creation_time=2020-07-25T13:38:46.000000Z"} {
		if !slices.Contains(meta, want) {
			t.Errorf("missing -metadata %s in %q", want, meta)
		}
	}
	if !reflect.DeepEqual(args[len(args)-2:], []string{"--", "/w/out/T.mp4"}) || slices.Contains(args, "-an") {
		t.Errorf("tail: %q", args[len(args)-4:])
	}
}

func TestAudioOnlyTakesTheAudioOfTheFirstInput(t *testing.T) {
	// A video with its audio inside (X): only its audio, so the cover is video stream 0.
	j := baseJob("/w")
	j.Audio = ""
	j.AudioOnly = true
	j.Cover = "/w/c.jpg"
	args := Arguments(j, "", false)
	if got := values(args, "-i"); !reflect.DeepEqual(got, []string{"/w/v.mp4", "/w/c.jpg"}) {
		t.Errorf("-i = %q", got)
	}
	if got := values(args, "-map"); !reflect.DeepEqual(got, []string{"0:a", "1"}) {
		t.Errorf("-map = %q", got)
	}
	if got := values(args, "-disposition:v:0"); !reflect.DeepEqual(got, []string{"attached_pic"}) {
		t.Errorf("disposition = %q", got)
	}
}

func TestPodcastMP3StaysMP3(t *testing.T) {
	j := Job{Audio: "/w/a.mp3", AudioOnly: true, Cover: "/w/c.jpg", Container: MP3, Output: "/w/out/E.mp3",
		Tags: Tags{Title: "E", Album: "T", Artist: "Host", Date: time.Unix(1595684326, 0)}}
	args := Arguments(j, "", false)
	if got := values(args, "-map"); !reflect.DeepEqual(got, []string{"0:a", "1"}) {
		t.Errorf("-map = %q (the MP3's own picture is left behind)", got)
	}
	if got := values(args, "-metadata:s:v:0"); !reflect.DeepEqual(got, []string{"comment=Cover (front)"}) {
		t.Errorf("cover comment = %q", got)
	}
	meta := values(args, "-metadata")
	if !slices.Contains(meta, "album=T") || !slices.ContainsFunc(meta, func(s string) bool { return strings.HasPrefix(s, "date=2020-07-2") }) {
		t.Errorf("meta = %q", meta)
	}
	if !reflect.DeepEqual(values(args, "-f"), []string{"mp3"}) || slices.Contains(args, "-movflags") {
		t.Errorf("args = %q", args)
	}
}

func TestVideoOnlyAndNoTags(t *testing.T) {
	j := baseJob("/w")
	j.VideoOnly = true
	j.Audio = ""
	j.Tags = Tags{}
	args := Arguments(j, "", false)
	if !slices.Contains(args, "-an") || len(values(args, "-metadata")) != 0 {
		t.Errorf("args = %q", args)
	}
}

func TestExtraAudioTracks(t *testing.T) {
	j := baseJob("/w")
	j.ExtraAudio = []Track{{Path: "/w/bg.m4a", Title: "Background audio"}, {Path: "/w/r.m4a", Title: "角色", Artist: "配音演员"}}
	args := Arguments(j, "", false)
	if got := values(args, "-i"); !reflect.DeepEqual(got, []string{"/w/v.mp4", "/w/a.m4a", "/w/bg.m4a", "/w/r.m4a"}) {
		t.Errorf("-i = %q", got)
	}
	if got := values(args, "-metadata:s:a:0"); !reflect.DeepEqual(got, []string{"title=Original audio"}) {
		t.Errorf("a:0 = %q", got)
	}
	if got := values(args, "-metadata:s:a:2"); !reflect.DeepEqual(got, []string{"title=角色", "artist=配音演员"}) {
		t.Errorf("a:2 = %q", got)
	}
}

func TestChapterFiles(t *testing.T) {
	c := []media.Chapter{{Title: "Intro", End: 12500 * time.Millisecond}, {Title: "Main", Start: 12500 * time.Millisecond, End: time.Minute}}
	if got := FFmetadata(c); !strings.Contains(got, "START=12500\nEND=60000\ntitle=Main") {
		t.Errorf("FFmetadata = %q", got)
	}
	if got := MP4BoxChapters(c); got != "00:00:00 Intro\n00:00:12 Main\n" {
		t.Errorf("MP4BoxChapters = %q", got)
	}
	args := MP4BoxArguments(Job{Video: "v", Audio: "a", Output: "o.mp4", Tags: Tags{Title: "a:b"}}, "", false)
	if got := strings.Join(args, " "); got != "-inter 500 -noprog -add v#video:name= -add a#audio:lang=und -itags title=a\\:b -new o.mp4" {
		t.Errorf("MP4Box args = %q", got)
	}
}

func TestAudioContainer(t *testing.T) {
	if AudioContainer("MP3") != MP3 || AudioContainer("M4A") != M4A || AudioContainer("OPUS") != M4A || AudioContainer("OGG") != M4A {
		t.Error("AudioContainer")
	}
}

func write(t *testing.T, path string, data []byte) string {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFFmpegEndToEnd(t *testing.T) {
	m := testkit.MakeMedia(t)
	dir := t.TempDir()
	sub := write(t, filepath.Join(dir, "en.srt"), []byte("1\n00:00:00,000 --> 00:00:00,900\nHello\n\n"))
	job := Job{
		Video: write(t, filepath.Join(dir, "v.mp4"), m.Video), Audio: write(t, filepath.Join(dir, "a.m4a"), m.Audio),
		Cover: write(t, filepath.Join(dir, "c.jpg"), m.Cover), Subtitles: []Subtitle{{Path: sub, Code: "eng", Name: "English"}},
		Chapters: []media.Chapter{{Title: "Intro", End: time.Second}}, Container: MP4, Output: filepath.Join(dir, "out", "T.mp4"),
		Tags: Tags{Title: "T", Artist: "UP", Comment: "https://example.test/v"},
	}
	os.MkdirAll(filepath.Dir(job.Output), 0o755)
	if err := (FFmpeg{Path: testkit.FFmpeg}).Mux(context.Background(), job); err != nil {
		t.Fatal(err)
	}
	p := testkit.ProbeFile(t, job.Output)
	if !reflect.DeepEqual(p.Streams, []string{"video:h264", "audio:aac", "subtitle:mov_text", "video:mjpeg"}) {
		t.Errorf("streams = %q", p.Streams)
	}
	if !reflect.DeepEqual(p.Chapters, []string{"Intro"}) || p.Tags["title"] != "T" || p.Tags["artist"] != "UP" || !reflect.DeepEqual(p.Languages, []string{"eng"}) {
		t.Errorf("probe = %+v", p)
	}

	mp3 := Job{Audio: write(t, filepath.Join(dir, "a.mp3"), m.MP3), AudioOnly: true, Cover: job.Cover, Container: MP3,
		Output: filepath.Join(dir, "out", "E.mp3"), Tags: Tags{Title: "E", Album: "Show"}}
	if err := (FFmpeg{Path: testkit.FFmpeg}).Mux(context.Background(), mp3); err != nil {
		t.Fatal(err)
	}
	if p := testkit.ProbeFile(t, mp3.Output); !reflect.DeepEqual(p.Streams, []string{"audio:mp3", "video:mjpeg"}) || p.Tags["album"] != "Show" {
		t.Errorf("mp3 probe = %+v", p)
	}
	if !SupportsDolbyVision(context.Background(), testkit.FFmpeg) {
		t.Error("a current ffmpeg supports Dolby Vision")
	}
}

func TestFailedMuxIsAnError(t *testing.T) {
	testkit.RequireFFmpeg(t)
	dir := t.TempDir()
	job := Job{Video: filepath.Join(dir, "missing.mp4"), Container: MP4, Output: filepath.Join(dir, "o.mp4")}
	if err := (FFmpeg{Path: testkit.FFmpeg}).Mux(context.Background(), job); err == nil || !strings.Contains(err.Error(), "exit code") {
		t.Errorf("err = %v", err)
	}
}
