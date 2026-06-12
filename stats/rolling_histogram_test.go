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
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/logger"
	svcMetric "github.com/rudderlabs/rudder-go-kit/stats/metric"
)

func TestRollingHistogramTrackerFromExponentialDeltas(t *testing.T) {
	now := time.Now()
	tracker := &rollingHistogramTracker{window: time.Minute, now: func() time.Time { return now }}

	tracker.observe(exponentialHistogramSnapshot{
		scale:         0,
		count:         10,
		positiveCount: map[int32]uint64{1: 10},
	}, now)
	require.EqualValues(t, 0, tracker.count(), "first cumulative snapshot is only the baseline")

	now = now.Add(10 * time.Second)
	tracker.observe(exponentialHistogramSnapshot{
		scale:         0,
		count:         20,
		positiveCount: map[int32]uint64{1: 10, 3: 10},
	}, now)

	require.EqualValues(t, 10, tracker.count())
	p95, ok := tracker.percentile(95)
	require.True(t, ok)
	require.Equal(t, 8.0, p95)

	now = now.Add(2 * time.Minute)
	require.EqualValues(t, 0, tracker.count())
	_, ok = tracker.percentile(95)
	require.False(t, ok)
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
	require.EqualValues(t, 0, tr.count())

	// Next cumulative snapshot has downscaled to 0 (as OTel does when data spreads). The delta must be
	// captured (10-4=6), not dropped as a spurious reset.
	now = now.Add(time.Second)
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 10, positiveCount: map[int32]uint64{1: 10}}, now)
	require.EqualValues(t, 6, tr.count())
}

func TestRollingHistogramCounterReset(t *testing.T) {
	now := time.Now()
	tr := &rollingHistogramTracker{window: time.Minute, now: func() time.Time { return now }}

	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 100, positiveCount: map[int32]uint64{1: 100}}, now)
	now = now.Add(time.Second)
	// Cumulative total dropped -> treated as a reset; the current snapshot becomes the delta.
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 30, positiveCount: map[int32]uint64{1: 30}}, now)
	require.EqualValues(t, 30, tr.count())
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

func TestRollingHistogramTrackerPercentileInvalidInputs(t *testing.T) {
	now := time.Now()
	tr := &rollingHistogramTracker{window: time.Minute, now: func() time.Time { return now }}
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 10, positiveCount: map[int32]uint64{1: 10}}, now)
	now = now.Add(time.Second)
	tr.observe(exponentialHistogramSnapshot{scale: 0, count: 20, positiveCount: map[int32]uint64{1: 20}}, now)
	require.EqualValues(t, 10, tr.count())

	for _, p := range []float64{-1, 101, math.NaN()} {
		_, ok := tr.percentile(p)
		require.Falsef(t, ok, "p=%v must be rejected", p)
	}
	_, ok := tr.percentile(50)
	require.True(t, ok, "a valid percentile still works")
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
	require.EqualValues(t, 0, trackerA.count())
	require.EqualValues(t, 0, trackerB.count())

	now = now.Add(time.Second)
	poll(30, 8) // each datapoint is routed to its own tracker
	require.EqualValues(t, 20, trackerA.count())
	require.EqualValues(t, 3, trackerB.count())
}

func TestRollingHistogramConcurrentAccess(t *testing.T) {
	tr := &rollingHistogramTracker{window: time.Minute, now: time.Now}

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
					_, _ = tr.percentile(95)
					_ = tr.count()
				}
			}
		})
	}

	wg.Wait()
}

func TestNewTrackedHistogramNonOTelBackend(t *testing.T) {
	// Backends that cannot track (e.g. NOP) still return a usable Measurement, but Percentile reports
	// no data.
	m := NOP.NewTrackedHistogram("latency", nil, time.Minute)
	require.NotNil(t, m)
	m.Observe(1)
	_, ok := m.Percentile(95)
	require.False(t, ok)
}

func TestWithTrackingHistogramOptions(t *testing.T) {
	// The options set the corresponding statsConfig fields.
	var cfg statsConfig
	WithTrackingHistogramPollInterval(250 * time.Millisecond)(&cfg)
	WithTrackingHistogramMaxEmptyPolls(7)(&cfg)
	require.Equal(t, 250*time.Millisecond, cfg.trackingHistogramPollInterval)
	require.Equal(t, 7, cfg.trackingHistogramMaxEmptyPolls)

	// And they flow through NewStats into the rolling-histogram registry, taking precedence over the
	// equivalent config values.
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.rollingHistogramPollInterval", time.Hour) // overridden by the option below
	c.Set("OpenTelemetry.metrics.rollingHistogramMaxEmptyPolls", 99)       // overridden by the option below
	s := NewStats(
		c, logger.NewFactory(c), svcMetric.NewManager(),
		WithTrackingHistogramPollInterval(250*time.Millisecond),
		WithTrackingHistogramMaxEmptyPolls(7),
	)
	registry := s.(*otelStats).rollingHistograms
	require.Equal(t, 250*time.Millisecond, registry.pollInterval)
	require.Equal(t, 7, registry.maxEmptyPolls)
}

// TestNewTrackedHistogramRoundRobin is a full end-to-end test (real OTel SDK + Prometheus exporter
// serving on :9102 + background poller). It creates a counter, a histogram, a gauge and a tracked
// histogram, observes them round-robin, and after each round scrapes the real /metrics HTTP endpoint to
// verify the exported values and checks the tracked histogram's percentile. Finally it verifies the
// percentile empties once the window elapses.
func TestNewTrackedHistogramRoundRobin(t *testing.T) {
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
	c.Set("OpenTelemetry.metrics.rollingHistogramPollInterval", 10*time.Millisecond)
	c.Set("RuntimeStats.enabled", false) // keep the exported metric set to exactly what we create

	reg := prometheus.NewRegistry()
	s := NewStats(c, logger.NewFactory(c), svcMetric.NewManager(), WithPrometheusRegistry(reg, reg))
	require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
	t.Cleanup(s.Stop)

	tags := Tags{"foo": "bar"}
	counter := s.NewTaggedStat("rr_counter", CountType, tags)
	histogram := s.NewTaggedStat("rr_histogram", HistogramType, tags)
	gauge := s.NewTaggedStat("rr_gauge", GaugeType, tags)
	tracked := s.NewTrackedHistogram("rr_tracked", tags, window)

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
				!trackedExported // the tracked histogram is in-process only, never exported
		}, eventuallyT, eventuallyI, "prometheus values not correct at round %d", round)

		// The tracker updates asynchronously via the poller, and its first cumulative snapshot is only a
		// baseline. Keep observing 1 until the percentile reflects it; every observation is 1, which lands
		// in the exponential bucket whose upper bound is exactly 1, so the percentile is exactly 1.
		require.Eventuallyf(t, func() bool {
			tracked.Observe(1)
			p, ok := tracked.Percentile(95)
			return ok && p == 1.0
		}, eventuallyT, eventuallyI, "tracked percentile not correct at round %d", round)
	}

	// With no further observations the rolling window empties after `window`, and Percentile reports no
	// data.
	require.Eventually(t, func() bool {
		_, ok := tracked.Percentile(95)
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
