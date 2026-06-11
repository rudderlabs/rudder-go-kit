package stats

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	svcMetric "github.com/rudderlabs/rudder-go-kit/stats/metric"
)

func TestRollingHistogramTrackerFromExponentialDeltas(t *testing.T) {
	now := time.Now()
	tracker := &rollingHistogramTracker{
		window:   time.Minute,
		quantile: 0.95,
		now:      func() time.Time { return now },
	}

	tracker.observe(exponentialHistogramSnapshot{
		scale:         0,
		count:         10,
		positiveCount: map[int32]uint64{1: 10},
	}, now)
	require.EqualValues(t, 0, tracker.Count(), "first cumulative snapshot is only the baseline")

	now = now.Add(10 * time.Second)
	tracker.observe(exponentialHistogramSnapshot{
		scale:         0,
		count:         20,
		positiveCount: map[int32]uint64{1: 10, 3: 10},
	}, now)

	require.EqualValues(t, 10, tracker.Count())
	p95, ok := tracker.Percentile()
	require.True(t, ok)
	require.Equal(t, 8.0, p95)

	now = now.Add(2 * time.Minute)
	require.EqualValues(t, 0, tracker.Count())
	_, ok = tracker.Percentile()
	require.False(t, ok)
}

func TestRollingHistogramValidation(t *testing.T) {
	_, err := validateRollingHistogramConfig(RollingHistogramConfig{})
	require.ErrorContains(t, err, "window must be positive")

	_, err = validateRollingHistogramConfig(RollingHistogramConfig{
		Window:     time.Minute,
		Percentile: 101,
	})
	require.ErrorContains(t, err, "percentile")

	cfg, err := validateRollingHistogramConfig(RollingHistogramConfig{Window: time.Minute})
	require.NoError(t, err)
	require.Equal(t, float64(defaultRollingHistogramPercentile), cfg.Percentile)
}

func TestTrackHistogramUnsupportedBackends(t *testing.T) {
	// Backends that do not support rolling histograms (e.g. NOP) degrade gracefully to a no-op
	// tracker instead of panicking.
	tracker, supported, err := TrackHistogram(NOP, "latency", nil, RollingHistogramConfig{Window: time.Minute})
	require.NoError(t, err)
	require.False(t, supported)
	require.NotNil(t, tracker)
	_, ok := tracker.Percentile()
	require.False(t, ok)
	require.EqualValues(t, 0, tracker.Count())

	// OpenTelemetry without the Prometheus exporter is unsupported: rolling histograms read from the
	// Prometheus reader only (we avoid a dedicated internal reader for performance). It must degrade
	// to a no-op tracker rather than panic.
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	s := NewStats(c, logger.NewFactory(c), svcMetric.NewManager())
	tracker, supported, err = TrackHistogram(s, "latency", nil, RollingHistogramConfig{Window: time.Minute})
	require.NoError(t, err)
	require.False(t, supported)
	require.NotNil(t, tracker)
	_, ok = tracker.Percentile()
	require.False(t, ok)
	require.EqualValues(t, 0, tracker.Count())
}

