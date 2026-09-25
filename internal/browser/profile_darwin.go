package browser

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/shell"
)

func dataDirectory(b Browser) string {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, "Library", "Application Support")
	if b == Edge {
		return filepath.Join(base, "Microsoft Edge")
	}
	return filepath.Join(base, "Google", "Chrome")
}

func readProfile(ctx context.Context, b Browser, profile, hostSuffix string) ([]Cookie, error) {
	dataDir := dataDirectory(b)
	if _, err := os.Stat(dataDir); err != nil {
		return nil, errs.NewAuth("%s data not found: %s", b.Name(), dataDir)
	}
	profileDir := filepath.Join(dataDir, profile)
	var dbPath string
	for _, candidate := range []string{filepath.Join(profileDir, "Network", "Cookies"), filepath.Join(profileDir, "Cookies")} {
		if _, err := os.Stat(candidate); err == nil {
			dbPath = candidate
			break
		}
	}
	if dbPath == "" {
		if _, err := os.ReadDir(profileDir); errors.Is(err, fs.ErrPermission) {
			return nil, fullDiskAccessError(b)
		}
		return nil, errs.NewAuth("No cookie database in the %s profile %q", b.Name(), profile)
	}
	// Probe readability before touching the keychain, so a permissions problem does not trigger a pointless prompt.
	probe, err := os.Open(dbPath)
	if err != nil {
		return nil, fullDiskAccessError(b)
	}
	probe.Close()
	password, err := safeStoragePassword(ctx, b)
	if err != nil {
		return nil, err
	}
	cookies, err := readDatabase(ctx, dbPath, password, hostSuffix)
	if errors.Is(err, fs.ErrPermission) {
		return nil, fullDiskAccessError(b)
	}
	return cookies, err
}

func fullDiskAccessError(b Browser) error {
	return errs.NewAuth("Cannot read the %s profile (Operation not permitted).\n"+
		"macOS protects browser data: allow your terminal app under System Settings → Privacy & Security → Full Disk Access, then run again.", b.Name())
}

// safeStoragePassword is the random password Chromium keeps in the login keychain. macOS asks the first time.
func safeStoragePassword(ctx context.Context, b Browser) ([]byte, error) {
	res, err := shell.Run(ctx, "/usr/bin/security", []string{"find-generic-password", "-w", "-s", b.keychainService(), "-a", b.keychainAccount()},
		shell.Options{Capture: true})
	if err == nil && res.Status == 0 {
		return []byte(strings.TrimRight(res.Output, "\r\n")), nil
	}
	reason := strings.TrimSpace(res.Errors)
	if err != nil {
		reason = err.Error()
	}
	return nil, errs.NewAuth("Cannot read %q from the keychain: %s\nChoose \"Always Allow\" when macOS asks.", b.keychainService(), reason)
}
