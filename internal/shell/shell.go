// Package shell finds and runs the external tools haul drives: ffmpeg, MP4Box, aria2c, yt-dlp.
package shell

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ac1982/haul/internal/console"
)

// FindExecutable looks in the current directory, next to haul itself, then on PATH.
func FindExecutable(name string) string {
	var dirs []string
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, wd)
	}
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		dirs = append(dirs, filepath.Dir(exe))
	}
	dirs = append(dirs, filepath.SplitList(os.Getenv("PATH"))...)
	for _, dir := range dirs {
		for _, candidate := range candidates(filepath.Join(dir, name)) {
			if isExecutable(candidate) {
				return candidate
			}
		}
	}
	return ""
}

// Resolve turns a user-given path or bare name into an executable path, or "".
func Resolve(pathOrName string) string {
	if strings.ContainsAny(pathOrName, `/\`) {
		for _, c := range candidates(pathOrName) {
			if isExecutable(c) {
				return c
			}
		}
		return ""
	}
	return FindExecutable(pathOrName)
}

func candidates(p string) []string {
	if runtime.GOOS != "windows" || filepath.Ext(p) != "" {
		return []string{p}
	}
	return []string{p + ".exe", p + ".cmd", p + ".bat", p}
}

func isExecutable(p string) bool {
	info, err := os.Stat(p)
	if err != nil || info.IsDir() {
		return false
	}
	return runtime.GOOS == "windows" || info.Mode()&0o111 != 0
}

// Result is how a tool ended and, when captured, what it said.
type Result struct {
	Status int
	Output string
	Errors string
}

// Options: Echo shows the tool's stderr lines as status lines (ffmpeg talks on stderr); Capture keeps stdout and
// stderr in the Result instead.
type Options struct {
	Echo    bool
	Capture bool
}

// Run runs a tool to completion. A non-zero exit is a Status, not an error; failing to start is an error.
// Cancelling ctx kills the tool.
func Run(ctx context.Context, executable string, args []string, opt Options) (Result, error) {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Stdin = nil
	var stdout, stderr bytes.Buffer
	var wg sync.WaitGroup
	if opt.Capture {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	} else {
		cmd.Stdout = console.Out()
		pipe, err := cmd.StderrPipe()
		if err != nil {
			return Result{}, err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			echoLines(pipe, opt.Echo)
		}()
	}
	if err := cmd.Start(); err != nil {
		return Result{}, err
	}
	wg.Wait()
	err := cmd.Wait()
	res := Result{Output: stdout.String(), Errors: stderr.String()}
	var exit *exec.ExitError
	switch {
	case err == nil:
	case errors.As(err, &exit):
		res.Status = exit.ExitCode()
		if ctx.Err() != nil {
			return res, ctx.Err()
		}
	default:
		return res, err
	}
	return res, nil
}

// echoLines reads a tool's stderr and shows each non-empty line (split on \n and \r) as a status line.
func echoLines(r io.Reader, show bool) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	sc.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
			return i + 1, data[:i], nil
		}
		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); show && line != "" {
			console.Status(line)
		}
	}
}

// InstallHint says how to install tools on this system: space-separated names (ffmpeg, yt-dlp, deno, gpac, aria2).
func InstallHint(names string) string {
	pkgs := strings.Fields(names)
	switch runtime.GOOS {
	case "darwin":
		return "brew install " + strings.Join(pkgs, " ")
	case "windows":
		var cmds []string
		for _, p := range pkgs {
			cmds = append(cmds, "winget install "+wingetIDs[p])
		}
		return strings.Join(cmds, "; ")
	}
	// Elsewhere yt-dlp and deno come from their projects: distributions ship yt-dlp too old for YouTube, and most
	// have no deno. The rest comes from the package manager. Every hint after the first names its tool.
	var managed, own []string
	for _, p := range pkgs {
		if _, ok := linuxHints[p]; ok {
			own = append(own, p)
		} else {
			managed = append(managed, p)
		}
	}
	var hints []string
	if len(managed) > 0 {
		hints = append(hints, "install "+strings.Join(managed, " ")+" with your package manager (apt, dnf, pacman…)")
	}
	for _, p := range own {
		hint := linuxHints[p]
		if p == "yt-dlp" && runtime.GOARCH == "arm64" {
			hint = strings.Replace(hint, "yt-dlp_linux", "yt-dlp_linux_aarch64", 1)
		}
		if len(hints) > 0 {
			hint = p + ": " + hint
		}
		hints = append(hints, hint)
	}
	return strings.Join(hints, "; ")
}

// linuxHints install the tools that should not come from a Linux distribution.
var linuxHints = map[string]string{
	"yt-dlp": `pipx install "yt-dlp[default]", or yt-dlp_linux from https://github.com/yt-dlp/yt-dlp/releases/latest ` +
		`(distribution packages are often too old for YouTube)`,
	"deno": "curl -fsSL https://deno.land/install.sh | sh -s -- -y",
}

// wingetIDs are the winget package ids of the tools haul drives.
var wingetIDs = map[string]string{
	"ffmpeg": "Gyan.FFmpeg", "yt-dlp": "yt-dlp.yt-dlp", "deno": "DenoLand.Deno", "gpac": "GPAC.GPAC", "aria2": "aria2.aria2",
}
