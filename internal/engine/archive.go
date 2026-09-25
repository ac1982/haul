package engine

import (
	"strings"

	"github.com/ac1982/haul/internal/media"
	"github.com/ac1982/haul/internal/storage"
)

// archiveKey is how --archive remembers an entry: site:id, so ids of different sites never collide.
func archiveKey(site media.Site, e *media.Entry) string { return string(site) + ":" + e.ID }

func archived(key string) bool {
	for _, line := range strings.Split(storage.Read(storage.ArchiveFile), "\n") {
		if strings.TrimSpace(line) == key {
			return true
		}
	}
	return false
}

func archive(key string) error { return storage.Append(storage.ArchiveFile, key+"\n") }
