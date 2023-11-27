package stats

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
	"github.com/rudderlabs/rudder-go-kit/stats/testhelper/tracemodel"
	"github.com/rudderlabs/rudder-go-kit/testhelper/assert"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

func TestSpanFromContext(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	zipkin, err := resource.SetupZipkin(pool, t)
	require.NoError(t, err)

	zipkinURL := "http://localhost:" + zipkin.Port + "/api/v2/spans"
	zipkinTracesURL := "http://localhost:" + zipkin.Port + "/api/v2/traces?serviceName=" + t.Name()

	c := config.New()
	c.Set("INSTANCE_ID", t.Name())
	c.Set("OpenTelemetry.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	c.Set("OpenTelemetry.traces.endpoint", zipkinURL)
	c.Set("OpenTelemetry.traces.samplingRate", 1.0)
	c.Set("OpenTelemetry.traces.withSyncer", true)
	c.Set("OpenTelemetry.traces.withZipkin", true)
	stats := NewStats(c, logger.NewFactory(c), metric.NewManager(),
		WithServiceName(t.Name()), WithServiceVersion("1.2.3"),
	)
	t.Cleanup(stats.Stop)

	require.NoError(t, stats.Start(context.Background(), DefaultGoRoutineFactory))

	tracer := stats.NewTracer("my-tracer")

	ctx, span := tracer.Start(context.Background(), "my-span-01", SpanKindInternal)
	spanFromCtx := tracer.SpanFromContext(ctx)
	require.Equalf(t, span, spanFromCtx, "SpanFromContext should return the span from the context")

	// let's add the attributes to the span from the ctx, we should see them on zipkin for the same span
	spanFromCtx.SetAttributes(Tags{"key1": "value1"})
	span.End()

	getTracesReq, err := http.NewRequest(http.MethodGet, zipkinTracesURL, nil)
	require.NoError(t, err)

	spansBody := assert.RequireEventuallyStatusCode(t, http.StatusOK, getTracesReq)

	var traces [][]tracemodel.ZipkinTrace
	require.NoError(t, json.Unmarshal([]byte(spansBody), &traces))

	require.Len(t, traces, 1)
	require.Len(t, traces[0], 1)
	require.Equal(t, traces[0][0].Name, "my-span-01")
	require.Equal(t, map[string]string{
		"key1":                   "value1", // if this is present then the attributes were added to the span
		"instanceName":           t.Name(),
		"service.name":           t.Name(),
		"service.version":        "1.2.3",
		"otel.library.name":      "my-tracer",
		"otel.library.version":   "1.2.3",
		"telemetry.sdk.language": "go",
		"telemetry.sdk.name":     "opentelemetry",
		"telemetry.sdk.version":  otel.Version(),
	}, traces[0][0].Tags)
}

func TestAsyncTracePropagation(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	zipkin, err := resource.SetupZipkin(pool, t)
	require.NoError(t, err)

	zipkinURL := "http://localhost:" + zipkin.Port + "/api/v2/spans"
	zipkinTracesURL := "http://localhost:" + zipkin.Port + "/api/v2/traces?serviceName=" + t.Name()

	c := config.New()
	c.Set("INSTANCE_ID", t.Name())
	c.Set("OpenTelemetry.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	c.Set("OpenTelemetry.traces.endpoint", zipkinURL)
	c.Set("OpenTelemetry.traces.samplingRate", 1.0)
	c.Set("OpenTelemetry.traces.withSyncer", true)
	c.Set("OpenTelemetry.traces.withZipkin", true)
	stats := NewStats(c, logger.NewFactory(c), metric.NewManager(), WithServiceName(t.Name()))
	t.Cleanup(stats.Stop)

	require.NoError(t, stats.Start(context.Background(), DefaultGoRoutineFactory))

	tracer := stats.NewTracer("my-tracer")

	// let's use an anonymous function to avoid sharing the context between the two spans
	traceParent := func() string {
		ctx, span := tracer.Start(context.Background(), "my-span-01", SpanKindInternal)
		time.Sleep(123 * time.Millisecond)
		span.End()

		return GetTraceParentFromContext(ctx)
	}()

	require.NotEmpty(t, traceParent, "traceParent should not be empty")
	t.Logf("traceParent my-span-01: %s", traceParent)

	// we are not sharing any context here, let's say we stored the traceParent in a database
	// now we want to continue the trace from the traceParent

	ctx := InjectTraceParentIntoContext(context.Background(), traceParent)
	// this span should show as a child of my-span-01 on zipkin
	_, span := tracer.Start(ctx, "my-span-02", SpanKindInternal)
	time.Sleep(234 * time.Millisecond)
	span.End()

	// let's verify that the two spans have the same traceID even though we did not share the context
	t.Logf("my-span-02 trace ID: %v", span.SpanContext().TraceID())
	require.Equalf(t, 0,
		strings.Index(traceParent, "00-"+span.SpanContext().TraceID().String()),
		"The 2nd span traceID should be the same as the 1st span traceID",
	)

	// let's check that the spans have the expected hierarchy on zipkin as well
	getTracesReq, err := http.NewRequest(http.MethodGet, zipkinTracesURL, nil)
	require.NoError(t, err)

	spansBody := assert.RequireEventuallyStatusCode(t, http.StatusOK, getTracesReq)

	var traces [][]tracemodel.ZipkinTrace
	require.NoError(t, json.Unmarshal([]byte(spansBody), &traces))

	require.Len(t, traces, 1)
	require.Len(t, traces[0], 2)
	require.NotEmpty(t, traces[0][1].ParentID)
	require.Equal(t, traces[0][0].ID, traces[0][1].ParentID)
}