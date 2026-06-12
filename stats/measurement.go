package stats

import (
	"fmt"
	"time"
)

// Counter represents a counter metric
type Counter interface {
	Count(n int)
	Increment()
}

// Gauge represents a gauge metric
type Gauge interface {
	Gauge(value any)
}

// Histogram represents a histogram metric
type Histogram interface {
	Observe(value float64)

	// Percentile returns the p-th percentile (p in [0,100]) over the rolling window of the most recent
	// observations, and true when data is available. It is supported only by the OpenTelemetry backend
	// (which retains recent observations as exemplars); every other backend returns (0, false).
	Percentile(p float64, window time.Duration) (float64, bool)
}

// Timer represents a timer metric
type Timer interface {
	SendTiming(duration time.Duration)
	Since(start time.Time)
	RecordDuration() func()
}

// Measurement provides all stat measurement functions
// TODO: the API should not return a union of measurement methods, but rather a distinct type for each measurement type
type Measurement interface {
	Counter
	Gauge
	Histogram
	Timer
}

type genericMeasurement struct {
	statType string
}

// Count default behavior is to panic as not supported operation
func (m *genericMeasurement) Count(_ int) {
	panic(fmt.Errorf("operation Count not supported for measurement type:%s", m.statType))
}

// Increment default behavior is to panic as not supported operation
func (m *genericMeasurement) Increment() {
	panic(fmt.Errorf("operation Increment not supported for measurement type:%s", m.statType))
}

// Gauge default behavior is to panic as not supported operation
func (m *genericMeasurement) Gauge(_ any) {
	panic(fmt.Errorf("operation Gauge not supported for measurement type:%s", m.statType))
}

// Observe default behavior is to panic as not supported operation
func (m *genericMeasurement) Observe(_ float64) {
	panic(fmt.Errorf("operation Observe not supported for measurement type:%s", m.statType))
}

// Start default behavior is to panic as not supported operation
func (m *genericMeasurement) Start() {
	panic(fmt.Errorf("operation Start not supported for measurement type:%s", m.statType))
}

func (m *genericMeasurement) End() {
	panic(fmt.Errorf("operation End not supported for measurement type:%s", m.statType))
}

// SendTiming default behavior is to panic as not supported operation
func (m *genericMeasurement) SendTiming(_ time.Duration) {
	panic(fmt.Errorf("operation SendTiming not supported for measurement type:%s", m.statType))
}

// Since default behavior is to panic as not supported operation
func (m *genericMeasurement) Since(_ time.Time) {
	panic(fmt.Errorf("operation Since not supported for measurement type:%s", m.statType))
}

// RecordDuration default behavior is to panic as not supported operation
func (m *genericMeasurement) RecordDuration() func() {
	panic(fmt.Errorf("operation RecordDuration not supported for measurement type:%s", m.statType))
}

// Percentile default behavior is to report no data: unlike the mutating operations above this is a
// read, so it returns (0, false) rather than panicking. Only the OpenTelemetry histogram overrides it.
func (m *genericMeasurement) Percentile(_ float64, _ time.Duration) (float64, bool) {
	return 0, false
}
