package metric_test

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stats/metric"
)

func TestPercentileTracker(t *testing.T) {
	t.Run("no observations report no data", func(t *testing.T) {
		tr := metric.NewPercentileTracker(0)

		// Valid p and positive window, but nothing observed yet: no slot is written, so there is
		// nothing to rank. Percentile must return (0, false) rather than rank an empty set.
		got, ok := tr.Percentile(50, time.Minute)
		require.False(t, ok)
		require.Zero(t, got)
	})

	t.Run("zero-based clock does not leak unwritten slots", func(t *testing.T) {
		// WithPercentileTrackerNow with a zero-value base (a common test setup) makes the cutoff
		// predate the zero instant. Validity must come from the written count, not from comparing
		// the unwritten slots' zero timestamp against the cutoff — otherwise empty/unwritten slots
		// would leak in as bogus 0.0 values.
		clock := time.Time{}
		tr := metric.NewPercentileTracker(8, metric.WithPercentileTrackerNow(func() time.Time { return clock }))

		_, ok := tr.Percentile(50, time.Minute)
		require.False(t, ok, "empty tracker must report no data even with a zero-based clock")

		// Partially filled: only the two observed values count; the six unwritten slots must not
		// contribute 0.0, so p0 is the smallest observed value rather than 0.
		tr.Observe(5)
		tr.Observe(7)
		got, ok := tr.Percentile(0, time.Minute)
		require.True(t, ok)
		require.Equal(t, 5.0, got, "p0 must be the smallest observed value, not a 0.0 from an unwritten slot")
	})

	t.Run("nearest-rank over the window", func(t *testing.T) {
		tr := metric.NewPercentileTracker(0) // default capacity

		// Observe out of order to confirm values are sorted before ranking.
		for _, v := range []float64{90, 33, 83, 6, 93, 41, 49, 24, 53, 63, 81, 41, 33, 49, 87, 36, 46, 29, 119, 116} {
			tr.Observe(v)
		}
		for _, tc := range []struct{ p, want float64 }{
			{0, 6}, {50, 49}, {95, 116}, {100, 119},
		} {
			got, ok := tr.Percentile(tc.p, time.Minute)
			require.Truef(t, ok, "p=%v", tc.p)
			require.Equalf(t, tc.want, got, "p=%v", tc.p)
		}
	})

	t.Run("invalid p or window report no data", func(t *testing.T) {
		tr := metric.NewPercentileTracker(0)
		tr.Observe(1)
		for _, p := range []float64{-1, 101, math.NaN()} {
			_, ok := tr.Percentile(p, time.Minute)
			require.Falsef(t, ok, "p=%v must be rejected", p)
		}
		_, ok := tr.Percentile(50, 0)
		require.False(t, ok, "non-positive window must be rejected")
		_, ok = tr.Percentile(50, time.Second)
		require.True(t, ok, "valid p and window should be accepted")
	})

	t.Run("observations outside the window are excluded", func(t *testing.T) {
		const window = time.Minute
		clock := time.Now()
		tr := metric.NewPercentileTracker(0, metric.WithPercentileTrackerNow(func() time.Time { return clock }))

		for range 5 {
			tr.Observe(7)
		}
		got, ok := tr.Percentile(50, window)
		require.True(t, ok)
		require.Equal(t, 7.0, got)

		// Advance the clock past the window: every observation is now older than now-window.
		clock = clock.Add(2 * window)
		_, ok = tr.Percentile(50, window)
		require.False(t, ok, "observations should have aged out of the window")
	})

	t.Run("only the in-window batch contributes when the window straddles two batches", func(t *testing.T) {
		const window = time.Minute
		clock := time.Now()
		tr := metric.NewPercentileTracker(0, metric.WithPercentileTrackerNow(func() time.Time { return clock }))

		// First batch, then advance 1.5x the window so it falls just outside it.
		for range 5 {
			tr.Observe(1)
		}
		clock = clock.Add(3 * window / 2)

		// Second batch lands inside the window.
		for range 5 {
			tr.Observe(100)
		}

		// cutoff = now-window sits between the two batches, so only the second (100) qualifies. If the
		// expired first batch leaked in, p0 would be 1 and the median would drop to 1 — this is the mixed
		// case that catches an off-by-one in the cutoff comparison or the ring rotation.
		got0, ok := tr.Percentile(0, window)
		require.True(t, ok)
		require.Equal(t, 100.0, got0, "smallest in-window value must come from the second batch only")
		got50, ok := tr.Percentile(50, window)
		require.True(t, ok)
		require.Equal(t, 100.0, got50, "the expired first batch must not drag the median down")
	})

	t.Run("capacity bounds retained observations", func(t *testing.T) {
		tr := metric.NewPercentileTracker(3) // keep only the 3 most recent
		for i := 1; i <= 10; i++ {
			tr.Observe(float64(i))
		}
		// Only 8, 9, 10 remain.
		lo, ok := tr.Percentile(0, time.Minute)
		require.True(t, ok)
		require.Equal(t, 8.0, lo)
		hi, ok := tr.Percentile(100, time.Minute)
		require.True(t, ok)
		require.Equal(t, 10.0, hi)
	})
}

func TestPercentileTrackerConcurrent(t *testing.T) {
	tr := metric.NewPercentileTracker(0)

	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 1000 {
			tr.Observe(float64(i))
		}
	})
	for range 1000 {
		_, _ = tr.Percentile(95, time.Minute)
	}
	wg.Wait()

	_, ok := tr.Percentile(95, time.Minute)
	require.True(t, ok)
}
