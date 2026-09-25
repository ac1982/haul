# haul

A command-line downloader for video and audio, built for people and AI agents alike. Give it a link from YouTube, X, bilibili, Xiaoyuzhou or Apple Podcasts: it picks the best streams, downloads them over several connections, and muxes video, audio, subtitles, chapters and cover into one file.

- **Agent-friendly.** `--json` prints one JSON document on stdout (the item, its pages, every stream with an index, the files written) while logs go to stderr. It never prompts without a terminal, and exit codes say what went wrong.
- **One tool, one vocabulary.** The same stream table, `-q` / `-c` priorities, page selection, file-name templates and config file for every site.
- **Native.** A single Swift binary for macOS on Apple Silicon.

```
$ haul info "https://youtu.be/DdCEmlAydcw"
haul v1.0.0

  Inside Anthropic's molecular biology lab
  channel Anthropic  ·  2026-09-24  ·  00:01:15

  Video
  ▶  0  1080p  2048×1240  VP9  24fps  5827 kbps   52.3 MB
     1  1080p  2048×1240  AV1  24fps  4241 kbps   38.1 MB
     2  1080p  1784×1080  AVC  24fps  2807 kbps   25.2 MB
     …
    19  144p   238×144    AVC  24fps    50 kbps  459.0 KB
  Audio
  ▶ 0  OPUS  131 kbps    1.2 MB
    1  M4A   130 kbps    1.2 MB
    …
```

> For personal, research and other non-commercial use. You are responsible for respecting copyright and each site's terms.

## Install

