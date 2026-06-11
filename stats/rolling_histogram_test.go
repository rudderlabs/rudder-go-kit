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

func TestTrackHistogramWithOTelPrometheus(t *testing.T) {
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	c.Set("OpenTelemetry.metrics.rollingHistogramPollInterval", 10*time.Millisecond)

	r := prometheus.NewRegistry()
	s := NewStats(
		c,
		logger.NewFactory(c),
		svcMetric.NewManager(),
		WithPrometheusRegistry(r, r),
		WithDefaultExponentialHistogram(160),
	)
	require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
	t.Cleanup(s.Stop)

	tracker, supported, err := TrackHistogram(s, "publish_latency", Tags{"dest": "pulsar"}, RollingHistogramConfig{
		Window:     time.Minute,
		Percentile: 95,
	})
	require.NoError(t, err)
	require.True(t, supported)

	h := s.NewTaggedStat("publish_latency", HistogramType, Tags{"dest": "pulsar"})
	typedTracker := tracker.(*rollingHistogramTracker)

	h.Observe(0.1)
	require.Eventually(t, func() bool {
		typedTracker.mu.Lock()
		defer typedTracker.mu.Unlock()
		return len(typedTracker.prevByKey) > 0
	}, time.Second, 10*time.Millisecond)

	for i := 0; i < 100; i++ {
		h.Observe(float64(i+1) / 100)
	}

	require.Eventually(t, func() bool {
		return tracker.Count() >= 100
	}, time.Second, 10*time.Millisecond)
	p95, ok := tracker.Percentile()
	require.True(t, ok)
	require.Greater(t, p95, 0.9)
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
