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
	SetWithConfig(config.New(),
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

func Set(cpuRequests string, opts ...Option) {
	conf := &conf{
		logger:                logger.NOP,
		minProcs:              1,
		cpuRequestsMultiplier: 3,
		roundQuotaFunc:        roundQuotaCeil,
	}
	for _, opt := range opts {
		opt(conf)
	}

	cpuRequest := 1.0
	if strings.HasSuffix(cpuRequests, "m") {
		value, err := strconv.Atoi(strings.TrimSuffix(cpuRequests, "m"))
		if err == nil {
			cpuRequest = float64(value) / 1000
		} else {
			conf.logger.Warnn("unable to parse CPU requests with Atoi, using default value")
		}
	} else {
		value, err := strconv.ParseFloat(cpuRequests, 64)
		if err == nil {
			cpuRequest = value
		} else {
			conf.logger.Warnn("unable to parse CPU requests with ParseFloat, using default value")
		}
	}

	// Calculate GOMAXPROCS
	gomaxprocs := conf.roundQuotaFunc(cpuRequest * conf.cpuRequestsMultiplier)

	if gomaxprocs < conf.minProcs {
		gomaxprocs = conf.minProcs
	}

	// Set GOMAXPROCS
	runtime.GOMAXPROCS(gomaxprocs)
}

func SetWithConfig(c *config.Config, opts ...Option) {
	conf := &conf{
		logger:                logger.NOP,
		minProcs:              c.GetInt("MaxProcs.MinProcs", 1),
		cpuRequestsMultiplier: c.GetFloat64("MaxProcs.CPURequestsMultiplier", 3),
		roundQuotaFunc:        roundQuotaCeil,
	}
	for _, opt := range opts {
		opt(conf)
	}

	Set(c.GetString("MaxProcs.CPURequests", "1"),
		WithLogger(conf.logger),
		WithMinProcs(conf.minProcs),
		WithCPURequestsMultiplier(conf.cpuRequestsMultiplier),
		WithRoundQuotaFunc(conf.roundQuotaFunc),
	)
}

func roundQuotaCeil(f float64) int {
	return int(math.Ceil(f))
}
