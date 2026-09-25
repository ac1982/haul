package media

import (
	"testing"
	"time"
)

func TestEstimatedSize(t *testing.T) {
	if EstimatedSize(123, 1000, time.Minute) != 123 {
		t.Error("a stated size wins")
	}
	if got := EstimatedSize(0, 1000, 8*time.Second); got != 1_000_000 {
		t.Errorf("1000 kbps for 8 s = %d bytes", got)
	}
	if EstimatedSize(0, 0, time.Minute) != 0 {
		t.Error("unknown bitrate")
	}
}
