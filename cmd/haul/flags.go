package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// kind is a flag's value type.
type kind int

const (
	boolFlag kind = iota
	stringFlag
	intFlag
)

// flagDef is one option: how it parses, how --help shows it, and its key in config.json.
type flagDef struct {
	name   string
	short  string
	kind   kind
	value  string // placeholder shown in help, e.g. <list>
	def    string // default for string / int flags
	group  string
	help   string
	detail string // a second help line
	hidden bool   // shown by --help-hidden only
	// config is the key in config.json ("" = the name in camelCase); "-" = not in the config.
	config string
}

// Help groups in the order --help shows them.
var groups = []string{"General", "Streams", "Pages", "Content", "Output", "bilibili", "Tools"}

// downloadFlags are the options of download and info.
var downloadFlags = []flagDef{
	{name: "json", kind: boolFlag, group: "General", config: "-", help: "Print one JSON document on stdout; logs go to stderr.",
		detail: "The item, its pages, their streams and, for a download, the output files (top-level `files`)."},
	{name: "config", kind: stringFlag, value: "<file>", group: "General", config: "-", help: "Config file.", detail: "Default: ~/.config/haul/config.json ($HAUL_HOME)."},
	{name: "debug", kind: boolFlag, group: "General", help: "Debug log: timestamps, requests, tool command lines."},

	{name: "quality", short: "q", kind: stringFlag, value: "<list>", group: "Streams", help: "Quality priority, comma separated, as the stream table labels them.",
		detail: `e.g. "1080p,720p" (YouTube), "4K,1080P+,1080P" (bilibili). Unlisted qualities follow, best first.`},
	{name: "codec", short: "c", kind: stringFlag, value: "<list>", group: "Streams", help: "Codec priority, comma separated.",
		detail: "Video: av1 vp9 hevc avc. Audio: m4a opus flac eac3 mp3."},
	{name: "video-stream", kind: intFlag, value: "<n>", def: "-1", group: "Streams", config: "-", help: "Take this video stream: its index in `haul info`.",
		detail: "Indexes follow the order -q, -c and --*-ascending give: pass the same ones to info and download."},
	{name: "audio-stream", kind: intFlag, value: "<n>", def: "-1", group: "Streams", config: "-", help: "Take this audio stream: its index in `haul info`."},
	{name: "interactive", short: "i", kind: boolFlag, group: "Streams", config: "-", help: "Choose the streams with the arrow keys (needs a terminal)."},
	{name: "video-ascending", kind: boolFlag, group: "Streams", help: "Prefer the lowest quality and the smallest video stream."},
	{name: "audio-ascending", kind: boolFlag, group: "Streams", help: "Prefer the smallest audio stream."},

	{name: "pages", short: "p", kind: stringFlag, value: "<spec>", group: "Pages", config: "-", help: "Pages / episodes / videos of a post to take: 8, 1,2, 3-5, ALL, LAST.",
		detail: "1-based, as `info` numbers them; LAST is the last page, the newest episode of a show. Default: the page the link points at, else all (info: the list of pages only)."},
	{name: "show-all", kind: boolFlag, group: "Pages", help: "List every page and every stream instead of a summary."},
	{name: "hide-streams", kind: boolFlag, group: "Pages", help: "Do not print the stream table."},

	{name: "audio-only", kind: boolFlag, group: "Content", help: "Audio only, as .m4a.", detail: "The best audio stream, which may be Opus; -c m4a gets AAC. Podcasts keep .mp3 / .m4a."},
	{name: "video-only", kind: boolFlag, group: "Content", help: "Video only, no audio."},
	{name: "subtitle-only", kind: boolFlag, group: "Content", help: "Only the subtitles, as .srt files next to where the video would be."},
	{name: "cover-only", kind: boolFlag, group: "Content", help: "Only the cover image."},
	{name: "skip-subtitle", kind: boolFlag, group: "Content", help: "Do not embed subtitles."},
	{name: "sub-lang", kind: stringFlag, value: "<list>", group: "Content", help: "Only these subtitle languages, comma separated, e.g. en,zh.",
		detail: "A prefix matches: en takes en-US. `haul info` lists them."},
	{name: "auto-subtitles", kind: boolFlag, group: "Content", help: "Also take auto-generated subtitles."},
	{name: "skip-cover", kind: boolFlag, group: "Content", help: "Do not embed the cover."},
	{name: "skip-mux", kind: boolFlag, group: "Content", help: "Keep the downloaded streams as they are, without muxing."},

	{name: "output", short: "o", kind: stringFlag, value: "<template>", group: "Output", help: "File-name template for an item with one page.",
		detail: "Default: <title>. Variables: `haul templates`. No extension: haul adds it."},
	{name: "multi-output", kind: stringFlag, value: "<template>", group: "Output", help: "File-name template for an item with several pages, even when -p takes one.",
		detail: "Default: <title>/[P<pageNumberWithZero>]<pageTitle>"},
	{name: "work-dir", short: "w", kind: stringFlag, value: "<dir>", group: "Output", help: "Directory to download into. Default: the current directory."},
	{name: "lang", kind: stringFlag, value: "<code>", group: "Output", help: "Audio language code written into the file, e.g. eng, jpn, chi."},
	{name: "no-tags", kind: boolFlag, group: "Output", help: "Do not write title, artist, description and date tags."},
	{name: "archive", kind: boolFlag, group: "Output", help: "Remember downloaded pages in the archive and skip them next time."},
	{name: "delay", kind: intFlag, value: "<seconds>", def: "0", group: "Output", help: "Seconds to wait between pages."},

	{name: "api", kind: stringFlag, value: "<api>", def: "web", group: "bilibili", config: "bilibili.api", help: "bilibili API: web, tv, app or intl (bilibili.tv)."},
	{name: "danmaku", kind: boolFlag, group: "bilibili", config: "bilibili.danmaku", help: "Also save the danmaku (bullet comments)."},
	{name: "danmaku-only", kind: boolFlag, group: "bilibili", config: "-", help: "Only the danmaku."},
	{name: "danmaku-format", kind: stringFlag, value: "<list>", def: "xml,ass", group: "bilibili", config: "bilibili.danmakuFormat", help: "Danmaku files to write with --danmaku / --danmaku-only: xml, ass."},
	{name: "cookie", kind: stringFlag, value: "<cookie>", group: "bilibili", config: "bilibili.cookie", help: "Web cookie (SESSDATA=…) instead of the stored login."},
	{name: "token", kind: stringFlag, value: "<token>", group: "bilibili", config: "bilibili.token", help: "TV / APP access token instead of the stored login."},
	{name: "user-agent", kind: stringFlag, value: "<ua>", group: "bilibili", config: "bilibili.userAgent", hidden: true, help: "User-Agent for bilibili API requests."},
	{name: "upos-host", kind: stringFlag, value: "<host>", group: "bilibili", config: "bilibili.uposHost", hidden: true, help: "Download from this upos CDN host."},
	{name: "no-replace-host", kind: boolFlag, group: "bilibili", config: "bilibili.noReplaceHost", hidden: true, help: "Keep the CDN host bilibili gives instead of a known-good mirror."},
	{name: "allow-pcdn", kind: boolFlag, group: "bilibili", config: "bilibili.allowPcdn", hidden: true, help: "Keep PCDN hosts."},
	{name: "no-force-http", kind: boolFlag, group: "bilibili", config: "bilibili.noForceHttp", hidden: true, help: "Fetch media over HTTPS rather than HTTP."},
	{name: "host", kind: stringFlag, value: "<host>", group: "bilibili", config: "bilibili.host", hidden: true, help: "BiliPlus-style API proxy host (needs a token)."},
	{name: "ep-host", kind: stringFlag, value: "<host>", group: "bilibili", config: "bilibili.epHost", hidden: true, help: "Proxy host for /pgc/view/web/season."},
	{name: "tv-host", kind: stringFlag, value: "<host>", group: "bilibili", config: "bilibili.tvHost", hidden: true, help: "TV API host."},
	{name: "area", kind: stringFlag, value: "<area>", group: "bilibili", config: "bilibili.area", hidden: true, help: "Proxy area: hk, tw or th."},

	{name: "ffmpeg", kind: stringFlag, value: "<path>", group: "Tools", help: "ffmpeg to use. Default: from PATH."},
	{name: "yt-dlp", kind: stringFlag, value: "<path>", group: "Tools", config: "ytDlp", help: "yt-dlp to use (YouTube, X). Default: from PATH."},
	{name: "use-mp4box", kind: boolFlag, group: "Tools", help: "Mux with MP4Box instead of ffmpeg."},
	{name: "mp4box", kind: stringFlag, value: "<path>", group: "Tools", help: "MP4Box to use."},
	{name: "use-aria2c", kind: boolFlag, group: "Tools", help: "Download with aria2c."},
	{name: "aria2c", kind: stringFlag, value: "<path>", group: "Tools", help: "aria2c to use."},
	{name: "aria2c-args", kind: stringFlag, value: "<args>", group: "Tools", help: "Extra aria2c arguments (on top of -x16 -s16 -j16 -k5M)."},
	{name: "single-connection", kind: boolFlag, group: "Tools", help: "Download over one connection instead of 16 parallel ranges."},
}

