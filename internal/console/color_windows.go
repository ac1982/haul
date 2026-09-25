//go:build windows

package console

import (
	"sync"

	"golang.org/x/sys/windows"
)

var (
	vtOnce sync.Once
	vtOK   bool
)

// enableColor turns on ANSI escape processing for the log stream (Windows 10 and later); without it, no colour.
func enableColor() bool {
	vtOnce.Do(func() {
		h := windows.Handle(Out().Fd())
		var mode uint32
		if windows.GetConsoleMode(h, &mode) != nil {
			return
		}
		vtOK = windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
	})
	return vtOK
}
