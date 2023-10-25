package otel

import (
	"context"
	"fmt"
	"time"

	promClient "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"golang.org/x/sync/errgroup"

	"github.com/rudderlabs/rudder-go-kit/stats/internal/otel/prometheus"
)

// DefaultRetryConfig represents the default retry configuration
var DefaultRetryConfig = RetryConfig{
	Enabled:         true,
	InitialInterval: 5 * time.Second,
	MaxInterval:     30 * time.Second,
	MaxElapsedTime:  time.Minute,
}

type Manager struct {
	tp *sdktrace.TracerProvider
	mp *sdkmetric.MeterProvider
}

// Setup simplifies the creation of tracer and meter providers with GRPC
func (m *Manager) Setup(
	ctx context.Context, res *resource.Resource, opts ...Option,
) (
	*sdktrace.TracerProvider,
	*sdkmetric.MeterProvider,
	error,
) {
	var c config
	for _, opt := range opts {
		opt(&c)
	}
	if c.retryConfig == nil {
		c.retryConfig = &DefaultRetryConfig
	}
	if c.logger == nil {
		c.logger = nopLogger{}
	}

	if !c.tracerProviderConfig.enabled && !c.meterProviderConfig.enabled {
		return nil, nil, fmt.Errorf("no trace provider or meter provider to initialize")
	}

	if c.tracerProviderConfig.enabled {
		if c.tracerProviderConfig.customSpanExporter != nil {
			opts := []sdktrace.TracerProviderOption{
				sdktrace.WithSampler(sdktrace.TraceIDRatioBased(c.tracerProviderConfig.samplingRate)),
			}
			if c.tracerProviderConfig.withSyncer {
				opts = append(opts, sdktrace.WithSyncer(c.tracerProviderConfig.customSpanExporter))
			} else {
				opts = append(opts, sdktrace.WithBatcher(c.tracerProviderConfig.customSpanExporter))
			}
			m.tp = sdktrace.NewTracerProvider(opts...)
		} else {
			tracerProviderOptions := []otlptracegrpc.Option{
				otlptracegrpc.WithEndpoint(c.tracesEndpoint),
				otlptracegrpc.WithRetry(otlptracegrpc.RetryConfig{
					Enabled:         c.retryConfig.Enabled,
					InitialInterval: c.retryConfig.InitialInterval,
					MaxInterval:     c.retryConfig.MaxInterval,
					MaxElapsedTime:  c.retryConfig.MaxElapsedTime,
				}),
			}
			if c.withInsecure {
				tracerProviderOptions = append(tracerProviderOptions, otlptracegrpc.WithInsecure())
			}
			traceExporter, err := otlptracegrpc.New(ctx, tracerProviderOptions...)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
			}

			m.tp = sdktrace.NewTracerProvider(
				sdktrace.WithResource(res),
				sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(traceExporter)),
				sdktrace.WithSampler(sdktrace.TraceIDRatioBased(c.tracerProviderConfig.samplingRate)),
			)
		}

		if c.tracerProviderConfig.textMapPropagator != nil {
			otel.SetTextMapPropagator(c.tracerProviderConfig.textMapPropagator)
		}

		if c.tracerProviderConfig.global {
			otel.SetTracerProvider(m.tp)
		}
	}

	if c.meterProviderConfig.enabled {
		var err error
		m.mp, err = m.buildMeterProvider(ctx, c, res)
		if err != nil {
			return nil, nil, err
		}
		if c.meterProviderConfig.global {
			otel.SetMeterProvider(m.mp)
		}
	}

	return m.tp, m.mp, nil
}

func (m *Manager) buildMeterProvider(
	ctx context.Context, c config, res *resource.Resource,
) (*sdkmetric.MeterProvider, error) {
	if c.meterProviderConfig.grpcEndpoint == nil && c.meterProviderConfig.prometheusRegisterer == nil {
		return nil, fmt.Errorf("no grpc endpoint or prometheus registerer to initialize meter provider")
	}
	if c.meterProviderConfig.grpcEndpoint != nil && c.meterProviderConfig.prometheusRegisterer != nil {
		return nil, fmt.Errorf("cannot initialize meter provider with both grpc endpoint and prometheus registerer")
	}
	if c.meterProviderConfig.prometheusRegisterer != nil {
		return m.buildPrometheusMeterProvider(c, res)
	}
	return m.buildOTLPMeterProvider(ctx, c, res)
}

func (m *Manager) buildPrometheusMeterProvider(c config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	exporterOptions := []prometheus.Option{
		prometheus.WithRegisterer(c.meterProviderConfig.prometheusRegisterer),
		prometheus.WithLogger(c.logger),
	}
	exp, err := prometheus.New(exporterOptions...)
	if err != nil {
		return nil, fmt.Errorf("prometheus: failed to create metric exporter: %w", err)
	}

	return sdkmetric.NewMeterProvider(m.getMeterProviderOptions(c, res, exp)...), nil
}