// TestTrackHistogramConfigurations is an end-to-end test (real OTel SDK + Prometheus reader +
// background poller) that exercises one case per histogram-aggregation configuration. All cases run
// with the Prometheus exporter enabled — the only mode that can support rolling histograms; the
// "OTel without Prometheus" combination is covered by TestTrackHistogramUnsupportedBackends.
func TestTrackHistogramConfigurations(t *testing.T) {
	const histogramName = "publish_latency"
	histogramTags := Tags{"dest": "pulsar"}
	explicitBuckets := []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10}

	type testCase struct {
		name              string
		options           []Option
		expectedSupported bool
	}

	runTest := func(t *testing.T, tc testCase) {
		c := config.New()
		c.Set("OpenTelemetry.enabled", true)
		c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
		c.Set("OpenTelemetry.metrics.rollingHistogramPollInterval", 10*time.Millisecond)

		r := prometheus.NewRegistry()
		opts := append([]Option{WithPrometheusRegistry(r, r)}, tc.options...)
		s := NewStats(c, logger.NewFactory(c), svcMetric.NewManager(), opts...)
		require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
		t.Cleanup(s.Stop)

		tracker, supported, err := TrackHistogram(s, histogramName, histogramTags, RollingHistogramConfig{
			Window:     time.Minute,
			Percentile: 95,
		})
		require.NoError(t, err)
		require.Equal(t, tc.expectedSupported, supported)
		require.NotNil(t, tracker)

		h := s.NewTaggedStat(histogramName, HistogramType, histogramTags)

		if !tc.expectedSupported {
			// A no-op tracker never records observations, even though the histogram is still emitted
			// (as a classic histogram, which the poller cannot read).
			for i := 0; i < 100; i++ {
				h.Observe(float64(i+1) / 100)
			}
			require.EqualValues(t, 0, tracker.Count())
			_, ok := tracker.Percentile()
			require.False(t, ok)
			return
		}

		// The first poll records the cumulative baseline without emitting a delta, so observe one
		// value and wait for that baseline before observing the data we want in the rolling window.
		h.Observe(0.1)
		typedTracker := tracker.(*rollingHistogramTracker)
		require.Eventually(t, func() bool {
			typedTracker.mu.Lock()
			defer typedTracker.mu.Unlock()
			return typedTracker.hasPrev
		}, 5*time.Second, 10*time.Millisecond)

		for i := 0; i < 100; i++ {
			h.Observe(float64(i+1) / 100)
		}
		require.Eventually(t, func() bool {
			return tracker.Count() >= 100
		}, 5*time.Second, 10*time.Millisecond)

		// Observations are 0.01..1.00, so the true p95 is ~0.95. With native histograms at full
		// resolution the window settles around scale 4 and the reported value is the containing
		// bucket's upper bound (~0.96) — not the coarse scale-0 value of 1.0.
		p95, ok := tracker.Percentile()
		require.True(t, ok)
		require.InDelta(t, 0.95, p95, 0.03)
	}

	for _, tc := range []testCase{
		{
			name:              "default exponential (native) histogram",
			options:           []Option{WithDefaultExponentialHistogram(160)},
			expectedSupported: true,
		},
		{
			name:              "per-histogram exponential (native) histogram",
			options:           []Option{WithExponentialHistogram(histogramName, 160)},
			expectedSupported: true,
		},
		{
			name: "per-histogram explicit buckets override default exponential",
			options: []Option{
				WithDefaultExponentialHistogram(160),
				WithHistogramBuckets(histogramName, explicitBuckets),
			},
			expectedSupported: false,
		},
		{
			name:              "default explicit buckets",
			options:           []Option{WithDefaultHistogramBuckets(explicitBuckets)},
			expectedSupported: false,
		},
		{
			name:              "no histogram aggregation configured (classic histogram)",
			options:           nil,
			expectedSupported: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runTest(t, tc)
		})
	}
}

func TestExponentialHistogramSnapshotFromDataPoint(t *testing.T) {
	snapshot := exponentialHistogramSnapshotFromDataPoint(metricdata.ExponentialHistogramDataPoint[float64]{
		Scale: 1,
		Count: 3,
		PositiveBucket: metricdata.ExponentialBucket{
			Offset: 1,
			Counts: []uint64{0, 2, 1},
		},
	})

	require.Equal(t, int32(1), snapshot.scale)
	require.EqualValues(t, 3, snapshot.count)
	require.Equal(t, map[int32]uint64{3: 2, 4: 1}, snapshot.positiveCount)
}

