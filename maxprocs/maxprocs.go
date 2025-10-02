package maxprocs

import (
	"math"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
)

const (
	defaultMinProcs              = 1
	defaultCPURequestsMultiplier = 1.5
)

func init() {
	setDefault()
}

func setDefault() {
	c := config.New(config.WithEnvPrefix("MAXPROCS"))
	l := logger.NewFactory(c).NewLogger().Child("maxprocs")
	SetWithConfig(c, WithLogger(l))
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
		minProcs:              defaultMinProcs,
		cpuRequestsMultiplier: defaultCPURequestsMultiplier,
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
		minProcs:              c.GetInt("MinProcs", defaultMinProcs),
		cpuRequestsMultiplier: c.GetFloat64("RequestsMultiplier", defaultCPURequestsMultiplier),
		roundQuotaFunc:        roundQuotaCeil,
	}
	for _, opt := range opts {
		opt(conf)
	}

	var (
		fileMode     bool
		requests     = c.GetString("Requests", "1")
		requestsFile = c.GetString("RequestsFile", "/etc/podinfo/cpu_requests")
	)
	if data, err := os.ReadFile(requestsFile); err == nil && len(data) > 0 {
		fileMode = true
		requests = strings.TrimSpace(string(data)) + "m"
		conf.logger.Infon("Using CPU requests from file",
			logger.NewStringField("requests", requests),
			logger.NewStringField("file", requestsFile),
		)
	}

	Set(requests,
		WithLogger(conf.logger),
		WithMinProcs(conf.minProcs),
		WithCPURequestsMultiplier(conf.cpuRequestsMultiplier),
		WithRoundQuotaFunc(conf.roundQuotaFunc),
	)

	if fileMode && c.GetBool("Watch", true) {
		conf.logger.Infon("Starting file watcher to monitor CPU requests changes",
			logger.NewStringField("file", requestsFile),
		)
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Kill, os.Interrupt, syscall.SIGTERM)
		go watchFile(requestsFile, conf, stop)
	}
}

func watchFile(file string, conf *conf, stop chan os.Signal) {
	log := conf.logger.Withn(logger.NewStringField("file", file))

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Warnn("Failed to create file watcher", obskit.Error(err))
		return
	}
	defer func() {
		err := watcher.Close()
		if err != nil {
			log.Warnn("Failed to close file watcher", obskit.Error(err))
		}
	}()

	if err := watcher.Add(file); err != nil {
		log.Warnn("Failed to watch file", obskit.Error(err))
		return
	}

	for {
		select {
		case <-stop:
			log.Infon("Received signal, stopping file watcher")
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if data, err := os.ReadFile(file); err == nil && len(data) > 0 {
					requests := strings.TrimSpace(string(data)) + "m"
					Set(requests,
						WithLogger(conf.logger),
						WithMinProcs(conf.minProcs),
						WithCPURequestsMultiplier(conf.cpuRequestsMultiplier),
						WithRoundQuotaFunc(conf.roundQuotaFunc),
					)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Warnn("File watcher error", obskit.Error(err))
		}
	}
}

func roundQuotaCeil(f float64) int {
	return int(math.Ceil(f))
}
