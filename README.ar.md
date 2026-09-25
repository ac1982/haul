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
  <bdi><strong>العربية</strong></bdi> ·
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

<h1 align="center">رابط واحد. وسائطك بين يديك.</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="حالة البناء"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="أحدث إصدار"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 أو أحدث">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS وLinux وWindows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="رخصة MIT"></a>
</p>

<p align="center">
  <strong>أداة سطر أوامر لتنزيل الفيديو والصوت، مصممة للأشخاص ولوكلاء الذكاء الاصطناعي على حد سواء.</strong><br>
  أعطها رابطًا، واختر التدفقات، واحصل على الفيديو والصوت والترجمات والفصول والغلاف في ملف واحد.
</p>

<p align="center"><a href="#install">التثبيت</a> · <a href="#usage">البدء السريع</a> · <a href="#sites">المواقع المدعومة</a> · <a href="#for-ai-agents-and-scripts">دليل الوكلاء</a> · <a href="#options">الخيارات</a> · <a href="../../releases">الإصدارات</a></p>

---

<a id="small-command-complete-download"></a>

## أمر صغير. تنزيل كامل.

| ↓ وسائطك كما تريدها | ⌘ ملف تنفيذي واحد لكل أنظمة سطح المكتب | { } جاهز للأتمتة |
| :--- | :--- | :--- |
| اختر الجودة والترميز والصفحات وأسماء الملفات بالمفردات نفسها عبر خمسة مواقع | ملف Go تنفيذي واحد لأنظمة macOS وLinux وWindows، مع تنزيل متوازٍ بالنطاقات يمكن استئنافه | مستند JSON واحد على stdout، والسجلات على stderr، ورموز خروج ذات معنى، ولا أسئلة تفاعلية دون طرفية |

| 01 / الاستعراض | 02 / الاختيار | 03 / التنزيل | 04 / الدمج |
| :--- | :--- | :--- | :--- |
| **كل تدفق أمام عينيك** | **حدّد أولوياتك** | **أحضرها إلى جهازك** | **اجمع كل شيء معًا** |
| الصفحات والترميزات والأحجام | الجودة والصوت والصفحات | بالنطاقات أو بالمقاطع | المسارات والترجمات والغلاف |
| <bdi dir="ltr">`haul info "<url>"`</bdi> | <bdi dir="ltr">`-q 720p -c avc,m4a`</bdi> | <bdi dir="ltr">`haul "<url>"`</bdi> | <bdi dir="ltr">`ffmpeg / MP4Box`</bdi> |

<details>
<summary><strong>نظرة داخل الطرفية</strong></summary>

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

> للاستخدام الشخصي والبحثي وغيره من الاستخدامات غير التجارية. أنت مسؤول عن احترام حقوق النشر وشروط كل موقع.

<a id="install"></a>

## التثبيت

