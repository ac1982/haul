package main

import (
	"context"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/engine"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/fetch"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/render"
	"github.com/ac1982/haul/internal/shell"
	"github.com/ac1982/haul/internal/sites/bilibili"
	"github.com/ac1982/haul/internal/sites/podcast"
	"github.com/ac1982/haul/internal/sites/ytdlp"
	"github.com/spf13/pflag"
)

// download is `haul <url>` and `haul info <url>`.
func download(ctx context.Context, command string, args []string) int {
	defs := downloadFlags
	if command == "info" {
		defs = infoFlags
	}
	fs := newFlagSet(command, defs)
	if code, ok := parse(fs, command, args); !ok {
		return code
	}
	jsonOut := flagBool(fs, "json")
	if jsonOut {
		console.SetToStderr(true)
	}
	console.SetDebug(flagBool(fs, "debug"))
	link := strings.Join(fs.Args(), " ")
	if len(fs.Args()) != 1 {
		if len(fs.Args()) == 0 {
			return usageError(command, errMissing("<url>"))
		}
		return usageError(command, errTooMany(fs.Args()))
	}
	if !jsonOut {
		console.Banner("haul " + console.CurrentStyle().Dim(version))
		console.Plain("")
	}
	result, err := runEngine(ctx, command, fs, defs, link)
	if jsonOut {
		console.Output(encodeDocument(document(command, link, result, err, command == "info" && flagBool(fs, "urls"))))
	}
	if err != nil {
		return fail(err)
	}
	return 0
}

// runEngine builds the extractors and the engine from the options, and runs the link.
func runEngine(ctx context.Context, command string, fs *pflag.FlagSet, defs []flagDef, link string) (*engine.Result, error) {
	if err := applyConfig(fs, defs); err != nil {
		return nil, err
	}
	console.SetDebug(flagBool(fs, "debug"))
	o, err := engineOptions(fs, command == "info")
	if err != nil {
		return nil, err
	}
	router, err := newRouter(fs, o)
	if err != nil {
		return nil, err
	}
	fetcher, err := newFetcher(fs)
	if err != nil {
		return nil, err
	}
	term := &render.Terminal{ShowAll: flagBool(fs, "show-all"), HideStreams: flagBool(fs, "hide-streams"), URLs: o.IncludeURLs,
		List: o.List, Interactive: o.Interactive}
	e := &engine.Engine{
		Router: router, Client: httpx.Default, Fetcher: fetcher, Observer: term, Chooser: term, Options: o,
		FFmpeg: tool(fs, "ffmpeg", "ffmpeg"), MP4Box: tool(fs, "mp4box", "MP4Box"), RetryWait: 3 * time.Second,
	}
	return e.Run(ctx, link)
}

// engineOptions maps the command line onto the engine's options.
func engineOptions(fs *pflag.FlagSet, info bool) (engine.Options, error) {
	o := engine.DefaultOptions()
	o.Dir = flagString(fs, "work-dir")
	o.Pages = flagString(fs, "pages")
	o.Quality = engine.SplitList(flagString(fs, "quality"))
	o.Codec = engine.SplitList(flagString(fs, "codec"))
	o.CodecFirst = appearsFirst(os.Args, []string{"-c", "--codec"}, []string{"-q", "--quality"})
	o.VideoAscending = flagBool(fs, "video-ascending")
	o.AudioAscending = flagBool(fs, "audio-ascending")
	o.VideoIndex, _ = fs.GetInt("video-stream")
	o.AudioIndex, _ = fs.GetInt("audio-stream")
	o.Interactive = flagBool(fs, "interactive")
	o.SubtitleLangs = engine.SplitList(flagString(fs, "sub-lang"))
	o.AutoSubtitles = flagBool(fs, "auto-subtitles")
	o.Template = flagString(fs, "output")
	o.ListTemplate = flagString(fs, "multi-output")
	o.AudioLanguage = flagString(fs, "lang")
	o.NoTags = flagBool(fs, "no-tags")
	o.UseMP4Box = flagBool(fs, "use-mp4box")
	o.Archive = flagBool(fs, "archive")
	delay, _ := fs.GetInt("delay")
	o.Delay = time.Duration(delay) * time.Second
	o.List = info
	o.IncludeURLs = info && flagBool(fs, "urls")

	c := engine.DefaultContent()
	switch {
	case flagBool(fs, "audio-only") && flagBool(fs, "video-only"):
		return o, errs.NewInput("--audio-only and --video-only exclude each other")
	case flagBool(fs, "audio-only"):
		c.Tracks = engine.AudioOnly
	case flagBool(fs, "video-only"):
		c.Tracks = engine.VideoOnly
	}
	if flagBool(fs, "skip-subtitle") {
		c.Subtitles = engine.Skip
	}
	if flagBool(fs, "skip-cover") {
		c.Cover = engine.Skip
	}
	c.NoMux = flagBool(fs, "skip-mux")
	if flagBool(fs, "danmaku") {
		c.Sidecars = append(c.Sidecars, "danmaku")
	}
	// The --*-only switches produce just that one thing.
	only := func() engine.Content {
		return engine.Content{Tracks: engine.NoTracks, Subtitles: engine.Skip, Cover: engine.Skip}
	}
	switch {
	case flagBool(fs, "subtitle-only"):
		if flagBool(fs, "skip-subtitle") {
			return o, errs.NewInput("--subtitle-only and --skip-subtitle exclude each other")
		}
		c = only()
		c.Subtitles = engine.Files
	case flagBool(fs, "cover-only"):
		c = only()
		c.Cover = engine.Files
	case flagBool(fs, "danmaku-only"):
		c = only()
		c.Sidecars = []string{"danmaku"}
	}
	o.Content = c
	return o, nil
}