func TestRollingHistogramEviction(t *testing.T) {
	now := time.Now()
	clock := func() time.Time { return now }

	// addActiveTracker registers a tracker and feeds it two cumulative snapshots so it holds one
	// delta sample (i.e. it is past warm-up).
	addActiveTracker := func(r *rollingHistogramRegistry, key string) {
		tr := &rollingHistogramTracker{window: time.Minute, now: clock}
		r.entries[key] = tr
		tr.observe(exponentialHistogramSnapshot{scale: 0, count: 10, positiveCount: map[int32]uint64{1: 10}}, now)
		tr.observe(exponentialHistogramSnapshot{scale: 0, count: 20, positiveCount: map[int32]uint64{1: 20}}, now)
	}

	t.Run("warming-up trackers are never evicted", func(t *testing.T) {
		r := newRollingHistogramRegistry(clock, time.Second, 1)
		r.entries["k"] = &rollingHistogramTracker{window: time.Minute, now: clock}
		r.evictEmpty()
		require.Contains(t, r.entries, "k", "a tracker that never held data must not be evicted")
	})

	t.Run("evicted once the window empties (N=1)", func(t *testing.T) {
		now = time.Now()
		r := newRollingHistogramRegistry(clock, time.Second, 1)
		addActiveTracker(r, "k")

		r.evictEmpty() // window still holds the sample
		require.Contains(t, r.entries, "k")

		now = now.Add(2 * time.Minute) // sample falls out of the window
		r.evictEmpty()
		require.NotContains(t, r.entries, "k")
	})

	t.Run("kept until N consecutive empty polls (N=2)", func(t *testing.T) {
		now = time.Now()
		r := newRollingHistogramRegistry(clock, time.Second, 2)
		addActiveTracker(r, "k")

		now = now.Add(2 * time.Minute) // sample falls out of the window
		r.evictEmpty()                 // 1st empty poll
		require.Contains(t, r.entries, "k")
		r.evictEmpty() // 2nd empty poll
		require.NotContains(t, r.entries, "k")
	})
}

func TestExponentialHistogramSnapshotDownscale(t *testing.T) {
	s := exponentialHistogramSnapshot{
		scale:         2,
		negativeCount: 1,
		zeroCount:     2,
		count:         10,
		positiveCount: map[int32]uint64{1: 5, 2: 3, 3: 2}, // keys are OTel bucket index + 1
	}

	got := s.downscale(1)
	require.Equal(t, int32(1), got.scale)
	require.EqualValues(t, 1, got.negativeCount, "non-positive counts are scale-independent")
	require.EqualValues(t, 2, got.zeroCount)
	require.EqualValues(t, 10, got.count)
	// index = key-1 -> 0,1,2; >>1 -> 0,0,1; +1 -> keys 1,1,2 (the first two merge).
	require.Equal(t, map[int32]uint64{1: 8, 2: 2}, got.positiveCount)

	// Downscaling to an equal or higher scale is a no-op.
	require.Equal(t, s.positiveCount, s.downscale(2).positiveCount)
	require.Equal(t, s.positiveCount, s.downscale(5).positiveCount)
}

func TestRollingHistogramDeltaAcrossScaleChange(t *testing.T) {
	now := time.Now()
	tr := &rollingHistogramTracker{window: time.Minute, now: func() time.Time { return now }}

	// Baseline at scale 1.
	tr.observe(exponentialHistogramSnapshot{scale: 1, count: 4, positiveCount: map[int32]uint64{2: 4}}, now)
	require.EqualValues(t, 0, tr.Count())

	// Next cumulative snapshot has downscaled to 0 (as OTel does when data spreads). The delta must be
	// captured (10-4=6), not dropped as a spurious reset.
	now = now.Add(time.Second)
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 10, positiveCount: map[int32]uint64{1: 10}}, now)
	require.EqualValues(t, 6, tr.Count())
}

func TestRollingHistogramCounterReset(t *testing.T) {
	now := time.Now()
	tr := &rollingHistogramTracker{window: time.Minute, now: func() time.Time { return now }}

	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 100, positiveCount: map[int32]uint64{1: 100}}, now)
	now = now.Add(time.Second)
	// Cumulative total dropped -> treated as a reset; the current snapshot becomes the delta.
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 30, positiveCount: map[int32]uint64{1: 30}}, now)
	require.EqualValues(t, 30, tr.Count())
}

func TestQuantileFromExponentialSnapshots(t *testing.T) {
	_, ok := quantileFromExponentialSnapshots(nil, 0.5)
	require.False(t, ok, "empty window has no quantile")

	// 6 of 10 observations are in the zero bucket, so the median falls in the zero/negative region.
	samples := []timedExponentialHistogram{{
		snapshot: exponentialHistogramSnapshot{
			scale: 0, count: 10, zeroCount: 6, positiveCount: map[int32]uint64{1: 4},
		},
	}}
	v, ok := quantileFromExponentialSnapshots(samples, 0.5) // rank 5 <= zeroCount 6
	require.True(t, ok)
	require.Equal(t, 0.0, v)

	// A higher quantile lands in the positive bucket (upper bound 2^1 = 2 at scale 0).
	v, ok = quantileFromExponentialSnapshots(samples, 0.95) // rank 10
	require.True(t, ok)
	require.Equal(t, 2.0, v)
}