func (m *Manager) buildOTLPMeterProvider(
	ctx context.Context, c config, res *resource.Resource,
) (*sdkmetric.MeterProvider, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(*c.meterProviderConfig.grpcEndpoint),
		otlpmetricgrpc.WithRetry(otlpmetricgrpc.RetryConfig{
			Enabled:         c.retryConfig.Enabled,
			InitialInterval: c.retryConfig.InitialInterval,
			MaxInterval:     c.retryConfig.MaxInterval,
			MaxElapsedTime:  c.retryConfig.MaxElapsedTime,
		}),
	}
	if c.withInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(c.meterProviderConfig.otlpMetricGRPCOptions) > 0 {
		opts = append(opts, c.meterProviderConfig.otlpMetricGRPCOptions...)
	}
	exp, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otlp: failed to create metric exporter: %w", err)
	}

	reader := sdkmetric.NewPeriodicReader(
		exp,
		sdkmetric.WithInterval(c.meterProviderConfig.exportsInterval),
	)

	return sdkmetric.NewMeterProvider(m.getMeterProviderOptions(c, res, reader)...), nil
}

func (m *Manager) getMeterProviderOptions(c config, res *resource.Resource, r sdkmetric.Reader) []sdkmetric.Option {
	opts := []sdkmetric.Option{
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(r),
	}
	var views []sdkmetric.View
	if len(c.meterProviderConfig.views) > 0 {
		views = append(views, c.meterProviderConfig.views...)
	}
	if c.meterProviderConfig.defaultHistogramBuckets != nil {
		views = append(views, c.meterProviderConfig.defaultHistogramBuckets)
	}
	if len(views) > 0 {
		opts = append(opts, sdkmetric.WithView(views...))
	}
	return opts
}

// Shutdown allows you to gracefully clean up after the OTel manager (e.g. close underlying gRPC connection)
func (m *Manager) Shutdown(ctx context.Context) error {
	var g errgroup.Group
	if m.tp != nil {
		g.Go(func() error {
			return m.tp.Shutdown(ctx)
		})
	}
	if m.mp != nil {
		g.Go(func() error {
			return m.mp.Shutdown(ctx)
		})
	}

	done := make(chan error)
	go func() {
		done <- g.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// NewResource allows the creation of an OpenTelemetry resource
// https://opentelemetry.io/docs/concepts/glossary/#resource
func NewResource(svcName, svcVersion string, attrs ...attribute.KeyValue) (*resource.Resource, error) {
	defaultAttrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(svcName),
		semconv.ServiceVersionKey.String(svcVersion),
	}
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, append(defaultAttrs, attrs...)...),
	)
}

// RetryConfig defines configuration for retrying batches in case of export failure
// using an exponential backoff.
type RetryConfig struct {
	// Enabled indicates whether to not retry sending batches in case of
	// export failure.
	Enabled bool
	// InitialInterval the time to wait after the first failure before
	// retrying.
	InitialInterval time.Duration
	// MaxInterval is the upper bound on backoff interval. Once this value is
	// reached the delay between consecutive retries will always be
	// `MaxInterval`.
	MaxInterval time.Duration
	// MaxElapsedTime is the maximum amount of time (including retries) spent
	// trying to send a request/batch.  Once this value is reached, the data
	// is discarded.
	MaxElapsedTime time.Duration
}

type config struct {
	retryConfig  *RetryConfig
	withInsecure bool

	tracesEndpoint       string
	tracerProviderConfig tracerProviderConfig
	meterProviderConfig  meterProviderConfig

	logger logger
}

type tracerProviderConfig struct {
	enabled            bool
	global             bool
	samplingRate       float64
	textMapPropagator  propagation.TextMapPropagator
	customSpanExporter SpanExporter
	withSyncer         bool
}

type meterProviderConfig struct {
	enabled         bool
	global          bool
	exportsInterval time.Duration
	views           []sdkmetric.View
	// defaultHistogramBuckets is not part of the above "views" because the order
	// by which we add views matter. We have to add the default view last because the
	// views criteria are applied in order and the default one is the more generic.
	// Thus, if we put it first it will be applied to all histogram instruments removing
	// the ability to customize the buckets of specific histograms.
	defaultHistogramBuckets sdkmetric.View
	grpcEndpoint            *string
	prometheusRegisterer    promClient.Registerer
	otlpMetricGRPCOptions   []otlpmetricgrpc.Option
}

type logger interface {
	Info(...interface{})
	Error(...interface{})
}

type nopLogger struct{}

func (nopLogger) Info(...interface{})  {}
func (nopLogger) Error(...interface{}) {}
