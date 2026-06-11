package stats

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	svcMetric "github.com/rudderlabs/rudder-go-kit/stats/metric"
)

func TestRollingHistogramTrackerFromExponentialDeltas(t *testing.T) {
	now := time.Now()
	tracker := &rollingHistogramTracker{
		window:    time.Minute,
		quantile:  0.95,
		now:       func() time.Time { return now },
		prevByKey: make(map[string]exponentialHistogramSnapshot),
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
	_, err := resolveRollingHistogramConfig(RollingHistogramConfig{})
	require.ErrorContains(t, err, "window must be positive")

	_, err = resolveRollingHistogramConfig(RollingHistogramConfig{
		Window:     time.Minute,
		Percentile: 101,
	})
	require.ErrorContains(t, err, "percentile")

	cfg, err := resolveRollingHistogramConfig(RollingHistogramConfig{Window: time.Minute})
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
			return len(typedTracker.prevByKey) > 0
		}, 5*time.Second, 10*time.Millisecond)

		for i := 0; i < 100; i++ {
			h.Observe(float64(i+1) / 100)
		}
		require.Eventually(t, func() bool {
			return tracker.Count() >= 100
		}, 5*time.Second, 10*time.Millisecond)

		p95, ok := tracker.Percentile()
		require.True(t, ok)
		require.Greater(t, p95, 0.9)
		t.Log("p95 latency:", p95)
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
