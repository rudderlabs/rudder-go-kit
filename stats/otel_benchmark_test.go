package stats

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
	statsTest "github.com/rudderlabs/rudder-go-kit/stats/testhelper"
	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"
)

func BenchmarkOTel(b *testing.B) {
	cwd, err := os.Getwd()
	require.NoError(b, err)
	_, grpcEndpoint := statsTest.StartOTelCollector(b, metricsPort,
		filepath.Join(cwd, "testdata", "otel-collector-config.yaml"),
	)

	type testCase struct {
		name         string
		grpcEndpoint string
	}
	for _, tc := range []testCase{
		{"reachable", grpcEndpoint},
		{"unreachable", "unreachable:4317"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			b.Run("counter creation", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				b.ResetTimer()
				for i := 1; i <= b.N; i++ {
					s.NewTaggedStat("test_counter_"+randString+strconv.Itoa(i), CountType, Tags{"tag1": "value1"})
				}
			})

			b.Run("timer creation", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				b.ResetTimer()
				for i := 1; i <= b.N; i++ {
					s.NewTaggedStat("test_timer_"+randString+strconv.Itoa(i), TimerType, Tags{"tag1": "value1"})
				}
			})

			b.Run("gauge creation", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				b.ResetTimer()
				for i := 1; i <= b.N; i++ {
					s.NewTaggedStat("test_gauge_"+randString+strconv.Itoa(i), GaugeType, Tags{"tag1": "value1"})
				}
			})

			b.Run("histogram creation", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				b.ResetTimer()
				for i := 1; i <= b.N; i++ {
					s.NewTaggedStat("test_histogram_"+randString+strconv.Itoa(i), HistogramType, Tags{"tag1": "value1"})
				}
			})

			b.Run("use counter", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				m := s.NewTaggedStat("test_counter_"+randString, CountType, Tags{"tag1": "value1"})
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					m.Count(i)
				}
			})

			b.Run("use timer", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				m := s.NewTaggedStat("test_timer_"+randString, TimerType, Tags{"tag1": "value1"})
				start := time.Now()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					m.Since(start)
				}
			})

			b.Run("use gauge", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				m := s.NewTaggedStat("test_gauge_"+randString, GaugeType, Tags{"tag1": "value1"})
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					m.Gauge(i)
				}
			})

			b.Run("use histogram", func(b *testing.B) {
				randString := rand.UniqueString(10)
				s := getStatsForBenchmark(tc.grpcEndpoint)
				m := s.NewTaggedStat("test_histogram_"+randString, HistogramType, Tags{"tag1": "value1"})
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					m.Observe(float64(i))
				}
			})
		})
	}
}

func getStatsForBenchmark(grpcEndpoint string) Stats {
	c := config.New()
	c.Set("INSTANCE_ID", "my-instance-id")
	c.Set("OpenTelemetry.enabled", true)
	c.Set("OpenTelemetry.metrics.endpoint", grpcEndpoint)
	c.Set("OpenTelemetry.metrics.exportInterval", 10*time.Second)
	c.Set("RuntimeStats.enabled", false)
	l := logger.NewFactory(c)
	m := metric.NewManager()
	return NewStats(c, l, m)
}
