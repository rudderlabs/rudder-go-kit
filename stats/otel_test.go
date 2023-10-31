package stats

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ory/dockertest/v3"
	"github.com/prometheus/client_golang/prometheus"
	promClient "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelMetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/logger/mock_logger"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
	statsTest "github.com/rudderlabs/rudder-go-kit/stats/testhelper"
	"github.com/rudderlabs/rudder-go-kit/testhelper"
	"github.com/rudderlabs/rudder-go-kit/testhelper/assert"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker"
	dockerRes "github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

const (
	metricsPort = "8889"
)

var globalDefaultAttrs = []*promClient.LabelPair{
	{Name: ptr("instanceName"), Value: ptr("my-instance-id")},
	{Name: ptr("service_version"), Value: ptr("v1.2.3")},
	{Name: ptr("telemetry_sdk_language"), Value: ptr("go")},
	{Name: ptr("telemetry_sdk_name"), Value: ptr("opentelemetry")},
	{Name: ptr("telemetry_sdk_version"), Value: ptr("1.19.0")},
}

func TestOTelMeasurementInvalidOperations(t *testing.T) {
	s := &otelStats{meter: otel.GetMeterProvider().Meter(t.Name())}

	t.Run("counter invalid operations", func(t *testing.T) {
		require.Panics(t, func() {
			s.NewStat("test", CountType).Gauge(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", CountType).Observe(1.2)
		})
		require.Panics(t, func() {
			s.NewStat("test", CountType).RecordDuration()
		})
		require.Panics(t, func() {
			s.NewStat("test", CountType).SendTiming(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", CountType).Since(time.Now())
		})
	})

	t.Run("gauge invalid operations", func(t *testing.T) {
		require.Panics(t, func() {
			s.NewStat("test", GaugeType).Increment()
		})
		require.Panics(t, func() {
			s.NewStat("test", GaugeType).Count(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", GaugeType).Observe(1.2)
		})
		require.Panics(t, func() {
			s.NewStat("test", GaugeType).RecordDuration()
		})
		require.Panics(t, func() {
			s.NewStat("test", GaugeType).SendTiming(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", GaugeType).Since(time.Now())
		})
	})

	t.Run("histogram invalid operations", func(t *testing.T) {
		require.Panics(t, func() {
			s.NewStat("test", HistogramType).Increment()
		})
		require.Panics(t, func() {
			s.NewStat("test", HistogramType).Count(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", HistogramType).Gauge(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", HistogramType).RecordDuration()
		})
		require.Panics(t, func() {
			s.NewStat("test", HistogramType).SendTiming(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", HistogramType).Since(time.Now())
		})
	})

	t.Run("timer invalid operations", func(t *testing.T) {
		require.Panics(t, func() {
			s.NewStat("test", TimerType).Increment()
		})
		require.Panics(t, func() {
			s.NewStat("test", TimerType).Count(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", TimerType).Gauge(1)
		})
		require.Panics(t, func() {
			s.NewStat("test", TimerType).Observe(1.2)
		})
	})
}

func TestOTelMeasurementOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("counter increment", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewStat("test-counter", CountType).Increment()
		md := getDataPoint[metricdata.Sum[int64]](ctx, t, r, "test-counter", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 1, md.DataPoints[0].Value)
	})

	t.Run("counter count", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewStat("test-counter", CountType).Count(10)
		md := getDataPoint[metricdata.Sum[int64]](ctx, t, r, "test-counter", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 10, md.DataPoints[0].Value)
	})

	t.Run("gauge", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewStat("test-gauge", GaugeType).Gauge(1234)
		md := getDataPoint[metricdata.Gauge[float64]](ctx, t, r, "test-gauge", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 1234, md.DataPoints[0].Value)
	})

	t.Run("tagged gauges", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewTaggedStat("test-tagged-gauge", GaugeType, Tags{"a": "b"}).Gauge(111)
		s.NewTaggedStat("test-tagged-gauge", GaugeType, Tags{"c": "d"}).Gauge(222)
		md := getDataPoint[metricdata.Gauge[float64]](ctx, t, r, "test-tagged-gauge", 0)
		require.Len(t, md.DataPoints, 2)
		// sorting data points by value since the collected time is the same
		sortDataPointsByValue(md.DataPoints)
		require.EqualValues(t, 111, md.DataPoints[0].Value)
		expectedAttrs1 := attribute.NewSet(attribute.String("a", "b"))
		require.True(t, expectedAttrs1.Equals(&md.DataPoints[0].Attributes))
		require.EqualValues(t, 222, md.DataPoints[1].Value)
		expectedAttrs2 := attribute.NewSet(attribute.String("c", "d"))
		require.True(t, expectedAttrs2.Equals(&md.DataPoints[1].Attributes))
	})

	t.Run("timer send timing", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewStat("test-timer-1", TimerType).SendTiming(10 * time.Second)
		md := getDataPoint[metricdata.Histogram[float64]](ctx, t, r, "test-timer-1", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 1, md.DataPoints[0].Count)
		require.InDelta(t, 10.0, md.DataPoints[0].Sum, 0.001)
	})

	t.Run("timer since", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewStat("test-timer-2", TimerType).Since(time.Now().Add(-time.Second))
		md := getDataPoint[metricdata.Histogram[float64]](ctx, t, r, "test-timer-2", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 1, md.DataPoints[0].Count)
		require.InDelta(t, 1.0, md.DataPoints[0].Sum, 0.001)
	})

	t.Run("timer RecordDuration", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		ot := s.NewStat("test-timer-3", TimerType)
		ot.(*otelTimer).now = func() time.Time {
			return time.Now().Add(-time.Second)
		}
		ot.RecordDuration()()
		md := getDataPoint[metricdata.Histogram[float64]](ctx, t, r, "test-timer-3", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 1, md.DataPoints[0].Count)
		require.InDelta(t, 1.0, md.DataPoints[0].Sum, 0.001)
	})

	t.Run("histogram", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewStat("test-hist-1", HistogramType).Observe(1.2)
		md := getDataPoint[metricdata.Histogram[float64]](ctx, t, r, "test-hist-1", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 1, md.DataPoints[0].Count)
		require.EqualValues(t, 1.2, md.DataPoints[0].Sum)
	})

	t.Run("tagged stats", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
		s.NewTaggedStat("test-tagged", CountType, Tags{"key": "value"}).Increment()
		md1 := getDataPoint[metricdata.Sum[int64]](ctx, t, r, "test-tagged", 0)
		require.Len(t, md1.DataPoints, 1)
		require.EqualValues(t, 1, md1.DataPoints[0].Value)
		expectedAttrs := attribute.NewSet(attribute.String("key", "value"))
		require.True(t, expectedAttrs.Equals(&md1.DataPoints[0].Attributes))

		// same measurement name, different measurement type
		s.NewTaggedStat("test-tagged", GaugeType, Tags{"key": "value"}).Gauge(1234)
		md2 := getDataPoint[metricdata.Gauge[float64]](ctx, t, r, "test-tagged", 1)
		require.Len(t, md2.DataPoints, 1)
		require.EqualValues(t, 1234, md2.DataPoints[0].Value)
		require.True(t, expectedAttrs.Equals(&md2.DataPoints[0].Attributes))
	})

	t.Run("measurement with empty name", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, logger: logger.NOP, config: statsConfig{enabled: atomicBool(true)}}
		s.NewStat("", CountType).Increment()
		md := getDataPoint[metricdata.Sum[int64]](ctx, t, r, "novalue", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 1, md.DataPoints[0].Value)
		require.True(t, md.DataPoints[0].Attributes.Equals(newAttributesSet(t)))
	})

	t.Run("measurement with empty name and empty tag key", func(t *testing.T) {
		r, m := newReaderWithMeter(t)
		s := &otelStats{meter: m, logger: logger.NOP, config: statsConfig{enabled: atomicBool(true)}}
		s.NewTaggedStat(" ", GaugeType, Tags{"key": "value", "": "value2", " ": "value3"}).Gauge(22)
		md := getDataPoint[metricdata.Gauge[float64]](ctx, t, r, "novalue", 0)
		require.Len(t, md.DataPoints, 1)
		require.EqualValues(t, 22, md.DataPoints[0].Value)
		require.True(t, md.DataPoints[0].Attributes.Equals(newAttributesSet(t,
			attribute.String("key", "value"),
		)))
	})
}

