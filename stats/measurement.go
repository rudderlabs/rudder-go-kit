package stats

import (
	"fmt"
	"time"

	"github.com/rudderlabs/rudder-go-kit/stats/internal/percentile"
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
	// observations, and true when data is available. It is supported by histogram and timer measurements on
	// every backend (for timers the values are the recorded durations in seconds); counters, gauges and
	// no-op measurements return (0, false).
	//
	// Observations are kept in an in-memory ring per series, bounded by WithHistogramPercentileMaxSamples.
	// That cost is per distinct series (name + tags), so call Percentile only on low-cardinality, important
	// measurements — not on high-cardinality series.
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

// requireTrackableType panics unless statType is one for which Percentile tracking is meaningful — a timer
// or a histogram. It guards NewTrackedStat on every backend.
func requireTrackableType(statType string) {
	if statType != TimerType && statType != HistogramType {
		panic(fmt.Errorf(
			"NewTrackedStat only supports %q and %q measurement types, got %q", HistogramType, TimerType, statType,
		))
	}
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

// percentileSupport gives a Measurement a rolling-window Percentile backed by an in-memory ring of recent
// observations (see stats/internal/percentile). It is embedded by the histogram- and timer-capable
// measurements of every backend; counters, gauges and disabled measurements leave buffer nil, so their
// Percentile reports no data. Embedding it (rather than genericMeasurement) is what supplies Percentile, so
// genericMeasurement deliberately does not define one — having both would make the method ambiguous.
type percentileSupport struct {
	buffer *percentile.Buffer
}

// observe records value into the rolling window when the measurement supports percentiles (buffer != nil).
func (s percentileSupport) observe(value float64) {
	if s.buffer != nil {
		s.buffer.Observe(value)
	}
}

// Percentile returns the p-th percentile (p in [0,100]) over the observations made within the last window,
// and true when data is available; (0, false) otherwise or for measurements that do not support it. Unlike
// the mutating operations on genericMeasurement this is a read, so it never panics.
func (s percentileSupport) Percentile(p float64, window time.Duration) (float64, bool) {
	if s.buffer == nil {
		return 0, false
	}
	return s.buffer.Percentile(p, window)
}
