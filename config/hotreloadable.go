package config

import (
	"sync"
	"time"
)

// RegisterIntConfigVariable registers int config variable
// Deprecated: use RegisterIntVar or RegisterAtomicIntVar instead
func RegisterIntConfigVariable(defaultValue int, ptr *int, isHotReloadable bool, valueScale int, keys ...string) {
	Default.registerIntVar(defaultValue, ptr, isHotReloadable, valueScale, func(v int) {
		*ptr = v
	}, keys...)
}

// RegisterIntVar registers a not hot-reloadable int config variable
func RegisterIntVar(defaultValue int, ptr *int, valueScale int, keys ...string) {
	Default.RegisterIntVar(defaultValue, ptr, valueScale, keys...)
}

// RegisterAtomicIntVar registers a hot-reloadable int config variable
func RegisterAtomicIntVar(defaultValue int, ptr *Atomic[int], valueScale int, keys ...string) {
	Default.RegisterAtomicIntVar(defaultValue, ptr, valueScale, keys...)
}

// RegisterIntConfigVariable registers int config variable
// Deprecated: use RegisterIntVar or RegisterAtomicIntVar instead
func (c *Config) RegisterIntConfigVariable(defaultValue int, ptr *int, isHotReloadable bool, valueScale int, keys ...string) {
	c.registerIntVar(defaultValue, ptr, isHotReloadable, valueScale, func(v int) {
		*ptr = v
	}, keys...)
}

// RegisterIntVar registers a not hot-reloadable int config variable
func (c *Config) RegisterIntVar(defaultValue int, ptr *int, valueScale int, keys ...string) {
	c.registerIntVar(defaultValue, ptr, false, valueScale, func(v int) {
		*ptr = v
	}, keys...)
}

// RegisterAtomicIntVar registers a hot-reloadable int config variable
// Copy of RegisterIntConfigVariable, but with a way to avoid data races for hot reloadable config variables
func (c *Config) RegisterAtomicIntVar(defaultValue int, ptr *Atomic[int], valueScale int, keys ...string) {
	c.registerIntVar(defaultValue, ptr, true, valueScale, func(v int) {
		ptr.Store(v)
	}, keys...)
}

func (c *Config) registerIntVar(defaultValue int, ptr any, isHotReloadable bool, valueScale int, store func(int), keys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   valueScale,
		defaultValue: defaultValue,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		if c.IsSet(key) {
			store(c.GetInt(key, defaultValue) * valueScale)
			return
		}
	}
	store(defaultValue * valueScale)
}

// RegisterBoolConfigVariable registers bool config variable
func RegisterBoolConfigVariable(defaultValue bool, ptr *bool, isHotReloadable bool, keys ...string) {
	Default.RegisterBoolConfigVariable(defaultValue, ptr, isHotReloadable, keys...)
}

// RegisterBoolConfigVariable registers bool config variable
func (c *Config) RegisterBoolConfigVariable(defaultValue bool, ptr *bool, isHotReloadable bool, keys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		c.bindEnv(key)
		if c.IsSet(key) {
			*ptr = c.GetBool(key, defaultValue)
			return
		}
	}
	*ptr = defaultValue
}

// RegisterFloat64ConfigVariable registers float64 config variable
func RegisterFloat64ConfigVariable(defaultValue float64, ptr *float64, isHotReloadable bool, keys ...string) {
	Default.RegisterFloat64ConfigVariable(defaultValue, ptr, isHotReloadable, keys...)
}

// RegisterFloat64ConfigVariable registers float64 config variable
func (c *Config) RegisterFloat64ConfigVariable(defaultValue float64, ptr *float64, isHotReloadable bool, keys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   1.0,
		defaultValue: defaultValue,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		c.bindEnv(key)
		if c.IsSet(key) {
			*ptr = c.GetFloat64(key, defaultValue)
			return
		}
	}
	*ptr = defaultValue
}

// RegisterInt64ConfigVariable registers int64 config variable
func RegisterInt64ConfigVariable(defaultValue int64, ptr *int64, isHotReloadable bool, valueScale int64, keys ...string) {
	Default.RegisterInt64ConfigVariable(defaultValue, ptr, isHotReloadable, valueScale, keys...)
}

// RegisterInt64ConfigVariable registers int64 config variable
func (c *Config) RegisterInt64ConfigVariable(defaultValue int64, ptr *int64, isHotReloadable bool, valueScale int64, keys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   valueScale,
		defaultValue: defaultValue,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		c.bindEnv(key)
		if c.IsSet(key) {
			*ptr = c.GetInt64(key, defaultValue) * valueScale
			return
		}
	}
	*ptr = defaultValue * valueScale
}

// RegisterDurationConfigVariable registers duration config variable
// Deprecated: use RegisterDurationVar or RegisterAtomicDurationVar instead
func RegisterDurationConfigVariable(defaultValueInTimescaleUnits int64, ptr *time.Duration, isHotReloadable bool, timeScale time.Duration, keys ...string) {
	Default.registerDurationVar(defaultValueInTimescaleUnits, ptr, isHotReloadable, timeScale, func(v time.Duration) {
		*ptr = v
	}, keys...)
}

// RegisterDurationVar registers a not hot-reloadable duration config variable
func RegisterDurationVar(defaultValueInTimescaleUnits int64, ptr *time.Duration, timeScale time.Duration, keys ...string) {
	Default.RegisterDurationVar(defaultValueInTimescaleUnits, ptr, timeScale, keys...)
}

