package main

import (
	"context"
	"os"
	"os/signal"
	"path"
	"time"

	"go.opentelemetry.io/otel/propagation"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	kitstats "github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
)

const (
	serviceName = "processor"
	zipkinURL   = "http://localhost:9411/api/v2/spans"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	conf := config.Default
	conf.Set("INSTANCE_ID", serviceName)
	conf.Set("OpenTelemetry.enabled", true)
	conf.Set("RuntimeStats.enabled", false)
	conf.Set("OpenTelemetry.traces.endpoint", zipkinURL)
	conf.Set("OpenTelemetry.traces.samplingRate", 1.0)
	conf.Set("OpenTelemetry.traces.withSyncer", true)
	conf.Set("OpenTelemetry.traces.withZipkin", true)
	conf.Set("OpenTelemetry.metrics.prometheus.enabled", true)
	conf.Set("OpenTelemetry.metrics.prometheus.port", 8889)
	l := logger.NewFactory(conf)
	m := metric.NewManager()
	log := l.NewLogger()

	stats := kitstats.NewStats(conf, l, m, kitstats.WithServiceName(serviceName))
	err := stats.Start(ctx, kitstats.DefaultGoRoutineFactory)
	if err != nil {
		log.Errorf("Error starting stats: %v", err)
		return
	}

	defer stats.Stop()

	tracer := stats.NewTracer("my-tracer")

	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Error getting CWD: %v", err)
		return
	}

	filename := path.Join(cwd, "../trace-context.txt")
	buf, err := os.ReadFile(filename)
	if err != nil {
		log.Errorf("Error reading %q: %v", filename, err)
		return
	}

	tc := propagation.TraceContext{}
	ctx = tc.Extract(ctx, propagation.MapCarrier{"traceparent": string(buf)})

	_, span := tracer.Start(ctx, "processor", kitstats.SpanKindInternal, time.Now(), kitstats.Tags{
		"component": "processor",
		"foo":       "qux",
	})
	time.Sleep(666 * time.Millisecond)
	span.End()

	if err != nil {
		log.Errorf("Error writing trace-context.txt: %v", err)
		return
	}
}
