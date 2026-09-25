package podcast

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// lookupLimit is the most episodes the iTunes lookup returns.
const lookupLimit = 200

// appleHeader: some podcast hosts turn away unknown clients; a browser is always welcome.
func appleHeader() http.Header { return httpx.Header("User-Agent", httpx.BrowserUserAgent) }

func (e *Extractor) resolveApple(ctx context.Context, ref appleRef) (*media.Item, error) {
	lookupURL := fmt.Sprintf("https://itunes.apple.com/lookup?id=%s&country=%s&media=podcast&entity=podcastEpisode&limit=%d",
		ref.show, ref.country, lookupLimit)
	lookup, err := e.client.JSON(ctx, lookupURL, appleHeader())
	if err != nil {
		return nil, err
	}
	item, err := parseAppleLookup(lookup, ref)
	if err != nil || item != nil {
		return item, err
	}
	return e.olderAppleEpisode(ctx, ref, appleShow(lookup.Get("results").Array()))
}

// olderAppleEpisode finds an episode the lookup did not return (it lists only the latest 200) in the show's RSS
// feed, by the title on its Apple page.
func (e *Extractor) olderAppleEpisode(ctx context.Context, ref appleRef, showNode jsonv.Value) (*media.Item, error) {
	feed := showNode.Get("feedUrl").String()
	if feed == "" {
		return nil, errs.New("Apple Podcasts gave no RSS feed for this show")
	}
	console.Status(fmt.Sprintf("The episode is not among the latest %d; looking in the RSS feed", lookupLimit))
	var (
		wg               sync.WaitGroup
		page, rss        string
		pageErr, feedErr error
	)
	wg.Go(func() { page, pageErr = e.client.Text(ctx, ref.url(), appleHeader()) })
	wg.Go(func() { rss, feedErr = e.client.Text(ctx, feed, appleHeader()) })
	wg.Wait()
	if err := cmp.Or(pageErr, feedErr); err != nil {
		return nil, err
	}
	showName := showNode.Get("collectionName").String()
	titles := appleEpisodeTitles(page, showName)
	if len(titles) == 0 {
		return nil, errs.New("No episode title on the Apple Podcasts page")
	}
	ep, ok := rssEpisode(rss, titles)
	if !ok {
		return nil, errs.New("Episode \"%s\" is not in the show's RSS feed (the feed may keep only newer episodes)", titles[0])
	}
	ep.id = ref.episode
	ep.show = showName
	ep.author = cmp.Or(ep.author, showNode.Get("artistName").String())
	ep.showID = ref.show
	ep.cover = cmp.Or(ep.cover, artwork(showNode))
	ep.pageURL = ref.url()
	return episodeItem(apple, ep, appleHeader()), nil
}

// parseAppleLookup reads the iTunes lookup: the show, or the linked episode. It is nil, without an error, when
// the link names an episode the lookup did not return.
func parseAppleLookup(lookup jsonv.Value, ref appleRef) (*media.Item, error) {
	results := lookup.Get("results").Array()
	showNode := appleShow(results)
	if showNode.IsNull() && len(results) == 0 {
		return nil, errs.New("Apple Podcasts has no show id%s (the country in the link may be wrong)", ref.show)
	}
	var episodes []episode
	for _, r := range results {
		if r.Get("wrapperType").String() != "podcastEpisode" {
			continue
		}
		if ep, ok := appleEpisode(r, showNode); ok {
			episodes = append(episodes, ep)
		}
	}
	if ref.episode != "" {
		for _, ep := range episodes {
			if ep.id == ref.episode {
				return episodeItem(apple, ep, appleHeader()), nil
			}
		}
		return nil, nil
	}
	if len(episodes) == 0 {
		return nil, errs.New("Apple Podcasts returned no episodes for this show")
	}
	if appleListIsCut(results) {
		console.Warn(fmt.Sprintf("Apple returns only the latest %d episodes; download older ones by their episode links", len(episodes)))
	}
	s := show{
		id:     ref.show,
		url:    ref.url(),
		title:  cmp.Or(showNode.Get("collectionName").String(), episodes[0].show),
		author: cmp.Or(showNode.Get("artistName").String(), episodes[0].author),
	}
	return showItem(apple, s, episodes, appleHeader()), nil
}

// appleListIsCut: the lookup stops at 200 episodes; the show's own entry does not count.
func appleListIsCut(results []jsonv.Value) bool {
	n := 0
	for _, r := range results {
		if r.Get("wrapperType").String() == "podcastEpisode" {
			n++
		}
	}
	return n >= lookupLimit
}

// appleShow is the show's own entry in the lookup, or Null.
func appleShow(results []jsonv.Value) jsonv.Value {
	for _, r := range results {
		if r.Get("wrapperType").String() == "track" && r.Get("kind").String() == "podcast" {
			return r
		}
	}
	return jsonv.Null
}

func appleEpisode(e, showNode jsonv.Value) (episode, bool) {
	id, ok := e.Get("trackId").Int64()
	mediaURL := firstString(e.Get("episodeUrl"), e.Get("previewUrl"))
	if !ok || mediaURL == "" {
		return episode{}, false
	}
	ep := episode{
		id:          strconv.FormatInt(id, 10),
		title:       strings.TrimSpace(e.Get("trackName").String()),
		mediaURL:    mediaURL,
		mediaType:   e.Get("episodeFileExtension").String(),
		duration:    seconds(e.Get("trackTimeMillis").Int64Or(0) / 1000),
		published:   parseISODate(e.Get("releaseDate").String()),
		cover:       cmp.Or(artwork(e), artwork(showNode)),
		description: firstString(e.Get("description"), e.Get("shortDescription")),
		show:        cmp.Or(e.Get("collectionName").String(), showNode.Get("collectionName").String()),
		author:      cmp.Or(e.Get("artistName").String(), showNode.Get("artistName").String()),
		pageURL:     e.Get("trackViewUrl").String(),
		video:       e.Get("episodeContentType").String() == "video",
	}
	if show, ok := e.Get("collectionId").Int64(); ok {
		ep.showID = strconv.FormatInt(show, 10)
	}
	return ep, true
}

var artworkSize = regexp.MustCompile(`/\d+x\d+bb\.(jpg|png|webp)$`)

// artwork is the artwork at 1400 px: the lookup offers 600 px, and the image server renders any size the path
// asks for.
func artwork(node jsonv.Value) string {
	u := firstString(node.Get("artworkUrl600"), node.Get("artworkUrl160"))
	return artworkSize.ReplaceAllString(u, "/1400x1400bb.jpg")
}

var titleTag = regexp.MustCompile(`<title[^>]*>([^<]*)</title>`)

// appleEpisodeTitles are the episode's title as its Apple page gives it: `og:title`, and `<title>` without
// " - <show> - Apple Podcasts".
func appleEpisodeTitles(html, showName string) []string {
	var titles []string
	if og := meta("og:title", html); og != "" {
		titles = append(titles, og)
	}
	if m := titleTag.FindStringSubmatch(html); m != nil {
		title := normalize(format.UnescapeEntities(m[1]))
		suffixes := []string{" - apple podcasts", " on apple podcasts"}
		if showName != "" {
			suffixes = append(suffixes, " - "+normalize(showName))
		}
		for _, s := range suffixes {
			title = strings.TrimSuffix(title, s)
		}
		titles = append(titles, title)
	}
	var out []string
	for _, t := range titles {
		if normalize(t) != "" {
			out = append(out, t)
		}
	}
	return out
}
