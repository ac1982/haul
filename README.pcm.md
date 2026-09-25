<!-- Translation of README.md; maintenance: docs/TRANSLATING.md -->

<p align="center">
  <img src="docs/assets/brand.svg" alt="haul" width="1200">
</p>

<!-- languages:start -->
<p align="center">
  <a href="README.md"><bdi>English</bdi></a> ·
  <a href="README.zh-CN.md"><bdi>简体中文</bdi></a> ·
  <a href="README.zh-TW.md"><bdi>繁體中文</bdi></a> ·
  <a href="README.ja.md"><bdi>日本語</bdi></a> ·
  <a href="README.hi.md"><bdi>हिन्दी</bdi></a> ·
  <a href="README.es.md"><bdi>Español</bdi></a> ·
  <a href="README.ar.md"><bdi>العربية</bdi></a> ·
  <a href="README.fr.md"><bdi>Français</bdi></a><br>
  <a href="README.bn.md"><bdi>বাংলা</bdi></a> ·
  <a href="README.pt.md"><bdi>Português</bdi></a> ·
  <a href="README.id.md"><bdi>Bahasa Indonesia</bdi></a> ·
  <a href="README.ur.md"><bdi>اردو</bdi></a> ·
  <a href="README.ru.md"><bdi>Русский</bdi></a> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <bdi><strong>Naijá</strong></bdi> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">One link. Your own media.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="Build status"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="Latest release"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 or newer one">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux and Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT licence"></a>
</p>

<p align="center">
  <strong>Command-line tool wey dey download video and audio, wey dem build for people and AI agents together.</strong><br>
  Give am link. Choose your streams. Collect video, audio, subtitles, chapters and cover inside one file.
</p>

<p align="center"><a href="#install">How to install</a> · <a href="#usage">How to start</a> · <a href="#sites">Sites wey e support</a> · <a href="#for-ai-agents-and-scripts">Guide for agents</a> · <a href="#options">Options</a> · <a href="../../releases">Releases</a></p>

---

<a id="small-command-complete-download"></a>

## Small command. Full download.

- **↓ Your media, your way:** Use the same words choose quality, codecs, pages and file names for five sites
- **⌘ One binary for every desktop:** Na one Go binary for macOS, Linux and Windows; e dey download many byte ranges at once, and e fit continue where e stop
- **{ } Ready for automation:** Na one JSON document dey stdout, logs dey stderr, exit codes get meaning, and e no go ask question if terminal no dey

| 01 / CHECK | 02 / CHOOSE | 03 / DOWNLOAD | 04 / JOIN |
| :--- | :--- | :--- | :--- |
| **See every stream** | **Put wetin you want first** | **Carry am come house** | **Keep am together** |
| Pages, codecs and sizes | Quality, audio and pages | Range by range or segment by segment | Tracks, subtitles and cover |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>See wetin dey inside terminal</strong></summary>

```
$ haul info -q 720p "https://youtu.be/DdCEmlAydcw"
haul v1.0.0

  Inside Anthropic's molecular biology lab
  channel Anthropic  ·  2026-09-24  ·  00:01:15

  Video
  ▶  0  720p   1190×720   AVC  24fps   755 kbps    6.8 MB
     1  720p   1190×720   VP9  24fps   528 kbps    4.7 MB
     2  720p   1190×720   AV1  24fps   388 kbps    3.5 MB
     3  1080p  2048×1240  VP9  24fps  5827 kbps   52.3 MB
     …
    19  144p   238×144    AVC  24fps    50 kbps  459.0 KB
  Audio
  ▶ 0  OPUS  131 kbps    1.2 MB
    1  M4A   130 kbps    1.2 MB
    …
```

</details>

> Na for personal use, research and other use wey no be commercial. Na you go make sure say you respect copyright and the terms for each site.

<a id="install"></a>

## How to install

