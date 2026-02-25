package stats

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cast"
	ootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	noopMetric "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/stats/internal/otel"
)

const (
	defaultMeterName = ""
)

// otelStats is an OTel-specific adapter that follows the Stats contract
type otelStats struct {
	config        statsConfig
	otelConfig    otelStatsConfig
	resourceAttrs map[string]struct{}

	tracerProvider      trace.TracerProvider
	traceBaseAttributes []attribute.KeyValue
	tracerMap           map[string]Tracer
	tracerMapMu         sync.Mutex

	meter        metric.Meter
	noopMeter    metric.Meter
	counters     map[string]metric.Int64Counter
	countersMu   sync.Mutex
	gauges       map[string]*otelGauge
	gaugesMu     sync.Mutex
	timers       map[string]metric.Float64Histogram
	timersMu     sync.Mutex
	histograms   map[string]metric.Float64Histogram
	histogramsMu sync.Mutex

	otelManager              otel.Manager
	collectorAggregator      *aggregatedCollector
	runtimeStatsCollector    runtimeStatsCollector
	metricsStatsCollector    metricStatsCollector
	stopBackgroundCollection func()
	logger                   logger.Logger

	httpServer                 *http.Server
	httpServerShutdownComplete chan struct{}
	prometheusRegisterer       prometheus.Registerer
	prometheusGatherer         prometheus.Gatherer
}

func OtelVersion() string {
	return ootel.Version()
}

