package podcast

import (
	"strconv"
	"strings"
	"time"
)

// parseISODate reads `2024-09-20T09:00:00Z`, with or without fractional seconds, as UTC; zero when it is not one.
func parseISODate(text string) time.Time {
	t, err := time.Parse(time.RFC3339, strings.TrimSpace(text))
	if err != nil {
		return time.Time{}
	}
	return t.UTC()
}

// rfc822Layouts are the shapes RSS dates come in: numeric zones, with and without the weekday or seconds.
var rfc822Layouts = []string{
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04 -0700",
}

// zoneOffsets are the zone names RFC 822 defines. Go's parser gives an unknown name a zero offset, which would
// shift American feeds' dates by hours.
var zoneOffsets = map[string]string{
	"UT": "+0000", "UTC": "+0000", "GMT": "+0000", "Z": "+0000",
	"EST": "-0500", "EDT": "-0400", "CST": "-0600", "CDT": "-0500",
	"MST": "-0700", "MDT": "-0600", "PST": "-0800", "PDT": "-0700",
}

// parseRFC822Date reads RSS dates, `Fri, 20 Sep 2024 09:00:00 +0000` or `… GMT`, as UTC; zero when it is not one.
func parseRFC822Date(text string) time.Time {
	text = strings.TrimSpace(text)
	if i := strings.LastIndexByte(text, ' '); i >= 0 {
		if offset, ok := zoneOffsets[strings.ToUpper(text[i+1:])]; ok {
			text = text[:i+1] + offset
		}
	}
	for _, layout := range rfc822Layouts {
		if t, err := time.Parse(layout, text); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// parseDuration reads `3600`, `59:30` or `1:02:03`; a part that is not a number counts as 0.
func parseDuration(text string) time.Duration {
	var secs int64
	for _, part := range strings.Split(text, ":") {
		f, _ := strconv.ParseFloat(strings.TrimSpace(part), 64)
		secs = secs*60 + int64(f)
	}
	return time.Duration(secs) * time.Second
}
