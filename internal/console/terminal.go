package console

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/term"
)

// Out is the stream log lines go to.
func Out() *os.File {
	if ToStderr() {
		return os.Stderr
	}
	return os.Stdout
}

// IsTTY reports whether log output is a terminal.
func IsTTY() bool { return term.IsTerminal(int(Out().Fd())) }

// StdinIsTTY reports whether someone can answer on stdin.
func StdinIsTTY() bool { return term.IsTerminal(int(os.Stdin.Fd())) }

// SupportsColor: a terminal, without NO_COLOR, and not TERM=dumb.
func SupportsColor() bool {
	return IsTTY() && os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") != "dumb" && enableColor()
}

// CurrentStyle is the style for log output right now.
func CurrentStyle() Style { return Style{Color: SupportsColor()} }

// Size is the terminal's columns and rows: 80×24, or $COLUMNS / $LINES, when unknown.
func Size() (width, height int) {
	if w, h, err := term.GetSize(int(Out().Fd())); err == nil && w > 0 && h > 0 {
		return w, h
	}
	width, height = 80, 24
	if c := atoiEnv("COLUMNS"); c > 0 {
		width = c
	}
	if l := atoiEnv("LINES"); l > 0 {
		height = l
	}
	return width, height
}

// Width is the terminal's columns.
func Width() int { w, _ := Size(); return w }

// Height is the terminal's rows.
func Height() int { _, h := Size(); return h }

func atoiEnv(name string) int {
	n := 0
	for _, c := range os.Getenv(name) {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

var wideRanges = [][2]rune{
	{0x1100, 0x115F}, {0x2E80, 0x303E}, {0x3041, 0x33FF}, {0x3400, 0x4DBF}, {0x4E00, 0x9FFF},
	{0xA000, 0xA4CF}, {0xAC00, 0xD7A3}, {0xF900, 0xFAFF}, {0xFE30, 0xFE4F}, {0xFF00, 0xFF60},
	{0xFFE0, 0xFFE6}, {0x1F300, 0x1F64F}, {0x1F900, 0x1F9FF}, {0x20000, 0x3FFFD},
}

// CellWidth is how many terminal cells a character takes: CJK and emoji two, combining marks none.
func CellWidth(r rune) int {
	if r < 0x20 || (r >= 0x7F && r <= 0x9F) || r == 0x200B || r == 0x200D {
		return 0
	}
	if unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf) {
		return 0
	}
	for _, w := range wideRanges {
		if r >= w[0] && r <= w[1] {
			return 2
		}
	}
	return 1
}

// DisplayWidth is how many cells s takes.
func DisplayWidth(s string) int {
	n := 0
	for _, r := range s {
		n += CellWidth(r)
	}
	return n
}

// Pad fills s with spaces to width cells, left-aligned unless right.
func Pad(s string, width int, right bool) string {
	w := DisplayWidth(s)
	if w >= width {
		return s
	}
	fill := strings.Repeat(" ", width-w)
	if right {
		return fill + s
	}
	return s + fill
}

// Truncate cuts s to width cells, ending in an ellipsis.
func Truncate(s string, width int) string {
	if DisplayWidth(s) <= width || width <= 1 {
		return s
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		w := CellWidth(r)
		if used+w > width-1 {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String() + "…"
}

// PrettyPath is an absolute path with the home directory shown as ~.
func PrettyPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = filepath.Clean(p)
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return abs
	}
	if abs == home {
		return "~"
	}
	if strings.HasPrefix(abs, home+string(filepath.Separator)) {
		return "~" + abs[len(home):]
	}
	return abs
}
