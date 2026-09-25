<!-- Translation of README.md; maintenance: docs/TRANSLATING.md -->

<p align="center">
  <img src="docs/assets/brand.svg" alt="haul" width="1200">
</p>

<!-- languages:start -->
<p align="center">
  <a href="README.md"><bdi>English</bdi></a> ·
  <a href="README.zh-CN.md"><bdi>简体中文</bdi></a> ·
  <bdi><strong>繁體中文</strong></bdi> ·
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
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">一個連結，收藏你的影音。</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="建置狀態"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="最新版本"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 或更新版本">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS、Linux 與 Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT 授權"></a>
</p>

<p align="center">
  <strong>為使用者與 AI 代理打造的命令列影音下載工具。</strong><br>
  輸入連結，選擇串流，將影片、音訊、字幕、章節與封面合併為一個檔案。
</p>

<p align="center">
  <a href="#install">安裝</a> ·
  <a href="#usage">快速開始</a> ·
  <a href="#sites">支援的平台</a> ·
  <a href="#for-ai-agents-and-scripts">代理指南</a> ·
  <a href="#options">選項</a> ·
  <a href="../../releases">發行版本</a>
</p>

---

<a id="small-command-complete-download"></a>

## 小指令，完整下載。

| ↓ 你的影音，由你決定 | ⌘ 一個執行檔，適用所有桌面系統 | { } 為自動化而生 |
| :--- | :--- | :--- |
| 五個平台使用同一套參數選擇畫質、編碼、分集與檔名 | 適用於 macOS、Linux 與 Windows 的單一 Go 執行檔，支援可續傳的平行分段下載 | 標準輸出僅含一份 JSON 文件；日誌寫入標準錯誤；結束代碼明確；沒有終端機就不進行互動 |

| 01 / 檢視 | 02 / 選擇 | 03 / 下載 | 04 / 封裝 |
| :--- | :--- | :--- | :--- |
| **盡覽所有串流** | **訂好優先順序** | **下載到本機** | **合而為一** |
| 分集、編碼與大小 | 畫質、音訊與分集 | 依範圍或分段下載 | 軌道、字幕與封面 |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>終端機預覽</strong></summary>

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

> 僅供個人、研究及其他非商業用途。請自行遵守著作權與各平台條款。

<a id="install"></a>

## 安裝

