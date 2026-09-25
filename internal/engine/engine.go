package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/fetch"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/mux"
	"github.com/ac1982/haul/internal/shell"
)

// Engine runs links through the pipeline.
type Engine struct {
	Router *extract.Router
	// Client fetches covers and subtitles.
	Client  *httpx.Client
	Fetcher fetch.Fetcher
	// FFmpeg and MP4Box are the tools' paths; "" when not found.
	FFmpeg   string
	MP4Box   string
	Observer Observer
	// Chooser answers -i; nil means no interaction.
	Chooser Chooser
	Options Options
	// RetryWait is the pause before an entry is tried again.
	RetryWait time.Duration
}

// attempts is how often an entry is tried when its download fails.
const attempts = 3

// Run resolves a link and works through the selected entries. The Result is returned whenever the item was read,
// also with an error, so a caller can report what was done before it.
func (e *Engine) Run(ctx context.Context, link string) (*Result, error) {
	o := e.Options
	x, normalized, err := e.Router.Route(link)
	if err != nil {
		return nil, err
	}
	if err := e.checkTools(); err != nil {
		return nil, err
	}
	dir, err := e.outputDir()
	if err != nil {
		return nil, err
	}
	item, err := x.Resolve(ctx, normalized)
	if err != nil {
		return nil, cancelled(ctx, err)
	}
	res := &Result{Item: item, Site: x.Info()}
	for _, entry := range item.Entries {
		res.Entries = append(res.Entries, &EntryResult{Entry: entry, ChosenVideo: -1, ChosenAudio: -1})
	}

	// info on a list without -p lists the entries only: reading every entry's streams can take minutes.
	if o.List && strings.TrimSpace(o.Pages) == "" && item.Focus == 0 && len(item.Entries) > 1 {
		e.observer().Resolved(item, res.Site, nil)
		console.Status(fmt.Sprintf("%d %ss; -p <n> lists the streams of one, -p ALL of all", len(item.Entries), res.Site.Unit))
		e.observer().Finished(res)
		return res, nil
	}
	pages, err := SelectPages(o.Pages, len(item.Entries), item.Focus)
	if err != nil {
		return res, err
	}
	var selected []*EntryResult
	for _, r := range res.Entries {
		if pages == nil || slices.Contains(pages, r.Entry.Index) {
			r.Selected = true
			selected = append(selected, r)
		}
	}
	if len(selected) == 0 {
		return res, errs.NewInput("-p %s matches none of the %d %ss", o.Pages, len(item.Entries), res.Site.Unit)
	}
	entries := make([]*media.Entry, len(selected))
	for i, r := range selected {
		entries[i] = r.Entry
	}
	e.observer().Resolved(item, res.Site, entries)

	for n, r := range selected {
		if n > 0 && o.Delay > 0 {
			console.Status(fmt.Sprintf("Waiting %s", o.Delay))
			if err := sleep(ctx, o.Delay); err != nil {
				return res, cancelled(ctx, err)
			}
		}
		e.observer().EntryStarted(n+1, len(selected), r.Entry)
		key := archiveKey(item.Site, r.Entry)
		if o.Archive && !o.List && archived(key) {
			console.Status("Already downloaded (" + key + "), skipping")
			r.Status, r.Reason = Skipped, "archive"
			e.observer().EntryFinished(r.Entry, r)
			continue
		}
		if err := e.entryWithRetries(ctx, x, item, r, dir); err != nil {
			return res, cancelled(ctx, err)
		}
		if o.Archive && !o.List && r.Status == Downloaded {
			if err := archive(key); err != nil {
				console.Warn("Could not update the archive: " + err.Error())
			}
		}
		e.observer().EntryFinished(r.Entry, r)
	}
	e.observer().Finished(res)
	return res, nil
}

func (e *Engine) observer() Observer {
	if e.Observer == nil {
		return NopObserver{}
	}
	return e.Observer
}

