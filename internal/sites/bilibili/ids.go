package bilibili

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// kind is what a link points at.
type kind int

const (
	video kind = iota
	episode
	cheese // a course
	space  // a user's uploads
	collection
	series
	favorites
)

// mediaID is a resolved link: an aid, an episode id, a user id, a list id; mid is the owner of a favorites list.
type mediaID struct {
	kind kind
	id   string
	mid  string
}

func (m mediaID) String() string {
	switch m.kind {
	case episode:
		return "ep:" + m.id
	case cheese:
		return "cheese:" + m.id
	case space:
		return "mid:" + m.id
	case collection:
		return "collection:" + m.id
	case series:
		return "series:" + m.id
	case favorites:
		return "favorites:" + m.id + ":" + m.mid
	}
	return "av" + m.id
}

// pgc: bangumi and courses have their own playurl endpoints and parameters.
func (k kind) pgc() bool { return k == episode || k == cheese }

var (
	bareID     = regexp.MustCompile(`(?i)^(bv1[0-9a-z]{9}|av\d+|ep\d+|ss\d+|md\d+)$`)
	bareCheese = regexp.MustCompile(`^cheese/(ep|ss)\d+$`)
	siteHosts  = []string{"bilibili.com", "b23.tv", "bilibili.tv", "biliintl.com"}
)

// matchLink recognises a bilibili link or bare id, and returns it in the form resolveID takes.
func matchLink(link string) (string, bool) {
	text := strings.TrimSpace(link)
	if bareID.MatchString(text) || bareCheese.MatchString(text) {
		return text, true
	}
	if !strings.HasPrefix(text, "http") {
		text = "https://" + text
	}
	u, err := url.Parse(text)
	if err != nil {
		return "", false
	}
	host := strings.ToLower(u.Hostname())
	for _, h := range siteHosts {
		if host == h || strings.HasSuffix(host, "."+h) {
			return text, true
		}
	}
	return "", false
}

// parseInt reads an unsigned decimal id.
func parseInt(s string) (int64, bool) {
	if s == "" || strings.TrimLeft(s, "0123456789") != "" {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
