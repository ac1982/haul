<!-- Translation of README.md; maintenance: docs/TRANSLATING.md -->

<p align="center">
  <img src="docs/assets/brand.svg" alt="haul" width="1200">
</p>

<!-- languages:start -->
<p align="center">
  <a href="README.md"><bdi>English</bdi></a> ·
  <a href="README.zh-CN.md"><bdi>简体中文</bdi></a> ·
  <a href="README.zh-TW.md"><bdi>繁體中文</bdi></a> ·
  <bdi><strong>日本語</strong></bdi> ·
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

<h1 align="center">リンクひとつで、メディアを手元に。</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="ビルド状況"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="最新リリース"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 以降">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS、Linux、Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT ライセンス"></a>
</p>

<p align="center">
  <strong>人にも AI エージェントにも使いやすい、動画・音声のコマンドラインダウンローダー。</strong><br>
  リンクを渡し、ストリームを選べば、動画、音声、字幕、チャプター、カバーがひとつのファイルにまとまります。
</p>

<p align="center">
  <a href="#install">インストール</a> ·
  <a href="#usage">クイックスタート</a> ·
  <a href="#sites">対応サイト</a> ·
  <a href="#for-ai-agents-and-scripts">エージェント向けガイド</a> ·
  <a href="#options">オプション</a> ·
  <a href="../../releases">リリース</a>
</p>

---

<a id="small-command-complete-download"></a>

## 小さなコマンドで、完全なダウンロード。

| ↓ メディアを思いどおりに | ⌘ 単一バイナリで、あらゆるデスクトップに | { } 自動化にすぐ使える |
| :--- | :--- | :--- |
| 5 サイトで共通のオプションを使い、画質・コーデック・ページ・ファイル名を指定 | macOS、Linux、Windows 向けの単一 Go バイナリ。再開可能な並列の範囲指定ダウンロードに対応 | 標準出力は JSON 文書ひとつ。ログは標準エラーへ。意味のある終了コードを返し、端末がなければ対話しません |

| 01 / 確認 | 02 / 選択 | 03 / ダウンロード | 04 / 結合 |
| :--- | :--- | :--- | :--- |
| **すべてのストリームを一覧** | **優先順位を決める** | **手元に持ち帰る** | **ひとつにまとめる** |
| ページ、コーデック、サイズ | 画質、音声、ページ | 範囲指定またはセグメント | トラック、字幕、カバー |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>端末での表示例</strong></summary>

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

> 個人利用、研究、その他の非商用目的に限ります。著作権と各サイトの利用規約は利用者の責任で守ってください。

<a id="install"></a>

## インストール