// checkTools fails early when a needed tool is missing.
func (e *Engine) checkTools() error {
	c := e.Options.Content
	if e.Options.List || c.Tracks == NoTracks {
		return nil
	}
	if e.FFmpeg == "" && !c.NoMux {
		return errs.NewDependency("ffmpeg not found: %s", shell.InstallHint("ffmpeg"))
	}
	if e.Options.UseMP4Box && e.MP4Box == "" && !c.NoMux {
		return errs.NewDependency("MP4Box not found: %s", shell.InstallHint("gpac"))
	}
	return nil
}

// outputDir is the absolute output directory, created.
func (e *Engine) outputDir() (string, error) {
	dir := expandPath(e.Options.Dir)
	if dir == "" {
		dir = "."
	}
	if !e.Options.List {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
	}
	return absPath(dir), nil
}

// entryWithRetries works on an entry, trying again (with fresh stream URLs) when the download fails.
// A bad option, a missing tool or a cancelled choice will not go away by trying again.
func (e *Engine) entryWithRetries(ctx context.Context, x extract.Extractor, item *media.Item, r *EntryResult, dir string) error {
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err = e.entry(ctx, x, item, r, dir); err == nil || errs.KindOf(err) != errs.Failed || ctx.Err() != nil {
			return err
		}
		if attempt < attempts {
			console.Error(err.Error())
			console.Warn(fmt.Sprintf("Retrying in %s (%d/%d)…", e.RetryWait, attempt, attempts-1))
			if serr := sleep(ctx, e.RetryWait); serr != nil {
				return serr
			}
		}
	}
	return err
}

// entry is one entry, from its formats to its files.
func (e *Engine) entry(ctx context.Context, x extract.Extractor, item *media.Item, r *EntryResult, dir string) error {
	o := e.Options
	entry := r.Entry
	formats, err := extract.LoadFormats(ctx, x, item, entry)
	if err != nil {
		return err
	}
	if formats.Raw != "" {
		console.Debugf("Formats: %s", formats.Raw)
	}
	if err := e.choose(entry, formats, r); err != nil {
		return err
	}
	e.observer().Streams(entry, r)
	if o.List {
		r.Status = Listed
		return nil
	}
	p, err := e.plan(item, entry, formats, r, dir)
	if err != nil {
		return err
	}
	return e.produce(ctx, x, item, entry, formats, r, p)
}

// choose sorts the streams, filters the subtitles and picks the streams: by index, interactively, or the first.
// A choice made interactively is kept when the entry is tried again.
func (e *Engine) choose(entry *media.Entry, f *media.Formats, r *EntryResult) error {
	o := e.Options
	video := f.Video
	if f.AudioOnly {
		video = nil
	}
	r.Video, r.Audio = o.SortVideo(video), o.SortAudio(f.Audio)
	r.VideoHasAudio = len(r.Audio) == 0 && len(r.Video) > 0 && r.Video[0].HasAudio
	r.Subtitles = o.filterSubtitles(f.Subtitles, o.List)
	if r.ChosenVideo >= 0 || r.ChosenAudio >= 0 {
		return nil
	}
	var err error
	if r.ChosenVideo, err = pick(o.VideoIndex, len(r.Video), "--video-stream"); err != nil {
		return err
	}
	if r.ChosenAudio, err = pick(o.AudioIndex, len(r.Audio), "--audio-stream"); err != nil {
		return err
	}
	if o.Interactive && !o.List && e.Chooser != nil && (len(r.Video) > 1 || len(r.Audio) > 1) {
		v, a, err := e.Chooser.Choose(entry, r)
		if err != nil {
			return err
		}
		r.ChosenVideo, r.ChosenAudio = v, a
	}
	return nil
}

// plan is where an entry's output goes and what it is made of.
type plan struct {
	video     *media.VideoFormat
	audio     *media.AudioFormat
	container mux.Container
	// output is the media file; base is the same path without extension, for side files.
	output string
	base   string
	// fromVideo: the audio is inside the video stream (X) — for audio-only it is taken out of it.
	fromVideo bool
}

