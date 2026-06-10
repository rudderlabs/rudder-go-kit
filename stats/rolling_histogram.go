package stats

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

const defaultRollingHistogramPercentile = 99

// RollingHistogramConfig configures an in-process rolling histogram tracker.
type RollingHistogramConfig struct {
	// Window is the amount of recent histogram delta data retained by the tracker.
	Window time.Duration
	// Percentile is the percentile returned by the tracker's Percentile method.
	// Values are expressed as 0-100. If unset, 99 is used.
	Percentile float64
}

// RollingHistogramTracker reads quantiles from a rolling in-process histogram.
type RollingHistogramTracker interface {
	// Percentile returns the configured percentile and true when the window has observations.
	Percentile() (float64, bool)
	// Quantile returns the requested quantile and true when the window has observations.
	Quantile(q float64) (float64, bool)
	// Count returns the number of observations in the current rolling window.
	Count() uint64
	// Reset clears all observations from the tracker.
	Reset()
}

// RollingHistogramStats is implemented by Stats backends that support in-process histogram quantiles.
type RollingHistogramStats interface {
	TrackHistogram(name string, tags Tags, cfg RollingHistogramConfig) (RollingHistogramTracker, error)
}

// TrackHistogram attaches an in-process rolling quantile tracker to an OTel+Prometheus histogram.
func TrackHistogram(
	s Stats, name string, tags Tags, cfg RollingHistogramConfig,
) (tracker RollingHistogramTracker, supported bool, err error) {
	rollingStats, ok := s.(RollingHistogramStats)
	if !ok {
		panic("rolling histogram percentiles require OpenTelemetry with Prometheus exporter enabled")
	}
	tracker, err = rollingStats.TrackHistogram(name, tags, cfg)
	return tracker, true, err
}

type rollingHistogramRegistry struct {
	mu           sync.RWMutex
	now          func() time.Time
	pollInterval time.Duration
	reader       sdkmetric.Reader
	entries      map[string]*rollingHistogramTracker
	started      bool
}

func newRollingHistogramRegistry(now func() time.Time, pollInterval time.Duration) *rollingHistogramRegistry {
	if now == nil {
		now = time.Now
	}
	if pollInterval <= 0 {
		pollInterval = 5 * time.Second
	}
	return &rollingHistogramRegistry{
		now:          now,
		pollInterval: pollInterval,
		entries:      make(map[string]*rollingHistogramTracker),
	}
}

func (r *rollingHistogramRegistry) start(ctx context.Context, goFactory GoRoutineFactory, reader sdkmetric.Reader) {
	if r == nil {
		return
	}
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return
	}
	r.reader = reader
	r.started = true
	r.mu.Unlock()

	goFactory.Go(func() {
		ticker := time.NewTicker(r.pollInterval)
		defer ticker.Stop()

		r.poll(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.poll(ctx)
			}
		}
	})
}

func (r *rollingHistogramRegistry) track(
	name string, tags Tags, cfg RollingHistogramConfig,
) (RollingHistogramTracker, error) {
	if r == nil {
		panic("rolling histogram percentiles require OpenTelemetry with Prometheus exporter enabled")
	}
	cfg, err := resolveRollingHistogramConfig(cfg)
	if err != nil {
		return nil, err
	}
	key := rollingHistogramKey(name, tags)

	r.mu.Lock()
	defer r.mu.Unlock()

	if entry, ok := r.entries[key]; ok {
		return entry, nil
	}
	tracker := &rollingHistogramTracker{
		window:    cfg.Window,
		quantile:  cfg.Percentile / 100,
		now:       r.now,
		prevByKey: make(map[string]exponentialHistogramSnapshot),
	}
	r.entries[key] = tracker
	return tracker, nil
}

func (r *rollingHistogramRegistry) poll(ctx context.Context) {
	r.mu.RLock()
	reader := r.reader
	hasEntries := len(r.entries) > 0
	r.mu.RUnlock()
	if reader == nil || !hasEntries {
		return
	}

	rm := metricdata.ResourceMetrics{}
	if err := reader.Collect(ctx, &rm); err != nil {
		return
	}
	now := r.now()

	for _, scopeMetrics := range rm.ScopeMetrics {
		for _, m := range scopeMetrics.Metrics {
			switch data := m.Data.(type) {
			case metricdata.ExponentialHistogram[float64]:
				pollExponentialHistogram(r, m.Name, data, now)
			case metricdata.ExponentialHistogram[int64]:
				pollExponentialHistogram(r, m.Name, data, now)
			}
		}
	}
}

func pollExponentialHistogram[N int64 | float64](
	r *rollingHistogramRegistry,
	name string, histogram metricdata.ExponentialHistogram[N], now time.Time,
) {
	for _, dp := range histogram.DataPoints {
		tags := tagsFromMetricAttributes(dp.Attributes)
		key := rollingHistogramKey(name, tags)

		r.mu.RLock()
		tracker := r.entries[key]
		r.mu.RUnlock()
		if tracker == nil {
			continue
		}
		tracker.observe(exponentialHistogramSnapshotFromDataPoint(dp), now)
	}
}

func rollingHistogramKey(name string, tags Tags) string {
	return name + "|" + tags.String()
}

func tagsFromMetricAttributes(attrs attribute.Set) Tags {
	if attrs.Len() == 0 {
		return nil
	}
	tags := make(Tags, attrs.Len())
	iter := attrs.Iter()
	for iter.Next() {
		kv := iter.Attribute()
		tags[string(kv.Key)] = kv.Value.AsString()
	}
	return tags
}

