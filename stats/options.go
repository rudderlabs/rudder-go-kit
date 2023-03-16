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
}

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