haul 可在 **macOS、Linux 與 Windows**（amd64 與 arm64）上執行。合併影音需要 [ffmpeg](https://ffmpeg.org)；YouTube 與 X 還需要 [yt-dlp](https://github.com/yt-dlp/yt-dlp)，YouTube 的播放器驗證也需要 [deno](https://deno.com)：

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg yt-dlp           # Debian / Ubuntu；deno 請見 https://deno.com
winget install Gyan.FFmpeg               # Windows，一次安裝一個套件
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

<a id="get-the-binary"></a><a id="binary"></a>

### 取得執行檔

從 [Releases](../../releases) 下載適合你系統的壓縮檔 `haul-<version>-<os>-<arch>.tar.gz`（Windows 為 `.zip`），再將 `haul` 放到 `PATH` 中的任一目錄：

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
xattr -d com.apple.quarantine ~/.local/bin/haul   # macOS：執行檔尚未經過公證
```

<details>
<summary><strong>從原始碼建置</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

或在複製下來的儲存庫中執行：

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## 快速開始

從一個連結開始。haul 預設選擇可用的最佳串流。

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

下載前先檢視，或指定畫質與編碼：

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC，適合 QuickTime
```

<details>
<summary><strong>指令參考</strong></summary>

| 指令 | 說明 |
|---|---|
| `haul <url> [options]` | 下載（與 `haul download <url>` 相同） |
| `haul info <url> [--urls]` | 檢視項目、分集與串流，不下載 |
| `haul login bilibili` | 登入 bilibili，取得更高畫質 |
| `haul templates` | 列出檔名範本變數 |
| `haul --help` | 平台、代理用法、結束代碼與範例 |
| `haul download --help` | 所有選項 |

</details>

<a id="sites"></a>

### 支援的平台

| 平台 | 連結 | 相依工具 |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`；`/video/<n>` 選擇貼文中的某部影片 | yt-dlp |
| **bilibili** | 影片、番劇、課程、合集、系列、收藏夾、使用者空間、`b23.tv`，以及直接輸入的 ID `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`；加上 `?i=<episode>` 選擇單集 | – |

<a id="make-it-yours"></a><a id="examples"></a>

### 依你的需求

**只下載音訊**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC 而非 Opus
```

**幾個分 P、整季，或最新一集**

```sh
# 選擇分 P，儲存到 Movies 目錄
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# 一季的所有劇集
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# 最新一集 Podcast
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# X 貼文中的所有影片
haul -p ALL "https://x.com/<user>/status/<id>"
```

**整理檔案、字幕或互動選擇**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # 只要這些語言的字幕
haul -i "BV1qt4y1X7TW"                                 # 用方向鍵選擇串流
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## AI 代理與指令碼

`haul --help` 包含代理所需的一切；[llms.txt](llms.txt) 是同一份指南的檔案版本。簡單來說：

```sh
haul info --json "<url>"                                # 1. 檢視
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. 精確下載這些串流
haul --json -q 720p -c avc,m4a "<url>"                  #    或依優先順序自動選擇
```

- 使用 `--json` 時，**stdout 只輸出一份 JSON 文件**，進度與日誌寫入 stderr。其中 `files` 列出輸出檔案，包含新下載與已存在的檔案。
- **沒有終端機就不進行互動。** 此時使用 `-i` 會以結束代碼 2 失敗，並指出應改用的參數。
- `info` 中的串流索引依 haul 的選擇順序排列；檢視與下載時請傳入相同的 `-q` / `-c`。
- 對播放清單、整季、節目或多影片貼文執行 `info` 會列出其分集；`-p <n>` 會加上該分集的串流與字幕。
- 磁碟上已存在的分集會回報為 `"status": "skipped", "reason": "exists"`，因此可安全地重複執行。

<details>
<summary><strong>JSON 回應範例</strong> · 一次成功的下載</summary>

下載會輸出：

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

以上來自 `haul --json -q 360p -c avc,m4a …`。此處的鍵依閱讀順序排列，串流清單也只保留所選的一項；實際文件會依字母順序排列鍵，並列出全部串流。

</details>

失敗時輸出 `"ok": false` 與 `"error": {"kind", "message", "exitCode"}`，並保留在此之前已完成的分集。

| 結束代碼 | 意義 | `error.kind` |
|---|---|---|
| 0 | 完成 | – |
| 1 | 下載或擷取失敗 | `failed` |
| 2 | 不支援的連結或無效的選項值 | `input` |
| 3 | 缺少必要工具（ffmpeg、yt-dlp） | `dependency` |
| 4 | 需要登入或登入已過期 | `auth` |
| 64 | 命令列錯誤（未知選項、缺少引數） | – |
| 130 | 已取消 | `cancelled` |

<a id="options"></a>

## 選項

完整清單請見 `haul download --help`。主要選項如下：

| 分類 | 選項 |
|---|---|
| 一般 | `--json`, `--config <file>`, `--debug` |
| 串流 | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| 分集 | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| 內容 | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| 輸出 | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`；更多選項請見 `--help-hidden` |
| 工具 | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>畫質與編碼優先順序</strong></summary>

**畫質與編碼。** `-q` 使用串流表中顯示的標籤：YouTube 為 `1080p`、`720p60`；bilibili 為 `8K`、`Dolby Vision`、`HDR`、`4K`、`1080P60`、`1080P+`、`1080P`、`720P`。`-c` 接受影片編碼 `av1 vp9 hevc avc` 與音訊編碼 `m4a opus flac eac3 mp3`。它們是優先順序而非篩選條件：未列出的項目排在後面，依最佳優先排序。`--video-ascending` / `--audio-ascending` 會反轉順序，以取得最小的下載量。

</details>

<details>
<summary><strong>分集選擇</strong></summary>

**分集。** `8`、`1,2`、`3-5`、`1-3,10`、`ALL`、`LAST`（最後一集，也就是節目的最新一集；也可以用 `LATEST`）。指向單集的連結或 `?p=N` 會單獨選取該分集。分集指 bilibili 影片的分 P，整季、節目或清單中的各集，以及 X 貼文中的影片。

</details>

<details>
<summary><strong>檔名與範本變數</strong></summary>

**檔名。** 只有一個分集的項目：`<title>`；有多個分集的項目（即使 `-p` 只選一個）：`<title>/[P<pageNumberWithZero>]<pageTitle>`，編號依分集總數補零。變數：`<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`，以及 bilibili 專用的 `<bvid>` `<aid>` `<cid>` `<api>`。日期可指定格式：`<publishDate:yyyy-MM-dd>`。副檔名會自動加上。

</details>

<a id="site-notes"></a><a id="notes"></a>

## 平台說明

<details>
<summary><strong>YouTube</strong> · 擷取、字幕與播放清單</summary>

YouTube 以需要 JavaScript 執行環境的播放器驗證保護串流網址，因此擷取交由 `yt-dlp -J` 處理（由它在 deno 中執行驗證）；之後的一切都由 haul 自行完成。上傳的字幕會合併；自動產生的字幕僅在使用 `--auto-subtitles` 時合併。googlevideo 只提供有界的位元組範圍，因此軌道每次以 10 MB 的範圍取得。`watch?v=…&list=…` 只下載該影片。請讓 yt-dlp 保持最新：針對 YouTube 變動的修正都在那裡發布。

</details>

<details>
<summary><strong>X</strong> · 公開貼文與音訊</summary>

X 的公開貼文無須登入。其影片是內含音訊的單一 MP4 檔案，因此串流表只列出影片；`--audio-only` 可擷取音訊。

</details>

<details>
<summary><strong>bilibili</strong> · 登入、更高畫質與彈幕</summary>

bilibili 透過其自身的 web、TV、APP（gRPC）及國際版 API 讀取。未登入時只提供較低畫質（通常最高 480P）；登入後可取得 1080P、4K、HDR、Dolby Vision 與 Hi-Res 音訊：

```sh
haul login bilibili                # 用 bilibili App 掃描 QR 碼
haul login bilibili --from-edge    # macOS：沿用 Microsoft Edge 的登入（或 --from-chrome；--profile "Profile 1"）
haul login bilibili --tv           # TV 存取權杖，供 --api tv / --api app 使用
```

讀取瀏覽器登入目前僅支援 macOS：終端機需要「完整磁碟取用權限」，macOS 會詢問一次「Safe Storage」鑰匙圈項目的存取權。`--danmaku` 將彈幕儲存為 XML 與 ASS；`--danmaku-format ass` 只保留其中一種。

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · 單集與中繼資料</summary>

小宇宙與 Apple Podcasts 無須 yt-dlp：小宇宙頁面本身帶有音訊連結，Apple 連結則透過公開的 iTunes API 與節目的 RSS 訂閱源取得。單集保留原格式（`.mp3` 或 `.m4a`），內嵌封面，並以節目名稱作為專輯、主持人作為演出者。節目連結依舊到新列出近期單集，因此 `-p LAST` 就是最新一集。

</details>

<a id="config"></a>

## 設定

`~/.config/haul/`（或 `$HAUL_HOME`）中存放：

| 檔案 | 用途 |
|---|---|
| `config.json` | 任何選項的預設值 |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili 登入資訊 |
| `archives.txt` | 已下載的分集（`--archive`） |

`config.json` 使用 camelCase 形式的選項名稱，bilibili 的選項放在 `"bilibili"` 之下，只需寫入要變更的項目；清單可以是陣列，也可以是以逗號分隔的字串。命令列參數會覆寫其中的設定：

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

## 開發

```sh
go build ./cmd/haul
go test ./...                                  # 離線執行，只需幾秒
HAUL_LIVE=1 go test -run Live ./internal/...   # 同時連線 bilibili、YouTube、X 與 Apple Podcasts
```

離線測試從不連上網路：測試透過一個替身 `http.RoundTripper` 通訊，它以錄製的 API 回應與模擬的 CDN（只支援範圍請求的伺服器、會中斷的連線、忽略範圍請求的伺服器）作答。端對端測試使用真實的 ffmpeg 合併，並以 ffprobe 檢查結果；未安裝 ffmpeg 時會略過。架構與慣例請見 [AGENTS.md](AGENTS.md)。

推送 `v*` 標籤後，GitHub Actions 會執行測試、交叉編譯，並為每個平台發布版本。

<a id="acknowledgements"></a>

## 致謝

[yt-dlp](https://github.com/yt-dlp/yt-dlp)、[FFmpeg](https://ffmpeg.org)、[GPAC](https://gpac.io)、[aria2](https://aria2.github.io)、[pflag](https://github.com/spf13/pflag)、[x/term](https://pkg.go.dev/golang.org/x/term)、[rsc.io/qr](https://pkg.go.dev/rsc.io/qr)，以及 [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) 與 [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api) 的 API 文件。

<a id="license"></a>

## 授權條款

[MIT](LICENSE)

---

<p align="center">
  <strong>一個連結，收藏你的影音。</strong><br>
  <a href="#install">取得 haul</a> · <a href="llms.txt">代理參考</a> · <a href="AGENTS.md">參與貢獻</a>
</p>
