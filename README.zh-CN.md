<!-- Translation of README.md; maintenance: docs/TRANSLATING.md -->

<p align="center">
  <img src="docs/assets/brand.svg" alt="haul" width="1200">
</p>

<!-- languages:start -->
<p align="center">
  <a href="README.md"><bdi>English</bdi></a> ·
  <bdi><strong>简体中文</strong></bdi> ·
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
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">一个链接，收藏你的影音。</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="构建状态"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="最新版本"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 或更高版本">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS、Linux 和 Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT 许可证"></a>
</p>

<p align="center">
  <strong>为用户与 AI 智能体打造的命令行影音下载器。</strong><br>
  输入链接，选择媒体流，将视频、音频、字幕、章节与封面合并为一个文件。
</p>

<p align="center">
  <a href="#install">安装</a> ·
  <a href="#usage">快速开始</a> ·
  <a href="#sites">支持的平台</a> ·
  <a href="#for-ai-agents-and-scripts">智能体指南</a> ·
  <a href="#options">选项</a> ·
  <a href="../../releases">发布版本</a>
</p>

---

<a id="small-command-complete-download"></a>

## 小命令，完整下载。

| ↓ 你的影音，由你决定 | ⌘ 一个二进制文件，适用所有桌面系统 | { } 为自动化而生 |
| :--- | :--- | :--- |
| 五个平台使用同一套参数选择画质、编码、分集和文件名 | 适用于 macOS、Linux 和 Windows 的单个 Go 二进制程序，支持可续传的并行分段下载 | 标准输出仅含一个 JSON 文档；日志写入标准错误；退出码明确；无终端不交互 |

| 01 / 查看 | 02 / 选择 | 03 / 下载 | 04 / 封装 |
| :--- | :--- | :--- | :--- |
| **尽览所有媒体流** | **定好优先级** | **下载到本地** | **合而为一** |
| 分集、编码与大小 | 画质、音频与分集 | 按范围或分段下载 | 轨道、字幕与封面 |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>终端预览</strong></summary>

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

> 仅供个人、研究及其他非商业用途。请自行遵守版权与各平台条款。

<a id="install"></a>

## 安装

