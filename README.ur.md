<!-- Translation of README.md; maintenance: docs/TRANSLATING.md -->

<div dir="rtl">

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
  <bdi><strong>اردو</strong></bdi> ·
  <a href="README.ru.md"><bdi>Русский</bdi></a> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <a href="README.arz.md"><bdi>مصري</bdi></a>
</p>
<!-- languages:end -->

<h1 align="center">ایک لنک۔ آپ کا میڈیا۔</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="بلڈ کی حالت"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="تازہ ترین ریلیز"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 یا نیا">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS، Linux اور Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT لائسنس"></a>
</p>

<p align="center">
  <strong>ویڈیو اور آڈیو ڈاؤن لوڈ کرنے کا کمانڈ لائن ٹول، جو لوگوں اور AI ایجنٹس دونوں کے لیے بنا ہے۔</strong><br>
  لنک دیں۔ اسٹریم منتخب کریں۔ ویڈیو، آڈیو، سب ٹائٹلز، ابواب اور سرورق ایک فائل میں حاصل کریں۔
</p>

<p align="center">
  <a href="#install">تنصیب</a> ·
  <a href="#usage">فوری آغاز</a> ·
  <a href="#sites">معاون سائٹس</a> ·
  <a href="#for-ai-agents-and-scripts">ایجنٹ گائیڈ</a> ·
  <a href="#options">اختیارات</a> ·
  <a href="../../releases">ریلیزز</a>
</p>

---

<a id="small-command-complete-download"></a>

## چھوٹی کمانڈ۔ مکمل ڈاؤن لوڈ۔

| ↓ آپ کا میڈیا، آپ کے انداز میں | ⌘ ایک بائنری، ہر ڈیسک ٹاپ | { } آٹومیشن کے لیے تیار |
| :--- | :--- | :--- |
| پانچوں سائٹس پر ایک ہی اصطلاحات سے معیار، کوڈیک، صفحات اور فائل کے نام منتخب کریں | macOS، Linux اور Windows کے لیے ایک Go بائنری، جو متوازی رینج ڈاؤن لوڈ کرتی ہے اور رکے ہوئے ڈاؤن لوڈ دوبارہ جاری کرتی ہے | stdout پر ایک JSON دستاویز، stderr پر لاگز، بامعنی ایگزٹ کوڈ؛ ٹرمینل کے بغیر کوئی سوال نہیں |

| 01 / جائزہ | 02 / انتخاب | 03 / ڈاؤن لوڈ | 04 / یکجا |
| :--- | :--- | :--- | :--- |
| **ہر اسٹریم دیکھیں** | **اپنی ترجیحات طے کریں** | **گھر لے آئیں** | **سب کچھ ایک ساتھ** |
| صفحات، کوڈیک اور سائز | معیار، آڈیو اور صفحات | رینج یا سیگمنٹ میں | ٹریک، سب ٹائٹلز اور سرورق |
| <bdi dir="ltr">`haul info "<url>"`</bdi> | <bdi dir="ltr">`-q 720p -c avc,m4a`</bdi> | <bdi dir="ltr">`haul "<url>"`</bdi> | <bdi dir="ltr">`ffmpeg / MP4Box`</bdi> |

<details>
<summary><strong>ٹرمینل کے اندر ایک جھلک</strong></summary>

<div dir="ltr">

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

</div>

</details>

> ذاتی، تحقیقی اور دیگر غیر تجارتی استعمال کے لیے۔ حقوقِ اشاعت اور ہر سائٹ کی شرائط کا احترام آپ کی ذمہ داری ہے۔

<a id="install"></a>

## تنصیب

