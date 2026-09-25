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
  <bdi><strong>हिन्दी</strong></bdi> ·
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

<h1 align="center">एक लिंक। आपका मीडिया।</h1>

<p align="center">
  <a href="https://github.com/ac1982/haul/actions/workflows/build.yml"><img src="https://github.com/ac1982/haul/actions/workflows/build.yml/badge.svg" alt="बिल्ड स्थिति"></a>
  <a href="https://github.com/ac1982/haul/releases"><img src="https://img.shields.io/github/v/release/ac1982/haul?color=63c9a5&amp;label=release" alt="नवीनतम रिलीज़"></a>
  <img src="https://img.shields.io/badge/Go-1.26%2B-00ADD8?logo=go&amp;logoColor=white" alt="Go 1.26 या नया">
  <img src="https://img.shields.io/badge/macOS%20%C2%B7%20Linux%20%C2%B7%20Windows-9bb5ca" alt="macOS, Linux और Windows">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-63c9a5" alt="MIT लाइसेंस"></a>
</p>

<p align="center">
  <strong>वीडियो और ऑडियो डाउनलोड करने वाला कमांड-लाइन टूल, जो लोगों और AI एजेंटों दोनों के लिए बना है।</strong><br>
  लिंक दें। स्ट्रीम चुनें। वीडियो, ऑडियो, उपशीर्षक, अध्याय और कवर एक ही फ़ाइल में पाएँ।
</p>

<p align="center">
  <a href="#install">इंस्टॉल करना</a> ·
  <a href="#usage">त्वरित शुरुआत</a> ·
  <a href="#sites">समर्थित साइटें</a> ·
  <a href="#for-ai-agents-and-scripts">एजेंट गाइड</a> ·
  <a href="#options">विकल्प</a> ·
  <a href="../../releases">रिलीज़</a>
</p>

---

<a id="small-command-complete-download"></a>

## छोटा कमांड। पूरा डाउनलोड।

| ↓ आपका मीडिया, आपके तरीक़े से | ⌘ एक बाइनरी, हर डेस्कटॉप | { } ऑटोमेशन के लिए तैयार |
| :--- | :--- | :--- |
| पाँचों साइटों पर एक जैसे शब्दों से गुणवत्ता, कोडेक, पेज और फ़ाइल नाम चुनें | macOS, Linux और Windows के लिए एक Go बाइनरी, जो समानांतर रेंज डाउनलोड करती है और रुके डाउनलोड फिर से शुरू करती है | stdout पर एक JSON दस्तावेज़, stderr पर लॉग, अर्थपूर्ण एग्ज़िट कोड, और टर्मिनल के बिना कोई सवाल नहीं |

| 01 / जाँचें | 02 / चुनें | 03 / डाउनलोड | 04 / जोड़ें |
| :--- | :--- | :--- | :--- |
| **हर स्ट्रीम देखें** | **अपनी प्राथमिकताएँ तय करें** | **घर ले आएँ** | **सब एक साथ** |
| पेज, कोडेक और आकार | गुणवत्ता, ऑडियो और पेज | रेंज या सेगमेंट में | ट्रैक, उपशीर्षक और कवर |
| `haul info "<url>"` | `-q 720p -c avc,m4a` | `haul "<url>"` | `ffmpeg / MP4Box` |

<details>
<summary><strong>टर्मिनल के अंदर एक झलक</strong></summary>

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

> व्यक्तिगत, शोध और अन्य गैर-व्यावसायिक उपयोग के लिए। कॉपीराइट और हर साइट की शर्तों का सम्मान करना आपकी ज़िम्मेदारी है।

<a id="install"></a>

## इंस्टॉल करना

