package podcast

import (
	"regexp"
	"strings"

	"github.com/ac1982/haul/internal/format"
)

var (
	rssItemRe = regexp.MustCompile(`(?s)<item[\s>].*?</item>`)
	enclosure = regexp.MustCompile(`<enclosure\s[^>]*>`)
	itunesImg = regexp.MustCompile(`<itunes:image\s[^>]*>`)
	cdata     = regexp.MustCompile(`(?s)^\s*<!\[CDATA\[(.*?)\]\]>\s*$`)
	htmlTag   = regexp.MustCompile(`<[^>]+>`)
	// rssTags match the tags rssItem reads, compiled once: feeds can have thousands of items.
	rssTags = func() map[string]*regexp.Regexp {
		m := map[string]*regexp.Regexp{}
		for _, name := range []string{"title", "guid", "pubDate", "description", "itunes:duration", "itunes:summary", "itunes:author"} {
			q := regexp.QuoteMeta(name)
			m[name] = regexp.MustCompile(`(?s)<` + q + `(?:\s[^>]*)?>(.*?)</` + q + `>`)
		}
		return m
	}()
)

// rssEpisode is the feed item titled exactly one of titles (case and spacing aside). A near miss is no match: it
// would download another episode (a trailer, a bonus) under this one's name.
func rssEpisode(rss string, titles []string) (episode, bool) {
	wanted := map[string]bool{}
	for _, t := range titles {
		if n := normalize(t); n != "" {
			wanted[n] = true
		}
	}
	for _, item := range rssItemRe.FindAllString(rss, -1) {
		// The title alone decides; only a match is parsed further.
		title, ok := rssTag("title", item)
		if !ok || !wanted[normalize(title)] {
			continue
		}
		if ep, ok := rssItem(item); ok {
			return ep, true
		}
	}
	return episode{}, false
}

// rssTag is the text of a tag in an item, CDATA unwrapped and entities decoded; ok is false when the tag is absent.
func rssTag(name, item string) (string, bool) {
	m := rssTags[name].FindStringSubmatch(item)
	if m == nil {
		return "", false
	}
	text := cdata.ReplaceAllString(m[1], "$1")
	return strings.TrimSpace(format.UnescapeEntities(text)), true
}

func rssItem(item string) (episode, bool) {
	tag := func(name string) string { s, _ := rssTag(name, item); return s }
	encl := enclosure.FindString(item)
	mediaURL := attr("url", encl)
	if mediaURL == "" {
		return episode{}, false
	}
	mediaType := attr("type", encl)
	ep := episode{
		id:        tag("guid"),
		title:     tag("title"),
		mediaURL:  mediaURL,
		mediaType: mediaType,
		duration:  parseDuration(tag("itunes:duration")),
		published: parseRFC822Date(tag("pubDate")),
		cover:     attr("href", itunesImg.FindString(item)),
		author:    tag("itunes:author"),
		video:     strings.HasPrefix(mediaType, "video/"),
	}
	if ep.id == "" {
		ep.id = mediaURL
	}
	// Show notes are HTML; the text is kept.
	notes, ok := rssTag("itunes:summary", item)
	if !ok {
		notes = tag("description")
	}
	ep.description = strings.TrimSpace(htmlTag.ReplaceAllString(notes, ""))
	return ep, true
}
