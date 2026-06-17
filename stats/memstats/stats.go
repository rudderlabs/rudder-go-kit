package memstats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/internal/percentile"
	"github.com/rudderlabs/rudder-go-kit/stats/testhelper/tracemodel"
)

var (
	_ stats.Stats       = (*Store)(nil)
	_ stats.Measurement = (*Measurement)(nil)
	_ stats.Measurement = (*trackedMeasurement)(nil)
)

type Store struct {
	mu    sync.Mutex
	byKey map[string]*Measurement
	now   func() time.Time

	withTracing       bool
	tracingBuffer     *bytes.Buffer
	tracingTimestamps bool
	tracerProvider    trace.TracerProvider
}

type Measurement struct {
	mu  sync.Mutex
	now func() time.Time

	tags  stats.Tags
	name  string
	mType string

	sum       float64
	values    []float64
	durations []time.Duration
	// percentileBuffer is the shared rolling-window ring for this series, created lazily by the first
	// NewTrackedStat call and reused by every tracked handle of the series so they share one ring. The
	// Measurement's own (untracked) Observe/SendTiming/Percentile never touch it — only the tracked handle
	// (trackedMeasurement) returned by NewTrackedStat feeds and reads it, mirroring the otel/statsd backends.
	percentileBuffer *percentile.Buffer
}

// Metric captures the name, tags and value(s) depending on type.
//
//	For Count and Gauge, Value is used.
//	For Histogram, Values is used.
//	For Timer, Durations is used.
type Metric struct {
	Name      string
	Tags      stats.Tags
	Value     float64         // Count, Gauge
	Values    []float64       // Histogram
	Durations []time.Duration // Timer
}

func (m *Measurement) LastValue() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.values) == 0 {
		return 0
	}

	return m.values[len(m.values)-1]
}

func (m *Measurement) Values() []float64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := make([]float64, len(m.values))
	copy(s, m.values)

	return s
}

func (m *Measurement) LastDuration() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.durations) == 0 {
		return 0
	}

	return m.durations[len(m.durations)-1]
}

func (m *Measurement) Durations() []time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := make([]time.Duration, len(m.durations))
	copy(s, m.durations)

	return s
}

// Count implements stats.Measurement
func (m *Measurement) Count(n int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mType != stats.CountType {
		panic("operation Count not supported for measurement type:" + m.mType)
	}

	m.sum += float64(n)
	m.values = append(m.values, m.sum)
}

// Increment implements stats.Measurement
func (m *Measurement) Increment() {
	if m.mType != stats.CountType {
		panic("operation Increment not supported for measurement type:" + m.mType)
	}

	m.Count(1)
}