// infoFlags are download's plus --urls.
var infoFlags = append([]flagDef{{name: "urls", kind: boolFlag, group: "General", config: "-", help: "Include each stream's download URL."}}, downloadFlags...)

var loginFlags = []flagDef{
	{name: "tv", kind: boolFlag, group: "Options", help: "Log in the TV account and save its access token (for --api tv / app)."},
	{name: "from-edge", kind: boolFlag, group: "Options", help: "Copy the login from Microsoft Edge."},
	{name: "from-chrome", kind: boolFlag, group: "Options", help: "Copy the login from Google Chrome."},
	{name: "profile", kind: stringFlag, value: "<name>", def: "Default", group: "Options", help: `Browser profile, e.g. "Profile 1".`},
	{name: "debug", kind: boolFlag, group: "Options", help: "Debug log."},
}

// newFlagSet registers defs on a pflag set that reports errors instead of exiting.
func newFlagSet(name string, defs []flagDef) *pflag.FlagSet {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.SetInterspersed(true)
	fs.Usage = func() {}
	fs.SetOutput(discard{})
	for _, d := range defs {
		switch d.kind {
		case boolFlag:
			fs.BoolP(d.name, d.short, false, d.help)
		case intFlag:
			var def int
			fmt.Sscan(d.def, &def)
			fs.IntP(d.name, d.short, def, d.help)
		default:
			fs.StringP(d.name, d.short, d.def, d.help)
		}
	}
	fs.BoolP("help", "h", false, "Show help.")
	fs.Bool("help-hidden", false, "Show help, including hidden options.")
	return fs
}

