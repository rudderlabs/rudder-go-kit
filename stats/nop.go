package stats

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace/noop"
)

var NOP Stats = &nop{}

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

func (*nop) NewTracer(_ string) Tracer {
	return NewTracerFromOpenTelemetry(noop.NewTracerProvider().Tracer(""))
}

func (*nop) Start(_ context.Context, _ GoRoutineFactory) error { return nil }
func (*nop) Stop()                                             {}

func (*nop) RegisterCollector(c Collector) error { return nil }
