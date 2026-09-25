package shell

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindAndResolve(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	dir := t.TempDir()
	tool := filepath.Join(dir, "haul-test-tool")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\necho out; echo err >&2; exit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	if FindExecutable("haul-test-tool") != tool {
		t.Errorf("FindExecutable = %q", FindExecutable("haul-test-tool"))
	}
	if Resolve(tool) != tool || Resolve(filepath.Join(dir, "missing")) != "" || FindExecutable("haul-no-such-tool") != "" {
		t.Error("Resolve")
	}
	res, err := Run(context.Background(), tool, nil, Options{Capture: true})
	if err != nil || res.Status != 3 || res.Output != "out\n" || res.Errors != "err\n" {
		t.Errorf("Run = %+v, %v", res, err)
	}
	if _, err := Run(context.Background(), filepath.Join(dir, "missing"), nil, Options{}); err == nil {
		t.Error("a missing tool is an error")
	}
}

func TestInstallHint(t *testing.T) {
	got := InstallHint("yt-dlp deno")
	switch runtime.GOOS {
	case "darwin":
		if got != "brew install yt-dlp deno" {
			t.Errorf("%q", got)
		}
	case "windows":
		if got != "winget install yt-dlp.yt-dlp; winget install DenoLand.Deno" {
			t.Errorf("%q", got)
		}
	default:
		binary := "yt-dlp_linux"
		if runtime.GOARCH == "arm64" {
			binary = "yt-dlp_linux_aarch64"
		}
		want := `pipx install "yt-dlp[default]", or ` + binary + ` from https://github.com/yt-dlp/yt-dlp/releases/latest ` +
			`(distribution packages are often too old for YouTube); deno: curl -fsSL https://deno.land/install.sh | sh -s -- -y`
		if got != want {
			t.Errorf("%q", got)
		}
		if got := InstallHint("ffmpeg deno"); got != "install ffmpeg with your package manager (apt, dnf, pacman…); "+
			"deno: curl -fsSL https://deno.land/install.sh | sh -s -- -y" {
			t.Errorf("%q", got)
		}
		if got := InstallHint("deno"); got != "curl -fsSL https://deno.land/install.sh | sh -s -- -y" {
			t.Errorf("%q", got)
		}
	}
}
