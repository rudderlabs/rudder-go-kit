package otel

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	promClient "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	statsTest "github.com/rudderlabs/rudder-go-kit/stats/testhelper"
	"github.com/rudderlabs/rudder-go-kit/testhelper"
	dt "github.com/rudderlabs/rudder-go-kit/testhelper/docker"
)

const (
	metricsPort = "8889"
)

var (
	globalDefaultAttrs = []*promClient.LabelPair{
		{Name: ptr("service_version"), Value: ptr("v1.2.3")},
		{Name: ptr("telemetry_sdk_language"), Value: ptr("go")},
		{Name: ptr("telemetry_sdk_name"), Value: ptr("opentelemetry")},
		{Name: ptr("telemetry_sdk_version"), Value: ptr(otel.Version())},
		{Name: ptr("instanceName"), Value: ptr("my-instance-id")},
	}
	globalGRPCDefaultAttrs = append(globalDefaultAttrs,
		// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
		&promClient.LabelPair{Name: ptr("label1"), Value: ptr("value1")},
	)
)

// see https://opentelemetry.io/docs/collector/getting-started/
func TestMetrics(t *testing.T) {
	var (
		ctx       = context.Background()
		meterName = "some-meter-name"
		svcName   = "TestMetrics"
	)
	scenarios := []testCase{
		{
			name:             "grpc",
			additionalLabels: globalGRPCDefaultAttrs,
			counterSuffix:    "_total", // OTel Collector Prometheus exporter adds _total suffix
			setupMeterProvider: func(t testing.TB, _ ...MeterProviderOption) (*sdkmetric.MeterProvider, string) {
				cwd, err := os.Getwd()
				require.NoError(t, err)
				container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
					filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
				)

				res, err := NewResource(svcName, "v1.2.3",
					attribute.String("instanceName", "my-instance-id"),
				)
				require.NoError(t, err)
				var om Manager
				_, mp, err := om.Setup(ctx, res,
					WithInsecure(),
					WithMeterProvider(
						WithGRPCMeterProvider(grpcEndpoint),
						WithMeterProviderExportsInterval(50*time.Millisecond),
						WithDefaultHistogramBucketBoundaries([]float64{1, 2, 3}),
						WithHistogramBucketBoundaries("baz", meterName, []float64{10, 20, 30}),
					),
				)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
				require.NotEqual(t, mp, otel.GetMeterProvider())

				metricsEndpoint := fmt.Sprintf("http://localhost:%d/metrics", dt.GetHostPort(t, metricsPort, container))
				return mp, metricsEndpoint
			},
		},
		{
			name:             "prometheus",
			additionalLabels: globalDefaultAttrs,
			counterSuffix:    "", // Direct Prometheus exporter doesn't add _total suffix
			setupMeterProvider: func(t testing.TB, _ ...MeterProviderOption) (*sdkmetric.MeterProvider, string) {
				registry := prometheus.NewRegistry()

				res, err := NewResource(svcName, "v1.2.3",
					attribute.String("instanceName", "my-instance-id"),
				)
				require.NoError(t, err)
				var om Manager
				tp, mp, err := om.Setup(ctx, res,
					WithInsecure(),
					WithMeterProvider(
						WithPrometheusExporter(registry),
						WithMeterProviderExportsInterval(100*time.Millisecond),
						WithDefaultHistogramBucketBoundaries([]float64{1, 2, 3}),
						WithHistogramBucketBoundaries("baz", meterName, []float64{10, 20, 30}),
					),
				)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
				require.Nil(t, tp)
				require.NotEqual(t, mp, otel.GetMeterProvider())

				ts := httptest.NewServer(promhttp.InstrumentMetricHandler(
					registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
				))
				t.Cleanup(ts.Close)

				return mp, ts.URL
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			mp, metricsEndpoint := scenario.setupMeterProvider(t)
			m := mp.Meter(meterName)
			// foo counter
			counter, err := m.Int64Counter("foo")
			require.NoError(t, err)
			counter.Add(ctx, 1, metric.WithAttributes(attribute.String("hello", "world")))
			// bar counter
			counter, err = m.Int64Counter("bar")
			require.NoError(t, err)
			counter.Add(ctx, 5)
			// baz histogram
			h1, err := m.Int64Histogram("baz")
			require.NoError(t, err)
			h1.Record(ctx, 20, metric.WithAttributes(attribute.String("a", "b")))
			// qux histogram
			h2, err := m.Int64Histogram("qux")
			require.NoError(t, err)
			h2.Record(ctx, 2, metric.WithAttributes(attribute.String("c", "d")))

			// Run assertions
			// Note: Counter metrics may have _total suffix depending on exporter
			fooName := "foo" + scenario.counterSuffix
			barName := "bar" + scenario.counterSuffix
			metrics := requireMetrics(t, metricsEndpoint, fooName, barName, "baz", "qux")

			require.EqualValues(t, ptr(fooName), metrics[fooName].Name)
			require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics[fooName].Type)
			require.Len(t, metrics[fooName].Metric, 1)
			require.EqualValues(t, &promClient.Counter{Value: ptr(1.0)}, metrics[fooName].Metric[0].Counter)
			require.ElementsMatch(t, append(
				scenario.additionalLabels,
				&promClient.LabelPair{Name: ptr("hello"), Value: ptr("world")},
				&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
				&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
			), metrics[fooName].Metric[0].Label)

			require.EqualValues(t, ptr(barName), metrics[barName].Name)
			require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics[barName].Type)
			require.Len(t, metrics[barName].Metric, 1)
			require.EqualValues(t, &promClient.Counter{Value: ptr(5.0)}, metrics[barName].Metric[0].Counter)
			require.ElementsMatch(t, append(
				scenario.additionalLabels,
				&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
				&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
			), metrics[barName].Metric[0].Label)

			requireHistogramEqual(t, metrics["baz"], histogram{
				name: "baz", count: 1, sum: 20,
				buckets: []*promClient.Bucket{
					{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
					{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(20.0)},
					{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
					{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
				},
				labels: append(
					scenario.additionalLabels,
					&promClient.LabelPair{Name: ptr("a"), Value: ptr("b")},
					&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
					&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
				),
			})

			requireHistogramEqual(t, metrics["qux"], histogram{
				name: "qux", count: 1, sum: 2,
				buckets: []*promClient.Bucket{
					{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(1.0)},
					{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(2.0)},
					{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(3.0)},
					{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
				},
				labels: append(
					scenario.additionalLabels,
					&promClient.LabelPair{Name: ptr("c"), Value: ptr("d")},
					&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
					&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
				),
			})
		})
	}
}

func TestHistogramBuckets(t *testing.T) {
	var (
		ctx     = context.Background()
		svcName = "TestHistogramBuckets"
	)
	scenarios := []testCase{
		{
			name:             "grpc",
			additionalLabels: globalGRPCDefaultAttrs,
			setupMeterProvider: func(t testing.TB, opts ...MeterProviderOption) (*sdkmetric.MeterProvider, string) {
				cwd, err := os.Getwd()
				require.NoError(t, err)
				container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
					filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
				)

				res, err := NewResource(svcName, "v1.2.3", attribute.String("instanceName", "my-instance-id"))
				require.NoError(t, err)
				var om Manager
				_, mp, err := om.Setup(ctx, res,
					WithInsecure(),
					WithMeterProvider(append(opts,
						WithGRPCMeterProvider(grpcEndpoint),
						WithMeterProviderExportsInterval(50*time.Millisecond),
					)...),
				)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
				require.NotEqual(t, mp, otel.GetMeterProvider())

				metricsEndpoint := fmt.Sprintf("http://localhost:%d/metrics", dt.GetHostPort(t, metricsPort, container))
				return mp, metricsEndpoint
			},
		},
		{
			name:             "prometheus",
			additionalLabels: globalDefaultAttrs,
			setupMeterProvider: func(t testing.TB, opts ...MeterProviderOption) (*sdkmetric.MeterProvider, string) {
				registry := prometheus.NewRegistry()

				res, err := NewResource(svcName, "v1.2.3", attribute.String("instanceName", "my-instance-id"))
				require.NoError(t, err)
				var om Manager
				tp, mp, err := om.Setup(ctx, res,
					WithInsecure(),
					WithMeterProvider(append(opts,
						WithPrometheusExporter(registry),
						WithMeterProviderExportsInterval(50*time.Millisecond),
					)...),
				)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
				require.Nil(t, tp)
				require.NotEqual(t, mp, otel.GetMeterProvider())

				ts := httptest.NewServer(promhttp.InstrumentMetricHandler(
					registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
				))
				t.Cleanup(ts.Close)

				return mp, ts.URL
			},
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Run("default applies to all meters", func(t *testing.T) {
				mp, metricsEndpoint := scenario.setupMeterProvider(t,
					WithDefaultHistogramBucketBoundaries([]float64{10, 20, 30}),
				)

				// foo histogram on meter-1
				h, err := mp.Meter("meter-1").Int64Histogram("foo")
				require.NoError(t, err)
				h.Record(ctx, 20, metric.WithAttributes(attribute.String("a", "b")))

				// bar histogram on meter-2
				h, err = mp.Meter("meter-2").Int64Histogram("bar")
				require.NoError(t, err)
				h.Record(ctx, 30, metric.WithAttributes(attribute.String("c", "d")))

				metrics := requireMetrics(t, metricsEndpoint, "foo", "bar")

				requireHistogramEqual(t, metrics["foo"], histogram{
					name: "foo", count: 1, sum: 20,
					buckets: []*promClient.Bucket{
						{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(20.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
					},
					labels: append(
						scenario.additionalLabels,
						&promClient.LabelPair{Name: ptr("a"), Value: ptr("b")},
						&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
						&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
					),
				})

				requireHistogramEqual(t, metrics["bar"], histogram{
					name: "bar", count: 1, sum: 30,
					buckets: []*promClient.Bucket{
						{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
						{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(20.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
					},
					labels: append(
						scenario.additionalLabels,
						&promClient.LabelPair{Name: ptr("c"), Value: ptr("d")},
						&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
						&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
					),
				})
			})

			t.Run("custom boundaries do not override default ones", func(t *testing.T) {
				mp, metricsEndpoint := scenario.setupMeterProvider(t,
					WithDefaultHistogramBucketBoundaries([]float64{10, 20, 30}),
					WithHistogramBucketBoundaries("bar", "meter-1", []float64{40, 50, 60}),
					WithHistogramBucketBoundaries("baz", "meter-1", []float64{70, 80, 90}),
				)

				// foo histogram
				h, err := mp.Meter("meter-1").Int64Histogram("foo")
				require.NoError(t, err)
				h.Record(ctx, 20, metric.WithAttributes(attribute.String("a", "b")))

				// bar histogram
				h, err = mp.Meter("meter-1").Int64Histogram("bar")
				require.NoError(t, err)
				h.Record(ctx, 50, metric.WithAttributes(attribute.String("c", "d")))

				// baz histogram
				h, err = mp.Meter("meter-1").Int64Histogram("baz")
				require.NoError(t, err)
				h.Record(ctx, 80, metric.WithAttributes(attribute.String("e", "f")))

				metrics := requireMetrics(t, metricsEndpoint, "foo", "bar", "baz")

				requireHistogramEqual(t, metrics["foo"], histogram{
					name: "foo", count: 1, sum: 20,
					buckets: []*promClient.Bucket{
						{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(20.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
					},
					labels: append(
						scenario.additionalLabels,
						&promClient.LabelPair{Name: ptr("a"), Value: ptr("b")},
						&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
						&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
					),
				})

				requireHistogramEqual(t, metrics["bar"], histogram{
					name: "bar", count: 1, sum: 50,
					buckets: []*promClient.Bucket{
						{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(40.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(50.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(60.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
					},
					labels: append(
						scenario.additionalLabels,
						&promClient.LabelPair{Name: ptr("c"), Value: ptr("d")},
						&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
						&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
					),
				})

				requireHistogramEqual(t, metrics["baz"], histogram{
					name: "baz", count: 1, sum: 80,
					buckets: []*promClient.Bucket{
						{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(70.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(80.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(90.0)},
						{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
					},
					labels: append(
						scenario.additionalLabels,
						&promClient.LabelPair{Name: ptr("e"), Value: ptr("f")},
						&promClient.LabelPair{Name: ptr("job"), Value: &svcName},
						&promClient.LabelPair{Name: ptr("service_name"), Value: &svcName},
					),
				})
			})
		})
	}
}

func TestCollectorGlobals(t *testing.T) {
	grpcPort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	collector, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "otel/opentelemetry-collector",
		Tag:        "0.115.0",
		PortBindings: map[docker.Port][]docker.PortBinding{
			"4317/tcp": {{HostPort: strconv.Itoa(grpcPort)}},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pool.Purge(collector); err != nil {
			t.Logf("Could not purge resource: %v", err)
		}
	})

	var (
		om       Manager
		ctx      = context.Background()
		endpoint = fmt.Sprintf("localhost:%d", grpcPort)
	)
	res, err := NewResource(t.Name(), "v1.2.3", attribute.String("instanceName", "my-instance-id"))
	require.NoError(t, err)
	tp, mp, err := om.Setup(ctx, res,
		WithInsecure(),
		WithTracerProvider(endpoint, WithTracingSamplingRate(1.0), WithGlobalTracerProvider()),
		WithMeterProvider(WithGRPCMeterProvider(endpoint), WithGlobalMeterProvider()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
	require.Equal(t, tp, otel.GetTracerProvider())
	require.Equal(t, mp, otel.GetMeterProvider())
}

func TestNonBlockingConnection(t *testing.T) {
	grpcPort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	res, err := NewResource(t.Name(), "v1.2.3",
		attribute.String("instanceName", "my-instance-id"),
	)
	require.NoError(t, err)

	var (
		om       Manager
		ctx      = context.Background()
		endpoint = fmt.Sprintf("localhost:%d", grpcPort)
	)
	_, mp, err := om.Setup(ctx, res,
		WithInsecure(),
		WithMeterProvider(
			WithGRPCMeterProvider(endpoint),
			WithMeterProviderExportsInterval(100*time.Millisecond),
		),
		WithRetryConfig(RetryConfig{
			Enabled:         true,
			InitialInterval: time.Second,
			MaxInterval:     time.Second,
			MaxElapsedTime:  time.Minute,
		}),
	)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, om.Shutdown(context.Background()))
	}()

	meter := mp.Meter("test")
	fooCounter, err := meter.Int64Counter("foo")
	require.NoError(t, err)
	barCounter, err := meter.Float64Counter("bar")
	require.NoError(t, err)

	// this counter will not be lost even though the container isn't even started. see MaxElapsedTime.
	fooCounter.Add(ctx, 123, metric.WithAttributes(attribute.String("hello", "world")))

	cwd, err := os.Getwd()
	require.NoError(t, err)

	container, _ := statsTest.StartOTelCollector(t, metricsPort,
		filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
		statsTest.WithStartCollectorPort(grpcPort),
	)
	barCounter.Add(ctx, 456) // this should be recorded

	metricsEndpoint := fmt.Sprintf("http://localhost:%d/metrics", dt.GetHostPort(t, metricsPort, container))
	// Note: Counter metrics have _total suffix due to OTel Collector Prometheus exporter
	metrics := requireMetrics(t, metricsEndpoint, "foo_total", "bar_total")

	defaultAttrs := append(globalGRPCDefaultAttrs,
		&promClient.LabelPair{Name: ptr("job"), Value: ptr("TestNonBlockingConnection")},
		&promClient.LabelPair{Name: ptr("service_name"), Value: ptr("TestNonBlockingConnection")},
	)

	require.EqualValues(t, ptr("foo_total"), metrics["foo_total"].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics["foo_total"].Type)
	require.Len(t, metrics["foo_total"].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(123.0)}, metrics["foo_total"].Metric[0].Counter)
	require.ElementsMatch(t, append(defaultAttrs,
		&promClient.LabelPair{Name: ptr("hello"), Value: ptr("world")},
	), metrics["foo_total"].Metric[0].Label)

	require.EqualValues(t, ptr("bar_total"), metrics["bar_total"].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics["bar_total"].Type)
	require.Len(t, metrics["bar_total"].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(456.0)}, metrics["bar_total"].Metric[0].Counter)
	require.ElementsMatch(t, defaultAttrs, metrics["bar_total"].Metric[0].Label)
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

func requireHistogramEqual(t *testing.T, mf *promClient.MetricFamily, h histogram) {
	t.Helper()
	require.EqualValues(t, &h.name, mf.Name)
	require.EqualValues(t, ptr(promClient.MetricType_HISTOGRAM), mf.Type)
	require.Len(t, mf.Metric, 1)
	require.EqualValuesf(t, &h.count, mf.Metric[0].Histogram.SampleCount,
		"Got %d, expected %d", *mf.Metric[0].Histogram.SampleCount, h.count,
	)
	require.EqualValuesf(t, &h.sum, mf.Metric[0].Histogram.SampleSum,
		"Got %.2f, expected %.2f", *mf.Metric[0].Histogram.SampleSum, h.sum,
	)
	require.ElementsMatchf(t, h.buckets, mf.Metric[0].Histogram.Bucket, "Buckets for %q do not match", h.name)
	require.ElementsMatch(t, h.labels, mf.Metric[0].Label)
}

func ptr[T any](v T) *T {
	return &v
}

type testCase struct {
	name               string
	additionalLabels   []*promClient.LabelPair
	counterSuffix      string // "_total" for gRPC/collector, "" for direct prometheus exporter
	setupMeterProvider func(testing.TB, ...MeterProviderOption) (*sdkmetric.MeterProvider, string)
}

type histogram struct {
	name    string
	count   uint64
	sum     float64
	buckets []*promClient.Bucket
	labels  []*promClient.LabelPair
}
