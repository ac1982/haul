// Package podcast reads podcast episodes from Xiaoyuzhou (小宇宙) and Apple Podcasts. Both lead to a plain audio
// file, so there is no yt-dlp here: Xiaoyuzhou pages carry their data as Next.js JSON, and Apple has the public
// iTunes lookup API plus each show's RSS feed.
package podcast

import (
	"context"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
)

// Extractor reads one podcast site.
type Extractor struct {
	client *httpx.Client
	site   media.Site
}

var _ extract.Extractor = (*Extractor)(nil)

const (
	xiaoyuzhou media.Site = "xiaoyuzhou"
	apple      media.Site = "applePodcasts"
)

// NewXiaoyuzhou reads Xiaoyuzhou episode and show pages.
func NewXiaoyuzhou(client *httpx.Client) *Extractor {
	return &Extractor{client: client, site: xiaoyuzhou}
}

// NewApple reads Apple Podcasts shows and episodes.
func NewApple(client *httpx.Client) *Extractor { return &Extractor{client: client, site: apple} }

// Info describes the site.
func (e *Extractor) Info() extract.Info {
	if e.site == xiaoyuzhou {
		return extract.Info{
			Site:       xiaoyuzhou,
			Name:       "Xiaoyuzhou",
			Links:      "xiaoyuzhoufm.com/episode/<id>, /podcast/<id>",
			Unit:       "episode",
			OwnerLabel: "host",
		}
	}
	return extract.Info{
		Site:       apple,
		Name:       "Apple Podcasts",
		Links:      "podcasts.apple.com/<cc>/podcast/<name>/id<show>[?i=<episode>]",
		Unit:       "episode",
		OwnerLabel: "host",
	}
}

// Match recognises the site's links and drops what does not select the content (slugs, share tokens).
func (e *Extractor) Match(link string) (string, bool) {
	if e.site == xiaoyuzhou {
		ref, ok := parseXiaoyuzhou(link)
		return ref.url(), ok
	}
	ref, ok := parseApple(link)
	return ref.url(), ok
}

// Resolve reads an episode or a show. Every entry comes with its formats.
func (e *Extractor) Resolve(ctx context.Context, link string) (*media.Item, error) {
	if e.site == xiaoyuzhou {
		ref, ok := parseXiaoyuzhou(link)
		if !ok {
			return nil, errs.NewInput("Not a Xiaoyuzhou link: %s", link)
		}
		return e.resolveXiaoyuzhou(ctx, ref)
	}
	ref, ok := parseApple(link)
	if !ok {
		return nil, errs.NewInput("Not an Apple Podcasts link: %s", link)
	}
	return e.resolveApple(ctx, ref)
}

// Formats returns what Resolve attached.
func (e *Extractor) Formats(_ context.Context, _ *media.Item, entry *media.Entry) (*media.Formats, error) {
	if entry.Formats == nil {
		return nil, errs.New("No formats for %s episode %s", e.Info().Name, entry.ID)
	}
	return entry.Formats, nil
}