haul 可在 **macOS、Linux 和 Windows**（amd64 与 arm64）上运行。合并媒体需要 [ffmpeg](https://ffmpeg.org)；YouTube 和 X 还需要 [yt-dlp](https://github.com/yt-dlp/yt-dlp)，YouTube 的播放器验证还需要 [deno](https://deno.com)：

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip curl  # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows，每次安装一个包
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

在 Linux 上，请用 pipx 安装 yt-dlp，或从它的 [Releases](https://github.com/yt-dlp/yt-dlp/releases/latest) 下载 `yt-dlp_linux`（arm64 为 `yt-dlp_linux_aarch64`），不要用发行版自带的包：YouTube 经常变化，打包的版本容易过时。安装完成后请打开新的终端，让这些工具出现在 `PATH` 中。

<a id="get-the-binary"></a><a id="binary"></a>

### 获取可执行文件

从 [Releases](../../releases) 下载适合你系统的压缩包 `haul-<version>-<os>-<arch>.tar.gz`（Windows 为 `.zip`），然后将 `haul` 放入 `PATH` 中的任意目录：

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

当前发布流程会对 macOS 版本签名并提交 Apple 公证。可选择 `haul-<version>-darwin-<arch>.pkg` 安装包，内附公证票据，将 `haul` 安装到 `/usr/local/bin`。压缩包包含同一份已签名程序，但首次验证可能需要联网。旧版本可能未签名。

<details>
<summary><strong>从源码构建</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

或者在克隆的仓库中执行：

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## 快速开始

从一个链接开始。haul 默认选择可用的最佳媒体流。

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

下载前先查看，或指定画质和编码：

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC，适合 QuickTime
```

<details>
<summary><strong>命令参考</strong></summary>

| 命令 | 说明 |
|---|---|
| `haul <url> [options]` | 下载（与 `haul download <url>` 相同） |
| `haul info <url> [--urls]` | 查看条目、分集与媒体流，不下载 |
| `haul login bilibili` | 登录 bilibili，获取更高画质 |
| `haul templates` | 列出文件名模板变量 |
| `haul --help` | 平台、智能体用法、退出码与示例 |
| `haul download --help` | 全部选项 |

</details>

<a id="sites"></a>

### 支持的平台

| 平台 | 链接 | 依赖 |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`；`/video/<n>` 选择帖子中的某个视频 | yt-dlp |
| **bilibili** | 视频、番剧、课程、合集、系列、收藏夹、用户空间、`b23.tv`，以及裸 ID `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`；加上 `?i=<episode>` 选择单集 | – |

<a id="make-it-yours"></a><a id="examples"></a>

### 按你的需要

**只下载音频**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC 而非 Opus
```

**几个分 P、整季，或最新一期**

```sh
# 选择分 P，保存到 Movies 目录
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# 一季的全部剧集
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# 最新一期播客
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# X 帖子中的全部视频
haul -p ALL "https://x.com/<user>/status/<id>"
```

**整理文件、字幕或交互选择**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # 只要这些语言的字幕
haul -i "BV1qt4y1X7TW"                                 # 用方向键选择媒体流
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## AI 智能体与脚本

`haul --help` 包含智能体所需的全部信息；[llms.txt](llms.txt) 是同一份指南的文件版。简而言之：

```sh
haul info --json "<url>"                                # 1. 查看
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. 精确下载这些媒体流
haul --json -q 720p -c avc,m4a "<url>"                  #    或按优先级自动选择
```

- 使用 `--json` 时，**stdout 只输出一个 JSON 文档**，进度和日志写入 stderr。其中 `files` 列出输出文件，包括新下载的和已存在的。
- **没有终端就不会交互。** 此时使用 `-i` 会以退出码 2 失败，并提示应改用的参数。
- `info` 中的媒体流索引按 haul 的选择顺序排列；查看和下载时请传入相同的 `-q` / `-c`。
- 对播放列表、整季、节目或多视频帖子执行 `info` 会列出其分集；`-p <n>` 会加上该分集的媒体流和字幕。
- 磁盘上已存在的分集会报告为 `"status": "skipped", "reason": "exists"`，因此可以安全地重复运行。

<details>
<summary><strong>JSON 响应示例</strong> · 一次成功的下载</summary>

下载会输出：

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

以上来自 `haul --json -q 360p -c avc,m4a …`。此处的键按阅读顺序排列，媒体流列表也只保留了所选的一项；实际文档会按字母顺序排列键，并列出全部媒体流。

</details>

失败时输出 `"ok": false` 和 `"error": {"kind", "message", "exitCode"}`，并保留此前已完成的分集。

| 退出码 | 含义 | `error.kind` |
|---|---|---|
| 0 | 完成 | – |
| 1 | 下载或提取失败 | `failed` |
| 2 | 链接不受支持或选项值无效 | `input` |
| 3 | 缺少必需的工具（ffmpeg、yt-dlp） | `dependency` |
| 4 | 需要登录或登录已过期 | `auth` |
| 64 | 命令行错误（未知参数、缺少参数） | – |
| 130 | 已取消 | `cancelled` |

<a id="options"></a>

## 选项

完整列表见 `haul download --help`。主要选项如下：

| 分类 | 选项 |
|---|---|
| 通用 | `--json`, `--config <file>`, `--debug` |
| 媒体流 | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| 分集 | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| 内容 | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| 输出 | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`；更多选项见 `--help-hidden` |
| 工具 | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>画质与编码优先级</strong></summary>

**画质与编码。** `-q` 使用媒体流表中显示的标签：YouTube 上为 `1080p`、`720p60`；bilibili 上为 `8K`、`Dolby Vision`、`HDR`、`4K`、`1080P60`、`1080P+`、`1080P`、`720P`。`-c` 接受视频编码 `av1 vp9 hevc avc` 和音频编码 `m4a opus flac eac3 mp3`。它们是优先级而非过滤条件：未列出的选项排在后面，按最佳优先排序。`--video-ascending` / `--audio-ascending` 会反转顺序，以获得最小的下载体积。

</details>

<details>
<summary><strong>分集选择</strong></summary>

**分集。** `8`、`1,2`、`3-5`、`1-3,10`、`ALL`、`LAST`（最后一集，即节目的最新一期；也可以用 `LATEST`）。指向单集的链接或 `?p=N` 会单独选中该分集。分集指 bilibili 视频的分 P，整季、节目或列表中的剧集，以及 X 帖子中的视频。

</details>

<details>
<summary><strong>文件名与模板变量</strong></summary>

**文件名。** 只有一个分集的条目：`<title>`；有多个分集的条目（即使 `-p` 只选一个）：`<title>/[P<pageNumberWithZero>]<pageTitle>`，编号按分集总数补零。变量：`<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`，以及 bilibili 专用的 `<bvid>` `<aid>` `<cid>` `<api>`。日期可指定格式：`<publishDate:yyyy-MM-dd>`。扩展名会自动添加。

</details>

<a id="site-notes"></a><a id="notes"></a>

## 平台说明

<details>
<summary><strong>YouTube</strong> · 提取、字幕与播放列表</summary>

YouTube 用需要 JavaScript 运行时的播放器验证保护媒体流地址，因此提取交给 `yt-dlp -J`（由它在 deno 中执行验证）；之后的一切都由 haul 自己完成。上传的字幕会被合并；自动生成的字幕仅在使用 `--auto-subtitles` 时合并。googlevideo 只提供有界的字节范围，因此轨道每次按 10 MB 的范围获取。`watch?v=…&list=…` 只下载该视频。请保持 yt-dlp 为最新版本：针对 YouTube 变化的修复都在那里发布。

</details>

<details>
<summary><strong>X</strong> · 公开帖子与音频</summary>

X 的公开帖子无需登录。其视频是内含音频的单个 MP4 文件，因此媒体流表只列出视频；`--audio-only` 可提取音频。

</details>

<details>
<summary><strong>bilibili</strong> · 登录、更高画质与弹幕</summary>

bilibili 通过其自身的 web、TV、APP（gRPC）及国际版 API 读取。未登录时只提供较低画质（通常最高 480P）；登录后可获取 1080P、4K、HDR、Dolby Vision 与 Hi-Res 音频：

```sh
haul login bilibili                # 用 bilibili App 扫描二维码
haul login bilibili --from-edge    # macOS：复用 Microsoft Edge 的登录（或 --from-chrome；--profile "Profile 1"）
haul login bilibili --tv           # TV 访问令牌，用于 --api tv / --api app
```

读取浏览器登录目前仅支持 macOS：终端需要“完全磁盘访问权限”，macOS 会询问一次“Safe Storage”钥匙串项目的访问权限。`--danmaku` 将弹幕保存为 XML 和 ASS；`--danmaku-format ass` 只保留其中一种。

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · 单集与元数据</summary>

小宇宙与 Apple Podcasts 无需 yt-dlp：小宇宙页面自带音频链接，Apple 链接则通过公开的 iTunes API 和节目的 RSS 订阅源获取。单集保留原格式（`.mp3` 或 `.m4a`），嵌入封面，并以节目名作为专辑、主播作为艺术家。节目链接按从旧到新列出最近的单集，因此 `-p LAST` 就是最新一期。

</details>

<a id="config"></a>

## 配置

`~/.config/haul/`（或 `$HAUL_HOME`）中存放：

| 文件 | 用途 |
|---|---|
| `config.json` | 任意选项的默认值 |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili 登录信息 |
| `archives.txt` | 已下载的分集（`--archive`） |

`config.json` 使用 camelCase 形式的选项名，bilibili 的选项放在 `"bilibili"` 下，只需写入要修改的项；列表可以是数组，也可以是逗号分隔的字符串。命令行参数会覆盖其中的设置：

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

## 开发

```sh
go build ./cmd/haul
go test ./...                                  # 离线运行，只需几秒
HAUL_LIVE=1 go test -run Live ./internal/...   # 同时访问 bilibili、YouTube、X 和 Apple Podcasts
```

离线测试从不访问网络：测试通过一个桩 `http.RoundTripper` 通信，它使用录制的 API 响应和模拟的 CDN（只支持范围请求的服务器、会断开的连接、忽略范围请求的服务器）作答。端到端测试使用真实的 ffmpeg 合并，并用 ffprobe 检查结果；未安装 ffmpeg 时会跳过。架构与约定见 [AGENTS.md](AGENTS.md)。

推送 `v*` 标签后，GitHub Actions 会运行测试、交叉编译，并为每个平台发布版本。

<a id="acknowledgements"></a>

## 致谢

[yt-dlp](https://github.com/yt-dlp/yt-dlp)、[FFmpeg](https://ffmpeg.org)、[GPAC](https://gpac.io)、[aria2](https://aria2.github.io)、[pflag](https://github.com/spf13/pflag)、[x/term](https://pkg.go.dev/golang.org/x/term)、[rsc.io/qr](https://pkg.go.dev/rsc.io/qr)，以及 [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) 和 [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api) 的 API 文档。

<a id="license"></a>

## 许可证

[MIT](LICENSE)

---

<p align="center">
  <strong>一个链接，收藏你的影音。</strong><br>
  <a href="#install">获取 haul</a> · <a href="llms.txt">智能体参考</a> · <a href="AGENTS.md">参与贡献</a>
</p>
