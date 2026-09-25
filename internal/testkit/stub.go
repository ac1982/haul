// Package testkit is shared by haul's tests: a stub network that answers from registered routes, and real media
// files made by ffmpeg. Import it from _test files only.
package testkit

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ac1982/haul/internal/httpx"
)

// Response is one canned answer.
type Response struct {
	Status int
	Header http.Header
	Body   []byte
	// TruncateAt sends only this many body bytes while promising all of them (a dropped connection).
	TruncateAt int
	// Redirect answers 302 to this URL instead.
	Redirect string
	// NoLength leaves Content-Length unknown.
	NoLength bool
}

// JSON answers a JSON text.
func JSON(text string) Response {
	return Response{Header: http.Header{"Content-Type": {"application/json"}}, Body: []byte(text)}
}

// Text answers a body.
func Text(text string) Response { return Response{Body: []byte(text)} }

// Status answers an empty body with a status.
func Status(code int) Response { return Response{Status: code} }

// Ranged serves data honouring a Range header, the way a CDN does.
func Ranged(data []byte, req *http.Request) Response {
	from, to, ok := Range(req)
	if !ok {
		if req.Header.Get("Range") != "" {
			return Status(416)
		}
		return Response{Body: data}
	}
	if from >= int64(len(data)) {
		return Status(416)
	}
	end := int64(len(data)) - 1
	if to >= 0 && to < end {
		end = to
	}
	return Response{
		Status: 206,
		Header: http.Header{"Content-Range": {fmt.Sprintf("bytes %d-%d/%d", from, end, len(data))}},
		Body:   data[from : end+1],
	}
}

var rangeRe = regexp.MustCompile(`^bytes=(\d+)-(\d*)$`)

// Range is a request's byte range; to is -1 for an open range.
func Range(req *http.Request) (from, to int64, ok bool) {
	m := rangeRe.FindStringSubmatch(req.Header.Get("Range"))
	if m == nil {
		return 0, 0, false
	}
	from, _ = strconv.ParseInt(m[1], 10, 64)
	to = -1
	if m[2] != "" {
		to, _ = strconv.ParseInt(m[2], 10, 64)
	}
	return from, to, true
}

// Handler answers a request.
type Handler func(*http.Request) Response

type route struct {
	specificity int
	seq         int
	match       func(string) bool
	respond     Handler
}

// Stub is a network that never leaves the process. Its routes and log are its own, so tests using separate stubs
// can run in parallel.
type Stub struct {
	mu     sync.Mutex
	routes []route
	log    []*http.Request
	seq    int
}

// NewStub is an empty stub network.
func NewStub() *Stub { return &Stub{} }

// Client is an httpx client over the stub.
func (s *Stub) Client() *httpx.Client { return httpx.New(s) }

// On answers URLs containing fragment. The longest matching fragment wins; among equals, the newest.
func (s *Stub) On(fragment string, h Handler) {
	s.OnFunc(len(fragment), func(u string) bool { return strings.Contains(u, fragment) }, h)
}

// OnFunc answers URLs match accepts, with a given specificity.
func (s *Stub) OnFunc(specificity int, match func(string) bool, h Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	s.routes = append(s.routes, route{specificity, s.seq, match, h})
}

// Requests are the requests so far whose URL contains fragment, in order.
func (s *Stub) Requests(fragment string) []*http.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*http.Request
	for _, r := range s.log {
		if strings.Contains(r.URL.String(), fragment) {
			out = append(out, r)
		}
	}
	return out
}

// Ranges are the Range headers of the requests whose URL contains fragment.
func (s *Stub) Ranges(fragment string) []string {
	var out []string
	for _, r := range s.Requests(fragment) {
		out = append(out, r.Header.Get("Range"))
	}
	return out
}

// RoundTrip answers from the routes; an unknown URL fails like an unreachable host.
func (s *Stub) RoundTrip(req *http.Request) (*http.Response, error) {
	u := req.URL.String()
	s.mu.Lock()
	s.log = append(s.log, req)
	var matches []route
	for _, r := range s.routes {
		if r.match(u) {
			matches = append(matches, r)
		}
	}
	s.mu.Unlock()
	if len(matches) == 0 {
		return nil, fmt.Errorf("stub: no route for %s", u)
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].specificity != matches[j].specificity {
			return matches[i].specificity > matches[j].specificity
		}
		return matches[i].seq > matches[j].seq
	})
	r := matches[0].respond(req)
	if r.Redirect != "" {
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": {r.Redirect}}, Body: http.NoBody, Request: req}, nil
	}
	status := r.Status
	if status == 0 {
		status = 200
	}
	header := r.Header.Clone()
	if header == nil {
		header = http.Header{}
	}
	body := r.Body
	length := int64(len(body))
	if r.TruncateAt > 0 && r.TruncateAt < len(body) {
		body = body[:r.TruncateAt]
	}
	if r.NoLength {
		length = -1
	} else {
		header.Set("Content-Length", strconv.FormatInt(length, 10))
	}
	resp := &http.Response{StatusCode: status, Header: header, ContentLength: length, Request: req, Proto: "HTTP/1.1", ProtoMajor: 1, ProtoMinor: 1}
	if req.Method == http.MethodHead {
		resp.Body = http.NoBody
	} else {
		resp.Body = &truncatedBody{r: bytes.NewReader(body), short: int64(len(body)) < length}
	}
	return resp, nil
}

// truncatedBody ends early with an unexpected EOF, as a dropped connection does.
type truncatedBody struct {
	r     *bytes.Reader
	short bool
}

func (b *truncatedBody) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	if err == io.EOF && b.short {
		return n, io.ErrUnexpectedEOF
	}
	return n, err
}

func (b *truncatedBody) Close() error { return nil }

// Pattern is n bytes with a pattern, so a misplaced range shows up as a mismatch.
func Pattern(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i*31 + i>>8)
	}
	return b
}
