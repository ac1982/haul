package console

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestDisplayWidthCountsCJKDouble(t *testing.T) {
	for s, want := range map[string]int{"abc": 3, "4K 超清": 7, "1080P 高帧率": 12, "é": 1, "é": 1} {
		if got := DisplayWidth(s); got != want {
			t.Errorf("DisplayWidth(%q) = %d, want %d", s, got, want)
		}
	}
}

func TestPadAndTruncateMeasureCells(t *testing.T) {
	cases := [][2]string{
		{Pad("超", 4, false), "超  "},
		{Pad("ab", 4, true), "  ab"},
		{Pad("toolong", 3, false), "toolong"},
		{Truncate("哔哩哔哩下载", 5), "哔哩…"},
		{Truncate("short", 10), "short"},
	}
	for _, c := range cases {
		if c[0] != c[1] {
			t.Errorf("got %q, want %q", c[0], c[1])
		}
	}
}

func TestPrettyPathUsesTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	if got := PrettyPath(filepath.Join(home, "Movies", "a.mp4")); got != filepath.Join("~", "Movies", "a.mp4") {
		t.Errorf("home: %q", got)
	}
	if runtime.GOOS == "windows" {
		return
	}
	if got := PrettyPath("/tmp/x.mp4"); got != "/tmp/x.mp4" {
		t.Errorf("absolute: %q", got)
	}
	if got := PrettyPath("rel/x.mp4"); !strings.HasSuffix(got, "/rel/x.mp4") {
		t.Errorf("relative: %q", got)
	}
}

func TestStyles(t *testing.T) {
	if PlainStyle.Bold("a") != "a" || PlainStyle.Cyan("b") != "b" || PlainStyle.Dim("c") != "c" {
		t.Error("plain style must not emit escapes")
	}
	if got := (Style{Color: true}).Red("x"); got != "\x1b[31mx\x1b[0m" {
		t.Errorf("red = %q", got)
	}
}

func TestKeyParsing(t *testing.T) {
	cases := map[string]Key{
		"\x1b[A": {Kind: KeyUp}, "\x1b[B": {Kind: KeyDown}, "k": {Kind: KeyUp}, "j": {Kind: KeyDown},
		"\r": {Kind: KeyEnter}, "\n": {Kind: KeyEnter}, "q": {Kind: KeyCancel}, "\x1b": {Kind: KeyCancel},
		"\x03": {Kind: KeyCancel}, "7": {Kind: KeyDigit, Digit: 7}, "x": {Kind: KeyOther}, "\x1b[4~": {Kind: KeyLast},
	}
	for in, want := range cases {
		if got := ParseKey([]byte(in)); got != want {
			t.Errorf("ParseKey(%q) = %+v, want %+v", in, got, want)
		}
	}
}

func TestSplitsCoalescedReads(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"\x1b[B\r", []string{"\x1b[B", "\r"}},
		{"\x1b[A\x1b[A3", []string{"\x1b[A", "\x1b[A", "3"}},
		{"\x1b", []string{"\x1b"}},
		{"\x1b[1~", []string{"\x1b[1~"}},
		{"", nil},
	}
	for _, c := range cases {
		var got []string
		for _, k := range SplitKeys([]byte(c.in)) {
			got = append(got, string(k))
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("SplitKeys(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCursorMovementClampsAndJumps(t *testing.T) {
	typed := ""
	check := func(got, want int, what string) {
		t.Helper()
		if got != want {
			t.Errorf("%s: got %d, want %d", what, got, want)
		}
	}
	check(Move(0, Key{Kind: KeyUp}, 5, &typed), 0, "up at top")
	check(Move(4, Key{Kind: KeyDown}, 5, &typed), 4, "down at bottom")
	check(Move(2, Key{Kind: KeyDown}, 5, &typed), 3, "down")
	check(Move(2, Key{Kind: KeyLast}, 5, &typed), 4, "last")
	check(Move(2, Key{Kind: KeyFirst}, 5, &typed), 0, "first")
	check(Move(2, Key{Kind: KeyDigit, Digit: 9}, 5, &typed), 2, "out of range digit")
	typed = ""
	check(Move(0, Key{Kind: KeyDigit, Digit: 1}, 15, &typed), 1, "1")
	check(Move(1, Key{Kind: KeyDigit, Digit: 4}, 15, &typed), 14, "then 4")
	if typed != "" {
		t.Errorf("typed buffer should clear, got %q", typed)
	}
	check(Move(0, Key{Kind: KeyOther}, 15, &typed), 0, "other")
}

func TestChooseWithoutATerminalIsAnInputError(t *testing.T) {
	// go test runs without a terminal on stdin.
	if StdinIsTTY() {
		t.Skip("stdin is a terminal")
	}
	if _, err := Choose(3, 0, "pick", 0, func(int) []string { return []string{"a", "b", "c"} }); err == nil ||
		!strings.Contains(err.Error(), "--video-stream") {
		t.Errorf("err = %v", err)
	}
}
