package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/ac1982/haul/internal/engine"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/media"
)

// capture runs f with stdout and stderr redirected, and returns what they got.
func capture(t *testing.T, f func()) (stdout, stderr string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	ro, wo, _ := os.Pipe()
	re, we, _ := os.Pipe()
	os.Stdout, os.Stderr = wo, we
	outc, errc := make(chan string), make(chan string)
	go func() { b, _ := io.ReadAll(ro); outc <- string(b) }()
	go func() { b, _ := io.ReadAll(re); errc <- string(b) }()
	f()
	wo.Close()
	we.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	return <-outc, <-errc
}

func TestHelpFitsEightyColumns(t *testing.T) {
	texts := map[string]string{"root": rootHelp(), "templates": templatesHelp()}
	for _, c := range []string{"download", "info", "login"} {
		texts[c] = commandHelp(c, true)
	}
	for name, text := range texts {
		for _, line := range strings.Split(text, "\n") {
			if n := utf8.RuneCountInString(line); n > 80 {
				t.Errorf("%s: %d columns: %q", name, n, line)
			}
		}
	}
	for _, want := range []string{"FOR SCRIPTS AND AI AGENTS", "EXIT CODES", "--json", "bilibili.com (video"} {
		if !strings.Contains(texts["root"], want) {
			t.Errorf("root help lacks %q", want)
		}
	}
	if strings.Contains(commandHelp("download", false), "--upos-host") || !strings.Contains(commandHelp("download", true), "--upos-host") {
		t.Error("hidden options belong to --help-hidden only")
	}
}

func TestExitCodes(t *testing.T) {
	t.Setenv("HAUL_HOME", t.TempDir())
	cases := []struct {
		args []string
		code int
		out  string
	}{
		{[]string{"--help"}, 0, "OVERVIEW"},
		{[]string{"download", "--help"}, 0, "--audio-only"},
		{[]string{"help", "info"}, 0, "--urls"},
		{[]string{"templates"}, 0, "<pageNumberWithZero>"},
		{[]string{"--version"}, 0, "haul "},
		{nil, exitUsage, "OVERVIEW"},
		{[]string{"--bogus", "x"}, exitUsage, ""},
		{[]string{"info"}, exitUsage, ""},
		{[]string{"a", "b"}, exitUsage, ""},
		{[]string{"login", "youtube"}, exitUsage, ""},
		{[]string{"https://example.com/v"}, 2, ""},
		{[]string{"--audio-only", "--video-only", "BV1xx411c7mD"}, 2, ""},
		{[]string{"-p", "x", "--api", "nope", "BV1xx411c7mD"}, 2, ""},
	}
	for _, c := range cases {
		var code int
		out, errOut := capture(t, func() { code = run(context.Background(), c.args) })
		if code != c.code || !strings.Contains(out, c.out) {
			t.Errorf("haul %q: exit %d (want %d), stdout %q, stderr %q", c.args, code, c.code, firstLine(out), firstLine(errOut))
		}
	}
}

