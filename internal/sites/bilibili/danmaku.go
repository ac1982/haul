package bilibili

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/media"
)

// danmakuMode is where a comment shows.
type danmakuMode int

const (
	scroll danmakuMode = iota
	top
	bottom
)

// comment is one danmaku.
type comment struct {
	seconds float64
	mode    danmakuMode
	// color is RRGGBB, upper-case hex.
	color   string
	content string
}

// danmaku is the sidecar that saves a page's comments next to the output: the XML as bilibili serves it and/or an
// ASS subtitle file players can overlay.
func (e *Extractor) danmaku(s *session, cid string) media.Sidecar {
	wantXML, wantASS := slices.Contains(e.opts.DanmakuFormats, "xml"), slices.Contains(e.opts.DanmakuFormats, "ass")
	return media.Sidecar{Kind: "danmaku", Write: func(ctx context.Context, base string) ([]string, error) {
		data, err := e.fetchDanmaku(ctx, s, "https://comment.bilibili.com/"+cid+".xml")
		if err != nil {
			return nil, err
		}
		comments, err := parseDanmaku(data)
		if err != nil {
			console.Warn("Could not parse the danmaku XML")
			console.Debugf("Danmaku XML: %v", err)
			return nil, nil
		}
		if len(comments) == 0 {
			console.Status("Danmaku  none")
			return nil, nil
		}
		var files, exts []string
		if wantASS {
			if err := os.WriteFile(base+".ass", []byte(danmakuASS(comments)), 0o644); err != nil {
				return files, err
			}
			files, exts = append(files, base+".ass"), append(exts, ".ass")
		}
		if wantXML {
			if err := os.WriteFile(base+".xml", data, 0o644); err != nil {
				return files, err
			}
			files, exts = append(files, base+".xml"), append(exts, ".xml")
		}
		console.Status(fmt.Sprintf("Danmaku  %d comments  →  %s", len(comments), strings.Join(exts, ", ")))
		return files, nil
	}}
}

