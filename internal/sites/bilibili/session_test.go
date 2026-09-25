package bilibili

import (
	"os"
	"testing"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/storage"
)

// Tests that write logins to the config directory are not parallel: they run before the parallel tests start and
// remove what they wrote.
func writeStored(t *testing.T, files map[string]string) {
	t.Helper()
	for name, text := range files {
		if err := storage.Write(name, text); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(storage.Path(name)) })
	}
}

func TestStoredCredentials(t *testing.T) {
	writeStored(t, map[string]string{storage.CookieFile: "SESSDATA=stored", storage.TVTokenFile: "access_token=tvtok", storage.AppTokenFile: "apptok"})
	s := &session{}
	loadStored(s, Web)
	if s.cookie != "SESSDATA=stored" || s.token != "" {
		t.Fatalf("web: %+v", s)
	}
	s = &session{}
	loadStored(s, TV)
	if s.token != "tvtok" {
		t.Fatalf("tv: %+v", s)
	}
	s = &session{}
	loadStored(s, App)
	if s.token != "apptok" {
		t.Fatalf("app: %+v", s)
	}
	// Explicit options win.
	s = &session{cookie: "SESSDATA=given", token: "given"}
	loadStored(s, App)
	if s.cookie != "SESSDATA=given" || s.token != "given" {
		t.Fatalf("explicit: %+v", s)
	}
	os.Remove(storage.Path(storage.AppTokenFile))
	s = &session{}
	loadStored(s, App)
	if s.token != "tvtok" {
		t.Fatalf("app from the TV token: %+v", s)
	}
}

func TestSessionChecksTheLoginOnce(t *testing.T) {
	t.Parallel()
	e, stub := stubbed(func(o *Options) { o.Cookie = "SESSDATA=x"; o.Token = "access_token=abc"; o.UserAgent = "Custom" })
	stub.On("x/web-interface/nav", answer(`{"code":0,"data":{"isLogin":true,"wbi_img":{"img_url":"https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b84077c.png","sub_url":"https://i0.hdslb.com/bfs/wbi/4932caff0ff746eab6f01bf08b70ac45.png"}}}`))
	s := e.session(ctx)
	if e.session(ctx) != s || len(stub.Requests("nav")) != 1 {
		t.Fatal("session set up twice")
	}
	if s.loggedIn == nil || !*s.loggedIn || s.currentWBI() != "ea1db124af3c7062474693fa704f4ff8" || s.token != "abc" || s.userAgent != "Custom" {
		t.Fatalf("%+v", s)
	}
	if h := stub.Requests("nav")[0].Header; h.Get("Cookie") != "SESSDATA=x" || h.Get("User-Agent") != "Custom" || h.Get("Referer") != "https://www.bilibili.com/" {
		t.Fatalf("%v", h)
	}
	if key, err := e.wbiKey(ctx, s); err != nil || key != s.wbi || len(stub.Requests("nav")) != 1 {
		t.Fatal("WBI key fetched again")
	}
}

func TestNoLoginCheckForTVIntlOrProxies(t *testing.T) {
	t.Parallel()
	for _, configure := range []func(*Options){
		func(o *Options) { o.API = TV },
		func(o *Options) { o.API = Intl },
		func(o *Options) { o.Area = "hk" },
	} {
		e, stub := stubbed(configure)
		if s := e.session(ctx); s.loggedIn != nil || len(stub.Requests("")) != 0 {
			t.Fatalf("%+v, %d requests", s, len(stub.Requests("")))
		}
	}
}

func TestAPIHeaders(t *testing.T) {
	t.Parallel()
	s := plainSession()
	h := s.apiHeader("https://www.bilibili.com/bangumi/play/ep5")
	if h.Get("Cookie") != ";CURRENT_FNVAL=4048;" || h.Get("Referer") != "" || h.Get("User-Agent") != "UA" || h.Get("Cache-Control") != "no-cache" {
		t.Fatalf("%v", h)
	}
	s.cookie = "SESSDATA=x"
	if h := s.apiHeader("https://api.bilibili.com/x/web-interface/view?aid=1"); h.Get("Cookie") != "SESSDATA=x" || h.Get("Referer") != "https://www.bilibili.com/" {
		t.Fatalf("%v", h)
	}
	if h := s.apiHeader("https://api.bilibili.tv/intl/x"); h.Get("sec-ch-ua") == "" {
		t.Fatalf("%v", h)
	}
}

func TestParseAPI(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]API{"": Web, "WEB": Web, " tv ": TV, "app": App, "Intl": Intl} {
		if got, err := ParseAPI(in); err != nil || got != want {
			t.Errorf("%q → %v %v", in, got, err)
		}
	}
	if _, err := ParseAPI("ios"); !errs.Is(err, errs.Input) {
		t.Fatal(err)
	}
	if App.Label() != "APP" {
		t.Fatal(App.Label())
	}
	e := New(nil, Options{})
	if e.opts.API != Web || e.opts.Host != defaultHost || e.opts.EpHost != defaultHost || e.opts.TVHost != defaultTVHost {
		t.Fatalf("%+v", e.opts)
	}
	if d := DefaultOptions(); !d.ReplaceHost || !d.ForceHTTP || len(d.DanmakuFormats) != 2 || d.API != Web {
		t.Fatalf("%+v", d)
	}
}