func (e *Engine) plan(item *media.Item, entry *media.Entry, f *media.Formats, r *EntryResult, dir string) (*plan, error) {
	o := e.Options
	p := &plan{}
	if r.ChosenVideo >= 0 {
		p.video = &r.Video[r.ChosenVideo]
	}
	if r.ChosenAudio >= 0 {
		p.audio = &r.Audio[r.ChosenAudio]
	}
	switch o.Content.Tracks {
	case AudioOnly:
		if p.audio == nil && p.video != nil && p.video.HasAudio {
			p.fromVideo = true
		} else {
			p.video = nil
		}
		if p.audio == nil && !p.fromVideo {
			return nil, errs.New("No audio stream found for %q", entry.Title)
		}
	case VideoOnly:
		if f.AudioOnly {
			return nil, errs.NewInput("%q is audio only; drop --video-only", entry.Title)
		}
		if p.video == nil {
			return nil, errs.New("No video stream found for %q", entry.Title)
		}
		p.audio = nil
	case VideoAndAudio:
		if p.video == nil && !f.AudioOnly {
			console.Warn("No video stream found")
		}
		if p.audio == nil && (p.video == nil || !p.video.HasAudio) {
			console.Warn("No audio stream found")
		}
		if p.video == nil && p.audio == nil {
			return nil, errs.New("No streams found for %q", entry.Title)
		}
	}

	switch {
	case f.AudioOnly && p.audio != nil:
		p.container = mux.AudioContainer(p.audio.Codec)
	case o.Content.Tracks == AudioOnly:
		p.container = mux.M4A
		if p.audio != nil {
			p.container = mux.AudioContainer(p.audio.Codec)
		}
	default:
		p.container = mux.MP4
	}
	tpl := o.Template
	if len(item.Entries) > 1 || item.Collection {
		tpl = o.ListTemplate
		if tpl == "" {
			tpl = DefaultListTemplate
		}
	} else if tpl == "" {
		tpl = DefaultTemplate
	}
	rel := templateData{item: item, entry: entry, video: p.video, audio: p.audio}.Render(tpl)
	p.base = filepath.Join(dir, filepath.FromSlash(rel))
	p.output = p.base + "." + string(p.container)
	return p, nil
}

