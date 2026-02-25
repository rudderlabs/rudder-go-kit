// Package prometheus is imported from the official OpenTelemetry package:
// https://github.com/open-telemetry/opentelemetry-go/tree/v1.38.0/exporters/prometheus
// The version of the exporter would be v0.60.0 (not v1.38.0, see releases).
//
// Customisations applied:
//
//  1. scope info keys are not "otel_scope_name", "otel_scope_version" but we're now using the semconv ones to be
//     consistent if we switch over to gRPC. we're propagating them via the *resource.Resource (see Collect method)
//     see here: https://github.com/open-telemetry/opentelemetry-go/blob/v1.14.0/exporters/prometheus/exporter.go#L48
//
//  2. prometheus counters MUST have a _total suffix but that breaks our dashboards, so we removed it
//     see here: https://github.com/open-telemetry/opentelemetry-go/blob/v1.14.0/exporters/prometheus/exporter.go#L73
//
//  3. a global logger was used, we made it injectable via options
//     see here: https://github.com/open-telemetry/opentelemetry-go/blob/v1.14.0/exporters/prometheus/exporter.go#L393
//
//  4. removed unnecessary otel_scope_info metric
//
// Features added from v1.38.0/v0.60.0:
//
//  5. exponential histogram support - converts OTel exponential histograms to Prometheus native histograms
//     with scale conversion (supports scales -4 to 8)
//
//  6. exemplar support - attaches trace_id and span_id to histogram and counter metrics for distributed
//     tracing correlation
package prometheus

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
)

const (
	targetInfoMetricName  = "target_info"
	targetInfoDescription = "Target metadata"
)

// Exporter is a Prometheus Exporter that embeds the OTel metric.Reader
// interface for easy instantiation with a MeterProvider.
type Exporter struct {
	metric.Reader
}

var _ metric.Reader = &Exporter{}

// collector is used to implement prometheus.Collector.
type collector struct {
	reader metric.Reader
	logger logger

	disableTargetInfo    bool
	withoutUnits         bool
	targetInfo           prometheus.Metric
	createTargetInfoOnce sync.Once
	scopeInfos           map[instrumentation.Scope]prometheus.Metric
	metricFamilies       map[string]*dto.MetricFamily
	namespace            string
}

// New returns a Prometheus Exporter.
func New(opts ...Option) (*Exporter, error) {
	cfg := newConfig(opts...)

	// this assumes that the default temporality selector will always return cumulative.
	// we only support cumulative temporality, so building our own reader enforces this.
	reader := metric.NewManualReader(cfg.manualReaderOptions()...)

	collector := &collector{
		reader:            reader,
		logger:            cfg.logger,
		disableTargetInfo: cfg.disableTargetInfo,
		withoutUnits:      cfg.withoutUnits,
		scopeInfos:        make(map[instrumentation.Scope]prometheus.Metric),
		metricFamilies:    make(map[string]*dto.MetricFamily),
		namespace:         cfg.namespace,
	}

	if err := cfg.registerer.Register(collector); err != nil {
		return nil, fmt.Errorf("cannot register the collector: %w", err)
	}

	e := &Exporter{
		Reader: reader,
	}

	return e, nil
}

// Describe implements prometheus.Collector.
func (c *collector) Describe(_ chan<- *prometheus.Desc) {}

// Collect implements prometheus.Collector.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	metrics := metricdata.ResourceMetrics{}
	err := c.reader.Collect(context.TODO(), &metrics)
	if err != nil {
		otel.Handle(err)
		if errors.Is(err, metric.ErrReaderNotRegistered) {
			return
		}
	}

	c.createTargetInfoOnce.Do(func() {
		// Resource should be immutable, we don't need to compute again
		targetInfo, err := c.createInfoMetric(targetInfoMetricName, targetInfoDescription, metrics.Resource)
		if err != nil {
			// If the target info metric is invalid, disable sending it.
			otel.Handle(err)
			c.disableTargetInfo = true
		}
		c.targetInfo = targetInfo
	})
	if !c.disableTargetInfo {
		ch <- c.targetInfo
	}

	var scopeKeys, scopeValues []string
	for _, attr := range metrics.Resource.Attributes() {
		if string(attr.Key) == string(semconv.ServiceNameKey) {
			scopeKeys = append(scopeKeys, "job")
			scopeValues = append(scopeValues, attr.Value.AsString())
		}
		scopeKeys = append(scopeKeys, strings.Map(sanitizeRune, string(attr.Key)))
		scopeValues = append(scopeValues, attr.Value.AsString())
	}

	for _, scopeMetrics := range metrics.ScopeMetrics {
		for _, m := range scopeMetrics.Metrics {
			switch v := m.Data.(type) {
			case metricdata.Histogram[int64]:
				addHistogramMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			case metricdata.Histogram[float64]:
				addHistogramMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			case metricdata.ExponentialHistogram[int64]:
				addExponentialHistogramMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			case metricdata.ExponentialHistogram[float64]:
				addExponentialHistogramMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			case metricdata.Sum[int64]:
				addSumMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			case metricdata.Sum[float64]:
				addSumMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			case metricdata.Gauge[int64]:
				addGaugeMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			case metricdata.Gauge[float64]:
				addGaugeMetric(ch, v, m, scopeKeys, scopeValues, c.getName(m), c.metricFamilies, c.logger)
			}
		}
	}
}

