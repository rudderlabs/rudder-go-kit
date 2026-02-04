package stats

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats/metric"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/jaeger"
)

func TestSpanFromContext(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	jaegerContainer, err := jaeger.Setup(pool, t)
	require.NoError(t, err)

	c := config.New()
	c.Set("INSTANCE_ID", t.Name())
	c.Set("OpenTelemetry.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	c.Set("OpenTelemetry.traces.endpoint", jaegerContainer.OTLPEndpoint)
	c.Set("OpenTelemetry.traces.samplingRate", 1.0)
	c.Set("OpenTelemetry.traces.withSyncer", true)
	c.Set("OpenTelemetry.traces.withOTLPHTTP", true)
	stats := NewStats(c, logger.NewFactory(c), metric.NewManager(),
		WithServiceName(t.Name()), WithServiceVersion("1.2.3"),
	)
	t.Cleanup(stats.Stop)

	require.NoError(t, stats.Start(context.Background(), DefaultGoRoutineFactory))

	tracer := stats.NewTracer("my-tracer")

	ctx, span := tracer.Start(context.Background(), "my-span-01", SpanKindInternal)
	spanFromCtx := tracer.SpanFromContext(ctx)
	require.Equalf(t, span, spanFromCtx, "SpanFromContext should return the same span as the one from Start()")

	// let's add the attributes to the span from the ctx, we should see them on Jaeger for the same span
	spanFromCtx.SetStatus(SpanStatusError, "some bad error")
	spanFromCtx.SetAttributes(Tags{"key1": "value1"})
	spanFromCtx.AddEvent("some-event",
		SpanWithTags(Tags{"key2": "value2"}),
		SpanWithTimestamp(time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC)),
	)
	span.End()

	var foundTrace *jaeger.Trace
	var foundSpan *jaeger.Span
	require.Eventually(t, func() bool {
		traces, err := jaegerContainer.GetTraces(t.Name())
		if err != nil || len(traces) == 0 {
			return false
		}
		for i := range traces {
			for j := range traces[i].Spans {
				if traces[i].Spans[j].OperationName == "my-span-01" {
					foundTrace = &traces[i]
					foundSpan = &traces[i].Spans[j]
					return true
				}
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond, "expected span 'my-span-01' not found in Jaeger")

	require.NotNil(t, foundSpan, "span 'my-span-01' not found")
	require.NotNil(t, foundTrace, "trace not found")

	// Build tag map for easier checking of span-level attributes
	tagMap := make(map[string]any)
	for _, tag := range foundSpan.Tags {
		tagMap[tag.Key] = tag.Value
	}

	// Verify span-level attributes
	require.Equal(t, "value1", tagMap["key1"])
	require.Equal(t, "ERROR", tagMap["otel.status_code"])
	require.Equal(t, "some bad error", tagMap["otel.status_description"])
	require.Equal(t, "my-tracer", tagMap["otel.scope.name"])
	require.Equal(t, "1.2.3", tagMap["otel.scope.version"])

	// Verify resource attributes from process info
	process, ok := foundTrace.Processes[foundSpan.ProcessID]
	require.True(t, ok, "process not found for span")
	require.Equal(t, t.Name(), process.ServiceName)

	// Build process tag map for resource attributes
	// Note: service.name is in process.ServiceName (already checked above), not in tags
	processTagMap := make(map[string]any)
	for _, tag := range process.Tags {
		processTagMap[tag.Key] = tag.Value
	}
	require.Equal(t, t.Name(), processTagMap["instanceName"])
	require.Equal(t, "1.2.3", processTagMap["service.version"])
	require.Equal(t, "go", processTagMap["telemetry.sdk.language"])
	require.Equal(t, "opentelemetry", processTagMap["telemetry.sdk.name"])
	require.Equal(t, otel.Version(), processTagMap["telemetry.sdk.version"])

	// Check for the event/log
	require.NotEmpty(t, foundSpan.Logs, "expected at least one log/event")
}

func TestAsyncTracePropagation(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	jaegerContainer, err := jaeger.Setup(pool, t)
	require.NoError(t, err)

	c := config.New()
	c.Set("INSTANCE_ID", t.Name())
	c.Set("OpenTelemetry.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	c.Set("OpenTelemetry.traces.endpoint", jaegerContainer.OTLPEndpoint)
	c.Set("OpenTelemetry.traces.samplingRate", 1.0)
	c.Set("OpenTelemetry.traces.withSyncer", true)
	c.Set("OpenTelemetry.traces.withOTLPHTTP", true)
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
	// this span should show as a child of my-span-01 on Jaeger
	_, span := tracer.Start(ctx, "my-span-02", SpanKindInternal)
	time.Sleep(234 * time.Millisecond)
	span.End()

	// let's verify that the two spans have the same traceID even though we did not share the context
	t.Logf("my-span-02 trace ID: %v", span.SpanContext().TraceID())
	require.Equalf(t, 0,
		strings.Index(traceParent, "00-"+span.SpanContext().TraceID().String()),
		"The 2nd span traceID should be the same as the 1st span traceID",
	)

	// let's check that the spans have the expected hierarchy on Jaeger as well
	require.Eventually(t, func() bool {
		traces, err := jaegerContainer.GetTraces(t.Name())
		if err != nil || len(traces) == 0 {
			return false
		}
		// We expect one trace with two spans
		for _, tr := range traces {
			if len(tr.Spans) >= 2 {
				return true
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond, "expected trace with 2 spans not found in Jaeger")

	traces, err := jaegerContainer.GetTraces(t.Name())
	require.NoError(t, err)
	require.NotEmpty(t, traces)

	// Find the trace with both spans
	var foundTrace *jaeger.Trace
	for i := range traces {
		if len(traces[i].Spans) >= 2 {
			foundTrace = &traces[i]
			break
		}
	}
	require.NotNil(t, foundTrace, "trace with 2 spans not found")

	// Find child span and verify it has a parent reference
	var childSpan *jaeger.Span
	for i := range foundTrace.Spans {
		if foundTrace.Spans[i].OperationName == "my-span-02" {
			childSpan = &foundTrace.Spans[i]
			break
		}
	}
	require.NotNil(t, childSpan, "child span 'my-span-02' not found")
	require.NotEmpty(t, childSpan.References, "child span should have a parent reference")
}

func TestTracingBackendDownIsNotBlocking(t *testing.T) {
	c := config.New()
	c.Set("INSTANCE_ID", t.Name())
	c.Set("OpenTelemetry.enabled", true)
	c.Set("RuntimeStats.enabled", false)
	// Use a non-existent endpoint to simulate backend being down
	c.Set("OpenTelemetry.traces.endpoint", "localhost:19999")
	c.Set("OpenTelemetry.traces.samplingRate", 1.0)
	c.Set("OpenTelemetry.traces.withSyncer", true)
	c.Set("OpenTelemetry.traces.withOTLPHTTP", true)
	stats := NewStats(c, logger.NewFactory(c), metric.NewManager(), WithServiceName(t.Name()))
	t.Cleanup(stats.Stop)
	require.NoError(t, stats.Start(context.Background(), DefaultGoRoutineFactory))

	tracer := stats.NewTracer("my-tracer")

	done := make(chan struct{})
	go func() {
		defer close(done)
		_, span := tracer.Start(context.Background(), "my-span-01", SpanKindInternal)
		span.End()
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("The Tracing API should not block if backend is down")
	}
}
