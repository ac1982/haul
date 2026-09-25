package ytdlp

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/media"
)

// fakeYtDlp writes a stand-in for yt-dlp that records its arguments, prints stdout and stderr, and exits with status.
func fakeYtDlp(t *testing.T, stdout, stderr string, status int) (tool, argsFile string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	dir := t.TempDir()
	for name, text := range map[string]string{"out": stdout, "err": stderr} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	argsFile = filepath.Join(dir, "args")
	script := "#!/bin/sh\n" +
		`for a in "$@"; do echo "$a"; done > '` + argsFile + "'\n" +
		"cat '" + filepath.Join(dir, "out") + "'\n" +
		"cat '" + filepath.Join(dir, "err") + "' >&2\n" +
		"exit " + strconv.Itoa(status) + "\n"
	tool = filepath.Join(dir, "yt-dlp")
	if err := os.WriteFile(tool, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return tool, argsFile
}

func TestResolveRunsYtDlp(t *testing.T) {
	t.Parallel()
	tool, argsFile := fakeYtDlp(t, xPost, "WARNING: something\n", 0)
	x := NewX(nil, Options{Path: tool})
	item, err := x.Resolve(context.Background(), "https://x.com/CTVJLaidlaw/status/1600649710662213632")
	if err != nil {
		t.Fatal(err)
	}
	if item.Site != "x" || len(item.Entries) != 2 || item.Entries[1].Title != "Video 2" {
		t.Errorf("item = %+v", item)
	}
	args, _ := os.ReadFile(argsFile)
	want := "-J\n--no-playlist\n--no-progress\n--\nhttps://x.com/CTVJLaidlaw/status/1600649710662213632\n"
	if string(args) != want {
		t.Errorf("args = %q", args)
	}
	// Entries carry their formats; Formats hands them back.
	f, err := x.Formats(context.Background(), item, item.Entries[0])
	if err != nil || f != item.Entries[0].Formats {
		t.Errorf("Formats = %v, %v", f, err)
	}
	if _, err := x.Formats(context.Background(), item, &media.Entry{ID: "1"}); err == nil {
		t.Error("an entry without formats is an error")
	}
}

func TestResolveReportsYtDlpErrors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		stdout, stderr string
		status         int
		want           string
	}{
		{"", "WARNING: [youtube] slow\nERROR: [youtube] abc: Video unavailable\n", 1, "yt-dlp failed: Video unavailable"},
		{"", "ERROR: Unsupported URL: https://x.test\n", 1, "yt-dlp failed: Unsupported URL: https://x.test"},
		{"", "something odd\n", 2, "yt-dlp failed: exit code 2"},
		// Exit 0 without an answer is a failure too.
		{"", "", 0, "yt-dlp failed: exit code 0"},
	}
	for _, c := range cases {
		tool, _ := fakeYtDlp(t, c.stdout, c.stderr, c.status)
		_, err := NewYouTube(nil, Options{Path: tool}).Resolve(context.Background(), "https://www.youtube.com/watch?v=DdCEmlAydcw")
		if err == nil || err.Error() != c.want || !errs.Is(err, errs.Failed) {
			t.Errorf("stderr %q: err = %v, want %q", c.stderr, err, c.want)
		}
	}
}

func TestMissingYtDlpIsADependencyError(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "yt-dlp")
	for _, c := range []struct {
		ex   *Extractor
		want string
	}{
		{NewYouTube(nil, Options{Path: missing}), "YouTube links need yt-dlp: brew install yt-dlp deno"},
		{NewX(nil, Options{Path: missing}), "X links need yt-dlp: brew install yt-dlp"},
	} {
		_, err := c.ex.Resolve(context.Background(), "u")
		if !errs.Is(err, errs.Dependency) || err.Error() != c.want {
			t.Errorf("err = %v", err)
		}
	}
}

func TestFailureReason(t *testing.T) {
	t.Parallel()
	lines := strings.Split("ERROR: [youtube:tab] abc: first\nERROR: [generic] https://a: last", "\n")
	if got := failureReason(lines, 1); got != "last" {
		t.Errorf("failureReason = %q", got)
	}
}
