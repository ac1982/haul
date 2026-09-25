package bilibili

import (
	"bytes"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"rsc.io/qr"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/storage"
	"github.com/ac1982/haul/internal/testkit"
)

// sequence answers each call with the next text, repeating the last.
func sequence(texts ...string) testkit.Handler {
	var mu sync.Mutex
	n := 0
	return func(*http.Request) testkit.Response {
		mu.Lock()
		defer mu.Unlock()
		text := texts[min(n, len(texts)-1)]
		n++
		return testkit.JSON(text)
	}
}

// loginExtractor polls fast and keeps its QR code in a temporary directory.
func loginExtractor(t *testing.T) (*Extractor, *testkit.Stub, string) {
	e, stub := stubbed(nil)
	e.pollEvery = time.Millisecond
	e.qrPath = filepath.Join(t.TempDir(), "qrcode.png")
	t.Cleanup(func() {
		os.Remove(storage.Path(storage.CookieFile))
		os.Remove(storage.Path(storage.TVTokenFile))
	})
	return e, stub, e.qrPath
}

func TestWebQRLogin(t *testing.T) {
	e, stub, qrPath := loginExtractor(t)
	stub.On("qrcode/generate", answer(`{"code":0,"data":{"url":"https://account.bilibili.com/h5/account-h5/auth/scan-web?navhide=1&qrcode_key=KEY1&from=","qrcode_key":"KEY1"}}`))
	polls := sequence(`{"code":0,"data":{"code":86101}}`, `{"code":0,"data":{"code":86090}}`, `{"code":0,"data":{"code":86090}}`,
		`{"code":0,"data":{"code":0,"url":"https://passport.biligame.com/crossDomain?DedeUserID=42&SESSDATA=a,b&bili_jct=c&gourl=x"}}`)
	stub.On("qrcode/poll?qrcode_key=KEY1&", func(r *http.Request) testkit.Response {
		if _, err := os.Stat(qrPath); err != nil {
			t.Error("the QR code is not saved while it is shown")
		}
		return polls(r)
	})
	if err := e.Login(ctx, extract.LoginRequest{Method: "qr"}); err != nil {
		t.Fatal(err)
	}
	if got := storage.Read(storage.CookieFile); got != "DedeUserID=42;SESSDATA=a%2Cb;bili_jct=c;gourl=x" {
		t.Fatalf("cookie %q", got)
	}
	if n := len(stub.Requests("qrcode/poll")); n != 4 {
		t.Fatalf("%d polls", n)
	}
	if _, err := os.Stat(qrPath); err == nil {
		t.Fatal("QR code left behind")
	}
}

func TestExpiredQRCode(t *testing.T) {
	e, stub, _ := loginExtractor(t)
	stub.On("qrcode/generate", answer(`{"code":0,"data":{"url":"https://x.test/?qrcode_key=K"}}`))
	stub.On("qrcode/poll", answer(`{"code":0,"data":{"code":86038}}`))
	if err := e.Login(ctx, extract.LoginRequest{}); !errs.Is(err, errs.Auth) {
		t.Fatalf("err = %v", err)
	}
	if storage.Read(storage.CookieFile) != "" {
		t.Fatal("cookie stored")
	}
}

func TestTVQRLogin(t *testing.T) {
	e, stub, _ := loginExtractor(t)
	stub.On("passport-tv-login/qrcode/auth_code", answer(`{"code":0,"data":{"url":"https://passport.snm0516.aisee.tv/x/passport-tv-login/h5/qrcode/auth?auth_code=AC","auth_code":"AC"}}`))
	var mu sync.Mutex
	var forms []url.Values
	polls := sequence(`{"code":86039,"message":"not scanned"}`, `{"code":0,"data":{"access_token":"TOKEN","mid":42}}`)
	stub.On("passport-tv-login/qrcode/poll", func(r *http.Request) testkit.Response {
		body, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(body))
		mu.Lock()
		forms = append(forms, form)
		mu.Unlock()
		return polls(r)
	})
	if err := e.Login(ctx, extract.LoginRequest{Method: "tv"}); err != nil {
		t.Fatal(err)
	}
	if got := storage.Read(storage.TVTokenFile); got != "access_token=TOKEN" {
		t.Fatalf("token %q", got)
	}
	if len(forms) != 2 || forms[0].Get("auth_code") != "AC" || forms[0].Get("appkey") != tvAppKey {
		t.Fatalf("%v", forms)
	}
	// The poll form is signed over everything but the signature, in order.
	body := stub.Requests("qrcode/poll")[0]
	if body.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Fatal(body.Header)
	}
	auth := stub.Requests("qrcode/auth_code")[0]
	raw, _ := io.ReadAll(auth.Body)
	unsigned, sign, _ := strings.Cut(string(raw), "&sign=")
	if sign != appSign(unsigned, tvAppSecret) || !strings.HasPrefix(unsigned, "appkey="+tvAppKey+"&auth_code=&") {
		t.Fatalf("auth form %s", raw)
	}
}

func TestTVLoginRefused(t *testing.T) {
	e, stub, _ := loginExtractor(t)
	stub.On("passport-tv-login/qrcode/auth_code", answer(`{"code":0,"data":{"url":"https://x.test","auth_code":"AC"}}`))
	stub.On("passport-tv-login/qrcode/poll", answer(`{"code":-3,"message":"API校验密匙错误"}`))
	if err := e.Login(ctx, extract.LoginRequest{Method: "tv"}); !errs.Is(err, errs.Auth) || !strings.Contains(err.Error(), "-3") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoginMethodsAndBrowsers(t *testing.T) {
	t.Parallel()
	e := New(httpx.New(testkit.NewStub()), DefaultOptions())
	if err := e.Login(ctx, extract.LoginRequest{Method: "sms"}); !errs.Is(err, errs.Input) {
		t.Fatalf("err = %v", err)
	}
	if err := e.Login(ctx, extract.LoginRequest{Method: "browser", Browser: "firefox"}); !errs.Is(err, errs.Input) {
		t.Fatalf("err = %v", err)
	}
	if cookieValue("SESSDATA=a; DedeUserID=42; x", "DedeUserID") != "42" || cookieValue("a=1", "b") != "" {
		t.Fatal("cookieValue")
	}
}

func TestQRCodeDrawing(t *testing.T) {
	t.Parallel()
	code, err := qr.Encode("https://example.test/login?key=abc", qr.Q)
	if err != nil {
		t.Fatal(err)
	}
	lines := qrLines(code)
	if len(lines) != code.Size+4 {
		t.Fatalf("%d lines for size %d", len(lines), code.Size)
	}
	// The quiet zone is light; the top-left finder pattern starts dark.
	light, dark := "\x1b[47m  ", "\x1b[40m  "
	if lines[0] != strings.Repeat(light, code.Size+4)+"\x1b[0m" || !strings.HasPrefix(lines[2], light+light+dark) {
		t.Fatalf("%q", lines[2][:40])
	}
	path := filepath.Join(t.TempDir(), "qr.png")
	if err := writeQRPNG(code, path, 7); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	img, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	side := (code.Size + 8) * 7
	if img.Bounds().Dx() != side || img.Bounds().Dy() != side {
		t.Fatalf("%v", img.Bounds())
	}
	if r, _, _, _ := img.At(0, 0).RGBA(); r != 0xffff {
		t.Fatal("quiet zone is not white")
	}
	if r, _, _, _ := img.At(4*7, 4*7).RGBA(); r != 0 {
		t.Fatal("finder pattern is not black")
	}
}
