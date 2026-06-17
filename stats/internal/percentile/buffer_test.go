package percentile

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBufferPercentile(t *testing.T) {
	b := NewBuffer(0, nil) // default capacity, real clock

	// No observations: any percentile reports no data.
	_, ok := b.Percentile(50, time.Minute)
	require.False(t, ok)

	// Observe out of order to confirm values are sorted before ranking.
	for _, v := range []float64{90, 33, 83, 6, 93, 41, 49, 24, 53, 63, 81, 41, 33, 49, 87, 36, 46, 29, 119, 116} {
		b.Observe(v)
	}

	for _, tc := range []struct {
		p    float64
		want float64
	}{
		{p: 0, want: 6},
		{p: 50, want: 49},
		{p: 95, want: 116},
		{p: 100, want: 119},
	} {
		got, ok := b.Percentile(tc.p, time.Minute)
		require.Truef(t, ok, "p=%v", tc.p)
		require.Equalf(t, tc.want, got, "p=%v", tc.p)
	}

	// Out-of-range / NaN / non-positive window report no data.
	for _, p := range []float64{-1, 101} {
		_, ok := b.Percentile(p, time.Minute)
		require.Falsef(t, ok, "p=%v must be rejected", p)
	}
	_, ok = b.Percentile(50, 0)
	require.False(t, ok, "non-positive window must be rejected")
}

func TestBufferWindowExpiry(t *testing.T) {
	const window = time.Minute
	base := time.Now()
	clock := base
	b := NewBuffer(0, func() time.Time { return clock })

	for range 5 {
		b.Observe(7)
	}
	got, ok := b.Percentile(50, window)
	require.True(t, ok)
	require.Equal(t, 7.0, got)

	// Advance the clock past the window: every observation is now older than now-window.
	clock = base.Add(2 * window)
	_, ok = b.Percentile(50, window)
	require.False(t, ok, "observations should have aged out of the window")
}

func TestBufferCapacityEviction(t *testing.T) {
	b := NewBuffer(3, nil) // keep only the 3 most recent observations
	for i := 1; i <= 10; i++ {
		b.Observe(float64(i))
	}
	// Only 8, 9, 10 remain.
	lo, ok := b.Percentile(0, time.Minute)
	require.True(t, ok)
	require.Equal(t, 8.0, lo)
	hi, ok := b.Percentile(100, time.Minute)
	require.True(t, ok)
	require.Equal(t, 10.0, hi)
}

func TestBufferConcurrent(t *testing.T) {
	b := NewBuffer(0, nil)
	done := make(chan struct{})
	go func() {
		for i := range 1000 {
			b.Observe(float64(i))
		}
		close(done)
	}()
	for range 1000 {
		_, _ = b.Percentile(95, time.Minute)
	}
	<-done
	_, ok := b.Percentile(95, time.Minute)
	require.True(t, ok)
}
