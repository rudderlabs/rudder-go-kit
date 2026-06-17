package stats

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	svcMetric "github.com/rudderlabs/rudder-go-kit/stats/metric"
)

const percentileWindow = time.Minute

// newOTelStats returns a started OpenTelemetry-backed Stats with a Prometheus exporter but no port, so it
// builds a real meter provider without standing up an HTTP server or touching the network.
func newOTelStats(t testing.TB, opts ...Option) Stats {
	t.Helper()
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	reg := prometheus.NewRegistry()
	s := NewStats(c, logger.NewFactory(c), svcMetric.NewManager(), append(opts, WithPrometheusRegistry(reg, reg))...)
	require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
	t.Cleanup(s.Stop)
	return s
}

// newStatsdStats returns a statsd-backed Stats. It is not started: percentiles are tracked in memory on
// Observe/SendTiming independently of statsd connectivity, so no server or Start is needed.
func newStatsdStats(t testing.TB, opts ...Option) Stats {
	t.Helper()
	c := config.New()
	c.Set("RuntimeStats.enabled", false) // OpenTelemetry.enabled defaults to false => statsd backend
	return NewStats(c, logger.NewFactory(c), svcMetric.NewManager(), opts...)
}

// backends returns one started Stats per backend that supports percentile tracking.
func backends(t *testing.T) map[string]Stats {
	return map[string]Stats{"otel": newOTelStats(t), "statsd": newStatsdStats(t)}
}

func TestPercentileHistogram(t *testing.T) {
	for name, s := range backends(t) {
		t.Run(name, func(t *testing.T) {
			h := s.NewTrackedStat("latency", HistogramType, nil)

			_, ok := h.Percentile(50, percentileWindow)
			require.False(t, ok, "no observations yet")

			for _, v := range []float64{90, 33, 83, 6, 93, 41, 49, 24, 53, 63, 81, 41, 33, 49, 87, 36, 46, 29, 119, 116} {
				h.Observe(v)
			}
			for _, tc := range []struct{ p, want float64 }{
				{0, 6}, {50, 49}, {95, 116}, {100, 119},
			} {
				got, ok := h.Percentile(tc.p, percentileWindow)
				require.Truef(t, ok, "p=%v", tc.p)
				require.Equalf(t, tc.want, got, "p=%v", tc.p)
			}
		})
	}
}

func TestPercentileTimer(t *testing.T) {
	// Timers track the percentile over the recorded durations in seconds, on every backend.
	for name, s := range backends(t) {
		t.Run(name, func(t *testing.T) {
			timer := s.NewTrackedStat("duration", TimerType, nil)
			for range 5 {
				timer.SendTiming(2 * time.Second)
			}
			got, ok := timer.Percentile(95, percentileWindow)
			require.True(t, ok)
			require.Equal(t, 2.0, got, "durations are recorded in seconds")
		})
	}
}

func TestPercentileOnlyTrackedStatsRetainObservations(t *testing.T) {
	// Tracking is opt-in: a histogram created via NewStat/NewTaggedStat reports no percentile data even
	// after observations, while the same series via NewTrackedStat does.
	for name, s := range backends(t) {
		t.Run(name, func(t *testing.T) {
			plain := s.NewTaggedStat("requests", HistogramType, Tags{"a": "b"})
			plain.Observe(10)
			plain.Observe(20)
			_, ok := plain.Percentile(50, percentileWindow)
			require.False(t, ok, "untracked histogram must not retain observations")

			tracked := s.NewTrackedStat("requests", HistogramType, Tags{"c": "d"})
			tracked.Observe(10)
			tracked.Observe(20)
			got, ok := tracked.Percentile(100, percentileWindow)
			require.True(t, ok)
			require.Equal(t, 20.0, got)
		})
	}
}