haul は **macOS、Linux、Windows**（amd64 と arm64）で動作します。結合には [ffmpeg](https://ffmpeg.org) が必要です。YouTube と X には [yt-dlp](https://github.com/yt-dlp/yt-dlp)、YouTube のプレーヤー検証には [deno](https://deno.com) も必要です。

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip curl  # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows。パッケージは 1 つずつ
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

Linux では、yt-dlp をディストリビューションのパッケージではなく、pipx か [リリース](https://github.com/yt-dlp/yt-dlp/releases/latest) の `yt-dlp_linux`（arm64 では `yt-dlp_linux_aarch64`）で入れてください。YouTube は頻繁に変わるため、パッケージ版はすぐ古くなります。インストール後は新しいターミナルを開き、ツールが `PATH` に入るようにしてください。

<a id="get-the-binary"></a><a id="binary"></a>

### バイナリの入手

[Releases](../../releases) からお使いのシステム用のアーカイブ `haul-<version>-<os>-<arch>.tar.gz`（Windows では `.zip`）をダウンロードし、`haul` を `PATH` 上の任意の場所に置きます。

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

現在のリリース処理では macOS 版に署名し、Apple の公証を受けます。`haul-<version>-darwin-<arch>.pkg` には公証チケットが添付され、`haul` を `/usr/local/bin` にインストールします。アーカイブも同じ署名済み実行ファイルを含みますが、初回の検証にはネット接続が必要な場合があります。過去のリリースは未署名の場合があります。

<details>
<summary><strong>ソースからビルド</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

または、クローンしたリポジトリで:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## クイックスタート

まずはリンクから。haul は既定で、利用できる最良のストリームを選びます。

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

ダウンロード前に確認したり、画質とコーデックを指定したりできます。

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # QuickTime 向けに H.264 + AAC
```

<details>
<summary><strong>コマンド一覧</strong></summary>

| コマンド | 意味 |
|---|---|
| `haul <url> [options]` | ダウンロード（`haul download <url>` と同じ） |
| `haul info <url> [--urls]` | 項目・ページ・ストリームを表示。ダウンロードはしない |
| `haul login bilibili` | bilibili にログインして高画質を利用 |
| `haul templates` | ファイル名テンプレートの変数 |
| `haul --help` | サイト、エージェント向けの使い方、終了コード、使用例 |
| `haul download --help` | すべてのオプション |

</details>

<a id="sites"></a>

### 対応サイト

| サイト | リンク | 必要なツール |
|---|---|---|
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`。`/video/<n>` で投稿内の動画をひとつ指定 | yt-dlp |
| **bilibili** | 動画、番組、講座、コレクション、シリーズ、お気に入り、ユーザーページ、`b23.tv`、ID の直接指定 `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`。`?i=<episode>` を付けると単一エピソード | – |

<a id="make-it-yours"></a><a id="examples"></a>

### 自分好みに

**音声だけ**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # Opus ではなく AAC
```

**一部のパート、シーズン全体、または最新回**

```sh
# 選んだパートを Movies フォルダーに保存
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# シーズンの全エピソード
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# ポッドキャストの最新回
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# X の投稿に含まれるすべての動画
haul -p ALL "https://x.com/<user>/status/<id>"
```

**ファイルの整理、字幕、対話的な選択**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # これらの言語の字幕のみ
haul -i "BV1qt4y1X7TW"                                 # 矢印キーでストリームを選択
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## AI エージェントとスクリプト

エージェントに必要な情報はすべて `haul --help` にあります。[llms.txt](llms.txt) は同じガイドのファイル版です。要点は次のとおりです。

```sh
haul info --json "<url>"                                # 1. 確認
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. そのストリームを正確にダウンロード
haul --json -q 720p -c avc,m4a "<url>"                  #    または優先順位に任せる
```

- `--json` 使用時、**stdout は JSON 文書ひとつのみ**です。進捗とログは stderr に出ます。`files` には新規・既存を問わず出力ファイルが並びます。
- **端末がなければ対話しません。** 端末なしの `-i` は終了コード 2 で失敗し、代わりに使うオプションを示します。
- `info` のストリーム番号は haul が選ぶ順です。`info` とダウンロードには同じ `-q` / `-c` を渡してください。
- プレイリスト、シーズン、番組、複数動画の投稿に対する `info` はページ一覧を表示します。`-p <n>` を付けるとそのページのストリームと字幕も表示します。
- すでにディスクにあるページは `"status": "skipped", "reason": "exists"` と報告されるため、安全に再実行できます。

<details>
<summary><strong>JSON 応答例</strong> · ダウンロードが成功した場合</summary>

ダウンロードの出力:

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

これは `haul --json -q 360p -c avc,m4a …` の出力です。ここではキーを読む順に並べ、ストリーム一覧を選ばれたものだけに省略しています。実際の文書ではキーがアルファベット順に並び、すべてのストリームが含まれます。

</details>

失敗時は `"ok": false` と `"error": {"kind", "message", "exitCode"}` を出力し、それまでに完了したページは保持されます。

| 終了コード | 意味 | `error.kind` |
|---|---|---|
| 0 | 完了 | – |
| 1 | ダウンロードまたは抽出に失敗 | `failed` |
| 2 | 未対応のリンク、または無効なオプション値 | `input` |
| 3 | 必要なツール（ffmpeg、yt-dlp）がない | `dependency` |
| 4 | ログインが必要、または期限切れ | `auth` |
| 64 | コマンドラインの誤り（不明なフラグ、引数の不足） | – |
| 130 | キャンセル | `cancelled` |

<a id="options"></a>

## オプション

すべてのオプションは `haul download --help` に一覧があります。主なもの:

| 分類 | オプション |
|---|---|
| 全般 | `--json`, `--config <file>`, `--debug` |
| ストリーム | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| ページ | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| 内容 | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| 出力 | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`。その他は `--help-hidden` で表示 |
| ツール | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>画質とコーデックの優先順位</strong></summary>

**画質とコーデック。** `-q` にはストリーム表に表示されるラベルを指定します。YouTube では `1080p`、`720p60`、bilibili では `8K`、`Dolby Vision`、`HDR`、`4K`、`1080P60`、`1080P+`、`1080P`、`720P` です。`-c` は動画に `av1 vp9 hevc avc`、音声に `m4a opus flac eac3 mp3` を受け付けます。これらは絞り込みではなく優先順位です。指定しなかったものはその後ろに、良いものから順に並びます。`--video-ascending` / `--audio-ascending` は順序を逆にし、最小サイズのダウンロードにします。

</details>

<details>
<summary><strong>ページ選択</strong></summary>

**ページ。** `8`、`1,2`、`3-5`、`1-3,10`、`ALL`、`LAST`（最後のページ。番組では最新回。`LATEST` も使えます）。単一エピソードへのリンクや `?p=N` は、そのページだけを選びます。ページとは、bilibili 動画のパート、シーズン・番組・リストのエピソード、X の投稿内の動画のことです。

</details>

<details>
<summary><strong>ファイル名とテンプレート変数</strong></summary>

**ファイル名。** ページがひとつの項目は `<title>`、複数ある項目は（`-p` でひとつだけ選んだ場合も）`<title>/[P<pageNumberWithZero>]<pageTitle>` で、番号はページ数に合わせてゼロ埋めされます。変数: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`、および bilibili 専用の `<bvid>` `<aid>` `<cid>` `<api>`。日付には書式を指定できます: `<publishDate:yyyy-MM-dd>`。拡張子は自動で付きます。

</details>

<a id="site-notes"></a><a id="notes"></a>

## サイト別の注意点

<details>
<summary><strong>YouTube</strong> · 抽出、字幕、プレイリスト</summary>

YouTube はストリーム URL を、JavaScript ランタイムが必要なプレーヤー検証で保護しています。そのため抽出は `yt-dlp -J`（検証を deno で実行）に任せ、それ以降はすべて haul 自身が処理します。アップロードされた字幕は結合されます。自動生成字幕は `--auto-subtitles` を指定した場合のみです。googlevideo は上限のあるバイト範囲しか返さないため、トラックは 10 MB ずつ範囲指定で取得します。`watch?v=…&list=…` はその動画だけをダウンロードします。yt-dlp は最新に保ってください。YouTube の変更への対応はそこに入ります。

</details>

<details>
<summary><strong>X</strong> · 公開投稿と音声</summary>

X の公開投稿にログインは不要です。動画は音声を含む単一の MP4 ファイルなので、表には動画のみが並びます。`--audio-only` で音声を取り出せます。

</details>

<details>
<summary><strong>bilibili</strong> · ログイン、高画質、弾幕</summary>

bilibili は独自の web、TV、APP（gRPC）、国際版 API から読み込みます。ログインしていない場合は低画質（通常 480P まで）のみです。1080P、4K、HDR、Dolby Vision、Hi-Res 音声にはログインしてください。

```sh
haul login bilibili                # bilibili アプリで QR コードをスキャン
haul login bilibili --from-edge    # macOS: Microsoft Edge のログインを再利用（または --from-chrome、--profile "Profile 1"）
haul login bilibili --tv           # TV アクセストークン。--api tv / --api app 用
```

ブラウザーのログインの読み取りは、今のところ macOS でのみ利用できます。端末に「フルディスクアクセス」が必要で、macOS が「Safe Storage」キーチェーン項目へのアクセスを一度だけ確認します。`--danmaku` は弾幕（コメント）を XML と ASS で保存します。`--danmaku-format ass` を指定すると片方だけを残します。

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · エピソードとメタデータ</summary>

Xiaoyuzhou と Apple Podcasts に yt-dlp は不要です。Xiaoyuzhou のページには音声リンクが含まれ、Apple のリンクは公開 iTunes API と番組の RSS フィードを経由します。エピソードは元の形式（`.mp3` または `.m4a`）のまま、カバーを埋め込み、番組名をアルバム、ホストをアーティストに設定します。番組リンクは最近のエピソードを古い順に並べるので、`-p LAST` が最新回です。

</details>

<a id="config"></a>

## 設定

`~/.config/haul/`（または `$HAUL_HOME`）には次のファイルがあります。

| ファイル | 用途 |
|---|---|
| `config.json` | 任意のオプションの既定値 |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili のログイン |
| `archives.txt` | ダウンロード済みのページ（`--archive`） |

`config.json` ではオプション名を camelCase で書き、bilibili 用のものは `"bilibili"` の下に置きます。変更する項目だけを書けばよく、リストは配列でもカンマ区切りの文字列でもかまいません。コマンドラインのフラグが優先されます。

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

## 開発

```sh
go build ./cmd/haul
go test ./...                                  # オフライン。数秒で完了
HAUL_LIVE=1 go test -run Live ./internal/...   # bilibili、YouTube、X、Apple Podcasts にも接続
```

オフラインのテストはネットワークに一切触れません。テストはスタブの `http.RoundTripper` と通信し、それが記録済みの API 応答と模擬 CDN（範囲リクエストのみのサーバー、切断される接続、範囲を無視するサーバー）から応答します。エンドツーエンドのテストは実際の ffmpeg で結合し、結果を ffprobe で検証します。ffmpeg がインストールされていない場合はスキップされます。アーキテクチャと規約は [AGENTS.md](AGENTS.md) を参照してください。

`v*` タグをプッシュすると、GitHub Actions がテスト、クロスコンパイルを行い、全プラットフォーム向けのリリースを公開します。

<a id="acknowledgements"></a>

## 謝辞

[yt-dlp](https://github.com/yt-dlp/yt-dlp)、[FFmpeg](https://ffmpeg.org)、[GPAC](https://gpac.io)、[aria2](https://aria2.github.io)、[pflag](https://github.com/spf13/pflag)、[x/term](https://pkg.go.dev/golang.org/x/term)、[rsc.io/qr](https://pkg.go.dev/rsc.io/qr)、そして [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) と [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api) の API 資料に感謝します。

<a id="license"></a>

## ライセンス

[MIT](LICENSE)

---

<p align="center">
  <strong>リンクひとつで、メディアを手元に。</strong><br>
  <a href="#install">haul を入手</a> · <a href="llms.txt">エージェント向けリファレンス</a> · <a href="AGENTS.md">コントリビュート</a>
</p>