func TestOTelTaggedGauges(t *testing.T) {
	ctx := context.Background()
	r, m := newReaderWithMeter(t)
	s := &otelStats{meter: m, config: statsConfig{enabled: atomicBool(true)}}
	s.NewTaggedStat("test-gauge", GaugeType, Tags{"a": "b"}).Gauge(1)
	s.NewStat("test-gauge", GaugeType).Gauge(2)
	s.NewTaggedStat("test-gauge", GaugeType, Tags{"c": "d"}).Gauge(3)

	rm := metricdata.ResourceMetrics{}
	err := r.Collect(ctx, &rm)
	require.NoError(t, err)

	var dp []metricdata.DataPoint[float64]
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			dp = append(dp, m.Data.(metricdata.Gauge[float64]).DataPoints...)
		}
	}
	sortDataPointsByValue(dp)

	require.Len(t, dp, 3)

	require.EqualValues(t, 1, dp[0].Value)
	expectedAttrs := attribute.NewSet(attribute.String("a", "b"))
	require.True(t, expectedAttrs.Equals(&dp[0].Attributes))

	require.EqualValues(t, 2, dp[1].Value)
	expectedAttrs = attribute.NewSet()
	require.True(t, expectedAttrs.Equals(&dp[1].Attributes))

	require.EqualValues(t, 3, dp[2].Value)
	expectedAttrs = attribute.NewSet(attribute.String("c", "d"))
	require.True(t, expectedAttrs.Equals(&dp[2].Attributes))
}

