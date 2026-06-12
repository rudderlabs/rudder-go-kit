package stats

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

const (
	// Tracking instruments use a fixed high-resolution exponential aggregation. Buckets are not actually
	// read (percentiles are computed from exemplars), but a histogram aggregation is what carries the
	// exemplars, so we keep it cheap and accurate.
	trackingHistogramMaxSize  = 160
	trackingHistogramMaxScale = 20
	// defaultTrackingHistogramMaxSamples bounds how many recent observations are retained per series.
	// It caps both memory and the number of samples a percentile is computed over.
	defaultTrackingHistogramMaxSamples = 2048
	rollingHistogramMeterName          = "github.com/rudderlabs/rudder-go-kit/stats/rollinghistogram"
)

// rollingHistogramRegistry backs Histogram.Percentile for the OpenTelemetry stats. Per histogram series
// it holds a small histogramTracking record, created when the measurement is created but otherwise
// dormant: the OTel pipeline that actually retains observations (a private meter provider with an
// exemplar reservoir) is provisioned lazily, on the first Percentile call for that series. A service
// that never calls Percentile therefore allocates no providers, no reservoirs and does no extra
// recording — only an atomic check per Observe.
type rollingHistogramRegistry struct {
	mu         sync.Mutex
	now        func() time.Time
	maxSamples int
	log        logger.Logger
	providers  map[string]*trackedHistogramProvider // one private provider per measurement name
	series     map[string]*histogramTracking        // one tracking record per series (name|tags)
}

// trackedHistogramProvider is the private OTel pipeline backing a single tracked histogram name. The
// reader, when collected, returns only that one instrument's data (the view matches it by name); both
// fields transitively keep the meter provider's pipeline alive.
type trackedHistogramProvider struct {
	reader     sdkmetric.Reader
	instrument metric.Float64Histogram
}

func newRollingHistogramRegistry(now func() time.Time, maxSamples int, log logger.Logger) *rollingHistogramRegistry {
	if now == nil {
		now = time.Now
	}
	if maxSamples <= 0 {
		maxSamples = defaultTrackingHistogramMaxSamples
	}
	return &rollingHistogramRegistry{
		now:        now,
		maxSamples: maxSamples,
		log:        log,
		providers:  make(map[string]*trackedHistogramProvider),
		series:     make(map[string]*histogramTracking),
	}
}

// tracking returns the shared tracking record for a series, creating it if necessary. The record is
// cheap and dormant until its first Percentile call; sharing it per series means every Measurement for
// the same series records into the same reservoir once tracking is enabled.
func (r *rollingHistogramRegistry) tracking(name string, tags Tags) *histogramTracking {
	if r == nil { // no registry wired (e.g. a directly-constructed otelStats): tracking is unavailable
		return nil
	}
	key := rollingHistogramKey(name, tags)

	r.mu.Lock()
	defer r.mu.Unlock()

	ht, ok := r.series[key]
	if !ok {
		ht = &histogramTracking{registry: r, name: name, tags: tags}
		r.series[key] = ht
	}
	return ht
}

// provider returns the private provider for a name, creating it on first use.
func (r *rollingHistogramRegistry) provider(name string) (*trackedHistogramProvider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tp, ok := r.providers[name]
	if !ok {
		var err error
		tp, err = newTrackedHistogramProvider(name, r.maxSamples)
		if err != nil {
			return nil, err
		}
		r.providers[name] = tp
	}
	return tp, nil
}

func newTrackedHistogramProvider(name string, maxSamples int) (*trackedHistogramProvider, error) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		// Record an exemplar for every observation, not only those made inside a sampled span.
		sdkmetric.WithExemplarFilter(exemplar.AlwaysOnFilter),
		sdkmetric.WithView(sdkmetric.NewView(
			// Match this one instrument by name: the reader then only ever yields this series' data.
			sdkmetric.Instrument{Name: name, Kind: sdkmetric.InstrumentKindHistogram},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationBase2ExponentialHistogram{
					MaxSize:  trackingHistogramMaxSize,
					MaxScale: trackingHistogramMaxScale,
				},
				// Retain the most recent maxSamples observations (with their timestamps) as exemplars,
				// so a rolling-window percentile can be read straight from them.
				ExemplarReservoirProviderSelector: func(sdkmetric.Aggregation) exemplar.ReservoirProvider {
					return func(attribute.Set) exemplar.Reservoir {
						return newWindowReservoir(maxSamples)
					}
				},
			},
		)),
	)

	instrument, err := provider.Meter(rollingHistogramMeterName).Float64Histogram(name)
	if err != nil {
		return nil, fmt.Errorf("creating rolling histogram instrument %q: %w", name, err)
	}
	return &trackedHistogramProvider{reader: reader, instrument: instrument}, nil
}

// histogramTracking is the per-series state behind Histogram.Percentile. It is shared across all
// Measurements for the same series. enabled gates the (lazy) tracking pipeline: until the first
// Percentile call it is false and Observe does no extra work; once enabled, instrument is the dedicated
// instrument observations are mirrored into and tracker reads them back.
type histogramTracking struct {
	registry *rollingHistogramRegistry
	name     string
	tags     Tags

	once       sync.Once
	enabled    atomic.Bool
	instrument metric.Float64Histogram
	tracker    *rollingHistogramTracker
}

// record mirrors an observation into the tracking instrument, but only once tracking has been enabled.
func (h *histogramTracking) record(ctx context.Context, value float64, opts ...metric.RecordOption) {
	if h.enabled.Load() {
		h.instrument.Record(ctx, value, opts...)
	}
}

