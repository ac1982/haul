// Package errs is haul's user-facing error: a message a person can read and a kind that decides the exit code
// and the `error.kind` of `--json` output, so scripts and agents can tell a bad link from a missing tool from a
// failed download.
package errs

import (
	"errors"
	"fmt"
)

// Kind classifies a failure.
type Kind string

const (
	// Failed: the download or extraction failed (network, site, muxing).
	Failed Kind = "failed"
	// Input: the input cannot be handled — an unsupported or malformed link, a bad option value.
	Input Kind = "input"
	// Dependency: an external tool (ffmpeg, yt-dlp…) is missing.
	Dependency Kind = "dependency"
	// Auth: the content needs a login, or the stored login is no longer valid.
	Auth Kind = "auth"
	// Cancelled: the person cancelled an interactive choice or interrupted the run.
	Cancelled Kind = "cancelled"
)

// ExitCode is 1 failed, 2 input, 3 dependency, 4 auth, 130 cancelled. Bad command-line flags exit with 64.
func (k Kind) ExitCode() int {
	switch k {
	case Input:
		return 2
	case Dependency:
		return 3
	case Auth:
		return 4
	case Cancelled:
		return 130
	default:
		return 1
	}
}

// Error is a failure with a message meant for people.
type Error struct {
	Kind    Kind
	Message string
}

func (e *Error) Error() string { return e.Message }

// New is a Failed error.
func New(format string, a ...any) error {
	return &Error{Kind: Failed, Message: fmt.Sprintf(format, a...)}
}

// NewInput is an Input error.
func NewInput(format string, a ...any) error {
	return &Error{Kind: Input, Message: fmt.Sprintf(format, a...)}
}

// NewDependency is a Dependency error.
func NewDependency(format string, a ...any) error {
	return &Error{Kind: Dependency, Message: fmt.Sprintf(format, a...)}
}

// NewAuth is an Auth error.
func NewAuth(format string, a ...any) error {
	return &Error{Kind: Auth, Message: fmt.Sprintf(format, a...)}
}

// ErrCancelled is what an interrupted choice or run returns.
var ErrCancelled error = &Error{Kind: Cancelled, Message: "Cancelled"}

// KindOf reports the kind of err: its own for an *Error anywhere in the chain, Failed for anything else.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return Failed
}

// Is reports whether err is an *Error of kind k.
func Is(err error, k Kind) bool { return err != nil && KindOf(err) == k }
