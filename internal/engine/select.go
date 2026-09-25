package engine

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/subtitle"
)

// SelectPages parses a -p spec against n entries: 8, 1,2, 3-5, ALL, LAST (also LATEST / NEW). An empty spec takes
// focus (the entry the link points at) when there is one. nil means all entries.
func SelectPages(spec string, n, focus int) ([]int, error) {
	text := strings.Trim(strings.ToUpper(strings.TrimSpace(spec)), ",")
	if text == "" {
		if focus > 0 {
			return []int{focus}, nil
		}
		return nil, nil
	}
	if text == "ALL" {
		return nil, nil
	}
	for _, last := range []string{"LATEST", "LAST", "NEW"} {
		text = strings.ReplaceAll(text, last, strconv.Itoa(n))
	}
	invalid := errs.NewInput("Invalid page selection -p %s: use 8, 1,2, 3-5, ALL or LAST", spec)
	var out []int
	for _, token := range strings.Split(text, ",") {
		part := strings.TrimSpace(token)
		if a, b, ok := strings.Cut(part, "-"); ok {
			from, err1 := strconv.Atoi(strings.TrimSpace(a))
			to, err2 := strconv.Atoi(strings.TrimSpace(b))
			if err1 != nil || err2 != nil || from > to {
				return nil, invalid
			}
			for i := from; i <= to; i++ {
				out = append(out, i)
			}
			continue
		}
		i, err := strconv.Atoi(part)
		if err != nil {
			return nil, invalid
		}
		out = append(out, i)
	}
	return out, nil
}

// priorities maps each listed name (upper case, dashes removed) to its rank.
func priorities(list []string) map[string]int {
	m := map[string]int{}
	for i, name := range list {
		key := normalizeName(name)
		if _, dup := m[key]; key != "" && !dup {
			m[key] = i
		}
	}
	return m
}

func normalizeName(s string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(s)), "-", "")
}

func rank(m map[string]int, name string) int {
	if r, ok := m[normalizeName(name)]; ok {
		return r
	}
	return len(m) + 100
}

// SortVideo orders video streams the way haul chooses: listed qualities, then listed codecs (the other way round
// with codecFirst), then quality rank and bitrate — highest first, or lowest first when ascending. A codec list
// alone outranks resolution: -c avc means AVC even when another codec goes higher.
func (o Options) SortVideo(in []media.VideoFormat) []media.VideoFormat {
	q, c := priorities(o.Quality), priorities(o.Codec)
	codecFirst := o.CodecFirst && len(q) > 0 && len(c) > 0
	out := slices.Clone(in)
	sign := -1
	if o.VideoAscending {
		sign = 1
	}
	slices.SortStableFunc(out, func(a, b media.VideoFormat) int {
		first, second := cmp.Compare(rank(q, a.Quality), rank(q, b.Quality)), cmp.Compare(rank(c, a.Codec), rank(c, b.Codec))
		if codecFirst {
			first, second = second, first
		}
		return cmp.Or(first, second, sign*cmp.Compare(a.Rank, b.Rank), sign*cmp.Compare(a.Bitrate, b.Bitrate))
	})
	return out
}

// SortAudio orders audio streams: listed codecs, then bitrate, highest first unless ascending.
func (o Options) SortAudio(in []media.AudioFormat) []media.AudioFormat {
	c := priorities(o.Codec)
	out := slices.Clone(in)
	sign := -1
	if o.AudioAscending {
		sign = 1
	}
	slices.SortStableFunc(out, func(a, b media.AudioFormat) int {
		return cmp.Or(cmp.Compare(rank(c, a.Codec), rank(c, b.Codec)), sign*cmp.Compare(a.Bitrate, b.Bitrate))
	})
	return out
}

// SplitList splits a comma-separated priority list (full-width commas too).
func SplitList(text string) []string {
	var out []string
	for _, p := range strings.Split(strings.ReplaceAll(text, "，", ","), ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// filterSubtitles keeps only the wanted languages (a prefix matches: en takes en-US) and drops machine-generated
// tracks unless they are wanted or everything is being listed.
func (o Options) filterSubtitles(subs []media.Subtitle, listing bool) []media.Subtitle {
	out := []media.Subtitle{}
	for _, s := range subs {
		if s.Auto && !o.AutoSubtitles && !listing {
			continue
		}
		if len(o.SubtitleLangs) > 0 && !slices.ContainsFunc(o.SubtitleLangs, func(l string) bool {
			lang, want := strings.ToLower(s.Lang), strings.ToLower(l)
			return lang == want || strings.HasPrefix(lang, want+"-")
		}) {
			continue
		}
		out = append(out, s)
	}
	return out
}

// pick is the chosen index: the requested one, checked against count, or 0.
func pick(index, count int, flag string) (int, error) {
	if count == 0 {
		return -1, nil
	}
	if index < 0 {
		return 0, nil
	}
	if index >= count {
		return 0, errs.NewInput("%s %d is out of range: this page has %d (0-%d); `haul info` lists them", flag, index, count, count-1)
	}
	return index, nil
}

// LanguageLabel is how a subtitle stream is tagged: ISO 639-2 code and a readable name.
func LanguageLabel(s media.Subtitle) subtitle.Language { return subtitle.LanguageOf(s.Lang, s.Auto) }

func convertSubtitle(f media.SubtitleFormat, data []byte) (string, error) {
	return subtitle.ToSRT(f, data)
}
