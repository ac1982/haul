package main

import (
	"context"
	"fmt"

	"github.com/ac1982/haul/internal/console"
	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/extract"
	"github.com/ac1982/haul/internal/httpx"
	"github.com/ac1982/haul/internal/sites/bilibili"
)

// login is `haul login <site>`.
func login(ctx context.Context, args []string) int {
	fs := newFlagSet("login", loginFlags)
	if code, ok := parse(fs, "login", args); !ok {
		return code
	}
	console.SetDebug(flagBool(fs, "debug"))
	if len(fs.Args()) != 1 {
		return usageError("login", errMissing("<site>"))
	}
	router := extract.NewRouter(bilibili.New(httpx.Default, bilibili.DefaultOptions()))
	x := router.ByName(fs.Args()[0])
	auth, ok := x.(extract.Authenticator)
	if !ok {
		return usageError("login", fmt.Errorf("the value %q is invalid for <site>: logging in works for bilibili", fs.Args()[0]))
	}
	req := extract.LoginRequest{Method: "qr", Profile: flagString(fs, "profile")}
	switch {
	case flagBool(fs, "from-edge") && flagBool(fs, "from-chrome"):
		return fail(errs.NewInput("--from-edge and --from-chrome exclude each other"))
	case flagBool(fs, "from-edge"):
		req.Method, req.Browser = "browser", "edge"
	case flagBool(fs, "from-chrome"):
		req.Method, req.Browser = "browser", "chrome"
	case flagBool(fs, "tv"):
		req.Method = "tv"
	}
	console.Banner("haul " + console.CurrentStyle().Dim(version))
	console.Plain("")
	if err := auth.Login(ctx, req); err != nil {
		if ctx.Err() != nil {
			err = errs.ErrCancelled
		}
		return fail(err)
	}
	return 0
}
