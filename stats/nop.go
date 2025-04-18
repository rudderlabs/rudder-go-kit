package stats

import (
	"context"
	"time"
)

var (
	NOP       Stats  = &nop{}
	NOPTracer Tracer = &nopTracer{}
)

type nop struct{}

type nopMeasurement struct{}

func (nopMeasurement) Count(_ int)                {}
func (nopMeasurement) Increment()                 {}
func (nopMeasurement) Gauge(_ any)                {}
func (nopMeasurement) Observe(_ float64)          {}
func (nopMeasurement) SendTiming(_ time.Duration) {}
func (nopMeasurement) Since(_ time.Time)          {}
func (nopMeasurement) RecordDuration() func()     { return func() {} }

func (*nop) NewStat(_, _ string) Measurement {
	return &nopMeasurement{}
}

func (*nop) NewTaggedStat(_, _ string, _ Tags) Measurement {
	return &nopMeasurement{}
}

func (*nop) NewSampledTaggedStat(_, _ string, _ Tags) Measurement {
	return &nopMeasurement{}
}

func (*nop) NewTracer(_ string) Tracer { return NOPTracer }

func (*nop) Start(_ context.Context, _ GoRoutineFactory) error { return nil }
func (*nop) Stop()                                             {}

func (*nop) RegisterCollector(c Collector) error { return nil }

type nopTracer struct{}

func (*nopTracer) Start(ctx context.Context, _ string, _ SpanKind, _ ...SpanOption) (context.Context, TraceSpan) {
	return ctx, &nopSpan{}
}

func (*nopTracer) SpanFromContext(_ context.Context) TraceSpan {
	return &nopSpan{}
}

type nopSpan struct{}

func (*nopSpan) AddEvent(_ string, _ ...SpanOption) {}
func (*nopSpan) SetStatus(_ SpanStatus, _ string)   {}
func (*nopSpan) SpanContext() SpanContext           { return SpanContext{} }
func (*nopSpan) SetAttributes(_ Tags)               {}
func (*nopSpan) End()                               {}
