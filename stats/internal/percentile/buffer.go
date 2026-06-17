// Package percentile provides an in-memory, fixed-capacity ring of recent observations used to compute
// rolling-window percentiles for histogram and timer measurements, independently of any metrics backend.
package percentile

import (
	"math"
	"sort"
	"sync"
	"time"
)

// DefaultCapacity is the number of most-recent observations a Buffer retains when no capacity is configured.
const DefaultCapacity = 512

// Buffer is a fixed-capacity ring of recent timestamped observations. Observe appends (evicting the oldest
// once full) and Percentile ranks the observations that fall within a window. It is safe for concurrent use,
// so a single Buffer can back every Measurement of one series across goroutines.
type Buffer struct {
	now      func() time.Time
	capacity int

	mu     sync.Mutex
	times  []time.Time
	values []float64
	next   int // next write position once the ring is full
}

// NewBuffer returns a Buffer retaining up to capacity most-recent observations. A non-positive capacity
// falls back to DefaultCapacity; a nil now falls back to time.Now.
func NewBuffer(capacity int, now func() time.Time) *Buffer {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	if now == nil {
		now = time.Now
	}
	return &Buffer{now: now, capacity: capacity}
}

// Observe records value stamped with the current time. The backing slices grow lazily up to capacity, after
// which the oldest observation is overwritten, so memory is proportional to the observations actually made.
func (b *Buffer) Observe(value float64) {
	t := b.now()

	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.values) < b.capacity {
		b.times = append(b.times, t)
		b.values = append(b.values, value)
		return
	}
	b.times[b.next] = t
	b.values[b.next] = value
	b.next++
	if b.next == b.capacity {
		b.next = 0
	}
}

// Percentile returns the nearest-rank p-th percentile (p in [0,100]) over the observations made within the
// last window, and true when at least one qualifies. It returns (0, false) for an invalid p (outside
// [0,100] or NaN), a non-positive window, or a window holding no observations.
func (b *Buffer) Percentile(p float64, window time.Duration) (float64, bool) {
	if p < 0 || p > 100 || math.IsNaN(p) || window <= 0 {
		return 0, false
	}
	cutoff := b.now().Add(-window)

	b.mu.Lock()
	values := make([]float64, 0, len(b.values))
	for i, t := range b.times {
		if !t.Before(cutoff) {
			values = append(values, b.values[i])
		}
	}
	b.mu.Unlock()

	if len(values) == 0 {
		return 0, false
	}
	return nearestRank(values, p), true
}

// nearestRank returns the p-th percentile of values using the nearest-rank method. values is sorted in
// place and must be non-empty.
func nearestRank(values []float64, p float64) float64 {
	sort.Float64s(values)
	rank := max(int(math.Ceil(p/100*float64(len(values))))-1, 0)
	if rank >= len(values) {
		rank = len(values) - 1
	}
	return values[rank]
}
