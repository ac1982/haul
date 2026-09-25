package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHomeAndFiles(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HAUL_HOME", dir)
	if Home() != dir {
		t.Fatalf("Home = %q", Home())
	}
	if Read(CookieFile) != "" {
		t.Error("a missing file reads as empty")
	}
	if err := Write(CookieFile, "  SESSDATA=x \n"); err != nil {
		t.Fatal(err)
	}
	if Read(CookieFile) != "SESSDATA=x" {
		t.Errorf("Read = %q", Read(CookieFile))
	}
	if info, _ := os.Stat(Path(CookieFile)); runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("a login file is private, got %v", info.Mode().Perm())
	}
	_ = Append(ArchiveFile, "a|")
	_ = Append(ArchiveFile, "b|")
	if Read(ArchiveFile) != "a|b|" {
		t.Errorf("Append: %q", Read(ArchiveFile))
	}

	t.Setenv("HAUL_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", dir)
	if Home() != filepath.Join(dir, "haul") {
		t.Errorf("XDG: %q", Home())
	}
}
