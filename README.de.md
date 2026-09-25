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
  <bdi><strong>Deutsch</strong></bdi> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">Ein Link. Deine Medien.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="Build-Status"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="Neueste Version"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 oder neuer">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux und Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT-Lizenz"></a>
</p>

<p align="center">
  <strong>Ein Kommandozeilenprogramm zum Herunterladen von Video und Audio, für Menschen und KI-Agenten gleichermaßen.</strong><br>
  Gib ihm einen Link. Wähle deine Streams. Erhalte Video, Audio, Untertitel, Kapitel und Cover in einer Datei.
</p>

<p align="center">
  <a href="#install">Installation</a> ·
  <a href="#usage">Schnellstart</a> ·
  <a href="#sites">Unterstützte Plattformen</a> ·
  <a href="#for-ai-agents-and-scripts">Anleitung für Agenten</a> ·
  <a href="#options">Optionen</a> ·
  <a href="../../releases">Releases</a>
</p>

---

<a id="small-command-complete-download"></a>

## Kleiner Befehl. Vollständiger Download.

| ↓ Deine Medien, wie du sie willst | ⌘ Eine Binärdatei für jeden Desktop | { } Bereit für Automatisierung |
| :--- | :--- | :--- |
| Qualität, Codecs, Seiten und Dateinamen mit denselben Begriffen auf fünf Plattformen wählen | Eine einzelne Go-Binärdatei für macOS, Linux und Windows, mit parallelen Bereichsdownloads, die sich fortsetzen lassen | Ein JSON-Dokument auf stdout, Protokolle auf stderr, aussagekräftige Exit-Codes und keine Rückfragen ohne Terminal |

| 01 / PRÜFEN | 02 / WÄHLEN | 03 / HERUNTERLADEN | 04 / ZUSAMMENFÜHREN |
| :--- | :--- | :--- | :--- |
| **Jeden Stream sehen** | **Prioritäten setzen** | **Ab auf die Platte** | **Alles beisammen** |
| Seiten, Codecs und Größen | Qualität, Audio und Seiten | In Bereichen oder Segmenten | Spuren, Untertitel und Cover |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>Ein Blick ins Terminal</strong></summary>

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

> Für persönliche Zwecke, Forschung und andere nicht kommerzielle Nutzung. Du bist dafür verantwortlich, Urheberrechte und die Bedingungen der jeweiligen Plattform zu beachten.

<a id="install"></a>

## Installation

