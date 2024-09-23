package stats_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	promClient "github.com/prometheus/client_model/go"
	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/collectors"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
	statsTest "github.com/rudderlabs/rudder-go-kit/stats/testhelper"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
)

func ptr[T any](v T) *T {
	return &v
}

const (
	metricsPort = "8889"
)

var globalDefaultAttrs = []*promClient.LabelPair{
	{Name: ptr("instanceName"), Value: ptr("my-instance-id")},
	{Name: ptr("service_version"), Value: ptr("v1.2.3")},
	{Name: ptr("telemetry_sdk_language"), Value: ptr("go")},
	{Name: ptr("telemetry_sdk_name"), Value: ptr("opentelemetry")},
	{Name: ptr("telemetry_sdk_version"), Value: ptr(otel.Version())},
}

func TestOTelPeriodicStats(t *testing.T) {
	type expectation struct {
		name string
		tags []*promClient.LabelPair
	}

	cwd, err := os.Getwd()
	require.NoError(t, err)

	runTest := func(t *testing.T, expected []expectation, cols ...stats.Collector) {
		container, grpcEndpoint := statsTest.StartOTelCollector(t, metricsPort,
			filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
		)

		c := config.New()
		c.Set("INSTANCE_ID", "my-instance-id")
		c.Set("OpenTelemetry.enabled", true)
		c.Set("OpenTelemetry.metrics.endpoint", grpcEndpoint)
		c.Set("OpenTelemetry.metrics.exportInterval", time.Millisecond)
		m := metric.NewManager()

		l := logger.NewFactory(c)
		s := stats.NewStats(c, l, m,
			stats.WithServiceName("TestOTelPeriodicStats"),
			stats.WithServiceVersion("v1.2.3"),
		)

		for _, col := range cols {
			s.RegisterCollector(col)
		}

		// start stats
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		require.NoError(t, s.Start(ctx, stats.DefaultGoRoutineFactory))
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

	t.Run("static stats", func(t *testing.T) {
		runTest(t,
			[]expectation{
				{name: "a_custom_metric"},
			},
			collectors.NewStaticMetric("a_custom_metric", nil, 1),
		)

		runTest(t,
			[]expectation{
				{name: "a_custom_metric", tags: []*promClient.LabelPair{
					{Name: ptr("foo"), Value: ptr("bar")},
				}},
			},
			collectors.NewStaticMetric("a_custom_metric", stats.Tags{"foo": "bar"}, 1),
		)
	})

	t.Run("multiple collectors", func(t *testing.T) {
		runTest(t,
			[]expectation{
				{name: "col_1"},
				{name: "col_2"},
				{name: "col_3"},
			},
			collectors.NewStaticMetric("col_1", nil, 1),
			collectors.NewStaticMetric("col_2", nil, 1),
			collectors.NewStaticMetric("col_3", nil, 1),
		)
	})

	t.Run("sql collector", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)

		runTest(t,
			[]expectation{
				{name: "sql_db_max_open_connections", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_open_connections", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_in_use_connections", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_idle_connections", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_wait_count_total", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_wait_duration_seconds_total", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_max_idle_closed_total", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_max_idle_time_closed_total", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
				{name: "sql_db_max_lifetime_closed_total", tags: []*promClient.LabelPair{
					{Name: ptr("name"), Value: ptr("test")},
				}},
			},
			collectors.NewDatabaseSQLStats("test", db),
		)

	})

}
