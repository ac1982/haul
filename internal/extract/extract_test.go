package extract

import (
	"context"
	"strings"
	"testing"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/media"
)

type site struct {
	name, prefix string
	loads        int
}

func (s *site) Info() Info { return Info{Site: media.Site(strings.ToLower(s.name)), Name: s.name} }
func (s *site) Match(link string) (string, bool) {
	return strings.TrimPrefix(link, s.prefix), strings.HasPrefix(link, s.prefix)
}
func (s *site) Resolve(context.Context, string) (*media.Item, error) { return nil, nil }
func (s *site) Formats(context.Context, *media.Item, *media.Entry) (*media.Formats, error) {
	s.loads++
	return &media.Formats{Raw: "loaded"}, nil
}

func TestRouter(t *testing.T) {
	a, b := &site{name: "Alpha", prefix: "a:"}, &site{name: "Beta", prefix: "b:"}
	r := NewRouter(a, b)
	x, link, err := r.Route("  b:123 ")
	if err != nil || x != b || link != "123" {
		t.Errorf("Route = %v %q %v", x, link, err)
	}
	if _, _, err := r.Route("c:1"); !errs.Is(err, errs.Input) || !strings.Contains(err.Error(), "Supported sites: Alpha, Beta") {
		t.Errorf("unsupported: %v", err)
	}
	if r.ByName("beta") != b || r.ByName("ALPHA") != a || r.ByName("gamma") != nil {
		t.Error("ByName")
	}
	inline := &media.Entry{Formats: &media.Formats{Raw: "inline"}}
	if f, _ := LoadFormats(context.Background(), a, nil, inline); f.Raw != "inline" || a.loads != 0 {
		t.Error("formats an entry carries are used as they are")
	}
	if f, _ := LoadFormats(context.Background(), a, nil, &media.Entry{}); f.Raw != "loaded" || a.loads != 1 {
		t.Error("formats are loaded on demand")
	}
}
