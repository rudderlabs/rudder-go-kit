package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	kitconfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conf := kitconfig.New()
	conf.Set("INSTANCE_ID", "histogram-demo")
	conf.Set("OpenTelemetry.enabled", true)
	conf.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	conf.Set("OpenTelemetry.metrics.prometheus.port", 8080)
	conf.Set("OpenTelemetry.metrics.exportInterval", 5*time.Second)
	conf.Set("RuntimeStats.enabled", false)

	logFactory := logger.NewFactory(conf)
	log := logFactory.NewLogger()
	m := metric.NewManager()
	r := prometheus.NewRegistry()

	s := stats.NewStats(conf, logFactory, m,
		stats.WithServiceName("histogram-demo"),
		stats.WithServiceVersion("v1.0.0"),
		stats.WithPrometheusRegistry(r, r),
		stats.WithDefaultExponentialHistogram(160),
	)

	if err := s.Start(ctx, stats.DefaultGoRoutineFactory); err != nil {
		log.Errorn("Failed to start stats", obskit.Error(err))
		os.Exit(1)
	}
	defer s.Stop()

	latencyHistogram := s.NewStat("request_latency_seconds", stats.HistogramType)

	log.Infon("Starting histogram data generator", logger.NewStringField("endpoint", "http://localhost:8080/metrics"))

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	requestCount := 0

	for {
		select {
		case <-ctx.Done():
			log.Warnn("Received interrupt signal, shutting down...")
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			latency := generateLatency(rng)
			latencyHistogram.Observe(latency)
			requestCount++

			if requestCount%100 == 0 {
				log.Infon("Generated 100 more requests",
					logger.NewIntField("requests", int64(requestCount)),
					logger.NewDurationField("latency", time.Duration(latency*1e6)*time.Microsecond),
				)
			}
		}
	}
}

func generateLatency(rng *rand.Rand) float64 {
	distribution := rng.Float64()

	switch {
	case distribution < 0.50:
		return rng.Float64()*0.050 + 0.001
	case distribution < 0.75:
		return rng.Float64()*0.450 + 0.050
	case distribution < 0.90:
		return rng.Float64()*1.500 + 0.500
	case distribution < 0.97:
		return rng.Float64()*2.000 + 2.000
	default:
		return rng.Float64()*1.000 + 4.000
	}
}