haul läuft unter **macOS, Linux und Windows** (amd64 und arm64). Zum Zusammenführen wird [ffmpeg](https://ffmpeg.org) benötigt; YouTube und X brauchen außerdem [yt-dlp](https://github.com/yt-dlp/yt-dlp), und für die Player-Prüfungen von YouTube ist [deno](https://deno.com) erforderlich:

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg yt-dlp           # Debian / Ubuntu; deno: https://deno.com
winget install Gyan.FFmpeg               # Windows, ein Paket nach dem anderen
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

<a id="get-the-binary"></a><a id="binary"></a>

### Binärdatei herunterladen

Lade das Archiv für dein System unter [Releases](../../releases) herunter – `haul-<version>-<os>-<arch>.tar.gz` (unter Windows `.zip`) – und lege `haul` in ein beliebiges Verzeichnis in deinem `PATH`:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
xattr -d com.apple.quarantine ~/.local/bin/haul   # macOS: die Binärdatei ist nicht notarisiert
```

<details>
<summary><strong>Aus dem Quellcode bauen</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

oder aus einem Klon des Repositorys:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## Schnellstart

Beginne mit einem Link. Standardmäßig wählt haul die besten verfügbaren Streams.

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

Vor dem Download ansehen oder Qualität und Codec wählen:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC für QuickTime
```

<details>
<summary><strong>Befehlsübersicht</strong></summary>

| Befehl | Bedeutung |
|---|---|
| `haul <url> [options]` | herunterladen (wie `haul download <url>`) |
| `haul info <url> [--urls]` | Eintrag, Seiten und Streams anzeigen; nichts herunterladen |
| `haul login bilibili` | bei bilibili für höhere Qualitäten anmelden |
| `haul templates` | Variablen für Dateinamenvorlagen |
| `haul --help` | Plattformen, Agentennutzung, Exit-Codes, Beispiele |
| `haul download --help` | alle Optionen |

</details>

<a id="sites"></a>

### Unterstützte Plattformen

| Plattform | Links | Benötigt |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` wählt ein Video eines Beitrags | yt-dlp |
| **bilibili** | Videos, Bangumi, Kurse, Sammlungen, Serien, Favoriten, Nutzerseiten, `b23.tv`, direkte IDs `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`, mit `?i=<episode>` für eine einzelne Folge | – |

<a id="make-it-yours"></a><a id="examples"></a>

### Nach deinen Wünschen

**Nur Audio**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC statt Opus
```

**Einige Teile, eine ganze Staffel oder die neueste Folge**

```sh
# Ausgewählte Teile, gespeichert im Ordner Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# Alle Folgen einer Staffel
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# Die neueste Podcastfolge
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# Alle Videos eines X-Beitrags
haul -p ALL "https://x.com/<user>/status/<id>"
```

**Geordnete Dateien, Untertitel oder interaktive Auswahl**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # nur diese Untertitelsprachen
haul -i "BV1qt4y1X7TW"                                 # Streams mit den Pfeiltasten wählen
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## Für KI-Agenten und Skripte

`haul --help` enthält alles, was ein Agent braucht; [llms.txt](llms.txt) ist dieselbe Anleitung als Datei. Kurzfassung:

```sh
haul info --json "<url>"                                # 1. ansehen
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. genau diese Streams herunterladen
haul --json -q 720p -c avc,m4a "<url>"                  #    oder die Prioritäten wählen lassen
```

- Mit `--json` enthält **stdout nur ein einziges JSON-Dokument**; Fortschritt und Protokolle gehen an stderr. Sein Feld `files` listet die Ausgabedateien auf, neue wie bereits vorhandene.
- **Ohne Terminal gibt es keine Interaktion.** `-i` scheitert dann mit Exit-Code 2 und nennt die stattdessen zu verwendenden Optionen.
- Die Streamindizes in `info` folgen der Reihenfolge, in der haul auswählt; gib `info` und dem Download dieselben `-q` / `-c` mit.
- Bei einer Playlist, Staffel, Sendung oder einem Beitrag mit mehreren Videos listet `info` die Seiten auf; `-p <n>` ergänzt Streams und Untertitel dieser Seite.
- Bereits vorhandene Seiten werden als `"status": "skipped", "reason": "exists"` gemeldet, daher ist erneutes Ausführen sicher.

<details>
<summary><strong>Beispiel einer JSON-Antwort</strong> · ein erfolgreicher Download</summary>

Ein Download gibt Folgendes aus:

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

Das entspricht `haul --json -q 360p -c avc,m4a …`. Hier stehen die Schlüssel in Lesereihenfolge, und die Streamlisten sind auf die ausgewählten Streams gekürzt; das echte Dokument sortiert seine Schlüssel alphabetisch und listet jeden Stream auf.

</details>

Ein Fehler liefert `"ok": false` mit `"error": {"kind", "message", "exitCode"}` und behält die zuvor abgeschlossenen Seiten.

| Exit-Code | Bedeutung | `error.kind` |
|---|---|---|
| 0 | fertig | – |
| 1 | Download oder Extraktion fehlgeschlagen | `failed` |
| 2 | nicht unterstützter Link oder ungültiger Optionswert | `input` |
| 3 | ein benötigtes Werkzeug fehlt (ffmpeg, yt-dlp) | `dependency` |
| 4 | Anmeldung erforderlich oder abgelaufen | `auth` |
| 64 | ungültige Kommandozeile (unbekannte Option, fehlendes Argument) | – |
| 130 | abgebrochen | `cancelled` |

<a id="options"></a>

## Optionen

`haul download --help` listet alle auf. Die wichtigsten:

| Gruppe | Optionen |
|---|---|
| Allgemein | `--json`, `--config <file>`, `--debug` |
| Streams | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| Seiten | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| Inhalt | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| Ausgabe | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; weitere mit `--help-hidden` |
| Werkzeuge | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>Prioritäten für Qualität und Codecs</strong></summary>

**Qualitäten und Codecs.** `-q` nimmt die Bezeichnungen aus der Streamtabelle: `1080p`, `720p60` bei YouTube; `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P` bei bilibili. `-c` nimmt `av1 vp9 hevc avc` für Video und `m4a opus flac eac3 mp3` für Audio. Das sind Prioritäten, keine Filter: Alles nicht Genannte folgt danach, das Beste zuerst. `--video-ascending` / `--audio-ascending` kehren die Reihenfolge um, für den kleinsten Download.

</details>

<details>
<summary><strong>Seitenauswahl</strong></summary>

**Seiten.** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (die letzte Seite, also die neueste Folge einer Sendung; `LATEST` geht auch). Ein Link auf eine einzelne Folge oder `?p=N` wählt diese Seite allein aus. Seiten sind die Teile eines bilibili-Videos, die Folgen einer Staffel, Sendung oder Liste und die Videos eines X-Beitrags.

</details>

<details>
<summary><strong>Dateinamen und Vorlagenvariablen</strong></summary>

**Dateinamen.** Ein Eintrag mit einer Seite: `<title>`; mit mehreren (auch wenn `-p` nur eine auswählt): `<title>/[P<pageNumberWithZero>]<pageTitle>`, mit Nullen aufgefüllt entsprechend der Seitenzahl. Variablen: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, dazu bilibilis `<bvid>` `<aid>` `<cid>` `<api>`. Datumsangaben nehmen ein Format an: `<publishDate:yyyy-MM-dd>`. Die Dateierweiterung wird ergänzt.

</details>

<a id="site-notes"></a><a id="notes"></a>

## Hinweise zu Plattformen

<details>
<summary><strong>YouTube</strong> · Extraktion, Untertitel und Playlists</summary>

YouTube schützt seine Stream-URLs mit Player-Prüfungen, die eine JavaScript-Laufzeit brauchen; die Extraktion übernimmt daher `yt-dlp -J` (das sie in deno ausführt), alles Weitere erledigt haul selbst. Hochgeladene Untertitel werden eingebunden, automatisch erzeugte nur mit `--auto-subtitles`. googlevideo liefert nur begrenzte Bytebereiche, daher werden Spuren in Bereichen von je 10 MB abgerufen. `watch?v=…&list=…` lädt nur das Video. Halte yt-dlp aktuell: Dort landen die Korrekturen für Änderungen bei YouTube.

</details>

<details>
<summary><strong>X</strong> · öffentliche Beiträge und Audio</summary>

Öffentliche Beiträge auf X brauchen keine Anmeldung. Die Videos sind einzelne MP4-Dateien mit enthaltener Tonspur, daher zeigt die Tabelle nur Video; `--audio-only` extrahiert das Audio.

</details>

<details>
<summary><strong>bilibili</strong> · Anmeldung, höhere Qualitäten und Danmaku</summary>

bilibili wird über seine eigenen Web-, TV-, APP- (gRPC) und internationalen APIs gelesen. Ohne Anmeldung gibt es nur niedrigere Qualitäten (meist bis 480P); melde dich für 1080P, 4K, HDR, Dolby Vision und Hi-Res-Audio an:

```sh
haul login bilibili                # QR-Code mit der bilibili-App scannen
haul login bilibili --from-edge    # macOS: Anmeldung aus Microsoft Edge übernehmen (oder --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # TV-Zugriffstoken, für --api tv / --api app
```

Das Übernehmen der Browseranmeldung funktioniert vorerst nur unter macOS: Das Terminal braucht dafür Festplattenvollzugriff, und macOS fragt einmal nach dem Schlüsselbundeintrag „Safe Storage“. `--danmaku` speichert die eingeblendeten Kommentare als XML und ASS; `--danmaku-format ass` behält nur eines der beiden Formate.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · Folgen und Metadaten</summary>

Xiaoyuzhou und Apple Podcasts brauchen kein yt-dlp: Xiaoyuzhou-Seiten enthalten den Audiolink, Apple-Links laufen über die öffentliche iTunes-API und den RSS-Feed der Sendung. Folgen behalten ihr Format (`.mp3` oder `.m4a`), mit eingebettetem Cover, der Sendung als Album und dem Moderator als Künstler. Ein Sendungslink listet die jüngsten Folgen, die älteste zuerst; `-p LAST` ist also die neueste.

</details>

<a id="config"></a>

## Konfiguration

`~/.config/haul/` (oder `$HAUL_HOME`) enthält:

| Datei | Zweck |
|---|---|
| `config.json` | Standardwerte für beliebige Optionen |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili-Anmeldung |
| `archives.txt` | bereits heruntergeladene Seiten (`--archive`) |

`config.json` verwendet die Optionsnamen in camelCase, die von bilibili unter `"bilibili"`, und braucht nur die geänderten Werte; eine Liste kann ein Array oder eine kommagetrennte Zeichenkette sein. Optionen auf der Kommandozeile haben Vorrang:

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

## Entwicklung

```sh
go build ./cmd/haul
go test ./...                                  # offline, wenige Sekunden
HAUL_LIVE=1 go test -run Live ./internal/...   # greift zusätzlich auf bilibili, YouTube, X und Apple Podcasts zu
```

Die Offline-Tests greifen nie auf das Netzwerk zu: Sie sprechen mit einem Stub-`http.RoundTripper`, der aus aufgezeichneten API-Antworten und simulierten CDNs antwortet (Server nur mit Bereichsanfragen, abgebrochene Verbindungen, Server, die Bereiche ignorieren). Ende-zu-Ende-Tests führen mit echtem ffmpeg zusammen und prüfen das Ergebnis mit ffprobe; ohne installiertes ffmpeg werden sie übersprungen. Aufbau und Konventionen beschreibt [AGENTS.md](AGENTS.md).

Ein gepushter `v*`-Tag testet, kompiliert für alle Plattformen und veröffentlicht über GitHub Actions ein Release.

<a id="acknowledgements"></a>

## Danksagung

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr) sowie die API-Notizen von [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) und [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## Lizenz

[MIT](LICENSE)

---

<p align="center">
  <strong>Ein Link. Deine Medien.</strong><br>
  <a href="#install">haul holen</a> · <a href="llms.txt">Referenz für Agenten</a> · <a href="AGENTS.md">Mitwirken</a>
</p>