type rollingHistogramTracker struct {
	mu        sync.Mutex
	window    time.Duration
	quantile  float64
	now       func() time.Time
	prevByKey map[string]exponentialHistogramSnapshot
	samples   []timedExponentialHistogram
}

type timedExponentialHistogram struct {
	at       time.Time
	snapshot exponentialHistogramSnapshot
}

func (t *rollingHistogramTracker) Percentile() (float64, bool) {
	return t.Quantile(t.quantile)
}

func (t *rollingHistogramTracker) Quantile(q float64) (float64, bool) {
	if q < 0 || q > 1 || math.IsNaN(q) {
		return 0, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked()
	return quantileFromExponentialSnapshots(t.samples, q)
}

func (t *rollingHistogramTracker) Count() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked()
	var count uint64
	for _, sample := range t.samples {
		count += sample.snapshot.count
	}
	return count
}

func (t *rollingHistogramTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.prevByKey = make(map[string]exponentialHistogramSnapshot)
	t.samples = nil
}

func (t *rollingHistogramTracker) observe(current exponentialHistogramSnapshot, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	seriesKey := current.seriesKey()
	previous, ok := t.prevByKey[seriesKey]
	t.prevByKey[seriesKey] = current
	if !ok {
		return
	}
	delta := current.delta(previous)
	if delta.count == 0 {
		return
	}
	t.samples = append(t.samples, timedExponentialHistogram{
		at:       now,
		snapshot: delta,
	})
	t.pruneLocked()
}

func (t *rollingHistogramTracker) pruneLocked() {
	cutoff := t.now().Add(-t.window)
	firstLive := 0
	for firstLive < len(t.samples) && t.samples[firstLive].at.Before(cutoff) {
		firstLive++
	}
	if firstLive > 0 {
		copy(t.samples, t.samples[firstLive:])
		t.samples = t.samples[:len(t.samples)-firstLive]
	}
}

type exponentialHistogramSnapshot struct {
	scale         int32
	negativeCount uint64
	zeroCount     uint64
	count         uint64
	positiveCount map[int32]uint64
}

func exponentialHistogramSnapshotFromDataPoint[N int64 | float64](
	dp metricdata.ExponentialHistogramDataPoint[N],
) exponentialHistogramSnapshot {
	s := exponentialHistogramSnapshot{
		scale:         dp.Scale,
		zeroCount:     dp.ZeroCount,
		count:         dp.Count,
		positiveCount: make(map[int32]uint64, len(dp.PositiveBucket.Counts)),
	}
	for _, count := range dp.NegativeBucket.Counts {
		s.negativeCount += count
	}
	for i, count := range dp.PositiveBucket.Counts {
		if count == 0 {
			continue
		}
		s.positiveCount[dp.PositiveBucket.Offset+int32(i)+1] = count
	}
	return s
}

func (s exponentialHistogramSnapshot) seriesKey() string {
	return fmt.Sprintf("%d", s.scale)
}

func (s exponentialHistogramSnapshot) delta(previous exponentialHistogramSnapshot) exponentialHistogramSnapshot {
	if s.scale != previous.scale ||
		s.count < previous.count ||
		s.negativeCount < previous.negativeCount ||
		s.zeroCount < previous.zeroCount {
		return s
	}
	delta := exponentialHistogramSnapshot{
		scale:         s.scale,
		negativeCount: s.negativeCount - previous.negativeCount,
		zeroCount:     s.zeroCount - previous.zeroCount,
		count:         s.count - previous.count,
		positiveCount: make(map[int32]uint64, len(s.positiveCount)),
	}
	for index, count := range s.positiveCount {
		prev := previous.positiveCount[index]
		if count < prev {
			return s
		}
		if count > prev {
			delta.positiveCount[index] = count - prev
		}
	}
	return delta
}

func quantileFromExponentialSnapshots(samples []timedExponentialHistogram, q float64) (float64, bool) {
	var total, negativeCount, zeroCount uint64
	positiveBounds := make(map[float64]uint64)
	for _, sample := range samples {
		snapshot := sample.snapshot
		total += snapshot.count
		negativeCount += snapshot.negativeCount
		zeroCount += snapshot.zeroCount
		for index, count := range snapshot.positiveCount {
			positiveBounds[exponentialBucketUpperBound(snapshot.scale, index)] += count
		}
	}
	if total == 0 {
		return 0, false
	}

	rank := uint64(math.Ceil(q * float64(total)))
	if rank == 0 {
		rank = 1
	}
	if negativeCount+zeroCount >= rank {
		return 0, true
	}

	bounds := make([]float64, 0, len(positiveBounds))
	for bound := range positiveBounds {
		bounds = append(bounds, bound)
	}
	sort.Float64s(bounds)

	cumulative := negativeCount + zeroCount
	for _, bound := range bounds {
		cumulative += positiveBounds[bound]
		if cumulative >= rank {
			return bound, true
		}
	}
	if len(bounds) == 0 {
		return 0, false
	}
	return bounds[len(bounds)-1], true
}

func exponentialBucketUpperBound(scale, index int32) float64 {
	return math.Exp2(math.Ldexp(float64(index), int(-scale)))
}

func resolveRollingHistogramConfig(cfg RollingHistogramConfig) (RollingHistogramConfig, error) {
	if cfg.Window <= 0 {
		return cfg, errors.New("rolling histogram window must be positive")
	}
	if cfg.Percentile == 0 {
		cfg.Percentile = defaultRollingHistogramPercentile
	}
	if cfg.Percentile <= 0 || cfg.Percentile > 100 {
		return cfg, errors.New("rolling histogram percentile must be greater than 0 and less than or equal to 100")
	}
	return cfg, nil
}
