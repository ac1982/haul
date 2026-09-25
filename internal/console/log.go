package console

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

var (
	debug    atomic.Bool
	toStderr atomic.Bool
	writeMu  sync.Mutex
)

// Debug reports whether debug output (--debug) is on.
func Debug() bool { return debug.Load() }

// SetDebug turns debug output on or off.
func SetDebug(on bool) { debug.Store(on) }

// ToStderr reports whether log lines go to stderr.
func ToStderr() bool { return toStderr.Load() }

// SetToStderr sends every log line to stderr, keeping stdout for data.
func SetToStderr(on bool) { toStderr.Store(on) }

// Write puts raw text on the log stream.
func Write(s string) {
	writeMu.Lock()
	defer writeMu.Unlock()
	_, _ = Out().WriteString(s)
}

// Output writes data (the --json document) to stdout, whatever ToStderr says.
func Output(s string) {
	writeMu.Lock()
	defer writeMu.Unlock()
	_, _ = os.Stdout.WriteString(s + "\n")
}

// Timestamp is [2006-01-02 15:04:05.000].
func Timestamp(t time.Time) string { return t.Format("[2006-01-02 15:04:05.000]") }

func emit(glyph, text string, newline bool) {
	prefix := ""
	if Debug() {
		prefix = CurrentStyle().Gray(Timestamp(time.Now())) + " "
	}
	end := ""
	if newline {
		end = "\n"
	}
	Write(prefix + glyph + text + end)
}

// Info is a neutral line.
func Info(text string) { emit("  ", text, true) }

// Status is low-key activity: a cookie loaded, a subtitle found…
func Status(text string) { emit("  ", CurrentStyle().Dim("· "+text), true) }

// Success is a ✓ line.
func Success(text string) { emit(CurrentStyle().Green("✓ "), text, true) }

// Warn is a ! line for a problem that does not stop the run.
func Warn(text string) { s := CurrentStyle(); emit(s.Yellow("! "), s.Yellow(text), true) }

// Error is a ✗ line.
func Error(text string) { s := CurrentStyle(); emit(s.Red("✗ "), s.Red(text), true) }

// Highlight is a cyan line.
func Highlight(text string) { emit("  ", CurrentStyle().Cyan(text), true) }

// Banner is a bold line without indent.
func Banner(text string) { Write(CurrentStyle().Bold(text) + "\n") }

// Plain is a line as it is.
func Plain(text string) { Write(text + "\n") }

// Lines prints each line.
func Lines(lines []string) {
	for _, l := range lines {
		Plain(l)
	}
}

// Prompt is a question; the cursor stays on the line.
func Prompt(text string) { emit("  ", CurrentStyle().Cyan(text), false) }

// Debugf prints only under --debug, with a timestamp.
func Debugf(format string, a ...any) {
	if !Debug() {
		return
	}
	Write(CurrentStyle().Gray(Timestamp(time.Now())+" "+fmt.Sprintf(format, a...)) + "\n")
}
