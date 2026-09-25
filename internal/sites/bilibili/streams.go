package bilibili

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// qualityNames are bilibili's quality ids, best first.
var qualityNames = map[string]string{
	"127": "8K", "126": "Dolby Vision", "125": "HDR", "120": "4K", "116": "1080P60",
	"112": "1080P+", "100": "AI Enhanced", "80": "1080P", "74": "720P60",
	"64": "720P", "48": "720P", "32": "480P", "16": "360P", "5": "144P", "6": "240P",
}

func qualityName(id string) string {
	if name, ok := qualityNames[id]; ok {
		return name
	}
	return "Unknown (" + id + ")"
}

func videoCodec(codecid string) string {
	switch codecid {
	case "7":
		return "AVC"
	case "12":
		return "HEVC"
	case "13":
		return "AV1"
	}
	return "UNKNOWN"
}

func audioCodec(codecs string) string {
	switch codecs {
	case "mp4a.40.2", "mp4a.40.5":
		return "M4A"
	case "ec-3":
		return "E-AC-3"
	case "fLaC":
		return "FLAC"
	}
	return codecs
}

var pcdnURL = regexp.MustCompile(`http.*:\d+`)

// pickURL prefers a URL that is not on a bare ip:port (PCDN) host.
func pickURL(node jsonv.Value) string {
	urls := []string{node.Get("base_url").String()}
	for _, u := range node.Get("backup_url").Array() {
		urls = append(urls, u.String())
	}
	for _, u := range urls {
		if !pcdnURL.MatchString(u) {
			return u
		}
	}
	return urls[0]
}

func videoFormat(id string, node jsonv.Value, bitrate int64) media.VideoFormat {
	return media.VideoFormat{
		ID:      id,
		Rank:    jsonv.Of(id).IntOr(0),
		Quality: qualityName(id),
		Codec:   videoCodec(node.Get("codecid").String()),
		Bitrate: bitrate,
		Size:    node.Get("size").Int64Or(0),
		Source:  media.Resource{URL: pickURL(node)},
	}
}

func audioFormat(node jsonv.Value, codec string) media.AudioFormat {
	if codec == "" {
		codec = audioCodec(node.Get("codecs").String())
	}
	return media.AudioFormat{
		ID:      node.Get("id").String(),
		Codec:   codec,
		Bitrate: node.Get("bandwidth").Int64Or(0) / 1000,
		Size:    node.Get("size").Int64Or(0),
		Source:  media.Resource{URL: pickURL(node)},
	}
}

// collector gathers formats over several answers, keeping the first of streams that look the same.
type collector struct {
	f    *media.Formats
	seen map[string]bool
}

func (c *collector) video(v media.VideoFormat) {
	if c.once(fmt.Sprintf("v|%s|%s|%d|%dx%d@%g", v.ID, v.Codec, v.Bitrate, v.Width, v.Height, v.FPS)) {
		c.f.Video = append(c.f.Video, v)
	}
}

func (c *collector) audio(a media.AudioFormat) {
	if c.once(fmt.Sprintf("a|%s|%s|%d", a.ID, a.Codec, a.Bitrate)) {
		c.f.Audio = append(c.f.Audio, a)
	}
}

func (c *collector) once(key string) bool {
	if c.seen[key] {
		return false
	}
	c.seen[key] = true
	return true
}

// streams asks the entry's API for its streams and turns the answer into formats with prepared resources.
func (e *Extractor) streams(ctx context.Context, s *session, r *ref) (*media.Formats, error) {
	raw, err := e.playJSON(ctx, s, r, "0")
	if err != nil {
		return nil, err
	}
	c := &collector{f: &media.Formats{}, seen: map[string]bool{}}
	switch {
	case strings.Contains(raw, `"stream_list"`):
		raw, err = e.intlStreams(ctx, s, r, raw, c)
	case strings.Contains(raw, `"dash":{`):
		raw, err = e.dashStreams(ctx, s, r, raw, c)
	case strings.Contains(raw, `"durl":[`):
		// FLV: segments, at the highest quality the account can see.
		if raw, err = e.playJSON(ctx, s, r, maxQuality); err == nil {
			err = flvStreams(raw, c)
		}
	}
	if err != nil {
		return nil, err
	}
	f := c.f
	f.Raw = raw
	if r.bangumi() {
		if root, err := jsonv.ParseString(raw); err == nil {
			f.Chapters = openingAndEnding(payloadRoot(root, raw).Get("clip_info_list"))
		}
	}
	e.prepare(s, f)
	return f, nil
}

// intlStreams reads the international API's stream_list, once per codec preference.
func (e *Extractor) intlStreams(ctx context.Context, s *session, r *ref, raw string, c *collector) (string, error) {
	for _, code := range []string{"0", "1"} {
		if code == "1" {
			var err error
			if raw, err = e.intlJSON(ctx, s, r, "0", code); err != nil {
				return "", err
			}
		}
		root, err := jsonv.ParseString(raw)
		if err != nil {
			return "", err
		}
		vi := root.Path("data", "video_info")
		for _, stream := range vi.Get("stream_list").Array() {
			dv := stream.Get("dash_video")
			if !dv.Exists() || dv.Get("base_url").String() == "" {
				continue
			}
			c.video(videoFormat(stream.Path("stream_info", "quality").String(), dv, dv.Get("bandwidth").Int64Or(0)/1000))
		}
		for _, node := range vi.Get("dash_audio").Array() {
			c.audio(audioFormat(node, "M4A"))
		}
	}
	return raw, nil
}

