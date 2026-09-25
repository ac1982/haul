// Package httpx is the HTTP client every extractor and the downloader share. It knows nothing about sites: callers
// pass the headers they need. Tests swap the transport for a stub.
package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
)

// BrowserUserAgent is a desktop Safari; some hosts turn away unknown clients.
const BrowserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15"

// Client wraps an http.Client.
type Client struct {
	HTTP *http.Client
}

// New is a client over transport (nil: the default transport, tuned for many parallel ranges).
func New(transport http.RoundTripper) *Client {
	if transport == nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConnsPerHost = 32
		t.ResponseHeaderTimeout = 60 * time.Second
		transport = t
	}
	return &Client{HTTP: &http.Client{Transport: transport}}
}

// Default is the client for real runs.
var Default = New(nil)

// Header builds a header from name / value pairs.
func Header(kv ...string) http.Header {
	h := http.Header{}
	for i := 0; i+1 < len(kv); i += 2 {
		if kv[i+1] != "" {
			h.Set(kv[i], kv[i+1])
		}
	}
	return h
}

// StatusError is a non-2xx answer.
type StatusError struct {
	Status int
	URL    string
}

func (e *StatusError) Error() string { return fmt.Sprintf("HTTP %d: %s", e.Status, e.URL) }

// Do sends a request built from method, URL, headers and body, and checks the status.
func (c *Client) Do(ctx context.Context, method, rawURL string, header http.Header, body []byte) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, r)
	if err != nil {
		return nil, errs.New("Invalid URL: %s", rawURL)
	}
	for k, v := range header {
		req.Header[k] = v
	}
	if host := header.Get("Host"); host != "" {
		req.Host = host
	}
	console.Debugf("%s %s", method, rawURL)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		resp.Body.Close()
		return nil, &StatusError{Status: resp.StatusCode, URL: rawURL}
	}
	return resp, nil
}

// Bytes GETs a body.
func (c *Client) Bytes(ctx context.Context, rawURL string, header http.Header) ([]byte, error) {
	resp, err := c.Do(ctx, http.MethodGet, rawURL, header, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Text GETs a body as text.
func (c *Client) Text(ctx context.Context, rawURL string, header http.Header) (string, error) {
	b, err := c.Bytes(ctx, rawURL, header)
	if err != nil {
		return "", err
	}
	console.Debugf("Response: %s", truncate(string(b), 4000))
	return string(b), nil
}

// JSON GETs and parses a body.
func (c *Client) JSON(ctx context.Context, rawURL string, header http.Header) (jsonv.Value, error) {
	b, err := c.Bytes(ctx, rawURL, header)
	if err != nil {
		return jsonv.Null, err
	}
	console.Debugf("Response: %s", truncate(string(b), 4000))
	return jsonv.Parse(b)
}

// Post sends a body and returns the answer's body.
func (c *Client) Post(ctx context.Context, rawURL string, header http.Header, body []byte) ([]byte, error) {
	resp, err := c.Do(ctx, http.MethodPost, rawURL, header, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// PostForm sends url-encoded fields in the order given.
func (c *Client) PostForm(ctx context.Context, rawURL string, header http.Header, fields [][2]string) ([]byte, error) {
	h := header.Clone()
	if h == nil {
		h = http.Header{}
	}
	h.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.Post(ctx, rawURL, h, []byte(FormEncode(fields)))
}

// FinalURL follows redirects with a HEAD and reports where they end.
func (c *Client) FinalURL(ctx context.Context, rawURL string, header http.Header) (string, error) {
	resp, err := c.Do(ctx, http.MethodHead, rawURL, header, nil)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	return resp.Request.URL.String(), nil
}

// FormEncode joins fields as a query, in order.
func FormEncode(fields [][2]string) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		parts[i] = escape(f[0]) + "=" + escape(f[1])
	}
	return strings.Join(parts, "&")
}

func escape(s string) string { return strings.ReplaceAll(url.QueryEscape(s), "+", "%20") }

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
