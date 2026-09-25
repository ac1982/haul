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
  <bdi><strong>বাংলা</strong></bdi> ·
  <a href="README.pt.md"><bdi>Português</bdi></a> ·
  <a href="README.id.md"><bdi>Bahasa Indonesia</bdi></a> ·
  <a href="README.ur.md"><bdi>اردو</bdi></a> ·
  <a href="README.ru.md"><bdi>Русский</bdi></a> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">একটি লিংক। আপনার মিডিয়া।</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="বিল্ডের অবস্থা"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="সর্বশেষ রিলিজ"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 বা নতুনতর">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux ও Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT লাইসেন্স"></a>
</p>

<p align="center">
  <strong>ভিডিও ও অডিও ডাউনলোডের কমান্ড-লাইন টুল, মানুষ ও AI এজেন্ট দুয়ের জন্যই তৈরি।</strong><br>
  একটি লিংক দিন। স্ট্রিম বেছে নিন। ভিডিও, অডিও, সাবটাইটেল, অধ্যায় ও প্রচ্ছদ পান একটি ফাইলে।
</p>

<p align="center">
  <a href="#install">ইনস্টলেশন</a> ·
  <a href="#usage">দ্রুত শুরু</a> ·
  <a href="#sites">সমর্থিত সাইট</a> ·
  <a href="#for-ai-agents-and-scripts">এজেন্ট নির্দেশিকা</a> ·
  <a href="#options">বিকল্প</a> ·
  <a href="../../releases">রিলিজ</a>
</p>

---

<a id="small-command-complete-download"></a>

## ছোট কমান্ড। সম্পূর্ণ ডাউনলোড।

| ↓ আপনার মিডিয়া, আপনার মতো করে | ⌘ একটি বাইনারি, সব ডেস্কটপ | { } অটোমেশনের জন্য প্রস্তুত |
| :--- | :--- | :--- |
| পাঁচটি সাইটেই একই শব্দভান্ডারে মান, কোডেক, পৃষ্ঠা ও ফাইলের নাম বাছুন | macOS, Linux ও Windows-এর জন্য একটি Go বাইনারি; সমান্তরাল রেঞ্জ ডাউনলোড, যা থেমে গেলে আবার চালু করা যায় | stdout-এ একটি JSON নথি, stderr-এ লগ, অর্থবহ এক্সিট কোড; টার্মিনাল না থাকলে কোনো প্রশ্ন নয় |

| 01 / পরীক্ষা | 02 / বাছাই | 03 / ডাউনলোড | 04 / একত্রীকরণ |
| :--- | :--- | :--- | :--- |
| **প্রতিটি স্ট্রিম দেখুন** | **আপনার অগ্রাধিকার ঠিক করুন** | **ঘরে নিয়ে আসুন** | **সব একসাথে** |
| পৃষ্ঠা, কোডেক ও আকার | মান, অডিও ও পৃষ্ঠা | রেঞ্জ বা সেগমেন্টে | ট্র্যাক, সাবটাইটেল ও প্রচ্ছদ |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>টার্মিনালের ভেতরে এক নজর</strong></summary>

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

> ব্যক্তিগত, গবেষণা ও অন্যান্য অবাণিজ্যিক ব্যবহারের জন্য। কপিরাইট ও প্রতিটি সাইটের শর্ত মেনে চলার দায়িত্ব আপনার।

<a id="install"></a>

## ইনস্টলেশন

