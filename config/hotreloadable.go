package config

import (
	"sync"
	"time"
)

// GetReloadableIntVar registers a hot-reloadable int config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableIntVar(defaultValue, valueScale int, orderedKeys ...string) *Reloadable[int] {
	return Default.GetReloadableIntVar(defaultValue, valueScale, orderedKeys...)
}

// GetReloadableIntVar registers a hot-reloadable int config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableIntVar(defaultValue, valueScale int, orderedKeys ...string) *Reloadable[int] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, defaultValue*valueScale,
		&configValue{
			multiplier:   valueScale,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.registerIntVar(defaultValue, ptr, true, valueScale, func(v int) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerIntVar(defaultValue int, ptr any, isHotReloadable bool, valueScale int, store func(int), orderedKeys ...string) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, defaultValue*valueScale, &configValue{
			value:        ptr,
			multiplier:   valueScale,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		})
	}
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getIntInternal(key, defaultValue) * valueScale)
			return
		}
	}
	store(defaultValue * valueScale)
}

// GetReloadableBoolVar registers a hot-reloadable bool config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableBoolVar(defaultValue bool, orderedKeys ...string) *Reloadable[bool] {
	return Default.GetReloadableBoolVar(defaultValue, orderedKeys...)
}

// GetReloadableBoolVar registers a hot-reloadable bool config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableBoolVar(defaultValue bool, orderedKeys ...string) *Reloadable[bool] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, defaultValue,
		&configValue{
			defaultValue: defaultValue,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.registerBoolVar(defaultValue, ptr, true, func(v bool) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerBoolVar(defaultValue bool, ptr any, isHotReloadable bool, store func(bool), orderedKeys ...string) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, defaultValue, &configValue{
			value:        ptr,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		})
	}
	for _, key := range orderedKeys {
		c.bindEnv(key)
		if c.IsSet(key) {
			store(c.getBoolInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// GetReloadableFloat64Var registers a hot-reloadable float64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableFloat64Var(defaultValue float64, orderedKeys ...string) *Reloadable[float64] {
	return Default.GetReloadableFloat64Var(defaultValue, orderedKeys...)
}

// GetReloadableFloat64Var registers a hot-reloadable float64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableFloat64Var(defaultValue float64, orderedKeys ...string) *Reloadable[float64] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, defaultValue,
		&configValue{
			multiplier:   1.0,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.registerFloat64Var(defaultValue, ptr, true, func(v float64) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerFloat64Var(defaultValue float64, ptr any, isHotReloadable bool, store func(float64), orderedKeys ...string) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, defaultValue, &configValue{
			value:        ptr,
			multiplier:   1.0,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		})
	}

	for _, key := range orderedKeys {
		c.bindEnv(key)
		if c.IsSet(key) {
			store(c.getFloat64Internal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// GetReloadableInt64Var registers a hot-reloadable int64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableInt64Var(defaultValue, valueScale int64, orderedKeys ...string) *Reloadable[int64] {
	return Default.GetReloadableInt64Var(defaultValue, valueScale, orderedKeys...)
}

// GetReloadableInt64Var registers a not hot-reloadable int64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableInt64Var(defaultValue, valueScale int64, orderedKeys ...string) *Reloadable[int64] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, defaultValue*valueScale,
		&configValue{
			multiplier:   valueScale,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.registerInt64Var(defaultValue, ptr, true, valueScale, func(v int64) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerInt64Var(defaultValue int64, ptr any, isHotReloadable bool, valueScale int64, store func(int64), orderedKeys ...string) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, defaultValue*valueScale, &configValue{
			value:        ptr,
			multiplier:   valueScale,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		})
	}
	for _, key := range orderedKeys {
		c.bindEnv(key)
		if c.IsSet(key) {
			store(c.getInt64Internal(key, defaultValue) * valueScale)
			return
		}
	}
	store(defaultValue * valueScale)
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
		c.registerDurationVar(defaultValueInTimescaleUnits, ptr, true, timeScale, func(v time.Duration) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerDurationVar(
	defaultValueInTimescaleUnits int64, ptr any, isHotReloadable bool, timeScale time.Duration,
	store func(time.Duration), orderedKeys ...string,
) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, time.Duration(defaultValueInTimescaleUnits)*timeScale, &configValue{
			value:        ptr,
			multiplier:   timeScale,
			defaultValue: defaultValueInTimescaleUnits,
			keys:         orderedKeys,
		})
	}
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getDurationInternal(key, defaultValueInTimescaleUnits, timeScale))
			return
		}
	}
	store(time.Duration(defaultValueInTimescaleUnits) * timeScale)
}

