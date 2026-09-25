package errs

import (
	"errors"
	"fmt"
	"testing"
)

func TestKindsAndExitCodes(t *testing.T) {
	cases := []struct {
		err  error
		kind Kind
		code int
	}{
		{New("x"), Failed, 1}, {NewInput("x"), Input, 2}, {NewDependency("x"), Dependency, 3},
		{NewAuth("x"), Auth, 4}, {ErrCancelled, Cancelled, 130}, {errors.New("plain"), Failed, 1},
		{fmt.Errorf("wrapped: %w", NewInput("bad %s", "link")), Input, 2},
	}
	for _, c := range cases {
		if KindOf(c.err) != c.kind || KindOf(c.err).ExitCode() != c.code || !Is(c.err, c.kind) {
			t.Errorf("%v: kind %s, exit %d", c.err, KindOf(c.err), KindOf(c.err).ExitCode())
		}
	}
	if NewInput("bad %s", "link").Error() != "bad link" || Is(nil, Failed) {
		t.Error("messages and nil")
	}
}
