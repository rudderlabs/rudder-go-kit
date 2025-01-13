package maxprocs

import (
	"math"
	"runtime"
	"strconv"
	"strings"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

func init() {
	setDefault()
}

func setDefault() {
	SetWithConfig(config.New(config.WithEnvPrefix("MAXPROCS")),
		WithLogger(logger.NewLogger().Child("maxprocs")),
	)
}

type conf struct {
	logger                logger.Logger
	minProcs              int
	cpuRequestsMultiplier float64
	roundQuotaFunc        func(float64) int
}

type Option func(*conf)

func WithLogger(logger logger.Logger) Option {
	return func(c *conf) { c.logger = logger }
}

func WithMinProcs(minProcs int) Option {
	return func(c *conf) { c.minProcs = minProcs }
}

func WithCPURequestsMultiplier(cpuRequestsMultiplier float64) Option {
	return func(c *conf) { c.cpuRequestsMultiplier = cpuRequestsMultiplier }
}

func WithRoundQuotaFunc(roundQuotaFunc func(float64) int) Option {
	return func(c *conf) { c.roundQuotaFunc = roundQuotaFunc }
}

func Set(raw string, opts ...Option) {
	conf := &conf{
		logger:                logger.NOP,
		minProcs:              1,
		cpuRequestsMultiplier: 3,
		roundQuotaFunc:        roundQuotaCeil,
	}
	for _, opt := range opts {
		opt(conf)
	}

	cpuRequests := 1.0
	if strings.HasSuffix(raw, "m") {
		value, err := strconv.Atoi(strings.TrimSuffix(raw, "m"))
		if err == nil {
			cpuRequests = float64(value) / 1000
		} else {
			conf.logger.Warnn("unable to parse CPU requests with Atoi, using default value")
		}
	} else {
		value, err := strconv.ParseFloat(raw, 64)
		if err == nil {
			cpuRequests = value
		} else {
			conf.logger.Warnn("unable to parse CPU requests with ParseFloat, using default value")
		}
	}

	// Calculate GOMAXPROCS
	gomaxprocs := conf.roundQuotaFunc(cpuRequests * conf.cpuRequestsMultiplier)
	if gomaxprocs < conf.minProcs {
		gomaxprocs = conf.minProcs
	}

	// Set GOMAXPROCS
	runtime.GOMAXPROCS(gomaxprocs)

	// Log new GOMAXPROCS
	conf.logger.Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", cpuRequests),
		logger.NewFloatField("multiplier", conf.cpuRequestsMultiplier),
		logger.NewIntField("minProcs", int64(conf.minProcs)),
		logger.NewIntField("result", int64(gomaxprocs)),
		logger.NewIntField("GOMAXPROCS", int64(runtime.GOMAXPROCS(0))),
	)
}

func SetWithConfig(c *config.Config, opts ...Option) {
	conf := &conf{
		logger:                logger.NOP,
		minProcs:              c.GetInt("MinProcs", 1),
		cpuRequestsMultiplier: c.GetFloat64("RequestsMultiplier", 3),
		roundQuotaFunc:        roundQuotaCeil,
	}
	for _, opt := range opts {
		opt(conf)
	}

	Set(c.GetString("Requests", "1"),
		WithLogger(conf.logger),
		WithMinProcs(conf.minProcs),
		WithCPURequestsMultiplier(conf.cpuRequestsMultiplier),
		WithRoundQuotaFunc(conf.roundQuotaFunc),
	)
}

func roundQuotaCeil(f float64) int {
	return int(math.Ceil(f))
}
