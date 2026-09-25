package mux

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/shell"
)

// MP4Box muxes with GPAC's MP4Box: MP4 only, and the way to Dolby Vision on an old ffmpeg.
type MP4Box struct{ Path string }

// Name is "MP4Box".
func (MP4Box) Name() string { return "MP4Box" }

// Mux runs MP4Box for the job.
func (m MP4Box) Mux(ctx context.Context, job Job) error {
	chapters := ""
	if len(job.Chapters) > 0 {
		tmp, err := os.CreateTemp("", "haul-chapters-*.txt")
		if err != nil {
			return err
		}
		chapters = tmp.Name()
		defer os.Remove(chapters)
		_, err = tmp.WriteString(MP4BoxChapters(job.Chapters))
		tmp.Close()
		if err != nil {
			return err
		}
	}
	args := MP4BoxArguments(job, chapters, console.Debug())
	console.Debugf("MP4Box %s", strings.Join(args, " "))
	res, err := shell.Run(ctx, m.Path, args, shell.Options{Echo: true})
	if err != nil {
		return err
	}
	if res.Status != 0 {
		os.Remove(job.Output)
		return errs.New("Muxing failed (MP4Box exit code %d); --debug shows its output", res.Status)
	}
	return nil
}

// MP4BoxArguments is the MP4Box command line for a job.
func MP4BoxArguments(job Job, chaptersFile string, verbose bool) []string {
	var args []string
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "-inter", "500", "-noprog")
	track := 0
	if job.Video != "" && !(job.AudioOnly && job.Audio != "") {
		args = append(args, "-add", job.Video+"#video:name=")
		track++
	}
	if job.Audio != "" && !job.VideoOnly {
		lang := job.AudioLanguage
		if lang == "" {
			lang = "und"
		}
		args = append(args, "-add", job.Audio+"#audio:lang="+lang)
		track++
	}
	if chaptersFile != "" {
		args = append(args, "-chap", chaptersFile)
	}
	for _, s := range job.Subtitles {
		track++
		args = append(args, "-add", s.Path+":name=:hdlr=sbtl:lang="+s.Code)
		args = append(args, "-udta", fmt.Sprintf("%d:type=name:str=%q", track, s.Name))
	}
	var tags []string
	t := job.Tags
	for _, kv := range [][2]string{{"cover", job.Cover}, {"title", t.Title}, {"album", t.Album}, {"artist", t.Artist},
		{"comment", t.Comment}, {"sdesc", t.Description}} {
		if kv[1] != "" {
			tags = append(tags, kv[0]+"="+strings.ReplaceAll(kv[1], ":", "\\:"))
		}
	}
	if !t.Date.IsZero() {
		tags = append(tags, "created="+format.FormatDate(t.Date, "yyyy-MM-dd"))
	}
	if len(tags) > 0 {
		args = append(args, "-itags", strings.Join(tags, ":"))
	}
	return append(args, "-new", job.Output)
}

// MP4BoxChapters is the chapter list as MP4Box reads it: one "hh:mm:ss title" per line.
func MP4BoxChapters(chapters []media.Chapter) string {
	var b strings.Builder
	for _, c := range chapters {
		fmt.Fprintf(&b, "%s %s\n", format.Duration(int(c.Start.Seconds()), true), c.Title)
	}
	return b.String()
}
