package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/baggage"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	kitstats "github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
)

const (
	serviceName = "gateway"
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
	conf.Set("OpenTelemetry.metrics.prometheus.port", 8888)
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
	_, span := tracer.Start(ctx, "my-span", kitstats.SpanKindServer, time.Now(), kitstats.Tags{"foo": "bar"})
	time.Sleep(123 * time.Millisecond)
	span.End()

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Infof("Handling request: %v", req.URL.Path)

		ctx := req.Context()
		span := tracer.SpanFromContext(ctx)
		bag := baggage.FromContext(ctx)

		spanTags := kitstats.Tags{"username": bag.Member("username").Value()}
		span.AddEvent("handling this...", spanTags, time.Now(), false)
		span.End()

		_, _ = io.WriteString(w, "Hello, world!\n")
	}

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "Hello")

	httpSrv := http.Server{Addr: ":7777", Handler: otelHandler}
	go func() {
		<-ctx.Done()
		log.Infof("Context cancelled: %v", ctx.Err())
		_ = httpSrv.Shutdown(context.Background())
	}()

	if err = httpSrv.ListenAndServe(); err != nil {
		log.Errorf("Listen and serve: %v", err)
	}
}
