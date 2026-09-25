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

## macOS signing and notarization

Signing is configured for **Qi Jiang**, Apple Developer team **89G3DBC6CS**.
Use these existing Developer ID identities, not an Apple Distribution or ad hoc
certificate. Both certificates use Apple's Developer ID G2 intermediate and
expire on **2031-09-17**.

The Apple portal showed the current membership term ending on **2026-09-30**
(checked 2026-09-25). Confirm renewal separately; the certificate expiration
date is not the membership renewal date.

| Purpose | Identity |
|---|---|
| Go executable | `Developer ID Application: Qi Jiang (89G3DBC6CS)` |
| Installer package | `Developer ID Installer: Qi Jiang (89G3DBC6CS)` |

### Credentials

The following GitHub Actions repository secrets are already installed in
`ac1982/haul`. Reuse them; never put their values in source, logs, artifacts,
issue comments, or this document.

| Secret | Contents |
|---|---|
| `MACOS_APPLICATION_P12_BASE64` | Base64-encoded application certificate and private key, in an encrypted PKCS#12 bundle |
| `MACOS_INSTALLER_P12_BASE64` | Base64-encoded installer certificate and private key, in an encrypted PKCS#12 bundle |
| `MACOS_CERTIFICATE_PASSWORD` | Password for both PKCS#12 bundles |
| `APPLE_API_KEY_P8` | App Store Connect API private key, as PEM text |
| `APPLE_API_KEY_ID` | API key ID |
| `APPLE_API_ISSUER_ID` | Team API issuer ID |

The team API key is named **haul notarization** and has the **Developer** role.
It has been validated with `notarytool`. Do not revoke or replace it, the signing
certificates, or unrelated App Store Connect keys during routine releases.

On the maintainer's Mac, signing material is outside the repository at:

```text
~/Library/Application Support/haul-signing/89G3DBC6CS/
  application.p12, installer.p12   encrypted certificate/private-key bundles
  application.cer, installer.cer   public certificates
  application.key, installer.key  private signing keys
  application.csr, installer.csr  certificate requests
  AuthKey_<key-id>.p8              notarization API private key
  certificate-password            PKCS#12 password (secret)
  keychain-password               dedicated keychain password (secret)
  haul.keychain-db                dedicated signing keychain
  metadata.json                   team, identities, and API identifiers
```

The directory is owner-only (`0700`); private files must be `0600`. Preserve this
directory and back it up securely. Read password files directly into the signing
process when needed; do not print them. The dedicated keychain is deliberately
outside the normal user keychain search list, so pass its path explicitly.
Notarization credentials are stored there under profile **haul-notary**.

### Release requirements

`.github/workflows/release.yml` runs on macOS so that all six Go targets can be
cross-compiled and the two macOS targets can be signed. Its signing entry point
is `scripts/sign-macos.sh <binary> <output.pkg> <version>`, with the six secrets
above supplied as environment variables. It produces a package for the binary's
architecture and rejects invalid versions or unexpected signing identities.

- Cross-compile with `CGO_ENABLED=0`, `GOOS=darwin`, and `GOARCH=arm64` or `amd64`.
  Sign the final executable on a macOS runner with `codesign --options runtime
  --timestamp`, the application identity above, and identifier `com.ac1982.haul`.
  Do not modify executable bytes after signing.
- Build a `.pkg` installing `haul` into `/usr/local/bin`, signed with the
  **installer** identity and a trusted timestamp. Keep the archive download too.
- Submit the signed package with `xcrun notarytool submit ... --wait` and require
  status **Accepted**. An upload succeeding is not a successful notarization.
- Run `xcrun stapler staple`, `xcrun stapler validate`, and
  `spctl --assess --type install --verbose=2` on the package before publishing.
  A standalone executable or archive cannot carry a stapled ticket; the `.pkg`
  provides that ticket for offline verification. The archive must contain the
  same signed executable that Apple checked in the package.
- Generate checksums after signing and stapling. A missing credential, failed
  signature, rejection, or pending notarization must stop publication; never
  silently fall back to unsigned macOS release files.
- Keep signing credentials in a temporary keychain on CI and remove the keychain
  and decoded private files on both success and failure. Apple signing covers
  macOS only; it does not sign Windows executables or remove OS permission prompts.

For local notarization commands, append:

```sh
--keychain-profile haul-notary \
--keychain "$HOME/Library/Application Support/haul-signing/89G3DBC6CS/haul.keychain-db"
```

On 2026-09-25, both Go verification packages passed notarization, stapling,
ticket validation, and Gatekeeper (`source=Notarized Developer ID`):

| Architecture | Apple submission ID |
|---|---|
| arm64 | `c9f33e5f-a464-4089-8121-11cf62b2bbac` |
| amd64 | `ba3a1679-ad1e-45c6-8131-bfbc4e0d6660` |

These were local verification builds, not published releases. Each new release
still needs its own validation. Signing credentials do not confer Windows trust;
Windows signing requires a separate Authenticode certificate or signing service.

When operating the Apple portals, keep one persistent Playwright browser/context
and reuse its tabs so the maintainer does not have to sign in repeatedly.

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