// fetchDanmaku downloads the comment XML, which the server may send deflated whatever was asked for.
func (e *Extractor) fetchDanmaku(ctx context.Context, s *session, url string) ([]byte, error) {
	resp, err := e.http.Do(ctx, "GET", url, s.mediaHeader(url), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.Header.Get("Content-Encoding") != "deflate" {
		return data, nil
	}
	// "deflate" is zlib-wrapped by the standard, raw by bilibili's habit; take whichever reads.
	if zr, err := zlib.NewReader(bytes.NewReader(data)); err == nil {
		if out, err := io.ReadAll(zr); err == nil {
			return out, nil
		}
	}
	return io.ReadAll(flate.NewReader(bytes.NewReader(data)))
}

// parseDanmaku reads the <d p="time,mode,size,color,timestamp,pool,uid,rowid">text</d> elements; malformed ones
// are skipped.
func parseDanmaku(data []byte) ([]comment, error) {
	var doc struct {
		D []struct {
			P    string `xml:"p,attr"`
			Text string `xml:",chardata"`
		} `xml:"d"`
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	var out []comment
	for _, d := range doc.D {
		parts := strings.Split(d.P, ",")
		if len(parts) < 8 {
			continue
		}
		c := comment{content: d.Text, mode: scroll}
		c.seconds, _ = strconv.ParseFloat(parts[0], 64)
		switch parts[1] {
		case "4":
			c.mode = bottom
		case "5":
			c.mode = top
		}
		color, err := strconv.ParseInt(parts[3], 10, 64)
		if err != nil {
			color = 0xFFFFFF
		}
		c.color = fmt.Sprintf("%06X", color)
		out = append(out, c)
	}
	return out, nil
}

// The ASS layout: a 1080p canvas, 40-pixel text, scrolling comments on screen for 8 s, fixed ones for 4 s, and
// only the top half used so the picture stays visible.
const (
	screenWidth      = 1920
	screenHeight     = 1080
	danmakuFontSize  = 40
	scrollSeconds    = 8.0
	staticSeconds    = 4.0
	protectedPercent = 50
)

// danmakuASS lays comments out as ASS dialogue lines.
func danmakuASS(comments []comment) string {
	var b strings.Builder
	b.WriteString("[Script Info]\n")
	b.WriteString("Script Updated By: haul\n")
	b.WriteString("ScriptType: v4.00+\n")
	fmt.Fprintf(&b, "PlayResX: %d\n", screenWidth)
	fmt.Fprintf(&b, "PlayResY: %d\n", screenHeight)
	fmt.Fprintf(&b, "Aspect Ratio: %d:%d\n", screenWidth, screenHeight)
	b.WriteString("Collisions: Normal\n")
	b.WriteString("WrapStyle: 2\n")
	b.WriteString("ScaledBorderAndShadow: yes\n")
	b.WriteString("YCbCr Matrix: TV.601\n")
	b.WriteString("[V4+ Styles]\n")
	b.WriteString("Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding\n")
	fmt.Fprintf(&b, "Style: Danmaku, 黑体, %d, &H00FFFFFF, &H00FFFFFF, &H00000000, &H00000000, 0, 0, 0, 0, 100, 100, 0.00, 0.00, 1, 2, 0, 7, 0, 0, 0, 0\n", danmakuFontSize)
	b.WriteString("[Events]\n")
	b.WriteString("Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")

	sorted := slices.Clone(comments)
	slices.SortStableFunc(sorted, func(a, b comment) int {
		switch {
		case a.seconds < b.seconds:
			return -1
		case a.seconds > b.seconds:
			return 1
		}
		return 0
	})
	rows := newRows()
	for _, c := range sorted {
		length := utf8.RuneCountInString(c.content)
		y := rows.place(c.mode, c.seconds, length)
		if y < 0 {
			continue
		}
		end := c.seconds + staticSeconds
		var effect string
		switch c.mode {
		case bottom:
			effect = fmt.Sprintf(`\an8\pos(%d, %d)`, screenWidth/2, screenHeight-danmakuFontSize-y)
		case top:
			effect = fmt.Sprintf(`\an8\pos(%d, %d)`, screenWidth/2, y)
		default:
			end = c.seconds + scrollSeconds
			effect = fmt.Sprintf(`\move(%d, %d, %d, %d)`, screenWidth, y, -length*danmakuFontSize, y)
		}
		if c.color != "FFFFFF" {
			effect += `\c&H` + assColor(c.color) + "&"
		}
		fmt.Fprintf(&b, "Dialogue: 2,%s,%s,Danmaku,,0000,0000,0000,,{%s}%s\n", assTime(c.seconds), assTime(end), effect, c.content)
	}
	return b.String()
}

// assColor turns RRGGBB into BBGGRR, the byte order of ASS colour overrides.
func assColor(rgb string) string {
	if len(rgb) != 6 {
		return rgb
	}
	return rgb[4:6] + rgb[2:4] + rgb[0:2]
}

// assTime is h:mm:ss.cc.
func assTime(seconds float64) string {
	h := int(seconds) / 3600
	m := int(seconds) % 3600 / 60
	s := seconds - float64(h*3600+m*60)
	return fmt.Sprintf("%d:%02d:%05.2f", h, m, s)
}

// rows tracks when each row of each mode frees up, so comments do not pile onto each other.
type rows struct {
	free [3][]float64 // by danmakuMode
}

func newRows() *rows {
	n := screenHeight * protectedPercent / danmakuFontSize / 100
	var r rows
	for i := range r.free {
		r.free[i] = make([]float64, n)
	}
	return &r
}

// place is the y offset for a comment shown at time, or -1 when every row is busy.
func (r *rows) place(mode danmakuMode, time float64, length int) int {
	display := staticSeconds
	if mode == scroll {
		display = scrollSeconds * float64(length+5) * danmakuFontSize / (screenWidth + float64(length)*scrollSeconds)
	}
	queue := r.free[mode]
	for i := range queue {
		if time >= queue[i] {
			queue[i] = time + display
			return i * danmakuFontSize
		}
	}
	return -1
}
