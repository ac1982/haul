// Package bilibili reads bilibili: videos, bangumi, courses, lists and uploads, through the web, TV, Android app
// (gRPC) and international playurl APIs, with bilibili's CDN quirks, subtitles, chapters, dubbing and danmaku.
package bilibili

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
)

// Site is bilibili's site id.
const Site media.Site = "bilibili"

// API is the playurl backend asked for streams.
type API string

const (
	Web  API = "web"
	TV   API = "tv"
	App  API = "app"
	Intl API = "intl"
)

// Label is the API as file-name templates show it: WEB, TV, APP, INTL.
func (a API) Label() string { return strings.ToUpper(string(a)) }

// ParseAPI reads an --api value.
func ParseAPI(s string) (API, error) {
	switch a := API(strings.ToLower(strings.TrimSpace(s))); a {
	case Web, TV, App, Intl:
		return a, nil
	case "":
		return Web, nil
	}
	return "", errs.NewInput("Unknown bilibili API %q: use web, tv, app or intl", s)
}

// Default endpoint hosts; anything else in Host is a BiliPlus-style proxy.
const (
	defaultHost   = "api.bilibili.com"
	defaultTVHost = "api.snm0516.aisee.tv"
	// backupHost is a known-good CDN mirror that PCDN, overseas and (by default) all streams are moved to.
	backupHost = "upos-sz-mirrorcoso1.bilivideo.com"
)

// Options configure the extractor.
type Options struct {
	API API
	// Cookie (web login) and Token (TV / APP access key); empty means the stored ones.
	Cookie string
	Token  string
	// UserAgent for API calls; empty means a desktop browser's.
	UserAgent string
	// Host, EpHost (for the bangumi season endpoint) and TVHost can point at BiliPlus-style proxies; Area
	// (hk, tw, th…) is the region such a proxy is asked for.
	Host   string
	EpHost string
	TVHost string
	Area   string
	// UposHost replaces the host of every stream.
	UposHost string
	// ReplaceHost moves every stream to the backup mirror, unless UposHost is set.
	ReplaceHost bool
	// AllowPCDN keeps streams on PCDN (ip:port) hosts, which are often slow or unreachable.
	AllowPCDN bool
	// ForceHTTP fetches streams over plain HTTP (except the *.mcdn.bilivideo.cn:port hosts, which only speak what
	// they advertise); bilibili's CDN is faster that way.
	ForceHTTP bool
	// DanmakuFormats are the danmaku files written: "xml", "ass".
	DanmakuFormats []string
	// PreferredCodec (HEVC, AVC, AV1) is asked of the APP API; "" is its default.
	PreferredCodec string
}

// DefaultOptions are the options of a plain run.
func DefaultOptions() Options {
	return Options{
		API:            Web,
		Host:           defaultHost,
		EpHost:         defaultHost,
		TVHost:         defaultTVHost,
		ReplaceHost:    true,
		ForceHTTP:      true,
		DanmakuFormats: []string{"xml", "ass"},
	}
}

// Extractor reads bilibili. It is safe for concurrent use.
type Extractor struct {
	http *httpx.Client
	opts Options

	// The session (credentials, WBI key, login state) is set up once, on first use.
	sessionOnce sync.Once
	sess        *session

	// pollEvery is how often a QR login asks whether the code was scanned.
	pollEvery time.Duration
	// qrPath is where the login QR code is saved as a PNG while it is shown.
	qrPath string
}

var (
	_ extract.Extractor     = (*Extractor)(nil)
	_ extract.Authenticator = (*Extractor)(nil)
)

// New is an extractor using client; empty hosts in opts fall back to bilibili's own.
func New(client *httpx.Client, opts Options) *Extractor {
	if opts.API == "" {
		opts.API = Web
	}
	if opts.Host == "" {
		opts.Host = defaultHost
	}
	if opts.EpHost == "" {
		opts.EpHost = defaultHost
	}
	if opts.TVHost == "" {
		opts.TVHost = defaultTVHost
	}
	return &Extractor{http: client, opts: opts, pollEvery: time.Second, qrPath: "qrcode.png"}
}

// Info describes bilibili.
func (e *Extractor) Info() extract.Info {
	return extract.Info{
		Site:       Site,
		Name:       "bilibili",
		Links:      "bilibili.com (video, bangumi, course, space, list), b23.tv, BV…, ep…",
		Unit:       "page",
		OwnerLabel: "uploader",
	}
}

// Match takes bilibili links (bilibili.com, b23.tv, bilibili.tv, biliintl.com) and bare ids (BV…, av…, ep…, ss…,
// md…, cheese/ep…). Links without a scheme get https://.
func (e *Extractor) Match(link string) (string, bool) { return matchLink(link) }

// Resolve reads what a link points at.
func (e *Extractor) Resolve(ctx context.Context, link string) (*media.Item, error) {
	s := e.session(ctx)
	link = strings.TrimSpace(link)
	id, resolved, err := e.resolveID(ctx, s, link)
	if err != nil {
		return nil, err
	}
	id, in, err := e.fetchInfo(ctx, s, id)
	if err != nil {
		return nil, err
	}
	api := e.opts.API
	if in.interactive && api == TV {
		console.Warn("Interactive videos are not available through the TV API; using the web API")
		api = Web
	}
	item := buildItem(link, resolved, id, in, api)
	item.LoggedIn = s.loggedIn
	return item, nil
}

// Formats loads the streams, subtitles, chapters, dubbing and danmaku of an entry.
func (e *Extractor) Formats(ctx context.Context, item *media.Item, entry *media.Entry) (*media.Formats, error) {
	if entry.Formats != nil {
		return entry.Formats, nil
	}
	r, ok := entry.Ref.(*ref)
	if !ok {
		return nil, errs.New("Not a bilibili entry: %s", entry.ID)
	}
	s := e.session(ctx)
	f, err := e.streams(ctx, s, r)
	if err != nil {
		return nil, err
	}
	if len(f.Video) == 0 && len(f.Audio) == 0 {
		return nil, noStreamsError(entry.Title, f.Raw)
	}
	player := e.playerInfo(ctx, s, r)
	if points := viewPoints(player); len(points) > 0 {
		f.Chapters = points
	}
	f.Subtitles = e.subtitles(ctx, s, r, player)
	f.Sidecars = []media.Sidecar{e.danmaku(s, r.cid)}
	return f, nil
}
