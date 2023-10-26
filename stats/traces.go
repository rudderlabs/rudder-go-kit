package stats

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceCode is an 32-bit representation of a status state.
type TraceCode uint32

const (
	// TraceErrorUnset is the default status code.
	TraceErrorUnset = TraceCode(codes.Unset)

	// TraceErrorError indicates the operation contains an error.
	TraceErrorError = TraceCode(codes.Error)

	// TraceErrorOk indicates operation has been validated by an Application developers
	// or Operator to have completed successfully, or contain no error.
	TraceErrorOk = TraceCode(codes.Ok)
)

type TraceKeyValue = attribute.KeyValue

type TraceSpanCtx = trace.SpanContext

type TraceSpanKind = trace.SpanKind

const (
	// SpanKindUnspecified is an unspecified SpanKind and is not a valid
	// SpanKind. SpanKindUnspecified should be replaced with SpanKindInternal
	// if it is received.
	SpanKindUnspecified = trace.SpanKindUnspecified
	// SpanKindInternal is a SpanKind for a Span that represents an internal
	// operation within an application.
	SpanKindInternal = trace.SpanKindInternal
	// SpanKindServer is a SpanKind for a Span that represents the operation
	// of handling a request from a client.
	SpanKindServer = trace.SpanKindServer
	// SpanKindClient is a SpanKind for a Span that represents the operation
	// of client making a request to a server.
	SpanKindClient = trace.SpanKindClient
	// SpanKindProducer is a SpanKind for a Span that represents the operation
	// of a producer sending a message to a message broker. Unlike
	// SpanKindClient and SpanKindServer, there is often no direct
	// relationship between this kind of Span and a SpanKindConsumer kind. A
	// SpanKindProducer Span will end once the message is accepted by the
	// message broker which might not overlap with the processing of that
	// message.
	SpanKindProducer = trace.SpanKindProducer
	// SpanKindConsumer is a SpanKind for a Span that represents the operation
	// of a consumer receiving a message from a message broker. Like
	// SpanKindProducer Spans, there is often no direct relationship between
	// this Span and the Span that produced the message.
	SpanKindConsumer = trace.SpanKindConsumer
)

type Tracer interface {
	Start( // @TODO add other options
		ctx context.Context, spanName string, spanKind TraceSpanKind,
		timestamp time.Time, attrs ...TraceKeyValue,
	) (context.Context, TraceSpan)
}

type TraceSpan interface {
	// AddEvent adds an event with the provided name and options.
	AddEvent(name string, attributes []TraceKeyValue, timestamp time.Time, stackTrace bool)

	// SetStatus sets the status of the Span in the form of a code and a
	// description, provided the status hasn't already been set to a higher
	// value before (OK > Error > Unset). The description is only included in a
	// status when the code is for an error.
	SetStatus(code TraceCode, description string)

	// SpanContext returns the SpanContext of the Span. The returned SpanContext
	// is usable even after the End method has been called for the Span.
	SpanContext() TraceSpanCtx

	// SetAttributes sets kv as attributes of the Span. If a key from kv
	// already exists for an attribute of the Span it will be overwritten with
	// the value contained in kv.
	SetAttributes(kv ...TraceKeyValue)

	// End terminates the span
	End() // @TODO add options (e.g. labels, timestamp, span kind, etc...)
}

// NewTracerFromOpenTelemetry creates a new go-kit Tracer from an OpenTelemetry Tracer.
func NewTracerFromOpenTelemetry(t trace.Tracer) Tracer {
	return &tracer{tracer: t}
}

type tracer struct {
	tracer trace.Tracer
}

func (t *tracer) Start(
	ctx context.Context, spanName string, spanKind TraceSpanKind,
	timestamp time.Time, attrs ...TraceKeyValue,
) (context.Context, TraceSpan) {
	var opts []trace.SpanStartOption
	if !timestamp.IsZero() {
		opts = append(opts, trace.WithTimestamp(timestamp))
	}
	if spanKind != SpanKindUnspecified {
		opts = append(opts, trace.WithSpanKind(spanKind))
	}
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}

	ctx, s := t.tracer.Start(ctx, spanName, opts...)
	return ctx, &span{span: s}
}

type span struct {
	span trace.Span
}

func (s *span) AddEvent(name string, attrs []TraceKeyValue, timestamp time.Time, stackTrace bool) {
	var opts []trace.EventOption
	if len(attrs) > 0 {
		opts = append(opts, trace.WithAttributes(attrs...))
	}
	if !timestamp.IsZero() {
		opts = append(opts, trace.WithTimestamp(timestamp))
	}
	if stackTrace {
		opts = append(opts, trace.WithStackTrace(true))
	}

	s.span.AddEvent(name, opts...)
}

func (s *span) SetStatus(code TraceCode, description string) {
	s.span.SetStatus(codes.Code(code), description)
}

func (s *span) SpanContext() TraceSpanCtx         { return s.span.SpanContext() }
func (s *span) SetAttributes(kv ...TraceKeyValue) { s.span.SetAttributes(kv...) }
func (s *span) End()                              { s.span.End() }
