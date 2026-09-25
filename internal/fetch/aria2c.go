package fetch

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/shell"
)

// Aria2c downloads through aria2c. Resources that allow only closed ranges one at a time go to Fallback instead:
// aria2c cannot be held to that.
type Aria2c struct {
	Path     string
	Args     []string
	Fallback Fetcher
}

// Fetch runs aria2c for res.
func (a *Aria2c) Fetch(ctx context.Context, res media.Resource, dst string, p Progress) error {
	if res.Policy == media.Sequential && a.Fallback != nil {
		return a.Fallback.Fetch(ctx, res, dst, p)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	result, err := shell.Run(ctx, a.Path, a.Arguments(res, dst), shell.Options{Echo: true})
	if err != nil {
		return err
	}
	if _, err := os.Stat(dst + ".aria2"); result.Status != 0 || err == nil {
		return errs.New("aria2c failed (exit code %d)", result.Status)
	}
	if _, err := os.Stat(dst); err != nil {
		return errs.New("aria2c wrote nothing")
	}
	return nil
}

// Arguments is the aria2c command line: the resource's headers, sorted, then the extra arguments.
func (a *Aria2c) Arguments(res media.Resource, dst string) []string {
	args := []string{"--auto-file-renaming=false", "--download-result=hide", "--allow-overwrite=true",
		"--console-log-level=warn", "-x16", "-s16", "-j16", "-k5M"}
	names := make([]string, 0, len(res.Header))
	for k := range res.Header {
		names = append(names, k)
	}
	sort.Strings(names)
	for _, k := range names {
		for _, v := range res.Header[k] {
			args = append(args, "--header="+k+": "+v)
		}
	}
	args = append(args, a.Args...)
	return append(args, res.URL, "-d", filepath.Dir(dst), "-o", filepath.Base(dst))
}

// SplitArgs is shell-style word splitting with quotes and backslashes, for --aria2c-args.
func SplitArgs(text string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	hasToken, escaped := false, false
	for _, ch := range text {
		switch {
		case escaped:
			cur.WriteRune(ch)
			escaped = false
		case ch == '\\' && quote != '\'':
			escaped = true
		case quote != 0:
			if ch == quote {
				quote = 0
			} else {
				cur.WriteRune(ch)
			}
		case ch == '"' || ch == '\'':
			quote = ch
			hasToken = true
		case unicode.IsSpace(ch):
			if hasToken {
				out = append(out, cur.String())
				cur.Reset()
				hasToken = false
			}
		default:
			cur.WriteRune(ch)
			hasToken = true
		}
	}
	if hasToken {
		out = append(out, cur.String())
	}
	return out
}