func (s *otelStats) Start(ctx context.Context, goFactory GoRoutineFactory) error {
	if !s.config.enabled.Load() {
		return nil
	}

	// Starting OpenTelemetry setup
	var attrs []attribute.KeyValue
	s.resourceAttrs = make(map[string]struct{})
	if s.config.instanceName != "" {
		sanitized := sanitizeTagKey("instanceName")
		attrs = append(attrs, attribute.String(sanitized, s.config.instanceName))
		s.resourceAttrs[sanitized] = struct{}{}
	}
	if s.config.namespaceIdentifier != "" {
		sanitized := sanitizeTagKey("namespace")
		attrs = append(attrs, attribute.String(sanitized, s.config.namespaceIdentifier))
		s.resourceAttrs[sanitized] = struct{}{}
	}
	res, err := otel.NewResource(s.config.serviceName, s.config.serviceVersion, attrs...)
	if err != nil {
		return fmt.Errorf("failed to create open telemetry resource: %w", err)
	}

	options := []otel.Option{otel.WithInsecure(), otel.WithLogger(s.logger)}
	if s.otelConfig.tracesEndpoint != "" {
		s.traceBaseAttributes = attrs
		tpOpts := []otel.TracerProviderOption{
			otel.WithTracingSamplingRate(s.otelConfig.tracingSamplingRate),
		}
		if s.otelConfig.withTracingSyncer {
			tpOpts = append(tpOpts, otel.WithTracingSyncer())
		}
		if s.otelConfig.withOTLPHTTP {
			tpOpts = append(tpOpts, otel.WithOTLPHTTP())
		}
		options = append(options,
			otel.WithTracerProvider(s.otelConfig.tracesEndpoint, tpOpts...),
			otel.WithTextMapPropagator(
				propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
			),
		)
	}

	meterProviderOptions := []otel.MeterProviderOption{
		otel.WithMeterProviderExportsInterval(s.otelConfig.metricsExportInterval),
	}

	// Configure default histogram aggregation (exponential takes precedence over explicit buckets)
	if s.config.useExponentialHistogram {
		meterProviderOptions = append(meterProviderOptions,
			otel.WithDefaultExponentialHistogram(s.config.exponentialHistogramMaxSize),
		)
	} else if len(s.config.defaultHistogramBuckets) > 0 {
		meterProviderOptions = append(meterProviderOptions,
			otel.WithDefaultHistogramBucketBoundaries(s.config.defaultHistogramBuckets),
		)
	}

	// Configure per-histogram aggregation (exponential histograms are configured first, then explicit buckets)
	if len(s.config.exponentialHistograms) > 0 {
		for histogramName, maxSize := range s.config.exponentialHistograms {
			meterProviderOptions = append(meterProviderOptions,
				otel.WithExponentialHistogram(histogramName, defaultMeterName, maxSize),
			)
		}
	}
	if len(s.config.histogramBuckets) > 0 {
		for histogramName, buckets := range s.config.histogramBuckets {
			// Only apply explicit bucket config if not already configured as exponential
			if _, isExponential := s.config.exponentialHistograms[histogramName]; !isExponential {
				meterProviderOptions = append(meterProviderOptions,
					otel.WithHistogramBucketBoundaries(histogramName, defaultMeterName, buckets),
				)
			}
		}
	}
	if s.otelConfig.metricsEndpoint != "" {
		options = append(options, otel.WithMeterProvider(append(meterProviderOptions,
			otel.WithGRPCMeterProvider(s.otelConfig.metricsEndpoint),
		)...))
	} else if s.otelConfig.enablePrometheusExporter {
		options = append(options, otel.WithMeterProvider(append(meterProviderOptions,
			otel.WithPrometheusExporter(s.prometheusRegisterer),
		)...))
	}

	tp, mp, err := s.otelManager.Setup(ctx, res, options...)
	if err != nil {
		return fmt.Errorf("failed to setup open telemetry: %w", err)
	}

	if tp != nil {
		s.tracerProvider = tp
	}

	s.noopMeter = noopMetric.NewMeterProvider().Meter(defaultMeterName)
	if mp != nil {
		s.meter = mp.Meter(defaultMeterName)
	} else {
		s.meter = s.noopMeter
	}

	if s.otelConfig.enablePrometheusExporter && s.otelConfig.prometheusMetricsPort > 0 {
		s.httpServerShutdownComplete = make(chan struct{})
		s.httpServer = &http.Server{
			Addr: fmt.Sprintf(":%d", s.otelConfig.prometheusMetricsPort),
			Handler: promhttp.InstrumentMetricHandler(
				s.prometheusRegisterer, promhttp.HandlerFor(s.prometheusGatherer, promhttp.HandlerOpts{
					ErrorLog: &prometheusLogger{l: s.logger},
				}),
			),
		}
		goFactory.Go(func() {
			defer close(s.httpServerShutdownComplete)
			if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				s.logger.Fataln("Prometheus exporter failed", obskit.Error(err))
			}
		})
	}

	// Starting background collection
	var backgroundCollectionCtx context.Context
	backgroundCollectionCtx, s.stopBackgroundCollection = context.WithCancel(context.Background())

	gaugeFunc := func(key string, val uint64) {
		s.getMeasurement("runtime_"+key, GaugeType, nil).Gauge(val)
	}
	s.metricsStatsCollector = newMetricStatsCollector(s, s.config.periodicStatsConfig.metricManager)
	goFactory.Go(func() {
		s.metricsStatsCollector.run(backgroundCollectionCtx)
	})

	gaugeTagsFunc := func(key string, tags Tags, val uint64) {
		s.getMeasurement(key, GaugeType, tags).Gauge(val)
	}
	s.collectorAggregator.gaugeFunc = gaugeTagsFunc
	goFactory.Go(func() {
		s.collectorAggregator.Run(backgroundCollectionCtx)
	})

	if s.config.periodicStatsConfig.enabled {
		s.runtimeStatsCollector = newRuntimeStatsCollector(gaugeFunc)
		s.runtimeStatsCollector.PauseDur = time.Duration(s.config.periodicStatsConfig.statsCollectionInterval) * time.Second
		s.runtimeStatsCollector.EnableCPU = s.config.periodicStatsConfig.enableCPUStats
		s.runtimeStatsCollector.EnableMem = s.config.periodicStatsConfig.enableMemStats
		s.runtimeStatsCollector.EnableGC = s.config.periodicStatsConfig.enableGCStats
		goFactory.Go(func() {
			s.runtimeStatsCollector.run(backgroundCollectionCtx)
		})
	}

	if s.otelConfig.enablePrometheusExporter {
		s.logger.Infon("Stats started in Prometheus mode", logger.NewIntField("port", int64(s.otelConfig.prometheusMetricsPort)))
	} else {
		s.logger.Infon("Stats started in OpenTelemetry mode",
			logger.NewStringField("metricsEndpoint", s.otelConfig.metricsEndpoint),
			logger.NewStringField("tracesEndpoint", s.otelConfig.tracesEndpoint),
		)
	}

	return nil
}

