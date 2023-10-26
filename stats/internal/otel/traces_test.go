package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/rudderlabs/rudder-go-kit/stats/testhelper/spanmodel"
)

func TestTraces(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	exp, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
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

	var data spanmodel.Span
	require.NoError(t, json.Unmarshal(buf.Bytes(), &data))
	require.Equal(t, "my-span", data.Name)
	require.Equal(t, "my-tracer", data.InstrumentationLibrary.Name)
	require.Equal(t, []spanmodel.Attributes{
		{
			Key:   "foo",
			Value: spanmodel.Value{Type: "STRING", Value: "bar"},
		},
	}, data.Attributes)
	require.Equal(t, "Ok", data.Status.Code)
}