// newRouter builds every site's extractor, in the order links are tried.
func newRouter(fs *pflag.FlagSet, o engine.Options) (*extract.Router, error) {
	client := httpx.Default
	b := bilibili.DefaultOptions()
	api, err := bilibili.ParseAPI(flagString(fs, "api"))
	if err != nil {
		return nil, err
	}
	b.API = api
	b.Cookie = flagString(fs, "cookie")
	b.Token = flagString(fs, "token")
	b.UserAgent = flagString(fs, "user-agent")
	b.UposHost = flagString(fs, "upos-host")
	b.ReplaceHost = !flagBool(fs, "no-replace-host")
	b.AllowPCDN = flagBool(fs, "allow-pcdn")
	b.ForceHTTP = !flagBool(fs, "no-force-http")
	if h := flagString(fs, "host"); h != "" {
		b.Host = h
	}
	if h := flagString(fs, "ep-host"); h != "" {
		b.EpHost = h
	}
	if h := flagString(fs, "tv-host"); h != "" {
		b.TVHost = h
	}
	b.Area = flagString(fs, "area")
	b.DanmakuFormats = engine.SplitList(strings.ToLower(flagString(fs, "danmaku-format")))
	for _, f := range b.DanmakuFormats {
		if f != "xml" && f != "ass" {
			return nil, errs.NewInput("Unknown danmaku format %q: use xml, ass", f)
		}
	}
	for _, c := range o.Codec {
		if u := strings.ToUpper(c); slices.Contains([]string{"HEVC", "AVC", "AV1"}, u) {
			b.PreferredCodec = u
			break
		}
	}
	y := ytdlp.Options{Path: flagString(fs, "yt-dlp")}
	return extract.NewRouter(
		ytdlp.NewYouTube(client, y),
		ytdlp.NewX(client, y),
		bilibili.New(client, b),
		podcast.NewXiaoyuzhou(client),
		podcast.NewApple(client),
	), nil
}

// newFetcher is the built-in downloader, or aria2c.
func newFetcher(fs *pflag.FlagSet) (fetch.Fetcher, error) {
	h := fetch.NewHTTP(httpx.Default)
	if flagBool(fs, "single-connection") {
		h.Parallel = 1
	}
	if !flagBool(fs, "use-aria2c") {
		return h, nil
	}
	path := tool(fs, "aria2c", "aria2c")
	if path == "" {
		return nil, errs.NewDependency("aria2c not found: %s", shell.InstallHint("aria2"))
	}
	return &fetch.Aria2c{Path: path, Args: fetch.SplitArgs(flagString(fs, "aria2c-args")), Fallback: h}, nil
}

// tool is the path of an external tool: the one given, else found on PATH; "" when missing.
func tool(fs *pflag.FlagSet, flag, name string) string {
	if given := flagString(fs, flag); given != "" {
		if p := shell.Resolve(given); p != "" {
			return p
		}
		console.Warn(name + " not found at " + given + "; looking on PATH")
	}
	return shell.FindExecutable(name)
}

// appearsFirst reports whether one of a comes before every one of b on the command line.
func appearsFirst(args, a, b []string) bool {
	ia, ib := -1, -1
	for i, arg := range args {
		name, _, _ := strings.Cut(arg, "=")
		if ia < 0 && slices.Contains(a, name) {
			ia = i
		}
		if ib < 0 && slices.Contains(b, name) {
			ib = i
		}
	}
	return ia >= 0 && ib >= 0 && ia < ib
}

func flagBool(fs *pflag.FlagSet, name string) bool {
	v, _ := fs.GetBool(name)
	return v
}

func flagString(fs *pflag.FlagSet, name string) string {
	v, _ := fs.GetString(name)
	return strings.TrimSpace(v)
}

type usageErr string

func (e usageErr) Error() string { return string(e) }

func errMissing(what string) error { return usageErr("missing expected argument '" + what + "'") }

func errTooMany(args []string) error {
	return usageErr("expected one link, got " + strings.Join(quoteAll(args), " ") + " (quote links that contain spaces)")
}

func quoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = "'" + s + "'"
	}
	return out
}
