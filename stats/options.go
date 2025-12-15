package stats

import (
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

type statsConfig struct {
	enabled             *atomic.Bool
	serviceName         string
	serviceVersion      string
	instanceName        string
	namespaceIdentifier string
	excludedTags        map[string]struct{}

	periodicStatsConfig     periodicStatsConfig
	defaultHistogramBuckets []float64
	histogramBuckets        map[string][]float64
	prometheusRegisterer    prometheus.Registerer
	prometheusGatherer      prometheus.Gatherer

	// Exponential histogram configuration
	useExponentialHistogram     bool
	exponentialHistogramMaxSize int32
	exponentialHistograms       map[string]int32 // per-histogram maxSize
}

// Option is a function used to configure the stats service.
type Option func(*statsConfig)

// WithServiceName sets the service name for the stats service.
func WithServiceName(name string) Option {
	return func(c *statsConfig) {
		c.serviceName = name
	}
}

// WithServiceVersion sets the service version for the stats service.
func WithServiceVersion(version string) Option {
	return func(c *statsConfig) {
		c.serviceVersion = version
	}
}

// WithDefaultHistogramBuckets sets the histogram buckets for the stats service.
func WithDefaultHistogramBuckets(buckets []float64) Option {
	return func(c *statsConfig) {
		c.defaultHistogramBuckets = buckets
	}
}

// WithHistogramBuckets sets the histogram buckets for a measurement.
func WithHistogramBuckets(histogramName string, buckets []float64) Option {
	return func(c *statsConfig) {
		if c.histogramBuckets == nil {
			c.histogramBuckets = make(map[string][]float64)
		}
		c.histogramBuckets[histogramName] = buckets
	}
}

// WithDefaultExponentialHistogram configures all histograms to use exponential bucketing.
// Exponential histograms provide better accuracy and lower memory usage for high-dynamic-range metrics.
// They automatically adapt to the data distribution and are exported as Prometheus native histograms.
// maxSize controls the maximum number of buckets (default: 160, min: 1, max: 160).
// Note: This option is mutually exclusive with WithDefaultHistogramBuckets. The last one applied wins.
func WithDefaultExponentialHistogram(maxSize int32) Option {
	return func(c *statsConfig) {
		c.useExponentialHistogram = true
		c.exponentialHistogramMaxSize = maxSize
	}
}

// WithExponentialHistogram configures a specific histogram to use exponential bucketing.
// Exponential histograms provide better accuracy and lower memory usage for high-dynamic-range metrics.
// They automatically adapt to the data distribution and are exported as Prometheus native histograms.
// maxSize controls the maximum number of buckets (default: 160, min: 1, max: 160).
// Note: This option takes precedence over both WithDefaultHistogramBuckets and WithHistogramBuckets for the specified
// histogram.
func WithExponentialHistogram(histogramName string, maxSize int32) Option {
	return func(c *statsConfig) {
		if c.exponentialHistograms == nil {
			c.exponentialHistograms = make(map[string]int32)
		}
		c.exponentialHistograms[histogramName] = maxSize
	}
}

// WithPrometheusRegistry sets the prometheus registerer and gatherer for the stats service.
// If nil is passed the default ones will be used.
func WithPrometheusRegistry(registerer prometheus.Registerer, gatherer prometheus.Gatherer) Option {
	return func(c *statsConfig) {
		c.prometheusRegisterer = registerer
		c.prometheusGatherer = gatherer
	}
}
