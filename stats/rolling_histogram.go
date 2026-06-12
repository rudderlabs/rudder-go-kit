package stats

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

const (
	defaultPollingInterval = 5 * time.Second
	// defaultMaxEmptyPolls is how many consecutive polls a tracker's window may stay empty
	// before the registry drops it to clean up memory; see rollingHistogramRegistry.poll.
	defaultMaxEmptyPolls = 1
	// Tracking instruments always use a fixed high-resolution exponential aggregation, regardless of how
	// the application configures its exported histograms. See newRollingHistogramRegistry.
	trackingHistogramMaxSize  = 160
	trackingHistogramMaxScale = 20
)

type rollingHistogramRegistry struct {
	mu            sync.RWMutex
	now           func() time.Time
	pollInterval  time.Duration
	maxEmptyPolls int
	logger        logger.Logger

	// A dedicated meter provider that holds ONLY the tracked histogram instruments: tracked stats
	// (created via NewTrackedHistogram) record into these and the poller reads them from this provider's
	// manual reader. This is what keeps polling O(tracked series) instead of O(all process metrics) —
	// Collect on this reader never sees counters, gauges, timers or untracked histograms.
	provider    *sdkmetric.MeterProvider
	reader      sdkmetric.Reader
	meter       metric.Meter
	instruments map[string]metric.Float64Histogram // tracking instrument per measurement name

	entries map[string]*rollingHistogramTracker
	started bool

	// collectFailing tracks whether the last poll's Collect failed, so we log the failure (and the
	// recovery) once per episode instead of every poll. Touched only by the poll goroutine.
	collectFailing bool
}

func newRollingHistogramRegistry(
	now func() time.Time, pollInterval time.Duration, maxEmptyPolls int,
) *rollingHistogramRegistry {
	if now == nil {
		now = time.Now
	}
	if pollInterval <= 0 {
		pollInterval = defaultPollingInterval
	}
	if maxEmptyPolls <= 0 {
		maxEmptyPolls = defaultMaxEmptyPolls
	}

	// A manual reader (cumulative temporality by default, which delta() relies on) on a private meter
	// provider. The view forces exponential aggregation so quantiles are accurate regardless of how the
	// exported histogram is bucketed.
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{Kind: sdkmetric.InstrumentKindHistogram},
			sdkmetric.Stream{Aggregation: sdkmetric.AggregationBase2ExponentialHistogram{
				MaxSize:  trackingHistogramMaxSize,
				MaxScale: trackingHistogramMaxScale,
			}},
		)),
	)

	return &rollingHistogramRegistry{
		now:           now,
		pollInterval:  pollInterval,
		maxEmptyPolls: maxEmptyPolls,
		provider:      provider,
		reader:        reader,
		meter:         provider.Meter("github.com/rudderlabs/rudder-go-kit/stats/rollinghistogram"),
		instruments:   make(map[string]metric.Float64Histogram),
		entries:       make(map[string]*rollingHistogramTracker),
	}
}

func (r *rollingHistogramRegistry) start(ctx context.Context, goFactory GoRoutineFactory, log logger.Logger) {
	if r == nil {
		return
	}
	r.mu.Lock()
	if r.started {
		r.mu.Unlock()
		return
	}
	r.logger = log
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

// track registers (or returns the existing) tracker for the given series and returns the dedicated
// instrument that the caller must record observations into. The instrument lives on the private meter
// provider, so the poller — which reads only that provider — never sees any other metric.
func (r *rollingHistogramRegistry) track(
	name string, tags Tags, window time.Duration,
) (*rollingHistogramTracker, metric.Float64Histogram, error) {
	if window <= 0 {
		return nil, nil, fmt.Errorf("rolling histogram window must be positive, got %s", window)
	}

	key := rollingHistogramKey(name, tags)

	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure the per-name tracking instrument exists on the dedicated meter (shared across tag sets).
	instrument, ok := r.instruments[name]
	if !ok {
		var err error
		instrument, err = r.meter.Float64Histogram(name)
		if err != nil {
			return nil, nil, fmt.Errorf("creating rolling histogram instrument %q: %w", name, err)
		}
		r.instruments[name] = instrument
	}

	if entry, ok := r.entries[key]; ok {
		return entry, instrument, nil
	}

	tracker := &rollingHistogramTracker{window: window, now: r.now}
	r.entries[key] = tracker

	return tracker, instrument, nil
}

func (r *rollingHistogramRegistry) poll(ctx context.Context) {
	r.mu.RLock()
	reader := r.reader
	hasEntries := len(r.entries) > 0
	r.mu.RUnlock()
	if reader == nil || !hasEntries {
		return
	}

	var resourceMetrics metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &resourceMetrics); err != nil {
		// Don't log while shutting down (the poller's context is cancelled in Stop): a Collect against
		// a reader being shut down is expected, not an operational failure.
		if ctx.Err() == nil && !r.collectFailing {
			r.collectFailing = true
			if r.logger != nil {
				r.logger.Warnn("Rolling histogram poll failed to collect metrics", obskit.Error(err))
			}
		}
		return
	}
	if r.collectFailing {
		r.collectFailing = false
		if r.logger != nil {
			r.logger.Infon("Rolling histogram poll resumed collecting metrics")
		}
	}

	now := r.now()
	for _, scopeMetrics := range resourceMetrics.ScopeMetrics {
		for _, m := range scopeMetrics.Metrics {
			switch data := m.Data.(type) {
			case metricdata.ExponentialHistogram[float64]:
				pollExponentialHistogram(r, m.Name, data, now)
			case metricdata.ExponentialHistogram[int64]:
				pollExponentialHistogram(r, m.Name, data, now)
			}
		}
	}

	r.evictEmpty()
}

