package stats

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// SpanStatus is an 32-bit representation of a status state.
type SpanStatus uint32

const (
	// SpanStatusUnset is the default status code.
	SpanStatusUnset = SpanStatus(codes.Unset)

	// SpanStatusError indicates the operation contains an error.
	SpanStatusError = SpanStatus(codes.Error)

	// SpanStatusOk indicates operation has been validated by an Application developers
	// or Operator to have completed successfully, or contain no error.
	SpanStatusOk = SpanStatus(codes.Ok)
)

type (
	SpanKind    = trace.SpanKind
	SpanContext = trace.SpanContext
)

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
	SpanFromContext(context.Context) TraceSpan
	Start(
		ctx context.Context, spanName string, spanKind SpanKind, options ...SpanOption,
	) (context.Context, TraceSpan)
}

type TraceSpan interface {
	// AddEvent adds an event with the provided name and options.
	AddEvent(name string, options ...SpanOption)

	// SetStatus sets the status of the Span in the form of a code and a
	// description, provided the status hasn't already been set to a higher
	// value before (OK > Error > Unset). The description is only included in a
	// status when the code is for an error.
	SetStatus(code SpanStatus, description string)

	// SpanContext returns the SpanContext of the Span. The returned SpanContext
	// is usable even after the End method has been called for the Span.
	SpanContext() SpanContext

	// SetAttributes sets kv as attributes of the Span. If a key from kv
	// already exists for an attribute of the Span it will be overwritten with
	// the value contained in kv.
	SetAttributes(tags Tags)

	// End terminates the span
	End()
}

// NewTracerFromOpenTelemetry creates a new go-kit Tracer from an OpenTelemetry Tracer.
func NewTracerFromOpenTelemetry(t trace.Tracer) Tracer {
	return &tracer{tracer: t}
}

type tracer struct {
	tracer trace.Tracer
}

func (t *tracer) SpanFromContext(ctx context.Context) TraceSpan {
	return &span{span: trace.SpanFromContext(ctx)}
}

func (t *tracer) Start(ctx context.Context, spanName string, spanKind SpanKind, opts ...SpanOption) (
	context.Context, TraceSpan,
) {
	var sc spanConfig
	for _, opt := range opts {
		opt(&sc)
	}

	var startOpts []trace.SpanStartOption
	if len(sc.tags) > 0 {
		startOpts = append(startOpts, trace.WithAttributes(sc.tags.otelAttributes()...))
	}
	if !sc.timestamp.IsZero() {
		startOpts = append(startOpts, trace.WithTimestamp(sc.timestamp))
	}
	if spanKind != SpanKindUnspecified {
		startOpts = append(startOpts, trace.WithSpanKind(spanKind))
	}

	ctx, s := t.tracer.Start(ctx, spanName, startOpts...)
	return ctx, &span{span: s}
}

type spanConfig struct {
	tags      Tags
	timestamp time.Time
}

// SpanOption can be used with span.Start() and span.AddEvent()
type SpanOption func(*spanConfig)

// SpanWithTags sets the tags for the span
func SpanWithTags(tags Tags) SpanOption {
	return func(c *spanConfig) {
		c.tags = tags
	}
}

// SpanWithTimestamp sets the timestamp for the span
func SpanWithTimestamp(timestamp time.Time) SpanOption {
	return func(c *spanConfig) {
		c.timestamp = timestamp
	}
}

type span struct {
	span trace.Span
}

func (s *span) AddEvent(name string, opts ...SpanOption) {
	var sc spanConfig
	for _, opt := range opts {
		opt(&sc)
	}

	var eventOpts []trace.EventOption
	if len(sc.tags) > 0 {
		eventOpts = append(eventOpts, trace.WithAttributes(sc.tags.otelAttributes()...))
	}
	if !sc.timestamp.IsZero() {
		eventOpts = append(eventOpts, trace.WithTimestamp(sc.timestamp))
	}

	s.span.AddEvent(name, eventOpts...)
}

func (s *span) SetStatus(code SpanStatus, description string) {
	s.span.SetStatus(codes.Code(code), description)
}

func (s *span) SpanContext() SpanContext { return s.span.SpanContext() }
func (s *span) SetAttributes(t Tags)     { s.span.SetAttributes(t.otelAttributes()...) }
func (s *span) End()                     { s.span.End() }

const (
	traceParentHeader = "traceparent"
)

// GetTraceParentFromContext returns the traceparent header from the context
func GetTraceParentFromContext(ctx context.Context) string {
	mapCarrier := propagation.MapCarrier{}
	(propagation.TraceContext{}).Inject(ctx, mapCarrier)
	return mapCarrier[traceParentHeader]
}

// InjectTraceParentIntoContext injects the traceparent header into the context
func InjectTraceParentIntoContext(ctx context.Context, traceParent string) context.Context {
	return (propagation.TraceContext{}).Extract(ctx, propagation.MapCarrier{
		traceParentHeader: traceParent,
	})
}
