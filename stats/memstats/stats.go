package memstats

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cast"

	"github.com/rudderlabs/rudder-go-kit/stats"
)

var _ stats.Stats = (*Store)(nil)

var _ stats.Measurement = (*Measurement)(nil)

type Store struct {
	mu    sync.Mutex
	byKey map[string]*Measurement
	now   func() time.Time
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
func (m *Measurement) Gauge(value interface{}) {
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

type Opts func(*Store)

func WithNow(nowFn func() time.Time) Opts {
	return func(s *Store) {
		s.now = nowFn
	}
}

func New(opts ...Opts) *Store {
	s := &Store{
		byKey: make(map[string]*Measurement),
		now:   time.Now,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
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

	m := &Measurement{
		name:  name,
		tags:  tags,
		mType: statType,

		now: ms.now,
	}

	ms.byKey[ms.getKey(name, tags)] = m
	return m
}

// Get the stored measurement with the name and tags.
// If no measurement is found, nil is returned.
func (ms *Store) Get(name string, tags stats.Tags) *Measurement {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	return ms.byKey[ms.getKey(name, tags)]
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

	metrics := make([]Metric, 0, len(ms.byKey))
	keys := make([]string, 0, len(ms.byKey))
	for k, m := range ms.byKey {
		if name != "" && m.name != name {
			continue
		}
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	for _, key := range keys {
		m := ms.byKey[key]
		switch m.mType {
		case stats.CountType:
			metrics = append(metrics, Metric{
				Name:  m.name,
				Tags:  m.tags,
				Value: m.LastValue(),
			})
		case stats.GaugeType:
			metrics = append(metrics, Metric{
				Name:  m.name,
				Tags:  m.tags,
				Value: m.LastValue(),
			})
		case stats.HistogramType:
			metrics = append(metrics, Metric{
				Name:   m.name,
				Tags:   m.tags,
				Values: m.Values(),
			})
		case stats.TimerType:
			metrics = append(metrics, Metric{
				Name:      m.name,
				Tags:      m.tags,
				Durations: m.Durations(),
			})
		default:
			panic("unknown measurement type:" + m.mType)
		}
	}

	return metrics
}

// Start implements stats.Stats
func (*Store) Start(_ context.Context, _ stats.GoRoutineFactory) error { return nil }

// Stop implements stats.Stats
func (*Store) Stop() {}

// getKey maps name and tags, to a store lookup key.
func (*Store) getKey(name string, tags stats.Tags) string {
	return name + tags.String()
}