haul চলে **macOS, Linux ও Windows**-এ (amd64 ও arm64)। ট্র্যাক একত্র (mux) করতে [ffmpeg](https://ffmpeg.org) লাগে; YouTube ও X-এর জন্য [yt-dlp](https://github.com/yt-dlp/yt-dlp)-ও লাগে, আর YouTube প্লেয়ারের যাচাইয়ের জন্য লাগে [deno](https://deno.com):

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg yt-dlp           # Debian / Ubuntu; deno: https://deno.com
winget install Gyan.FFmpeg               # Windows, একবারে একটি প্যাকেজ
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

<a id="get-the-binary"></a><a id="binary"></a>

### বাইনারি সংগ্রহ

[Releases](../../releases) থেকে আপনার সিস্টেমের আর্কাইভ ডাউনলোড করুন — `haul-<version>-<os>-<arch>.tar.gz` (Windows-এ `.zip`) — এবং `haul` আপনার `PATH`-এর যেকোনো ফোল্ডারে রাখুন:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
xattr -d com.apple.quarantine ~/.local/bin/haul   # macOS: বাইনারিটি নোটারাইজড নয়
```

<details>
<summary><strong>সোর্স থেকে বিল্ড</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

অথবা ক্লোন করা রিপোজিটরি থেকে:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## দ্রুত শুরু

একটি লিংক দিয়ে শুরু করুন। haul ডিফল্টভাবে সেরা উপলব্ধ স্ট্রিম বেছে নেয়।

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

ডাউনলোডের আগে তথ্য দেখুন, অথবা মান ও কোডেক বেছে নিন:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # QuickTime-এর জন্য H.264 + AAC
```

<details>
<summary><strong>কমান্ড নির্দেশিকা</strong></summary>

| কমান্ড | অর্থ |
| --- | --- |
| `haul <url> [options]` | ডাউনলোড (`haul download <url>`-এর সমান) |
| `haul info <url> [--urls]` | আইটেম, তার পৃষ্ঠা ও স্ট্রিম দেখায়; কিছু ডাউনলোড করে না |
| `haul login bilibili` | উচ্চতর মানের জন্য bilibili-তে লগ ইন |
| `haul templates` | ফাইলের নামের টেমপ্লেট ভেরিয়েবল |
| `haul --help` | সাইট, এজেন্টের ব্যবহার, এক্সিট কোড, উদাহরণ |
| `haul download --help` | সব বিকল্প |

</details>

<a id="sites"></a>

### সমর্থিত সাইট

| সাইট | লিংক | প্রয়োজন |
| --- | --- | --- |
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` পোস্টের একটি ভিডিও বাছে | yt-dlp |
| **bilibili** | ভিডিও, bangumi, কোর্স, সংকলন, সিরিজ, পছন্দের তালিকা, ব্যবহারকারীর পাতা, `b23.tv`, সরাসরি `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`; একটি পর্বের জন্য `?i=<episode>` সহ | – |

<a id="make-it-yours"></a><a id="examples"></a>

### নিজের মতো করে নিন

**শুধু অডিও**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # Opus-এর বদলে AAC
```

**কয়েকটি অংশ, পুরো সিজন বা সর্বশেষ পর্ব**

```sh
# বাছাই করা অংশ, আপনার Movies ফোল্ডারে
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# সিজনের সব পর্ব
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# পডকাস্টের সবচেয়ে নতুন পর্ব
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# একটি X পোস্টের সব ভিডিও
haul -p ALL "https://x.com/<user>/status/<id>"
```

**গোছানো ফাইল, সাবটাইটেল বা ইন্টার‌্যাক্টিভ বাছাই**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # শুধু এই ভাষাগুলোর সাবটাইটেল
haul -i "BV1qt4y1X7TW"                                 # তীর-কী দিয়ে স্ট্রিম বাছুন
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## AI এজেন্ট ও স্ক্রিপ্টের জন্য

একটি এজেন্টের যা দরকার সবই `haul --help`-এ আছে; [llms.txt](llms.txt) একই নির্দেশিকার ফাইল রূপ। সংক্ষেপে:

```sh
haul info --json "<url>"                                # 1. তথ্য দেখুন
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. ঠিক এই স্ট্রিমগুলো ডাউনলোড করুন
haul --json -q 720p -c avc,m4a "<url>"                  #    অথবা অগ্রাধিকার অনুযায়ী বাছাই হতে দিন
```

- `--json` দিলে **stdout-এ কেবল একটি JSON নথি** আসে; অগ্রগতি ও লগ যায় stderr-এ। এর `files` তালিকায় থাকে আউটপুট ফাইলগুলো, নতুন হোক বা আগে থেকে থাকা।
- **টার্মিনাল ছাড়া কিছুই ইন্টার‌্যাক্টিভ নয়।** টার্মিনাল না থাকলে `-i` এক্সিট কোড 2 দিয়ে ব্যর্থ হয় এবং বদলে কোন ফ্ল্যাগ ব্যবহার করতে হবে তা জানায়।
- `info`-তে স্ট্রিমের সূচক haul যে ক্রমে বাছাই করে সেই ক্রমে থাকে; `info` ও ডাউনলোড দুটিতেই একই `-q` / `-c` দিন।
- প্লেলিস্ট, সিজন, শো বা একাধিক ভিডিওর পোস্টে `info` তার পৃষ্ঠাগুলো দেখায়; `-p <n>` সেই পৃষ্ঠার স্ট্রিম ও সাবটাইটেলও যোগ করে।
- ডিস্কে আগে থেকে থাকা পৃষ্ঠা `"status": "skipped", "reason": "exists"` হিসেবে জানানো হয়, তাই আবার চালানো নিরাপদ।

<details>
<summary><strong>JSON উত্তরের উদাহরণ</strong> · একটি সফল ডাউনলোড</summary>

একটি ডাউনলোড এটি ছাপে:

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

এটি `haul --json -q 360p -c avc,m4a …`-এর উত্তর। এখানে কী-গুলো পড়ার ক্রমে সাজানো এবং স্ট্রিমের তালিকা কেবল বাছাই করা স্ট্রিমে ছোট করা হয়েছে; আসল নথি তার কী-গুলো বর্ণানুক্রমে সাজায় এবং প্রতিটি স্ট্রিম দেখায়।

</details>

ব্যর্থ হলে `"ok": false` সহ `"error": {"kind", "message", "exitCode"}` ছাপা হয়, এবং তার আগে শেষ হওয়া পৃষ্ঠাগুলো উত্তরে থেকে যায়।

| এক্সিট কোড | অর্থ | `error.kind` |
| --- | --- | --- |
| 0 | সম্পন্ন | – |
| 1 | ডাউনলোড বা এক্সট্র্যাকশন ব্যর্থ | `failed` |
| 2 | অসমর্থিত লিংক বা ভুল বিকল্প-মান | `input` |
| 3 | প্রয়োজনীয় টুল নেই (ffmpeg, yt-dlp) | `dependency` |
| 4 | লগইন দরকার বা মেয়াদ শেষ | `auth` |
| 64 | ভুল কমান্ড লাইন (অজানা ফ্ল্যাগ, অনুপস্থিত আর্গুমেন্ট) | – |
| 130 | বাতিল | `cancelled` |

<a id="options"></a>

## বিকল্প

`haul download --help` সব বিকল্প দেখায়। প্রধানগুলো:

| শ্রেণি | বিকল্প |
| --- | --- |
| সাধারণ | `--json`, `--config <file>`, `--debug` |
| স্ট্রিম | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| পৃষ্ঠা | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| বিষয়বস্তু | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| আউটপুট | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; আরও বিকল্প `--help-hidden`-এ |
| টুল | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>মান ও কোডেকের অগ্রাধিকার</strong></summary>

**মান ও কোডেক।** `-q` স্ট্রিম টেবিলে দেখানো লেবেলগুলো নেয়: YouTube-এ `1080p`, `720p60`; bilibili-তে `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P`। `-c` ভিডিওর জন্য `av1 vp9 hevc avc` এবং অডিওর জন্য `m4a opus flac eac3 mp3` নেয়। এগুলো অগ্রাধিকার, ফিল্টার নয়: তালিকায় যা নেই তা পরে আসে, সেরাটি আগে। `--video-ascending` / `--audio-ascending` সবচেয়ে ছোট ডাউনলোডের জন্য ক্রম উল্টে দেয়।

</details>

<details>
<summary><strong>পৃষ্ঠা নির্বাচন</strong></summary>

**পৃষ্ঠা।** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (শেষ পৃষ্ঠা, অর্থাৎ শো-এর সবচেয়ে নতুন পর্ব; `LATEST`-ও কাজ করে)। একটি পর্বের লিংক, বা `?p=N`, সেই পৃষ্ঠাটি নিজেই বেছে নেয়। পৃষ্ঠা মানে bilibili ভিডিওর অংশ, সিজন, শো বা তালিকার পর্ব, এবং X পোস্টের ভিডিও।

</details>

<details>
<summary><strong>ফাইলের নাম ও টেমপ্লেট ভেরিয়েবল</strong></summary>

**ফাইলের নাম।** এক পৃষ্ঠার আইটেম: `<title>`; একাধিক পৃষ্ঠার (এমনকি `-p` একটি নিলেও): `<title>/[P<pageNumberWithZero>]<pageTitle>`, মোট পৃষ্ঠাসংখ্যা অনুযায়ী সামনে শূন্য বসিয়ে। ভেরিয়েবল: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, সঙ্গে bilibili-র `<bvid>` `<aid>` `<cid>` `<api>`। তারিখে ফরম্যাট দেওয়া যায়: `<publishDate:yyyy-MM-dd>`। এক্সটেনশন নিজে থেকেই যোগ হয়।

</details>

<a id="site-notes"></a><a id="notes"></a>

## সাইটভিত্তিক তথ্য

<details>
<summary><strong>YouTube</strong> · এক্সট্র্যাকশন, সাবটাইটেল ও প্লেলিস্ট</summary>

YouTube তার স্ট্রিম URL প্লেয়ারের এমন যাচাই দিয়ে সুরক্ষিত রাখে যার জন্য JavaScript রানটাইম লাগে, তাই এক্সট্র্যাকশন `yt-dlp -J`-কে দেওয়া হয় (যা সেগুলো deno-তে চালায়); তার পরের সব কাজ haul নিজেই করে। আপলোড করা সাবটাইটেল ফাইলে যুক্ত হয়; স্বয়ংক্রিয়ভাবে তৈরি সাবটাইটেল কেবল `--auto-subtitles` দিলে। googlevideo কেবল সীমিত বাইট রেঞ্জ দেয়, তাই ট্র্যাক একবারে 10 MB-এর একটি রেঞ্জ করে আনা হয়। `watch?v=…&list=…` শুধু সেই ভিডিওটি ডাউনলোড করে। yt-dlp হালনাগাদ রাখুন: YouTube-এর পরিবর্তনের সমাধান সেখানেই আসে।

</details>

<details>
<summary><strong>X</strong> · সর্বজনীন পোস্ট ও অডিও</summary>

X-এর সর্বজনীন পোস্টের জন্য লগইন লাগে না। এর ভিডিও অডিওসহ একটিমাত্র MP4 ফাইল, তাই টেবিলে কেবল ভিডিও দেখায়; `--audio-only` অডিও বের করে।

</details>

<details>
<summary><strong>bilibili</strong> · লগইন, উচ্চতর মান ও danmaku</summary>

bilibili পড়া হয় তার নিজস্ব web, TV, APP (gRPC) ও আন্তর্জাতিক API দিয়ে। লগ আউট অবস্থায় কেবল নিম্ন মান পাওয়া যায় (সাধারণত 480P পর্যন্ত); 1080P, 4K, HDR, Dolby Vision ও Hi-Res অডিওর জন্য লগ ইন করুন:

```sh
haul login bilibili                # bilibili অ্যাপ দিয়ে QR কোড স্ক্যান করুন
haul login bilibili --from-edge    # macOS: Microsoft Edge-এর লগইন আবার ব্যবহার করুন (বা --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # TV অ্যাক্সেস টোকেন, --api tv / --api app-এর জন্য
```

ব্রাউজারের লগইন পড়া আপাতত শুধু macOS-এ কাজ করে: এর জন্য টার্মিনালের Full Disk Access লাগে, এবং macOS একবার “Safe Storage” কিচেন আইটেমের অনুমতি চায়। `--danmaku` পর্দায় ভেসে চলা মন্তব্য (danmaku) XML ও ASS হিসেবে সংরক্ষণ করে; `--danmaku-format ass` এর মধ্যে কেবল একটি রাখে।

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · পর্ব ও মেটাডেটা</summary>

Xiaoyuzhou ও Apple Podcasts-এর yt-dlp লাগে না: Xiaoyuzhou-এর পৃষ্ঠায় অডিওর লিংক থাকে, আর Apple-এর লিংক সর্বজনীন iTunes API ও শো-এর RSS ফিডের মাধ্যমে যায়। পর্বগুলো নিজের ফরম্যাট (`.mp3` বা `.m4a`) বজায় রাখে; প্রচ্ছদ যুক্ত হয়, শো হয় অ্যালবাম আর উপস্থাপক শিল্পী। শো-এর লিংক সাম্প্রতিক পর্বগুলো পুরোনো থেকে নতুন ক্রমে দেখায়, তাই `-p LAST` হলো সবচেয়ে নতুন পর্ব।

</details>

<a id="config"></a>

## কনফিগারেশন

`~/.config/haul/` (বা `$HAUL_HOME`)-এ থাকে:

| ফাইল | উদ্দেশ্য |
| --- | --- |
| `config.json` | যেকোনো বিকল্পের ডিফল্ট মান |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili লগইন |
| `archives.txt` | আগে ডাউনলোড করা পৃষ্ঠা (`--archive`) |

`config.json` বিকল্পের নাম camelCase-এ ব্যবহার করে, bilibili-র বিকল্প `"bilibili"`-এর ভেতরে, এবং এতে শুধু যা বদলাতে চান তা-ই লিখতে হয়; তালিকা একটি array বা কমা দিয়ে আলাদা করা string হতে পারে। কমান্ড-লাইন ফ্ল্যাগ একে ওভাররাইড করে:

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

## ডেভেলপমেন্ট

```sh
go build ./cmd/haul
go test ./...                                  # অফলাইন, কয়েক সেকেন্ড
HAUL_LIVE=1 go test -run Live ./internal/...   # bilibili, YouTube, X ও Apple Podcasts-এও সংযোগ করে
```

অফলাইন টেস্ট কখনো নেটওয়ার্ক ছোঁয় না: টেস্টগুলো একটি স্টাব `http.RoundTripper`-এর সঙ্গে কথা বলে, যা রেকর্ড করা API উত্তর এবং অনুকরণ করা CDN (শুধু রেঞ্জ দেওয়া সার্ভার, বিচ্ছিন্ন সংযোগ, রেঞ্জ উপেক্ষা করা সার্ভার) থেকে উত্তর দেয়। শুরু থেকে শেষ পর্যন্ত টেস্টগুলো আসল ffmpeg দিয়ে ট্র্যাক একত্র করে এবং ffprobe দিয়ে ফলাফল যাচাই করে; ffmpeg ইনস্টল না থাকলে সেগুলো বাদ যায়। কাঠামো ও রীতির জন্য [AGENTS.md](AGENTS.md) দেখুন।

একটি `v*` ট্যাগ পুশ করলে GitHub Actions টেস্ট চালায়, প্রতিটি প্ল্যাটফর্মের জন্য ক্রস-কম্পাইল করে এবং রিলিজ প্রকাশ করে।

<a id="acknowledgements"></a>

## কৃতজ্ঞতা

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr), এবং [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) ও [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api)-এর API নোট।

<a id="license"></a>

## লাইসেন্স

[MIT](LICENSE)

---

<p align="center">
  <strong>একটি লিংক। আপনার মিডিয়া।</strong><br>
  <a href="#install">haul সংগ্রহ করুন</a> · <a href="llms.txt">এজেন্ট রেফারেন্স</a> · <a href="AGENTS.md">অবদান রাখুন</a>
</p>
