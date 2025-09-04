package reconcileutil

import (
	"testing"
	"time"
)

func TestBackoffMonotonicCap(t *testing.T) {
	base := 10 * time.Millisecond
	maxDelay := 200 * time.Millisecond
	for retry := 0; retry < 6; retry++ {
		d := Backoff(retry, base, maxDelay)
		if d < 0 || d > maxDelay {
			t.Fatalf("backoff out of range: %v", d)
		}
	}
}