func TestRollingHistogramTrackerQuantileInvalidInputs(t *testing.T) {
	now := time.Now()
	tr := &rollingHistogramTracker{window: time.Minute, now: func() time.Time { return now }}
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 10, positiveCount: map[int32]uint64{1: 10}}, now)
	now = now.Add(time.Second)
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 20, positiveCount: map[int32]uint64{1: 20}}, now)
	require.EqualValues(t, 10, tr.Count())

	for _, q := range []float64{-0.1, 1.1, math.NaN()} {
		_, ok := tr.Quantile(q)
		require.Falsef(t, ok, "q=%v must be rejected", q)
	}
	_, ok := tr.Quantile(0.5)
	require.True(t, ok, "valid quantile still works")
}

func TestRollingHistogramReset(t *testing.T) {
	now := time.Now()
	tr := &rollingHistogramTracker{window: time.Minute, now: func() time.Time { return now }}
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 10, positiveCount: map[int32]uint64{1: 10}}, now)
	now = now.Add(time.Second)
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 20, positiveCount: map[int32]uint64{1: 20}}, now)
	require.EqualValues(t, 10, tr.Count())

	tr.Reset()
	require.EqualValues(t, 0, tr.Count())
	_, ok := tr.Percentile()
	require.False(t, ok)

	// After Reset the next observation is a fresh baseline (no delta yet).
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 100, positiveCount: map[int32]uint64{1: 100}}, now)
	require.EqualValues(t, 0, tr.Count())
}

func TestRollingHistogramConcurrentAccess(t *testing.T) {
	tr := &rollingHistogramTracker{window: time.Minute, quantile: 0.95, now: time.Now}

	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Go(func() { // single writer, mirroring the poll goroutine
		defer close(done)
		var c uint64
		for i := 0; i < 2000; i++ {
			c += 10
			tr.observe(
				exponentialHistogramSnapshot{scale: 0, count: c, positiveCount: map[int32]uint64{1: c}},
				time.Now(),
			)
		}
	})

	for i := 0; i < 4; i++ { // concurrent readers
		wg.Go(func() {
			for {
				select {
				case <-done:
					return
				default:
					_, _ = tr.Percentile()
					_ = tr.Count()
				}
			}
		})
	}

	wg.Go(func() { // concurrent resetter
		for {
			select {
			case <-done:
				return
			default:
				tr.Reset()
			}
		}
	})

	wg.Wait()
}

func TestPollExponentialHistogramMultipleSeries(t *testing.T) {
	now := time.Now()
	clock := func() time.Time { return now }
	r := newRollingHistogramRegistry(clock, time.Second, 1)

	attrsA := attribute.NewSet(attribute.String("dest", "a"))
	attrsB := attribute.NewSet(attribute.String("dest", "b"))
	trackerA := &rollingHistogramTracker{window: time.Minute, now: clock}
	trackerB := &rollingHistogramTracker{window: time.Minute, now: clock}
	r.entries[rollingHistogramKey("lat", tagsFromMetricAttributes(attrsA))] = trackerA
	r.entries[rollingHistogramKey("lat", tagsFromMetricAttributes(attrsB))] = trackerB

	dp := func(attrs attribute.Set, count uint64) metricdata.ExponentialHistogramDataPoint[float64] {
		return metricdata.ExponentialHistogramDataPoint[float64]{
			Attributes:     attrs,
			Scale:          0,
			Count:          count,
			PositiveBucket: metricdata.ExponentialBucket{Offset: 0, Counts: []uint64{count}},
		}
	}
	poll := func(a, b uint64) {
		pollExponentialHistogram(r, "lat", metricdata.ExponentialHistogram[float64]{
			DataPoints: []metricdata.ExponentialHistogramDataPoint[float64]{dp(attrsA, a), dp(attrsB, b)},
		}, now)
	}

	poll(10, 5) // baselines only
	require.EqualValues(t, 0, trackerA.Count())
	require.EqualValues(t, 0, trackerB.Count())

	now = now.Add(time.Second)
	poll(30, 8) // each datapoint is routed to its own tracker
	require.EqualValues(t, 20, trackerA.Count())
	require.EqualValues(t, 3, trackerB.Count())
}
