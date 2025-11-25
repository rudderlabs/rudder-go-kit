package mem

import (
	"context"
	"runtime/debug"
	"sync"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"
)

// SetMemoryLimit sets the GOMEMLIMIT based on the configured percentage of total system memory.
// By default, it uses 90% of total memory and uses a NOP logger.
// Use SetWithPercentage/SetWithPercentageLoader and SetWithLogger options to customize behavior.
func SetMemoryLimit(opts ...SetOption) {
	cfg := &SetConfig{
		limitPercent: config.SingleValueLoader(90),
		log:          logger.NOP,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	memoryLimit, err := calculateMemoryLimit(cfg.limitPercent)
	if err != nil {
		cfg.log.Errorn("Error calculating GOMEMLIMIT", obskit.Error(err))
		return
	}
	debug.SetMemoryLimit(memoryLimit)
	cfg.log.Infon("Set GOMEMLIMIT", logger.NewIntField("limitBytes", memoryLimit))
}

// WatchMemoryLimit continuously monitors the memory limit and updates GOMEMLIMIT whenever there is a change.
// By default, it uses 90% of total memory, checks every 10 seconds, and uses a NOP logger.
// Use WithPercentage/WithPercentageLoader, WithInterval, and WithLogger options to customize behavior.
func WatchMemoryLimit(ctx context.Context, wg *sync.WaitGroup, opts ...WatchOption) {
	cfg := &WatchConfig{
		limitPercent: config.SingleValueLoader(90),
		interval:     10 * time.Second,
		log:          logger.NOP,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	memoryLimits := distinctMemoryLimitValues(ctx, wg, cfg)
	wg.Go(func() {
		for memoryLimit := range memoryLimits {
			debug.SetMemoryLimit(memoryLimit)
			cfg.log.Infon("Set GOMEMLIMIT", logger.NewIntField("limitBytes", memoryLimit))
		}
	})
}

// calculateMemoryLimit returns the memory limit in bytes based on the given percentage of total system memory.
func calculateMemoryLimit(limitPercent config.ValueLoader[int]) (int64, error) {
	memStat, err := Get()
	if err != nil {
		return 0, err
	}
	memoryLimit := int64(uint64(limitPercent.Load()) * memStat.Total / 100)
	return memoryLimit, nil
}

// distinctMemoryLimitValues returns a channel that emits distinct memory limit values
// calculated based on the provided limitPercent, checking at the specified interval.
func distinctMemoryLimitValues(ctx context.Context, wg *sync.WaitGroup, cfg *WatchConfig) <-chan int64 {
	lastMemoryLimit := int64(-1)
	ch := make(chan int64, 1)

	poll := func() (ctxDone bool) {
		memoryLimit, err := calculateMemoryLimit(cfg.limitPercent)
		if err != nil {
			cfg.log.Errorn("Error calculating GOMEMLIMIT", obskit.Error(err))
			return false
		}
		if memoryLimit != lastMemoryLimit {
			lastMemoryLimit = memoryLimit
			select {
			case ch <- memoryLimit:
				return false
			case <-ctx.Done():
				return true
			}
		}
		return false
	}

	wg.Go(func() {
		defer close(ch)
		if poll() {
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(cfg.interval):
				if poll() {
					return
				}
			}
		}
	})
	return ch
}
