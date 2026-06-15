package stats

import (
	"context"
	"math"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/logger"
	svcMetric "github.com/rudderlabs/rudder-go-kit/stats/metric"
)

// newOTelStats returns a started OpenTelemetry stats instance. The Prometheus exporter is enabled (the
// SDK needs at least one) but no port is set, so no HTTP server runs and the test needs no network;
// percentile tracking is in-process and independent of the export path anyway.
func newOTelStats(t *testing.T) Stats {
	t.Helper()
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	reg := prometheus.NewRegistry()
	s := NewStats(c, logger.NewFactory(c), svcMetric.NewManager(), WithPrometheusRegistry(reg, reg))
	require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
	t.Cleanup(s.Stop)
	return s
}

// TestHistogramPercentile checks the exact percentiles a histogram reports over its rolling window,
// end to end through the public API.
func TestHistogramPercentile(t *testing.T) {
	const window = time.Minute
	h := newOTelStats(t).NewStat("latency", HistogramType)

	// The first call enables tracking; with no observations yet there is no data.
	_, ok := h.Percentile(50, window)
	require.False(t, ok)

	values := []float64{90, 33, 83, 6, 93, 41, 49, 24, 53, 63, 81, 41, 33, 49, 87, 36, 46, 29, 119, 116}
	for _, v := range values {
		h.Observe(v)
	}

	for p, want := range map[float64]float64{0: 6, 50: 49, 95: 116, 100: 119} {
		got, ok := h.Percentile(p, window)
		require.Truef(t, ok, "p=%v", p)
		require.Equalf(t, want, got, "p=%v", p)
	}

	for _, bad := range []float64{-1, 101, math.NaN()} {
		_, ok := h.Percentile(bad, window)
		require.Falsef(t, ok, "p=%v must be rejected", bad)
	}
	_, ok = h.Percentile(50, 0)
	require.False(t, ok, "a non-positive window has no percentile")
}

// TestHistogramPercentileNoCollision proves percentiles don't leak across series: same name with
// different tags, and a different name, all stay independent.
func TestHistogramPercentileNoCollision(t *testing.T) {
	const window = time.Minute
	s := newOTelStats(t)

	a := s.NewTaggedStat("latency", HistogramType, Tags{"dest": "a"}) // same name...
	b := s.NewTaggedStat("latency", HistogramType, Tags{"dest": "b"}) // ...different tag
	other := s.NewStat("size", HistogramType)                         // different name

	for _, m := range []Measurement{a, b, other} {
		_, _ = m.Percentile(95, window) // enable tracking for each series
	}
	for i := 0; i < 5; i++ {
		a.Observe(1)      // latency{dest=a}: only 1s
		b.Observe(100)    // latency{dest=b}: only 100s
		other.Observe(50) // size: only 50s
	}

	requirePercentile := func(m Measurement, want float64) {
		t.Helper()
		got, ok := m.Percentile(95, window)
		require.True(t, ok)
		require.Equal(t, want, got)
	}
	requirePercentile(a, 1)
	requirePercentile(b, 100)
	requirePercentile(other, 50)
}

func TestHistogramPercentileConcurrent(t *testing.T) {
	const window = time.Minute
	h := newOTelStats(t).NewTaggedStat("latency", HistogramType, Tags{"dest": "a"})
	_, _ = h.Percentile(95, window) // enable before the race starts

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Go(func() { // writer
		defer close(done)
		for i := 0; i < 2000; i++ {
			h.Observe(float64(i % 100))
		}
	})
	for i := 0; i < 4; i++ { // concurrent readers
		wg.Go(func() {
			for {
				select {
				case <-done:
					return
				default:
					_, _ = h.Percentile(95, window)
				}
			}
		})
	}
	wg.Wait()
}

func TestHistogramPercentileUnsupportedBackend(t *testing.T) {
	// Backends that cannot track (e.g. NOP) still return a usable Measurement, but Percentile reports
	// no data.
	m := NOP.NewStat("latency", HistogramType)
	require.NotNil(t, m)
	m.Observe(1)
	_, ok := m.Percentile(95, time.Minute)
	require.False(t, ok)
}