func TestOTelPeriodicStats(t *testing.T) {
	type expectation struct {
		name string
		tags []*promClient.LabelPair
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)

	runTest := func(t *testing.T, prepareFunc func(c *config.Config, m metric.Manager), expected []expectation) {
		container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
			filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
		)

		c := config.New()
		c.Set("INSTANCE_ID", "my-instance-id")
		c.Set("OpenTelemetry.enabled", true)
		c.Set("OpenTelemetry.metrics.endpoint", grpcEndpoint)
		c.Set("OpenTelemetry.metrics.exportInterval", time.Millisecond)
		m := metric.NewManager()
		prepareFunc(c, m)

		l := logger.NewFactory(c)
		s := NewStats(c, l, m,
			WithServiceName("TestOTelPeriodicStats"),
			WithServiceVersion("v1.2.3"),
		)

		// start stats
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		require.NoError(t, s.Start(ctx, DefaultGoRoutineFactory))
		defer s.Stop()

		var (
			resp            *http.Response
			metrics         map[string]*promClient.MetricFamily
			metricsEndpoint = fmt.Sprintf("http://localhost:%d/metrics", docker.GetHostPort(t, metricsPort, container))
		)

		require.Eventuallyf(t, func() bool {
			resp, err = http.Get(metricsEndpoint)
			if err != nil {
				return false
			}
			defer func() { httputil.CloseResponse(resp) }()
			metrics, err = statsTest.ParsePrometheusMetrics(resp.Body)
			if err != nil {
				return false
			}
			for _, exp := range expected {
				expectedMetricName := strings.ReplaceAll(exp.name, ".", "_")
				if _, ok := metrics[expectedMetricName]; !ok {
					return false
				}
			}
			return true
		}, 10*time.Second, 100*time.Millisecond, "err: %v, metrics: %+v", err, metrics)

		for _, exp := range expected {
			metricName := strings.ReplaceAll(exp.name, ".", "_")
			require.EqualValues(t, &metricName, metrics[metricName].Name)
			require.EqualValues(t, ptr(promClient.MetricType_GAUGE), metrics[metricName].Type)
			require.Len(t, metrics[metricName].Metric, 1)

			expectedLabels := append(globalDefaultAttrs,
				// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
				&promClient.LabelPair{Name: ptr("label1"), Value: ptr("value1")},
				&promClient.LabelPair{Name: ptr("job"), Value: ptr("TestOTelPeriodicStats")},
				&promClient.LabelPair{Name: ptr("service_name"), Value: ptr("TestOTelPeriodicStats")},
			)
			if exp.tags != nil {
				expectedLabels = append(expectedLabels, exp.tags...)
			}
			require.ElementsMatchf(t, expectedLabels, metrics[metricName].Metric[0].Label,
				"Got %+v", metrics[metricName].Metric[0].Label,
			)
		}
	}

	t.Run("CPU stats", func(t *testing.T) {
		runTest(t, func(c *config.Config, m metric.Manager) {
			c.Set("RuntimeStats.enableCPUStats", true)
			c.Set("RuntimeStats.enabledMemStats", false)
			c.Set("RuntimeStats.enableGCStats", false)
		}, []expectation{
			{name: "runtime_cpu.goroutines"},
			{name: "runtime_cpu.cgo_calls"},
		})
	})

	t.Run("Mem stats", func(t *testing.T) {
		runTest(t, func(c *config.Config, m metric.Manager) {
			c.Set("RuntimeStats.enableCPUStats", false)
			c.Set("RuntimeStats.enabledMemStats", true)
			c.Set("RuntimeStats.enableGCStats", false)
		}, []expectation{
			{name: "runtime_mem.alloc"},
			{name: "runtime_mem.total"},
			{name: "runtime_mem.sys"},
			{name: "runtime_mem.lookups"},
			{name: "runtime_mem.malloc"},
			{name: "runtime_mem.frees"},
			{name: "runtime_mem.heap.alloc"},
			{name: "runtime_mem.heap.sys"},
			{name: "runtime_mem.heap.idle"},
			{name: "runtime_mem.heap.inuse"},
			{name: "runtime_mem.heap.released"},
			{name: "runtime_mem.heap.objects"},
			{name: "runtime_mem.stack.inuse"},
			{name: "runtime_mem.stack.sys"},
			{name: "runtime_mem.stack.mspan_inuse"},
			{name: "runtime_mem.stack.mspan_sys"},
			{name: "runtime_mem.stack.mcache_inuse"},
			{name: "runtime_mem.stack.mcache_sys"},
			{name: "runtime_mem.othersys"},
		})
	})

	t.Run("MemGC stats", func(t *testing.T) {
		runTest(t, func(c *config.Config, m metric.Manager) {
			c.Set("RuntimeStats.enableCPUStats", false)
			c.Set("RuntimeStats.enabledMemStats", true)
			c.Set("RuntimeStats.enableGCStats", true)
		}, []expectation{
			{name: "runtime_mem.alloc"},
			{name: "runtime_mem.total"},
			{name: "runtime_mem.sys"},
			{name: "runtime_mem.lookups"},
			{name: "runtime_mem.malloc"},
			{name: "runtime_mem.frees"},
			{name: "runtime_mem.heap.alloc"},
			{name: "runtime_mem.heap.sys"},
			{name: "runtime_mem.heap.idle"},
			{name: "runtime_mem.heap.inuse"},
			{name: "runtime_mem.heap.released"},
			{name: "runtime_mem.heap.objects"},
			{name: "runtime_mem.stack.inuse"},
			{name: "runtime_mem.stack.sys"},
			{name: "runtime_mem.stack.mspan_inuse"},
			{name: "runtime_mem.stack.mspan_sys"},
			{name: "runtime_mem.stack.mcache_inuse"},
			{name: "runtime_mem.stack.mcache_sys"},
			{name: "runtime_mem.othersys"},
			{name: "runtime_mem.gc.sys"},
			{name: "runtime_mem.gc.next"},
			{name: "runtime_mem.gc.last"},
			{name: "runtime_mem.gc.pause_total"},
			{name: "runtime_mem.gc.pause"},
			{name: "runtime_mem.gc.count"},
			{name: "runtime_mem.gc.cpu_percent"},
		})
	})

	t.Run("Pending events", func(t *testing.T) {
		runTest(t, func(c *config.Config, m metric.Manager) {
			c.Set("RuntimeStats.enableCPUStats", false)
			c.Set("RuntimeStats.enabledMemStats", false)
			c.Set("RuntimeStats.enableGCStats", false)
			m.GetRegistry(metric.PublishedMetrics).MustGetGauge(
				TestMeasurement{tablePrefix: "table", workspace: "workspace", destType: "destType"},
			).Set(1.0)
		}, []expectation{
			{name: "test_measurement_table", tags: []*promClient.LabelPair{
				{Name: ptr("destType"), Value: ptr("destType")},
				{Name: ptr("workspaceId"), Value: ptr("workspace")},
			}},
		})
	})
}

