package otel

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	promClient "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	statsTest "github.com/rudderlabs/rudder-go-kit/stats/testhelper"
	"github.com/rudderlabs/rudder-go-kit/testhelper"
	dt "github.com/rudderlabs/rudder-go-kit/testhelper/docker"
)

const (
	metricsPort = "8889"
)

// see https://opentelemetry.io/docs/collector/getting-started/
func TestCollector(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
		filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
	)

	ctx := context.Background()
	res, err := NewResource(t.Name(), "my-instance-id", "1.0.0")
	require.NoError(t, err)
	var om Manager
	tp, mp, err := om.Setup(ctx, res,
		WithInsecure(),
		WithTracerProvider(grpcEndpoint, 1.0),
		WithMeterProvider(grpcEndpoint,
			WithMeterProviderExportsInterval(100*time.Millisecond),
			WithHistogramBucketBoundaries("baz", "some-test", []float64{10, 20, 30}),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
	require.NotEqual(t, tp, otel.GetTracerProvider())
	require.NotEqual(t, mp, global.MeterProvider())

	m := mp.Meter("some-test")
	// foo counter
	counter, err := m.Int64Counter("foo")
	require.NoError(t, err)
	counter.Add(ctx, 1, attribute.String("hello", "world"))
	// bar counter
	counter, err = m.Int64Counter("bar")
	require.NoError(t, err)
	counter.Add(ctx, 5)
	// baz histogram
	h, err := m.Int64Histogram("baz")
	require.NoError(t, err)
	h.Record(ctx, 20, attribute.String("a", "b"))

	metrics := requireMetrics(t, container, 5*time.Second, "foo", "bar", "baz")

	require.EqualValues(t, ptr("foo"), metrics["foo"].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics["foo"].Type)
	require.Len(t, metrics["foo"].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(1.0)}, metrics["foo"].Metric[0].Counter)
	require.ElementsMatch(t, []*promClient.LabelPair{
		// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
		{Name: ptr("label1"), Value: ptr("value1")},
		{Name: ptr("hello"), Value: ptr("world")},
		{Name: ptr("job"), Value: ptr("TestCollector")},
		{Name: ptr("instance"), Value: ptr("my-instance-id")},
	}, metrics["foo"].Metric[0].Label)

	require.EqualValues(t, ptr("bar"), metrics["bar"].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics["bar"].Type)
	require.Len(t, metrics["bar"].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(5.0)}, metrics["bar"].Metric[0].Counter)
	require.ElementsMatch(t, []*promClient.LabelPair{
		// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
		{Name: ptr("label1"), Value: ptr("value1")},
		{Name: ptr("job"), Value: ptr("TestCollector")},
		{Name: ptr("instance"), Value: ptr("my-instance-id")},
	}, metrics["bar"].Metric[0].Label)

	requireHistogramEqual(t, metrics["baz"], histogram{
		name: "baz", count: 1, sum: 20,
		buckets: []*promClient.Bucket{
			{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
			{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(20.0)},
			{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
			{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
		},
		labels: []*promClient.LabelPair{
			{Name: ptr("label1"), Value: ptr("value1")},
			{Name: ptr("a"), Value: ptr("b")},
			{Name: ptr("job"), Value: ptr("TestCollector")},
			{Name: ptr("instance"), Value: ptr("my-instance-id")},
		},
	})
}

func TestHistogramBuckets(t *testing.T) {
	setup := func(t *testing.T, opts ...MeterProviderOption) (*docker.Container, *sdkmetric.MeterProvider) {
		cwd, err := os.Getwd()
		require.NoError(t, err)
		container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
			filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
		)

		ctx := context.Background()
		res, err := NewResource("TestHistogramBuckets", "my-instance-id", "1.0.0")
		require.NoError(t, err)
		var om Manager
		_, mp, err := om.Setup(ctx, res,
			WithInsecure(),
			WithMeterProvider(grpcEndpoint,
				append(opts, WithMeterProviderExportsInterval(50*time.Millisecond))...,
			),
		)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
		require.NotEqual(t, mp, global.MeterProvider())

		return container, mp
	}

	t.Run("default", func(t *testing.T) {
		ctx := context.Background()
		container, mp := setup(t,
			WithDefaultHistogramBucketBoundaries("meter-1", []float64{10, 20, 30}),
			WithDefaultHistogramBucketBoundaries("meter-2", []float64{40, 50, 60}),
		)

		// foo histogram
		h, err := mp.Meter("meter-1").Int64Histogram("foo")
		require.NoError(t, err)
		h.Record(ctx, 20, attribute.String("a", "b"))

		// bar histogram
		h, err = mp.Meter("meter-2").Int64Histogram("bar")
		require.NoError(t, err)
		h.Record(ctx, 50, attribute.String("c", "d"))

		metrics := requireMetrics(t, container, 5*time.Second, "foo", "bar")

		requireHistogramEqual(t, metrics["foo"], histogram{
			name: "foo", count: 1, sum: 20,
			buckets: []*promClient.Bucket{
				{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(10.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(20.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(30.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
			},
			labels: []*promClient.LabelPair{
				{Name: ptr("label1"), Value: ptr("value1")},
				{Name: ptr("a"), Value: ptr("b")},
				{Name: ptr("job"), Value: ptr("TestHistogramBuckets")},
				{Name: ptr("instance"), Value: ptr("my-instance-id")},
			},
		})

		requireHistogramEqual(t, metrics["bar"], histogram{
			name: "bar", count: 1, sum: 50,
			buckets: []*promClient.Bucket{
				{CumulativeCount: ptr(uint64(0)), UpperBound: ptr(40.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(50.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(60.0)},
				{CumulativeCount: ptr(uint64(1)), UpperBound: ptr(math.Inf(1))},
			},
			labels: []*promClient.LabelPair{
				{Name: ptr("label1"), Value: ptr("value1")},
				{Name: ptr("c"), Value: ptr("d")},
				{Name: ptr("job"), Value: ptr("TestHistogramBuckets")},
				{Name: ptr("instance"), Value: ptr("my-instance-id")},
			},
		})
	})
}

func TestCollectorGlobals(t *testing.T) {
	grpcPort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	collector, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "otel/opentelemetry-collector",
		Tag:        "0.67.0",
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
	res, err := NewResource(t.Name(), "my-instance-id", "1.0.0")
	require.NoError(t, err)
	tp, mp, err := om.Setup(ctx, res,
		WithInsecure(),
		WithTracerProvider(endpoint, 1.0, WithGlobalTracerProvider()),
		WithMeterProvider(endpoint, WithGlobalMeterProvider()),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })
	require.Equal(t, tp, otel.GetTracerProvider())
	require.Equal(t, mp, global.MeterProvider())
}

func TestNonBlockingConnection(t *testing.T) {
	grpcPort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	res, err := NewResource(t.Name(), "my-instance-id", "1.0.0")
	require.NoError(t, err)

	var (
		om       Manager
		ctx      = context.Background()
		endpoint = fmt.Sprintf("localhost:%d", grpcPort)
	)
	_, mp, err := om.Setup(ctx, res,
		WithInsecure(),
		WithMeterProvider(endpoint, WithMeterProviderExportsInterval(100*time.Millisecond)),
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
	fooCounter.Add(ctx, 123, attribute.String("hello", "world"))

	cwd, err := os.Getwd()
	require.NoError(t, err)

	container, _ := statsTest.StartOTelCollector(t, metricsPort,
		filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
		statsTest.WithStartCollectorPort(grpcPort),
	)
	barCounter.Add(ctx, 456) // this should be recorded

	metrics := requireMetrics(t, container, 5*time.Second, "foo", "bar")

	require.EqualValues(t, ptr("foo"), metrics["foo"].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics["foo"].Type)
	require.Len(t, metrics["foo"].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(123.0)}, metrics["foo"].Metric[0].Counter)
	require.ElementsMatch(t, []*promClient.LabelPair{
		// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
		{Name: ptr("label1"), Value: ptr("value1")},
		{Name: ptr("hello"), Value: ptr("world")},
		{Name: ptr("job"), Value: ptr("TestNonBlockingConnection")},
		{Name: ptr("instance"), Value: ptr("my-instance-id")},
	}, metrics["foo"].Metric[0].Label)

	require.EqualValues(t, ptr("bar"), metrics["bar"].Name)
	require.EqualValues(t, ptr(promClient.MetricType_COUNTER), metrics["bar"].Type)
	require.Len(t, metrics["bar"].Metric, 1)
	require.EqualValues(t, &promClient.Counter{Value: ptr(456.0)}, metrics["bar"].Metric[0].Counter)
	require.ElementsMatch(t, []*promClient.LabelPair{
		// the label1=value1 is coming from the otel-collector-config.yaml (see const_labels)
		{Name: ptr("label1"), Value: ptr("value1")},
		{Name: ptr("job"), Value: ptr("TestNonBlockingConnection")},
		{Name: ptr("instance"), Value: ptr("my-instance-id")},
	}, metrics["bar"].Metric[0].Label)
}

func requireMetrics(
	t *testing.T, container *docker.Container, timeout time.Duration, requiredKeys ...string,
) map[string]*promClient.MetricFamily {
	t.Helper()

	var (
		err             error
		resp            *http.Response
		metrics         map[string]*promClient.MetricFamily
		metricsEndpoint = fmt.Sprintf("http://localhost:%d/metrics", dt.GetHostPort(t, metricsPort, container))
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
	}, timeout, 100*time.Millisecond, "err: %v, metrics: %+v", err, metrics)

	return metrics
}

func requireHistogramEqual(t *testing.T, mf *promClient.MetricFamily, h histogram) {
	t.Helper()
	require.EqualValues(t, &h.name, mf.Name)
	require.EqualValues(t, ptr(promClient.MetricType_HISTOGRAM), mf.Type)
	require.Len(t, mf.Metric, 1)
	require.EqualValues(t, &h.count, mf.Metric[0].Histogram.SampleCount)
	require.EqualValues(t, &h.sum, mf.Metric[0].Histogram.SampleSum)
	require.ElementsMatchf(t, h.buckets, mf.Metric[0].Histogram.Bucket, "Buckets for %q do not match", h.name)
	require.ElementsMatch(t, h.labels, mf.Metric[0].Label)

}

func ptr[T any](v T) *T {
	return &v
}

type histogram struct {
	name    string
	count   uint64
	sum     float64
	buckets []*promClient.Bucket
	labels  []*promClient.LabelPair
}
