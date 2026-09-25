package main

import (
	"fmt"
	"strings"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/engine"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/sites/bilibili"
	"github.com/ac1982/haul/internal/sites/podcast"
	"github.com/ac1982/haul/internal/sites/ytdlp"
	"github.com/ac1982/haul/internal/storage"
)

// sites describes every supported site, for help texts.
func sites() []extract.Info {
	c := httpx.Default
	var out []extract.Info
	for _, x := range []extract.Extractor{ytdlp.NewYouTube(c, ytdlp.Options{}), ytdlp.NewX(c, ytdlp.Options{}),
		bilibili.New(c, bilibili.DefaultOptions()), podcast.NewXiaoyuzhou(c), podcast.NewApple(c)} {
		out = append(out, x.Info())
	}
	return out
}

func rootHelp() string {
	var siteLines []string
	for _, s := range sites() {
		siteLines = append(siteLines, "  "+s.Name, "      "+s.Links)
	}
	home := console.PrettyPath(storage.Home())
	return `OVERVIEW: Download video and audio from YouTube, X, bilibili, Xiaoyuzhou and
Apple Podcasts.

USAGE: haul <url> [options]  |  haul <command> [options]

COMMANDS:
  download (default)      Download a link. ` + "`haul <url>`" + ` is the same.
  info                    Show the item, its pages and their streams; download
                          nothing.
  login                   Log in to a site. bilibili: higher qualities and
                          members-only content.
  templates               List the variables of -o / --multi-output file-name
                          templates.
  help <command>          Show a command's options.

SITES
` + strings.Join(siteLines, "\n") + `
  YouTube and X need yt-dlp (and deno for YouTube). Muxing needs ffmpeg.

FOR SCRIPTS AND AI AGENTS
  haul info <url> --json   inspect: the item and its pages; for one
                           page (or -p N) its streams and subtitles
  haul <url> --json        download; ` + "`files`" + ` lists the output files
  With --json, stdout carries only that document; progress and logs
  go to stderr. haul never prompts without a terminal: choose streams
  with --video-stream N / --audio-stream N (indexes from info, with
  the same -q / -c), or let -q / -c decide. Re-runs skip files that
  exist. A list saves into a folder named after it: read the path
  from the JSON, do not build it.

EXIT CODES
  0 done              1 download failed    2 bad link or option
  3 missing tool      4 login needed       64 bad command line
  130 cancelled

EXAMPLES
  haul "https://youtu.be/DdCEmlAydcw"
  haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"
  haul --audio-only "https://youtu.be/DdCEmlAydcw"
  haul -p ALL "https://x.com/<user>/status/<id>"
  haul -p 1-3 -w ~/Movies "https://www.bilibili.com/video/BV1qt4y1X7TW"
  haul -p LAST "https://podcasts.apple.com/us/podcast/id1200361736"
  haul info --json --urls "https://youtu.be/DdCEmlAydcw"

FILES
  ` + home + `/ (or $HAUL_HOME) holds config.json (defaults for any
  option, camelCase keys, bilibili's under "bilibili"), the bilibili
  login and the archive. Files are saved in the current directory,
  or -w <dir>.

  ` + "`haul download --help`" + ` lists every option; ` + "`--help-hidden`" + ` adds
  the rarely needed bilibili network options.

  haul ` + version + `
`
}

func usageLine(command string) string {
	switch command {
	case "info":
		return "haul info <url> [options]"
	case "login":
		return "haul login <site> [--tv | --from-edge | --from-chrome [--profile <name>]]"
	case "templates":
		return "haul templates"
	}
	return "haul [download] <url> [options]"
}

func commandHelp(command string, hidden bool) string {
	var b strings.Builder
	switch command {
	case "info":
		b.WriteString(`OVERVIEW: Show the item, its pages and their streams; download nothing.

Streams are listed in the order haul would choose them under the same -q / -c /
--*-ascending options; the index of each is what --video-stream /
--audio-stream take, so pass the same options to both. For an item with several
pages (a playlist, season, show, multi-video post) info lists the pages only;
-p <n> lists the streams and subtitles of page n, -p ALL of every page.
Download options are accepted and ignored. With --json, stdout gets one JSON
document.
`)
	case "login":
		b.WriteString(`OVERVIEW: Log in to a site. bilibili: higher qualities and members-only content.

  haul login bilibili              scan a QR code with the bilibili app
  haul login bilibili --from-edge  reuse Microsoft Edge's login
                                   (--from-chrome; --profile "Profile 1")
  haul login bilibili --tv         a TV access token (--api tv / app)

Reading a browser's login needs Full Disk Access for the terminal; macOS asks
once for the keychain. The login is saved under ` + console.PrettyPath(storage.Home()) + `/.
`)
	case "templates":
		return templatesHelp()
	default:
		b.WriteString(`OVERVIEW: Download a link. The default: ` + "`haul <url>`" + ` is the same.

Picks the best streams (or those -q / -c / --video-stream / --audio-stream ask
for), downloads them and muxes video, audio, subtitles, chapters and cover into
one file. Files that already exist are skipped. With --json, stdout gets one
JSON document: every page, the streams chosen and the files written.
`)
	}
	fmt.Fprintf(&b, "\nUSAGE: %s\n", usageLine(command))
	switch command {
	case "login":
		b.WriteString("\nARGUMENTS:\n  <site>                  The site: bilibili.\n")
		b.WriteString(flagHelp(loginFlags, false))
	case "info":
		b.WriteString("\nARGUMENTS:\n  <url>                   A YouTube, X, bilibili, Xiaoyuzhou or Apple Podcasts\n                          link, or a bilibili id (BV…, av…, ep…, ss…).\n")
		b.WriteString(flagHelp(infoFlags, hidden))
	default:
		b.WriteString("\nARGUMENTS:\n  <url>                   A YouTube, X, bilibili, Xiaoyuzhou or Apple Podcasts\n                          link, or a bilibili id (BV…, av…, ep…, ss…).\n")
		b.WriteString(flagHelp(downloadFlags, hidden))
	}
	return b.String()
}

func templatesHelp() string {
	var b strings.Builder
	b.WriteString("File-name template variables (-o, --multi-output):\n")
	for _, v := range engine.TemplateVariables {
		fmt.Fprintf(&b, "  %-22s%s\n", "<"+v[0]+">", v[1])
	}
	b.WriteString("  <publishDate:yyyy-MM-dd>  a date in any format; also <pageDate:…>\n\n")
	fmt.Fprintf(&b, "Single page default:   %s\nSeveral pages default: %s\n", engine.DefaultTemplate, engine.DefaultListTemplate)
	b.WriteString("The extension (.mp4, .m4a, .mp3) is added; a / makes folders.\n")
	return b.String()
}
