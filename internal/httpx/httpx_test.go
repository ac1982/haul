package httpx_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/testkit"
)

func TestClient(t *testing.T) {
	stub := testkit.NewStub()
	c := stub.Client()
	ctx := context.Background()
	stub.On("api.test/json", func(r *http.Request) testkit.Response {
		return testkit.JSON(`{"code":0,"ua":"` + r.Header.Get("User-Agent") + `"}`)
	})
	stub.On("api.test/missing", func(*http.Request) testkit.Response { return testkit.Status(404) })
	stub.On("short.test/x", func(*http.Request) testkit.Response { return testkit.Response{Redirect: "https://api.test/final?x=1"} })
	stub.On("api.test/final", func(*http.Request) testkit.Response { return testkit.Text("ok") })
	stub.On("api.test/form", func(r *http.Request) testkit.Response {
		r.ParseForm()
		return testkit.Text(r.Header.Get("Content-Type") + " " + r.PostForm.Get("b") + " " + r.Host)
	})

	j, err := c.JSON(ctx, "https://api.test/json", httpx.Header("User-Agent", "UA", "Referer", ""))
	if err != nil || j.Get("ua").String() != "UA" {
		t.Errorf("JSON = %v, %v", j, err)
	}
	var se *httpx.StatusError
	if _, err := c.Text(ctx, "https://api.test/missing", nil); !errors.As(err, &se) || se.Status != 404 {
		t.Errorf("404: %v", err)
	}
	if u, err := c.FinalURL(ctx, "https://short.test/x", nil); err != nil || u != "https://api.test/final?x=1" {
		t.Errorf("FinalURL = %q, %v", u, err)
	}
	body, err := c.PostForm(ctx, "https://api.test/form", httpx.Header("Host", "grpc.test"), [][2]string{{"a", "1 2"}, {"b", "x&y"}})
	if err != nil || string(body) != "application/x-www-form-urlencoded x&y grpc.test" {
		t.Errorf("PostForm = %q, %v", body, err)
	}
	if got := httpx.FormEncode([][2]string{{"z", "1"}, {"a", "a b/c"}}); got != "z=1&a=a%20b%2Fc" {
		t.Errorf("FormEncode = %q", got)
	}
	if _, err := c.Text(ctx, "://bad", nil); err == nil {
		t.Error("an invalid URL is an error")
	}
}
