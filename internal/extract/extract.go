// Package extract is the contract between haul and the sites it reads. Each site is an Extractor in its own
// package; the Router picks one by link. Everything site-specific — link forms, APIs, headers, CDN quirks, logins —
// stays behind this interface.
package extract

import (
	"context"
	"strings"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/media"
)

// Info describes a site for help texts and messages.
type Info struct {
	Site media.Site
	// Name is how people call it: YouTube, bilibili, Apple Podcasts.
	Name string
	// Links are the link forms it accepts, one line for --help.
	Links string
	// Unit is what one entry is called: video, page, episode.
	Unit string
	// OwnerLabel is how the uploader is introduced: channel, uploader, host.
	OwnerLabel string
	// Requires names external tools the site cannot work without (yt-dlp), with how to install them.
	Requires []Tool
}

// Tool is an external program an extractor needs.
type Tool struct {
	Name    string
	Install string
}

// Extractor reads one site.
type Extractor interface {
	Info() Info
	// Match reports whether the link belongs to the site, and the link in the form Resolve takes.
	Match(link string) (string, bool)
	// Resolve reads the item a matched link points at: its metadata and entries.
	Resolve(ctx context.Context, link string) (*media.Item, error)
	// Formats loads an entry's formats. Entries that came with Formats set are returned as they are.
	Formats(ctx context.Context, item *media.Item, entry *media.Entry) (*media.Formats, error)
}

// Router picks the extractor for a link.
type Router struct {
	extractors []Extractor
}

// NewRouter tries extractors in the order given.
func NewRouter(extractors ...Extractor) *Router { return &Router{extractors: extractors} }

// Extractors are the registered extractors, in order.
func (r *Router) Extractors() []Extractor { return r.extractors }

// Route finds the extractor for a link; an unrecognised link is an input error naming the supported sites.
func (r *Router) Route(link string) (Extractor, string, error) {
	link = strings.TrimSpace(link)
	for _, x := range r.extractors {
		if normalized, ok := x.Match(link); ok {
			return x, normalized, nil
		}
	}
	return nil, "", r.Unsupported(link)
}

// Unsupported is the error for a link no extractor takes.
func (r *Router) Unsupported(link string) error {
	names := make([]string, len(r.extractors))
	for i, x := range r.extractors {
		names[i] = x.Info().Name
	}
	return errs.NewInput("Unsupported link: %s\nSupported sites: %s (run `haul --help` for the link forms)", link, strings.Join(names, ", "))
}

// ByName finds an extractor by site id or name, case-insensitively.
func (r *Router) ByName(name string) Extractor {
	for _, x := range r.extractors {
		info := x.Info()
		if strings.EqualFold(string(info.Site), name) || strings.EqualFold(info.Name, name) {
			return x
		}
	}
	return nil
}

// LoadFormats is an entry's formats: what it carried, or what its extractor loads.
func LoadFormats(ctx context.Context, x Extractor, item *media.Item, entry *media.Entry) (*media.Formats, error) {
	if entry.Formats != nil {
		return entry.Formats, nil
	}
	return x.Formats(ctx, item, entry)
}
