package stats

import (
	"context"
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
	// Percentile instruments use a fixed high-resolution exponential aggregation.
	// Buckets are not actually read (percentiles are computed from exemplars), but a histogram aggregation is what
	// carries the exemplars, so we keep it cheap and accurate.
	percentileHistogramMaxSize  = 160
	percentileHistogramMaxScale = 20
	// defaultPercentileMaxSamples bounds how many recent observations (exemplars) are retained per series.
	// It caps both memory and the number of samples a percentile is computed over.
	defaultPercentileMaxSamples = 512
	percentileMeterName         = "github.com/rudderlabs/rudder-go-kit/stats/percentile"
)

// percentileRegistry holds the shared configuration for Histogram.Percentile and builds per-series state on demand.
// The per-series records themselves are owned by the cached otelHistogram measurements (see otelStats.getHistogram),
// so a service that never calls Percentile() allocates no providers, no reservoirs and does no extra recording, only an
// atomic check per Observe.
type percentileRegistry struct {
	now        func() time.Time
	maxSamples int
	log        logger.Logger
}

func newPercentileRegistry(now func() time.Time, maxSamples int, log logger.Logger) *percentileRegistry {
	if now == nil {
		now = time.Now
	}
	if maxSamples <= 0 {
		maxSamples = defaultPercentileMaxSamples
	}
	return &percentileRegistry{
		now:        now,
		maxSamples: maxSamples,
		log:        log,
	}
}

// newSeries builds a dormant percentileSeries. Sharing one series across every Measurement of the same
// series is handled by the measurement cache in otelStats (keyed by name|tags).
func (r *percentileRegistry) newSeries(name string) *percentileSeries {
	if r == nil { // no registry wired (e.g. a directly-constructed otelStats): percentiles unavailable
		return nil
	}
	return &percentileSeries{registry: r, name: name}
}

// percentileSeries is the per-series state behind Histogram.Percentile, shared across all Measurements
// for the same series. It owns a private, single-instrument meter provider so that collecting its reader
// yields exactly one data point — the series' own — with no attribute matching on the read path. enabled
// gates the (lazy) pipeline: until the first Percentile call it is false and Observe does no extra work;
// once enabled, observations are mirrored into instrument and read back from reader.
type percentileSeries struct {
	registry *percentileRegistry
	name     string

	once       sync.Once
	enabled    atomic.Bool
	provider   *sdkmetric.MeterProvider
	instrument metric.Float64Histogram
	reader     sdkmetric.Reader
}

// record mirrors an observation into the percentile instrument, but only once the series is enabled. No
// attributes are recorded: the provider is private to this one series, so a single data point holds all
// of its observations.
func (ps *percentileSeries) record(ctx context.Context, value float64) {
	if ps.enabled.Load() {
		ps.instrument.Record(ctx, value)
	}
}

// compute enables the series on first use, then returns the p-th percentile (p in [0,100]) over the last
// window and true when the window holds observations; (0, false) otherwise. It collects the private
// reader, walks the series' exemplars newest → oldest stopping at now-window, and computes a nearest-rank
// percentile. Nothing is retained between calls.
func (ps *percentileSeries) compute(p float64, window time.Duration) (float64, bool) {
	if p < 0 || p > 100 || math.IsNaN(p) || window <= 0 {
		return 0, false
	}
	ps.once.Do(ps.enable)
	if !ps.enabled.Load() {
		return 0, false
	}

	var rm metricdata.ResourceMetrics
	if err := ps.reader.Collect(context.Background(), &rm); err != nil {
		return 0, false
	}
	values := windowValues(&rm, ps.registry.now().Add(-window))
	if len(values) == 0 {
		return 0, false
	}
	return nearestRankPercentile(values, p), true
}

