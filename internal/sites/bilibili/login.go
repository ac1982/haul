package bilibili

import (
	"context"
	"fmt"
	"strings"

	"github.com/ac1982/haul/internal/browser"
	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/format"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/storage"
)

// Login logs in and stores the login: "qr" scans a code for the web account (cookie.txt), "tv" for the TV account
// (tv-token.txt), "browser" copies the login of Edge or Chrome.
func (e *Extractor) Login(ctx context.Context, req extract.LoginRequest) error {
	switch strings.ToLower(req.Method) {
	case "", "qr":
		return e.loginWeb(ctx)
	case "tv":
		return e.loginTV(ctx)
	case "browser":
		return e.loginBrowser(ctx, req.Browser, req.Profile)
	}
	return errs.NewInput("Unknown login method %q: use qr, tv or browser", req.Method)
}

// loginSession is a session with no stored login, for the login endpoints.
func (e *Extractor) loginSession() *session {
	ua := e.opts.UserAgent
	if ua == "" {
		ua = httpx.BrowserUserAgent
	}
	return &session{userAgent: ua, host: defaultHost, epHost: defaultHost, tvHost: defaultTVHost}
}

func saved(what, file string) {
	console.Success("Logged in  " + console.CurrentStyle().Dim(what+" saved to "+console.PrettyPath(storage.Path(file))))
}

var errExpired = errs.NewAuth("The QR code expired; run the login again")

// loginWeb is the web QR login: the confirmed login's redirect carries the cookies in its query.
func (e *Extractor) loginWeb(ctx context.Context) error {
	s := e.loginSession()
	console.Status("Requesting a login QR code")
	gen, err := e.getJSON(ctx, s, "https://passport.bilibili.com/x/passport-login/web/qrcode/generate?source=main-fe-header")
	if err != nil {
		return err
	}
	data, err := field(gen, "data")
	if err != nil {
		return err
	}
	link := data.Get("url").String()
	key := format.QueryValue("qrcode_key", link)
	remove, err := showQR(link, e.qrPath)
	if err != nil {
		return err
	}
	defer remove()

	scanned := false
	for {
		if err := sleep(ctx, e.pollEvery); err != nil {
			return err
		}
		poll, err := e.getJSON(ctx, s, "https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key="+key+"&source=main-fe-header")
		if err != nil {
			return err
		}
		switch poll.Path("data", "code").IntOr(-1) {
		case 86038:
			return errExpired
		case 86101:
			continue
		case 86090:
			if !scanned {
				console.Status("Scanned; confirm on your phone")
				scanned = true
			}
			continue
		}
		redirect := poll.Path("data", "url").String()
		console.Debugf("SESSDATA=%s", format.QueryValue("SESSDATA", redirect))
		_, query, _ := strings.Cut(redirect, "?")
		if query == "" {
			return errs.NewAuth("The login was not confirmed: %s", poll.Path("data", "message").String())
		}
		// Commas would break the Cookie header, so they stay escaped.
		cookie := strings.NewReplacer("&", ";", ",", "%2C").Replace(query)
		if err := storage.Write(storage.CookieFile, cookie); err != nil {
			return err
		}
		saved("cookie", storage.CookieFile)
		return nil
	}
}

// tvLoginParams is the TV client's login form, in the (sorted) order the signature is computed over, signed.
func tvLoginParams() [][2]string {
	deviceID := format.RandomString(20)
	buvid := format.RandomString(37)
	fingerprint := format.CompactTimestamp() + format.RandomString(45)
	p := [][2]string{
		{"appkey", tvAppKey}, {"auth_code", ""}, {"bili_local_id", deviceID}, {"build", "102801"}, {"buvid", buvid},
		{"channel", "master"}, {"device", "OnePlus"}, {"device_id", deviceID}, {"device_name", "OnePlus7TPro"},
		{"device_platform", "Android10OnePlusHD1910"}, {"fingerprint", fingerprint}, {"guid", buvid},
		{"local_fingerprint", fingerprint}, {"local_id", buvid}, {"mobi_app", "android_tv_yst"}, {"networkstate", "wifi"},
		{"platform", "android"}, {"sys_ver", "29"}, {"ts", format.UnixSeconds()},
	}
	return signTV(p)
}

