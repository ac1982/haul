// Package fetch downloads media.Resources to files. It follows the policy each resource states — parallel byte
// ranges, closed ranges one at a time, or one stream — resumes what an earlier attempt left behind, and knows
// nothing about sites.
package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/media"
)

// Progress receives a transfer's progress: bytes so far and the total when known (≤ 0 unknown).
type Progress interface {
	Update(done, total int64)
	Finish(ok bool)
}

// Fetcher downloads a resource to dst.
type Fetcher interface {
	Fetch(ctx context.Context, res media.Resource, dst string, p Progress) error
}

// HTTP is the built-in downloader.
type HTTP struct {
	Client *httpx.Client
	// Parallel is how many ranges a Parallel resource is fetched over; 1 disables splitting.
	Parallel int
	// ChunkSize is the size of one range of a Parallel resource.
	ChunkSize int64
	// Attempts is how often a range or stream is tried before giving up.
	Attempts int
}

// NewHTTP is the downloader with its defaults: 16 connections, 20 MiB ranges, 3 attempts.
func NewHTTP(client *httpx.Client) *HTTP {
	return &HTTP{Client: client, Parallel: 16, ChunkSize: 20 << 20, Attempts: 3}
}

// errRangeIgnored: the server answered a range request with the whole file.
var errRangeIgnored = errors.New("range ignored")

type noProgress struct{}

func (noProgress) Update(int64, int64) {}
func (noProgress) Finish(bool)         {}

// Fetch downloads res to dst, replacing it.
func (h *HTTP) Fetch(ctx context.Context, res media.Resource, dst string, p Progress) (err error) {
	if p == nil {
		p = noProgress{}
	}
	if dir := filepath.Dir(dst); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	defer func() { p.Finish(err == nil) }()
	switch {
	case res.Policy == media.Sequential:
		return h.ranged(ctx, res, dst, p, true)
	case res.Policy == media.Whole || h.Parallel <= 1:
		return h.whole(ctx, res, dst, p)
	}
	err = h.ranged(ctx, res, dst, p, false)
	if errors.Is(err, errRangeIgnored) {
		console.Warn("The server refuses byte ranges; downloading over one connection")
		removeClips(dst)
		return h.whole(ctx, res, dst, p)
	}
	return err
}

// whole fetches one stream into dst.part, resuming it with an open range, then moves it into place.
func (h *HTTP) whole(ctx context.Context, res media.Resource, dst string, p Progress) error {
	part := dst + ".part"
	var err error
	for attempt := 1; attempt <= h.attempts(); attempt++ {
		if err = h.stream(ctx, res, part, 0, -1, false, p.Update); err == nil {
			return os.Rename(part, dst)
		}
		if ctx.Err() != nil || isFatal(err) {
			break
		}
		console.Debugf("Download failed, retrying (%d/%d): %v", attempt, h.attempts(), err)
	}
	return err
}

// Clip is one byte range [From, To] of a file; To is -1 for "to the end".
type Clip struct {
	Index    int
	From, To int64
}

// Clips covers a file of size bytes in ranges of chunk+1 bytes (bytes=a-(a+chunk)). The last one reads to the end,
// or ends at the last byte when closed.
func Clips(size, chunk int64, closed bool) []Clip {
	var clips []Clip
	for cursor := int64(0); cursor < size; {
		end := cursor + chunk
		if end >= size-1 {
			to := int64(-1)
			if closed {
				to = size - 1
			}
			return append(clips, Clip{len(clips), cursor, to})
		}
		clips = append(clips, Clip{len(clips), cursor, end})
		cursor = end + 1
	}
	return clips
}

// ranged fetches a file in clips, each into its own part file, then joins them. Sequential resources get closed
// ranges of at most MaxRange bytes, one at a time.
func (h *HTTP) ranged(ctx context.Context, res media.Resource, dst string, p Progress, sequential bool) error {
	size := res.Size
	if size <= 0 {
		var err error
		if size, err = h.probe(ctx, res, sequential); err != nil {
			return err
		}
	}
	console.Debugf("Size: %d bytes", size)
	if size <= 0 {
		if sequential {
			return errs.New("The server did not state the file size: %s", res.URL)
		}
		return errRangeIgnored
	}
	if info, err := os.Stat(dst); err == nil && info.Size() == size {
		console.Debugf("Already downloaded: %s", dst)
		return nil
	}
	parallel := h.Parallel
	clips := Clips(size, h.ChunkSize, false)
	if sequential {
		parallel = 1
		clips = Clips(size, res.MaxRange-1, true)
	}
	console.Debugf("Clips: %d", len(clips))

	var mu sync.Mutex
	done := make([]int64, len(clips))
	report := func(i int, n int64) {
		mu.Lock()
		done[i] = n
		var sum int64
		for _, d := range done {
			sum += d
		}
		mu.Unlock()
		p.Update(sum, size)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	sem := make(chan struct{}, parallel)
	errc := make(chan error, len(clips))
	var wg sync.WaitGroup
	for _, c := range clips {
		sem <- struct{}{}
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		go func(c Clip) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := h.clip(ctx, res, clipPath(dst, c.Index), c, func(n, _ int64) { report(c.Index, n) }); err != nil {
				errc <- err
				cancel()
			}
		}(c)
	}
	wg.Wait()
	close(errc)
	if err := firstError(errc); err != nil {
		return err
	}
	parts := make([]string, len(clips))
	for i := range clips {
		parts[i] = clipPath(dst, i)
	}
	if err := Join(parts, dst); err != nil {
		return err
	}
	for _, f := range parts {
		os.Remove(f)
	}
	return nil
}

