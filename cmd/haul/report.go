package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/ac1982/haul/internal/engine"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/media"
)

// The --json document. Keys are part of haul's interface for scripts and agents: add fields, never rename them.
type jsonDocument struct {
	OK          bool       `json:"ok"`
	Command     string     `json:"command"`
	Input       string     `json:"input"`
	Error       *jsonError `json:"error,omitempty"`
	Site        media.Site `json:"site,omitempty"`
	Title       string     `json:"title,omitempty"`
	Uploader    string     `json:"uploader,omitempty"`
	Published   string     `json:"published,omitempty"`
	Description string     `json:"description,omitempty"`
	LoggedIn    *bool      `json:"loggedIn,omitempty"`
	PageCount   int        `json:"pageCount,omitempty"`
	Files       []string   `json:"files"`
	Pages       []jsonPage `json:"pages,omitempty"`
}

type jsonError struct {
	Kind     errs.Kind `json:"kind"`
	Message  string    `json:"message"`
	ExitCode int       `json:"exitCode"`
}

type jsonPage struct {
	Index           int            `json:"index"`
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	DurationSeconds *int           `json:"durationSeconds,omitempty"`
	Published       string         `json:"published,omitempty"`
	Selected        bool           `json:"selected"`
	Status          engine.Status  `json:"status,omitempty"`
	Reason          string         `json:"reason,omitempty"`
	File            string         `json:"file,omitempty"`
	SizeBytes       int64          `json:"sizeBytes,omitempty"`
	ExtraFiles      []string       `json:"extraFiles,omitempty"`
	Video           []jsonVideo    `json:"video,omitempty"`
	Audio           []jsonAudio    `json:"audio,omitempty"`
	VideoHasAudio   bool           `json:"videoHasAudio,omitempty"`
	SelectedVideo   *int           `json:"selectedVideo,omitempty"`
	SelectedAudio   *int           `json:"selectedAudio,omitempty"`
	Subtitles       []jsonSubtitle `json:"subtitles,omitempty"`
}

type jsonVideo struct {
	Index       int     `json:"index"`
	Quality     string  `json:"quality,omitempty"`
	Resolution  string  `json:"resolution,omitempty"`
	Codec       string  `json:"codec"`
	FPS         float64 `json:"fps,omitempty"`
	BitrateKbps int64   `json:"bitrateKbps,omitempty"`
	SizeBytes   int64   `json:"sizeBytes,omitempty"`
	HasAudio    bool    `json:"hasAudio,omitempty"`
	URL         string  `json:"url,omitempty"`
}

type jsonAudio struct {
	Index       int    `json:"index"`
	Codec       string `json:"codec"`
	BitrateKbps int64  `json:"bitrateKbps,omitempty"`
	SizeBytes   int64  `json:"sizeBytes,omitempty"`
	URL         string `json:"url,omitempty"`
}

type jsonSubtitle struct {
	Lang string `json:"lang"`
	Auto bool   `json:"auto,omitempty"`
}

// document is the --json document of a run: what it found and did, and the error it stopped with.
func document(command, input string, r *engine.Result, err error, urls bool) jsonDocument {
	doc := jsonDocument{OK: err == nil, Command: command, Input: input, Files: []string{}}
	if err != nil {
		k := errs.KindOf(err)
		doc.Error = &jsonError{Kind: k, Message: err.Error(), ExitCode: k.ExitCode()}
	}
	if r == nil {
		return doc
	}
	it := r.Item
	doc.Site, doc.Title, doc.Description, doc.LoggedIn = it.Site, it.Title, it.Description, it.LoggedIn
	doc.Uploader = it.Uploader.Name
	if doc.Uploader == "" && len(it.Entries) > 0 {
		doc.Uploader = it.Entries[0].Uploader.Name
	}
	if !it.Published.IsZero() {
		doc.Published = format.ISO(it.Published.Unix())
	}
	doc.PageCount = len(it.Entries)
	if files := r.Files(); files != nil {
		doc.Files = files
	}
	for _, e := range r.Entries {
		doc.Pages = append(doc.Pages, page(e, urls))
	}
	return doc
}

func page(r *engine.EntryResult, urls bool) jsonPage {
	e := r.Entry
	p := jsonPage{Index: e.Index, ID: e.ID, Title: e.Title, Selected: r.Selected, Status: r.Status, Reason: r.Reason,
		File: r.File, SizeBytes: r.Size, ExtraFiles: r.ExtraFiles, VideoHasAudio: r.VideoHasAudio}
	if e.Duration > 0 {
		d := int(math.Round(e.Duration.Seconds()))
		p.DurationSeconds = &d
	}
	if !e.Published.IsZero() {
		p.Published = format.ISO(e.Published.Unix())
	}
	for i, v := range r.Video {
		jv := jsonVideo{Index: i, Quality: v.Quality, Codec: v.Codec, FPS: math.Round(v.FPS*1000) / 1000, BitrateKbps: v.Bitrate,
			SizeBytes: media.EstimatedSize(v.Size, v.Bitrate, e.Duration), HasAudio: v.HasAudio}
		if v.Width > 0 {
			jv.Resolution = fmt.Sprintf("%dx%d", v.Width, v.Height)
		}
		if urls {
			jv.URL = v.Source.URL
		}
		p.Video = append(p.Video, jv)
	}
	for i, a := range r.Audio {
		ja := jsonAudio{Index: i, Codec: a.Codec, BitrateKbps: a.Bitrate, SizeBytes: media.EstimatedSize(a.Size, a.Bitrate, e.Duration)}
		if urls {
			ja.URL = a.Source.URL
		}
		p.Audio = append(p.Audio, ja)
	}
	if r.Video != nil && r.ChosenVideo >= 0 {
		p.SelectedVideo = &r.ChosenVideo
	}
	if r.Audio != nil && r.ChosenAudio >= 0 {
		p.SelectedAudio = &r.ChosenAudio
	}
	if r.Status != "" {
		p.Subtitles = []jsonSubtitle{}
		for _, s := range r.Subtitles {
			p.Subtitles = append(p.Subtitles, jsonSubtitle{Lang: s.Lang, Auto: s.Auto})
		}
	}
	return p
}

// encodeDocument is pretty JSON with sorted keys and slashes and HTML left as they are.
func encodeDocument(doc jsonDocument) string {
	var generic any
	raw, _ := json.Marshal(doc)
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	_ = dec.Decode(&generic)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	_ = enc.Encode(generic)
	return string(bytes.TrimRight(buf.Bytes(), "\n"))
}
