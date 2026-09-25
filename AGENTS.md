# Working on haul

Notes for coding agents (and people) changing this repository. Using haul as a tool is covered by [llms.txt](llms.txt).

## Build and test

```sh
swift build                                   # debug
swift build -c release                        # .build/release/haul
swift test                                    # offline suite, ~2 s; end-to-end tests need ffmpeg + ffprobe
HAUL_LIVE=1 swift test --filter LiveTests     # hits the real sites
```

Swift 6 language mode, Swift 6.2+ toolchain (Xcode 26), macOS 15+, Apple Silicon only. CI (`.github/workflows`) runs the offline suite on every push; a `v*` tag builds and publishes a release. Tests must stay offline: point `HTTPClient` at `Stub.client` and register answers with `Stub.on(...)`; fixtures in `Tests/HaulCoreTests/Fixtures` are compact recorded JSON. Pipeline tests set `HAUL_HOME` to a temporary directory, so they never read a real login.

## Layout

```
Sources/haul                  the CLI (swift-argument-parser): commands, flag groups, --help text
Sources/HaulCore
  Core/                       logging, terminal, picker, JSON, gzip, shell, formatting, storage, errors
  Net/                        HTTP client (URLSession, streaming, stub-able)
  Model/                      Page, tracks, VideoInfo, Site (link recognition for every site)
  Download/                   ranged + segmented downloader, aria2c, progress bar
  Mux/                        ffmpeg / MP4Box command lines
  Subtitles/                  subtitle discovery and conversion to SRT
  Pipeline/                   options, the run, the per-page flow, file-name templates, UI, the --json Report
  Sites/MediaSource.swift     everything that differs per site, behind one protocol
  Sites/Bilibili/             ids, signing, info fetchers, playurl (web/tv/intl) and gRPC APP clients, danmaku, login
  Sites/YtDlp/                YouTube and X: yt-dlp -J → pages and streams
  Sites/Podcast/              Xiaoyuzhou pages, Apple Podcasts lookup and RSS
Tests/HaulCoreTests           swift-testing suites
```

## Conventions

- **Sites are equal.** The pipeline and `PageDownloader` never check which site they are on; anything site-specific goes behind `MediaSource`. A new site is a `Site` case (with `match`, `name`, `linkForms`), a source, and tests.
- **Output contract.** Human output goes through `Log` (stdout, or stderr under `--json`). Data for `--json` goes into `Report`; only the CLI prints it, with `Log.output`. Keep the JSON keys stable: agents depend on them. Add fields, do not rename them.
- **Errors.** Throw `HaulError` with the right kind: `.input` for bad links and options, `.dependency` for missing tools, `.auth` for login problems, the default `.failed` for everything else. The kind decides the exit code. Never end a failed download with exit code 0: a failed mux throws.
- **No prompts without a terminal.** Anything interactive goes through `Picker`, which refuses when stdin or the log stream is not a TTY.
- **Messages** are short English sentences without a trailing period; `Log.status` for progress, `Log.warn` for a problem that does not stop the run.
- **Help text** lives in `Sources/haul`: keep `haul --help` lines under 78 columns, and update README.md and llms.txt when flags or JSON keys change.
- **Commits**: English, conventional prefixes (`feat(scope):`, `fix:`, `docs:`, `test:`), one change each.
- Regenerate protobuf code after editing `Sources/HaulCore/Sites/Bilibili/Proto/*.proto` (needs `brew install swift-protobuf`):

  ```sh
  protoc --proto_path=Sources/HaulCore/Sites/Bilibili/Proto --swift_out=Sources/HaulCore/Sites/Bilibili/Generated \
    --swift_opt=Visibility=Public Sources/HaulCore/Sites/Bilibili/Proto/*.proto
  ```
