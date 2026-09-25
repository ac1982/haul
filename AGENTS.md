# Working on haul

Notes for coding agents (and people) changing this repository. Using haul as a tool is covered by [llms.txt](llms.txt).

## Build and test

```sh
go build ./cmd/haul                            # the binary
go vet ./... && gofmt -l .                     # must be clean
go test -race ./...                            # offline suite; end-to-end tests need ffmpeg + ffprobe
HAUL_LIVE=1 go test -run Live ./internal/...   # hits the real sites
GOOS=windows go vet ./... && GOOS=linux go vet ./...
```

Go 1.26+, no cgo: haul must cross-compile for macOS, Linux and Windows (amd64, arm64) with `CGO_ENABLED=0`. Dependencies are deliberately few (`spf13/pflag`, `golang.org/x/term`, `rsc.io/qr`); talk to external tools (`ffmpeg`, `yt-dlp`, `security`, `sqlite3`) through `internal/shell` instead of linking libraries. CI (`.github/workflows`) runs vet, gofmt and the tests on all three systems; a `v*` tag builds and publishes a release for every platform.

## Architecture

```
cmd/haul                 CLI: commands, the option table (flags.go) that drives parsing, help and config.json,
                         wiring of extractors and engine, the --json document (report.go)
internal/media           the domain model, the same for every site: Item → Entries → Formats (video / audio
                         formats, subtitles, chapters, extra audio, sidecars), each stream a ready Resource
internal/extract         the Extractor interface (Match, Resolve, Formats), the Router that picks one by link,
                         and Authenticator for logins
internal/sites/…         one package per site family: bilibili, ytdlp (YouTube, X), podcast (Xiaoyuzhou, Apple)
internal/engine          the pipeline: resolve → select pages → sort and choose streams → plan the output →
                         download → mux; returns a Result, reports to an Observer, asks a Chooser
internal/render          what people see: header, stream table, progress, summary, the arrow-key chooser
internal/fetch           the downloader: parallel ranges, sequential closed ranges, whole streams, resume;
                         aria2c as an alternative Fetcher
internal/mux             ffmpeg / MP4Box command lines for a Job
internal/subtitle        subtitle formats → SRT, language tags → ISO 639-2
internal/httpx           the shared HTTP client; no site knowledge
internal/browser         Chromium cookie reader (macOS keychain via `security`, the DB via `sqlite3`)
internal/console, errs, jsonv, format, storage, shell    support
internal/testkit         stub network and ffmpeg-made media for tests (test-only)
```

Dependencies point one way: `cmd` → `engine`, `render`, `sites` → `extract`, `fetch`, `mux`, `subtitle` → `media`, `httpx` → support packages. Sites never import the engine; the engine never imports a site.

## Conventions

- **Sites are equal.** The engine never asks which site it is on. Everything a site needs — link forms, APIs, headers, cookies, CDN rewriting, range rules — is decided inside its extractor and handed over as `media.Resource`s. A new site is a new package implementing `extract.Extractor`, registered in `cmd/haul/download.go`, with its tests.
- **Output contract.** People-facing lines go through `internal/console` (stdout, or stderr under `--json`) and `internal/render`. The `--json` document is built from `engine.Result` in `cmd/haul/report.go`; its keys are part of haul's interface. Add fields, never rename them, and update README.md and llms.txt with them.
- **Errors.** Return `errs` errors with the right kind: `errs.NewInput` for bad links and options, `errs.NewDependency` for missing tools (with `shell.InstallHint`), `errs.NewAuth` for login problems, `errs.New` for failures. The kind is the exit code and the JSON `error.kind`. Never let a failed download end with exit code 0.
- **No prompts without a terminal.** Anything interactive goes through `console.Choose`, which refuses when stdin or the log stream is not a terminal.
- **Tests stay offline.** Use `testkit.NewStub()` per test (routes by URL fragment, request log, `Ranged`, `TruncateAt`, `Redirect`), `t.TempDir()`, and `t.Setenv("HAUL_HOME", …)` wherever stored logins or the archive could be read. Live tests live in `live_test.go` files behind `HAUL_LIVE=1`.
- **Messages** are short English sentences without a trailing period; `console.Status` for progress, `console.Warn` for a problem that does not stop the run.
- **Help text** comes from `cmd/haul/flags.go` and `help.go`; every line stays within 80 columns (a test checks).
- **Commits**: English, conventional prefixes (`feat(scope):`, `fix:`, `docs:`, `test:`), one change each.
