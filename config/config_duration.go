package config

import (
	"strconv"
	"time"
)

// GetDuration gets duration value from config
func GetDuration(key string, defaultValueInTimescaleUnits int64, timeScale time.Duration) (value time.Duration) {
	return Default.GetDuration(key, defaultValueInTimescaleUnits, timeScale)
}

// GetDuration gets duration value from config
func (c *Config) GetDuration(key string, defaultValueInTimescaleUnits int64, timeScale time.Duration) (value time.Duration) {
	return c.GetDurationVar(defaultValueInTimescaleUnits, timeScale, key)
}

// GetDurationVar registers a not hot-reloadable duration config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetDurationVar(
	defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string,
) time.Duration {
	return Default.GetDurationVar(defaultValueInTimescaleUnits, timeScale, orderedKeys...)
}

// GetDurationVar registers a not hot-reloadable duration config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetDurationVar(
	defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string,
) time.Duration {
	var ret time.Duration
	c.storeAndRegisterDurationVar(defaultValueInTimescaleUnits, &ret, timeScale, func(v time.Duration) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getDurationInternal gets duration value from config
func (c *Config) getDurationInternal(key string, defaultValueInTimescaleUnits int64, timeScale time.Duration) (value time.Duration) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return time.Duration(defaultValueInTimescaleUnits) * timeScale
	} else {
		v := c.v.GetString(key)
		parseDuration, err := time.ParseDuration(v)
		if err == nil {
			return parseDuration
		} else {
			_, err = strconv.ParseFloat(v, 64)
			if err == nil {
				return c.v.GetDuration(key) * timeScale
			} else {
				return time.Duration(defaultValueInTimescaleUnits) * timeScale
			}
		}
	}
}

// GetReloadableDurationVar registers a not hot-reloadable duration config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableDurationVar(defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string) *Reloadable[time.Duration] {
	return Default.GetReloadableDurationVar(defaultValueInTimescaleUnits, timeScale, orderedKeys...)
}

// GetReloadableDurationVar registers a hot-reloadable duration config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableDurationVar(
	defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string,
) *Reloadable[time.Duration] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, time.Duration(defaultValueInTimescaleUnits)*timeScale,
		&configValue{
			multiplier:   timeScale,
			defaultValue: defaultValueInTimescaleUnits,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.storeDurationVar(defaultValueInTimescaleUnits, timeScale, func(v time.Duration) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeDurationVar(
	defaultValueInTimescaleUnits int64, timeScale time.Duration, store func(time.Duration), orderedKeys ...string,
) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getDurationInternal(key, defaultValueInTimescaleUnits, timeScale))
			return
		}
	}
	store(time.Duration(defaultValueInTimescaleUnits) * timeScale)
}

func (c *Config) storeAndRegisterDurationVar(
	defaultValueInTimescaleUnits int64, ptr any, timeScale time.Duration,
	store func(time.Duration), orderedKeys ...string,
) {
	c.storeDurationVar(defaultValueInTimescaleUnits, timeScale, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, time.Duration(defaultValueInTimescaleUnits)*timeScale, &configValue{
		value:        ptr,
		multiplier:   timeScale,
		defaultValue: defaultValueInTimescaleUnits,
		keys:         orderedKeys,
	})
}