func TestNearestRankPercentile(t *testing.T) {
	values := []float64{90, 33, 83, 6, 93, 41, 49, 24, 53, 63, 81, 41, 33, 49, 87, 36, 46, 29, 119, 116}
	cp := func() []float64 { return append([]float64(nil), values...) }
	require.Equal(t, 6.0, nearestRankPercentile(cp(), 0))
	require.Equal(t, 49.0, nearestRankPercentile(cp(), 50))
	require.Equal(t, 116.0, nearestRankPercentile(cp(), 95))
	require.Equal(t, 119.0, nearestRankPercentile(cp(), 100))
	require.Equal(t, 7.0, nearestRankPercentile([]float64{7}, 50), "single value")
}

func TestExemplarValuesSince(t *testing.T) {
	now := time.Now()
	ex := func(ageSec int, v float64) metricdata.Exemplar[float64] {
		return metricdata.Exemplar[float64]{Time: now.Add(-time.Duration(ageSec) * time.Second), Value: v}
	}
	// One data point (a per-series provider only ever has one), exemplars oldest-first.
	dps := []metricdata.ExponentialHistogramDataPoint[float64]{{
		Exemplars: []metricdata.Exemplar[float64]{ex(90, 1), ex(30, 2), ex(10, 3), ex(5, 4)},
	}}

	// Walks newest→oldest and stops at the 90s-old observation, outside a 60s window.
	require.Equal(t, []float64{4, 3, 2}, exemplarValuesSince(dps, now.Add(-time.Minute)))
	// A wider window keeps everything.
	require.Equal(t, []float64{4, 3, 2, 1}, exemplarValuesSince(dps, now.Add(-2*time.Minute)))
	// No data points → no values.
	require.Empty(t, exemplarValuesSince([]metricdata.ExponentialHistogramDataPoint[float64]{}, now.Add(-time.Minute)))
}

func TestWindowReservoir(t *testing.T) {
	r := newWindowReservoir(3)
	base := time.Now()
	offer := func(i int, v float64) {
		r.Offer(context.Background(), base.Add(time.Duration(i)*time.Second), exemplar.NewValue(v), nil)
	}
	vals := func(dest []exemplar.Exemplar) []float64 {
		out := make([]float64, len(dest))
		for i, e := range dest {
			out[i] = e.Value.Float64()
		}
		return out
	}

	var dest []exemplar.Exemplar
	offer(0, 10)
	offer(1, 20)
	r.Collect(&dest)
	require.Equal(t, []float64{10, 20}, vals(dest), "before the ring fills, all offers are kept oldest-first")

	// Capacity is 3; offering 3 more keeps only the most recent 3, still chronological.
	offer(2, 30) // [10,20,30]
	offer(3, 40) // overwrites 10 -> [20,30,40]
	offer(4, 50) // overwrites 20 -> [30,40,50]
	r.Collect(&dest)
	require.Equal(t, []float64{30, 40, 50}, vals(dest))
	require.True(t, dest[0].Time.Before(dest[1].Time) && dest[1].Time.Before(dest[2].Time), "times stay ordered")

	// Collect is non-destructive: reading again yields the same observations.
	r.Collect(&dest)
	require.Equal(t, []float64{30, 40, 50}, vals(dest))
}

func TestWithHistogramPercentileMaxSamples(t *testing.T) {
	// The option sets the corresponding statsConfig field.
	var cfg statsConfig
	WithHistogramPercentileMaxSamples(256)(&cfg)
	require.Equal(t, 256, cfg.histogramPercentileMaxSamples)

	// And it flows through NewStats into the percentile registry, taking precedence over the equivalent
	// config value (256 is neither the default nor the config value below, so it proves the option won).
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.histogramPercentileMaxSamples", 99) // overridden by the option below
	s := NewStats(
		c, logger.NewFactory(c), svcMetric.NewManager(),
		WithHistogramPercentileMaxSamples(256),
	)
	require.Equal(t, 256, s.(*otelStats).percentileRegistry.maxSamples)
}

