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
  <a href="README.ur.md"><bdi>اردو</bdi></a> ·
  <a href="README.ru.md"><bdi>Русский</bdi></a> ·
  <a href="README.de.md"><bdi>Deutsch</bdi></a> ·
  <a href="README.pcm.md"><bdi>Naijá</bdi></a> ·
  <bdi><strong>مصري</strong></bdi>
</p>
<!-- languages:end -->

<h1 align="center">لينك واحد. الميديا بتاعتك.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="حالة البناء"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="آخر إصدار"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 أو أحدث">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS وLinux وWindows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="رخصة MIT"></a>
</p>

<p align="center">
  <strong>أداة من سطر الأوامر لتنزيل الفيديو والصوت، معمولة للناس ولوكلاء الذكاء الاصطناعي زي بعض.</strong><br>
  اديها لينك، اختار المسارات، وخد الفيديو والصوت والترجمة والفصول والغلاف في ملف واحد.
</p>

<p align="center"><a href="#install">التثبيت</a> · <a href="#usage">ابدأ بسرعة</a> · <a href="#sites">المواقع المدعومة</a> · <a href="#for-ai-agents-and-scripts">دليل الوكلاء</a> · <a href="#options">الاختيارات</a> · <a href="../../releases">الإصدارات</a></p>

---

<a id="small-command-complete-download"></a>

## أمر صغير. تنزيل كامل.

| ↓ الميديا بتاعتك على مزاجك | ⌘ ملف تنفيذي واحد لكل الأجهزة | { } جاهز للأتمتة |
| :--- | :--- | :--- |
| اختار الجودة والترميز والصفحات وأسامي الملفات بنفس الكلمات على خمس مواقع | ملف Go تنفيذي واحد لـ macOS وLinux وWindows، بيحمّل بالنطاقات على التوازي وبيكمّل من مكان ما وقف | مستند JSON واحد في stdout، والسجلات في stderr، وأكواد خروج ليها معنى، ومفيش أسئلة من غير طرفية |

| 01 / شوف | 02 / اختار | 03 / نزّل | 04 / اجمع |
| :--- | :--- | :--- | :--- |
| **شوف كل المسارات** | **حدد أولوياتك** | **هاتها على جهازك** | **خلّي كله مع بعض** |
| الصفحات والترميز والأحجام | الجودة والصوت والصفحات | بالنطاقات أو بالأجزاء | المسارات والترجمة والغلاف |
| <bdi dir="ltr">`haul info "<url>"`</bdi> | <bdi dir="ltr">`-q 720p -c avc,m4a`</bdi> | <bdi dir="ltr">`haul "<url>"`</bdi> | <bdi dir="ltr">`ffmpeg / MP4Box`</bdi> |

<details>
<summary><strong>بص جوه الطرفية</strong></summary>

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

> للاستخدام الشخصي والبحثي وأي استخدام تاني مش تجاري. إنت المسؤول عن احترام حقوق النشر وشروط كل موقع.

<a id="install"></a>

## التثبيت

