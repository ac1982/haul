package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/shell"
	"github.com/ac1982/haul/internal/testkit"
)

func runFFmpeg(args ...string) (string, error) {
	res, err := shell.Run(context.Background(), testkit.FFmpeg, append([]string{"-v", "error", "-y"}, args...), shell.Options{Capture: true})
	if err == nil && res.Status != 0 {
		err = fmt.Errorf("ffmpeg exit %d", res.Status)
	}
	return res.Errors, err
}

func TestPageSpecs(t *testing.T) {
	cases := []struct {
		spec  string
		n     int
		focus int
		want  []int
	}{
		{"ALL", 5, 0, nil}, {"", 5, 0, nil}, {"", 5, 4, []int{4}}, {"all", 5, 4, nil},
		{"8", 9, 0, []int{8}}, {"1,2", 5, 0, []int{1, 2}}, {"3-5", 5, 0, []int{3, 4, 5}},
		{"last", 5, 0, []int{5}}, {"3,5,LATEST", 9, 0, []int{3, 5, 9}}, {"1-2,4", 5, 0, []int{1, 2, 4}}, {" 2, ", 5, 0, []int{2}},
	}
	for _, c := range cases {
		got, err := SelectPages(c.spec, c.n, c.focus)
		if err != nil || fmt.Sprint(got) != fmt.Sprint(c.want) {
			t.Errorf("SelectPages(%q) = %v, %v; want %v", c.spec, got, err, c.want)
		}
	}
	for _, bad := range []string{"x", "5-3", "1,,a", "1-x"} {
		if _, err := SelectPages(bad, 5, 0); !errs.Is(err, errs.Input) {
			t.Errorf("SelectPages(%q) err = %v", bad, err)
		}
	}
}

func videos() []media.VideoFormat {
	return []media.VideoFormat{
		{ID: "a", Rank: 80, Quality: "1080P", Codec: "AVC", Bitrate: 2000},
		{ID: "b", Rank: 80, Quality: "1080P", Codec: "HEVC", Bitrate: 1000},
		{ID: "c", Rank: 120, Quality: "4K", Codec: "HEVC", Bitrate: 9000},
		{ID: "d", Rank: 32, Quality: "480P", Codec: "AVC", Bitrate: 500},
	}
}

func order(v []media.VideoFormat) string {
	var ids []string
	for _, f := range v {
		ids = append(ids, f.ID)
	}
	return strings.Join(ids, "")
}

func TestVideoOrder(t *testing.T) {
	cases := []struct {
		name string
		o    Options
		want string
	}{
		{"quality then bitrate", Options{}, "cabd"},
		// A codec list alone outranks resolution: -c avc means AVC even when another codec goes higher.
		{"codec", Options{Codec: []string{"avc", "hevc"}}, "adcb"},
		{"quality list", Options{Quality: []string{"480p", "1080P"}, Codec: []string{"hevc"}}, "dbac"},
		{"codec first", Options{Quality: []string{"1080P"}, Codec: []string{"avc"}, CodecFirst: true}, "adbc"},
		{"ascending", Options{VideoAscending: true}, "dbac"},
	}
	for _, c := range cases {
		if got := order(c.o.SortVideo(videos())); got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
	}
}

func TestAudioOrder(t *testing.T) {
	audio := []media.AudioFormat{{ID: "1", Codec: "M4A", Bitrate: 130}, {ID: "2", Codec: "E-AC-3", Bitrate: 640}, {ID: "3", Codec: "M4A", Bitrate: 320}}
	ids := func(a []media.AudioFormat) string {
		s := ""
		for _, f := range a {
			s += f.ID
		}
		return s
	}
	if got := ids(Options{}.SortAudio(audio)); got != "231" {
		t.Errorf("plain: %s", got)
	}
	if got := ids(Options{Codec: []string{"m4a", "eac3"}}.SortAudio(audio)); got != "312" {
		t.Errorf("codec: %s", got)
	}
	if got := ids(Options{AudioAscending: true}.SortAudio(audio)); got != "132" {
		t.Errorf("ascending: %s", got)
	}
	if got := SplitList(" hevc，e-ac-3, ,av1"); strings.Join(got, "|") != "hevc|e-ac-3|av1" {
		t.Errorf("SplitList: %q", got)
	}
}

func TestTemplates(t *testing.T) {
	entry := &media.Entry{Index: 3, ID: "BV17x411w7KC", Title: "第三话/标题?", Uploader: media.Person{Name: "UP:主", ID: "1"},
		Published: time.Date(2024, 5, 6, 7, 8, 9, 0, time.Local), Fields: map[string]string{"bvid": "BV17x411w7KC", "aid": "170001", "cid": "279786", "api": "TV"}}
	item := &media.Item{Site: "bilibili", Title: "标题.", Published: time.Date(2023, 1, 2, 3, 4, 5, 0, time.Local)}
	for i := 0; i < 12; i++ {
		item.Entries = append(item.Entries, entry)
	}
	v := &media.VideoFormat{Quality: "1080P", Width: 1920, Height: 1080, FPS: 29.97, Codec: "HEVC", Bitrate: 1200}
	a := &media.AudioFormat{Codec: "M4A", Bitrate: 130}
	d := templateData{item: item, entry: entry, video: v, audio: a}
	cases := map[string]string{
		DefaultListTemplate: "标题/[P03]第三话_标题_",
		"<uploader>-<bvid>-<aid>-<cid>-<quality>-<resolution>-<fps>-<videoCodec>-<videoBitrate>-<audioCodec>-<audioBitrate>-<api>-<id>-<site>-<unknown>": "UP_主-BV17x411w7KC-170001-279786-1080P-1920x1080-29.97-HEVC-1200-M4A-130-TV-BV17x411w7KC-bilibili-<unknown>",
		"<publishDate:yyyy-MM-dd>/<pageDate:MMdd> <uploaderId>": "2023-01-02/0506 1",
		`out\<title>.mp4`:      "out/标题",
		"../<title>/./.hidden": "标题/_hidden",
		"<pageTitle> <cid2>":   "第三话_标题_ <cid2>",
		"<bitrate>":            "<bitrate>",
	}
	for tpl, want := range cases {
		if got := d.Render(tpl); got != want {
			t.Errorf("Render(%q) = %q, want %q", tpl, got, want)
		}
	}
	if got := (templateData{item: &media.Item{Title: ""}, entry: &media.Entry{}}).Render("<title>"); got != "untitled" {
		t.Errorf("empty = %q", got)
	}
}

func TestTags(t *testing.T) {
	entry := &media.Entry{Title: "P2", URL: "https://e", Uploader: media.Person{Name: "UP"}}
	single := &media.Item{Title: "T", Entries: []*media.Entry{entry}, URL: "https://i", Description: "D"}
	if got := tags(single, entry); got.Title != "T" || got.Album != "" || got.Comment != "https://e" || got.Artist != "UP" || got.Description != "D" {
		t.Errorf("single: %+v", got)
	}
	list := &media.Item{Title: "T", Entries: []*media.Entry{entry, entry}}
	if got := tags(list, entry); got.Title != "P2" || got.Album != "T" {
		t.Errorf("list: %+v", got)
	}
	ep := &media.Entry{Title: "E", Album: "Show"}
	if got := tags(&media.Item{Title: "E", Entries: []*media.Entry{ep}}, ep); got.Title != "E" || got.Album != "Show" {
		t.Errorf("podcast: %+v", got)
	}
}
