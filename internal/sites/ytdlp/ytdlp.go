// Package ytdlp reads YouTube and X through yt-dlp. YouTube's player challenges need yt-dlp's JavaScript solver, and
// X changes its API often; letting yt-dlp find the streams keeps those fixes a `brew upgrade` away. haul still
// downloads the streams itself, so it only asks yt-dlp for metadata (`-J`).
package ytdlp

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/shell"
)

// Options configure how yt-dlp is found.
type Options struct {
	// Path is an explicit yt-dlp executable (or a bare name); empty looks for yt-dlp next to haul and on PATH.
	Path string
}

// Extractor reads one yt-dlp backed site: YouTube or X.
type Extractor struct {
	profile  *profile
	opts     Options
	denoOnce sync.Once
}

var _ extract.Extractor = (*Extractor)(nil)

// NewYouTube reads YouTube videos. The client is not used: yt-dlp fetches the metadata itself, and the streams are
// downloaded by the engine with the headers the formats carry. It is taken so every site is built the same way.
func NewYouTube(_ *httpx.Client, opts Options) *Extractor {
	return &Extractor{profile: &youtubeProfile, opts: opts}
}

// NewX reads X (Twitter) posts; see NewYouTube about the client.
func NewX(_ *httpx.Client, opts Options) *Extractor {
	return &Extractor{profile: &xProfile, opts: opts}
}

// Info describes the site.
func (e *Extractor) Info() extract.Info { return e.profile.info }

// Match recognises the site's links.
func (e *Extractor) Match(link string) (string, bool) { return e.profile.match(link) }

// Resolve runs yt-dlp on the link. Every entry comes back with its formats, so Formats is never needed after it.
func (e *Extractor) Resolve(ctx context.Context, link string) (*media.Item, error) {
	ytdlp := e.executable()
	if ytdlp == "" {
		info := e.profile.info
		return nil, errs.NewDependency("%s links need yt-dlp: %s", info.Name, info.Requires[0].Install)
	}
	if e.profile.needsDeno {
		e.denoOnce.Do(func() {
			if shell.FindExecutable("deno") == "" {
				console.Warn("deno not found; yt-dlp may miss most YouTube formats: " + shell.InstallHint("deno"))
			}
		})
	}
	root, err := run(ctx, ytdlp, link)
	if err != nil {
		return nil, err
	}
	return parse(root, e.profile)
}

// Formats returns what Resolve attached; an entry without formats did not come from Resolve.
func (e *Extractor) Formats(_ context.Context, _ *media.Item, entry *media.Entry) (*media.Formats, error) {
	if entry.Formats == nil {
		return nil, errs.New("No formats for %s entry %s", e.profile.info.Name, entry.ID)
	}
	return entry.Formats, nil
}

func (e *Extractor) executable() string {
	if e.opts.Path != "" {
		return shell.Resolve(e.opts.Path)
	}
	return shell.FindExecutable("yt-dlp")
}

// errorPrefix is what yt-dlp puts before its reason: `ERROR: [youtube] abc: `.
var errorPrefix = regexp.MustCompile(`^ERROR:\s*(\[[\w:]+\]\s*\S+:\s*)?`)

// run asks yt-dlp for the link's metadata. Only the link itself is fetched (--no-playlist): a watch link inside a
// playlist is one video.
func run(ctx context.Context, ytdlp, link string) (jsonv.Value, error) {
	args := []string{"-J", "--no-playlist", "--no-progress", "--", link}
	console.Debugf("%s %s", ytdlp, strings.Join(args, " "))
	res, err := shell.Run(ctx, ytdlp, args, shell.Options{Capture: true})
	if err != nil {
		if ctx.Err() != nil {
			return jsonv.Null, ctx.Err()
		}
		return jsonv.Null, errs.New("yt-dlp failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(res.Errors), "\n")
	for _, line := range lines {
		if line != "" {
			console.Debugf("yt-dlp: %s", line)
		}
	}
	if res.Status != 0 || strings.TrimSpace(res.Output) == "" {
		return jsonv.Null, errs.New("yt-dlp failed: %s", failureReason(lines, res.Status))
	}
	return jsonv.ParseString(res.Output)
}

// failureReason is yt-dlp's last ERROR line without its prefix, or the exit code when it gave none.
func failureReason(lines []string, status int) string {
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); strings.HasPrefix(line, "ERROR") {
			return errorPrefix.ReplaceAllString(line, "")
		}
	}
	return "exit code " + strconv.Itoa(status)
}