haul **macOS, Linux और Windows** (amd64 और arm64) पर चलता है। ट्रैक जोड़ने (mux) के लिए [ffmpeg](https://ffmpeg.org) चाहिए; YouTube और X को [yt-dlp](https://github.com/yt-dlp/yt-dlp) भी चाहिए, और YouTube के प्लेयर की जाँचों के लिए [deno](https://deno.com) चाहिए:

```sh
brew install ffmpeg yt-dlp deno          # macOS
sudo apt install ffmpeg yt-dlp           # Debian / Ubuntu; deno: https://deno.com
winget install Gyan.FFmpeg               # Windows, एक बार में एक पैकेज
winget install yt-dlp.yt-dlp
winget install DenoLand.Deno
```

<a id="get-the-binary"></a><a id="binary"></a>

### बाइनरी प्राप्त करें

[Releases](../../releases) से अपने सिस्टम का आर्काइव डाउनलोड करें — `haul-<version>-<os>-<arch>.tar.gz` (Windows पर `.zip`) — और `haul` को अपने `PATH` के किसी भी फ़ोल्डर में रखें:

```sh
tar -xzf haul-*-darwin-arm64.tar.gz
mkdir -p ~/.local/bin && mv haul-*/haul ~/.local/bin/
```

वर्तमान रिलीज़ प्रक्रिया macOS संस्करणों पर हस्ताक्षर करती है और Apple से नोटरीकरण कराती है। `haul-<version>-darwin-<arch>.pkg` चुनें: इसमें नोटरीकरण टिकट जुड़ा है और यह `haul` को `/usr/local/bin` में इंस्टॉल करता है। आर्काइव में वही हस्ताक्षरित प्रोग्राम है, लेकिन पहली जाँच के लिए इंटरनेट लग सकता है। पुराने संस्करण बिना हस्ताक्षर के हो सकते हैं।

<details>
<summary><strong>सोर्स से बिल्ड करें</strong> · Go 1.26+</summary>

```sh
go install github.com/ac1982/haul/cmd/haul@latest
```

या क्लोन की गई रिपॉज़िटरी से:

```sh
go build -o haul ./cmd/haul
```

</details>

<a id="usage"></a>

## त्वरित शुरुआत

एक लिंक से शुरू करें। haul डिफ़ॉल्ट रूप से उपलब्ध सर्वोत्तम स्ट्रीम चुनता है।

```sh
haul "https://youtu.be/DdCEmlAydcw"
```

डाउनलोड से पहले जानकारी देखें, या गुणवत्ता और कोडेक चुनें:

```sh
haul info "https://youtu.be/DdCEmlAydcw"
haul -q 720p -c avc,m4a "https://youtu.be/DdCEmlAydcw"  # QuickTime के लिए H.264 + AAC
```

<details>
<summary><strong>कमांड संदर्भ</strong></summary>

| कमांड | अर्थ |
| --- | --- |
| `haul <url> [options]` | डाउनलोड करें (`haul download <url>` के समान) |
| `haul info <url> [--urls]` | आइटम, उसके पेज और स्ट्रीम दिखाएँ; कुछ डाउनलोड नहीं करता |
| `haul login bilibili` | बेहतर गुणवत्ता के लिए bilibili में लॉग इन करें |
| `haul templates` | फ़ाइल नाम टेम्पलेट के वेरिएबल |
| `haul --help` | साइटें, एजेंट उपयोग, एग्ज़िट कोड, उदाहरण |
| `haul download --help` | सभी विकल्प |

</details>

<a id="sites"></a>

### समर्थित साइटें

| साइट | लिंक | ज़रूरी टूल |
| --- | --- | --- |
| **YouTube** | `youtube.com/watch?v=…`, `youtu.be/…`, `/shorts/…`, `/embed/…`, `/live/…` | yt-dlp, deno |
| **X** | `x.com/<user>/status/<id>`, `twitter.com/…`; `/video/<n>` पोस्ट का एक वीडियो चुनता है | yt-dlp |
| **bilibili** | वीडियो, bangumi, पाठ्यक्रम, संग्रह, सीरीज़, पसंदीदा, उपयोगकर्ता स्पेस, `b23.tv`, सीधे `BV…` `av…` `ep…` `ss…` `md…` | – |
| **Xiaoyuzhou** | `xiaoyuzhoufm.com/episode/<id>`, `/podcast/<id>` | – |
| **Apple Podcasts** | `podcasts.apple.com/<cc>/podcast/<name>/id<show>`; एक एपिसोड के लिए `?i=<episode>` के साथ | – |

<a id="make-it-yours"></a><a id="examples"></a>

### अपने हिसाब से

**सिर्फ़ ऑडियो**

```sh
haul --audio-only "https://youtu.be/DdCEmlAydcw"
haul --audio-only -c m4a "https://youtu.be/DdCEmlAydcw"   # Opus की जगह AAC
```

**कुछ भाग, पूरा सीज़न या नवीनतम एपिसोड**

```sh
# चुने हुए भाग, आपके Movies फ़ोल्डर में
haul -p 1-3,10 -w ~/Movies "https://www.bilibili.com/video/BV1Wv411h7kN"

# सीज़न के सभी एपिसोड
haul -p ALL "https://www.bilibili.com/bangumi/play/ss33073"

# पॉडकास्ट का सबसे नया एपिसोड
haul -p LAST "https://podcasts.apple.com/us/podcast/the-daily/id1200361736"

# X पोस्ट के सभी वीडियो
haul -p ALL "https://x.com/<user>/status/<id>"
```

**व्यवस्थित फ़ाइलें, उपशीर्षक या इंटरैक्टिव चयन**

```sh
haul -o "<uploader>/<title> [<quality>]" "https://youtu.be/DdCEmlAydcw"
haul --sub-lang en,zh "https://youtu.be/DdCEmlAydcw"   # सिर्फ़ इन भाषाओं के उपशीर्षक
haul -i "BV1qt4y1X7TW"                                 # तीर वाली कुंजियों से स्ट्रीम चुनें
```

<a id="for-ai-agents-and-scripts"></a><a id="agents"></a>

## AI एजेंटों और स्क्रिप्ट के लिए

`haul --help` में वह सब है जो किसी एजेंट को चाहिए; [llms.txt](llms.txt) वही गाइड फ़ाइल के रूप में है। संक्षेप में:

```sh
haul info --json "<url>"                                # 1. जानकारी देखें
haul --json --video-stream 2 --audio-stream 0 "<url>"   # 2. ठीक यही स्ट्रीम डाउनलोड करें
haul --json -q 720p -c avc,m4a "<url>"                  #    या प्राथमिकताओं को चुनने दें
```

- `--json` के साथ **stdout पर केवल एक JSON दस्तावेज़** आता है; प्रगति और लॉग stderr पर जाते हैं। उसकी `files` सूची में आउटपुट फ़ाइलें होती हैं, नई या पहले से मौजूद।
- **टर्मिनल के बिना कुछ भी इंटरैक्टिव नहीं।** टर्मिनल न हो तो `-i` एग्ज़िट कोड 2 से विफल होता है और उसकी जगह इस्तेमाल होने वाले फ़्लैग बताता है।
- `info` में स्ट्रीम इंडेक्स उसी क्रम में होते हैं जिसमें haul चुनता है; `info` और डाउनलोड दोनों को वही `-q` / `-c` दें।
- प्लेलिस्ट, सीज़न, शो या कई वीडियो वाली पोस्ट पर `info` उसके पेज दिखाता है; `-p <n>` उस पेज की स्ट्रीम और उपशीर्षक भी जोड़ता है।
- डिस्क पर पहले से मौजूद पेज `"status": "skipped", "reason": "exists"` के रूप में बताए जाते हैं, इसलिए दोबारा चलाना सुरक्षित है।

<details>
<summary><strong>JSON उत्तर का उदाहरण</strong> · एक सफल डाउनलोड</summary>

डाउनलोड यह छापता है:

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

यह `haul --json -q 360p -c avc,m4a …` का उत्तर है। यहाँ कुंजियाँ पढ़ने के क्रम में हैं और स्ट्रीम सूचियाँ सिर्फ़ चुनी गई स्ट्रीम तक छोटी की गई हैं; असली दस्तावेज़ अपनी कुंजियाँ वर्णानुक्रम में रखता है और हर स्ट्रीम दिखाता है।

</details>

विफलता पर `"ok": false` के साथ `"error": {"kind", "message", "exitCode"}` छपता है, और उससे पहले पूरे हुए पेज उत्तर में बने रहते हैं।

| एग्ज़िट कोड | अर्थ | `error.kind` |
| --- | --- | --- |
| 0 | पूरा हुआ | – |
| 1 | डाउनलोड या एक्सट्रैक्शन विफल | `failed` |
| 2 | असमर्थित लिंक या गलत विकल्प मान | `input` |
| 3 | ज़रूरी टूल नहीं मिला (ffmpeg, yt-dlp) | `dependency` |
| 4 | लॉगिन ज़रूरी है या समाप्त हो गया | `auth` |
| 64 | गलत कमांड लाइन (अज्ञात फ़्लैग, गायब आर्ग्युमेंट) | – |
| 130 | रद्द किया गया | `cancelled` |

<a id="options"></a>

## विकल्प

`haul download --help` सभी विकल्प दिखाता है। मुख्य विकल्प:

| श्रेणी | विकल्प |
| --- | --- |
| सामान्य | `--json`, `--config <file>`, `--debug` |
| स्ट्रीम | `-q, --quality <list>`, `-c, --codec <list>`, `--video-stream <n>`, `--audio-stream <n>`, `-i, --interactive`, `--video-ascending`, `--audio-ascending` |
| पेज | `-p, --pages <spec>`, `--show-all`, `--hide-streams` |
| सामग्री | `--audio-only`, `--video-only`, `--subtitle-only`, `--cover-only`, `--skip-subtitle`, `--sub-lang <list>`, `--auto-subtitles`, `--skip-cover`, `--skip-mux` |
| आउटपुट | `-o, --output <template>`, `--multi-output <template>`, `-w, --work-dir <dir>`, `--lang <code>`, `--no-tags`, `--archive`, `--delay <seconds>` |
| bilibili | `--api web\|tv\|app\|intl`, `--danmaku`, `--danmaku-only`, `--danmaku-format xml,ass`, `--cookie`, `--token`; और विकल्प `--help-hidden` में |
| टूल | `--ffmpeg`, `--yt-dlp`, `--use-mp4box`, `--mp4box`, `--use-aria2c`, `--aria2c`, `--aria2c-args`, `--single-connection` |

<details>
<summary><strong>गुणवत्ता और कोडेक की प्राथमिकता</strong></summary>

**गुणवत्ता और कोडेक।** `-q` वही लेबल लेता है जो स्ट्रीम तालिका दिखाती है: YouTube पर `1080p`, `720p60`; bilibili पर `8K`, `Dolby Vision`, `HDR`, `4K`, `1080P60`, `1080P+`, `1080P`, `720P`। `-c` वीडियो के लिए `av1 vp9 hevc avc` और ऑडियो के लिए `m4a opus flac eac3 mp3` लेता है। ये प्राथमिकताएँ हैं, फ़िल्टर नहीं: जो सूची में नहीं है वह बाद में आता है, सर्वोत्तम पहले। `--video-ascending` / `--audio-ascending` सबसे छोटे डाउनलोड के लिए क्रम उलट देते हैं।

</details>

<details>
<summary><strong>पेज चुनना</strong></summary>

**पेज।** `8`, `1,2`, `3-5`, `1-3,10`, `ALL`, `LAST` (अंतिम पेज, यानी शो का सबसे नया एपिसोड; `LATEST` भी चलता है)। किसी एक एपिसोड का लिंक, या `?p=N`, वह पेज अपने आप चुन लेता है। पेज का अर्थ है bilibili वीडियो के भाग, सीज़न, शो या सूची के एपिसोड, और X पोस्ट के वीडियो।

</details>

<details>
<summary><strong>फ़ाइल नाम और टेम्पलेट वेरिएबल</strong></summary>

**फ़ाइल नाम।** एक पेज वाला आइटम: `<title>`; कई पेज वाला (भले `-p` सिर्फ़ एक चुने): `<title>/[P<pageNumberWithZero>]<pageTitle>`, कुल पेजों की संख्या के अनुसार आगे शून्य लगाकर। वेरिएबल: `<title>` `<pageNumber>` `<pageNumberWithZero>` `<pageTitle>` `<id>` `<site>` `<uploader>` `<uploaderId>` `<quality>` `<resolution>` `<fps>` `<videoCodec>` `<videoBitrate>` `<audioCodec>` `<audioBitrate>` `<publishDate>` `<pageDate>`, साथ ही bilibili के `<bvid>` `<aid>` `<cid>` `<api>`। तारीख़ों को फ़ॉर्मैट दिया जा सकता है: `<publishDate:yyyy-MM-dd>`। एक्सटेंशन अपने आप जुड़ता है।

</details>

<a id="site-notes"></a><a id="notes"></a>

## साइट संबंधी जानकारी

<details>
<summary><strong>YouTube</strong> · एक्सट्रैक्शन, उपशीर्षक और प्लेलिस्ट</summary>

YouTube अपने स्ट्रीम URL को प्लेयर की ऐसी जाँचों से सुरक्षित रखता है जिनके लिए JavaScript रनटाइम चाहिए, इसलिए एक्सट्रैक्शन `yt-dlp -J` पर छोड़ा गया है (जो उन्हें deno में चलाता है); उसके बाद का सारा काम haul खुद करता है। अपलोड किए गए उपशीर्षक फ़ाइल में जोड़े जाते हैं; स्वचालित रूप से बने उपशीर्षक सिर्फ़ `--auto-subtitles` के साथ। googlevideo केवल सीमित बाइट रेंज देता है, इसलिए ट्रैक एक बार में 10 MB की एक रेंज करके लाए जाते हैं। `watch?v=…&list=…` सिर्फ़ वही वीडियो डाउनलोड करता है। yt-dlp को अपडेट रखें: YouTube के बदलावों के सुधार वहीं आते हैं।

</details>

<details>
<summary><strong>X</strong> · सार्वजनिक पोस्ट और ऑडियो</summary>

X की सार्वजनिक पोस्ट के लिए लॉगिन नहीं चाहिए। उसके वीडियो ऑडियो सहित एक ही MP4 फ़ाइल होते हैं, इसलिए तालिका में सिर्फ़ वीडियो दिखता है; `--audio-only` ऑडियो निकालता है।

</details>

<details>
<summary><strong>bilibili</strong> · लॉगिन, बेहतर गुणवत्ता और danmaku</summary>

bilibili को उसके अपने web, TV, APP (gRPC) और अंतरराष्ट्रीय API से पढ़ा जाता है। लॉग आउट रहने पर सिर्फ़ कम गुणवत्ता मिलती है (आम तौर पर 480P तक); 1080P, 4K, HDR, Dolby Vision और Hi-Res ऑडियो के लिए लॉग इन करें:

```sh
haul login bilibili                # bilibili ऐप से QR कोड स्कैन करें
haul login bilibili --from-edge    # macOS: Microsoft Edge का लॉगिन दोबारा इस्तेमाल करें (या --from-chrome; --profile "Profile 1")
haul login bilibili --tv           # TV एक्सेस टोकन, --api tv / --api app के लिए
```

ब्राउज़र का लॉगिन पढ़ना फ़िलहाल सिर्फ़ macOS पर काम करता है: इसके लिए टर्मिनल को Full Disk Access चाहिए, और macOS एक बार “Safe Storage” कीचेन आइटम की अनुमति माँगता है। `--danmaku` स्क्रीन पर चलने वाले कमेंट (danmaku) XML और ASS में सहेजता है; `--danmaku-format ass` इनमें से सिर्फ़ एक रखता है।

</details>

<details>
<summary><strong>Xiaoyuzhou &amp; Apple Podcasts</strong> · एपिसोड और मेटाडेटा</summary>

Xiaoyuzhou और Apple Podcasts को yt-dlp नहीं चाहिए: Xiaoyuzhou के पेजों में ऑडियो लिंक होता है, और Apple के लिंक सार्वजनिक iTunes API और शो की RSS फ़ीड से होकर जाते हैं। एपिसोड अपना फ़ॉर्मैट (`.mp3` या `.m4a`) बनाए रखते हैं, कवर जुड़ता है, शो एल्बम और होस्ट कलाकार बनता है। शो लिंक हाल के एपिसोड पुराने से नए क्रम में दिखाता है, इसलिए `-p LAST` सबसे नया एपिसोड है।

</details>

<a id="config"></a>

## कॉन्फ़िगरेशन

`~/.config/haul/` (या `$HAUL_HOME`) में ये फ़ाइलें रहती हैं:

| फ़ाइल | उद्देश्य |
| --- | --- |
| `config.json` | किसी भी विकल्प के डिफ़ॉल्ट मान |
| `cookie.txt`, `tv-token.txt`, `app-token.txt` | bilibili लॉगिन |
| `archives.txt` | पहले से डाउनलोड किए गए पेज (`--archive`) |

`config.json` में विकल्पों के नाम camelCase में होते हैं, bilibili वाले `"bilibili"` के अंदर, और इसमें सिर्फ़ वही लिखना होता है जो बदलना है; सूची array भी हो सकती है और कॉमा से अलग की गई string भी। कमांड-लाइन फ़्लैग इसे ओवरराइड करते हैं:

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

## डेवलपमेंट

```sh
go build ./cmd/haul
go test ./...                                  # ऑफ़लाइन, कुछ सेकंड
HAUL_LIVE=1 go test -run Live ./internal/...   # bilibili, YouTube, X और Apple Podcasts से भी संपर्क करता है
```

ऑफ़लाइन टेस्ट कभी नेटवर्क को नहीं छूते: टेस्ट एक स्टब `http.RoundTripper` से बात करते हैं, जो रिकॉर्ड किए गए API उत्तरों और नकली CDN (सिर्फ़ रेंज देने वाले सर्वर, टूटते कनेक्शन, रेंज को अनदेखा करने वाले सर्वर) से जवाब देता है। शुरू से अंत तक के टेस्ट असली ffmpeg से ट्रैक जोड़ते हैं और नतीजे को ffprobe से जाँचते हैं; ffmpeg इंस्टॉल न हो तो वे छोड़ दिए जाते हैं। संरचना और नियमों के लिए [AGENTS.md](AGENTS.md) देखें।

`v*` टैग पुश करने पर GitHub Actions टेस्ट चलाता है, हर प्लेटफ़ॉर्म के लिए क्रॉस-कंपाइल करता है और रिलीज़ प्रकाशित करता है।

<a id="acknowledgements"></a>

## आभार

[yt-dlp](https://github.com/yt-dlp/yt-dlp), [FFmpeg](https://ffmpeg.org), [GPAC](https://gpac.io), [aria2](https://aria2.github.io), [pflag](https://github.com/spf13/pflag), [x/term](https://pkg.go.dev/golang.org/x/term), [rsc.io/qr](https://pkg.go.dev/rsc.io/qr), तथा [bilibili-API-collect](https://github.com/SocialSisterYi/bilibili-API-collect) और [bilibili-grpc-api](https://github.com/SeeFlowerX/bilibili-grpc-api) के API नोट्स।

<a id="license"></a>

## लाइसेंस

[MIT](LICENSE)

---

<p align="center">
  <strong>एक लिंक। आपका मीडिया।</strong><br>
  <a href="#install">haul पाएँ</a> · <a href="llms.txt">एजेंट संदर्भ</a> · <a href="AGENTS.md">योगदान करें</a>
</p>
