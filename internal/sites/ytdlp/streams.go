package ytdlp

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// streams picks the formats haul can download: DASH video and audio when the site has them (YouTube), otherwise
// whole files with the audio inside (X). It also returns the headers yt-dlp says the streams need; the first
// non-empty set applies to all of them.
func streams(list jsonv.Value, p *profile) (*media.Formats, http.Header, error) {
	type audioCandidate struct {
		format     media.AudioFormat
		preference int
	}
	var (
		header      http.Header
		video       []media.VideoFormat
		progressive []media.VideoFormat
		audio       []audioCandidate
	)
	for _, f := range list.Array() {
		url, _ := f.Get("url").Str()
		proto, _ := f.Get("protocol").Str()
		drm, _ := f.Get("has_drm").Bool()
		id := f.Get("format_id").String()
		// Plain HTTPS only: no HLS, storyboards, DRC (dynamic range compressed audio) or DRM variants.
		if proto != "https" || url == "" || drm || strings.Contains(id, "drc") {
			continue
		}
		vcodec, hasVcodec := f.Get("vcodec").Str()
		acodec, hasAcodec := f.Get("acodec").Str()
		// Only exact sizes: X's `filesize_approx` is a guess from the bitrate.
		size := f.Get("filesize").Int64Or(0)
		switch {
		case vcodec == "none" && hasAcodec && acodec != "none":
			audio = append(audio, audioCandidate{
				format: media.AudioFormat{
					ID:      id,
					Codec:   audioCodec(acodec),
					Bitrate: rounded(f, "abr", "tbr"),
					Size:    size,
					Source:  p.resource(url, size),
				},
				preference: f.Get("language_preference").IntOr(0),
			})
		case vcodec != "none" && f.Get("width").Exists():
			codec := urlCodec(url)
			if hasVcodec {
				codec = videoCodec(vcodec)
			}
			v := videoFormat(f, id, url, codec, size, p)
			if acodec == "none" {
				video = append(video, v)
			} else {
				v.HasAudio = true
				progressive = append(progressive, v)
			}
		default:
			continue
		}
		if len(header) == 0 {
			header = headerOf(f.Get("http_headers"))
		}
	}
	if len(video) == 0 && len(audio) == 0 {
		video = progressive
	}
	// Dubbed videos carry an audio track per language; the original has the highest preference.
	best := math.MinInt
	for _, a := range audio {
		best = max(best, a.preference)
	}
	out := &media.Formats{Video: video}
	for _, a := range audio {
		if a.preference == best {
			out.Audio = append(out.Audio, a.format)
		}
	}
	if len(out.Video) == 0 && len(out.Audio) == 0 {
		return nil, nil, errs.New("yt-dlp found no downloadable streams")
	}
	for i := range out.Video {
		out.Video[i].Source.Header = header.Clone()
	}
	for i := range out.Audio {
		out.Audio[i].Source.Header = header.Clone()
	}
	return out, header, nil
}

func videoFormat(f jsonv.Value, id, url, codec string, size int64, p *profile) media.VideoFormat {
	w, h := f.Get("width").IntOr(0), f.Get("height").IntOr(0)
	fps := f.Get("fps").FloatOr(0)
	short := min(w, h)
	quality := f.Get("format_note").String()
	if quality == "" {
		quality = strconv.Itoa(short) + "p"
	}
	return media.VideoFormat{
		ID: id,
		// The short side, so portrait and landscape videos of one quality rank alike.
		Rank:    short*1000 + int(math.Round(fps)),
		Quality: quality,
		Width:   w,
		Height:  h,
		FPS:     fps,
		Codec:   codec,
		Bitrate: rounded(f, "vbr", "tbr"),
		Size:    size,
		Source:  p.resource(url, size),
	}
}

// resource is a stream as the site's CDN wants it fetched. yt-dlp's filesize is exact, so it is trusted.
func (p *profile) resource(url string, size int64) media.Resource {
	return media.Resource{URL: url, Size: size, Policy: p.policy, MaxRange: p.maxRange}
}

// rounded is the first of the keys that is a number, rounded to a whole kbps.
func rounded(f jsonv.Value, keys ...string) int64 {
	for _, k := range keys {
		if v, ok := f.Get(k).Float(); ok {
			return int64(math.Round(v))
		}
	}
	return 0
}

func headerOf(v jsonv.Value) http.Header {
	h := http.Header{}
	for k, val := range v.Object() {
		if s, ok := val.Str(); ok {
			h.Set(k, s)
		}
	}
	return h
}

// subtitles are the uploaded tracks plus the auto-generated one in the spoken language. Machine translations and
// the transcripts of auto-dubbed audio are left out, as is the live chat replay.
func subtitles(n jsonv.Value, header http.Header) []media.Subtitle {
	json3 := func(tracks jsonv.Value) string {
		for _, t := range tracks.Array() {
			if t.Get("ext").String() == "json3" {
				return t.Get("url").String()
			}
		}
		return ""
	}
	base := func(lang string) string { b, _, _ := strings.Cut(lang, "-"); return b }
	sub := func(lang, url string, auto bool) media.Subtitle {
		return media.Subtitle{Lang: lang, Auto: auto, Format: media.YouTubeJSON3, Source: media.Resource{URL: url, Header: header.Clone()}}
	}
	var out []media.Subtitle
	uploaded := n.Get("subtitles")
	for _, lang := range uploaded.Keys() {
		if lang == "live_chat" {
			continue
		}
		if url := json3(uploaded.Get(lang)); url != "" {
			out = append(out, sub(lang, url, false))
		}
	}
	spoken, hasSpoken := n.Get("language").Str()
	auto := n.Get("automatic_captions")
	for _, key := range auto.Keys() {
		lang, ok := strings.CutSuffix(key, "-orig")
		if !ok || (hasSpoken && base(lang) != base(spoken)) {
			continue
		}
		if url := json3(auto.Get(key)); url != "" {
			out = append(out, sub(lang, url, true))
		}
	}
	return out
}

// videoCodec names a vcodec: avc1.64001F, hev1…, av01…, vp09…
func videoCodec(vcodec string) string {
	c := strings.ToLower(vcodec)
	switch {
	case strings.HasPrefix(c, "avc"):
		return "AVC"
	case strings.HasPrefix(c, "hev") || strings.HasPrefix(c, "hvc"):
		return "HEVC"
	case strings.HasPrefix(c, "av01"):
		return "AV1"
	case strings.HasPrefix(c, "vp9") || strings.HasPrefix(c, "vp09"):
		return "VP9"
	}
	return strings.ToUpper(vcodec)
}

// urlCodec is the codec of a format yt-dlp gives none for (X): its file path names it (`/vid/avc1/…`), else it is
// just an MP4.
func urlCodec(url string) string {
	switch {
	case strings.Contains(url, "/avc1/"):
		return "AVC"
	case strings.Contains(url, "/hevc/"):
		return "HEVC"
	}
	return "MP4"
}

func audioCodec(acodec string) string {
	c := strings.ToLower(acodec)
	switch {
	case strings.HasPrefix(c, "mp4a"):
		return "M4A"
	case strings.HasPrefix(c, "opus"):
		return "OPUS"
	case strings.HasPrefix(c, "ec-3"):
		return "E-AC-3"
	case strings.HasPrefix(c, "ac-3"):
		return "AC-3"
	}
	return strings.ToUpper(acodec)
}
