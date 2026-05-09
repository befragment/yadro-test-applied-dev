package clock_test

import (
	"testing"
	"time"

	"github.com/befragment/yadro-test-applied-dev/pkg/clock"
)

func TestNewClock(t *testing.T) {
	t.Parallel()

	c := clock.NewClock()
	if c == nil {
		t.Fatalf("expected non-nil clock")
	}
}

func TestClockNow(t *testing.T) {
	t.Parallel()

	c := clock.NewClock()
	if c == nil {
		t.Fatalf("expected non-nil clock")
	}

	before := time.Now()
	got := c.Now()
	after := time.Now()

	if got.Location() != time.UTC {
		t.Fatalf("expected location UTC, got %v", got.Location())
	}

	// Ensure returned time is within a small window.
	beforeNs := before.UTC().UnixNano()
	afterNs := after.UTC().UnixNano()
	gotNs := got.UnixNano()

	const maxSkewNs = int64(2 * time.Second) // generous, still catches obvious issues
	minNs := beforeNs - maxSkewNs
	maxNs := afterNs + maxSkewNs

	if gotNs < minNs || gotNs > maxNs {
		t.Fatalf("expected got time to be within [%d, %d], got %d", minNs, maxNs, gotNs)
	}
}