func (s *otelStats) RegisterCollector(c Collector) error {
	return s.collectorAggregator.Add(c)
}

func (s *otelStats) Stop() {
	if !s.config.enabled.Load() {
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	if err := s.otelManager.Shutdown(ctx); err != nil {
		s.logger.Errorn("failed to shutdown open telemetry", obskit.Error(err))
	}

	s.stopBackgroundCollection()
	if s.metricsStatsCollector.done != nil {
		<-s.metricsStatsCollector.done
	}
	if s.config.periodicStatsConfig.enabled && s.runtimeStatsCollector.done != nil {
		<-s.runtimeStatsCollector.done
	}

	if s.httpServer != nil && s.httpServerShutdownComplete != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Errorn("failed to shutdown prometheus exporter", obskit.Error(err))
		}
		<-s.httpServerShutdownComplete
	}
}

// NewTracer allows you to create a tracer for creating spans
func (s *otelStats) NewTracer(name string) Tracer {
	s.tracerMapMu.Lock()
	defer s.tracerMapMu.Unlock()

	if s.tracerMap == nil {
		s.tracerMap = make(map[string]Tracer)
	} else if t, ok := s.tracerMap[name]; ok {
		return t
	}

	var attrs []attribute.KeyValue
	if len(s.traceBaseAttributes) > 0 {
		attrs = append(attrs, s.traceBaseAttributes...)
	}
	if s.config.serviceName != "" {
		attrs = append(attrs, semconv.ServiceNameKey.String(s.config.serviceName))
	}

	opts := []trace.TracerOption{
		trace.WithInstrumentationVersion(s.config.serviceVersion),
	}
	if len(attrs) > 0 {
		opts = append(opts, trace.WithInstrumentationAttributes(attrs...))
	}

	s.tracerMap[name] = &tracer{
		tracer: s.tracerProvider.Tracer(name, opts...),
	}
	return s.tracerMap[name]
}

// NewStat creates a new Measurement with provided Name and Type
func (s *otelStats) NewStat(name, statType string) (m Measurement) {
	return s.getMeasurement(name, statType, nil)
}

// NewTaggedStat creates a new Measurement with provided Name, Type and Tags
func (s *otelStats) NewTaggedStat(name, statType string, tags Tags) (m Measurement) {
	return s.getMeasurement(name, statType, tags)
}

// NewSampledTaggedStat creates a new Measurement with provided Name, Type and Tags
// Deprecated: use NewTaggedStat instead
func (s *otelStats) NewSampledTaggedStat(name, statType string, tags Tags) (m Measurement) {
	return s.NewTaggedStat(name, statType, tags)
}

func (*otelStats) getNoOpMeasurement(statType string) Measurement {
	om := &otelMeasurement{
		genericMeasurement: genericMeasurement{statType: statType},
		disabled:           true,
	}
	switch statType {
	case CountType:
		return &otelCounter{otelMeasurement: om}
	case GaugeType:
		return &otelGauge{otelMeasurement: om}
	case TimerType:
		return &otelTimer{otelMeasurement: om}
	case HistogramType:
		return &otelHistogram{otelMeasurement: om}
	}
	panic(fmt.Errorf("unsupported measurement type %s", statType))
}