haul dey run for **macOS, Linux and Windows** (amd64 and arm64). E need [ffmpeg](https://ffmpeg.org) to join tracks; YouTube and X still need [yt-dlp](https://github.com/yt-dlp/yt-dlp), and YouTube player checks need [deno](https://deno.com):

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg yt-dlp           # Debian / Ubuntu; deno: https://deno.com
winget install Gyan.FFmpeg               # Windows, one package at a time
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

<a id="get-the-binary"></a><a id="binary"></a>

### Get the binary

Download the archive for your system from [Releases](../../releases), `haul-<version>-<os>-<arch>.tar.gz` (`.zip` for Windows), then put `haul` for any folder wey dey your `PATH`:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
xattr -d com.apple.quarantine ~/.local/bin/haul   # macOS: the binary never get Apple notarization
```

<details>
<summary><strong>Build from source</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

or, from repo wey you clone:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## How to start

Start with one link. Normally, haul go choose the best streams wey dey available.

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

Check first before you download, or choose quality and codec:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC so QuickTime fit play am
```

<details>
<summary><strong>Command guide</strong></summary>

| Command | Wetin e mean |
| --- | --- |
| `haul <url> [options]` | Download; na the same as `haul download <url>` |
| `haul info <url> [--urls]` | Show the item, im pages and streams; e no go download anything |
| `haul login bilibili` | Log in for bilibili to get higher quality |
| `haul templates` | Variables for file-name templates |
| `haul --help` | Sites, how agents fit use am, exit codes and examples |
| `haul download --help` | All the options |

</details>

<a id="sites"></a>

### Sites wey e support

| Site | Links | Wetin e need |
| --- | --- | --- |
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` go choose one video inside the post | yt-dlp |
| **bilibili** | Videos, bangumi, courses, collections, series, favourites, user spaces, `b23.tv`, and bare `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, with `?i=<episode>` for one episode | – |

<a id="make-it-yours"></a><a id="examples"></a>

### Make am your own

**Only the audio**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC instead of Opus
```

**Some parts, one whole season, or the newest episode**

```sh
# Parts wey you choose, save am inside your Movies folder
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# All the episodes for one season
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# The newest podcast episode
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# All the videos inside one X post
haul -p ALL "https://x.com/<user>/status/<id>"
```

**Arranged files, subtitles, or choose by yourself**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # only these subtitle languages
haul -i "BV1qt4y1X7TW"                                 # use arrow keys choose streams
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## For AI agents and scripts

`haul --help` get everything wey agent need; [llms.txt](llms.txt) na the same guide as file. The short version:

```sh
haul info --json "<url>"                                # 1. check
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. download exactly these streams
haul --json -q 720p -c avc,m4a "<url>"                  #    or make priorities choose
```

- With `--json`, **na only one JSON document go dey stdout**; progress and logs go stderr. Im `files` list the output files, new ones and the ones wey don already dey.
- **Nothing dey interactive without terminal.** If terminal no dey, `-i` go fail with exit code 2 and e go talk the flags wey you suppose use instead.
- Stream indexes for `info` follow the order wey haul dey take choose; pass the same `-q` / `-c` to `info` and to the download.
- `info` for playlist, season, show or post wey get many videos go list im pages; `-p <n>` go add the streams and subtitles for that page.
- Pages wey don already dey disk go show as `"status": "skipped", "reason": "exists"`, so e safe to run am again.

<details>
<summary><strong>JSON response example</strong> · download wey work</summary>

Download go print:

```json
{
  "ok": true,
  "command": "download",
  "site": "youtube",
  "input": "https://youtu.be/DdCEmlAydcw",
  "title": "Inside Anthropic's molecular biology lab",
  "uploader": "Anthropic",
  "published": "2026-09-23T18:01:32Z",
  "description": "…",
  "pageCount": 1,
  "files": ["/Users/me/Movies/Inside Anthropic's molecular biology lab.mp4"],
  "pages": [
    {
      "index": 1,
      "id": "DdCEmlAydcw",
      "title": "Inside Anthropic's molecular biology lab",
      "durationSeconds": 75,
      "published": "2026-09-23T18:01:32Z",
      "selected": true,
      "status": "downloaded",
      "file": "/Users/me/Movies/Inside Anthropic's molecular biology lab.mp4",
      "sizeBytes": 3434260,
      "selectedVideo": 0,
      "selectedAudio": 0,
      "video": [{ "index": 0, "quality": "360p", "resolution": "594x360", "codec": "AVC", "fps": 24, "bitrateKbps": 214, "sizeBytes": 2014782 }],
      "audio": [{ "index": 0, "codec": "M4A", "bitrateKbps": 130, "sizeBytes": 1220994 }],
      "subtitles": [{ "lang": "en" }]
    }
  ]
}
```

Na `haul --json -q 360p -c avc,m4a …` be dis. Here the keys dey for the order wey you go read am, and we cut the stream lists to only the ones wey e choose; the real document dey sort im keys from A to Z and e dey list every stream.

</details>

If e fail, e go print `"ok": false` with `"error": {"kind", "message", "exitCode"}`, and e go still keep the pages wey don finish before the failure.

| Exit code | Wetin e mean | `error.kind` |
| --- | --- | --- |
| 0 | E don finish | – |
| 1 | Download or extraction fail | `failed` |
| 2 | Link no get support or option value no correct | `input` |
| 3 | Tool wey e need no dey (ffmpeg, yt-dlp) | `dependency` |
| 4 | You need log in or login don expire | `auth` |
| 64 | Command line no correct (unknown flag, missing argument) | – |
| 130 | Dem cancel am | `cancelled` |

<a id="options"></a>

## Options

`haul download --help` go list all of them. The main ones:

| Group | Options |
| --- | --- |
| General | `--json`, `--config <file>`, `--debug` |
| Streams | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Pages | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Content | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| Output | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; more dey `--help-hidden` |
| Tools | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>Quality and codec priority</strong></summary>

**Qualities and codecs.** `-q` dey take the labels wey stream table show: `1080p`, `720p60` for YouTube; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` for bilibili. `-c` dey take `av1 vp9 hevc avc` for video and `m4a opus flac eac3 mp3` for audio. Dem be priorities, no be filters: anything wey you no list go follow after, best one first. `--video-ascending` / `--audio-ascending` go turn the order around so you get the smallest download.

</details>

<details>
<summary><strong>Choose pages</strong></summary>

**Pages.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (the last page, wey be the newest episode for one show; `LATEST` go work too). Link to one episode, or `?p=N`, go choose only that page. Pages na the parts of bilibili video, the episodes for season, show or list, and the videos inside X post.

</details>

<details>
<summary><strong>File names and template variables</strong></summary>

**File names.** Item wey get one page: `<title>`; if e get many pages (even when `-p` choose only one): `<title>/[P<pageNumberWithZero>]<pageTitle>`, with zeros for front of the number according to how many pages dey. Variables: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, plus bilibili own: `<bvid>` `<aid>` `<cid>` `<api>`. Dates fit take format: `<publishDate:yyyy-MM-dd>`. E go add the extension by imself.

</details>

<a id="site-notes"></a><a id="notes"></a>

## Things to know for each site

<details>
<summary><strong>YouTube</strong> · extraction, subtitles and playlists</summary>

YouTube dey protect im stream URLs with player checks wey need JavaScript runtime, so na `yt-dlp -J` dey do the extraction (e dey run the checks inside deno); everything wey follow after na haul own work. E go join uploaded subtitles inside; automatic ones only if you use `--auto-subtitles`. googlevideo dey give only limited byte ranges, so e dey fetch tracks one 10 MB range at a time. `watch?v=…&list=…` go download only that video. Keep yt-dlp fresh: na there fixes for YouTube changes dey land.

</details>

<details>
<summary><strong>X</strong> · public posts and audio</summary>

Public posts for X no need login. Their videos na single MP4 files wey get audio inside, so the table show only video; `--audio-only` go bring out the audio.

</details>

<details>
<summary><strong>bilibili</strong> · login, higher quality and danmaku</summary>

haul dey read bilibili through im own web, TV, APP (gRPC) and international APIs. If you no log in, e go give only lower quality (usually reach 480P); log in for 1080P, 4K, HDR, Dolby Vision and Hi-Res audio:

```sh
haul login bilibili                # scan QR code with bilibili app
haul login bilibili --from-edge    # macOS: use the login wey dey Microsoft Edge (or --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # TV access token, for --api tv / --api app
```

For now, to read browser login dey work only for macOS: the terminal need Full Disk Access, and macOS go ask one time for the "Safe Storage" keychain item. `--danmaku` go save the comments wey dey fly across the video (bullet comments) as XML and ASS; `--danmaku-format ass` go keep only one of them.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · episodes and metadata</summary>

Xiaoyuzhou and Apple Podcasts no need yt-dlp: Xiaoyuzhou pages carry the audio link, Apple links dey pass through the public iTunes API and the show RSS feed. Episodes keep their format (`.mp3` or `.m4a`) with the cover inside, show name as album and host as artist. Show link go list im latest episodes from oldest to newest, so `-p LAST` na the newest.

</details>

<a id="config"></a>

## Settings

`~/.config/haul/` (or `$HAUL_HOME`) dey hold:

| File | Wetin e do |
| --- | --- |
| `config.json` | Default values for any option |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili login |
| `archives.txt` | Pages wey you don download before (`--archive`) |

`config.json` dey use option names for camelCase, bilibili own dey under `"bilibili"`, and e need only the ones wey you wan change; one list fit be array or string wey comma separate. Command-line flags go override am:

```json
{
  "workDir": "~/Movies/haul",
  "codec": "avc,hevc,m4a",
  "quality": "1080p,1080P",
  "output": "<uploader>/<title>",
  "archive": true,
  "bilibili": { "danmaku": true, "danmakuFormat": ["ass"] }
}
```

<a id="development"></a>

## Development

```sh
go build ./cmd/haul
go test ./...                                  # offline, few seconds
HAUL_LIVE=1 go test -run Live ./internal/...   # e go still reach bilibili, YouTube, X and Apple Podcasts
```

The offline tests no dey touch network at all: tests dey talk to stub `http.RoundTripper` wey dey answer from recorded API responses and from simulated CDNs (servers wey accept only ranges, connections wey cut, servers wey ignore ranges). End-to-end tests dey join tracks with real ffmpeg and check the result with ffprobe; dem go skip if ffmpeg no dey installed. Check [AGENTS.md](AGENTS.md) for the architecture and rules.

If you push `v*` tag, GitHub Actions go test, cross-compile and publish release for every platform.

<a id="acknowledgements"></a>

## People and projects we thank

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr), and the API notes for [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) and [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## Licence

[MIT](LICENSE)

---

<p align="center">
  <strong>One link. Your own media.</strong><br>
  <a href="#install">Get haul</a> · <a href="llms.txt">Reference for agents</a> · <a href="AGENTS.md">Join hand</a>
</p>
