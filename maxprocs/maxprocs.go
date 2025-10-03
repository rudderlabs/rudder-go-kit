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
	setWithConfig(c, withLogger(l))
}

func defaultMaxProcs() int {
	return runtime.NumCPU()
}

type conf struct {
	logger                logger.Logger
	minProcs              int
	maxProcs              int
	cpuRequestsMultiplier float64
	roundQuotaFunc        func(float64) int
	stop                  chan os.Signal
}

type option func(*conf)

func withLogger(logger logger.Logger) option {
	return func(c *conf) { c.logger = logger }
}

func withMinProcs(minProcs int) option {
	return func(c *conf) { c.minProcs = minProcs }
}

func withMaxProcs(maxProcs int) option {
	return func(c *conf) { c.maxProcs = maxProcs }
}

func withCPURequestsMultiplier(cpuRequestsMultiplier float64) option {
	return func(c *conf) { c.cpuRequestsMultiplier = cpuRequestsMultiplier }
}

func withRoundQuotaFunc(roundQuotaFunc func(float64) int) option {
	return func(c *conf) { c.roundQuotaFunc = roundQuotaFunc }
}

func withStopFileWatcher(stop chan os.Signal) option {
	return func(c *conf) { c.stop = stop }
}

func set(raw string, opts ...option) {
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
	if conf.maxProcs > 0 && gomaxprocs > conf.maxProcs {
		gomaxprocs = conf.maxProcs
	}

	// Set GOMAXPROCS
	runtime.GOMAXPROCS(gomaxprocs)

	// Log new GOMAXPROCS
	conf.logger.Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", cpuRequests),
		logger.NewFloatField("multiplier", conf.cpuRequestsMultiplier),
		logger.NewIntField("minProcs", int64(conf.minProcs)),
		logger.NewIntField("maxProcs", int64(conf.maxProcs)),
		logger.NewIntField("result", int64(gomaxprocs)),
		logger.NewIntField("GOMAXPROCS", int64(runtime.GOMAXPROCS(0))),
	)
}

func setWithConfig(c *config.Config, opts ...option) {
	conf := &conf{
		logger:                logger.NOP,
		minProcs:              c.GetInt("MinProcs", defaultMinProcs),
		maxProcs:              c.GetInt("MaxProcs", defaultMaxProcs()),
		cpuRequestsMultiplier: c.GetFloat64("RequestsMultiplier", defaultCPURequestsMultiplier),
		roundQuotaFunc:        roundQuotaCeil,
		stop:                  make(chan os.Signal, 1),
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

	set(requests,
		withLogger(conf.logger),
		withMinProcs(conf.minProcs),
		withMaxProcs(conf.maxProcs),
		withCPURequestsMultiplier(conf.cpuRequestsMultiplier),
		withRoundQuotaFunc(conf.roundQuotaFunc),
	)

	if fileMode && c.GetBool("Watch", true) {
		conf.logger.Infon("Starting file watcher to monitor CPU requests changes",
			logger.NewStringField("file", requestsFile),
		)
		signal.Notify(conf.stop, os.Interrupt, syscall.SIGTERM)
		go watchFile(conf, requestsFile)
	}
}

func watchFile(conf *conf, file string) {
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

	log.Debugn("Watching file for changes")
	defer log.Debugn("Stopped watching file for changes")

	for {
		select {
		case <-conf.stop:
			log.Infon("Received signal, stopping file watcher")
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if data, err := os.ReadFile(file); err == nil && len(data) > 0 {
					requests := strings.TrimSpace(string(data)) + "m"
					set(requests,
						withLogger(conf.logger),
						withMinProcs(conf.minProcs),
						withMaxProcs(conf.maxProcs),
						withCPURequestsMultiplier(conf.cpuRequestsMultiplier),
						withRoundQuotaFunc(conf.roundQuotaFunc),
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