type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }

// configKey is where a flag lives in config.json.
func (d flagDef) configKey() string {
	if d.config != "" {
		return d.config
	}
	parts := strings.Split(d.name, "-")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

// flagHelp renders the options of defs by group, wrapped for an 80-column terminal.
func flagHelp(defs []flagDef, withHidden bool) string {
	var b strings.Builder
	order := append(append([]string{}, groups...), "Options")
	for _, g := range order {
		var lines []string
		for _, d := range defs {
			if d.group != g || d.hidden && !withHidden {
				continue
			}
			lines = append(lines, flagLines(d)...)
		}
		if len(lines) == 0 {
			continue
		}
		heading := strings.ToUpper(g)
		fmt.Fprintf(&b, "\n%s:\n%s\n", heading, strings.Join(lines, "\n"))
	}
	fmt.Fprintf(&b, "\nHELP:\n%s\n", strings.Join(append(flagLines(flagDef{name: "help", short: "h", help: "Show help."}),
		flagLines(flagDef{name: "help-hidden", help: "Show help with the rarely needed bilibili network options."})...), "\n"))
	return b.String()
}

const helpColumn = 26

func flagLines(d flagDef) []string {
	left := "  "
	if d.short != "" {
		left += "-" + d.short + ", "
	}
	left += "--" + d.name
	if d.value != "" {
		left += " " + d.value
	}
	text := d.help
	if d.kind != boolFlag && d.def != "" && d.def != "-1" && d.def != "0" {
		text += " Default: " + d.def + "."
	}
	var lines []string
	wrapped := wrap(text, 80-helpColumn)
	if len(left) >= helpColumn-1 {
		lines = append(lines, left)
		for _, w := range wrapped {
			lines = append(lines, strings.Repeat(" ", helpColumn)+w)
		}
	} else {
		for i, w := range wrapped {
			if i == 0 {
				lines = append(lines, left+strings.Repeat(" ", helpColumn-len(left))+w)
			} else {
				lines = append(lines, strings.Repeat(" ", helpColumn)+w)
			}
		}
	}
	if d.detail != "" {
		for _, w := range wrap(d.detail, 80-8) {
			lines = append(lines, "        "+w)
		}
	}
	return lines
}

// wrap breaks text into lines of at most width runes, at spaces.
func wrap(text string, width int) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		if line != "" && len([]rune(line))+1+len([]rune(word)) > width {
			lines = append(lines, line)
			line = word
			continue
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
