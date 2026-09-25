// Package storage is where haul keeps its config, logins and download archive: $HAUL_HOME, else
// $XDG_CONFIG_HOME/haul, else ~/.config/haul.
package storage

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	CookieFile   = "cookie.txt"
	TVTokenFile  = "tv-token.txt"
	AppTokenFile = "app-token.txt"
	ConfigFile   = "config.json"
	ArchiveFile  = "archives.txt"
)

// Home is the directory.
func Home() string {
	if h := os.Getenv("HAUL_HOME"); h != "" {
		return h
	}
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "haul")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".config", "haul")
	}
	return filepath.Join(home, ".config", "haul")
}

// Path is a file in it.
func Path(name string) string { return filepath.Join(Home(), name) }

// Read is a file's trimmed text; "" when missing or empty.
func Read(name string) string {
	b, err := os.ReadFile(Path(name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

// Write replaces a file, readable by its owner only: it may hold a login.
func Write(name, text string) error {
	if err := os.MkdirAll(Home(), 0o755); err != nil {
		return err
	}
	return os.WriteFile(Path(name), []byte(text), 0o600)
}

// Append adds text at the end of a file.
func Append(name, text string) error {
	if err := os.MkdirAll(Home(), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(Path(name), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(text)
	return err
}
