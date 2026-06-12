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
)

const (
	// defaultPollInterval is the minimum time between two on-read snapshots of a tracked histogram.
	defaultPollInterval = 5 * time.Second
	// Tracking instruments always use a fixed high-resolution exponential aggregation, regardless of how
	// the application configures its exported histograms. See newRollingHistogramRegistry.
	trackingHistogramMaxSize  = 160
	trackingHistogramMaxScale = 20
)

// rollingHistogramRegistry owns a private meter provider that holds ONLY the tracked histogram
// instruments. Tracked stats (created via NewTrackedHistogram) record into these instruments, and each
// tracker reads them back on demand from the provider's manual reader — there is no background poller,
// so a service that never calls NewTrackedHistogram does no extra work at all.
type rollingHistogramRegistry struct {
	mu           sync.Mutex
	now          func() time.Time
	pollInterval time.Duration

	provider    *sdkmetric.MeterProvider
	reader      sdkmetric.Reader
	meter       metric.Meter
	instruments map[string]metric.Float64Histogram // tracking instrument per measurement name
}

func newRollingHistogramRegistry(now func() time.Time, pollInterval time.Duration) *rollingHistogramRegistry {
	if now == nil {
		now = time.Now
	}
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}

	// A manual reader (cumulative temporality by default, which delta() relies on) on a private meter
	// provider. A manual reader runs no background goroutine — it is only collected on demand. The view
	// forces exponential aggregation so quantiles are accurate regardless of how the exported histogram
	// is bucketed.
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
		now:          now,
		pollInterval: pollInterval,
		provider:     provider,
		reader:       reader,
		meter:        provider.Meter("github.com/rudderlabs/rudder-go-kit/stats/rollinghistogram"),
		instruments:  make(map[string]metric.Float64Histogram),
	}
}

// track returns a tracker for the given series plus the dedicated instrument the caller must record
// observations into. The instrument lives on the private meter provider (and is shared across tag sets
// of the same name), so the tracker — which reads only that provider — never sees any other metric.
func (r *rollingHistogramRegistry) track(
	name string, tags Tags, window time.Duration,
) (*rollingHistogramTracker, metric.Float64Histogram, error) {
	if window <= 0 {
		return nil, nil, fmt.Errorf("rolling histogram window must be positive, got %s", window)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	instrument, ok := r.instruments[name]
	if !ok {
		var err error
		instrument, err = r.meter.Float64Histogram(name)
		if err != nil {
			return nil, nil, fmt.Errorf("creating rolling histogram instrument %q: %w", name, err)
		}
		r.instruments[name] = instrument
	}

	tracker := &rollingHistogramTracker{
		reader:       r.reader,
		name:         name,
		key:          rollingHistogramKey(name, tags),
		window:       window,
		pollInterval: r.pollInterval,
		now:          r.now,
	}
	return tracker, instrument, nil
}

// rollingHistogramTracker maintains a rolling window of recent observations for a single series by
// diffing successive cumulative snapshots of its dedicated instrument. Snapshots are taken lazily, when
// Percentile is read, at most once per pollInterval — so there is no background goroutine and frequent
// reads don't grow the window unboundedly.
type rollingHistogramTracker struct {
	mu sync.Mutex

	reader       sdkmetric.Reader // private reader to collect on read; nil only in unit tests
	name         string           // instrument name to find in the collected metrics
	key          string           // name|tags identity of the exact series this tracker follows
	window       time.Duration
	pollInterval time.Duration
	now          func() time.Time

	// prev is the last cumulative snapshot seen for this series; deltas are computed against it.
	prev         exponentialHistogramSnapshot
	hasPrev      bool
	lastSnapshot time.Time
	samples      []timedExponentialHistogram
}

type timedExponentialHistogram struct {
	at       time.Time
	snapshot exponentialHistogramSnapshot
}

// percentile takes a fresh snapshot if one is due, then returns the p-th percentile (p in [0,100]) over
// the rolling window and true when the window holds observations; (0, false) otherwise.
func (t *rollingHistogramTracker) percentile(p float64) (float64, bool) {
	if p < 0 || p > 100 || math.IsNaN(p) {
		return 0, false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if t.reader != nil && (t.lastSnapshot.IsZero() || now.Sub(t.lastSnapshot) >= t.pollInterval) {
		t.snapshotLocked(now)
	}
	t.pruneLocked(now)
	return quantileFromExponentialSnapshots(t.samples, p/100)
}

// snapshotLocked collects the current cumulative value of this tracker's series and folds the delta
// since the previous snapshot into the rolling window. A collect error or a missing series is treated
// as "no new data": the existing window is left untouched.
func (t *rollingHistogramTracker) snapshotLocked(now time.Time) {
	t.lastSnapshot = now

	var rm metricdata.ResourceMetrics
	if err := t.reader.Collect(context.Background(), &rm); err != nil {
		return
	}
	snapshot, ok := t.findSnapshot(&rm)
	if !ok {
		return
	}
	t.observeLocked(snapshot, now)
}

// findSnapshot locates this tracker's series among the collected metrics.
func (t *rollingHistogramTracker) findSnapshot(rm *metricdata.ResourceMetrics) (exponentialHistogramSnapshot, bool) {
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != t.name {
				continue
			}
			switch data := m.Data.(type) {
			case metricdata.ExponentialHistogram[float64]:
				return matchExponentialDataPoint(t, data.DataPoints)
			case metricdata.ExponentialHistogram[int64]:
				return matchExponentialDataPoint(t, data.DataPoints)
			}
		}
	}
	return exponentialHistogramSnapshot{}, false
}

func matchExponentialDataPoint[N int64 | float64](
	t *rollingHistogramTracker, dps []metricdata.ExponentialHistogramDataPoint[N],
) (exponentialHistogramSnapshot, bool) {
	for _, dp := range dps {
		if rollingHistogramKey(t.name, tagsFromMetricAttributes(dp.Attributes)) == t.key {
			return exponentialHistogramSnapshotFromDataPoint(dp), true
		}
	}
	return exponentialHistogramSnapshot{}, false
}

// observeLocked folds a cumulative snapshot into the window as a delta against the previous snapshot.
func (t *rollingHistogramTracker) observeLocked(current exponentialHistogramSnapshot, now time.Time) {
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
	t.samples = append(t.samples, timedExponentialHistogram{at: now, snapshot: delta})
	t.pruneLocked(now)
}

// observe folds a cumulative snapshot into the window. Used by tests to drive the tracker without a
// real reader.
func (t *rollingHistogramTracker) observe(current exponentialHistogramSnapshot, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.observeLocked(current, now)
}

// count returns the number of observations currently in the rolling window. Used by tests.
func (t *rollingHistogramTracker) count() uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	t.pruneLocked(now)
	var count uint64
	for _, sample := range t.samples {
		count += sample.snapshot.count
	}
	return count
}

func (t *rollingHistogramTracker) pruneLocked(now time.Time) {
	cutoff := now.Add(-t.window)
	firstLive := 0
	for firstLive < len(t.samples) && t.samples[firstLive].at.Before(cutoff) {
		firstLive++
	}
	if firstLive > 0 {
		copy(t.samples, t.samples[firstLive:])
		t.samples = t.samples[:len(t.samples)-firstLive]
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