func TestOTelExcludedTags(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)
	container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
		filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
	)

	c := config.New()
	c.Set("INSTANCE_ID", "my-instance-id")
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.endpoint", grpcEndpoint)
	c.Set("OpenTelemetry.metrics.exportInterval", time.Millisecond)
	c.Set("RuntimeStats.enabled", false)
	c.Set("statsExcludedTags", []string{"workspaceId"})
	l := logger.NewFactory(c)
	m := metric.NewManager()
	s := NewStats(c, l, m, WithServiceName(t.Name()), WithServiceVersion("v1.2.3"))

	// start stats
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, s.Start(ctx, DefaultGoRoutineFactory))
	defer s.Stop()

	metricName := "test_workspaceId"
	s.NewTaggedStat(metricName, CountType, Tags{
		"workspaceId":            "nice-value",
		"should_not_be_filtered": "fancy-value",
	}).Increment()

	metricsEndpoint := fmt.Sprintf("http://localhost:%d/metrics", docker.GetHostPort(t, metricsPort, container))
	metrics := requireMetrics(t, metricsEndpoint, metricName)

	require.EqualValues(t, &metricName, metrics[metricName].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics[metricName].Type)
	require.Len(t, metrics[metricName].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(1.0)}, metrics[metricName].Metric[0].Counter)
	require.ElementsMatchf(t, append(globalDefaultAttrs,
		// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
		&promClient.LabelPair{Name: ptr("label1"), Value: ptr("value1")},
		&promClient.LabelPair{Name: ptr("should_not_be_filtered"), Value: ptr("fancy-value")},
		&promClient.LabelPair{Name: ptr("job"), Value: ptr("TestOTelExcludedTags")},
		&promClient.LabelPair{Name: ptr("service_name"), Value: ptr("TestOTelExcludedTags")},
	), metrics[metricName].Metric[0].Label, "Got %+v", metrics[metricName].Metric[0].Label)
}

func TestOTelStartStopError(t *testing.T) {
	c := config.New()
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.enabled", false)
	l := logger.NewFactory(c)
	m := metric.NewManager()
	s := NewStats(c, l, m)

	ctx := context.Background()
	require.Error(t, s.Start(ctx, DefaultGoRoutineFactory),
		"we should error if no endpoint is provided but stats are enabled",
	)

	done := make(chan struct{})
	go func() {
		s.Stop() // this should not panic/block even if we couldn't start
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for Stop()")
	}
}