// evictEmpty drops trackers whose rolling window has been empty for maxEmptyPolls consecutive polls
// (after having held data at least once), bounding registry memory when callers create trackers for
// series that later go idle. A tracker that is still warming up (never held data) is never evicted.
// Note: an evicted series that later resumes will not be re-attached to a tracker the caller still
// holds — that tracker stays empty until the caller registers it again.
func (r *rollingHistogramRegistry) evictEmpty() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for key, tracker := range r.entries {
		if tracker.evictable(r.maxEmptyPolls) {
			delete(r.entries, key)
		}
	}
}

func pollExponentialHistogram[N int64 | float64](
	r *rollingHistogramRegistry, name string, histogram metricdata.ExponentialHistogram[N], now time.Time,
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
	mu     sync.Mutex
	window time.Duration
	now    func() time.Time
	// prev is the last cumulative snapshot observed for this (single) series; deltas are computed
	// against it. A tracker only ever sees one series, so a single snapshot is enough.
	prev    exponentialHistogramSnapshot
	hasPrev bool
	samples []timedExponentialHistogram

	// eviction bookkeeping, maintained by the registry poll goroutine under mu.
	hadSamples bool // the window has held at least one sample (i.e. the tracker is past warm-up)
	emptyPolls int  // consecutive polls the window has been empty since it last held data
}

type timedExponentialHistogram struct {
	at       time.Time
	snapshot exponentialHistogramSnapshot
}

// percentile returns the p-th percentile (p in [0,100]) over the rolling window and true when the
// window holds observations; (0, false) otherwise.
func (t *rollingHistogramTracker) percentile(p float64) (float64, bool) {
	if p < 0 || p > 100 || math.IsNaN(p) {
		return 0, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked()
	return quantileFromExponentialSnapshots(t.samples, p/100)
}

// count returns the number of observations currently in the rolling window. Used by tests.
func (t *rollingHistogramTracker) count() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked()
	var count uint64
	for _, sample := range t.samples {
		count += sample.snapshot.count
	}
	return count
}

func (t *rollingHistogramTracker) observe(current exponentialHistogramSnapshot, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	previous := t.prev
	hadPrev := t.hasPrev
	t.prev = current
	t.hasPrev = true
	if !hadPrev {
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
	t.hadSamples = true
	t.pruneLocked()
}

// evictable prunes expired samples and reports whether the tracker should be dropped: it has held
// data at some point and its window has now been empty for at least maxEmptyPolls consecutive polls.
// A tracker that has never held data (still warming up) is kept regardless.
func (t *rollingHistogramTracker) evictable(maxEmptyPolls int) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.pruneLocked()
	if len(t.samples) > 0 {
		t.emptyPolls = 0
		return false
	}
	if !t.hadSamples {
		return false
	}
	t.emptyPolls++
	return t.emptyPolls >= maxEmptyPolls
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

// downscale reduces the snapshot's resolution to targetScale (which must be <= s.scale), merging
// positive buckets the way OTel does when it downscales: bucket index i maps to i>>delta. Counts that
// are not scale-dependent (negative, zero, total) are preserved. It is used to align two cumulative
// snapshots taken at different scales before differencing, since OTel histograms downscale over time.
func (s exponentialHistogramSnapshot) downscale(targetScale int32) exponentialHistogramSnapshot {
	if targetScale >= s.scale {
		return s
	}
	shift := uint(s.scale - targetScale)
	out := exponentialHistogramSnapshot{
		scale:         targetScale,
		negativeCount: s.negativeCount,
		zeroCount:     s.zeroCount,
		count:         s.count,
		positiveCount: make(map[int32]uint64, len(s.positiveCount)),
	}
	for key, count := range s.positiveCount {
		// key is the upper-bound exponent (OTel bucket index + 1); shift the index, not the exponent.
		out.positiveCount[((key-1)>>shift)+1] += count
	}
	return out
}

func (s exponentialHistogramSnapshot) delta(previous exponentialHistogramSnapshot) exponentialHistogramSnapshot {
	// Align scales before differencing: OTel exponential histograms only ever downscale, so bring the
	// higher-resolution snapshot down to the lower scale (previously a scale change was treated as a
	// reset, which dropped a poll interval of data on every rescale).
	cur, prev := s, previous
	if cur.scale > prev.scale {
		cur = cur.downscale(prev.scale)
	} else if prev.scale > cur.scale {
		prev = prev.downscale(cur.scale)
	}

	// A drop in any cumulative total means the series reset (e.g. process restart); the current
	// cumulative snapshot is then the best estimate of the delta.
	if cur.count < prev.count ||
		cur.negativeCount < prev.negativeCount ||
		cur.zeroCount < prev.zeroCount {
		return cur
	}
	delta := exponentialHistogramSnapshot{
		scale:         cur.scale,
		negativeCount: cur.negativeCount - prev.negativeCount,
		zeroCount:     cur.zeroCount - prev.zeroCount,
		count:         cur.count - prev.count,
		positiveCount: make(map[int32]uint64, len(cur.positiveCount)),
	}
	for index, count := range cur.positiveCount {
		prevCount := prev.positiveCount[index]
		if count < prevCount {
			return cur
		}
		if count > prevCount {
			delta.positiveCount[index] = count - prevCount
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