يعمل haul على **macOS وLinux وWindows** (amd64 وarm64). يحتاج إلى [ffmpeg](https://ffmpeg.org) لدمج المسارات؛ ويحتاج YouTube وX أيضًا إلى [yt-dlp](https://github.com/yt-dlp/yt-dlp)، وتحتاج اختبارات مشغّل YouTube إلى [deno](https://deno.com):

<div dir="ltr">

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg pipx unzip curl  # Linux (Debian / Ubuntu)
pipx install "yt-dlp[default]" && pipx ensurepath
curl -fsSL https://deno.land/install.sh | sh -s -- -y
winget install Gyan.FFmpeg               # Windows: حزمة واحدة في كل مرة
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

</div>

على Linux، ثبّت yt-dlp عبر pipx أو من الملف <bdi dir="ltr">`yt-dlp_linux`</bdi> في [إصداراته](https://github.com/yt-dlp/yt-dlp/releases/latest) (<bdi dir="ltr">`yt-dlp_linux_aarch64`</bdi> على arm64)، لا من حزم توزيعتك: يتغيّر YouTube كثيرًا وتتأخر الحزم عنه. بعد التثبيت افتح طرفية جديدة لتصبح الأدوات ضمن <bdi dir="ltr">`PATH`</bdi>.

<a id="get-the-binary"></a><a id="binary"></a>

### الحصول على الملف التنفيذي

نزّل أرشيف نظامك من [الإصدارات](../../releases)، أي <bdi dir="ltr">`haul-<version>-<os>-<arch>.tar.gz`</bdi> (أو <bdi dir="ltr">`.zip`</bdi> على Windows)، ثم ضع <bdi dir="ltr">`haul`</bdi> في أي مجلد ضمن <bdi dir="ltr">`PATH`</bdi>:

<div dir="ltr">

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

</div>

يوقّع مسار الإصدار الحالي نسخ macOS ويحصل على توثيق Apple لها. اختر <bdi dir="ltr">`haul-<version>-darwin-<arch>.pkg`</bdi> لتثبيت <bdi dir="ltr">`haul`</bdi> في <bdi dir="ltr">`/usr/local/bin`</bdi> مع تذكرة التوثيق المرفقة. يحتوي الأرشيف الملف التنفيذي الموقّع نفسه، لكن التحقق الأول قد يحتاج إلى الإنترنت. قد تكون الإصدارات القديمة غير موقّعة.

<details>
<summary><strong>البناء من المصدر</strong> · Go 1.26+</summary>

<div dir="ltr">

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

</div>

أو من نسخة مستنسخة من المستودع:

<div dir="ltr">

```sh
go build -o haul ./cmd/haul
```

</div>

</details>

<a id="usage"></a>

## البدء السريع

ابدأ برابط. يختار haul أفضل التدفقات المتاحة افتراضيًا.

<div dir="ltr">

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

</div>

استعرض المعلومات قبل التنزيل، أو اختر الجودة والترميز:

<div dir="ltr">

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # H.264 + AAC للتشغيل في QuickTime
```

</div>

<details>
<summary><strong>مرجع الأوامر</strong></summary>

| الأمر | المعنى |
| --- | --- |
| <bdi dir="ltr">`haul <url> [options]`</bdi> | التنزيل؛ يعادل <bdi dir="ltr">`haul download <url>`</bdi> |
| <bdi dir="ltr">`haul info <url> [--urls]`</bdi> | عرض العنصر وصفحاته وتدفقاته دون تنزيل أي شيء |
| <bdi dir="ltr">`haul login bilibili`</bdi> | تسجيل الدخول إلى bilibili للحصول على جودة أعلى |
| <bdi dir="ltr">`haul templates`</bdi> | متغيرات قوالب أسماء الملفات |
| <bdi dir="ltr">`haul --help`</bdi> | المواقع واستخدام الوكلاء ورموز الخروج والأمثلة |
| <bdi dir="ltr">`haul download --help`</bdi> | كل الخيارات |

</details>

<a id="sites"></a>

### المواقع المدعومة

| الموقع | الروابط | المتطلبات |
| --- | --- | --- |
| **YouTube** | <bdi dir="ltr">`youtube.com/watch?v=…`</bdi>, <bdi dir="ltr">`youtu.be/…`</bdi>, <bdi dir="ltr">`/shorts/…`</bdi>, <bdi dir="ltr">`/embed/…`</bdi>, <bdi dir="ltr">`/live/…`</bdi> | yt-dlp, deno |
| **X** | <bdi dir="ltr">`x.com/<user>/status/<id>`</bdi>, <bdi dir="ltr">`twitter.com/…`</bdi>؛ يختار <bdi dir="ltr">`/video/<n>`</bdi> فيديو واحدًا من المنشور | yt-dlp |
| **bilibili** | الفيديوهات والمسلسلات (bangumi) والدورات والمجموعات والسلاسل والمفضلة وصفحات المستخدمين و<bdi dir="ltr">`b23.tv`</bdi> والمعرّفات المجردة <bdi dir="ltr">`BV…`</bdi> <bdi dir="ltr">`av…`</bdi> <bdi dir="ltr">`ep…`</bdi> <bdi dir="ltr">`ss…`</bdi> <bdi dir="ltr">`md…`</bdi> | – |
| **Xiaoyuzhou** | <bdi dir="ltr">`xiaoyuzhoufm.com/episode/<id>`</bdi>, <bdi dir="ltr">`/podcast/<id>`</bdi> | – |
| **Apple Podcasts** | <bdi dir="ltr">`podcasts.apple.com/<cc>/podcast/<name>/id<show>`</bdi>، مع <bdi dir="ltr">`?i=<episode>`</bdi> لحلقة واحدة | – |

<a id="make-it-yours"></a><a id="examples"></a>

### على مقاسك

**الصوت فقط**

<div dir="ltr">

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # AAC بدلًا من Opus
```

</div>

**بعض الأجزاء، أو موسم كامل، أو أحدث حلقة**

<div dir="ltr">

```sh
# أجزاء مختارة، تُحفظ في مجلد Movies
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# كل حلقات الموسم
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# أحدث حلقة بودكاست
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# كل الفيديوهات في منشور X
haul -p ALL "https://x.com/<user>/status/<id>"
```

</div>

**ملفات منظمة، أو ترجمات، أو اختيار تفاعلي**

<div dir="ltr">

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # لغات الترجمة هذه فقط
haul -i "BV1qt4y1X7TW"                                 # اختر التدفقات بمفاتيح الأسهم
```

</div>

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## لوكلاء الذكاء الاصطناعي والبرامج النصية

يتضمن <bdi dir="ltr">`haul --help`</bdi> كل ما يحتاج إليه الوكيل، و[llms.txt](llms.txt) هو الدليل نفسه في ملف. باختصار:

<div dir="ltr">

```sh
haul info --json "<url>"                                # 1. الاستعراض
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. تنزيل هذه التدفقات تحديدًا
haul --json -q 720p -c avc,m4a "<url>"                  #    أو اترك الأولويات تختار
```

</div>

- مع <bdi dir="ltr">`--json`</bdi>، **لا يحمل stdout إلا مستند JSON واحدًا**؛ ويذهب التقدم والسجلات إلى stderr. يسرد الحقل <bdi dir="ltr">`files`</bdi> ملفات الإخراج، الجديدة منها والموجودة مسبقًا.
- **لا شيء تفاعلي دون طرفية.** يفشل <bdi dir="ltr">`-i`</bdi> حينها برمز الخروج 2 ويذكر الخيارات التي تُستخدم بدلًا منه.
- تتبع فهارس التدفقات في <bdi dir="ltr">`info`</bdi> الترتيب الذي يختار به haul؛ مرّر <bdi dir="ltr">`-q`</bdi> / <bdi dir="ltr">`-c`</bdi> نفسيهما إلى <bdi dir="ltr">`info`</bdi> وإلى التنزيل.
- يسرد <bdi dir="ltr">`info`</bdi> صفحات قائمة التشغيل أو الموسم أو البرنامج أو المنشور متعدد الفيديوهات؛ ويضيف <bdi dir="ltr">`-p <n>`</bdi> تدفقات تلك الصفحة وترجماتها.
- تُبلَّغ الصفحات الموجودة على القرص بالحالة <bdi dir="ltr">`"status": "skipped", "reason": "exists"`</bdi>، لذا فإعادة التشغيل آمنة.

<details>
<summary><strong>مثال لاستجابة JSON</strong> · تنزيل ناجح</summary>

يطبع التنزيل:

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

هذا ناتج <bdi dir="ltr">`haul --json -q 360p -c avc,m4a …`</bdi>. رُتّبت المفاتيح هنا بترتيب القراءة، واختُصرت قوائم التدفقات إلى المختار منها؛ أما المستند الفعلي فيرتّب مفاتيحه أبجديًا ويسرد كل التدفقات.

</details>

عند الفشل يُطبع <bdi dir="ltr">`"ok": false`</bdi> مع <bdi dir="ltr">`"error": {"kind", "message", "exitCode"}`</bdi>، مع الاحتفاظ بالصفحات التي اكتملت قبله.

| رمز الخروج | المعنى | <bdi dir="ltr">`error.kind`</bdi> |
| --- | --- | --- |
| 0 | اكتمل | – |
| 1 | فشل التنزيل أو الاستخراج | <bdi dir="ltr">`failed`</bdi> |
| 2 | رابط غير مدعوم أو قيمة خيار غير صالحة | <bdi dir="ltr">`input`</bdi> |
| 3 | أداة مطلوبة مفقودة (ffmpeg أو yt-dlp) | <bdi dir="ltr">`dependency`</bdi> |
| 4 | يلزم تسجيل الدخول أو انتهت صلاحيته | <bdi dir="ltr">`auth`</bdi> |
| 64 | سطر أوامر غير صحيح (خيار مجهول أو وسيطة مفقودة) | – |
| 130 | أُلغي | <bdi dir="ltr">`cancelled`</bdi> |

<a id="options"></a>

## الخيارات

يسردها <bdi dir="ltr">`haul download --help`</bdi> كلها. أهمها:

| الفئة | الخيارات |
| --- | --- |
| عام | <bdi dir="ltr">`--json`</bdi>, <bdi dir="ltr">`--config <file>`</bdi>, <bdi dir="ltr">`--debug`</bdi> |
| التدفقات | <bdi dir="ltr">`-q, --quality <list>`</bdi>, <bdi dir="ltr">`-c, --codec <list>`</bdi>, <bdi dir="ltr">`--video-stream <n>`</bdi>, <bdi dir="ltr">`--audio-stream <n>`</bdi>, <bdi dir="ltr">`-i, --interactive`</bdi>, <bdi dir="ltr">`--video-ascending`</bdi>, <bdi dir="ltr">`--audio-ascending`</bdi> |
| الصفحات | <bdi dir="ltr">`-p, --pages <spec>`</bdi>, <bdi dir="ltr">`--show-all`</bdi>, <bdi dir="ltr">`--hide-streams`</bdi> |
| المحتوى | <bdi dir="ltr">`--audio-only`</bdi>, <bdi dir="ltr">`--video-only`</bdi>, <bdi dir="ltr">`--subtitle-only`</bdi>, <bdi dir="ltr">`--cover-only`</bdi>, <bdi dir="ltr">`--skip-subtitle`</bdi>, <bdi dir="ltr">`--sub-lang <list>`</bdi>, <bdi dir="ltr">`--auto-subtitles`</bdi>, <bdi dir="ltr">`--skip-cover`</bdi>, <bdi dir="ltr">`--skip-mux`</bdi> |
| الإخراج | <bdi dir="ltr">`-o, --output <template>`</bdi>, <bdi dir="ltr">`--multi-output <template>`</bdi>, <bdi dir="ltr">`-w, --work-dir <dir>`</bdi>, <bdi dir="ltr">`--lang <code>`</bdi>, <bdi dir="ltr">`--no-tags`</bdi>, <bdi dir="ltr">`--archive`</bdi>, <bdi dir="ltr">`--delay <seconds>`</bdi> |
| bilibili | <bdi dir="ltr">`--api web\|tv\|app\|intl`</bdi>, <bdi dir="ltr">`--danmaku`</bdi>, <bdi dir="ltr">`--danmaku-only`</bdi>, <bdi dir="ltr">`--danmaku-format xml,ass`</bdi>, <bdi dir="ltr">`--cookie`</bdi>, <bdi dir="ltr">`--token`</bdi>؛ والمزيد عبر <bdi dir="ltr">`--help-hidden`</bdi> |
| الأدوات | <bdi dir="ltr">`--ffmpeg`</bdi>, <bdi dir="ltr">`--yt-dlp`</bdi>, <bdi dir="ltr">`--use-mp4box`</bdi>, <bdi dir="ltr">`--mp4box`</bdi>, <bdi dir="ltr">`--use-aria2c`</bdi>, <bdi dir="ltr">`--aria2c`</bdi>, <bdi dir="ltr">`--aria2c-args`</bdi>, <bdi dir="ltr">`--single-connection`</bdi> |

<details>
<summary><strong>أولوية الجودة والترميز</strong></summary>

**الجودة والترميز.** يقبل <bdi dir="ltr">`-q`</bdi> التسميات التي يعرضها جدول التدفقات: <bdi dir="ltr">`1080p`</bdi> و<bdi dir="ltr">`720p60`</bdi> على YouTube؛ و<bdi dir="ltr">`8K`</bdi> و<bdi dir="ltr">`Dolby Vision`</bdi> و<bdi dir="ltr">`HDR`</bdi> و<bdi dir="ltr">`4K`</bdi> و<bdi dir="ltr">`1080P60`</bdi> و<bdi dir="ltr">`1080P+`</bdi> و<bdi dir="ltr">`1080P`</bdi> و<bdi dir="ltr">`720P`</bdi> على bilibili. ويقبل <bdi dir="ltr">`-c`</bdi> القيم <bdi dir="ltr">`av1 vp9 hevc avc`</bdi> للفيديو و<bdi dir="ltr">`m4a opus flac eac3 mp3`</bdi> للصوت. هذه أولويات لا مرشّحات: ما لم يُذكر يأتي بعدها، من الأفضل إلى الأقل. يعكس <bdi dir="ltr">`--video-ascending`</bdi> / <bdi dir="ltr">`--audio-ascending`</bdi> الترتيب للحصول على أصغر تنزيل.

</details>

<details>
<summary><strong>اختيار الصفحات</strong></summary>

**الصفحات.** <bdi dir="ltr">`8`</bdi>، <bdi dir="ltr">`1,2`</bdi>، <bdi dir="ltr">`3-5`</bdi>، <bdi dir="ltr">`1-3,10`</bdi>، <bdi dir="ltr">`ALL`</bdi>، <bdi dir="ltr">`LAST`</bdi> (آخر صفحة، وهي أحدث حلقة في البرنامج؛ ويعمل <bdi dir="ltr">`LATEST`</bdi> أيضًا). رابط حلقة واحدة، أو <bdi dir="ltr">`?p=N`</bdi>، يختار تلك الصفحة وحدها. الصفحات هي أجزاء فيديو bilibili، وحلقات الموسم أو البرنامج أو القائمة، وفيديوهات منشور X.

</details>

<details>
<summary><strong>أسماء الملفات ومتغيرات القوالب</strong></summary>

**أسماء الملفات.** لعنصر بصفحة واحدة: <bdi dir="ltr">`<title>`</bdi>؛ ولعدة صفحات (حتى إن اختار <bdi dir="ltr">`-p`</bdi> واحدة فقط): <bdi dir="ltr">`<title>/[P<pageNumberWithZero>]<pageTitle>`</bdi>، مع إضافة أصفار بادئة حسب عدد الصفحات. المتغيرات:

<bdi dir="ltr">`<title>`</bdi> <bdi dir="ltr">`<pageNumber>`</bdi> <bdi dir="ltr">`<pageNumberWithZero>`</bdi> <bdi dir="ltr">`<pageTitle>`</bdi> <bdi dir="ltr">`<id>`</bdi> <bdi dir="ltr">`<site>`</bdi> <bdi dir="ltr">`<uploader>`</bdi> <bdi dir="ltr">`<uploaderId>`</bdi> <bdi dir="ltr">`<quality>`</bdi> <bdi dir="ltr">`<resolution>`</bdi> <bdi dir="ltr">`<fps>`</bdi> <bdi dir="ltr">`<videoCodec>`</bdi> <bdi dir="ltr">`<videoBitrate>`</bdi> <bdi dir="ltr">`<audioCodec>`</bdi> <bdi dir="ltr">`<audioBitrate>`</bdi> <bdi dir="ltr">`<publishDate>`</bdi> <bdi dir="ltr">`<pageDate>`</bdi>، إضافة إلى متغيرات bilibili: <bdi dir="ltr">`<bvid>`</bdi> <bdi dir="ltr">`<aid>`</bdi> <bdi dir="ltr">`<cid>`</bdi> <bdi dir="ltr">`<api>`</bdi>

تقبل التواريخ تنسيقًا: <bdi dir="ltr">`<publishDate:yyyy-MM-dd>`</bdi>. يُضاف الامتداد تلقائيًا.

</details>

<a id="site-notes"></a><a id="notes"></a>

## ملاحظات المواقع

<details>
<summary><strong>YouTube</strong> · الاستخراج والترجمات وقوائم التشغيل</summary>

يحمي YouTube روابط تدفقاته باختبارات مشغّل تحتاج إلى بيئة تشغيل JavaScript، لذا يُترك الاستخراج لـ <bdi dir="ltr">`yt-dlp -J`</bdi> (الذي يشغّلها في deno)؛ وكل ما بعد ذلك ينفّذه haul بنفسه. تُدمج الترجمات المرفوعة، أما المولَّدة تلقائيًا فلا تُدمج إلا مع <bdi dir="ltr">`--auto-subtitles`</bdi>. لا يقدّم googlevideo إلا نطاقات بايت محدودة، لذا تُجلب المسارات نطاقًا واحدًا من 10 MB في كل مرة. ينزّل <bdi dir="ltr">`watch?v=…&list=…`</bdi> الفيديو وحده. أبقِ yt-dlp محدّثًا، ففيه تصل إصلاحات تغييرات YouTube.

</details>

<details>
<summary><strong>X</strong> · المنشورات العامة والصوت</summary>

لا تحتاج منشورات X العامة إلى تسجيل دخول. فيديوهاته ملفات MP4 مفردة تحتوي الصوت، لذا يعرض الجدول الفيديو فقط؛ ويستخرج <bdi dir="ltr">`--audio-only`</bdi> الصوت.

</details>

<details>
<summary><strong>bilibili</strong> · تسجيل الدخول والجودة الأعلى والتعليقات المتحركة (danmaku)</summary>

يُقرأ bilibili عبر واجهاته الخاصة: web وTV وAPP ‏(gRPC) والدولية. دون تسجيل دخول لا تتاح إلا جودات أدنى (عادةً حتى 480P)؛ سجّل الدخول للحصول على 1080P و4K وHDR وDolby Vision وصوت Hi-Res:

<div dir="ltr">

```sh
haul login bilibili                # امسح رمز QR بتطبيق bilibili
haul login bilibili --from-edge    # macOS: إعادة استخدام جلسة Microsoft Edge (أو --from-chrome؛ --profile "Profile 1")
haul login bilibili --tv           # رمز وصول TV، لـ --api tv / --api app
```

</div>

تعمل قراءة جلسة المتصفح حاليًا على macOS فقط: تحتاج الطرفية إلى الوصول الكامل إلى القرص (Full Disk Access)، ويطلب macOS مرة واحدة الوصول إلى عنصر «Safe Storage» في سلسلة المفاتيح. يحفظ <bdi dir="ltr">`--danmaku`</bdi> التعليقات المتحركة بصيغتي XML وASS؛ ويُبقي <bdi dir="ltr">`--danmaku-format ass`</bdi> إحداهما فقط.

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · الحلقات والبيانات الوصفية</summary>

لا يحتاج Xiaoyuzhou وApple Podcasts إلى yt-dlp: تحمل صفحات Xiaoyuzhou رابط الصوت، وتمرّ روابط Apple عبر واجهة iTunes العامة وخلاصة RSS الخاصة بالبرنامج. تحتفظ الحلقات بصيغتها (<bdi dir="ltr">`.mp3`</bdi> أو <bdi dir="ltr">`.m4a`</bdi>) مع غلاف مضمّن، واسم البرنامج كألبوم والمقدّم كفنان. يسرد رابط البرنامج أحدث حلقاته من الأقدم إلى الأحدث، لذا فإن <bdi dir="ltr">`-p LAST`</bdi> هو أحدثها.

</details>

<a id="config"></a>

## الإعدادات

يحتوي <bdi dir="ltr">`~/.config/haul/`</bdi> (أو <bdi dir="ltr">`$HAUL_HOME`</bdi>) على:

| الملف | الغرض |
| --- | --- |
| <bdi dir="ltr">`config.json`</bdi> | القيم الافتراضية لأي خيار |
| <bdi dir="ltr">`cookie.txt`</bdi>, <bdi dir="ltr">`tv-token.txt`</bdi>, <bdi dir="ltr">`app-token.txt`</bdi> | جلسة bilibili |
| <bdi dir="ltr">`archives.txt`</bdi> | الصفحات المنزّلة سابقًا (<bdi dir="ltr">`--archive`</bdi>) |

يستخدم <bdi dir="ltr">`config.json`</bdi> أسماء الخيارات بصيغة camelCase، وتوضع خيارات bilibili تحت <bdi dir="ltr">`"bilibili"`</bdi>، ولا يلزم أن يحتوي إلا ما تغيّره؛ ويمكن أن تكون القائمة مصفوفة أو سلسلة نصية مفصولة بفواصل. وتتقدم خيارات سطر الأوامر عليه:

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
go test ./...                                  # دون شبكة، بضع ثوانٍ
HAUL_LIVE=1 go test -run Live ./internal/...   # يتصل أيضًا بـ bilibili وYouTube وX وApple Podcasts
```

</div>

لا تلمس مجموعة الاختبارات دون اتصال الشبكةَ أبدًا: تتحدث الاختبارات إلى بديل <bdi dir="ltr">`http.RoundTripper`</bdi> يجيب من ردود API مسجّلة ومن شبكات CDN محاكاة (خوادم تقبل النطاقات فقط، واتصالات منقطعة، وخوادم تتجاهل النطاقات). تدمج الاختبارات الشاملة المسارات باستخدام ffmpeg حقيقي وتتحقق من النتيجة بـ ffprobe، وتُتخطى إن لم يكن ffmpeg مثبتًا. البنية والقواعد في [AGENTS.md](AGENTS.md).

يؤدي دفع وسم <bdi dir="ltr">`v*`</bdi> إلى اختبار الإصدار وبنائه لكل المنصات ونشره عبر GitHub Actions.

<a id="acknowledgements"></a>

## شكر وتقدير

[yt-dlp](https://github.com/yt-dlp/yt-dlp)، [FFmpeg](https://ffmpeg.org)، [GPAC](https://gpac.io)، [aria2](https://aria2.github.io)، [pflag](https://github.com/spf13/pflag)، [x/term](https://pkg.go.dev/golang.org/x/term)، [rsc.io/qr](https://pkg.go.dev/rsc.io/qr)، وملاحظات واجهات bilibili في [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) و[bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api).

<a id="license"></a>

## الترخيص

[MIT](LICENSE)

---

<p align="center">
  <strong>رابط واحد. وسائطك بين يديك.</strong><br>
  <a href="#install">احصل على haul</a> · <a href="llms.txt">مرجع الوكلاء</a> · <a href="AGENTS.md">ساهم</a>
</p>

</div>
