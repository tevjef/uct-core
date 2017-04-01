package harmony

import (
	"math/rand"
	"testing"
	"time"
)

func TestAddJitter(t *testing.T) {
	t.Run("10 second jitter", func(t *testing.T) {
		jitter := 10
		interval := 1 * time.Minute

		newOffset := addJitter(jitter, interval, rand.NewSource(1))
		if newOffset < interval || newOffset > interval*2 {
			t.Error("jitter is out of range")
		}
	})
	t.Run("0 second jitter", func(t *testing.T) {
		jitter := 0
		interval := 30 * time.Second
		newOffset := addJitter(jitter, interval, rand.NewSource(1))
		if newOffset != interval {
			t.Error("jitter is out of range")
		}
	})
}