// TestHistogramPercentileEndToEnd is a full end-to-end test with the real Prometheus exporter on :9102.
// It confirms a histogram is exported normally while its percentile is read in-process, and that the
// percentile empties once the window elapses.
func TestHistogramPercentileEndToEnd(t *testing.T) {
	const (
		window      = time.Second
		metricsURL  = "http://localhost:9102/metrics"
		eventuallyT = 10 * time.Second
		eventuallyI = 20 * time.Millisecond
	)

	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.port", 9102)
	c.Set("RuntimeStats.enabled", false) // keep the exported metric set to exactly what we create

	reg := prometheus.NewRegistry()
	s := NewStats(c, logger.NewFactory(c), svcMetric.NewManager(), WithPrometheusRegistry(reg, reg))
	require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
	t.Cleanup(s.Stop)

	tags := Tags{"foo": "bar"}
	counter := s.NewTaggedStat("rr_counter", CountType, tags)
	gauge := s.NewTaggedStat("rr_gauge", GaugeType, tags)
	histogram := s.NewTaggedStat("rr_histogram", HistogramType, tags)
	// tracked is observed an unpredictable number of times (percentile warm-up below), so it is kept
	// separate from histogram, whose exported count must match the round exactly.
	tracked := s.NewTaggedStat("rr_tracked", HistogramType, tags)

	for round := 1; round <= 10; round++ {
		counter.Increment()
		gauge.Gauge(round)
		histogram.Observe(1)

		// Scrape the real /metrics endpoint until it reflects this round's cumulative values. The HTTP
		// server starts asynchronously, hence require.Eventually.
		require.Eventuallyf(t, func() bool {
			families, err := scrapePrometheus(metricsURL)
			if err != nil {
				return false
			}
			return metricValue(families["rr_counter"], dtoCounterValue) == float64(round) &&
				metricValue(families["rr_gauge"], dtoGaugeValue) == float64(round) &&
				metricValue(families["rr_histogram"], dtoHistogramCount) == float64(round)
		}, eventuallyT, eventuallyI, "prometheus values not correct at round %d", round)

		// Every observation is 1, so the p95 over the window is exactly 1. The first Percentile call
		// enables tracking (reservoir still empty), so keep observing until it shows.
		require.Eventuallyf(t, func() bool {
			tracked.Observe(1)
			p, ok := tracked.Percentile(95, window)
			return ok && p == 1.0
		}, eventuallyT, eventuallyI, "tracked percentile not correct at round %d", round)
	}

	// With no further observations every exemplar ages past the window, so Percentile reports no data.
	require.Eventually(t, func() bool {
		_, ok := tracked.Percentile(95, window)
		return !ok
	}, 5*window, eventuallyI, "percentile should be empty once the window elapses")
}

// scrapePrometheus fetches and parses the Prometheus text exposition from the given URL.
func scrapePrometheus(url string) (map[string]*dto.MetricFamily, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { httputil.CloseResponse(resp) }()
	parser := expfmt.NewTextParser(model.UTF8Validation)
	return parser.TextToMetricFamilies(resp.Body)
}

// metricValue extracts a single float value from the first data point of a metric family using the
// provided accessor; it returns NaN when the family is missing or empty so callers can simply compare.
func metricValue(mf *dto.MetricFamily, accessor func(*dto.Metric) float64) float64 {
	if mf == nil || len(mf.GetMetric()) == 0 {
		return math.NaN()
	}
	return accessor(mf.GetMetric()[0])
}

func dtoCounterValue(m *dto.Metric) float64   { return m.GetCounter().GetValue() }
func dtoGaugeValue(m *dto.Metric) float64     { return m.GetGauge().GetValue() }
func dtoHistogramCount(m *dto.Metric) float64 { return float64(m.GetHistogram().GetSampleCount()) }