Requirements: macOS 15+ on Apple Silicon, and [ffmpeg](https://ffmpeg.org) for muxing. YouTube and X also need [yt-dlp](https://github.com/yt-dlp/yt-dlp) and [deno](https://deno.com):

```sh
brew install ffmpeg yt-dlp deno
```

Download `haul-<version>-macos-arm64.zip` from [Releases](../../releases), then:

```sh
unzip haul-*-macos-arm64.zip
xattr -d com.apple.quarantine haul    # the binary is not notarized
mkdir -p ~/.local/bin && mv haul ~/.local/bin/   # or anywhere on your PATH
```

Or build from source (Xcode 26 / Swift 6.2+):

```sh
swift build -c release
cp .build/release/haul ~/.local/bin/
```

## Usage

```
haul <url> [options]          download (same as `haul download <url>`)
haul info <url> [--urls]      show the item, its pages and streams; download nothing
haul login bilibili           log in to bilibili for higher qualities
haul templates                file-name template variables
haul --help                   sites, agent usage, exit codes, examples
haul download --help          every option
```

### Sites

| Site | Links | Needs |
|---|---|---|
| YouTube | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| X | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` picks one video of a post | yt-dlp |
| bilibili | videos, bangumi, courses, collections, series, favorites, user spaces, `b23.tv`, bare `BV…` `av…` `ep…` `ss…` `md…` | – |
| Xiaoyuzhou | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| Apple Podcasts | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, with `?i=<episode>` for one episode | – |

### Examples

```sh
haul "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"          # H.264 + AAC, plays in QuickTime
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul -p ALL "https://x.com/<user>/status/<id>"                  # every video of a post
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"      # a whole season
haul -p LATEST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul -i "BV1qt4y1X7TW"                                           # choose streams with the arrow keys
```

## For AI agents and scripts

`haul --help` carries everything an agent needs; [llms.txt](llms.txt) is the same guide as a file. The short version:

```sh
haul info --json "<url>"                         # 1. inspect
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. download exactly those streams
haul --json -q 720p -c avc,m4a "<url>"           #    or let priorities choose
```

- With `--json`, **stdout carries only one JSON document**; progress and logs go to stderr.
- **Nothing is interactive without a terminal.** `-i` without one fails with exit code 2 and names the flags to use instead.
- Stream indexes in `info` are the order haul chooses in under the same `-q` / `-c`, so they can be passed straight back.
- Pages already on disk are reported as `"status": "skipped", "reason": "exists"`, so re-running is safe.

A download prints:

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
  "pages": [
    {
      "index": 1,
      "id": "DdCEmlAydcw",
      "title": "Inside Anthropic's molecular biology lab",
      "durationSeconds": 75,
      "selected": true,
      "status": "downloaded",
      "file": "/Users/me/Movies/Inside Anthropic's molecular biology lab.mp4",
      "sizeBytes": 3434260,
      "selectedVideo": 0,
      "selectedAudio": 0,
      "video": [{ "index": 0, "quality": "360p", "resolution": "594x360", "codec": "AVC", "fps": 24, "bitrateKbps": 214, "sizeBytes": 2014782 }],
      "audio": [{ "index": 0, "codec": "M4A", "bitrateKbps": 130, "sizeBytes": 1220994 }],
      "subtitles": ["en"]
    }
  ]
}
```

That is `haul --json -q 360p -c avc,m4a …`; the stream lists are cut to the chosen ones here, the real document lists them all. Keys come out sorted.

A failure prints `"ok": false` with `"error": {"kind", "message", "exitCode"}`, and keeps the pages done before it.

| Exit code | Meaning | `error.kind` |
|---|---|---|
| 0 | done | – |
| 1 | the download or extraction failed | `failed` |
| 2 | unsupported link or bad option value | `input` |
| 3 | a required tool is missing (ffmpeg, yt-dlp) | `dependency` |
| 4 | login needed or expired | `auth` |
| 64 | bad command line (unknown flag, missing argument) | – |
| 130 | cancelled | `cancelled` |

## Options

`haul download --help` lists them all. The main ones:

| Group | Options |
|---|---|
| General | `--json`, `--config <file>`, `--debug` |
| Streams | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Pages | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Content | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--[no-]skip-ai-subtitle`, `--skip-cover`, `--skip-mux` |
| Output | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--simple-mux`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml ass`, `--cookie`, `--token`; more with `--help-hidden` |
| Tools | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--[no-]multi-thread` |

**Qualities and codecs.** `-q` takes the labels the stream table shows: `1080p`, `720p60` on YouTube; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` on bilibili. `-c` takes `av1 vp9 hevc avc` for video and `m4a opus flac eac3 mp3` for audio. Whatever is not listed comes after, best first.

**Pages.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST`, `LATEST`. A link to one episode, or `?p=N`, selects that page by itself. Pages are the parts of a bilibili video, the episodes of a season, show or list, and the videos of an X post.

**File names.** Single page: `<title>`; several: `<title>/[P<pageNumberWithZero>]<pageTitle>`. Variables: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, plus bilibili's `<bvid>` `<aid>` `<cid>` `<api>`. Dates take a format: `<publishDate:yyyy-MM-dd>`. The extension is added.

## Site notes

**YouTube** protects its stream URLs with player challenges that need a JavaScript runtime, so extraction is left to `yt-dlp -J` (which runs them in deno); everything after that is haul's own. Uploaded subtitles are muxed in; the auto-generated track is treated as an AI subtitle (`--no-skip-ai-subtitle` keeps it). googlevideo serves only bounded byte ranges, so tracks are fetched one 10 MB range at a time. `watch?v=…&list=…` downloads just the video. Keep yt-dlp current: `brew upgrade yt-dlp`.

**X** needs no login for public posts. Its videos are single MP4 files with the audio inside, so the table lists video only; `--audio-only` extracts the audio.

**bilibili** is read through its own web, TV, APP (gRPC) and international APIs. Logged out, it offers only lower qualities (typically up to 480P); log in for 1080P, 4K, HDR, Dolby Vision and Hi-Res audio:

```sh
haul login bilibili --from-edge    # reuse Microsoft Edge's login (or --from-chrome; --profile "Profile 1")
haul login bilibili                # or scan a QR code with the bilibili app
haul login bilibili --tv           # TV access token, for --api tv / --api app
```

Reading a browser's login needs Full Disk Access for the terminal, and macOS asks once for the "Safe Storage" keychain item. `--danmaku` saves the bullet comments as XML and ASS.

**Xiaoyuzhou and Apple Podcasts** need no yt-dlp: Xiaoyuzhou pages carry the audio link, Apple links go through the public iTunes API and the show's RSS feed. Episodes keep their format (`.mp3` or `.m4a`) with the cover embedded, the show as album and the host as artist. A show link lists its latest episodes oldest first, so `-p LATEST` is the newest.

## Config

`~/.config/haul/` (or `$HAUL_HOME`) holds:

| File | Purpose |
|---|---|
| `config.json` | defaults for any option |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili login |
| `archives.txt` | pages already downloaded (`--archive`) |

`config.json` uses the option names in camelCase and needs only what it changes; command-line flags override it:

```json
{
  "workDir": "~/Movies/haul",
  "codecPriority": "avc,hevc,m4a",
  "qualityPriority": "1080p,1080P",
  "filePattern": "<uploader>/<title>",
  "saveArchive": true
}
```

## Development

```sh
swift build
swift test                                   # offline, ~2 s
HAUL_LIVE=1 swift test --filter LiveTests    # also hits bilibili, YouTube, X and Apple Podcasts
```

The offline suite never touches the network: `HTTPClient` is pointed at a stub `URLProtocol` that answers from recorded API responses (`Tests/HaulCoreTests/Fixtures`) and from simulated CDNs. End-to-end tests mux with a real ffmpeg and check the result with ffprobe; they are skipped when ffmpeg is not installed. See [AGENTS.md](AGENTS.md) for the layout and conventions.

Pushing a `v*` tag builds, tests and publishes a release through GitHub Actions.

## Acknowledgements

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [swift-argument-parser](https://github.com/apple/swift-argument-parser), [swift-protobuf](https://github.com/apple/swift-protobuf), [swift-qrcode-generator](https://github.com/dagronf/swift-qrcode-generator), and the API notes of [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) and [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

## License

MIT