// GetReloadableStringVar registers a hot-reloadable string config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableStringVar(defaultValue string, orderedKeys ...string) *Reloadable[string] {
	return Default.GetReloadableStringVar(defaultValue, orderedKeys...)
}

// GetReloadableStringVar registers a hot-reloadable string config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableStringVar(defaultValue string, orderedKeys ...string) *Reloadable[string] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, defaultValue,
		&configValue{
			defaultValue: defaultValue,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.registerStringVar(defaultValue, ptr, true, func(v string) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerStringVar(defaultValue string, ptr any, isHotReloadable bool, store func(string), orderedKeys ...string) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, defaultValue, &configValue{
			value:        ptr,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		})
	}
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getStringInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// GetReloadableStringSliceVar registers a hot-reloadable string slice config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableStringSliceVar(defaultValue []string, orderedKeys ...string) *Reloadable[[]string] {
	return Default.GetReloadableStringSliceVar(defaultValue, orderedKeys...)
}

// GetReloadableStringSliceVar registers a hot-reloadable string slice config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableStringSliceVar(defaultValue []string, orderedKeys ...string) *Reloadable[[]string] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, defaultValue,
		&configValue{
			defaultValue: defaultValue,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.registerStringSliceVar(defaultValue, ptr, true, func(v []string) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerStringSliceVar(defaultValue []string, ptr any, isHotReloadable bool, store func([]string), orderedKeys ...string) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, defaultValue, &configValue{
			value:        ptr,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		})
	}
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getStringSliceInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// GetReloadableStringMapVar registers a hot-reloadable string map config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetReloadableStringMapVar(
	defaultValue map[string]any, orderedKeys ...string,
) *Reloadable[map[string]any] {
	return Default.GetReloadableStringMapVar(defaultValue, orderedKeys...)
}

// GetReloadableStringMapVar registers a hot-reloadable string map config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetReloadableStringMapVar(
	defaultValue map[string]any, orderedKeys ...string,
) *Reloadable[map[string]any] {
	ptr, exists := getOrCreatePointer(
		c.hotReloadableVars, c.hotReloadableVarsDefaults, c.hotReloadableConfig,
		&c.hotReloadableConfigLock, defaultValue,
		&configValue{
			defaultValue: defaultValue,
			keys:         orderedKeys,
		},
		orderedKeys...,
	)
	if !exists {
		c.registerStringMapVar(defaultValue, ptr, true, func(v map[string]any) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) registerStringMapVar(defaultValue map[string]any, ptr any, isHotReloadable bool, store func(map[string]any), orderedKeys ...string) {
	if !isHotReloadable {
		registerNonReloadableConfigKeys(c, defaultValue, &configValue{
			value:        ptr,
			defaultValue: defaultValue,
			keys:         orderedKeys,
		})
	}
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getStringMapInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// Reloadable is used as a wrapper for hot-reloadable config variables
type Reloadable[T configTypes] struct {
	value T
	lock  sync.RWMutex
}

// Load should be used to read the underlying value without worrying about data races
func (a *Reloadable[T]) Load() T {
	a.lock.RLock()
	v := a.value
	a.lock.RUnlock()
	return v
}

func (a *Reloadable[T]) store(v T) {
	a.lock.Lock()
	a.value = v
	a.lock.Unlock()
}

// swapIfNotEqual is used internally to swap the value of a hot-reloadable config variable
// if the new value is not equal to the old value
func (a *Reloadable[T]) swapIfNotEqual(new T, compare func(old, new T) bool) (old T, swapped bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if !compare(a.value, new) {
		old := a.value
		a.value = new
		return old, true
	}
	return a.value, false
}

func compare[T comparable]() func(a, b T) bool {
	return func(a, b T) bool {
		return a == b
	}
}