func addHistogramMetric[N int64 | float64](
	ch chan<- prometheus.Metric, histogram metricdata.Histogram[N], m metricdata.Metrics,
	ks, vs []string, name string, mfs map[string]*dto.MetricFamily, l logger,
) {
	drop, help := validateMetrics(name, m.Description, dto.MetricType_HISTOGRAM.Enum(), mfs, l)
	if drop {
		return
	}
	if help != "" {
		m.Description = help
	}

	for _, dp := range histogram.DataPoints {
		keys, values := getAttrs(dp.Attributes, ks, vs)

		desc := prometheus.NewDesc(name, m.Description, keys, nil)
		buckets := make(map[float64]uint64, len(dp.Bounds))

		cumulativeCount := uint64(0)
		for i, bound := range dp.Bounds {
			cumulativeCount += dp.BucketCounts[i]
			buckets[bound] = cumulativeCount
		}
		m, err := prometheus.NewConstHistogram(desc, dp.Count, float64(dp.Sum), buckets, values...)
		if err != nil {
			otel.Handle(err)
			continue
		}
		m = addExemplars(m, dp.Exemplars)
		ch <- m
	}
}

func addSumMetric[N int64 | float64](
	ch chan<- prometheus.Metric, sum metricdata.Sum[N], m metricdata.Metrics,
	ks, vs []string, name string, mfs map[string]*dto.MetricFamily, l logger,
) {
	valueType := prometheus.CounterValue
	metricType := dto.MetricType_COUNTER
	if !sum.IsMonotonic {
		valueType = prometheus.GaugeValue
		metricType = dto.MetricType_GAUGE
	}

	drop, help := validateMetrics(name, m.Description, metricType.Enum(), mfs, l)
	if drop {
		return
	}
	if help != "" {
		m.Description = help
	}

	for _, dp := range sum.DataPoints {
		keys, values := getAttrs(dp.Attributes, ks, vs)

		desc := prometheus.NewDesc(name, m.Description, keys, nil)
		m, err := prometheus.NewConstMetric(desc, valueType, float64(dp.Value), values...)
		if err != nil {
			otel.Handle(err)
			continue
		}
		// Only add exemplars for counters (monotonic metrics), not gauges
		if valueType != prometheus.GaugeValue {
			m = addExemplars(m, dp.Exemplars)
		}
		ch <- m
	}
}

func addGaugeMetric[N int64 | float64](
	ch chan<- prometheus.Metric, gauge metricdata.Gauge[N], m metricdata.Metrics,
	ks, vs []string, name string, mfs map[string]*dto.MetricFamily, l logger,
) {
	drop, help := validateMetrics(name, m.Description, dto.MetricType_GAUGE.Enum(), mfs, l)
	if drop {
		return
	}
	if help != "" {
		m.Description = help
	}

	for _, dp := range gauge.DataPoints {
		keys, values := getAttrs(dp.Attributes, ks, vs)

		desc := prometheus.NewDesc(name, m.Description, keys, nil)
		m, err := prometheus.NewConstMetric(desc, prometheus.GaugeValue, float64(dp.Value), values...)
		if err != nil {
			otel.Handle(err)
			continue
		}
		ch <- m
	}
}

// addExemplars attaches exemplars (trace context) to a Prometheus metric.
// Exemplars enable correlation between metrics and distributed traces in Prometheus.
func addExemplars[N int64 | float64](
	m prometheus.Metric,
	exemplars []metricdata.Exemplar[N],
) prometheus.Metric {
	if len(exemplars) == 0 {
		return m
	}

	promExemplars := make([]prometheus.Exemplar, len(exemplars))
	for i, exemplar := range exemplars {
		labels := make(map[string]string)

		// Convert filtered attributes to labels
		for _, kv := range exemplar.FilteredAttributes {
			key := strings.Map(sanitizeRune, string(kv.Key))
			labels[key] = kv.Value.Emit()
		}

		// Add trace context - these are the standard Prometheus exemplar labels
		// for distributed tracing correlation
		if len(exemplar.TraceID) > 0 {
			labels["trace_id"] = hex.EncodeToString(exemplar.TraceID)
		}
		if len(exemplar.SpanID) > 0 {
			labels["span_id"] = hex.EncodeToString(exemplar.SpanID)
		}

		promExemplars[i] = prometheus.Exemplar{
			Value:     float64(exemplar.Value),
			Timestamp: exemplar.Time,
			Labels:    labels,
		}
	}

	metricWithExemplar, err := prometheus.NewMetricWithExemplars(m, promExemplars...)
	if err != nil {
		otel.Handle(err)
		return m
	}
	return metricWithExemplar
}