func TestPercentileSharedAcrossLookups(t *testing.T) {
	// Resolving the same tracked series via separate NewTrackedStat calls must share one ring, so
	// observations made through one handle are visible to a percentile read on another.
	for name, s := range backends(t) {
		t.Run(name, func(t *testing.T) {
			tags := Tags{"a": "b"}
			s.NewTrackedStat("shared", HistogramType, tags).Observe(10)
			s.NewTrackedStat("shared", HistogramType, tags).Observe(20)

			got, ok := s.NewTrackedStat("shared", HistogramType, tags).Percentile(100, percentileWindow)
			require.True(t, ok)
			require.Equal(t, 20.0, got)
		})
	}
}

func TestNewTrackedStatPanicsForUnsupportedTypes(t *testing.T) {
	for name, s := range map[string]Stats{"otel": newOTelStats(t), "statsd": newStatsdStats(t), "nop": NOP} {
		t.Run(name, func(t *testing.T) {
			require.Panics(t, func() { s.NewTrackedStat("c", CountType, nil) })
			require.Panics(t, func() { s.NewTrackedStat("g", GaugeType, nil) })
			require.NotPanics(t, func() { s.NewTrackedStat("h", HistogramType, nil) })
			require.NotPanics(t, func() { s.NewTrackedStat("t", TimerType, nil) })
		})
	}
}

func TestPercentileInvalidArguments(t *testing.T) {
	h := newOTelStats(t).NewTrackedStat("latency", HistogramType, nil)
	h.Observe(1)
	for _, p := range []float64{-1, 101} {
		_, ok := h.Percentile(p, percentileWindow)
		require.Falsef(t, ok, "p=%v must be rejected", p)
	}
	_, ok := h.Percentile(50, 0)
	require.False(t, ok, "non-positive window must be rejected")
}

func TestWithHistogramPercentileMaxSamples(t *testing.T) {
	// The option sets the corresponding statsConfig field.
	var cfg statsConfig
	WithHistogramPercentileMaxSamples(256)(&cfg)
	require.Equal(t, 256, cfg.histogramPercentileMaxSamples)

	// And it bounds the ring end to end: with capacity 3, only the last three observations survive, so the
	// minimum is 8 (of 8, 9, 10) rather than 1.
	for name, s := range map[string]Stats{
		"otel":   newOTelStats(t, WithHistogramPercentileMaxSamples(3)),
		"statsd": newStatsdStats(t, WithHistogramPercentileMaxSamples(3)),
	} {
		t.Run(name, func(t *testing.T) {
			h := s.NewTrackedStat("latency", HistogramType, nil)
			for i := 1; i <= 10; i++ {
				h.Observe(float64(i))
			}
			lo, ok := h.Percentile(0, percentileWindow)
			require.True(t, ok)
			require.Equal(t, 8.0, lo, "only the most recent maxSamples observations are retained")
			hi, ok := h.Percentile(100, percentileWindow)
			require.True(t, ok)
			require.Equal(t, 10.0, hi)
		})
	}
}

func TestHistogramPercentileMaxSamplesConfig(t *testing.T) {
	// The config value is read for every backend; here statsd (no OTel) picks it up and bounds the ring.
	c := config.New()
	c.Set("Stats.histogramPercentileMaxSamples", 3)
	s := NewStats(c, logger.NewFactory(c), svcMetric.NewManager())

	h := s.NewTrackedStat("latency", HistogramType, nil)
	for i := 1; i <= 10; i++ {
		h.Observe(float64(i))
	}
	lo, ok := h.Percentile(0, percentileWindow)
	require.True(t, ok)
	require.Equal(t, 8.0, lo)
}

func TestPercentileConcurrent(t *testing.T) {
	h := newOTelStats(t).NewTrackedStat("latency", HistogramType, nil)

	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 1000 {
			h.Observe(float64(i))
		}
	})
	for range 1000 {
		_, _ = h.Percentile(95, percentileWindow)
	}
	wg.Wait()

	_, ok := h.Percentile(95, percentileWindow)
	require.True(t, ok)
}
