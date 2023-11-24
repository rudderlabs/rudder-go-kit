package otel

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type (
	// Option allows to configure the OpenTelemetry initialization
	Option func(*config)
	// TracerProviderOption allows to configure the tracer provider
	TracerProviderOption func(providerConfig *tracerProviderConfig)
	// MeterProviderOption allows to configure the meter provider
	MeterProviderOption func(providerConfig *meterProviderConfig)
	// SpanExporter can be used to set a custom span exporter (e.g. for testing purposes)
	SpanExporter = sdktrace.SpanExporter
)

// WithRetryConfig allows to set the retry configuration
func WithRetryConfig(rc RetryConfig) Option {
	return func(c *config) {
		c.retryConfig = &rc
	}
}

// WithInsecure allows to set the GRPC connection to be insecure
func WithInsecure() Option {
	return func(c *config) {
		// Note the use of insecure transport here. TLS is recommended in production.
		c.withInsecure = true
	}
}

// WithTextMapPropagator allows to set the text map propagator
// e.g. propagation.TraceContext{}
func WithTextMapPropagator(tmp propagation.TextMapPropagator) Option {
	return func(c *config) {
		c.tracerProviderConfig.textMapPropagator = tmp
	}
}

// WithCustomTracerProvider forces the usage of a custom exporter for the tracer provider
func WithCustomTracerProvider(se SpanExporter, opts ...TracerProviderOption) Option {
	return func(c *config) {
		c.tracerProviderConfig.enabled = true
		c.tracerProviderConfig.customSpanExporter = se
		for _, opt := range opts {
			opt(&c.tracerProviderConfig)
		}
	}
}

// WithTracerProvider allows to set the tracer provider and specify if it should be the global one
func WithTracerProvider(endpoint string, opts ...TracerProviderOption) Option {
	return func(c *config) {
		c.tracerProviderConfig.enabled = true
		c.tracesEndpoint = endpoint
		for _, opt := range opts {
			opt(&c.tracerProviderConfig)
		}
	}
}

// WithTracingSamplingRate allows to set the sampling rate for the tracer provider
// samplingRate >= 1 will always sample.
// samplingRate < 0 is treated as zero.
func WithTracingSamplingRate(rate float64) TracerProviderOption {
	return func(c *tracerProviderConfig) {
		c.samplingRate = rate
	}
}

// WithTracingSyncer lets you register the exporter with a synchronous SimpleSpanProcessor (e.g. instead of a batching
// asynchronous one).
// NOT RECOMMENDED FOR PRODUCTION USE (use for testing and debugging only).
func WithTracingSyncer() TracerProviderOption {
	return func(c *tracerProviderConfig) {
		c.withSyncer = true
	}
}

// WithZipkin allows to set the tracer provider to use Zipkin
// This means that the SDK will send the data to Zipkin directly instead of using the collector.
func WithZipkin() TracerProviderOption {
	return func(c *tracerProviderConfig) {
		c.withZipkin = true
	}
}

// WithGlobalTracerProvider allows to set the tracer provider as the global one
func WithGlobalTracerProvider() TracerProviderOption {
	return func(c *tracerProviderConfig) {
		c.global = true
	}
}

// WithMeterProvider allows to set the meter provider and specify if it should be the global one plus other options.
func WithMeterProvider(opts ...MeterProviderOption) Option {
	return func(c *config) {
		c.meterProviderConfig.enabled = true
		for _, opt := range opts {
			opt(&c.meterProviderConfig)
		}
	}
}

// WithGlobalMeterProvider allows to set the meter provider as the global one
func WithGlobalMeterProvider() MeterProviderOption {
	return func(c *meterProviderConfig) {
		c.global = true
	}
}

// WithGRPCMeterProvider allows to set the meter provider to use GRPC
func WithGRPCMeterProvider(grpcEndpoint string) MeterProviderOption {
	return func(c *meterProviderConfig) {
		c.grpcEndpoint = &grpcEndpoint
	}
}

// WithMeterProviderExportsInterval configures the intervening time between exports (if less than or equal to zero,
// 60 seconds is used)
func WithMeterProviderExportsInterval(interval time.Duration) MeterProviderOption {
	return func(c *meterProviderConfig) {
		c.exportsInterval = interval
	}
}

// WithPrometheusExporter allows to enable the Prometheus exporter
func WithPrometheusExporter(registerer prometheus.Registerer) MeterProviderOption {
	return func(c *meterProviderConfig) {
		c.prometheusRegisterer = registerer
	}
}

// WithDefaultHistogramBucketBoundaries lets you overwrite the default buckets for all histograms.
func WithDefaultHistogramBucketBoundaries(boundaries []float64) MeterProviderOption {
	return func(c *meterProviderConfig) {
		c.defaultHistogramBuckets = sdkmetric.NewView(
			sdkmetric.Instrument{
				Kind: sdkmetric.InstrumentKindHistogram,
			},
			sdkmetric.Stream{
				Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
					Boundaries: boundaries,
				},
			},
		)
	}
}

// WithHistogramBucketBoundaries allows the creation of a view to overwrite the default buckets of a given histogram.
// meterName is optional.
func WithHistogramBucketBoundaries(instrumentName, meterName string, boundaries []float64) MeterProviderOption {
	var scope instrumentation.Scope
	if meterName != "" {
		scope.Name = meterName
	}
	newView := sdkmetric.NewView(
		sdkmetric.Instrument{
			Name:  instrumentName,
			Scope: scope,
			Kind:  sdkmetric.InstrumentKindHistogram,
		},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: boundaries,
			},
		},
	)
	return func(c *meterProviderConfig) {
		c.views = append(c.views, newView)
	}
}

// WithLogger allows to set the logger
func WithLogger(l logger) Option {
	return func(c *config) { c.logger = l }
}