func (s *otelStats) getMeasurement(name, statType string, tags Tags) Measurement {
	if !s.config.enabled.Load() {
		return s.getNoOpMeasurement(statType)
	}

	if strings.Trim(name, " ") == "" {
		byteArr := make([]byte, 2048)
		n := runtime.Stack(byteArr, false)
		stackTrace := string(byteArr[:n])
		s.logger.Warnn("detected missing stat measurement name, using 'novalue'",
			logger.NewStringField("stacktrace", stackTrace),
		)
		name = "novalue"
	}

	// Clean up tags based on deployment type. No need to send workspace id tag for free tier customers.
	newTags := make(Tags)
	for k, v := range tags {
		if strings.Trim(k, " ") == "" {
			s.logger.Warnn("removing empty tag key for measurement",
				logger.NewStringField("value", v),
				logger.NewStringField("measurement", name),
			)
			continue
		}
		if _, ok := s.config.excludedTags[k]; ok {
			continue
		}
		sanitizedKey := sanitizeTagKey(k)
		if _, ok := s.config.excludedTags[sanitizedKey]; ok {
			continue
		}
		if _, ok := s.resourceAttrs[sanitizedKey]; ok {
			s.logger.Warnn("removing tag for measurement since it is a resource attribute",
				logger.NewStringField("key", k),
				logger.NewStringField("measurement", name),
			)
			continue
		}
		newTags[sanitizedKey] = v
	}

	om := &otelMeasurement{
		genericMeasurement: genericMeasurement{statType: statType},
		attributes:         newTags.otelAttributes(),
	}

	switch statType {
	case CountType:
		instr := buildOTelInstrument(s.meter, s.noopMeter, name, s.counters, &s.countersMu, s.logger)
		return &otelCounter{counter: instr, otelMeasurement: om}
	case GaugeType:
		return s.getGauge(name, om.attributes, newTags.String())
	case TimerType:
		instr := buildOTelInstrument(s.meter, s.noopMeter, name, s.timers, &s.timersMu, s.logger)
		return &otelTimer{timer: instr, otelMeasurement: om}
	case HistogramType:
		instr := buildOTelInstrument(s.meter, s.noopMeter, name, s.histograms, &s.histogramsMu, s.logger)
		return &otelHistogram{histogram: instr, otelMeasurement: om}
	default:
		panic(fmt.Errorf("unsupported measurement type %s", statType))
	}
}

func (s *otelStats) getGauge(
	name string, attributes []attribute.KeyValue, tagsKey string,
) *otelGauge {
	var (
		ok     bool
		og     *otelGauge
		mapKey = name + "|" + tagsKey
	)

	s.gaugesMu.Lock()
	defer s.gaugesMu.Unlock()

	if s.gauges == nil {
		s.gauges = make(map[string]*otelGauge)
	} else {
		og, ok = s.gauges[mapKey]
	}

	if !ok {
		og = &otelGauge{otelMeasurement: &otelMeasurement{
			genericMeasurement: genericMeasurement{statType: GaugeType},
			attributes:         attributes,
		}}

		g, err := s.meter.Float64ObservableGauge(name)
		if err != nil {
			s.logger.Warnn("failed to create gauge",
				logger.NewStringField("name", name),
				obskit.Error(err),
			)
			g, _ = s.noopMeter.Float64ObservableGauge(name)
		} else {
			_, err = s.meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
				if value := og.getValue(); value != nil {
					o.ObserveFloat64(g, cast.ToFloat64(value), metric.WithAttributes(attributes...))
				}
				return nil
			}, g)
			if err != nil {
				panic(fmt.Errorf("failed to register callback for gauge %s: %w", name, err))
			}
		}

		s.gauges[mapKey] = og
	}

	return og
}

func buildOTelInstrument[T any](
	meter, noopMeter metric.Meter,
	name string, m map[string]T, mu *sync.Mutex,
	l logger.Logger,
) T {
	var (
		ok    bool
		instr T
	)

	mu.Lock()
	defer mu.Unlock()
	if m == nil {
		m = make(map[string]T)
	} else {
		instr, ok = m[name]
	}

	if !ok {
		var err error
		var value any
		switch any(m).(type) {
		case map[string]metric.Int64Counter:
			if value, err = meter.Int64Counter(name); err != nil {
				value, _ = noopMeter.Int64Counter(name)
			}
		case map[string]metric.Float64Histogram:
			if value, err = meter.Float64Histogram(name); err != nil {
				value, _ = noopMeter.Float64Histogram(name)
			}
		default:
			panic(fmt.Errorf("unknown instrument type %T", instr))
		}
		if err != nil {
			l.Warnn("failed to create instrument",
				logger.NewStringField("type", fmt.Sprintf("%T", instr)),
				logger.NewStringField("name", name),
				obskit.Error(err),
			)
		}
		instr = value.(T)
		m[name] = instr
	}

	return instr
}

type otelStatsConfig struct {
	tracesEndpoint           string
	tracingSamplingRate      float64
	withTracingSyncer        bool
	withOTLPHTTP             bool
	metricsEndpoint          string
	metricsExportInterval    time.Duration
	enablePrometheusExporter bool
	prometheusMetricsPort    int
}

type prometheusLogger struct{ l logger.Logger }

func (p *prometheusLogger) Println(v ...any) {
	p.l.Error(v...) //nolint:forbidigo
}
