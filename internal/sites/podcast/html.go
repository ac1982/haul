package podcast

import (
	"regexp"
	"strings"

	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/jsonv"
)

var (
	metaTag   = regexp.MustCompile(`<meta\s[^>]*>`)
	attribute = regexp.MustCompile(`([\w:-]+)\s*=\s*(?:"([^"]*)"|'([^']*)')`)
)

// meta is the content of `<meta property="og:audio" content="…">` (or name=), attributes in either order.
func meta(property, html string) string {
	for _, tag := range metaTag.FindAllString(html, -1) {
		if attr("property", tag) != property && attr("name", tag) != property {
			continue
		}
		if content := attr("content", tag); content != "" {
			return content
		}
	}
	return ""
}

// attr is an attribute's value in a tag, entities decoded.
func attr(name, tag string) string {
	for _, m := range attribute.FindAllStringSubmatch(tag, -1) {
		if m[1] == name {
			// Only one of the two quote styles matched; the other group is empty.
			return format.UnescapeEntities(m[2] + m[3])
		}
	}
	return ""
}

// walk visits every object in the tree depth first, a parent before its children and keys in sorted order, so
// the result does not depend on map order.
func walk(node jsonv.Value, visit func(jsonv.Value)) {
	switch {
	case node.IsObject():
		visit(node)
		for _, k := range node.Keys() {
			walk(node.Get(k), visit)
		}
	case node.IsArray():
		for _, item := range node.Array() {
			walk(item, visit)
		}
	}
}

// normalize is a title for comparing: no invisible direction marks (Apple prefixes U+200E), single spaces, no case.
func normalize(text string) string {
	text = strings.NewReplacer("‎", "", "‏", "").Replace(text)
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}