// RegisterAtomicDurationVar registers a not hot-reloadable duration config variable
func RegisterAtomicDurationVar(defaultValueInTimescaleUnits int64, ptr *Atomic[time.Duration], timeScale time.Duration, keys ...string) {
	Default.RegisterAtomicDurationVar(defaultValueInTimescaleUnits, ptr, timeScale, keys...)
}

// RegisterDurationConfigVariable registers duration config variable
// Deprecated: use RegisterDurationVar or RegisterAtomicDurationVar instead
func (c *Config) RegisterDurationConfigVariable(
	defaultValueInTimescaleUnits int64, ptr *time.Duration, isHotReloadable bool, timeScale time.Duration, keys ...string,
) {
	c.registerDurationVar(defaultValueInTimescaleUnits, ptr, isHotReloadable, timeScale, func(v time.Duration) {
		*ptr = v
	}, keys...)
}

// RegisterDurationVar registers a not hot-reloadable duration config variable
func (c *Config) RegisterDurationVar(
	defaultValueInTimescaleUnits int64, ptr *time.Duration, timeScale time.Duration, keys ...string,
) {
	c.registerDurationVar(defaultValueInTimescaleUnits, ptr, false, timeScale, func(v time.Duration) {
		*ptr = v
	}, keys...)
}

// RegisterAtomicDurationVar registers a hot-reloadable duration config variable
func (c *Config) RegisterAtomicDurationVar(
	defaultValueInTimescaleUnits int64, ptr *Atomic[time.Duration], timeScale time.Duration, keys ...string,
) {
	c.registerDurationVar(defaultValueInTimescaleUnits, ptr, true, timeScale, func(v time.Duration) {
		ptr.Store(v)
	}, keys...)
}

func (c *Config) registerDurationVar(
	defaultValueInTimescaleUnits int64, ptr any, isHotReloadable bool, timeScale time.Duration,
	store func(time.Duration), keys ...string,
) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   timeScale,
		defaultValue: defaultValueInTimescaleUnits,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		if c.IsSet(key) {
			store(c.GetDuration(key, defaultValueInTimescaleUnits, timeScale))
			return
		}
	}
	store(time.Duration(defaultValueInTimescaleUnits) * timeScale)
}

// RegisterStringConfigVariable registers string config variable
func RegisterStringConfigVariable(defaultValue string, ptr *string, isHotReloadable bool, keys ...string) {
	Default.RegisterStringConfigVariable(defaultValue, ptr, isHotReloadable, keys...)
}

// RegisterStringConfigVariable registers string config variable
func (c *Config) RegisterStringConfigVariable(defaultValue string, ptr *string, isHotReloadable bool, keys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		if c.IsSet(key) {
			*ptr = c.GetString(key, defaultValue)
			return
		}
	}
	*ptr = defaultValue
}

// RegisterStringSliceConfigVariable registers string slice config variable
func RegisterStringSliceConfigVariable(defaultValue []string, ptr *[]string, isHotReloadable bool, keys ...string) {
	Default.RegisterStringSliceConfigVariable(defaultValue, ptr, isHotReloadable, keys...)
}

// RegisterStringSliceConfigVariable registers string slice config variable
func (c *Config) RegisterStringSliceConfigVariable(defaultValue []string, ptr *[]string, isHotReloadable bool, keys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		if c.IsSet(key) {
			*ptr = c.GetStringSlice(key, defaultValue)
			return
		}
	}
	*ptr = defaultValue
}

// RegisterStringMapConfigVariable registers string map config variable
func RegisterStringMapConfigVariable(defaultValue map[string]interface{}, ptr *map[string]interface{}, isHotReloadable bool, keys ...string) {
	Default.RegisterStringMapConfigVariable(defaultValue, ptr, isHotReloadable, keys...)
}

// RegisterStringMapConfigVariable registers string map config variable
func (c *Config) RegisterStringMapConfigVariable(defaultValue map[string]interface{}, ptr *map[string]interface{}, isHotReloadable bool, keys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         keys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(keys[0], &configVar)
	}

	for _, key := range keys {
		if c.IsSet(key) {
			*ptr = c.GetStringMap(key, defaultValue)
			return
		}
	}
	*ptr = defaultValue
}

func (c *Config) appendVarToConfigMaps(key string, configVar *configValue) {
	if _, ok := c.hotReloadableConfig[key]; !ok {
		c.hotReloadableConfig[key] = make([]*configValue, 0)
	}
	c.hotReloadableConfig[key] = append(c.hotReloadableConfig[key], configVar)
}

// Atomic is used as a wrapper for hot-reloadable config variables
type Atomic[T comparable] struct {
	value T
	lock  sync.RWMutex
}

// Load should be used to read the underlying value without worrying about data races
func (a *Atomic[T]) Load() T {
	a.lock.RLock()
	v := a.value
	a.lock.RUnlock()
	return v
}

// Store should be used to write a value without worrying about data races
func (a *Atomic[T]) Store(v T) {
	a.lock.Lock()
	a.value = v
	a.lock.Unlock()
}

// swapIfNotEqual is used internally to swap the value of a hot-reloadable config variable
func (a *Atomic[T]) swapIfNotEqual(new T) (old T, swapped bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if a.value != new {
		old := a.value
		a.value = new
		return old, true
	}
	return a.value, false
}