// percentile enables tracking on first use, then returns the windowed percentile.
func (h *histogramTracking) percentile(p float64, window time.Duration) (float64, bool) {
	h.once.Do(h.enable)
	if !h.enabled.Load() {
		return 0, false
	}
	return h.tracker.percentile(p, window)
}

func (h *histogramTracking) enable() {
	tp, err := h.registry.provider(h.name)
	if err != nil {
		if h.registry.log != nil {
			h.registry.log.Warnn("enabling rolling histogram tracking",
				logger.NewStringField("measurement", h.name), obskit.Error(err))
		}
		return
	}
	h.instrument = tp.instrument
	h.tracker = &rollingHistogramTracker{
		reader: tp.reader,
		name:   h.name,
		key:    rollingHistogramKey(h.name, h.tags),
		now:    h.registry.now,
	}
	h.enabled.Store(true)
}

// rollingHistogramTracker is immutable after creation, so it needs no locking: percentile only reads
// the (concurrency-safe) reader and the exemplars it returns.
type rollingHistogramTracker struct {
	reader sdkmetric.Reader // private reader for this measurement's name; nil only in unit tests
	name   string           // instrument name to find in the collected metrics
	key    string           // name|tags identity of the exact series this tracker follows
	now    func() time.Time
}

// percentile returns the p-th percentile (p in [0,100]) over the last window and true when the window
// holds observations; (0, false) otherwise. It collects the private reader, walks this series'
// exemplars from newest to oldest, stops at now-window, and computes a nearest-rank percentile over the
// observed values. Nothing is retained between calls.
func (t *rollingHistogramTracker) percentile(p float64, window time.Duration) (float64, bool) {
	if p < 0 || p > 100 || math.IsNaN(p) || window <= 0 {
		return 0, false
	}
	if t.reader == nil {
		return 0, false
	}

	var rm metricdata.ResourceMetrics
	if err := t.reader.Collect(context.Background(), &rm); err != nil {
		return 0, false
	}
	values := t.windowValues(&rm, window)
	if len(values) == 0 {
		return 0, false
	}
	return nearestRankPercentile(values, p), true
}

// windowValues finds this tracker's series among the collected metrics and returns the values of its
// observations made within the last window.
func (t *rollingHistogramTracker) windowValues(rm *metricdata.ResourceMetrics, window time.Duration) []float64 {
	cutoff := t.now().Add(-window)
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != t.name {
				continue
			}
			switch data := m.Data.(type) {
			case metricdata.ExponentialHistogram[float64]:
				return windowedExemplarValues(t, data.DataPoints, cutoff)
			case metricdata.ExponentialHistogram[int64]:
				return windowedExemplarValues(t, data.DataPoints, cutoff)
			}
		}
	}
	return nil
}

// windowedExemplarValues returns the values of this tracker's series' exemplars that are not older than
// cutoff. Exemplars are stored in observation order, so it walks them from the most recent backwards and
// stops as soon as one falls outside the window.
func windowedExemplarValues[N int64 | float64](
	t *rollingHistogramTracker, dps []metricdata.ExponentialHistogramDataPoint[N], cutoff time.Time,
) []float64 {
	for _, dp := range dps {
		if rollingHistogramKey(t.name, tagsFromMetricAttributes(dp.Attributes)) != t.key {
			continue
		}
		values := make([]float64, 0, len(dp.Exemplars))
		for i := len(dp.Exemplars) - 1; i >= 0; i-- {
			if dp.Exemplars[i].Time.Before(cutoff) {
				break
			}
			values = append(values, float64(dp.Exemplars[i].Value))
		}
		return values
	}
	return nil
}

// nearestRankPercentile returns the p-th percentile (p in [0,100]) of values using the nearest-rank
// method. values is sorted in place; it must be non-empty.
func nearestRankPercentile(values []float64, p float64) float64 {
	sort.Float64s(values)
	rank := int(math.Ceil(p/100*float64(len(values)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(values) {
		rank = len(values) - 1
	}
	return values[rank]
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

// windowReservoir is a fixed-capacity ring of the most recent observations (timestamp + value),
// exposed to OTel as an exemplar reservoir. OTel offers every observation to it (AlwaysOn filter) and
// reads it back on Collect; the reader-side window filter (see windowedExemplarValues) is what makes
// stale observations drop out, so this only needs to bound memory. It does not embed the SDK's internal
// reservoir.ConcurrentSafe marker, so the SDK already serializes Offer/Collect — no locking is needed
// here.
type windowReservoir struct {
	times  []time.Time
	values []exemplar.Value
	next   int  // next write position
	full   bool // whether the ring has wrapped at least once
}

func newWindowReservoir(capacity int) *windowReservoir {
	return &windowReservoir{
		times:  make([]time.Time, capacity),
		values: make([]exemplar.Value, capacity),
	}
}

func (r *windowReservoir) Offer(_ context.Context, t time.Time, v exemplar.Value, _ []attribute.KeyValue) {
	r.times[r.next] = t
	r.values[r.next] = v
	r.next++
	if r.next == len(r.times) {
		r.next = 0
		r.full = true
	}
}

// Collect emits the held observations oldest-first. It is non-destructive: state is preserved so
// successive reads see the same (windowed) observations until they are overwritten by newer ones.
func (r *windowReservoir) Collect(dest *[]exemplar.Exemplar) {
	*dest = (*dest)[:0]
	emit := func(i int) {
		*dest = append(*dest, exemplar.Exemplar{Time: r.times[i], Value: r.values[i]})
	}
	if r.full {
		for i := r.next; i < len(r.times); i++ {
			emit(i)
		}
	}
	for i := 0; i < r.next; i++ {
		emit(i)
	}
}