// produce downloads and writes an entry's files.
func (e *Engine) produce(ctx context.Context, x extract.Extractor, item *media.Item, entry *media.Entry, f *media.Formats, r *EntryResult, p *plan) error {
	c := e.Options.Content
	if c.Tracks != NoTracks {
		if info, err := os.Stat(p.output); err == nil && info.Size() > 0 {
			console.Status("Exists, skipping  " + console.PrettyPath(p.output))
			r.Status, r.Reason, r.File, r.Size = Skipped, "exists", absPath(p.output), info.Size()
			return nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(p.base), 0o755); err != nil {
		return err
	}
	work := filepath.Join(filepath.Dir(p.base), ".haul-"+safeName(string(item.Site)+"-"+entry.ID))
	if err := os.MkdirAll(work, 0o755); err != nil {
		return err
	}

	var subs []mux.Subtitle
	if c.Subtitles != Skip && (c.Subtitles == Files || c.Tracks == NoTracks || p.container != mux.MP3) {
		subs = e.subtitles(ctx, r.Subtitles, work, p.base, c.Subtitles == Files || c.Tracks == NoTracks, r)
	}
	cover := ""
	if c.Cover != Skip {
		cover = e.cover(ctx, item, entry, work, p.base, c.Cover == Files || c.Tracks == NoTracks, r)
	}
	for _, kind := range c.Sidecars {
		e.sidecar(ctx, x.Info().Name, f, kind, p.base, r)
	}
	if c.Tracks == NoTracks {
		os.RemoveAll(work)
		r.Status = Downloaded
		if len(r.ExtraFiles) == 0 {
			console.Warn("Nothing to save for " + fmt.Sprintf("%q", entry.Title))
			r.Status, r.Reason = Skipped, "empty"
		}
		return nil
	}

	job := mux.Job{Container: p.container, Output: p.output, AudioLanguage: e.Options.AudioLanguage,
		AudioOnly: c.Tracks == AudioOnly || f.AudioOnly, VideoOnly: c.Tracks == VideoOnly}
	if c.Subtitles == Embed {
		job.Subtitles = subs
	}
	if c.Cover == Embed {
		job.Cover = cover
	}
	var err error
	if p.video != nil {
		if job.Video, err = e.video(ctx, *p.video, work); err != nil {
			return err
		}
		job.HEVC = p.video.Codec == "HEVC"
	}
	if p.audio != nil {
		if job.Audio, err = e.download(ctx, p.audio.Source, filepath.Join(work, "audio-"+safeName(p.audio.ID)+audioExt(p.audio.Codec)), "audio"); err != nil {
			return err
		}
	}
	if c.Tracks == VideoAndAudio {
		for i, x := range f.ExtraAudio {
			path, err := e.download(ctx, x.Audio.Source, filepath.Join(work, fmt.Sprintf("extra-%d.m4a", i)), "dub")
			if err != nil {
				return err
			}
			job.ExtraAudio = append(job.ExtraAudio, mux.Track{Path: path, Title: x.Title, Artist: x.Artist})
		}
	}
	if c.NoMux {
		return e.keepStreams(job, p, work, r)
	}
	job.Chapters = entry.Chapters
	if len(job.Chapters) == 0 {
		job.Chapters = f.Chapters
	}
	if !e.Options.NoTags {
		job.Tags = tags(item, entry)
	}
	m := e.muxer(ctx, p)
	console.Status("Muxing  " + m.Name() + muxExtras(job))
	if err := m.Mux(ctx, job); err != nil {
		return err
	}
	info, err := os.Stat(p.output)
	if err != nil || info.Size() == 0 {
		return errs.New("Muxing wrote no file: %s", p.output)
	}
	os.RemoveAll(work)
	r.Status, r.File, r.Size = Downloaded, absPath(p.output), info.Size()
	return nil
}

// video downloads a video stream, joining its parts when it comes in segments.
func (e *Engine) video(ctx context.Context, v media.VideoFormat, work string) (string, error) {
	dst := filepath.Join(work, "video-"+safeName(v.ID)+".mp4")
	if len(v.Parts) == 0 {
		return e.download(ctx, v.Source, dst, "video")
	}
	var parts []string
	for i, part := range v.Parts {
		path, err := e.download(ctx, part, filepath.Join(work, fmt.Sprintf("part-%03d.flv", i)), fmt.Sprintf("part %d/%d", i+1, len(v.Parts)))
		if err != nil {
			return "", err
		}
		parts = append(parts, path)
	}
	console.Status(fmt.Sprintf("Joining %d segments", len(parts)))
	return dst, mux.ConcatFLV(ctx, e.FFmpeg, parts, dst)
}

func (e *Engine) download(ctx context.Context, res media.Resource, dst, label string) (string, error) {
	if info, err := os.Stat(dst); err == nil && res.Size > 0 && info.Size() == res.Size {
		return dst, nil
	}
	return dst, e.Fetcher.Fetch(ctx, res, dst, e.observer().Transfer(label))
}

// subtitles downloads the subtitles as SRT (ASS stays ASS: ffmpeg converts it when muxing). asFiles puts them next
// to the output as <base>.<lang>.srt instead of the work directory.
func (e *Engine) subtitles(ctx context.Context, subs []media.Subtitle, work, base string, asFiles bool, r *EntryResult) []mux.Subtitle {
	var out []mux.Subtitle
	for _, s := range subs {
		lang := LanguageLabel(s)
		console.Status("Subtitle  " + s.Lang + "  " + lang.Name)
		data, err := e.Client.Bytes(ctx, s.Source.URL, s.Source.Header)
		if err != nil {
			console.Warn(fmt.Sprintf("Subtitle %s failed, skipping: %v", s.Lang, err))
			continue
		}
		text, ext := string(data), ".ass"
		if s.Format != media.ASS {
			if text, err = convertSubtitle(s.Format, data); err != nil {
				console.Warn(fmt.Sprintf("Subtitle %s cannot be read, skipping: %v", s.Lang, err))
				continue
			}
			ext = ".srt"
		}
		if strings.TrimSpace(text) == "" {
			continue
		}
		name := s.Lang
		if s.Auto {
			name += ".auto"
		}
		path := filepath.Join(work, "sub-"+safeName(name)+ext)
		if asFiles {
			path = base + "." + safeName(name) + ext
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			console.Warn(fmt.Sprintf("Subtitle %s cannot be saved: %v", s.Lang, err))
			continue
		}
		if asFiles {
			r.ExtraFiles = append(r.ExtraFiles, absPath(path))
			console.Success("Subtitle  " + console.PrettyPath(path))
		}
		out = append(out, mux.Subtitle{Path: path, Code: lang.Code, Name: lang.Name})
	}
	return out
}

// cover downloads the entry's picture (or the item's). A missing cover is not worth the video: it only warns.
func (e *Engine) cover(ctx context.Context, item *media.Item, entry *media.Entry, work, base string, asFile bool, r *EntryResult) string {
	src := entry.Thumbnail
	if src == "" {
		src = item.Thumbnail
	}
	if src == "" {
		return ""
	}
	data, err := e.Client.Bytes(ctx, src, httpx.Header("User-Agent", httpx.BrowserUserAgent))
	if err != nil {
		console.Warn("Cover download failed, skipping: " + err.Error())
		return ""
	}
	ext := imageExt(src, data)
	path := filepath.Join(work, "cover"+ext)
	if asFile {
		path = base + ext
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		console.Warn("Cover cannot be saved: " + err.Error())
		return ""
	}
	if asFile {
		r.ExtraFiles = append(r.ExtraFiles, absPath(path))
		console.Success("Cover  " + console.PrettyPath(path))
	}
	return path
}

// sidecar writes a side file of a kind next to the output, when the site has that kind.
func (e *Engine) sidecar(ctx context.Context, site string, f *media.Formats, kind, base string, r *EntryResult) {
	for _, s := range f.Sidecars {
		if s.Kind != kind {
			continue
		}
		files, err := s.Write(ctx, base)
		if err != nil {
			console.Warn(fmt.Sprintf("%s failed, skipping: %v", kind, err))
			return
		}
		for _, file := range files {
			r.ExtraFiles = append(r.ExtraFiles, absPath(file))
		}
		return
	}
	console.Warn(fmt.Sprintf("%s has no %s", site, kind))
}

// keepStreams is --skip-mux: the streams are moved next to where the output would be, as they are.
func (e *Engine) keepStreams(job mux.Job, p *plan, work string, r *EntryResult) error {
	move := func(src, suffix string) error {
		if src == "" {
			return nil
		}
		dst := p.base + suffix + filepath.Ext(src)
		if err := os.Rename(src, dst); err != nil {
			return err
		}
		r.ExtraFiles = append(r.ExtraFiles, absPath(dst))
		return nil
	}
	if err := move(job.Video, ".video"); err != nil {
		return err
	}
	if err := move(job.Audio, ".audio"); err != nil {
		return err
	}
	for i, t := range job.ExtraAudio {
		if err := move(t.Path, fmt.Sprintf(".audio%d", i+2)); err != nil {
			return err
		}
	}
	for _, s := range job.Subtitles {
		if err := move(s.Path, "."+strings.TrimPrefix(strings.TrimSuffix(filepath.Base(s.Path), filepath.Ext(s.Path)), "sub-")); err != nil {
			return err
		}
	}
	os.RemoveAll(work)
	r.Status, r.Reason = Downloaded, "no-mux"
	return nil
}

// muxer is ffmpeg, or MP4Box when asked for (MP4 only) or when Dolby Vision needs it and ffmpeg is too old.
func (e *Engine) muxer(ctx context.Context, p *plan) mux.Muxer {
	useMP4Box := e.Options.UseMP4Box && p.container == mux.MP4 && e.MP4Box != ""
	if !useMP4Box && p.video != nil && strings.Contains(p.video.Quality, "Dolby Vision") && !mux.SupportsDolbyVision(ctx, e.FFmpeg) {
		if e.MP4Box != "" {
			console.Warn("Dolby Vision needs ffmpeg ≥ 5.0; muxing with MP4Box")
			useMP4Box = true
		} else {
			console.Warn("Dolby Vision needs ffmpeg ≥ 5.0 or MP4Box (" + shell.InstallHint("gpac") + "); muxing may fail")
		}
	}
	if useMP4Box {
		return mux.MP4Box{Path: e.MP4Box}
	}
	return mux.FFmpeg{Path: e.FFmpeg}
}

func muxExtras(job mux.Job) string {
	s := ""
	if len(job.Subtitles) > 0 {
		s += "  +subtitles"
	}
	if len(job.Chapters) > 0 {
		s += "  +chapters"
	}
	if job.Cover != "" {
		s += "  +cover"
	}
	return s
}

// tags are the file's metadata. An entry of a list is titled by itself with the list as album; a podcast episode
// has its show as album.
func tags(item *media.Item, entry *media.Entry) mux.Tags {
	t := mux.Tags{
		Title:       item.Title,
		Artist:      cmpOr(entry.Uploader.Name, item.Uploader.Name),
		Comment:     cmpOr(entry.URL, item.URL),
		Description: cmpOr(entry.Description, item.Description),
		Date:        cmpOrTime(entry.Published, item.Published),
	}
	if len(item.Entries) > 1 || item.Collection {
		t.Title, t.Album = entry.Title, item.Title
	}
	if entry.Album != "" {
		t.Title, t.Album = entry.Title, entry.Album
	}
	return t
}

func audioExt(codec string) string {
	if mux.AudioContainer(codec) == mux.MP3 {
		return ".mp3"
	}
	return ".m4a"
}

// imageExt is .jpg, .png or .webp, from the bytes or the URL.
func imageExt(src string, data []byte) string {
	switch {
	case len(data) > 3 && data[0] == 0x89 && string(data[1:4]) == "PNG":
		return ".png"
	case len(data) > 12 && string(data[8:12]) == "WEBP":
		return ".webp"
	case len(data) > 2 && data[0] == 0xFF && data[1] == 0xD8:
		return ".jpg"
	}
	u := strings.ToLower(src)
	if i := strings.IndexAny(u, "?#"); i >= 0 {
		u = u[:i]
	}
	if ext := filepath.Ext(u); ext == ".png" || ext == ".webp" || ext == ".jpg" || ext == ".jpeg" {
		return ext
	}
	return ".jpg"
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func safeName(s string) string { return unsafeChars.ReplaceAllString(s, "_") }

// absPath is absolute with symlinks resolved (/tmp → /private/tmp), so it compares equal to what realpath gives.
func absPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return real
	}
	if dir, err := filepath.EvalSymlinks(filepath.Dir(abs)); err == nil {
		return filepath.Join(dir, filepath.Base(abs))
	}
	return abs
}

// expandPath expands ~ and $VARS.
func expandPath(p string) string {
	if p == "" {
		return ""
	}
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return os.ExpandEnv(p)
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// cancelled turns whatever an interrupted run failed with into a Cancelled error.
func cancelled(ctx context.Context, err error) error {
	if err != nil && ctx.Err() != nil {
		return errs.ErrCancelled
	}
	return err
}