func firstError(errc <-chan error) error {
	var first error
	for err := range errc {
		if first == nil || errors.Is(first, context.Canceled) {
			first = err
		}
	}
	return first
}

func (h *HTTP) clip(ctx context.Context, res media.Resource, path string, c Clip, update func(done, total int64)) error {
	var err error
	for attempt := 1; attempt <= h.attempts(); attempt++ {
		if err = h.stream(ctx, res, path, c.From, c.To, true, update); err == nil {
			return nil
		}
		if ctx.Err() != nil || isFatal(err) || errors.Is(err, errRangeIgnored) {
			break
		}
	}
	if errors.Is(err, errRangeIgnored) || ctx.Err() != nil {
		return err
	}
	return errs.New("Clip %d failed: %v", c.Index, err)
}

// probe asks for the size: the Content-Length of a plain GET, or for sequential servers the total of the
// Content-Range answering bytes=0-0 (the Content-Length of that probe is 1).
func (h *HTTP) probe(ctx context.Context, res media.Resource, sequential bool) (int64, error) {
	header := res.Header.Clone()
	if header == nil {
		header = http.Header{}
	}
	if sequential {
		header.Set("Range", "bytes=0-0")
	}
	resp, err := h.Client.Do(ctx, http.MethodGet, res.URL, header, nil)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	if sequential {
		_, total, _ := strings.Cut(resp.Header.Get("Content-Range"), "/")
		n, _ := strconv.ParseInt(total, 10, 64)
		return n, nil
	}
	return resp.ContentLength, nil
}

// stream fetches [from, to] of res into path, continuing after what path already holds. update gets bytes of this
// range so far and the range's size when known. In strict mode a server that ignores the range is an error;
// otherwise the file starts over.
func (h *HTTP) stream(ctx context.Context, res media.Resource, path string, from, to int64, strict bool, update func(done, total int64)) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	have, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}
	if to >= 0 && have == to-from+1 {
		update(have, have)
		return nil
	}
	header := res.Header.Clone()
	if header == nil {
		header = http.Header{}
	}
	rng := "bytes=" + strconv.FormatInt(from+have, 10) + "-"
	if to >= 0 {
		rng += strconv.FormatInt(to, 10)
	}
	header.Set("Range", rng)
	resp, err := h.Client.Do(ctx, http.MethodGet, res.URL, header, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK && (from+have > 0 || to >= 0) {
		if strict {
			return errRangeIgnored
		}
		// The whole file came back: start over.
		have = 0
		if err := f.Truncate(0); err != nil {
			return err
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}
	total := int64(-1)
	if resp.ContentLength >= 0 {
		total = have + resp.ContentLength
	}
	buf := make([]byte, 256<<10)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				return err
			}
			have += int64(n)
			update(have, total)
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return fmt.Errorf("incomplete download: %w", rerr)
		}
	}
	if total >= 0 && have != total {
		return errs.New("Incomplete download (%d/%d bytes)", have, total)
	}
	return f.Sync()
}

func (h *HTTP) attempts() int { return max(1, h.Attempts) }

// isFatal: errors another attempt will not fix (4xx answers).
func isFatal(err error) bool {
	var se *httpx.StatusError
	return errors.As(err, &se) && se.Status >= 400 && se.Status < 500 && se.Status != 408 && se.Status != 429
}

func clipPath(dst string, i int) string { return fmt.Sprintf("%s.part%05d", dst, i) }

func removeClips(dst string) {
	matches, _ := filepath.Glob(dst + ".part[0-9][0-9][0-9][0-9][0-9]")
	for _, m := range matches {
		os.Remove(m)
	}
}

// Join concatenates files into dst (a rename when there is one).
func Join(files []string, dst string) error {
	if len(files) == 1 {
		return os.Rename(files[0], dst)
	}
	tmp := dst + ".joining"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	for _, name := range files {
		in, err := os.Open(name)
		if err != nil {
			out.Close()
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		if err != nil {
			out.Close()
			return err
		}
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}