// Gauge implements stats.Measurement
func (m *Measurement) Gauge(value any) {
	if m.mType != stats.GaugeType {
		panic("operation Gauge not supported for measurement type:" + m.mType)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.values = append(m.values, cast.ToFloat64(value))
}

// Observe implements stats.Measurement
func (m *Measurement) Observe(value float64) {
	if m.mType != stats.HistogramType {
		panic("operation Observe not supported for measurement type:" + m.mType)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.values = append(m.values, value)
}

// Percentile implements stats.Measurement. A measurement obtained from NewStat/NewTaggedStat is untracked and
// always reports no data here, exactly like the otel/statsd backends; the percentile is reported only by the
// tracked handle returned by NewTrackedStat (see trackedMeasurement.Percentile).
func (m *Measurement) Percentile(float64, time.Duration) (float64, bool) {
	return 0, false
}

// Since implements stats.Measurement
func (m *Measurement) Since(start time.Time) {
	if m.mType != stats.TimerType {
		panic("operation Since not supported for measurement type:" + m.mType)
	}

	m.SendTiming(m.now().Sub(start))
}

// SendTiming implements stats.Measurement
func (m *Measurement) SendTiming(duration time.Duration) {
	if m.mType != stats.TimerType {
		panic("operation SendTiming not supported for measurement type:" + m.mType)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.durations = append(m.durations, duration)
}

// RecordDuration implements stats.Measurement
func (m *Measurement) RecordDuration() func() {
	if m.mType != stats.TimerType {
		panic("operation RecordDuration not supported for measurement type:" + m.mType)
	}

	start := m.now()
	return func() {
		m.Since(start)
	}
}

// trackedMeasurement is the handle returned by Store.NewTrackedStat. It embeds the shared per-series
// Measurement — so it records values/durations and supports introspection (Get/Values/Durations) like any
// other handle — and adds the percentile ring: Observe/SendTiming also feed the ring and Percentile reads it.
// This mirrors the otel/statsd backends, where the tracked and untracked handles are distinct objects and
// only the tracked one carries the ring, so observations through a plain handle never affect Percentile.
type trackedMeasurement struct {
	*Measurement
	buffer *percentile.Buffer
}

// Observe records the value on the underlying histogram and into the percentile ring.
func (t *trackedMeasurement) Observe(value float64) {
	t.Measurement.Observe(value)
	t.buffer.Observe(value)
}

// SendTiming records the duration on the underlying timer and its seconds into the percentile ring.
func (t *trackedMeasurement) SendTiming(duration time.Duration) {
	t.Measurement.SendTiming(duration)
	// Percentile is computed over the recorded durations in seconds, like the other backends.
	t.buffer.Observe(duration.Seconds())
}

// Since records the time elapsed since start, routing through the tracked SendTiming so the ring is fed.
func (t *trackedMeasurement) Since(start time.Time) {
	t.SendTiming(t.now().Sub(start))
}

// RecordDuration records the elapsed time when the returned function is called, feeding the ring via Since.
func (t *trackedMeasurement) RecordDuration() func() {
	start := t.now()
	return func() {
		t.Since(start)
	}
}

// Percentile reports the rolling-window percentile over this tracked series' observations.
func (t *trackedMeasurement) Percentile(p float64, window time.Duration) (float64, bool) {
	return t.buffer.Percentile(p, window)
}

type Opts func(*Store)

func WithNow(nowFn func() time.Time) Opts {
	return func(s *Store) {
		s.now = nowFn
	}
}

func WithTracing() Opts {
	return func(s *Store) {
		s.withTracing = true
	}
}

func WithTracingTimestamps() Opts {
	return func(s *Store) {
		s.tracingTimestamps = true
	}
}

func New(opts ...Opts) (*Store, error) {
	s := &Store{
		byKey: make(map[string]*Measurement),
		now:   time.Now,
	}
	for _, opt := range opts {
		opt(s)
	}
	if !s.withTracing {
		s.tracerProvider = noop.NewTracerProvider()
		return s, nil
	}

	s.tracingBuffer = &bytes.Buffer{}

	tracingOpts := []stdouttrace.Option{
		stdouttrace.WithWriter(s.tracingBuffer),
		stdouttrace.WithPrettyPrint(),
	}
	if !s.tracingTimestamps {
		tracingOpts = append(tracingOpts, stdouttrace.WithoutTimestamps())
	}

	traceExporter, err := stdouttrace.New(tracingOpts...)
	if err != nil {
		return nil, fmt.Errorf("cannot create trace exporter: %w", err)
	}
	s.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(traceExporter),
	)

	return s, nil
}

func (ms *Store) NewTracer(name string) stats.Tracer {
	return stats.NewTracerFromOpenTelemetry(ms.tracerProvider.Tracer(name))
}

// NewStat implements stats.Stats
func (ms *Store) NewStat(name, statType string) (m stats.Measurement) {
	return ms.NewTaggedStat(name, statType, nil)
}

// NewTaggedStat implements stats.Stats
func (ms *Store) NewTaggedStat(name, statType string, tags stats.Tags) stats.Measurement {
	return ms.NewSampledTaggedStat(name, statType, tags)
}

// NewSampledTaggedStat implements stats.Stats
func (ms *Store) NewSampledTaggedStat(name, statType string, tags stats.Tags) stats.Measurement {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if m, found := ms.byKey[ms.getKey(name, tags)]; found {
		return m
	}

	m := &Measurement{
		name:  name,
		tags:  tags,
		mType: statType,

		now: ms.now,
	}

	ms.byKey[ms.getKey(name, tags)] = m
	return m
}

// NewTrackedStat implements stats.Stats. It returns a tracked handle for the series that, beyond recording
// like NewTaggedStat, retains recent observations (stamped with the store's clock) in a per-series ring so
// Percentile reports data. statType must be HistogramType or TimerType; any other type panics.
//
// As on the otel/statsd backends, only this tracked handle feeds and reads the ring: observations made through
// a plain NewStat/NewTaggedStat handle for the same series are still recorded as values/durations (and visible
// via Get) but are not reflected in Percentile. Repeated NewTrackedStat calls for the same series share one ring.
func (ms *Store) NewTrackedStat(name, statType string, tags stats.Tags) stats.Measurement {
	if statType != stats.HistogramType && statType != stats.TimerType {
		panic("NewTrackedStat only supports histogram and timer measurement types, got: " + statType)
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	m, found := ms.byKey[ms.getKey(name, tags)]
	if !found {
		m = &Measurement{name: name, tags: tags, mType: statType, now: ms.now}
		ms.byKey[ms.getKey(name, tags)] = m
	}

	m.mu.Lock()
	if m.percentileBuffer == nil {
		m.percentileBuffer = percentile.NewBuffer(0, ms.now)
	}
	buf := m.percentileBuffer
	m.mu.Unlock()

	return &trackedMeasurement{Measurement: m, buffer: buf}
}

// Get the stored measurement with the name and tags.
// If no measurement is found, nil is returned.
func (ms *Store) Get(name string, tags stats.Tags) *Measurement {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.byKey[ms.getKey(name, tags)]
}

func (ms *Store) Spans() ([]tracemodel.Span, error) {
	if ms.tracingBuffer.Len() == 0 {
		return nil, nil
	}

	// adding missing curly brackets and converting to a valid JSON array
	jsonData := "[" + ms.tracingBuffer.String() + "]"
	jsonData = strings.ReplaceAll(jsonData, "}\n{", "},{")

	var spans []tracemodel.Span
	err := json.Unmarshal([]byte(jsonData), &spans)

	return spans, err
}

// GetAll returns the metric for all name/tags register in the store.
func (ms *Store) GetAll() []Metric {
	return ms.getAllByName("")
}

// GetByName returns the metric for each tag variation with the given name.
func (ms *Store) GetByName(name string) []Metric {
	if name == "" {
		panic("name cannot be empty")
	}
	return ms.getAllByName(name)
}

func (ms *Store) getAllByName(name string) []Metric {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	keys := lo.Filter(lo.Keys(ms.byKey), func(k string, index int) bool {
		return name == "" || ms.byKey[k].name == name
	})
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return lo.Map(keys, func(key string, index int) Metric {
		m := ms.byKey[key]
		switch m.mType {
		case stats.CountType, stats.GaugeType:
			return Metric{
				Name:  m.name,
				Tags:  m.tags,
				Value: m.LastValue(),
			}
		case stats.HistogramType:
			return Metric{
				Name:   m.name,
				Tags:   m.tags,
				Values: m.Values(),
			}
		case stats.TimerType:
			return Metric{
				Name:      m.name,
				Tags:      m.tags,
				Durations: m.Durations(),
			}
		default:
			panic("unknown measurement type:" + m.mType)
		}
	})
}

// Start implements stats.Stats
func (*Store) Start(_ context.Context, _ stats.GoRoutineFactory) error { return nil }

// Stop implements stats.Stats
func (*Store) Stop() {}

func (ms *Store) RegisterCollector(c stats.Collector) error {
	c.Collect(func(key string, tags stats.Tags, val uint64) {
		ms.NewTaggedStat(key, stats.GaugeType, tags).Gauge(val)
	})
	return nil
}

// getKey maps name and tags, to a store lookup key.
func (*Store) getKey(name string, tags stats.Tags) string {
	return name + tags.String()
}
