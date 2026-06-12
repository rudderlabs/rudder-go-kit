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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/logger"
	svcMetric "github.com/rudderlabs/rudder-go-kit/stats/metric"
)

func TestNearestRankPercentile(t *testing.T) {
	values := []float64{90, 33, 83, 6, 93, 41, 49, 24, 53, 63, 81, 41, 33, 49, 87, 36, 46, 29, 119, 116}
	cp := func() []float64 { return append([]float64(nil), values...) }
	require.Equal(t, 6.0, nearestRankPercentile(cp(), 0))
	require.Equal(t, 49.0, nearestRankPercentile(cp(), 50))
	require.Equal(t, 116.0, nearestRankPercentile(cp(), 95))
	require.Equal(t, 119.0, nearestRankPercentile(cp(), 100))
	require.Equal(t, 7.0, nearestRankPercentile([]float64{7}, 50), "single value")
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

func TestWindowedExemplarValues(t *testing.T) {
	now := time.Now()
	tracker := &rollingHistogramTracker{
		name: "lat",
		key:  rollingHistogramKey("lat", Tags{"dest": "b"}),
		now:  func() time.Time { return now },
	}
	ex := func(ageSec int, v float64) metricdata.Exemplar[float64] {
		return metricdata.Exemplar[float64]{Time: now.Add(-time.Duration(ageSec) * time.Second), Value: v}
	}
	dp := func(dest string, exs ...metricdata.Exemplar[float64]) metricdata.ExponentialHistogramDataPoint[float64] {
		return metricdata.ExponentialHistogramDataPoint[float64]{
			Attributes: attribute.NewSet(attribute.String("dest", dest)),
			Exemplars:  exs,
		}
	}
	rm := metricdata.ResourceMetrics{ScopeMetrics: []metricdata.ScopeMetrics{{
		Metrics: []metricdata.Metrics{
			{Name: "other", Data: metricdata.ExponentialHistogram[float64]{ // wrong name, ignored
				DataPoints: []metricdata.ExponentialHistogramDataPoint[float64]{dp("b", ex(1, 999))},
			}},
			{Name: "lat", Data: metricdata.ExponentialHistogram[float64]{
				DataPoints: []metricdata.ExponentialHistogramDataPoint[float64]{
					dp("a", ex(1, 7)), // wrong tag set, ignored
					dp("b", ex(90, 1), ex(30, 2), ex(10, 3), ex(5, 4)), // oldest-first; 90s is outside the window
				},
			}},
		},
	}}}

	// Walks back from the newest (4,3,2) and stops at the 90s-old observation, which is outside the window.
	require.ElementsMatch(t, []float64{2, 3, 4}, tracker.windowValues(&rm, time.Minute))

	// A series this tracker does not follow yields nothing.
	missing := &rollingHistogramTracker{
		name: "lat", key: rollingHistogramKey("lat", Tags{"dest": "z"}), now: func() time.Time { return now },
	}
	require.Empty(t, missing.windowValues(&rm, time.Minute))
}

// TestHistogramTrackingInMemory drives the real registry pipeline (private meter provider + exemplar
// reservoir + manual reader) without any network, asserting lazy enablement and exact percentiles read
// straight from the recorded observations.
func TestHistogramTrackingInMemory(t *testing.T) {
	reg := newRollingHistogramRegistry(time.Now, 0, nil)
	tracking := reg.tracking("lat", nil)

	// Before the first Percentile call tracking is dormant.
	require.False(t, tracking.enabled.Load())
	_, ok := tracking.percentile(50, time.Minute)
	require.False(t, ok, "first call enables tracking but the reservoir is still empty")
	require.True(t, tracking.enabled.Load())

	values := []float64{90, 33, 83, 6, 93, 41, 49, 24, 53, 63, 81, 41, 33, 49, 87, 36, 46, 29, 119, 116}
	for _, v := range values {
		tracking.instrument.Record(context.Background(), v)
	}

	for p, want := range map[float64]float64{0: 6, 50: 49, 95: 116, 100: 119} {
		got, ok := tracking.percentile(p, time.Minute)
		require.Truef(t, ok, "p=%v", p)
		require.Equalf(t, want, got, "p=%v", p)
	}

	for _, bad := range []float64{-1, 101, math.NaN()} {
		_, ok := tracking.percentile(bad, time.Minute)
		require.Falsef(t, ok, "p=%v must be rejected", bad)
	}
	_, ok = tracking.percentile(50, 0)
	require.False(t, ok, "a non-positive window has no percentile")
}

func TestHistogramTrackingConcurrentAccess(t *testing.T) {
	reg := newRollingHistogramRegistry(time.Now, 0, nil)
	tracking := reg.tracking("lat", nil)
	_, _ = tracking.percentile(95, time.Minute) // enable before the race starts

	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Go(func() { // writer: drive observations through the tracking instrument
		defer close(done)
		for i := 0; i < 2000; i++ {
			tracking.record(context.Background(), float64(i%100))
		}
	})
	for i := 0; i < 4; i++ { // concurrent readers
		wg.Go(func() {
			for {
				select {
				case <-done:
					return
				default:
					_, _ = tracking.percentile(95, time.Minute)
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

func TestWithTrackingHistogramMaxSamples(t *testing.T) {
	// The option sets the corresponding statsConfig field.
	var cfg statsConfig
	WithTrackingHistogramMaxSamples(512)(&cfg)
	require.Equal(t, 512, cfg.trackingHistogramMaxSamples)

	// And it flows through NewStats into the rolling-histogram registry, taking precedence over the
	// equivalent config value.
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.rollingHistogramMaxSamples", 99) // overridden by the option below
	s := NewStats(
		c, logger.NewFactory(c), svcMetric.NewManager(),
		WithTrackingHistogramMaxSamples(512),
	)
	require.Equal(t, 512, s.(*otelStats).rollingHistograms.maxSamples)
}

// TestHistogramPercentileEndToEnd is a full end-to-end test (real OTel SDK + Prometheus exporter serving
// on :9102). It creates a counter, a histogram, a gauge and a (plain) histogram whose percentile it also
// reads. After each round it scrapes the real /metrics HTTP endpoint to verify the exported values and
// checks the percentile (read on demand from the exemplars). Finally it verifies the percentile empties
// once the window elapses.
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
	histogram := s.NewTaggedStat("rr_histogram", HistogramType, tags)
	gauge := s.NewTaggedStat("rr_gauge", GaugeType, tags)
	tracked := s.NewTaggedStat("rr_tracked", HistogramType, tags)

	for round := 1; round <= 10; round++ {
		counter.Increment()
		histogram.Observe(1)
		gauge.Gauge(round)
		tracked.Observe(1)

		// Scrape the real /metrics endpoint until it reflects this round's cumulative values. The HTTP
		// server starts asynchronously, hence require.Eventually.
		require.Eventuallyf(t, func() bool {
			families, err := scrapePrometheus(metricsURL)
			if err != nil {
				return false
			}
			_, trackedExported := families["rr_tracked"]
			return metricValue(families["rr_counter"], dtoCounterValue) == float64(round) &&
				metricValue(families["rr_gauge"], dtoGaugeValue) == float64(round) &&
				metricValue(families["rr_histogram"], dtoHistogramCount) == float64(round) &&
				trackedExported // a tracked histogram is also a normal, exported histogram
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
	}, 5*window, eventuallyI, "tracked percentile should be empty once the window elapses")
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