// downscaleExponentialBucket downscales an exponential histogram bucket by the given scale delta.
// This is used to convert high-scale OTel exponential histograms to Prometheus native histograms
// which support scales from -4 to 8.
func downscaleExponentialBucket(bucket metricdata.ExponentialBucket, scaleDelta int32) metricdata.ExponentialBucket {
	if len(bucket.Counts) == 0 || scaleDelta < 1 {
		return metricdata.ExponentialBucket{
			Offset: bucket.Offset >> scaleDelta,
			Counts: append([]uint64(nil), bucket.Counts...),
		}
	}

	newOffset := bucket.Offset >> scaleDelta
	lastBucketIdx := bucket.Offset + int32(len(bucket.Counts)) - 1
	lastNewIdx := lastBucketIdx >> scaleDelta
	newBucketCount := int(lastNewIdx - newOffset + 1)

	if newBucketCount <= 0 {
		return metricdata.ExponentialBucket{
			Offset: newOffset,
			Counts: []uint64{},
		}
	}

	newCounts := make([]uint64, newBucketCount)

	for i, count := range bucket.Counts {
		if count == 0 {
			continue
		}
		originalIdx := bucket.Offset + int32(i)
		newIdx := originalIdx >> scaleDelta
		position := newIdx - newOffset
		if position >= 0 && position < int32(len(newCounts)) {
			newCounts[position] += count
		}
	}

	return metricdata.ExponentialBucket{
		Offset: newOffset,
		Counts: newCounts,
	}
}

// addExponentialHistogramMetric converts OTel exponential histograms to Prometheus native histograms.
// Exponential histograms provide better accuracy and performance for high-dynamic-range data.
func addExponentialHistogramMetric[N int64 | float64](
	ch chan<- prometheus.Metric, histogram metricdata.ExponentialHistogram[N], m metricdata.Metrics,
	ks, vs []string, name string, mfs map[string]*dto.MetricFamily, l logger,
) {
	drop, help := validateMetrics(name, m.Description, dto.MetricType_HISTOGRAM.Enum(), mfs, l)
	if drop {
		return
	}
	if help != "" {
		m.Description = help
	}

	for _, dp := range histogram.DataPoints {
		keys, values := getAttrs(dp.Attributes, ks, vs)
		desc := prometheus.NewDesc(name, m.Description, keys, nil)

		scale := dp.Scale
		if scale < -4 {
			otel.Handle(fmt.Errorf(
				"exponential histogram scale %d is below minimum supported scale -4, skipping data point",
				scale))
			continue
		}

		positiveBucket := dp.PositiveBucket
		negativeBucket := dp.NegativeBucket
		if scale > 8 {
			scaleDelta := scale - 8
			positiveBucket = downscaleExponentialBucket(dp.PositiveBucket, scaleDelta)
			negativeBucket = downscaleExponentialBucket(dp.NegativeBucket, scaleDelta)
			scale = 8
		}

		positiveBuckets := make(map[int]int64)
		for i, c := range positiveBucket.Counts {
			if c > math.MaxInt64 {
				otel.Handle(fmt.Errorf("positive count %d is too large to be represented as int64", c))
				continue
			}
			positiveBuckets[int(positiveBucket.Offset)+i+1] = int64(c)
		}

		negativeBuckets := make(map[int]int64)
		for i, c := range negativeBucket.Counts {
			if c > math.MaxInt64 {
				otel.Handle(fmt.Errorf("negative count %d is too large to be represented as int64", c))
				continue
			}
			negativeBuckets[int(negativeBucket.Offset)+i+1] = int64(c)
		}

		m, err := prometheus.NewConstNativeHistogram(
			desc,
			dp.Count,
			float64(dp.Sum),
			positiveBuckets,
			negativeBuckets,
			dp.ZeroCount,
			scale,
			dp.ZeroThreshold,
			dp.StartTime,
			values...)
		if err != nil {
			otel.Handle(err)
			continue
		}
		m = addExemplars(m, dp.Exemplars)
		ch <- m
	}
}