func TestJSONOnFailure(t *testing.T) {
	t.Setenv("HAUL_HOME", t.TempDir())
	var code int
	out, errOut := capture(t, func() { code = run(context.Background(), []string{"info", "--json", "https://example.com/v"}) })
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("stdout is not one JSON document: %v\n%s", err, out)
	}
	e := doc["error"].(map[string]any)
	if code != 2 || doc["ok"] != false || e["kind"] != "input" || e["exitCode"] != float64(2) || doc["files"] == nil {
		t.Errorf("doc = %v (exit %d)", doc, code)
	}
	if !strings.Contains(errOut, "Unsupported link") {
		t.Errorf("stderr = %q", errOut)
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

func options(t *testing.T, args ...string) (engine.Options, error) {
	t.Helper()
	fs := newFlagSet("download", downloadFlags)
	if err := fs.Parse(args); err != nil {
		t.Fatal(err)
	}
	return engineOptions(fs, false)
}

func TestContentFlags(t *testing.T) {
	o, _ := options(t)
	if o.Content.Tracks != engine.VideoAndAudio || o.Content.Subtitles != engine.Embed || o.Content.Cover != engine.Embed || o.VideoIndex != -1 {
		t.Errorf("default: %+v", o)
	}
	o, _ = options(t, "--audio-only", "--skip-cover", "--danmaku", "--skip-mux")
	if o.Content.Tracks != engine.AudioOnly || o.Content.Cover != engine.Skip || o.Content.Sidecars[0] != "danmaku" || !o.Content.NoMux {
		t.Errorf("audio-only: %+v", o.Content)
	}
	o, _ = options(t, "--subtitle-only", "--sub-lang", "en, zh", "--auto-subtitles")
	if o.Content.Tracks != engine.NoTracks || o.Content.Subtitles != engine.Files || o.Content.Cover != engine.Skip ||
		strings.Join(o.SubtitleLangs, "|") != "en|zh" || !o.AutoSubtitles {
		t.Errorf("subtitle-only: %+v", o)
	}
	o, _ = options(t, "--cover-only")
	if o.Content.Tracks != engine.NoTracks || o.Content.Cover != engine.Files || o.Content.Subtitles != engine.Skip {
		t.Errorf("cover-only: %+v", o.Content)
	}
	o, _ = options(t, "--danmaku-only")
	if o.Content.Tracks != engine.NoTracks || len(o.Content.Sidecars) != 1 {
		t.Errorf("danmaku-only: %+v", o.Content)
	}
	o, _ = options(t, "-q", "4K, 1080P+", "-c", "hevc，m4a", "--video-stream", "3", "--delay", "2", "-w", "~/x", "-p", "1-3")
	if strings.Join(o.Quality, "|") != "4K|1080P+" || strings.Join(o.Codec, "|") != "hevc|m4a" || o.VideoIndex != 3 || o.Delay != 2*time.Second ||
		o.Dir != "~/x" || o.Pages != "1-3" {
		t.Errorf("options: %+v", o)
	}
	if _, err := options(t, "--subtitle-only", "--skip-subtitle"); !errs.Is(err, errs.Input) {
		t.Errorf("conflict: %v", err)
	}
}

func TestConfigFillsWhatTheCommandLineLeaves(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HAUL_HOME", dir)
	cfg := `{"workDir": "/movies", "codec": "avc", "archive": true, "delay": 2, "danmaku": true,
	  "bilibili": {"api": "tv", "danmakuFormat": ["ass"], "noForceHttp": true}, "nonsense": 1}`
	os.WriteFile(filepath.Join(dir, "config.json"), []byte(cfg), 0o644)
	fs := newFlagSet("download", downloadFlags)
	fs.Parse([]string{"-c", "hevc", "--delay", "5"})
	var err error
	_, errOut := capture(t, func() { err = applyConfig(fs, downloadFlags) })
	if err != nil {
		t.Fatal(err)
	}
	checks := map[string]string{"work-dir": "/movies", "codec": "hevc", "api": "tv", "danmaku-format": "ass", "delay": "5",
		"archive": "true", "no-force-http": "true", "danmaku": "false"}
	for name, want := range checks {
		if got := fs.Lookup(name).Value.String(); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	if !strings.Contains(errOut, `Unknown key "danmaku"`) || !strings.Contains(errOut, `Unknown key "nonsense"`) {
		t.Errorf("unknown keys are warned about: %q", errOut)
	}

	fs = newFlagSet("download", downloadFlags)
	fs.Parse([]string{"--config", filepath.Join(dir, "missing.json")})
	if err := applyConfig(fs, downloadFlags); !errs.Is(err, errs.Input) {
		t.Errorf("a missing --config is an input error: %v", err)
	}
	os.WriteFile(filepath.Join(dir, "config.json"), []byte("{"), 0o644)
	fs = newFlagSet("download", downloadFlags)
	if err := applyConfig(fs, downloadFlags); !errs.Is(err, errs.Input) {
		t.Errorf("bad JSON: %v", err)
	}
}

func TestDocument(t *testing.T) {
	yes := true
	entry := &media.Entry{Index: 1, ID: "BV1", Title: "P1", Duration: 113490 * time.Millisecond, Published: time.Unix(1700000000, 0)}
	item := &media.Item{Site: "bilibili", Title: "T", Uploader: media.Person{Name: "UP"}, LoggedIn: &yes, Entries: []*media.Entry{entry, {Index: 2, ID: "BV2", Title: "P2"}}}
	r := &engine.Result{Item: item, Entries: []*engine.EntryResult{
		{Entry: entry, Selected: true, Status: engine.Downloaded, File: "/o/T/[P1]P1.mp4", Size: 10, ChosenVideo: 0, ChosenAudio: -1,
			Video: []media.VideoFormat{{Quality: "1080P", Width: 1920, Height: 1080, FPS: 29.412, Codec: "AVC", Bitrate: 1000, HasAudio: true,
				Source: media.Resource{URL: "https://v"}}}, VideoHasAudio: true, Subtitles: []media.Subtitle{{Lang: "zh-CN", Auto: true}}},
		{Entry: item.Entries[1], ChosenVideo: -1, ChosenAudio: -1},
	}}
	out := encodeDocument(document("download", "BV1", r, nil, false))
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatal(err)
	}
	p := doc["pages"].([]any)[0].(map[string]any)
	v := p["video"].([]any)[0].(map[string]any)
	if doc["ok"] != true || doc["loggedIn"] != true || doc["pageCount"] != float64(2) || doc["files"].([]any)[0] != "/o/T/[P1]P1.mp4" ||
		p["durationSeconds"] != float64(113) || p["selectedVideo"] != float64(0) || p["selectedAudio"] != nil ||
		v["resolution"] != "1920x1080" || v["fps"] != 29.412 || v["url"] != nil || p["published"] != "2023-11-14T22:13:20Z" {
		t.Errorf("doc:\n%s", out)
	}
	if sub := p["subtitles"].([]any)[0].(map[string]any); sub["lang"] != "zh-CN" || sub["auto"] != true {
		t.Errorf("subtitles: %v", sub)
	}
	second := doc["pages"].([]any)[1].(map[string]any)
	if second["selected"] != false || second["status"] != nil || second["subtitles"] != nil {
		t.Errorf("an untaken page: %v", second)
	}
	if !strings.Contains(out, `"files": [`) || strings.Contains(out, `<`) {
		t.Errorf("encoding: %s", out)
	}

	withURLs := encodeDocument(document("info", "BV1", r, nil, true))
	if !strings.Contains(withURLs, `"url": "https://v"`) {
		t.Errorf("--urls: %s", withURLs)
	}
	failed := encodeDocument(document("download", "x", r, errs.NewAuth("login expired"), false))
	if !strings.Contains(failed, `"kind": "auth"`) || !strings.Contains(failed, `"exitCode": 4`) || !strings.Contains(failed, `"pages"`) {
		t.Errorf("failure keeps the pages: %s", failed)
	}
	bare := encodeDocument(document("download", "x", nil, errors.New("boom"), false))
	if !strings.Contains(bare, `"kind": "failed"`) || strings.Contains(bare, "pages") {
		t.Errorf("bare failure: %s", bare)
	}
}

func TestAppearsFirst(t *testing.T) {
	c, q := []string{"-c", "--codec"}, []string{"-q", "--quality"}
	if !appearsFirst([]string{"haul", "--codec=avc", "-q", "720p"}, c, q) || appearsFirst([]string{"-q", "x", "-c", "y"}, c, q) ||
		appearsFirst([]string{"-c", "y"}, c, q) {
		t.Error("appearsFirst")
	}
}