haul بيشتغل على **macOS وLinux وWindows** (amd64 وarm64). محتاج [ffmpeg](https://ffmpeg.org) لدمج المسارات؛ وYouTube وX محتاجين كمان [yt-dlp](https://github.com/yt-dlp/yt-dlp)، واختبارات مشغّل YouTube محتاجة [deno](https://deno.com):

<div dir="ltr">

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip curl  # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows: باكدج واحدة كل مرة
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

</div>

على Linux، نزّل yt-dlp بـ pipx أو من ملف <bdi dir="ltr">`yt-dlp_linux`</bdi> في [الإصدارات بتاعته](https://github.com/yt-dlp/yt-dlp/releases/latest) (<bdi dir="ltr">`yt-dlp_linux_aarch64`</bdi> على arm64)، مش من باكدجات التوزيعة: YouTube بيتغيّر كتير والباكدجات بتبقى قديمة. بعد التثبيت افتح ترمينال جديد عشان الأدوات تبقى في <bdi dir="ltr">`PATH`</bdi>.

<a id="get-the-binary"></a><a id="binary"></a>

### نزّل الملف التنفيذي

نزّل الأرشيف اللي على قد نظامك من [الإصدارات](../../releases)، يعني <bdi dir="ltr">`haul-<version>-<os>-<arch>.tar.gz`</bdi> (أو <bdi dir="ltr">`.zip`</bdi> على Windows)، وحط <bdi dir="ltr">`haul`</bdi> في أي فولدر جوه <bdi dir="ltr">`PATH`</bdi>:

<div dir="ltr">

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

</div>

مسار الإصدار الحالي بيوقّع نسخ macOS وبيوثّقها عند Apple. اختار <bdi dir="ltr">`haul-<version>-darwin-<arch>.pkg`</bdi> عشان يثبت <bdi dir="ltr">`haul`</bdi> في <bdi dir="ltr">`/usr/local/bin`</bdi> ومعاه تذكرة التوثيق. الأرشيف فيه نفس البرنامج الموقّع، بس أول تحقق ممكن يحتاج إنترنت. الإصدارات القديمة ممكن تكون من غير توقيع.

<details>
<summary><strong>ابني من الكود المصدري</strong> · Go 1.26+</summary>

<div dir="ltr">

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

</div>

أو من نسخة متنسخة من الريبو:

<div dir="ltr">

```sh
go build -o haul ./cmd/haul
```

</div>

</details>

<a id="usage"></a>

## ابدأ بسرعة

ابدأ بلينك. haul بيختار أحسن مسارات متاحة لوحده.

<div dir="ltr">

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

</div>

شوف المعلومات قبل ما تنزّل، أو اختار الجودة والترميز:

<div dir="ltr">

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC عشان QuickTime
```

</div>

<details>
<summary><strong>دليل الأوامر</strong></summary>

| الأمر | معناه |
| --- | --- |
| <bdi dir="ltr">`haul <url> [options]`</bdi> | تنزيل؛ زي <bdi dir="ltr">`haul download <url>`</bdi> بالظبط |
| <bdi dir="ltr">`haul info <url> [--urls]`</bdi> | يعرض العنصر وصفحاته ومساراته؛ من غير ما ينزّل حاجة |
| <bdi dir="ltr">`haul login bilibili`</bdi> | تسجيل دخول bilibili عشان جودة أعلى |
| <bdi dir="ltr">`haul templates`</bdi> | متغيرات قوالب أسامي الملفات |
| <bdi dir="ltr">`haul --help`</bdi> | المواقع واستخدام الوكلاء وأكواد الخروج والأمثلة |
| <bdi dir="ltr">`haul download --help`</bdi> | كل الاختيارات |

</details>

<a id="sites"></a>

### المواقع المدعومة

| الموقع | اللينكات | محتاج |
| --- | --- | --- |
| **YouTube** | <bdi dir="ltr">`youtube.com/watch?v=…`</bdi>, <bdi dir="ltr">`youtu.be/…`</bdi>, <bdi dir="ltr">`/shorts/…`</bdi>, <bdi dir="ltr">`/embed/…`</bdi>, <bdi dir="ltr">`/live/…`</bdi> | yt-dlp, deno |
| **X** | <bdi dir="ltr">`x.com/<user>/status/<id>`</bdi>, <bdi dir="ltr">`twitter.com/…`</bdi>؛ <bdi dir="ltr">`/video/<n>`</bdi> بيختار فيديو واحد من البوست | yt-dlp |
| **bilibili** | فيديوهات ومسلسلات (bangumi) وكورسات ومجموعات وسلاسل ومفضلة وصفحات مستخدمين و<bdi dir="ltr">`b23.tv`</bdi> ومعرّفات لوحدها <bdi dir="ltr">`BV…`</bdi> <bdi dir="ltr">`av…`</bdi> <bdi dir="ltr">`ep…`</bdi> <bdi dir="ltr">`ss…`</bdi> <bdi dir="ltr">`md…`</bdi> | – |
| **Xiaoyuzhou** | <bdi dir="ltr">`xiaoyuzhoufm.com/episode/<id>`</bdi>, <bdi dir="ltr">`/podcast/<id>`</bdi> | – |
| **Apple Podcasts** | <bdi dir="ltr">`podcasts.apple.com/<cc>/podcast/<name>/id<show>`</bdi>، ومعاه <bdi dir="ltr">`?i=<episode>`</bdi> لحلقة واحدة | – |

<a id="make-it-yours"></a><a id="examples"></a>

### على مزاجك

**الصوت بس**

<div dir="ltr">

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC بدل Opus
```

</div>

**كام جزء، أو موسم كامل، أو آخر حلقة**

<div dir="ltr">

```sh
# أجزاء مختارة، تتحفظ في فولدر Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# كل حلقات الموسم
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# أحدث حلقة بودكاست
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# كل الفيديوهات اللي في بوست X
haul -p ALL "https://x.com/<user>/status/<id>"
```

</div>

**ملفات مترتبة، أو ترجمة، أو اختيار بإيدك**

<div dir="ltr">

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # لغات الترجمة دي بس
haul -i "BV1qt4y1X7TW"                                 # اختار المسارات بأسهم الكيبورد
```

</div>

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## لوكلاء الذكاء الاصطناعي والسكربتات

<bdi dir="ltr">`haul --help`</bdi> فيه كل اللي الوكيل محتاجه، و[llms.txt](llms.txt) هو نفس الدليل في ملف. من الآخر:

<div dir="ltr">

```sh
haul info --json "<url>"                                # 1. شوف
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. نزّل المسارات دي بالظبط
haul --json -q 720p -c avc,m4a "<url>"                  #    أو سيب الأولويات تختار
```

</div>

- مع <bdi dir="ltr">`--json`</bdi>، **stdout فيه مستند JSON واحد بس**؛ التقدم والسجلات بيروحوا stderr. <bdi dir="ltr">`files`</bdi> بتعرض ملفات النتيجة، الجديدة واللي كانت موجودة.
- **مفيش حاجة تفاعلية من غير طرفية.** <bdi dir="ltr">`-i`</bdi> من غير طرفية بيفشل بكود خروج 2 وبيقولك تستخدم أنهي اختيارات بداله.
- أرقام المسارات في <bdi dir="ltr">`info`</bdi> ماشية بنفس ترتيب اختيار haul؛ استخدم نفس <bdi dir="ltr">`-q`</bdi> / <bdi dir="ltr">`-c`</bdi> مع <bdi dir="ltr">`info`</bdi> ومع التنزيل.
- <bdi dir="ltr">`info`</bdi> على قائمة تشغيل أو موسم أو برنامج أو بوست فيه كذا فيديو بيعرض الصفحات؛ <bdi dir="ltr">`-p <n>`</bdi> بيضيف مسارات وترجمات الصفحة دي.
- الصفحات اللي موجودة على الديسك بتظهر كـ <bdi dir="ltr">`"status": "skipped", "reason": "exists"`</bdi>، فتقدر تشغّل الأمر تاني من غير قلق.

<details>
<summary><strong>مثال لرد JSON</strong> · تنزيل نجح</summary>

التنزيل بيطبع:

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

ده ناتج <bdi dir="ltr">`haul --json -q 360p -c avc,m4a …`</bdi>. هنا المفاتيح مترتبة بترتيب القراية، وقوائم المسارات متقصّرة على المختار بس؛ أما المستند الحقيقي فبيرتب مفاتيحه أبجديًا وبيعرض كل المسارات.

</details>

لو حصل فشل، بيطبع <bdi dir="ltr">`"ok": false`</bdi> ومعاه <bdi dir="ltr">`"error": {"kind", "message", "exitCode"}`</bdi>، وبيفضل محتفظ بالصفحات اللي خلصت قبله.

| كود الخروج | معناه | <bdi dir="ltr">`error.kind`</bdi> |
| --- | --- | --- |
| 0 | خلص | – |
| 1 | التنزيل أو الاستخراج فشل | <bdi dir="ltr">`failed`</bdi> |
| 2 | لينك مش مدعوم أو قيمة اختيار غلط | <bdi dir="ltr">`input`</bdi> |
| 3 | أداة مطلوبة مش موجودة (ffmpeg أو yt-dlp) | <bdi dir="ltr">`dependency`</bdi> |
| 4 | محتاج تسجيل دخول أو الجلسة انتهت | <bdi dir="ltr">`auth`</bdi> |
| 64 | سطر أوامر غلط (اختيار مش معروف أو وسيط ناقص) | – |
| 130 | اتلغى | <bdi dir="ltr">`cancelled`</bdi> |

<a id="options"></a>

## الاختيارات

<bdi dir="ltr">`haul download --help`</bdi> بيعرضهم كلهم. أهمهم:

| الفئة | الاختيارات |
| --- | --- |
| عام | <bdi dir="ltr">`--json`</bdi>, <bdi dir="ltr">`--config <file>`</bdi>, <bdi dir="ltr">`--debug`</bdi> |
| المسارات | <bdi dir="ltr">`-q, --quality <list>`</bdi>, <bdi dir="ltr">`-c, --codec <list>`</bdi>, <bdi dir="ltr">`--video-stream <n>`</bdi>, <bdi dir="ltr">`--audio-stream <n>`</bdi>, <bdi dir="ltr">`-i, --interactive`</bdi>, <bdi dir="ltr">`--video-ascending`</bdi>, <bdi dir="ltr">`--audio-ascending`</bdi> |
| الصفحات | <bdi dir="ltr">`-p, --pages <spec>`</bdi>, <bdi dir="ltr">`--show-all`</bdi>, <bdi dir="ltr">`--hide-streams`</bdi> |
| المحتوى | <bdi dir="ltr">`--audio-only`</bdi>, <bdi dir="ltr">`--video-only`</bdi>, <bdi dir="ltr">`--subtitle-only`</bdi>, <bdi dir="ltr">`--cover-only`</bdi>, <bdi dir="ltr">`--skip-subtitle`</bdi>, <bdi dir="ltr">`--sub-lang <list>`</bdi>, <bdi dir="ltr">`--auto-subtitles`</bdi>, <bdi dir="ltr">`--skip-cover`</bdi>, <bdi dir="ltr">`--skip-mux`</bdi> |
| الناتج | <bdi dir="ltr">`-o, --output <template>`</bdi>, <bdi dir="ltr">`--multi-output <template>`</bdi>, <bdi dir="ltr">`-w, --work-dir <dir>`</bdi>, <bdi dir="ltr">`--lang <code>`</bdi>, <bdi dir="ltr">`--no-tags`</bdi>, <bdi dir="ltr">`--archive`</bdi>, <bdi dir="ltr">`--delay <seconds>`</bdi> |
| bilibili | <bdi dir="ltr">`--api web\|tv\|app\|intl`</bdi>, <bdi dir="ltr">`--danmaku`</bdi>, <bdi dir="ltr">`--danmaku-only`</bdi>, <bdi dir="ltr">`--danmaku-format xml,ass`</bdi>, <bdi dir="ltr">`--cookie`</bdi>, <bdi dir="ltr">`--token`</bdi>؛ والباقي في <bdi dir="ltr">`--help-hidden`</bdi> |
| الأدوات | <bdi dir="ltr">`--ffmpeg`</bdi>, <bdi dir="ltr">`--yt-dlp`</bdi>, <bdi dir="ltr">`--use-mp4box`</bdi>, <bdi dir="ltr">`--mp4box`</bdi>, <bdi dir="ltr">`--use-aria2c`</bdi>, <bdi dir="ltr">`--aria2c`</bdi>, <bdi dir="ltr">`--aria2c-args`</bdi>, <bdi dir="ltr">`--single-connection`</bdi> |

<details>
<summary><strong>أولوية الجودة والترميز</strong></summary>

**الجودة والترميز.** <bdi dir="ltr">`-q`</bdi> بياخد نفس الأسامي اللي في جدول المسارات: <bdi dir="ltr">`1080p`</bdi> و<bdi dir="ltr">`720p60`</bdi> على YouTube؛ و<bdi dir="ltr">`8K`</bdi> و<bdi dir="ltr">`Dolby Vision`</bdi> و<bdi dir="ltr">`HDR`</bdi> و<bdi dir="ltr">`4K`</bdi> و<bdi dir="ltr">`1080P60`</bdi> و<bdi dir="ltr">`1080P+`</bdi> و<bdi dir="ltr">`1080P`</bdi> و<bdi dir="ltr">`720P`</bdi> على bilibili. و<bdi dir="ltr">`-c`</bdi> بياخد <bdi dir="ltr">`av1 vp9 hevc avc`</bdi> للفيديو و<bdi dir="ltr">`m4a opus flac eac3 mp3`</bdi> للصوت. دي أولويات مش فلاتر: اللي مش مكتوب بييجي بعدهم، الأحسن الأول. <bdi dir="ltr">`--video-ascending`</bdi> / <bdi dir="ltr">`--audio-ascending`</bdi> بيعكسوا الترتيب عشان أصغر تنزيل.

</details>

<details>
<summary><strong>اختيار الصفحات</strong></summary>

**الصفحات.** <bdi dir="ltr">`8`</bdi>، <bdi dir="ltr">`1,2`</bdi>، <bdi dir="ltr">`3-5`</bdi>، <bdi dir="ltr">`1-3,10`</bdi>، <bdi dir="ltr">`ALL`</bdi>، <bdi dir="ltr">`LAST`</bdi> (آخر صفحة، يعني أحدث حلقة في البرنامج؛ و<bdi dir="ltr">`LATEST`</bdi> بيشتغل برضه). لينك حلقة واحدة، أو <bdi dir="ltr">`?p=N`</bdi>، بيختار الصفحة دي لوحدها. الصفحات هي أجزاء فيديو bilibili، وحلقات الموسم أو البرنامج أو القائمة، وفيديوهات بوست X.

</details>

<details>
<summary><strong>أسامي الملفات ومتغيرات القوالب</strong></summary>

**أسامي الملفات.** عنصر بصفحة واحدة: <bdi dir="ltr">`<title>`</bdi>؛ وبكذا صفحة (حتى لو <bdi dir="ltr">`-p`</bdi> اختار واحدة بس): <bdi dir="ltr">`<title>/[P<pageNumberWithZero>]<pageTitle>`</bdi>، والرقم بيتكمّل بأصفار على قد عدد الصفحات. المتغيرات:

<bdi dir="ltr">`<title>`</bdi> <bdi dir="ltr">`<pageNumber>`</bdi> <bdi dir="ltr">`<pageNumberWithZero>`</bdi> <bdi dir="ltr">`<pageTitle>`</bdi> <bdi dir="ltr">`<id>`</bdi> <bdi dir="ltr">`<site>`</bdi> <bdi dir="ltr">`<uploader>`</bdi> <bdi dir="ltr">`<uploaderId>`</bdi> <bdi dir="ltr">`<quality>`</bdi> <bdi dir="ltr">`<resolution>`</bdi> <bdi dir="ltr">`<fps>`</bdi> <bdi dir="ltr">`<videoCodec>`</bdi> <bdi dir="ltr">`<videoBitrate>`</bdi> <bdi dir="ltr">`<audioCodec>`</bdi> <bdi dir="ltr">`<audioBitrate>`</bdi> <bdi dir="ltr">`<publishDate>`</bdi> <bdi dir="ltr">`<pageDate>`</bdi>، وكمان بتوع bilibili: <bdi dir="ltr">`<bvid>`</bdi> <bdi dir="ltr">`<aid>`</bdi> <bdi dir="ltr">`<cid>`</bdi> <bdi dir="ltr">`<api>`</bdi>

التواريخ بتقبل صيغة: <bdi dir="ltr">`<publishDate:yyyy-MM-dd>`</bdi>. الامتداد بيتضاف لوحده.

</details>

<a id="site-notes"></a><a id="notes"></a>

## ملاحظات عن المواقع

<details>
<summary><strong>YouTube</strong> · الاستخراج والترجمة وقوائم التشغيل</summary>

YouTube بيحمي لينكات المسارات باختبارات مشغّل محتاجة بيئة JavaScript، فالاستخراج متساب لـ <bdi dir="ltr">`yt-dlp -J`</bdi> (اللي بيشغّلها في deno)؛ وكل اللي بعد كده haul بيعمله بنفسه. الترجمة المرفوعة بتندمج، والتلقائية ما بتندمجش غير مع <bdi dir="ltr">`--auto-subtitles`</bdi>. googlevideo بيدّي نطاقات بايت محدودة بس، فالمسارات بتتجاب نطاق 10 MB كل مرة. <bdi dir="ltr">`watch?v=…&list=…`</bdi> بينزّل الفيديو ده بس. خلّي yt-dlp متحدّث دايمًا: هو ده المكان اللي بتوصل فيه إصلاحات تغييرات YouTube.

</details>

<details>
<summary><strong>X</strong> · البوستات العامة والصوت</summary>

بوستات X العامة مش محتاجة تسجيل دخول. الفيديوهات ملفات MP4 لوحدها والصوت جواها، فالجدول بيعرض الفيديو بس؛ و<bdi dir="ltr">`--audio-only`</bdi> بيطلّع الصوت.

</details>

<details>
<summary><strong>bilibili</strong> · تسجيل الدخول والجودة الأعلى والتعليقات الطايرة (danmaku)</summary>

haul بيقرا bilibili من واجهاته هو: web وTV وAPP ‏(gRPC) والدولية. من غير دخول، بيدّي جودة أقل بس (غالبًا لحد 480P)؛ سجّل دخول عشان 1080P و4K وHDR وDolby Vision وصوت Hi-Res:

<div dir="ltr">

```sh
haul login bilibili                # امسح كود QR بتطبيق bilibili
haul login bilibili --from-edge    # macOS: استخدم دخول Microsoft Edge (أو --from-chrome؛ --profile "Profile 1")
haul login bilibili --tv           # رمز دخول TV، لـ --api tv / --api app
```

</div>

قراءة دخول المتصفح بتشتغل على macOS بس دلوقتي: الطرفية محتاجة وصول كامل للديسك (Full Disk Access)، وmacOS بيطلب مرة واحدة الوصول لعنصر “Safe Storage” في سلسلة المفاتيح. <bdi dir="ltr">`--danmaku`</bdi> بيحفظ التعليقات الطايرة بصيغ XML وASS؛ و<bdi dir="ltr">`--danmaku-format ass`</bdi> بيسيب واحدة بس منهم.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · الحلقات وبياناتها</summary>

Xiaoyuzhou وApple Podcasts مش محتاجين yt-dlp: صفحات Xiaoyuzhou فيها لينك الصوت، ولينكات Apple بتعدّي على واجهة iTunes العامة وRSS البرنامج. الحلقات بتفضل بصيغتها (<bdi dir="ltr">`.mp3`</bdi> أو <bdi dir="ltr">`.m4a`</bdi>) والغلاف جواها، واسم البرنامج كألبوم والمقدّم كفنان. لينك البرنامج بيعرض آخر حلقاته من الأقدم للأحدث، فـ <bdi dir="ltr">`-p LAST`</bdi> هي الأحدث.

</details>

<a id="config"></a>

## الإعدادات

<bdi dir="ltr">`~/.config/haul/`</bdi> (أو <bdi dir="ltr">`$HAUL_HOME`</bdi>) فيه:

| الملف | استخدامه |
| --- | --- |
| <bdi dir="ltr">`config.json`</bdi> | القيم الافتراضية لأي اختيار |
| <bdi dir="ltr">`cookie.txt`</bdi>, <bdi dir="ltr">`tv-token.txt`</bdi>, <bdi dir="ltr">`app-token.txt`</bdi> | تسجيل دخول bilibili |
| <bdi dir="ltr">`archives.txt`</bdi> | الصفحات اللي اتنزّلت قبل كده (<bdi dir="ltr">`--archive`</bdi>) |

<bdi dir="ltr">`config.json`</bdi> بيستخدم أسامي الاختيارات بصيغة camelCase، واختيارات bilibili تحت <bdi dir="ltr">`"bilibili"`</bdi>، ومحتاج بس اللي هتغيّره؛ والقائمة ممكن تبقى مصفوفة أو نص مفصول بفواصل. واختيارات سطر الأوامر بتغلب عليه:

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

## التطوير

<div dir="ltr">

```sh
go build ./cmd/haul
go test ./...                                  # من غير إنترنت، كام ثانية
HAUL_LIVE=1 go test -run Live ./internal/...   # بيكلّم كمان bilibili وYouTube وX وApple Podcasts
```

</div>

الاختبارات اللي من غير إنترنت عمرها ما بتلمس الشبكة: بتكلّم بديل <bdi dir="ltr">`http.RoundTripper`</bdi> بيرد من ردود API متسجلة ومن CDN متقلّدة (سيرفرات بتقبل النطاقات بس، واتصالات بتقطع، وسيرفرات بتتجاهل النطاقات). الاختبارات الشاملة بتدمج بـ ffmpeg حقيقي وبتتأكد من النتيجة بـ ffprobe، وبتتخطى لو ffmpeg مش متثبت. الهيكل والقواعد في [AGENTS.md](AGENTS.md).

رفع وسم <bdi dir="ltr">`v*`</bdi> بيختبر ويبني لكل المنصات وينشر إصدار عن طريق GitHub Actions.

<a id="acknowledgements"></a>

## شكر وتقدير

[yt-dlp](https://github.com/yt-dlp/yt-dlp)، [FFmpeg](https://ffmpeg.org)، [GPAC](https://gpac.io)، [aria2](https://aria2.github.io)، [pflag](https://github.com/spf13/pflag)، [x/term](https://pkg.go.dev/golang.org/x/term)، [rsc.io/qr](https://pkg.go.dev/rsc.io/qr)، وملاحظات واجهات bilibili في [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) و[bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## الرخصة

[MIT](LICENSE)

---

<p align="center">
  <strong>لينك واحد. الميديا بتاعتك.</strong><br>
  <a href="#install">هات haul</a> · <a href="llms.txt">مرجع الوكلاء</a> · <a href="AGENTS.md">شارك</a>
</p>

</div>