// signTV appends the TV appkey signature over the fields (without any previous sign).
func signTV(fields [][2]string) [][2]string {
	var p [][2]string
	for _, f := range fields {
		if f[0] != "sign" {
			p = append(p, f)
		}
	}
	return append(p, [2]string{"sign", appSign(httpx.FormEncode(p), tvAppSecret)})
}

// loginTV is the TV QR login; it yields an access token for the TV and APP APIs.
func (e *Extractor) loginTV(ctx context.Context) error {
	s := e.loginSession()
	params := tvLoginParams()
	console.Status("Requesting a login QR code")
	body, err := e.http.PostForm(ctx, "https://passport.snm0516.aisee.tv/x/passport-tv-login/qrcode/auth_code", s.apiHeader(""), params)
	if err != nil {
		return err
	}
	auth, err := jsonv.Parse(body)
	if err != nil {
		return err
	}
	data, err := field(auth, "data")
	if err != nil {
		return err
	}
	remove, err := showQR(data.Get("url").String(), e.qrPath)
	if err != nil {
		return err
	}
	defer remove()

	for i, f := range params {
		switch f[0] {
		case "auth_code":
			params[i][1] = data.Get("auth_code").String()
		case "ts":
			params[i][1] = format.UnixSeconds()
		}
	}
	params = signTV(params)
	for {
		if err := sleep(ctx, e.pollEvery); err != nil {
			return err
		}
		body, err := e.http.PostForm(ctx, "https://passport.bilibili.com/x/passport-tv-login/qrcode/poll", s.apiHeader(""), params)
		if err != nil {
			return err
		}
		poll, err := jsonv.Parse(body)
		if err != nil {
			return err
		}
		switch poll.Get("code").String() {
		case "86038":
			return errExpired
		case "86039":
			continue
		}
		token := poll.Path("data", "access_token").String()
		if token == "" {
			return errs.NewAuth("The TV login failed (code %s): %s", poll.Get("code").String(), poll.Get("message").String())
		}
		console.Debugf("access_token=%s", token)
		if err := storage.Write(storage.TVTokenFile, "access_token="+token); err != nil {
			return err
		}
		saved("token", storage.TVTokenFile)
		return nil
	}
}

// loginBrowser copies the bilibili cookies of a browser profile, after checking they are logged in.
func (e *Extractor) loginBrowser(ctx context.Context, name, profile string) error {
	b := browser.Browser(strings.ToLower(name))
	if profile == "" {
		profile = "Default"
	}
	console.Status(fmt.Sprintf("Reading the bilibili cookie of %s (%s)", b, profile))
	cookies, err := browser.Cookies(b, profile, "bilibili.com")
	if err != nil {
		return err
	}
	cookie := browser.Header(cookies)
	if cookieValue(cookie, "SESSDATA") == "" {
		return errs.NewAuth("The %s profile %q is not logged in to bilibili (no SESSDATA cookie); log in there first", b.Name(), profile)
	}
	s := e.loginSession()
	s.cookie = cookie
	console.Status("Checking the login")
	loggedIn, _, err := e.nav(ctx, s)
	if err != nil {
		return err
	}
	if !loggedIn {
		return errs.NewAuth("The browser's bilibili login has expired; log in again there and retry")
	}
	if err := storage.Write(storage.CookieFile, cookie); err != nil {
		return err
	}
	console.Success("Logged in  uid " + cookieValue(cookie, "DedeUserID") + "  " +
		console.CurrentStyle().Dim("cookie saved to "+console.PrettyPath(storage.Path(storage.CookieFile))))
	return nil
}

// cookieValue is one cookie of a Cookie header.
func cookieValue(header, name string) string {
	for _, part := range strings.Split(header, ";") {
		if k, v, ok := strings.Cut(strings.TrimSpace(part), "="); ok && k == name {
			return v
		}
	}
	return ""
}
