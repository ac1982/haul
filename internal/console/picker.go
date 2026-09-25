package console

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/ac1982/haul/internal/errs"
	"golang.org/x/term"
)

// Key is one decoded key press.
type Key struct {
	Kind  KeyKind
	Digit int
}

// KeyKind names what a key does in the picker.
type KeyKind int

const (
	KeyOther KeyKind = iota
	KeyUp
	KeyDown
	KeyFirst
	KeyLast
	KeyEnter
	KeyCancel
	KeyDigit
)

// PickerHint is the help line under a chooser.
const PickerHint = "↑↓ move   Enter choose   digits jump   q cancel"

// ParseKey decodes one key press from raw terminal bytes.
func ParseKey(b []byte) Key {
	if len(b) == 0 {
		return Key{Kind: KeyOther}
	}
	switch string(b) {
	case "\x1b[A", "\x1bOA":
		return Key{Kind: KeyUp}
	case "\x1b[B", "\x1bOB":
		return Key{Kind: KeyDown}
	case "\x1b[H", "\x1b[1~":
		return Key{Kind: KeyFirst}
	case "\x1b[F", "\x1b[4~":
		return Key{Kind: KeyLast}
	}
	switch c := b[0]; {
	case c == '\r' || c == '\n':
		return Key{Kind: KeyEnter}
	case c == 0x1b || c == 'q' || c == 0x03:
		return Key{Kind: KeyCancel}
	case c == 'k':
		return Key{Kind: KeyUp}
	case c == 'j':
		return Key{Kind: KeyDown}
	case c == 'g':
		return Key{Kind: KeyFirst}
	case c == 'G':
		return Key{Kind: KeyLast}
	case c >= '0' && c <= '9':
		return Key{Kind: KeyDigit, Digit: int(c - '0')}
	}
	return Key{Kind: KeyOther}
}

// SplitKeys splits one read into key presses: escape sequences stay together, everything else is one byte each.
func SplitKeys(b []byte) [][]byte {
	var keys [][]byte
	for i := 0; i < len(b); {
		if b[i] == 0x1b && i+1 < len(b) && (b[i+1] == '[' || b[i+1] == 'O') {
			j := i + 2
			for j < len(b) && (b[j] < 0x40 || b[j] > 0x7e) {
				j++
			}
			end := min(j, len(b)-1)
			keys = append(keys, b[i:end+1])
			i = j + 1
			continue
		}
		keys = append(keys, b[i:i+1])
		i++
	}
	return keys
}

// Move is the cursor after a key. Typed digits accumulate in typed, so 1 then 4 reaches row 14.
func Move(cursor int, k Key, count int, typed *string) int {
	if count <= 0 {
		return 0
	}
	switch k.Kind {
	case KeyUp:
		*typed = ""
		return max(0, cursor-1)
	case KeyDown:
		*typed = ""
		return min(count-1, cursor+1)
	case KeyFirst:
		*typed = ""
		return 0
	case KeyLast:
		*typed = ""
		return count - 1
	case KeyDigit:
		candidate := *typed + strconv.Itoa(k.Digit)
		if n, err := strconv.Atoi(candidate); err == nil && n < count {
			*typed = candidate
			if len(candidate) >= len(strconv.Itoa(count-1)) {
				*typed = ""
			}
			return n
		}
		*typed = ""
		if k.Digit < count {
			return k.Digit
		}
	}
	return cursor
}

// Indent is the left margin of every block.
const Indent = "  "

var (
	rawMu    sync.Mutex
	rawState *term.State
)

// RestoreTerminal leaves raw mode, if the picker is in it. Safe to call from a signal handler goroutine.
func RestoreTerminal() {
	rawMu.Lock()
	defer rawMu.Unlock()
	if rawState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), rawState)
		rawState = nil
	}
}

// Choose runs the arrow-key chooser and returns the chosen index. render draws the whole block for a cursor;
// overwriting is how many lines of a previous frame are on screen to draw over, so consecutive choosers share one
// block. q / Esc / Ctrl-C return errs.ErrCancelled. Without a terminal on both ends nobody can answer, so it returns
// an input error instead of waiting on stdin; a block too tall to redraw asks for the number instead.
func Choose(count, initial int, prompt string, overwriting int, render func(cursor int) []string) (int, error) {
	if count <= 0 {
		return 0, nil
	}
	if !StdinIsTTY() || !IsTTY() {
		return 0, errs.NewInput("-i needs a terminal. Pick streams with --video-stream / --audio-stream instead (the indexes `haul info` lists)")
	}
	if len(render(initial))+1 >= Height() {
		if overwriting == 0 {
			Lines(render(initial))
		}
		Prompt(fmt.Sprintf("%s [%d]: ", prompt, initial))
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		if n, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && n >= 0 && n < count {
			return n, nil
		}
		return initial, nil
	}

	cursor := min(max(0, initial), count-1)
	typed := ""
	drawn := overwriting
	draw := func() {
		var b strings.Builder
		if drawn > 0 {
			fmt.Fprintf(&b, "\x1b[%dA", drawn)
		}
		lines := append(render(cursor), Indent+CurrentStyle().Dim(prompt+"   "+PickerHint))
		for _, l := range lines {
			b.WriteString("\r\x1b[K" + l + "\r\n")
		}
		Write(b.String())
		drawn = len(lines)
	}

	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return 0, errs.NewInput("cannot read keys from the terminal: %v", err)
	}
	rawMu.Lock()
	rawState = state
	rawMu.Unlock()
	defer RestoreTerminal()

	draw()
	buf := make([]byte, 64)
	for {
		n, err := os.Stdin.Read(buf)
		keys := [][]byte{{0x1b}}
		if err == nil && n > 0 {
			keys = SplitKeys(buf[:n])
		}
		for _, raw := range keys {
			k := ParseKey(raw)
			switch k.Kind {
			case KeyEnter:
				// Leave the final state on screen without the hint line.
				Write("\x1b[1A\r\x1b[K")
				return cursor, nil
			case KeyCancel:
				Write("\x1b[1A\r\x1b[K")
				return 0, errs.ErrCancelled
			default:
				if next := Move(cursor, k, count, &typed); next != cursor {
					cursor = next
					draw()
				}
			}
		}
	}
}
