package mem

import (
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

// SetConfig holds configuration for setting memory limits
type SetConfig struct {
	limitPercent config.ValueLoader[int]
	log          logger.Logger
}

// SetOption configures the SetMemoryLimit behavior
type SetOption func(*SetConfig)

// WatchConfig holds configuration for watching memory limits
type WatchConfig struct {
	limitPercent config.ValueLoader[int]
	interval     time.Duration
	log          logger.Logger
}

// WatchOption configures the WatchMemoryLimit behavior
type WatchOption func(*WatchConfig)

// WatchWithInterval sets the polling interval for checking memory limit changes
func WatchWithInterval(interval time.Duration) WatchOption {
	return func(c *WatchConfig) {
		c.interval = interval
	}
}

// WatchWithPercentage sets the memory limit as a percentage of total system memory
func WatchWithPercentage(percent int) WatchOption {
	return func(c *WatchConfig) {
		c.limitPercent = config.SingleValueLoader(percent)
	}
}

// WatchWithPercentageLoader sets the memory limit using a config.ValueLoader
func WatchWithPercentageLoader(loader config.ValueLoader[int]) WatchOption {
	return func(c *WatchConfig) {
		c.limitPercent = loader
	}
}

// WatchWithLogger sets the logger to use
func WatchWithLogger(log logger.Logger) WatchOption {
	return func(c *WatchConfig) {
		c.log = log
	}
}

// SetWithPercentage sets the memory limit as a percentage of total system memory
func SetWithPercentage(percent int) SetOption {
	return func(c *SetConfig) {
		c.limitPercent = config.SingleValueLoader(percent)
	}
}

// SetWithPercentageLoader sets the memory limit using a config.ValueLoader
func SetWithPercentageLoader(loader config.ValueLoader[int]) SetOption {
	return func(c *SetConfig) {
		c.limitPercent = loader
	}
}

// SetWithLogger sets the logger to use
func SetWithLogger(log logger.Logger) SetOption {
	return func(c *SetConfig) {
		c.log = log
	}
}
