package bilibili

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/storage"
)

// session is the credentials and endpoints of a run, and what the login check found. Set up once; only the WBI
// key can still be filled in later (a space listing needs it even where the login check was skipped).
type session struct {
	cookie    string
	token     string
	userAgent string
	host      string
	epHost    string
	tvHost    string
	area      string
	// loggedIn is nil when the login was not checked (tv and intl APIs, proxies) or the check failed.
	loggedIn *bool

	mu  sync.Mutex
	wbi string
}

// biliPlus reports whether the main API goes through a proxy.
func (s *session) biliPlus() bool { return s.host != defaultHost }

// session sets up the session on first use: stored credentials, then the nav login check, which also yields the
// WBI key the web playurl needs. The check is skipped for the tv and intl APIs and behind proxies, where it means
// nothing; when it fails the run goes on logged out.
func (e *Extractor) session(ctx context.Context) *session {
	e.sessionOnce.Do(func() {
		s := &session{
			cookie:    e.opts.Cookie,
			token:     strings.ReplaceAll(e.opts.Token, "access_token=", ""),
			userAgent: e.opts.UserAgent,
			host:      e.opts.Host,
			epHost:    e.opts.EpHost,
			tvHost:    e.opts.TVHost,
			area:      e.opts.Area,
		}
		if s.userAgent == "" {
			s.userAgent = httpx.BrowserUserAgent
		}
		loadStored(s, e.opts.API)
		e.sess = s
		if e.opts.API == TV || e.opts.API == Intl || s.area != "" {
			return
		}
		console.Debugf("Checking the bilibili login")
		loggedIn, key, err := e.nav(ctx, s)
		if err != nil {
			console.Warn("Could not check the bilibili login; some streams may be missing")
			return
		}
		s.wbi = key
		s.loggedIn = &loggedIn
	})
	return e.sess
}

// loadStored fills in the stored login for the API in use; explicit options win.
func loadStored(s *session, api API) {
	if s.cookie == "" {
		if c := storage.Read(storage.CookieFile); c != "" {
			console.Debugf("Using the stored cookie: %s", storage.Path(storage.CookieFile))
			s.cookie = c
		}
	}
	var token string
	switch api {
	case TV:
		token = storage.Read(storage.TVTokenFile)
	case App:
		if token = storage.Read(storage.AppTokenFile); token == "" {
			token = storage.Read(storage.TVTokenFile)
		}
	}
	if s.token == "" && token != "" {
		console.Debugf("Using the stored %s token", api.Label())
		s.token = strings.ReplaceAll(token, "access_token=", "")
	}
}

// nav is x/web-interface/nav: whether the cookie is logged in, and the WBI key from its image keys.
func (e *Extractor) nav(ctx context.Context, s *session) (loggedIn bool, wbiKey string, err error) {
	j, err := e.getJSON(ctx, s, "https://api.bilibili.com/x/web-interface/nav")
	if err != nil {
		return false, "", err
	}
	data, err := field(j, "data")
	if err != nil {
		return false, "", err
	}
	img := data.Get("wbi_img")
	key := wbiMixinKey(imageKey(img.Get("img_url").String()), imageKey(img.Get("sub_url").String()))
	console.Debugf("wbi: %s", key)
	loggedIn, _ = data.Get("isLogin").Bool()
	return loggedIn, key, nil
}

// wbiKey is the WBI key, asking nav for it when the login check did not run.
func (e *Extractor) wbiKey(ctx context.Context, s *session) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.wbi != "" {
		return s.wbi, nil
	}
	_, key, err := e.nav(ctx, s)
	if err != nil {
		return "", err
	}
	s.wbi = key
	return key, nil
}

// currentWBI is the WBI key as it stands, possibly "" (then the signature is merely wrong, as bilibili tolerates).
func (s *session) currentWBI() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.wbi
}

// apiHeader is what an API call sends: the user agent, the cookie (episode and season endpoints also need
// CURRENT_FNVAL to answer with DASH), and the Referer api.bilibili.com checks.
func (s *session) apiHeader(url string) http.Header {
	h := http.Header{}
	h.Set("User-Agent", s.userAgent)
	h.Set("Cache-Control", "no-cache")
	cookie := s.cookie
	if strings.Contains(url, "/ep") || strings.Contains(url, "/ss") {
		cookie += ";CURRENT_FNVAL=4048;"
	}
	if cookie != "" {
		h.Set("Cookie", cookie)
	}
	if strings.Contains(url, "api.bilibili.com") {
		h.Set("Referer", "https://www.bilibili.com/")
	}
	if strings.Contains(url, "api.bilibili.tv") {
		h.Set("sec-ch-ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	}
	return h
}

// mediaHeader is what a stream or file download sends. The CDN wants bilibili's Referer, except for the
// Android and TV clients' URLs, which it refuses with one.
func (s *session) mediaHeader(url string) http.Header {
	h := http.Header{}
	if !strings.Contains(url, "platform=android") { // also platform=android_tv_yst
		h.Set("Referer", "https://www.bilibili.com")
	}
	h.Set("User-Agent", "Mozilla/5.0")
	if s.cookie != "" {
		h.Set("Cookie", s.cookie)
	}
	return h
}

func (e *Extractor) getJSON(ctx context.Context, s *session, url string) (jsonv.Value, error) {
	return e.http.JSON(ctx, url, s.apiHeader(url))
}

func (e *Extractor) getText(ctx context.Context, s *session, url string) (string, error) {
	return e.http.Text(ctx, url, s.apiHeader(url))
}

// finalURL is where a link's redirects end.
func (e *Extractor) finalURL(ctx context.Context, s *session, url string) (string, error) {
	h := http.Header{}
	h.Set("User-Agent", s.userAgent)
	h.Set("Cache-Control", "no-cache")
	location, err := e.http.FinalURL(ctx, url, h)
	if err == nil {
		console.Debugf("Location: %s", location)
	}
	return location, err
}
