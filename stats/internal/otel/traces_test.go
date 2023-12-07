package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/rudderlabs/rudder-go-kit/stats/testhelper/tracemodel"
	"github.com/rudderlabs/rudder-go-kit/testhelper/assert"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

func TestTraces(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	exp, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(buf),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, exp.Shutdown(ctx))
	})

	res, err := NewResource(t.Name(), "v1.2.3",
		attribute.String("instanceName", "my-instance-id"),
	)
	require.NoError(t, err)

	var om Manager
	tp, _, err := om.Setup(ctx, res,
		WithCustomTracerProvider(exp, WithTracingSamplingRate(1.0), WithTracingSyncer()),
		WithTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })

	fooTracer := tp.Tracer("my-tracer")
	ctx, span := fooTracer.Start(ctx, "my-span", trace.WithAttributes(attribute.String("foo", "bar")))

	span.SetStatus(codes.Ok, "okey-dokey")
	time.Sleep(123 * time.Millisecond)
	span.End()

	var data tracemodel.Span
	require.NoError(t, json.Unmarshal(buf.Bytes(), &data))
	require.Equal(t, "my-span", data.Name)
	require.Equal(t, "my-tracer", data.InstrumentationLibrary.Name)
	require.Equal(t, []tracemodel.Attributes{
		{
			Key:   "foo",
			Value: tracemodel.Value{Type: "STRING", Value: "bar"},
		},
	}, data.Attributes)
	require.Equal(t, "Ok", data.Status.Code)
	require.InDelta(t, 123, data.EndTime.Sub(data.StartTime).Milliseconds(), 50)
}

func TestZipkinIntegration(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	zipkin, err := resource.SetupZipkin(pool, t)
	require.NoError(t, err)

	res, err := NewResource(t.Name(), "v1.2.3",
		attribute.String("instanceName", "my-instance-id"),
	)
	require.NoError(t, err)

	var (
		om        Manager
		ctx       = context.Background()
		zipkinURL = "http://localhost:" + zipkin.Port + "/api/v2/spans"
	)
	tp, _, err := om.Setup(ctx, res,
		WithTracerProvider(zipkinURL,
			WithTracingSamplingRate(1.0),
			WithTracingSyncer(),
			WithZipkin(),
		),
		WithTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, om.Shutdown(context.Background())) })

	_, span := tp.Tracer("my-tracer").Start(ctx, "my-span")
	time.Sleep(123 * time.Millisecond)
	span.End()

	zipkinSpansURL := zipkinURL + "?serviceName=" + t.Name()
	getSpansReq, err := http.NewRequest(http.MethodGet, zipkinSpansURL, nil)
	require.NoError(t, err)

	spansBody := assert.RequireEventuallyStatusCode(t, http.StatusOK, getSpansReq)
	require.Equal(t, `["my-span"]`, spansBody)
}
