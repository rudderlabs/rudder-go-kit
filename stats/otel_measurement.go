package stats

import (
	"context"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/metric"
)

// otelMeasurement is the statsd-specific implementation of Measurement
type otelMeasurement struct {
	genericMeasurement
	disabled bool
	// recordOption carries this measurement's attribute set, prebuilt once via metric.WithAttributeSet and
	// reused on every Count/Observe/SendTiming so the OTel SDK does not rebuild the attribute Set on each
	// call (which metric.WithAttributes would). nil for no-op (disabled) measurements, which never record.
	recordOption metric.MeasurementOption
	// percentile is set for the Float64Histogram-backed measurements that support it — histograms and timers.
	// It is dormant until the first Percentile call. nil for counters, gauges and no-op measurements.
	percentile *percentileSeries
}

// Percentile returns the p-th percentile over the last window for measurements that support it
// (histograms and timers on the OpenTelemetry backend), or (0, false) otherwise or when the window is
// empty. The first call enables in-process tracking for the series, so percentiles become available
// shortly after.
func (m *otelMeasurement) Percentile(p float64, window time.Duration) (float64, bool) {
	if m.percentile == nil {
		return 0, false
	}
	return m.percentile.compute(p, window)
}

// otelCounter represents a counter stat
type otelCounter struct {
	*otelMeasurement
	counter metric.Int64Counter
}

func (c *otelCounter) Count(n int) {
	if !c.disabled {
		c.counter.Add(context.TODO(), int64(n), c.recordOption)
	}
}

// Increment increases the stat by 1. Is the Equivalent of Count(1). Only applies to CountType stats
func (c *otelCounter) Increment() {
	if !c.disabled {
		c.counter.Add(context.TODO(), 1, c.recordOption)
	}
}

// otelGauge represents a gauge stat
type otelGauge struct {
	*otelMeasurement
	value atomic.Value
}

// Gauge records an absolute value for this stat. Only applies to GaugeType stats
func (g *otelGauge) Gauge(value any) {
	if g.disabled {
		return
	}
	g.value.Store(value)
}

func (g *otelGauge) getValue() any {
	if g.disabled {
		return nil
	}
	return g.value.Load()
}

// otelTimer represents a timer stat
type otelTimer struct {
	*otelMeasurement
	now   func() time.Time
	timer metric.Float64Histogram
}

// Since sends the time elapsed since duration start. Only applies to TimerType stats
func (t *otelTimer) Since(start time.Time) {
	if !t.disabled {
		t.SendTiming(time.Since(start))
	}
}

// SendTiming sends a timing for this stat. Only applies to TimerType stats
func (t *otelTimer) SendTiming(duration time.Duration) {
	if t.disabled {
		return
	}
	t.timer.Record(context.TODO(), duration.Seconds(), t.recordOption)
	if t.percentile != nil {
		// Percentile is computed over the recorded durations in seconds (no attributes: the percentile
		// provider is private to this series).
		t.percentile.record(context.TODO(), duration.Seconds())
	}
}

// RecordDuration records the duration of time between
// the call to this function and the execution of the function it returns.
// Only applies to TimerType stats
func (t *otelTimer) RecordDuration() func() {
	if t.disabled {
		return func() {}
	}
	var start time.Time
	if t.now == nil {
		start = time.Now()
	} else {
		start = t.now()
	}
	return func() {
		t.Since(start)
	}
}

// otelHistogram represents a histogram stat
type otelHistogram struct {
	*otelMeasurement
	histogram metric.Float64Histogram
}

// Observe sends an observation
func (h *otelHistogram) Observe(value float64) {
	if h.disabled {
		return
	}
	h.histogram.Record(context.TODO(), value, h.recordOption)
	if h.percentile != nil {
		// No attributes: the percentile provider is private to this series, so its single data point
		// already isolates these observations.
		h.percentile.record(context.TODO(), value)
	}
}