// dashStreams reads a DASH answer. The web and TV APIs are asked twice, with the default and then the maximum
// quality: the second answer surfaces streams the first hid (the "no re-encode" originals). The APP answer
// already has everything.
func (e *Extractor) dashStreams(ctx context.Context, s *session, r *ref, raw string, c *collector) (string, error) {
	passes := 2
	if r.api == App {
		passes = 1
	}
	var root jsonv.Value
	for pass := range passes {
		if pass == 1 {
			var err error
			if raw, err = e.playJSON(ctx, s, r, maxQuality); err != nil {
				return "", err
			}
		}
		var err error
		if root, err = jsonv.ParseString(raw); err != nil {
			return "", err
		}
		dash := payloadRoot(root, raw).Get("dash")
		for _, node := range dash.Get("video").Array() {
			v := videoFormat(node.Get("id").String(), node, node.Get("bandwidth").Int64Or(0)/1000)
			if r.api == Web || r.api == Intl {
				v.Width = node.Get("width").IntOr(0)
				v.Height = node.Get("height").IntOr(0)
				v.FPS = node.Get("frame_rate").FloatOr(0)
			}
			c.video(v)
		}
		audio := dash.Get("audio").Array()
		if r.api != TV {
			audio = append(audio, dash.Path("dolby", "audio").Array()...)
			if flac := dash.Path("flac", "audio"); flac.IsObject() {
				audio = append(audio, flac)
			}
		}
		for _, node := range audio {
			c.audio(audioFormat(node, ""))
		}
	}
	if r.api == App && r.bangumi() {
		c.f.ExtraAudio = dubbing(root.Get("dubbing_info"))
	}
	return raw, nil
}

// dubbing turns the APP API's dubbing info into extra tracks: the background, then each voice role, each with its
// best stream. A background without roles (or the reverse) is no dubbing.
func dubbing(info jsonv.Value) []media.ExtraAudio {
	background := info.Get("background_audio").Array()
	roles := info.Get("role_audio_list").Array()
	if len(background) == 0 || len(roles) == 0 {
		return nil
	}
	best := func(nodes []jsonv.Value) (media.AudioFormat, bool) {
		var out []media.AudioFormat
		for _, n := range nodes {
			out = append(out, audioFormat(n, ""))
		}
		if len(out) == 0 {
			return media.AudioFormat{}, false
		}
		return slices.MaxFunc(out, func(a, b media.AudioFormat) int { return int(a.Bitrate - b.Bitrate) }), true
	}
	var extra []media.ExtraAudio
	if a, ok := best(background); ok {
		extra = append(extra, media.ExtraAudio{Title: "Background audio", Audio: a})
	}
	for _, role := range roles {
		if a, ok := best(role.Get("audio").Array()); ok {
			extra = append(extra, media.ExtraAudio{Title: role.Get("title").String(), Artist: role.Get("person_name").String(), Audio: a})
		}
	}
	return extra
}

// flvStreams reads a durl answer: one format made of segments, with the audio inside.
func flvStreams(raw string, c *collector) error {
	root, err := jsonv.ParseString(raw)
	if err != nil {
		return err
	}
	payload := payloadRoot(root, raw)
	var parts []media.Resource
	var size, length int64
	for _, node := range payload.Get("durl").Array() {
		parts = append(parts, media.Resource{URL: node.Get("url").String()})
		size += node.Get("size").Int64Or(0)
		length += node.Get("length").Int64Or(0)
	}
	quality := payload.Get("quality").String()
	v := media.VideoFormat{
		ID:       quality,
		Rank:     jsonv.Of(quality).IntOr(0),
		Quality:  qualityName(quality),
		Codec:    videoCodec(payload.Get("video_codecid").String()),
		Size:     size,
		HasAudio: true,
		Parts:    parts,
	}
	if length > 0 {
		v.Bitrate = size * 8 / length // bytes per millisecond × 8 = kbps
	}
	c.video(v)
	return nil
}

// openingAndEnding turns bangumi opening / ending markers into chapters, with "Main" filling the gaps before them.
func openingAndEnding(clips jsonv.Value) []media.Chapter {
	type span struct {
		title      string
		start, end int
	}
	var spans []span
	for _, c := range clips.Array() {
		spans = append(spans, span{strings.ReplaceAll(c.Get("toastText").String(), "即将跳过", ""), c.Get("start").IntOr(0), c.Get("end").IntOr(0)})
	}
	slices.SortStableFunc(spans, func(a, b span) int { return a.start - b.start })
	sec := func(n int) time.Duration { return time.Duration(n) * time.Second }
	var chapters []media.Chapter
	lastEnd := 0
	for _, s := range spans {
		if lastEnd < s.start {
			chapters = append(chapters, media.Chapter{Title: "Main", Start: sec(lastEnd), End: sec(s.start)})
		}
		chapters = append(chapters, media.Chapter{Title: s.title, Start: sec(s.start), End: sec(s.end)})
		lastEnd = s.end
	}
	return chapters
}

// viewPoints are the chapters an uploader set, from the player info.
func viewPoints(player jsonv.Value) []media.Chapter {
	var chapters []media.Chapter
	for _, p := range player.Path("data", "view_points").Array() {
		chapters = append(chapters, media.Chapter{
			Title: p.Get("content").String(),
			Start: seconds(p.Get("from")),
			End:   seconds(p.Get("to")),
		})
	}
	return chapters
}

// playerInfo is x/player/wbi/v2: chapters and, for a logged-in cookie, subtitles. Null when it fails; both are
// extras.
func (e *Extractor) playerInfo(ctx context.Context, s *session, r *ref) jsonv.Value {
	j, err := e.getJSON(ctx, s, "https://api.bilibili.com/x/player/wbi/v2?cid="+r.cid+"&aid="+r.aid)
	if err != nil {
		return jsonv.Null
	}
	return j
}