func TestOTelMeasurementsConsistency(t *testing.T) {
	type testCase struct {
		name               string
		additionalLabels   []*promClient.LabelPair
		setupMeterProvider func(testing.TB) (Stats, string)
	}
	scenarios := []testCase{
		{
			name: "grpc",
			additionalLabels: append(globalDefaultAttrs,
				// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
				&promClient.LabelPair{Name: ptr("label1"), Value: ptr("value1")},
			),
			setupMeterProvider: func(t testing.TB) (Stats, string) {
				cwd, err := os.Getwd()
				require.NoError(t, err)
				container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
					filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
				)

				c := config.New()
				c.Set("INSTANCE_ID", "my-instance-id")
				c.Set("OpenTelemetry.enabled", true)
				c.Set("OpenTelemetry.metrics.endpoint", grpcEndpoint)
				c.Set("OpenTelemetry.metrics.exportInterval", time.Millisecond)
				c.Set("RuntimeStats.enabled", false)
				l := logger.NewFactory(c)
				m := metric.NewManager()
				s := NewStats(c, l, m,
					WithServiceName("TestOTelHistogramBuckets"),
					WithServiceVersion("v1.2.3"),
					WithDefaultHistogramBuckets([]float64{10, 20, 30}),
					WithHistogramBuckets("bar", []float64{40, 50, 60}),
				)
				t.Cleanup(s.Stop)

				return s, fmt.Sprintf("http://localhost:%d/metrics", docker.GetHostPort(t, metricsPort, container))
			},
		},
		{
			name:             "prometheus",
			additionalLabels: globalDefaultAttrs,
			setupMeterProvider: func(t testing.TB) (Stats, string) {
				freePort, err := testhelper.GetFreePort()
				require.NoError(t, err)

				c := config.New()
				c.Set("INSTANCE_ID", "my-instance-id")
				c.Set("OpenTelemetry.enabled", true)
				c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
				c.Set("OpenTelemetry.metrics.prometheus.port", freePort)
				c.Set("OpenTelemetry.metrics.exportInterval", time.Millisecond)
				c.Set("RuntimeStats.enabled", false)
				l := logger.NewFactory(c)
				m := metric.NewManager()
				s := NewStats(c, l, m,
					WithServiceName("TestOTelHistogramBuckets"),
					WithServiceVersion("v1.2.3"),
					WithDefaultHistogramBuckets([]float64{10, 20, 30}),
					WithHistogramBuckets("bar", []float64{40, 50, 60}),
				)
				t.Cleanup(s.Stop)

				return s, fmt.Sprintf("http://localhost:%d/metrics", freePort)
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			s, metricsEndpoint := scenario.setupMeterProvider(t)

			// start stats
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			require.NoError(t, s.Start(ctx, DefaultGoRoutineFactory))
			defer s.Stop()

			s.NewTaggedStat("foo", HistogramType, Tags{"a": "b"}).Observe(20)
			s.NewTaggedStat("bar", HistogramType, Tags{"c": "d"}).Observe(50)
			s.NewTaggedStat("baz", CountType, Tags{"e": "f"}).Count(7)
			s.NewTaggedStat("qux", GaugeType, Tags{"g": "h"}).Gauge(13)
			s.NewTaggedStat("asd", TimerType, Tags{"i": "l"}).SendTiming(20 * time.Second)

			metrics := requireMetrics(t, metricsEndpoint, "foo", "bar", "baz", "qux", "asd")

			require.EqualValues(t, ptr("foo"), metrics["foo"].Name)
			require.EqualValues(t, ptr(promClient.MetricType_HISTOGRAM), metrics["foo"].Type)
			require.Len(t, metrics["foo"].Metric, 1)
			require.EqualValues(t, ptr(uint64(1)), metrics["foo"].Metric[0].Histogram.SampleCount)
			require.EqualValues(t, ptr(20.0), metrics["foo"].Metric[0].Histogram.SampleSum)
			require.EqualValues(t, []*promClient.Bucket{
				{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(20.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
			}, metrics["foo"].Metric[0].Histogram.Bucket)
			require.ElementsMatchf(t, append([]*promClient.LabelPair{
				{Name: ptr("a"), Value: ptr("b")},
				{Name: ptr("job"), Value: ptr("TestOTelHistogramBuckets")},
				{Name: ptr("service_name"), Value: ptr("TestOTelHistogramBuckets")},
			}, scenario.additionalLabels...), metrics["foo"].Metric[0].Label, "Got %+v", metrics["foo"].Metric[0].Label)

			require.EqualValues(t, ptr("bar"), metrics["bar"].Name)
			require.EqualValues(t, ptr(promClient.MetricType_HISTOGRAM), metrics["bar"].Type)
			require.Len(t, metrics["bar"].Metric, 1)
			require.EqualValues(t, ptr(uint64(1)), metrics["bar"].Metric[0].Histogram.SampleCount)
			require.EqualValues(t, ptr(50.0), metrics["bar"].Metric[0].Histogram.SampleSum)
			require.EqualValues(t, []*promClient.Bucket{
				{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(40.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(50.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(60.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
			}, metrics["bar"].Metric[0].Histogram.Bucket)
			require.ElementsMatchf(t, append([]*promClient.LabelPair{
				{Name: ptr("c"), Value: ptr("d")},
				{Name: ptr("job"), Value: ptr("TestOTelHistogramBuckets")},
				{Name: ptr("service_name"), Value: ptr("TestOTelHistogramBuckets")},
			}, scenario.additionalLabels...), metrics["bar"].Metric[0].Label, "Got %+v", metrics["bar"].Metric[0].Label)

			require.EqualValues(t, ptr("baz"), metrics["baz"].Name)
			require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics["baz"].Type)
			require.Len(t, metrics["baz"].Metric, 1)
			require.EqualValues(t, ptr(7.0), metrics["baz"].Metric[0].Counter.Value)
			require.ElementsMatchf(t, append([]*promClient.LabelPair{
				{Name: ptr("e"), Value: ptr("f")},
				{Name: ptr("job"), Value: ptr("TestOTelHistogramBuckets")},
				{Name: ptr("service_name"), Value: ptr("TestOTelHistogramBuckets")},
			}, scenario.additionalLabels...), metrics["baz"].Metric[0].Label, "Got %+v", metrics["baz"].Metric[0].Label)

			require.EqualValues(t, ptr("qux"), metrics["qux"].Name)
			require.EqualValues(t, ptr(promClient.MetricType_GAUGE), metrics["qux"].Type)
			require.Len(t, metrics["qux"].Metric, 1)
			require.EqualValues(t, ptr(13.0), metrics["qux"].Metric[0].Gauge.Value)
			require.ElementsMatchf(t, append([]*promClient.LabelPair{
				{Name: ptr("g"), Value: ptr("h")},
				{Name: ptr("job"), Value: ptr("TestOTelHistogramBuckets")},
				{Name: ptr("service_name"), Value: ptr("TestOTelHistogramBuckets")},
			}, scenario.additionalLabels...), metrics["qux"].Metric[0].Label, "Got %+v", metrics["qux"].Metric[0].Label)

			require.EqualValues(t, ptr("asd"), metrics["asd"].Name)
			require.EqualValues(t, ptr(promClient.MetricType_HISTOGRAM), metrics["asd"].Type)
			require.Len(t, metrics["asd"].Metric, 1)
			require.EqualValues(t, []*promClient.Bucket{
				{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(20.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
			}, metrics["asd"].Metric[0].Histogram.Bucket)
			require.ElementsMatchf(t, append([]*promClient.LabelPair{
				{Name: ptr("i"), Value: ptr("l")},
				{Name: ptr("job"), Value: ptr("TestOTelHistogramBuckets")},
				{Name: ptr("service_name"), Value: ptr("TestOTelHistogramBuckets")},
			}, scenario.additionalLabels...), metrics["asd"].Metric[0].Label, "Got %+v", metrics["asd"].Metric[0].Label)
		})
	}
}

func TestPrometheusCustomRegistry(t *testing.T) {
	metricName := "foo"
	setup := func(t testing.TB) (prometheus.Registerer, int) {
		freePort, err := testhelper.GetFreePort()
		require.NoError(t, err)

		c := config.New()
		c.Set("INSTANCE_ID", "my-instance-id")
		c.Set("OpenTelemetry.enabled", true)
		c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
		c.Set("OpenTelemetry.metrics.prometheus.port", freePort)
		c.Set("OpenTelemetry.metrics.exportInterval", time.Millisecond)
		c.Set("RuntimeStats.enabled", false)
		l := logger.NewFactory(c)
		m := metric.NewManager()
		r := prometheus.NewRegistry()
		s := NewStats(c, l, m,
			WithServiceName("TestPrometheusCustomRegistry"),
			WithServiceVersion("v1.2.3"),
			WithPrometheusRegistry(r, r),
		)
		require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
		t.Cleanup(s.Stop)

		s.NewTaggedStat(metricName, CountType, Tags{"a": "b"}).Count(7)

		return r, freePort
	}

	t.Run("http", func(t *testing.T) {
		var (
			err             error
			resp            *http.Response
			metrics         map[string]*promClient.MetricFamily
			_, serverPort   = setup(t)
			metricsEndpoint = fmt.Sprintf("http://localhost:%d/metrics", serverPort)
		)
		require.Eventuallyf(t, func() bool {
			resp, err = http.Get(metricsEndpoint)
			if err != nil {
				return false
			}
			defer func() { httputil.CloseResponse(resp) }()
			metrics, err = statsTest.ParsePrometheusMetrics(resp.Body)
			if err != nil {
				return false
			}
			if _, ok := metrics[metricName]; !ok {
				return false
			}
			return true
		}, 10*time.Second, 100*time.Millisecond, "err: %v, metrics: %+v", err, metrics)

		require.EqualValues(t, &metricName, metrics[metricName].Name)
		require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics[metricName].Type)
		require.Len(t, metrics[metricName].Metric, 1)
		require.EqualValues(t, &promClient.Counter{Value: ptr(7.0)}, metrics[metricName].Metric[0].Counter)
		require.ElementsMatchf(t, append(globalDefaultAttrs,
			&promClient.LabelPair{Name: ptr("a"), Value: ptr("b")},
			&promClient.LabelPair{Name: ptr("job"), Value: ptr("TestPrometheusCustomRegistry")},
			&promClient.LabelPair{Name: ptr("service_name"), Value: ptr("TestPrometheusCustomRegistry")},
		), metrics[metricName].Metric[0].Label, "Got %+v", metrics[metricName].Metric[0].Label)
	})

	t.Run("collector", func(t *testing.T) {
		r, _ := setup(t)
		metrics, err := r.(prometheus.Gatherer).Gather()
		require.NoError(t, err)

		var mf *promClient.MetricFamily
		for _, m := range metrics {
			if m.GetName() == metricName {
				mf = m
				break
			}
		}
		require.NotNilf(t, mf, "Metric not found in %+v", metrics)
		require.EqualValues(t, metricName, mf.GetName())
		require.EqualValues(t, promClient.MetricType_COUNTER, mf.GetType())
		require.Len(t, mf.GetMetric(), 1)
		require.ElementsMatch(t, append(globalDefaultAttrs,
			&promClient.LabelPair{Name: ptr("a"), Value: ptr("b")},
			&promClient.LabelPair{Name: ptr("job"), Value: ptr("TestPrometheusCustomRegistry")},
			&promClient.LabelPair{Name: ptr("service_name"), Value: ptr("TestPrometheusCustomRegistry")},
		), mf.GetMetric()[0].GetLabel())
		require.EqualValues(t, ptr(7.0), mf.GetMetric()[0].GetCounter().Value)
	})
}

func TestPrometheusDuplicatedAttributes(t *testing.T) {
	freePort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	c := config.New()
	c.Set("INSTANCE_ID", "my-instance-id")
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.port", freePort)
	c.Set("OpenTelemetry.metrics.exportInterval", time.Millisecond)
	c.Set("RuntimeStats.enabled", false)
	ctrl := gomock.NewController(t)
	loggerSpy := mock_logger.NewMockLogger(ctrl)
	loggerSpy.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	loggerSpy.EXPECT().Warnf(
		"removing tag %q for measurement %q since it is a resource attribute",
		"instanceName", "foo",
	).Times(1)
	loggerFactory := mock_logger.NewMockLogger(ctrl)
	loggerFactory.EXPECT().Child(gomock.Any()).Times(1).Return(loggerSpy)
	l := newLoggerSpyFactory(loggerFactory)
	m := metric.NewManager()
	r := prometheus.NewRegistry()
	s := NewStats(c, l, m,
		WithServiceName(t.Name()),
		WithServiceVersion("v1.2.3"),
		WithPrometheusRegistry(r, r),
	)
	require.NoError(t, s.Start(context.Background(), DefaultGoRoutineFactory))
	t.Cleanup(s.Stop)

	metricName := "foo"
	s.NewTaggedStat(metricName, CountType, Tags{"a": "b", "instanceName": "from-metric"}).Count(7)

	var (
		resp            *http.Response
		metrics         map[string]*promClient.MetricFamily
		metricsEndpoint = fmt.Sprintf("http://localhost:%d/metrics", freePort)
	)
	require.Eventuallyf(t, func() bool {
		resp, err = http.Get(metricsEndpoint)
		if err != nil {
			return false
		}
		defer func() { httputil.CloseResponse(resp) }()
		metrics, err = statsTest.ParsePrometheusMetrics(resp.Body)
		if err != nil {
			return false
		}
		if _, ok := metrics[metricName]; !ok {
			return false
		}
		return true
	}, 10*time.Second, 100*time.Millisecond, "err: %v, metrics: %+v", err, metrics)

	require.EqualValues(t, &metricName, metrics[metricName].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics[metricName].Type)
	require.Len(t, metrics[metricName].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(7.0)}, metrics[metricName].Metric[0].Counter)
	require.ElementsMatchf(t, append(globalDefaultAttrs,
		&promClient.LabelPair{Name: ptr("a"), Value: ptr("b")},
		&promClient.LabelPair{Name: ptr("job"), Value: ptr(t.Name())},
		&promClient.LabelPair{Name: ptr("service_name"), Value: ptr(t.Name())},
	), metrics[metricName].Metric[0].Label, "Got %+v", metrics[metricName].Metric[0].Label)
}

func TestZipkin(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	zipkin, err := dockerRes.SetupZipkin(pool, t)
	require.NoError(t, err)

	prometheusPort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	zipkinURL := "http://localhost:" + zipkin.Port + "/api/v2/spans"

	c := config.New()
	c.Set("INSTANCE_ID", t.Name())
	c.Set("OpenTelemetry.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	c.Set("OpenTelemetry.traces.endpoint", zipkinURL)
	c.Set("OpenTelemetry.traces.samplingRate", 1.0)
	c.Set("OpenTelemetry.traces.withSyncer", true)
	c.Set("OpenTelemetry.traces.withZipkin", true)
	// @TODO remove the fact that the metrics have to be up and we can't use traces alone
	c.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	c.Set("OpenTelemetry.metrics.prometheus.port", prometheusPort)
	l := logger.NewFactory(c)
	m := metric.NewManager()
	s := NewStats(c, l, m, WithServiceName(t.Name()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, s.Start(ctx, DefaultGoRoutineFactory))
	t.Cleanup(s.Stop)

	_, span := s.NewTracer("my-tracer").Start(
		ctx, "my-span", SpanKindServer, time.Now(), Tags{"foo": "bar"},
	)
	time.Sleep(123 * time.Millisecond)
	span.End()

	zipkinSpansURL := zipkinURL + "?serviceName=" + t.Name()
	getSpansReq, err := http.NewRequest(http.MethodGet, zipkinSpansURL, nil)
	require.NoError(t, err)

	spansBody := assert.RequireEventuallyResponse(
		t, http.StatusOK, getSpansReq, 10*time.Second, 100*time.Millisecond,
	)
	require.Equal(t, `["my-span"]`, spansBody)
}

func getDataPoint[T any](ctx context.Context, t *testing.T, rdr sdkmetric.Reader, name string, idx int) (zero T) {
	t.Helper()
	rm := metricdata.ResourceMetrics{}
	err := rdr.Collect(ctx, &rm)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rm.ScopeMetrics), 1)
	require.GreaterOrEqual(t, len(rm.ScopeMetrics[0].Metrics), idx+1)
	require.Equal(t, name, rm.ScopeMetrics[0].Metrics[idx].Name)
	md, ok := rm.ScopeMetrics[0].Metrics[idx].Data.(T)
	require.Truef(t, ok, "Metric data is not of type %T but %T", zero, rm.ScopeMetrics[0].Metrics[idx].Data)
	return md
}

func sortDataPointsByValue[N int64 | float64](dp []metricdata.DataPoint[N]) {
	sort.Slice(dp, func(i, j int) bool {
		return dp[i].Value < dp[j].Value
	})
}

func newAttributesSet(t *testing.T, attrs ...attribute.KeyValue) *attribute.Set {
	t.Helper()
	set := attribute.NewSet(attrs...)
	return &set
}

func newReaderWithMeter(t *testing.T) (sdkmetric.Reader, otelMetric.Meter) {
	t.Helper()
	manualRdr := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(t.Name()))),
		sdkmetric.WithReader(manualRdr),
	)
	t.Cleanup(func() {
		_ = meterProvider.Shutdown(context.Background())
	})
	return manualRdr, meterProvider.Meter(t.Name())
}

func requireMetrics(
	t *testing.T, metricsEndpoint string, requiredKeys ...string,
) map[string]*promClient.MetricFamily {
	t.Helper()

	var (
		err     error
		resp    *http.Response
		metrics map[string]*promClient.MetricFamily
	)
	require.Eventuallyf(t, func() bool {
		resp, err = http.Get(metricsEndpoint)
		if err != nil {
			return false
		}
		defer func() { httputil.CloseResponse(resp) }()
		metrics, err = statsTest.ParsePrometheusMetrics(resp.Body)
		if err != nil {
			return false
		}
		for _, k := range requiredKeys {
			if _, ok := metrics[k]; !ok {
				return false
			}
		}
		return true
	}, 5*time.Second, 100*time.Millisecond, "err: %v, metrics: %+v", err, metrics)

	return metrics
}

func ptr[T any](v T) *T {
	return &v
}

func atomicBool(b bool) *atomic.Bool { // nolint:unparam
	a := atomic.Bool{}
	a.Store(b)
	return &a
}

type TestMeasurement struct {
	tablePrefix string
	workspace   string
	destType    string
}

func (r TestMeasurement) GetName() string {
	return fmt.Sprintf("test_measurement_%s", r.tablePrefix)
}

func (r TestMeasurement) GetTags() map[string]string {
	return map[string]string{
		"workspaceId": r.workspace,
		"destType":    r.destType,
	}
}

type loggerSpyFactory struct{ spy logger.Logger }

func (f loggerSpyFactory) NewLogger() logger.Logger { return f.spy }

func newLoggerSpyFactory(l logger.Logger) loggerFactory {
	return &loggerSpyFactory{spy: l}
}
