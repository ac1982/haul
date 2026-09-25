//go:build !darwin

package browser

import (
	"context"

	"github.com/ac1982/haul/internal/errs"
)

func readProfile(context.Context, Browser, string, string) ([]Cookie, error) {
	return nil, errs.NewAuth("Reading browser cookies works on macOS only for now")
}
