// Package browser reads login cookies out of a Chromium-based browser's profile (Microsoft Edge, Google Chrome), so
// a site's login can be copied from the browser instead of scanning a QR code.
//
// Chromium stores cookie values AES-128-CBC encrypted with a key derived (PBKDF2-HMAC-SHA1, salt "saltysalt",
// 1003 rounds) from a random password kept in the login keychain as "<Browser> Safe Storage". Newer builds also
// prefix the plaintext with SHA-256(host_key). Everything is done with the standard library and two command-line
// tools macOS ships with, `security` and `sqlite3`, so haul needs no cgo.
package browser

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"

	"github.com/ac1982/haul/internal/errs"
	"github.com/ac1982/haul/internal/jsonv"
	"github.com/ac1982/haul/internal/shell"
)

// Browser is a supported Chromium-based browser.
type Browser string

const (
	Edge   Browser = "edge"
	Chrome Browser = "chrome"
)

// Name is how the browser calls itself.
func (b Browser) Name() string {
	if b == Edge {
		return "Microsoft Edge"
	}
	return "Google Chrome"
}

func (b Browser) keychainService() string {
	if b == Edge {
		return "Microsoft Edge Safe Storage"
	}
	return "Chrome Safe Storage"
}

func (b Browser) keychainAccount() string {
	if b == Edge {
		return "Microsoft Edge"
	}
	return "Chrome"
}

// Cookie is one decrypted cookie.
type Cookie struct {
	Host  string
	Name  string
	Value string
}

// Header is a Cookie header from cookies, `name=value; …`, keeping the first cookie of each name and skipping
// empty values.
func Header(cookies []Cookie) string {
	seen := map[string]bool{}
	var parts []string
	for _, c := range cookies {
		if c.Value == "" || seen[c.Name] {
			continue
		}
		seen[c.Name] = true
		parts = append(parts, c.Name+"="+c.Value)
	}
	return strings.Join(parts, "; ")
}

// Cookies reads every cookie whose host ends with hostSuffix from a browser profile ("Default", "Profile 1"…).
func Cookies(b Browser, profile, hostSuffix string) ([]Cookie, error) {
	if b != Edge && b != Chrome {
		return nil, errs.NewInput("Unknown browser %q: use edge or chrome", string(b))
	}
	if profile == "" {
		profile = "Default"
	}
	return readProfile(context.Background(), b, profile, hostSuffix)
}

// deriveKey turns the Safe Storage password into the AES key.
func deriveKey(password []byte) []byte {
	key, err := pbkdf2.Key(sha1.New, string(password), []byte("saltysalt"), 1003, 16)
	if err != nil {
		panic(err) // only for key lengths out of range, and 16 is not
	}
	return key
}

// iv is Chromium's fixed initialisation vector: 16 spaces.
var iv = bytes.Repeat([]byte{' '}, aes.BlockSize)

// decrypt reverses Chromium's `v10` + AES-128-CBC(PKCS#7); a leading SHA-256(host) is dropped.
func decrypt(encrypted, key []byte, host string) (string, error) {
	body, ok := bytes.CutPrefix(encrypted, []byte("v10"))
	if !ok || len(body) == 0 {
		return "", errors.New("unsupported cookie encryption")
	}
	if len(body)%aes.BlockSize != 0 {
		return "", errors.New("cookie value is not a whole number of blocks")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	plain := make([]byte, len(body))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, body)
	pad := int(plain[len(plain)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(plain) || !bytes.Equal(plain[len(plain)-pad:], bytes.Repeat([]byte{byte(pad)}, pad)) {
		return "", errors.New("cookie decryption failed: bad padding")
	}
	plain = plain[:len(plain)-pad]
	hostHash := sha256.Sum256([]byte(host))
	plain, _ = bytes.CutPrefix(plain, hostHash[:])
	return string(plain), nil
}

// readDatabase decrypts the cookies of hostSuffix in a Chromium Cookies database. The file is copied first because
// the running browser keeps it locked, and read with the sqlite3 tool.
func readDatabase(ctx context.Context, path string, password []byte, hostSuffix string) ([]Cookie, error) {
	sqlite := shell.FindExecutable("sqlite3")
	if sqlite == "" {
		return nil, errs.NewDependency("sqlite3 not found on PATH; it reads the browser's cookie database")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp("", "haul-cookies-*.db")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	_, err = tmp.Write(data)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return nil, err
	}

	pattern := strings.ReplaceAll("%"+hostSuffix, "'", "''")
	query := "SELECT host_key, name, value, hex(encrypted_value) AS encrypted FROM cookies WHERE host_key LIKE '" + pattern + "' ORDER BY host_key, name"
	res, err := shell.Run(ctx, sqlite, []string{"-readonly", "-json", tmp.Name(), query}, shell.Options{Capture: true})
	if err != nil {
		return nil, err
	}
	if res.Status != 0 {
		return nil, errs.New("Cannot read the cookie database %s: %s", path, strings.TrimSpace(res.Errors))
	}
	if strings.TrimSpace(res.Output) == "" {
		return nil, nil // sqlite3 prints nothing for no rows
	}
	rows, err := jsonv.ParseString(res.Output)
	if err != nil {
		return nil, err
	}
	key := deriveKey(password)
	var out []Cookie
	for _, row := range rows.Array() {
		c := Cookie{Host: row.Get("host_key").String(), Name: row.Get("name").String(), Value: row.Get("value").String()}
		if c.Value == "" {
			if blob, err := hex.DecodeString(row.Get("encrypted").String()); err == nil && len(blob) > 0 {
				c.Value, _ = decrypt(blob, key, c.Host) // an undecryptable cookie is left empty, as the browser would
			}
		}
		out = append(out, c)
	}
	return out, nil
}
