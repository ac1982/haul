// Package subtitle converts the subtitle formats sites serve into SRT, and names languages for the container.
package subtitle

import (
	"fmt"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// ToSRT converts a subtitle file to SRT. SRT passes through; ASS is not SRT and is refused.
func ToSRT(format media.SubtitleFormat, data []byte) (string, error) {
	switch format {
	case media.SRT:
		return string(data), nil
	case media.YouTubeJSON3:
		j, err := jsonv.Parse(data)
		if err != nil {
			return "", err
		}
		return fromJSON3(j), nil
	case media.BilibiliJSON:
		j, err := jsonv.Parse(data)
		if err != nil {
			return "", err
		}
		if j.Has("events") {
			return fromJSON3(j), nil
		}
		return fromBilibili(j), nil
	}
	return "", errs.New("cannot convert %s subtitles to SRT", format)
}

// fromBilibili converts bilibili's {"body": [{"from", "to", "content"}]}.
func fromBilibili(j jsonv.Value) string {
	var b strings.Builder
	for i, line := range j.Get("body").Array() {
		from := line.Get("from").FloatOr(0)
		to := line.Get("to").FloatOr(0)
		fmt.Fprintf(&b, "%d\n%s --> %s\n", i+1, Time(from), Time(to))
		if line.Has("content") {
			b.WriteString(line.Get("content").String() + "\n")
		}
		b.WriteString("\n")
	}
	return b.String()
}

// fromJSON3 converts YouTube's json3 captions. Auto-generated tracks interleave newline-only events; those are dropped.
func fromJSON3(j jsonv.Value) string {
	var b strings.Builder
	n := 0
	for _, e := range j.Get("events").Array() {
		var text strings.Builder
		for _, seg := range e.Get("segs").Array() {
			text.WriteString(seg.Get("utf8").String())
		}
		t := strings.TrimSpace(text.String())
		start, ok := e.Get("tStartMs").Float()
		if t == "" || !ok {
			continue
		}
		end := start + e.Get("dDurationMs").FloatOr(0)
		n++
		fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n\n", n, Time(start/1000), Time(end/1000), t)
	}
	return b.String()
}

// Time is an SRT timestamp, hh:mm:ss,fff.
func Time(seconds float64) string {
	d := time.Duration(seconds*1000+0.5) * time.Millisecond
	h := int(d / time.Hour)
	m := int(d % time.Hour / time.Minute)
	s := int(d % time.Minute / time.Second)
	ms := int(d % time.Second / time.Millisecond)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

// Language is how a subtitle stream is tagged in the output: its ISO 639-2 code and a readable name.
type Language struct {
	Code string
	Name string
}

// LanguageOf maps a BCP 47 tag (zh-Hans, en-US, pt-BR) to its ISO 639-2 code and name; unknown tags are "und".
func LanguageOf(tag string, auto bool) Language {
	norm := strings.ToLower(strings.ReplaceAll(tag, "_", "-"))
	l, ok := languages[norm]
	if !ok {
		base, _, _ := strings.Cut(norm, "-")
		l, ok = languages[base]
		if ok && base != norm {
			l.Name += " (" + tag[len(base)+1:] + ")"
		}
	}
	if !ok {
		l = Language{Code: "und", Name: tag}
	}
	if auto {
		l.Name += ", auto-generated"
	}
	return l
}

var languages = map[string]Language{
	"zh": {"chi", "中文"}, "zh-cn": {"chi", "中文（简体）"}, "zh-hans": {"chi", "中文（简体）"}, "zh-sg": {"chi", "中文（新加坡）"},
	"zh-tw": {"chi", "中文（台灣繁體）"}, "zh-hk": {"chi", "中文（香港繁體）"}, "zh-hant": {"chi", "中文（繁體）"},
	"yue": {"chi", "粵語"}, "nan": {"nan", "閩南語"}, "hak": {"hak", "Hak-kâ-fa"},
	"en": {"eng", "English"}, "ja": {"jpn", "日本語"}, "ko": {"kor", "한국어"}, "fr": {"fre", "Français"},
	"de": {"ger", "Deutsch"}, "es": {"spa", "Español"}, "pt": {"por", "Português"}, "it": {"ita", "Italiano"},
	"ru": {"rus", "Русский"}, "ar": {"ara", "العربية"}, "hi": {"hin", "हिन्दी"}, "th": {"tha", "ไทย"},
	"vi": {"vie", "Tiếng Việt"}, "id": {"ind", "Bahasa Indonesia"}, "ms": {"may", "Bahasa Melayu"},
	"tr": {"tur", "Türkçe"}, "pl": {"pol", "Polski"}, "nl": {"dut", "Nederlands"}, "sv": {"swe", "Svenska"},
	"no": {"nor", "Norsk"}, "nb": {"nob", "Norsk bokmål"}, "da": {"dan", "Dansk"}, "fi": {"fin", "Suomi"},
	"cs": {"cze", "Čeština"}, "sk": {"slo", "Slovenčina"}, "hu": {"hun", "Magyar"}, "ro": {"rum", "Română"},
	"el": {"gre", "Ελληνικά"}, "he": {"heb", "עברית"}, "iw": {"heb", "עברית"}, "uk": {"ukr", "Українська"},
	"bg": {"bul", "Български"}, "hr": {"hrv", "Hrvatski"}, "sr": {"srp", "Српски"}, "sl": {"slv", "Slovenščina"},
	"lt": {"lit", "Lietuvių"}, "lv": {"lav", "Latviešu"}, "et": {"est", "Eesti"}, "fa": {"per", "فارسی"},
	"ur": {"urd", "اردو"}, "bn": {"ben", "বাংলা"}, "ta": {"tam", "தமிழ்"}, "te": {"tel", "తెలుగు"},
	"mr": {"mar", "मराठी"}, "gu": {"guj", "ગુજરાતી"}, "kn": {"kan", "ಕನ್ನಡ"}, "ml": {"mal", "മലയാളം"},
	"fil": {"fil", "Filipino"}, "tl": {"tgl", "Tagalog"}, "ca": {"cat", "Català"}, "eu": {"baq", "Euskara"},
	"gl": {"glg", "Galego"}, "af": {"afr", "Afrikaans"}, "sw": {"swa", "Kiswahili"}, "mn": {"mon", "Монгол"},
	"km": {"khm", "ខ្មែរ"}, "lo": {"lao", "ລາວ"}, "my": {"bur", "မြန်မာ"}, "ne": {"nep", "नेपाली"},
	"si": {"sin", "සිංහල"}, "ka": {"geo", "ქართული"}, "hy": {"arm", "Հայերեն"}, "az": {"aze", "Azərbaycan"},
	"kk": {"kaz", "Қазақ"}, "uz": {"uzb", "Oʻzbek"}, "is": {"ice", "Íslenska"}, "ga": {"gle", "Gaeilge"},
	"cy": {"wel", "Cymraeg"}, "mt": {"mlt", "Malti"}, "sq": {"alb", "Shqip"}, "mk": {"mac", "Македонски"},
	"bs": {"bos", "Bosanski"}, "be": {"bel", "Беларуская"}, "la": {"lat", "Latina"}, "eo": {"epo", "Esperanto"},
}
