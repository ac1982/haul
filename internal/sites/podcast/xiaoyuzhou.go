package podcast

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/media"
)

// xiaoyuzhouHeader is what the pages and the audio CDN expect: a browser coming from the site.
func xiaoyuzhouHeader() http.Header {
	return httpx.Header("User-Agent", httpx.BrowserUserAgent, "Referer", "https://www.xiaoyuzhoufm.com/")
}

func (e *Extractor) resolveXiaoyuzhou(ctx context.Context, ref xiaoyuzhouRef) (*media.Item, error) {
	html, err := e.client.Text(ctx, ref.url(), httpx.Header("User-Agent", httpx.BrowserUserAgent))
	if err != nil {
		return nil, err
	}
	return parseXiaoyuzhouPage(html, ref)
}

var nextData = regexp.MustCompile(`(?s)<script[^>]*id="__NEXT_DATA__"[^>]*>(.*?)</script>`)

// parseXiaoyuzhouPage reads a page: an episode page gives that episode; a show page gives the episodes it lists
// (the latest ones).
func parseXiaoyuzhouPage(html string, ref xiaoyuzhouRef) (*media.Item, error) {
	var episodes []episode
	var showNode jsonv.Value
	if m := nextData.FindStringSubmatch(html); m != nil {
		if data, err := jsonv.ParseString(m[1]); err == nil {
			episodes, showNode = xiaoyuzhouData(data, ref)
		}
	}
	// Episodes nested in their show's page do not repeat the show.
	for i := range episodes {
		ep := &episodes[i]
		ep.show = cmp.Or(ep.show, showNode.Get("title").String())
		ep.author = cmp.Or(ep.author, showNode.Get("author").String())
		ep.cover = cmp.Or(ep.cover, xiaoyuzhouImage(showNode.Get("image")))
	}

	if ref.episode {
		for _, ep := range episodes {
			if ep.id == ref.id {
				return episodeItem(xiaoyuzhou, ep, xiaoyuzhouHeader()), nil
			}
		}
		// The page data moved: the Open Graph tags still name the audio.
		if audio := meta("og:audio", html); audio != "" {
			ep := episode{
				id:          ref.id,
				title:       cmp.Or(meta("og:title", html), ref.id),
				mediaURL:    audio,
				cover:       meta("og:image", html),
				description: meta("og:description", html),
				pageURL:     ref.url(),
			}
			return episodeItem(xiaoyuzhou, ep, xiaoyuzhouHeader()), nil
		}
		return nil, errs.New("No audio on the Xiaoyuzhou page (a paid episode, or the page changed; --debug shows more)")
	}

	var listed []episode
	for _, ep := range episodes {
		if ep.showID == "" || ep.showID == ref.id {
			listed = append(listed, ep)
		}
	}
	if len(listed) == 0 {
		return nil, errs.New("No episodes on the Xiaoyuzhou show page (the page may have changed; --debug shows more)")
	}
	console.Status(fmt.Sprintf("The show page lists its latest %d episodes; download older ones by their episode links", len(listed)))
	s := show{
		id:          ref.id,
		url:         ref.url(),
		title:       cmp.Or(showNode.Get("title").String(), listed[0].show),
		description: showNode.Get("description").String(),
		author:      cmp.Or(showNode.Get("author").String(), listed[0].author),
	}
	return showItem(xiaoyuzhou, s, listed, xiaoyuzhouHeader()), nil
}

// xiaoyuzhouData finds the episodes and the show in the page data by their shape rather than by path, so a
// reshuffled page still parses. The show is the object with the show id (for an episode page, the episode's show)
// that is not itself an episode.
func xiaoyuzhouData(data jsonv.Value, ref xiaoyuzhouRef) ([]episode, jsonv.Value) {
	var episodes []episode
	var showNode jsonv.Value
	seen := map[string]bool{}
	wantedShow := func() string {
		if !ref.episode {
			return ref.id
		}
		for _, ep := range episodes {
			if ep.id == ref.id {
				return ep.showID
			}
		}
		return ""
	}
	walk(data, func(node jsonv.Value) {
		if ep, ok := xiaoyuzhouEpisode(node); ok && !seen[ep.id] {
			seen[ep.id] = true
			episodes = append(episodes, ep)
		}
		if showNode.IsNull() && !node.Has("eid") {
			pid, _ := node.Get("pid").Str()
			if _, titled := node.Get("title").Str(); titled && pid != "" && pid == wantedShow() {
				showNode = node
			}
		}
	})
	return episodes, showNode
}

func xiaoyuzhouEpisode(n jsonv.Value) (episode, bool) {
	eid, _ := n.Get("eid").Str()
	title, titled := n.Get("title").Str()
	audio := firstString(n.Path("enclosure", "url"), n.Path("media", "source", "url"))
	if eid == "" || !titled || audio == "" {
		return episode{}, false
	}
	podcast := n.Get("podcast")
	return episode{
		id:          eid,
		title:       strings.TrimSpace(title),
		mediaURL:    audio,
		mediaType:   n.Path("media", "mimeType").String(),
		size:        n.Path("media", "size").Int64Or(0),
		duration:    seconds(n.Get("duration").Int64Or(0)),
		published:   parseISODate(n.Get("pubDate").String()),
		cover:       cmp.Or(xiaoyuzhouImage(n.Get("image")), xiaoyuzhouImage(podcast.Get("image"))),
		description: n.Get("description").String(),
		show:        podcast.Get("title").String(),
		author:      podcast.Get("author").String(),
		showID:      firstString(n.Get("pid"), podcast.Get("pid")),
		pageURL:     "https://www.xiaoyuzhoufm.com/episode/" + eid,
	}, true
}

func xiaoyuzhouImage(image jsonv.Value) string {
	return firstString(image.Get("picUrl"), image.Get("largePicUrl"), image.Get("middlePicUrl"))
}