haul **macOS، Linux اور Windows** (amd64 اور arm64) پر چلتا ہے۔ ٹریک یکجا (mux) کرنے کے لیے [ffmpeg](https://ffmpeg.org) درکار ہے؛ YouTube اور X کے لیے [yt-dlp](https://github.com/yt-dlp/yt-dlp) بھی، اور YouTube پلیئر کی جانچ کے لیے [deno](https://deno.com) چاہیے:

<div dir="ltr">

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip       # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows، ایک وقت میں ایک پیکیج
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

</div>

Linux پر yt-dlp کو ڈسٹری بیوشن کے پیکیج سے نہیں بلکہ pipx سے، یا اس کی [ریلیزز](https://github.com/yt-dlp/yt-dlp/releases/latest) سے <bdi dir="ltr">`yt-dlp_linux`</bdi> (arm64 پر <bdi dir="ltr">`yt-dlp_linux_aarch64`</bdi>) لے کر انسٹال کریں: YouTube اکثر بدلتا رہتا ہے اور پیکیج والے ورژن پیچھے رہ جاتے ہیں۔ انسٹال کرنے کے بعد نیا ٹرمینل کھولیں تاکہ یہ ٹولز <bdi dir="ltr">`PATH`</bdi> میں آ جائیں۔

<a id="get-the-binary"></a><a id="binary"></a>

### بائنری حاصل کریں

[Releases](../../releases) سے اپنے سسٹم کا آرکائیو ڈاؤن لوڈ کریں — <bdi dir="ltr">`haul-<version>-<os>-<arch>.tar.gz`</bdi> (Windows پر <bdi dir="ltr">`.zip`</bdi>) — اور <bdi dir="ltr">`haul`</bdi> کو اپنے <bdi dir="ltr">`PATH`</bdi> کے کسی بھی فولڈر میں رکھیں:

<div dir="ltr">

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

</div>

موجودہ ریلیز طریقہ macOS ورژن پر دستخط کرتا ہے اور Apple سے نوٹرائزیشن حاصل کرتا ہے۔ <bdi dir="ltr">`haul-<version>-darwin-<arch>.pkg`</bdi> منتخب کریں؛ اس کے ساتھ نوٹرائزیشن ٹکٹ منسلک ہے اور یہ <bdi dir="ltr">`haul`</bdi> کو <bdi dir="ltr">`/usr/local/bin`</bdi> میں نصب کرتا ہے۔ آرکائیو میں وہی دستخط شدہ پروگرام ہے، مگر پہلی تصدیق کے لیے انٹرنیٹ درکار ہو سکتا ہے۔ پرانے ورژن غیر دستخط شدہ ہو سکتے ہیں۔

<details>
<summary><strong>سورس سے بلڈ کریں</strong> · Go 1.26+</summary>

<div dir="ltr">

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

</div>

یا کلون کی گئی ریپوزٹری سے:

<div dir="ltr">

```sh
go build -o haul ./cmd/haul
```

</div>

</details>

<a id="usage"></a>

## فوری آغاز

ایک لنک سے شروع کریں۔ haul طے شدہ طور پر دستیاب بہترین اسٹریم منتخب کرتا ہے۔

<div dir="ltr">

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

</div>

ڈاؤن لوڈ سے پہلے معلومات دیکھیں، یا معیار اور کوڈیک منتخب کریں:

<div dir="ltr">

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # QuickTime کے لیے H.264 + AAC
```

</div>

<details>
<summary><strong>کمانڈ کا حوالہ</strong></summary>

| کمانڈ | مطلب |
| --- | --- |
| <bdi dir="ltr">`haul <url> [options]`</bdi> | ڈاؤن لوڈ (<bdi dir="ltr">`haul download <url>`</bdi> کے برابر) |
| <bdi dir="ltr">`haul info <url> [--urls]`</bdi> | آئٹم، اس کے صفحات اور اسٹریم دکھاتا ہے؛ کچھ ڈاؤن لوڈ نہیں کرتا |
| <bdi dir="ltr">`haul login bilibili`</bdi> | بہتر معیار کے لیے bilibili میں لاگ ان |
| <bdi dir="ltr">`haul templates`</bdi> | فائل نام کے ٹیمپلیٹ متغیرات |
| <bdi dir="ltr">`haul --help`</bdi> | سائٹس، ایجنٹ استعمال، ایگزٹ کوڈ، مثالیں |
| <bdi dir="ltr">`haul download --help`</bdi> | تمام اختیارات |

</details>

<a id="sites"></a>

### معاون سائٹس

| سائٹ | لنکس | ضرورت |
| --- | --- | --- |
| **YouTube** | <bdi dir="ltr">`youtube.com/watch?v=…`</bdi>, <bdi dir="ltr">`youtu.be/…`</bdi>, <bdi dir="ltr">`/shorts/…`</bdi>, <bdi dir="ltr">`/embed/…`</bdi>, <bdi dir="ltr">`/live/…`</bdi> | yt-dlp, deno |
| **X** | <bdi dir="ltr">`x.com/<user>/status/<id>`</bdi>, <bdi dir="ltr">`twitter.com/…`</bdi>؛ <bdi dir="ltr">`/video/<n>`</bdi> پوسٹ کی ایک ویڈیو منتخب کرتا ہے | yt-dlp |
| **bilibili** | ویڈیوز، bangumi، کورسز، مجموعے، سیریز، پسندیدہ، صارف صفحات، <bdi dir="ltr">`b23.tv`</bdi>، براہ راست <bdi dir="ltr">`BV…`</bdi> <bdi dir="ltr">`av…`</bdi> <bdi dir="ltr">`ep…`</bdi> <bdi dir="ltr">`ss…`</bdi> <bdi dir="ltr">`md…`</bdi> | – |
| **Xiaoyuzhou** | <bdi dir="ltr">`xiaoyuzhoufm.com/episode/<id>`</bdi>, <bdi dir="ltr">`/podcast/<id>`</bdi> | – |
| **Apple Podcasts** | <bdi dir="ltr">`podcasts.apple.com/<cc>/podcast/<name>/id<show>`</bdi>؛ ایک قسط کے لیے <bdi dir="ltr">`?i=<episode>`</bdi> کے ساتھ | – |

<a id="make-it-yours"></a><a id="examples"></a>

### اپنی مرضی کے مطابق

**صرف آڈیو**

<div dir="ltr">

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # Opus کے بجائے AAC
```

</div>

**چند حصے، پورا سیزن یا تازہ ترین قسط**

<div dir="ltr">

```sh
# منتخب حصے، آپ کے Movies فولڈر میں
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# سیزن کی تمام اقساط
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# پوڈکاسٹ کی سب سے نئی قسط
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# X پوسٹ کی تمام ویڈیوز
haul -p ALL "https://x.com/<user>/status/<id>"
```

</div>

**منظم فائلیں، سب ٹائٹلز یا تعاملی انتخاب**

<div dir="ltr">

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # صرف ان زبانوں کے سب ٹائٹلز
haul -i "BV1qt4y1X7TW"                                 # تیر والی کلیدوں سے اسٹریم منتخب کریں
```

</div>

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## AI ایجنٹس اور اسکرپٹس کے لیے

ایجنٹ کو جو کچھ چاہیے وہ سب <bdi dir="ltr">`haul --help`</bdi> میں ہے؛ [llms.txt](llms.txt) اسی رہنمائی کا فائل نسخہ ہے۔ مختصراً:

<div dir="ltr">

```sh
haul info --json "<url>"                                # 1. معلومات دیکھیں
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. بالکل یہی اسٹریم ڈاؤن لوڈ کریں
haul --json -q 720p -c avc,m4a "<url>"                  #    یا ترجیحات کو انتخاب کرنے دیں
```

</div>

- <bdi dir="ltr">`--json`</bdi> کے ساتھ **stdout پر صرف ایک JSON دستاویز** آتی ہے؛ پیش رفت اور لاگز stderr پر جاتے ہیں۔ اس کی <bdi dir="ltr">`files`</bdi> فہرست میں آؤٹ پٹ فائلیں ہوتی ہیں، نئی یا پہلے سے موجود۔
- **ٹرمینل کے بغیر کچھ بھی تعاملی نہیں۔** ٹرمینل نہ ہو تو <bdi dir="ltr">`-i`</bdi> ایگزٹ کوڈ 2 کے ساتھ ناکام ہوتا ہے اور اس کی جگہ استعمال ہونے والے فلیگ بتاتا ہے۔
- <bdi dir="ltr">`info`</bdi> میں اسٹریم انڈیکس اسی ترتیب میں ہوتے ہیں جس میں haul انتخاب کرتا ہے؛ <bdi dir="ltr">`info`</bdi> اور ڈاؤن لوڈ دونوں کو وہی <bdi dir="ltr">`-q`</bdi> / <bdi dir="ltr">`-c`</bdi> دیں۔
- پلے لسٹ، سیزن، شو یا کئی ویڈیوز والی پوسٹ پر <bdi dir="ltr">`info`</bdi> اس کے صفحات دکھاتا ہے؛ <bdi dir="ltr">`-p <n>`</bdi> اس صفحے کے اسٹریم اور سب ٹائٹلز بھی شامل کرتا ہے۔
- ڈسک پر پہلے سے موجود صفحات <bdi dir="ltr">`"status": "skipped", "reason": "exists"`</bdi> کے طور پر بتائے جاتے ہیں، اس لیے دوبارہ چلانا محفوظ ہے۔

<details>
<summary><strong>JSON جواب کی مثال</strong> · ایک کامیاب ڈاؤن لوڈ</summary>

ڈاؤن لوڈ یہ پرنٹ کرتا ہے:

<div dir="ltr">

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

</div>

یہ <bdi dir="ltr">`haul --json -q 360p -c avc,m4a …`</bdi> کا جواب ہے۔ یہاں کلیدیں پڑھنے کی ترتیب میں ہیں اور اسٹریم کی فہرستیں صرف منتخب اسٹریم تک مختصر کی گئی ہیں؛ اصل دستاویز اپنی کلیدیں حروفِ تہجی کی ترتیب سے رکھتی ہے اور ہر اسٹریم دکھاتی ہے۔

</details>

ناکامی پر <bdi dir="ltr">`"ok": false`</bdi> کے ساتھ <bdi dir="ltr">`"error": {"kind", "message", "exitCode"}`</bdi> پرنٹ ہوتا ہے، اور اس سے پہلے مکمل ہونے والے صفحات جواب میں برقرار رہتے ہیں۔

| ایگزٹ کوڈ | مطلب | <bdi dir="ltr">`error.kind`</bdi> |
| --- | --- | --- |
| 0 | مکمل | – |
| 1 | ڈاؤن لوڈ یا اخذ کرنے میں ناکامی | <bdi dir="ltr">`failed`</bdi> |
| 2 | غیر معاون لنک یا اختیار کی غلط قدر | <bdi dir="ltr">`input`</bdi> |
| 3 | ضروری ٹول موجود نہیں (ffmpeg، yt-dlp) | <bdi dir="ltr">`dependency`</bdi> |
| 4 | لاگ ان ضروری ہے یا مدت ختم ہو گئی | <bdi dir="ltr">`auth`</bdi> |
| 64 | غلط کمانڈ لائن (نامعلوم فلیگ، غائب آرگیومنٹ) | – |
| 130 | منسوخ | <bdi dir="ltr">`cancelled`</bdi> |

<a id="options"></a>

## اختیارات

<bdi dir="ltr">`haul download --help`</bdi> تمام اختیارات دکھاتا ہے۔ اہم اختیارات:

| گروپ | اختیارات |
| --- | --- |
| عمومی | <bdi dir="ltr">`--json`</bdi>, <bdi dir="ltr">`--config <file>`</bdi>, <bdi dir="ltr">`--debug`</bdi> |
| اسٹریم | <bdi dir="ltr">`-q, --quality <list>`</bdi>, <bdi dir="ltr">`-c, --codec <list>`</bdi>, <bdi dir="ltr">`--video-stream <n>`</bdi>, <bdi dir="ltr">`--audio-stream <n>`</bdi>, <bdi dir="ltr">`-i, --interactive`</bdi>, <bdi dir="ltr">`--video-ascending`</bdi>, <bdi dir="ltr">`--audio-ascending`</bdi> |
| صفحات | <bdi dir="ltr">`-p, --pages <spec>`</bdi>, <bdi dir="ltr">`--show-all`</bdi>, <bdi dir="ltr">`--hide-streams`</bdi> |
| مواد | <bdi dir="ltr">`--audio-only`</bdi>, <bdi dir="ltr">`--video-only`</bdi>, <bdi dir="ltr">`--subtitle-only`</bdi>, <bdi dir="ltr">`--cover-only`</bdi>, <bdi dir="ltr">`--skip-subtitle`</bdi>, <bdi dir="ltr">`--sub-lang <list>`</bdi>, <bdi dir="ltr">`--auto-subtitles`</bdi>, <bdi dir="ltr">`--skip-cover`</bdi>, <bdi dir="ltr">`--skip-mux`</bdi> |
| آؤٹ پٹ | <bdi dir="ltr">`-o, --output <template>`</bdi>, <bdi dir="ltr">`--multi-output <template>`</bdi>, <bdi dir="ltr">`-w, --work-dir <dir>`</bdi>, <bdi dir="ltr">`--lang <code>`</bdi>, <bdi dir="ltr">`--no-tags`</bdi>, <bdi dir="ltr">`--archive`</bdi>, <bdi dir="ltr">`--delay <seconds>`</bdi> |
| bilibili | <bdi dir="ltr">`--api web\|tv\|app\|intl`</bdi>, <bdi dir="ltr">`--danmaku`</bdi>, <bdi dir="ltr">`--danmaku-only`</bdi>, <bdi dir="ltr">`--danmaku-format xml,ass`</bdi>, <bdi dir="ltr">`--cookie`</bdi>, <bdi dir="ltr">`--token`</bdi>؛ مزید اختیارات <bdi dir="ltr">`--help-hidden`</bdi> میں |
| ٹولز | <bdi dir="ltr">`--ffmpeg`</bdi>, <bdi dir="ltr">`--yt-dlp`</bdi>, <bdi dir="ltr">`--use-mp4box`</bdi>, <bdi dir="ltr">`--mp4box`</bdi>, <bdi dir="ltr">`--use-aria2c`</bdi>, <bdi dir="ltr">`--aria2c`</bdi>, <bdi dir="ltr">`--aria2c-args`</bdi>, <bdi dir="ltr">`--single-connection`</bdi> |

<details>
<summary><strong>معیار اور کوڈیک کی ترجیحات</strong></summary>

**معیار اور کوڈیک۔** <bdi dir="ltr">`-q`</bdi> وہی لیبل لیتا ہے جو اسٹریم جدول دکھاتا ہے: YouTube پر <bdi dir="ltr">`1080p`</bdi>، <bdi dir="ltr">`720p60`</bdi>؛ bilibili پر <bdi dir="ltr">`8K`</bdi>، <bdi dir="ltr">`Dolby Vision`</bdi>، <bdi dir="ltr">`HDR`</bdi>، <bdi dir="ltr">`4K`</bdi>، <bdi dir="ltr">`1080P60`</bdi>، <bdi dir="ltr">`1080P+`</bdi>، <bdi dir="ltr">`1080P`</bdi>، <bdi dir="ltr">`720P`</bdi>۔ <bdi dir="ltr">`-c`</bdi> ویڈیو کے لیے <bdi dir="ltr">`av1 vp9 hevc avc`</bdi> اور آڈیو کے لیے <bdi dir="ltr">`m4a opus flac eac3 mp3`</bdi> لیتا ہے۔ یہ ترجیحات ہیں، فلٹر نہیں: جو فہرست میں نہیں وہ بعد میں آتا ہے، بہترین پہلے۔ <bdi dir="ltr">`--video-ascending`</bdi> / <bdi dir="ltr">`--audio-ascending`</bdi> سب سے چھوٹے ڈاؤن لوڈ کے لیے ترتیب الٹ دیتے ہیں۔

</details>

<details>
<summary><strong>صفحات کا انتخاب</strong></summary>

**صفحات۔** <bdi dir="ltr">`8`</bdi>، <bdi dir="ltr">`1,2`</bdi>، <bdi dir="ltr">`3-5`</bdi>، <bdi dir="ltr">`1-3,10`</bdi>، <bdi dir="ltr">`ALL`</bdi>، <bdi dir="ltr">`LAST`</bdi> (آخری صفحہ، یعنی شو کی سب سے نئی قسط؛ <bdi dir="ltr">`LATEST`</bdi> بھی چلتا ہے)۔ کسی ایک قسط کا لنک، یا <bdi dir="ltr">`?p=N`</bdi>، وہ صفحہ خود منتخب کر لیتا ہے۔ صفحات سے مراد bilibili ویڈیو کے حصے، سیزن، شو یا فہرست کی اقساط، اور X پوسٹ کی ویڈیوز ہیں۔

</details>

<details>
<summary><strong>فائل کے نام اور ٹیمپلیٹ متغیرات</strong></summary>

**فائل کے نام۔** ایک صفحے والا آئٹم: <bdi dir="ltr">`<title>`</bdi>؛ کئی صفحات والا (چاہے <bdi dir="ltr">`-p`</bdi> صرف ایک لے): <bdi dir="ltr">`<title>/[P<pageNumberWithZero>]<pageTitle>`</bdi>، کل صفحات کی تعداد کے مطابق آگے صفر لگا کر۔ متغیرات:

<bdi dir="ltr">`<title>`</bdi> <bdi dir="ltr">`<pageNumber>`</bdi> <bdi dir="ltr">`<pageNumberWithZero>`</bdi> <bdi dir="ltr">`<pageTitle>`</bdi> <bdi dir="ltr">`<id>`</bdi> <bdi dir="ltr">`<site>`</bdi> <bdi dir="ltr">`<uploader>`</bdi> <bdi dir="ltr">`<uploaderId>`</bdi> <bdi dir="ltr">`<quality>`</bdi> <bdi dir="ltr">`<resolution>`</bdi> <bdi dir="ltr">`<fps>`</bdi> <bdi dir="ltr">`<videoCodec>`</bdi> <bdi dir="ltr">`<videoBitrate>`</bdi> <bdi dir="ltr">`<audioCodec>`</bdi> <bdi dir="ltr">`<audioBitrate>`</bdi> <bdi dir="ltr">`<publishDate>`</bdi> <bdi dir="ltr">`<pageDate>`</bdi>، اور bilibili کے <bdi dir="ltr">`<bvid>`</bdi> <bdi dir="ltr">`<aid>`</bdi> <bdi dir="ltr">`<cid>`</bdi> <bdi dir="ltr">`<api>`</bdi>۔

تاریخوں کو فارمیٹ دیا جا سکتا ہے: <bdi dir="ltr">`<publishDate:yyyy-MM-dd>`</bdi>۔ توسیع خود شامل ہوتی ہے۔

</details>

<a id="site-notes"></a><a id="notes"></a>

## سائٹ سے متعلق معلومات

<details>
<summary><strong>YouTube</strong> · اخذ، سب ٹائٹلز اور پلے لسٹس</summary>

YouTube اپنے اسٹریم URL کو پلیئر کی ایسی جانچوں سے محفوظ رکھتا ہے جن کے لیے JavaScript رن ٹائم درکار ہے، اس لیے اخذ کرنے کا کام <bdi dir="ltr">`yt-dlp -J`</bdi> پر چھوڑا گیا ہے (جو انہیں deno میں چلاتا ہے)؛ اس کے بعد کا سارا کام haul خود کرتا ہے۔ اپ لوڈ کیے گئے سب ٹائٹلز فائل میں شامل کیے جاتے ہیں؛ خودکار طور پر بنے سب ٹائٹلز صرف <bdi dir="ltr">`--auto-subtitles`</bdi> کے ساتھ۔ googlevideo صرف محدود بائٹ رینج دیتا ہے، اس لیے ٹریک ایک وقت میں 10 MB کی ایک رینج کر کے لائے جاتے ہیں۔ <bdi dir="ltr">`watch?v=…&list=…`</bdi> صرف وہی ویڈیو ڈاؤن لوڈ کرتا ہے۔ yt-dlp کو تازہ رکھیں: YouTube کی تبدیلیوں کے حل وہیں آتے ہیں۔

</details>

<details>
<summary><strong>X</strong> · عوامی پوسٹس اور آڈیو</summary>

X کی عوامی پوسٹس کے لیے لاگ ان نہیں چاہیے۔ اس کی ویڈیوز آڈیو سمیت ایک ہی MP4 فائل ہوتی ہیں، اس لیے جدول میں صرف ویڈیو دکھائی دیتی ہے؛ <bdi dir="ltr">`--audio-only`</bdi> آڈیو الگ کرتا ہے۔

</details>

<details>
<summary><strong>bilibili</strong> · لاگ ان، بہتر معیار اور danmaku</summary>

bilibili کو اس کے اپنے web، TV، APP (gRPC) اور بین الاقوامی API کے ذریعے پڑھا جاتا ہے۔ لاگ آؤٹ حالت میں صرف کم معیار ملتا ہے (عام طور پر 480P تک)؛ 1080P، 4K، HDR، Dolby Vision اور Hi-Res آڈیو کے لیے لاگ ان کریں:

<div dir="ltr">

```sh
haul login bilibili                # bilibili ایپ سے QR کوڈ اسکین کریں
haul login bilibili --from-edge    # macOS: Microsoft Edge کا لاگ ان دوبارہ استعمال کریں (یا --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # TV ایکسیس ٹوکن، --api tv / --api app کے لیے
```

</div>

براؤزر کا لاگ ان پڑھنا فی الحال صرف macOS پر کام کرتا ہے: اس کے لیے ٹرمینل کو Full Disk Access چاہیے، اور macOS ایک بار “Safe Storage” کی چین آئٹم کی اجازت مانگتا ہے۔ <bdi dir="ltr">`--danmaku`</bdi> اسکرین پر چلتے تبصرے (danmaku) XML اور ASS میں محفوظ کرتا ہے؛ <bdi dir="ltr">`--danmaku-format ass`</bdi> ان میں سے صرف ایک رکھتا ہے۔

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · اقساط اور میٹاڈیٹا</summary>

Xiaoyuzhou اور Apple Podcasts کو yt-dlp نہیں چاہیے: Xiaoyuzhou کے صفحات میں آڈیو لنک ہوتا ہے، اور Apple کے لنکس عوامی iTunes API اور شو کی RSS فیڈ کے ذریعے جاتے ہیں۔ اقساط اپنا فارمیٹ (<bdi dir="ltr">`.mp3`</bdi> یا <bdi dir="ltr">`.m4a`</bdi>) برقرار رکھتی ہیں؛ سرورق شامل ہوتا ہے، شو البم اور میزبان فنکار بنتا ہے۔ شو کا لنک حالیہ اقساط پرانی سے نئی ترتیب میں دکھاتا ہے، اس لیے <bdi dir="ltr">`-p LAST`</bdi> سب سے نئی قسط ہے۔

</details>

<a id="config"></a>

## ترتیبات

<bdi dir="ltr">`~/.config/haul/`</bdi> (یا <bdi dir="ltr">`$HAUL_HOME`</bdi>) میں یہ فائلیں ہوتی ہیں:

| فائل | مقصد |
| --- | --- |
| <bdi dir="ltr">`config.json`</bdi> | کسی بھی اختیار کی طے شدہ قدر |
| <bdi dir="ltr">`cookie.txt`</bdi>, <bdi dir="ltr">`tv-token.txt`</bdi>, <bdi dir="ltr">`app-token.txt`</bdi> | bilibili لاگ ان |
| <bdi dir="ltr">`archives.txt`</bdi> | پہلے سے ڈاؤن لوڈ شدہ صفحات (<bdi dir="ltr">`--archive`</bdi>) |

<bdi dir="ltr">`config.json`</bdi> میں اختیارات کے نام camelCase میں ہوتے ہیں، bilibili والے <bdi dir="ltr">`"bilibili"`</bdi> کے اندر، اور اس میں صرف وہی لکھنا ہوتا ہے جو بدلنا ہو؛ فہرست array بھی ہو سکتی ہے اور کاما سے الگ کی گئی string بھی۔ کمانڈ لائن فلیگ اسے اوور رائیڈ کرتے ہیں:

<div dir="ltr">

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

</div>

<a id="development"></a>

## ڈیولپمنٹ

<div dir="ltr">

```sh
go build ./cmd/haul
go test ./...                                  # آف لائن، چند سیکنڈ
HAUL_LIVE=1 go test -run Live ./internal/...   # bilibili، YouTube، X اور Apple Podcasts سے بھی رابطہ کرتا ہے
```

</div>

آف لائن ٹیسٹ کبھی نیٹ ورک کو نہیں چھوتے: ٹیسٹ ایک فرضی <bdi dir="ltr">`http.RoundTripper`</bdi> سے بات کرتے ہیں، جو ریکارڈ شدہ API جوابات اور نقلی CDN (صرف رینج دینے والے سرور، ٹوٹتے کنکشن، رینج کو نظرانداز کرنے والے سرور) سے جواب دیتا ہے۔ مکمل سلسلے کے ٹیسٹ اصل ffmpeg سے ٹریک یکجا کرتے ہیں اور نتیجہ ffprobe سے جانچتے ہیں؛ ffmpeg انسٹال نہ ہو تو یہ چھوڑ دیے جاتے ہیں۔ ساخت اور اصولوں کے لیے [AGENTS.md](AGENTS.md) دیکھیں۔

<bdi dir="ltr">`v*`</bdi> ٹیگ پش کرنے پر GitHub Actions ٹیسٹ چلاتا ہے، ہر پلیٹ فارم کے لیے کراس کمپائل کرتا ہے اور ریلیز شائع کرتا ہے۔

<a id="acknowledgements"></a>

## اظہار تشکر

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr)، اور [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) اور [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api) کے API نوٹس۔

<a id="license"></a>

## لائسنس

[MIT](LICENSE)

---

<p align="center">
  <strong>ایک لنک۔ آپ کا میڈیا۔</strong><br>
  <a href="#install">haul حاصل کریں</a> · <a href="llms.txt">ایجنٹ حوالہ</a> · <a href="AGENTS.md">تعاون کریں</a>
</p>

</div>