// getAttrs parses the attribute.Set to two lists of matching Prometheus-style
// keys and values. It sanitizes invalid characters and handles duplicate keys
// (due to sanitization) by sorting and concatenating the values following the spec.
func getAttrs(attrs attribute.Set, ks, vs []string) ([]string, []string) {
	keysMap := make(map[string][]string)
	itr := attrs.Iter()
	for itr.Next() {
		kv := itr.Attribute()
		key := strings.Map(sanitizeRune, string(kv.Key))
		if _, ok := keysMap[key]; !ok {
			keysMap[key] = []string{kv.Value.Emit()}
		} else {
			// if the sanitized key is a duplicate, append to the list of keys
			keysMap[key] = append(keysMap[key], kv.Value.Emit())
		}
	}

	keys := make([]string, 0, attrs.Len())
	values := make([]string, 0, attrs.Len())
	for key, vals := range keysMap {
		keys = append(keys, key)
		sort.Slice(vals, func(i, j int) bool {
			return i < j
		})
		values = append(values, strings.Join(vals, ";"))
	}

	if len(ks) > 0 {
		keys = append(keys, ks[:]...)
		values = append(values, vs[:]...)
	}
	return keys, values
}

func (c *collector) createInfoMetric(name, description string, res *resource.Resource) (prometheus.Metric, error) {
	keys, values := getAttrs(*res.Set(), []string{}, []string{})
	desc := prometheus.NewDesc(name, description, keys, nil)
	return prometheus.NewConstMetric(desc, prometheus.GaugeValue, float64(1), values...)
}

// BEWARE that we are already sanitizing metric names in the OTel adapter, see sanitizeTagKey function,
// but we still need this function to sanitize metrics coming from the internal OpenTelemetry client
func sanitizeRune(r rune) rune {
	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ':' || r == '_' {
		return r
	}
	return '_'
}

var unitSuffixes = map[string]string{
	"1":  "_ratio",
	"By": "_bytes",
	"ms": "_milliseconds",
}

// getName returns the sanitized name, prefixed with the namespace and suffixed with unit.
func (c *collector) getName(m metricdata.Metrics) string {
	name := sanitizeName(m.Name)
	if c.namespace != "" {
		name = c.namespace + name
	}
	if c.withoutUnits {
		return name
	}
	if suffix, ok := unitSuffixes[m.Unit]; ok {
		name += suffix
	}
	return name
}

func sanitizeName(n string) string {
	// This algorithm is based on strings.Map from Go 1.19.
	const replacement = '_'

	valid := func(i int, r rune) bool {
		// Taken from
		// https://github.com/prometheus/common/blob/dfbc25bd00225c70aca0d94c3c4bb7744f28ace0/model/metric.go#L92-L102
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == ':' || (r >= '0' && r <= '9' && i > 0) {
			return true
		}
		return false
	}

	// This output buffer b is initialized on demand, the first time a
	// character needs to be replaced.
	var b strings.Builder
	for i, c := range n {
		if valid(i, c) {
			continue
		}

		if i == 0 && c >= '0' && c <= '9' {
			// Prefix leading number with replacement character.
			b.Grow(len(n) + 1)
			_ = b.WriteByte(byte(replacement))
			break
		}
		b.Grow(len(n))
		_, _ = b.WriteString(n[:i])
		_ = b.WriteByte(byte(replacement))
		width := utf8.RuneLen(c)
		n = n[i+width:]
		break
	}

	// Fast path for unchanged input.
	if b.Cap() == 0 { // b.Grow was not called above.
		return n
	}

	for _, c := range n {
		// Due to inlining, it is more performant to invoke WriteByte rather then
		// WriteRune.
		if valid(1, c) { // We are guaranteed to not be at the start.
			_ = b.WriteByte(byte(c))
		} else {
			_ = b.WriteByte(byte(replacement))
		}
	}

	return b.String()
}

func validateMetrics(
	name, description string, metricType *dto.MetricType, mfs map[string]*dto.MetricFamily, l logger,
) (drop bool, help string) {
	emf, exist := mfs[name]
	if !exist {
		mfs[name] = &dto.MetricFamily{
			Name: new(name),
			Help: new(description),
			Type: metricType,
		}
		return false, ""
	}
	if emf.GetType() != *metricType {
		l.Error(
			errors.New("instrument type conflict"),
			"Using existing type definition.",
			"instrument", name,
			"existing", emf.GetType(),
			"dropped", *metricType,
		)
		return true, ""
	}
	if emf.GetHelp() != description {
		l.Info(
			"Instrument description conflict, using existing",
			"instrument", name,
			"existing", emf.GetHelp(),
			"dropped", description,
		)
		return false, emf.GetHelp()
	}

	return false, ""
}
