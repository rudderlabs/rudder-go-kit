package stats

import (
	"sync/atomic"
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
