package browser

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ac1982/haul/internal/errs"
)

// encrypt is decrypt's inverse, in the newer layout that prefixes SHA-256(host) unless host is "".
func encrypt(t *testing.T, value string, key []byte, host string) []byte {
	t.Helper()
	var plain []byte
	if host != "" {
		sum := sha256.Sum256([]byte(host))
		plain = append(plain, sum[:]...)
	}
	plain = append(plain, value...)
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	plain = append(plain, bytes.Repeat([]byte{byte(pad)}, pad)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]byte, len(plain))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, plain)
	return append([]byte("v10"), out...)
}

func TestDeriveKeyIsChromiumsPBKDF2(t *testing.T) {
	t.Parallel()
	// PBKDF2-HMAC-SHA1("peanuts", "saltysalt", 1003, 16), as Python's hashlib.pbkdf2_hmac computes it.
	if got := hex.EncodeToString(deriveKey([]byte("peanuts"))); got != "d9a09d499b4e1b7461f28e67972c6dbd" {
		t.Fatalf("key = %s", got)
	}
}

func TestDecryptsChromiumCookies(t *testing.T) {
	t.Parallel()
	sqlite, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Skip("needs sqlite3")
	}
	password := []byte("fake-safe-storage-password")
	key := deriveKey(password)
	db := filepath.Join(t.TempDir(), "Cookies")
	var sql strings.Builder
	sql.WriteString("CREATE TABLE cookies (host_key TEXT, name TEXT, value TEXT, encrypted_value BLOB);\n")
	insert := func(host, name, value string) {
		fmt.Fprintf(&sql, "INSERT INTO cookies VALUES ('%s', '%s', '', X'%s');\n", host, name, hex.EncodeToString(encrypt(t, value, key, host)))
	}
	insert(".bilibili.com", "SESSDATA", "abc%2Cdef")
	insert(".bilibili.com", "bili_jct", "csrf123")
	insert(".bilibili.com", "DedeUserID", "42")
	insert(".example.com", "other", "no")
	sql.WriteString("INSERT INTO cookies VALUES ('www.bilibili.com', 'plain', 'kept', X'');\n")
	if out, err := exec.Command(sqlite, db, sql.String()).CombinedOutput(); err != nil {
		t.Fatalf("sqlite3: %v %s", err, out)
	}

	cookies, err := readDatabase(context.Background(), db, password, "bilibili.com")
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, c := range cookies {
		names = append(names, c.Name)
	}
	if strings.Join(names, ",") != "DedeUserID,SESSDATA,bili_jct,plain" {
		t.Fatalf("names = %v", names)
	}
	if cookies[1].Value != "abc%2Cdef" || cookies[3].Value != "kept" {
		t.Fatalf("cookies = %+v", cookies)
	}
	if h := Header(cookies); h != "DedeUserID=42; SESSDATA=abc%2Cdef; bili_jct=csrf123; plain=kept" {
		t.Fatalf("header = %q", h)
	}

	none, err := readDatabase(context.Background(), db, password, "nowhere.test")
	if err != nil || len(none) != 0 {
		t.Fatalf("no rows: %v %v", none, err)
	}
}

func TestDecryptWithAndWithoutHostHash(t *testing.T) {
	t.Parallel()
	key := deriveKey([]byte("pw"))
	hashed := encrypt(t, "plain", key, "h")
	if got, err := decrypt(hashed, key, "h"); err != nil || got != "plain" {
		t.Fatalf("hashed: %q %v", got, err)
	}
	// Another host's hash is not stripped: the value comes through with the 32 bytes in front.
	if got, _ := decrypt(hashed, key, "other-host"); !strings.HasSuffix(got, "plain") || len(got) != 32+5 {
		t.Fatalf("other host: %q", got)
	}
	// Older Chromium builds do not prefix the hash; the plaintext comes through untouched.
	if got, err := decrypt(encrypt(t, "old", key, ""), key, "h"); err != nil || got != "old" {
		t.Fatalf("unhashed: %q %v", got, err)
	}
	for _, bad := range [][]byte{[]byte("v11abc"), []byte("v10"), []byte("v10short")} {
		if _, err := decrypt(bad, key, "h"); err == nil {
			t.Fatalf("%q decrypted", bad)
		}
	}
	if _, err := decrypt(hashed, deriveKey([]byte("wrong")), "h"); err == nil {
		t.Fatal("wrong key decrypted")
	}
}

func TestHeaderKeepsTheFirstOfEachName(t *testing.T) {
	t.Parallel()
	h := Header([]Cookie{{Name: "a", Value: "1"}, {Name: "b"}, {Name: "a", Value: "2"}, {Name: "c", Value: "3"}})
	if h != "a=1; c=3" {
		t.Fatalf("header = %q", h)
	}
}

func TestUnknownBrowserIsAnInputError(t *testing.T) {
	t.Parallel()
	if _, err := Cookies("firefox", "", "bilibili.com"); !errs.Is(err, errs.Input) {
		t.Fatalf("err = %v", err)
	}
}

func TestOtherSystemsAreRefused(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "darwin" {
		t.Skip("macOS reads real profiles")
	}
	if _, err := Cookies(Chrome, "Default", "bilibili.com"); !errs.Is(err, errs.Auth) {
		t.Fatalf("err = %v", err)
	}
}
