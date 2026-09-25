// Package format turns numbers, dates and titles into the text haul prints and the names it gives files.
package format

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// FileSize is 512 bytes, 1.50 KB, 3.00 MB… with decimals after the point.
func FileSize(bytes float64, decimals int) string {
	b := max(0, bytes)
	const kb, mb, gb = 1024.0, 1024.0 * 1024, 1024.0 * 1024 * 1024
	switch {
	case b >= gb:
		return fmt.Sprintf("%.*f GB", decimals, b/gb)
	case b >= mb:
		return fmt.Sprintf("%.*f MB", decimals, b/mb)
	case b >= kb:
		return fmt.Sprintf("%.*f KB", decimals, b/kb)
	}
	return fmt.Sprintf("%d bytes", int64(b))
}

// Duration is 1h02m03s / 02m03s, or 01:02:03 when absolute.
func Duration(seconds int, absolute bool) string {
	h, m, s := seconds/3600, (seconds%3600)/60, seconds%60
	if absolute {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	if h == 0 {
		return fmt.Sprintf("%02dm%02ds", m, s)
	}
	return fmt.Sprintf("%dh%02dm%02ds", h, m, s)
}

// Timestamp renders unix seconds in local time with a Unicode date pattern (yyyy-MM-dd_HH-mm-ss). 0 renders as "null".
func Timestamp(ts int64, pattern string) string {
	if ts == 0 {
		return "null"
	}
	return FormatDate(time.Unix(ts, 0), pattern)
}

// FormatDate renders t with a Unicode date pattern: yyyy yy MM M dd d HH H hh h mm m ss s SSS a, and 'quoted' text.
func FormatDate(t time.Time, pattern string) string {
	var b strings.Builder
	rs := []rune(pattern)
	for i := 0; i < len(rs); {
		c := rs[i]
		if c == '\'' {
			j := i + 1
			for j < len(rs) && rs[j] != '\'' {
				j++
			}
			if j == i+1 {
				b.WriteRune('\'')
			} else {
				b.WriteString(string(rs[i+1 : j]))
			}
			i = j + 1
			continue
		}
		if !unicode.IsLetter(c) {
			b.WriteRune(c)
			i++
			continue
		}
		n := 1
		for i+n < len(rs) && rs[i+n] == c {
			n++
		}
		b.WriteString(dateField(t, c, n))
		i += n
	}
	return b.String()
}

func dateField(t time.Time, c rune, n int) string {
	pad := func(v int) string { return fmt.Sprintf("%0*d", n, v) }
	switch c {
	case 'y':
		if n == 2 {
			return fmt.Sprintf("%02d", t.Year()%100)
		}
		return pad(t.Year())
	case 'M':
		if n >= 3 {
			if n == 3 {
				return t.Month().String()[:3]
			}
			return t.Month().String()
		}
		return pad(int(t.Month()))
	case 'd':
		return pad(t.Day())
	case 'H':
		return pad(t.Hour())
	case 'h':
		h := t.Hour() % 12
		if h == 0 {
			h = 12
		}
		return pad(h)
	case 'm':
		return pad(t.Minute())
	case 's':
		return pad(t.Second())
	case 'S':
		ms := fmt.Sprintf("%09d", t.Nanosecond())
		if n > 9 {
			n = 9
		}
		return ms[:n]
	case 'a':
		if t.Hour() < 12 {
			return "AM"
		}
		return "PM"
	case 'E':
		if n >= 4 {
			return t.Weekday().String()
		}
		return t.Weekday().String()[:3]
	}
	return strings.Repeat(string(c), n)
}

// FFmpegCreationTime is ISO 8601 UTC with microseconds, the shape ffmpeg accepts for creation_time.
func FFmpegCreationTime(ts int64) string {
	return time.Unix(ts, 0).UTC().Format("2006-01-02T15:04:05.000000Z")
}

// ISO is ISO 8601 UTC with seconds.
func ISO(ts int64) string { return time.Unix(ts, 0).UTC().Format(time.RFC3339) }

// ParseDateTime reads `yyyy-MM-dd HH:mm:ss` (bilibili's pgc pub_time) as local time.
func ParseDateTime(text string) (int64, bool) {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", text, time.Local)
	if err != nil {
		return 0, false
	}
	return t.Unix(), true
}

// ValidFileName replaces the characters no file system likes, slashes included (they separate directories).
func ValidFileName(input string) string {
	var b strings.Builder
	for _, r := range input {
		if r < 32 || strings.ContainsRune(`"<>|:*?\/`, r) {
			b.WriteByte('_')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CleanName is the cleanup every name component gets: invalid characters out, spaces and trailing dots trimmed.
func CleanName(input string) string {
	s := strings.TrimSpace(ValidFileName(input))
	s = strings.TrimRight(s, ".")
	return strings.TrimSpace(s)
}

// ChangeExtension swaps the extension of a path: a/b.mp4 → a/b.xml.
func ChangeExtension(p, ext string) string {
	return strings.TrimSuffix(p, path.Ext(p)) + "." + ext
}

var (
	decimalEntity = regexp.MustCompile(`&#(\d+);`)
	hexEntity     = regexp.MustCompile(`&#[xX]([0-9a-fA-F]+);`)
	namedEntities = strings.NewReplacer("&quot;", `"`, "&apos;", "'", "&lt;", "<", "&gt;", ">", "&nbsp;", " ", "&amp;", "&")
)

// UnescapeEntities turns &amp; &lt;… and numeric entities (&#26159; &#x4E2D;) back into characters.
func UnescapeEntities(text string) string {
	if !strings.Contains(text, "&") {
		return text
	}
	numeric := func(re *regexp.Regexp, base int) func(string) string {
		return func(m string) string {
			n, err := strconv.ParseInt(re.FindStringSubmatch(m)[1], base, 32)
			if err != nil || n < 0 || n > unicode.MaxRune {
				return m
			}
			return string(rune(n))
		}
	}
	text = decimalEntity.ReplaceAllStringFunc(text, numeric(decimalEntity, 10))
	text = hexEntity.ReplaceAllStringFunc(text, numeric(hexEntity, 16))
	return namedEntities.Replace(text)
}

// QueryValue is the raw value of name in a URL's query, or "".
func QueryValue(name, u string) string {
	q := strings.IndexByte(u, '?')
	if q < 0 {
		return ""
	}
	query := u[q+1:]
	if h := strings.IndexByte(query, '#'); h >= 0 {
		query = query[:h]
	}
	for _, pair := range strings.Split(query, "&") {
		k, v, ok := strings.Cut(pair, "=")
		if ok && k == name {
			return v
		}
	}
	return ""
}

// RandomString is length characters of the alphabet bilibili's device ids use.
func RandomString(length int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[rand.IntN(len(chars))]
	}
	return string(b)
}

// UnixSeconds is the current time in seconds, as text.
func UnixSeconds() string { return strconv.FormatInt(time.Now().Unix(), 10) }

// CompactTimestamp is yyyyMMddHHmmssSSS, used to seed the TV login fingerprint.
func CompactTimestamp() string { return FormatDate(time.Now(), "yyyyMMddHHmmssSSS") }

// PercentEncode escapes everything but letters, digits and -._~.
func PercentEncode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}