func (ps *percentileSeries) enable() {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		// Record an exemplar for every observation, not only those made inside a sampled span.
		sdkmetric.WithExemplarFilter(exemplar.AlwaysOnFilter),
		sdkmetric.WithView(sdkmetric.NewView(
			sdkmetric.Instrument{Name: ps.name, Kind: sdkmetric.InstrumentKindHistogram},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationBase2ExponentialHistogram{
					MaxSize:  percentileHistogramMaxSize,
					MaxScale: percentileHistogramMaxScale,
				},
				// Retain the most recent maxSamples observations (with their timestamps) as exemplars,
				// so a rolling-window percentile can be read straight from them.
				ExemplarReservoirProviderSelector: func(sdkmetric.Aggregation) exemplar.ReservoirProvider {
					return func(attribute.Set) exemplar.Reservoir {
						return newWindowReservoir(ps.registry.maxSamples)
					}
				},
			},
		)),
	)

	instrument, err := provider.Meter(percentileMeterName).Float64Histogram(ps.name)
	if err != nil {
		if ps.registry.log != nil {
			ps.registry.log.Warnn("Enabling histogram percentile tracking",
				logger.NewStringField("measurement", ps.name), obskit.Error(err))
		}
		return
	}
	ps.provider = provider
	ps.instrument = instrument
	ps.reader = reader
	ps.enabled.Store(true)
}

// shutdown releases the series' private meter provider (and its reader). It is safe to call on a nil
// series or one that was never enabled (both no-op). After shutdown the reader is closed, so compute()
// simply reports no data rather than panicking.
func (ps *percentileSeries) shutdown(ctx context.Context) error {
	if ps == nil || !ps.enabled.Load() {
		return nil
	}
	return ps.provider.Shutdown(ctx)
}

// windowValues returns the values of the tracked series' exemplars made within the last window (cutoff =
// now-window). The reader belongs to a private, single-instrument provider, so the collected metric is
// always this series' — there is exactly one data point and no attributes to match.
func windowValues(rm *metricdata.ResourceMetrics, cutoff time.Time) []float64 {
	for _, scope := range rm.ScopeMetrics {
		for _, m := range scope.Metrics {
			switch data := m.Data.(type) {
			case metricdata.ExponentialHistogram[float64]:
				return exemplarValuesSince(data.DataPoints, cutoff)
			case metricdata.ExponentialHistogram[int64]:
				return exemplarValuesSince(data.DataPoints, cutoff)
			}
		}
	}
	return nil
}

// exemplarValuesSince returns the values of the single data point's exemplars not older than cutoff.
// Exemplars are stored in observation order, so it walks them from the most recent backwards and stops
// as soon as one falls outside the window.
func exemplarValuesSince[N int64 | float64](
	dps []metricdata.ExponentialHistogramDataPoint[N], cutoff time.Time,
) []float64 {
	if len(dps) == 0 {
		return nil
	}
	exemplars := dps[0].Exemplars // a per-series provider yields exactly one data point
	values := make([]float64, 0, len(exemplars))
	for i := len(exemplars) - 1; i >= 0; i-- {
		if exemplars[i].Time.Before(cutoff) {
			break
		}
		values = append(values, float64(exemplars[i].Value))
	}
	return values
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

// windowReservoir is a fixed-capacity ring of the most recent observations (timestamp + value), exposed
// to OTel as an exemplar reservoir. OTel offers every observation to it (AlwaysOn filter) and reads it
// back on Collect; the reader-side window filter (see windowValues) is what makes stale observations
// drop out, so this only needs to bound memory.
// It does not embed the SDK's internal reservoir.ConcurrentSafe marker, so the SDK already serializes Offer/Collect,
// no locking is needed here.
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
//
// NOTE: Exemplars are concrete example measurements that an aggregated metric keeps on the side.
//
// The problem they solve:
//
//	When you record into a histogram, the individual observations are thrown away. That's the whole point of aggregation.
//	A bucket ends up saying "47 observations fell between 100ms and 200ms," but it has lost which 47, and any detail about them.
//	That's great for cheap storage and dashboards, but terrible when you're staring at a spike and asking "okay, but
//	show me an actual one of those slow requests."
//
//	An exemplar is the metrics SDK keeping a few real, individual observations attached to the aggregate so you don't
//	lose that thread entirely.
//
//	So each exemplar is one recorded sample — not a bucket count, not an average — with its timestamp and the value.
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
