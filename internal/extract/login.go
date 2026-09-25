package extract

import "context"

// LoginRequest says how to log in: Method is "qr" (scan with the site's app), "tv" (a TV-app token) or "browser"
// (copy a browser's login; Browser and Profile say which).
type LoginRequest struct {
	Method  string
	Browser string
	Profile string
}

// Authenticator is an extractor that can log in and keep the login for later runs.
type Authenticator interface {
	Login(ctx context.Context, req LoginRequest) error
}
