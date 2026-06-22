package metric

import (
	"math"
	"sort"
	"sync"
	"time"
)

// defaultPercentileTrackerCapacity is the number of most-recent observations a PercentileTracker retains
// when it is created with a non-positive capacity.
const defaultPercentileTrackerCapacity = 512

// PercentileTracker keeps a fixed-capacity, time-stamped rolling window of recent observations and computes
// nearest-rank percentiles over a trailing time window. It is a standalone primitive, like the moving
// averages in this package: it does not know about stats.Histogram or any metrics backend.
// To attach percentiles to a measurement, wrap both in your own type and feed the tracker alongside the measurement.
//
// It is safe for concurrent use.
type PercentileTracker interface {
	// Observe records a value, stamped with the current time. Once capacity observations are held the
	// oldest is overwritten.
	Observe(value float64)
	// Percentile returns the nearest-rank p-th percentile (p in [0,100]) over the observations made within
	// the last window, and true when at least one qualifies. It returns (0, false) for an invalid p
	// (outside [0,100] or NaN), a non-positive window, or a window holding no observations.
	Percentile(p float64, window time.Duration) (float64, bool)
}

// PercentileTrackerOption configures a PercentileTracker created by NewPercentileTracker.
type PercentileTrackerOption func(*percentileTracker)

// WithPercentileTrackerNow overrides the clock used to stamp and expire observations.
// It is intended for tests; a nil function is ignored.
// When unset the tracker uses time.Now.
func WithPercentileTrackerNow(now func() time.Time) PercentileTrackerOption {
	return func(t *percentileTracker) {
		if now != nil {
			t.now = now
		}
	}
}

// NewPercentileTracker returns a PercentileTracker retaining up to capacity most-recent observations.
// A non-positive capacity falls back to defaultPercentileTrackerCapacity.
func NewPercentileTracker(capacity int, opts ...PercentileTrackerOption) PercentileTracker {
	if capacity <= 0 {
		capacity = defaultPercentileTrackerCapacity
	}
	t := &percentileTracker{
		capacity: capacity,
		now:      time.Now,
		times:    make([]time.Time, capacity),
		values:   make([]float64, capacity),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

type percentileTracker struct {
	now      func() time.Time
	capacity int

	mu     sync.Mutex
	times  []time.Time
	values []float64
	next   int // next write position
}

// Observe implements PercentileTracker. It overwrites the slot at next and advances the ring; slots not yet
// written hold the zero time, which Percentile's window filter excludes.
func (t *percentileTracker) Observe(value float64) {
	now := t.now()

	t.mu.Lock()
	defer t.mu.Unlock()

	t.times[t.next] = now
	t.values[t.next] = value
	t.next++
	if t.next == t.capacity {
		t.next = 0
	}
}

// Percentile implements PercentileTracker.
func (t *percentileTracker) Percentile(p float64, window time.Duration) (float64, bool) {
	if p < 0 || p > 100 || math.IsNaN(p) || window <= 0 {
		return 0, false
	}

	cutoff := t.now().Add(-window)

	t.mu.Lock()
	values := make([]float64, 0, len(t.values))
	for i, ts := range t.times {
		if !ts.Before(cutoff) {
			values = append(values, t.values[i])
		}
	}
	t.mu.Unlock()

	if len(values) == 0 {
		return 0, false
	}
	return nearestRankPercentile(values, p), true
}

// nearestRankPercentile returns the p-th percentile of values using the nearest-rank method. values is
// sorted in place and must be non-empty.
func nearestRankPercentile(values []float64, p float64) float64 {
	sort.Float64s(values)
	rank := max(int(math.Ceil(p/100*float64(len(values))))-1, 0)
	if rank >= len(values) {
		rank = len(values) - 1
	}
	return values[rank]
}
