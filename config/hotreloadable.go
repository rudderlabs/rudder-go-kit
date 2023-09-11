package config

import (
	"strings"
	"sync"
	"time"
)

// RegisterIntConfigVariable registers int config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterIntVar or RegisterReloadableIntVar instead
func RegisterIntConfigVariable(defaultValue int, ptr *int, isHotReloadable bool, valueScale int, orderedKeys ...string) {
	Default.RegisterIntConfigVariable(defaultValue, ptr, isHotReloadable, valueScale, orderedKeys...)
}

// RegisterIntVar registers a not hot-reloadable int config variable
// WARNING: a different order of keys will result in a different variable
func RegisterIntVar(defaultValue, valueScale int, orderedKeys ...string) *int {
	return Default.RegisterIntVar(defaultValue, valueScale, orderedKeys...)
}

// RegisterReloadableIntVar registers a hot-reloadable int config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableIntVar(defaultValue, valueScale int, orderedKeys ...string) *Reloadable[int] {
	return Default.RegisterReloadableIntVar(defaultValue, valueScale, orderedKeys...)
}

// RegisterIntConfigVariable registers int config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterIntVar or RegisterReloadableIntVar instead
func (c *Config) RegisterIntConfigVariable(
	defaultValue int, ptr *int, isHotReloadable bool, valueScale int, orderedKeys ...string,
) {
	c.registerIntVar(defaultValue, ptr, isHotReloadable, valueScale, func(v int) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterIntVar registers a not hot-reloadable int config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterIntVar(defaultValue, valueScale int, orderedKeys ...string) *int {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, false, orderedKeys...).(*int)
	c.registerIntVar(defaultValue, ptr, false, valueScale, func(v int) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableIntVar registers a hot-reloadable int config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableIntVar(defaultValue, valueScale int, orderedKeys ...string) *Reloadable[int] {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, true, orderedKeys...).(*Reloadable[int])
	c.registerIntVar(defaultValue, ptr, true, valueScale, func(v int) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerIntVar(
	defaultValue int, ptr any, isHotReloadable bool, valueScale int, store func(int), orderedKeys ...string,
) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   valueScale,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.GetInt(key, defaultValue) * valueScale)
			return
		}
	}
	store(defaultValue * valueScale)
}

// RegisterBoolConfigVariable registers bool config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterBoolVar or RegisterReloadableBoolVar instead
func RegisterBoolConfigVariable(defaultValue bool, ptr *bool, isHotReloadable bool, orderedKeys ...string) {
	Default.RegisterBoolConfigVariable(defaultValue, ptr, isHotReloadable, orderedKeys...)
}

// RegisterBoolVar registers a not hot-reloadable bool config variable
// WARNING: a different order of keys will result in a different variable
func RegisterBoolVar(defaultValue bool, orderedKeys ...string) *bool {
	return Default.RegisterBoolVar(defaultValue, orderedKeys...)
}

// RegisterReloadableBoolVar registers a hot-reloadable bool config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableBoolVar(defaultValue bool, orderedKeys ...string) *Reloadable[bool] {
	return Default.RegisterReloadableBoolVar(defaultValue, orderedKeys...)
}

// RegisterBoolConfigVariable registers bool config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterBoolVar or RegisterReloadableBoolVar instead
func (c *Config) RegisterBoolConfigVariable(defaultValue bool, ptr *bool, isHotReloadable bool, orderedKeys ...string) {
	c.registerBoolVar(defaultValue, ptr, isHotReloadable, func(v bool) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterBoolVar registers a not hot-reloadable bool config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterBoolVar(defaultValue bool, orderedKeys ...string) *bool {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, false, orderedKeys...).(*bool)
	c.registerBoolVar(defaultValue, ptr, false, func(v bool) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableBoolVar registers a hot-reloadable bool config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableBoolVar(defaultValue bool, orderedKeys ...string) *Reloadable[bool] {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, true, orderedKeys...).(*Reloadable[bool])
	c.registerBoolVar(defaultValue, ptr, true, func(v bool) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerBoolVar(defaultValue bool, ptr any, isHotReloadable bool, store func(bool), orderedKeys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		c.bindEnv(key)
		if c.IsSet(key) {
			store(c.GetBool(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// RegisterFloat64ConfigVariable registers float64 config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterFloat64Var or RegisterReloadableFloat64Var instead
func RegisterFloat64ConfigVariable(defaultValue float64, ptr *float64, isHotReloadable bool, orderedKeys ...string) {
	Default.RegisterFloat64ConfigVariable(defaultValue, ptr, isHotReloadable, orderedKeys...)
}

// RegisterFloat64Var registers a not hot-reloadable float64 config variable
// WARNING: a different order of keys will result in a different variable
func RegisterFloat64Var(defaultValue float64, orderedKeys ...string) *float64 {
	return Default.RegisterFloat64Var(defaultValue, orderedKeys...)
}

// RegisterReloadableFloat64Var registers a hot-reloadable float64 config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableFloat64Var(defaultValue float64, orderedKeys ...string) *Reloadable[float64] {
	return Default.RegisterReloadableFloat64Var(defaultValue, orderedKeys...)
}

// RegisterFloat64ConfigVariable registers float64 config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterFloat64Var or RegisterReloadableFloat64Var instead
func (c *Config) RegisterFloat64ConfigVariable(defaultValue float64, ptr *float64, isHotReloadable bool, orderedKeys ...string) {
	c.registerFloat64Var(defaultValue, ptr, isHotReloadable, func(v float64) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterFloat64Var registers a not hot-reloadable float64 config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterFloat64Var(defaultValue float64, orderedKeys ...string) *float64 {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, false, orderedKeys...).(*float64)
	c.registerFloat64Var(defaultValue, ptr, false, func(v float64) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableFloat64Var registers a hot-reloadable float64 config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableFloat64Var(defaultValue float64, orderedKeys ...string) *Reloadable[float64] {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, true, orderedKeys...).(*Reloadable[float64])
	c.registerFloat64Var(defaultValue, ptr, true, func(v float64) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerFloat64Var(
	defaultValue float64, ptr any, isHotReloadable bool, store func(float64), orderedKeys ...string,
) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   1.0,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		c.bindEnv(key)
		if c.IsSet(key) {
			store(c.GetFloat64(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// RegisterInt64ConfigVariable registers int64 config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterInt64Var or RegisterReloadableInt64Var instead
func RegisterInt64ConfigVariable(defaultValue int64, ptr *int64, isHotReloadable bool, valueScale int64, orderedKeys ...string) {
	Default.RegisterInt64ConfigVariable(defaultValue, ptr, isHotReloadable, valueScale, orderedKeys...)
}

// RegisterInt64Var registers a not hot-reloadable int64 config variable
// WARNING: a different order of keys will result in a different variable
func RegisterInt64Var(defaultValue, valueScale int64, orderedKeys ...string) *int64 {
	return Default.RegisterInt64Var(defaultValue, valueScale, orderedKeys...)
}

// RegisterReloadableInt64Var registers a hot-reloadable int64 config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableInt64Var(defaultValue, valueScale int64, orderedKeys ...string) *Reloadable[int64] {
	return Default.RegisterReloadableInt64Var(defaultValue, valueScale, orderedKeys...)
}

// RegisterInt64ConfigVariable registers int64 config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterInt64Var or RegisterReloadableInt64Var instead
func (c *Config) RegisterInt64ConfigVariable(
	defaultValue int64, ptr *int64, isHotReloadable bool, valueScale int64, orderedKeys ...string,
) {
	c.registerInt64Var(defaultValue, ptr, isHotReloadable, valueScale, func(v int64) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterInt64Var registers a not hot-reloadable int64 config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterInt64Var(defaultValue, valueScale int64, orderedKeys ...string) *int64 {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, false, orderedKeys...).(*int64)
	c.registerInt64Var(defaultValue, ptr, false, valueScale, func(v int64) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableInt64Var registers a not hot-reloadable int64 config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableInt64Var(defaultValue, valueScale int64, orderedKeys ...string) *Reloadable[int64] {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, true, orderedKeys...).(*Reloadable[int64])
	c.registerInt64Var(defaultValue, ptr, true, valueScale, func(v int64) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerInt64Var(
	defaultValue int64, ptr any, isHotReloadable bool, valueScale int64, store func(int64), orderedKeys ...string,
) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   valueScale,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		c.bindEnv(key)
		if c.IsSet(key) {
			store(c.GetInt64(key, defaultValue) * valueScale)
			return
		}
	}
	store(defaultValue * valueScale)
}

// RegisterDurationConfigVariable registers duration config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterDurationVar or RegisterReloadableDurationVar instead
func RegisterDurationConfigVariable(
	defaultValueInTimescaleUnits int64, ptr *time.Duration, isHotReloadable bool, timeScale time.Duration, orderedKeys ...string,
) {
	Default.RegisterDurationConfigVariable(defaultValueInTimescaleUnits, ptr, isHotReloadable, timeScale, orderedKeys...)
}

// RegisterDurationVar registers a not hot-reloadable duration config variable
// WARNING: a different order of keys will result in a different variable
func RegisterDurationVar(
	defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string,
) *time.Duration {
	return Default.RegisterDurationVar(defaultValueInTimescaleUnits, timeScale, orderedKeys...)
}

// RegisterReloadableDurationVar registers a not hot-reloadable duration config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableDurationVar(defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string) *Reloadable[time.Duration] {
	return Default.RegisterReloadableDurationVar(defaultValueInTimescaleUnits, timeScale, orderedKeys...)
}

// RegisterDurationConfigVariable registers duration config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterDurationVar or RegisterReloadableDurationVar instead
func (c *Config) RegisterDurationConfigVariable(
	defaultValueInTimescaleUnits int64, ptr *time.Duration, isHotReloadable bool, timeScale time.Duration, orderedKeys ...string,
) {
	c.registerDurationVar(defaultValueInTimescaleUnits, ptr, isHotReloadable, timeScale, func(v time.Duration) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterDurationVar registers a not hot-reloadable duration config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterDurationVar(
	defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string,
) *time.Duration {
	ptr := getOrCreatePointer(
		c.vars, c.varsMisuses, &c.varsLock, time.Duration(defaultValueInTimescaleUnits)*timeScale, false, orderedKeys...,
	).(*time.Duration)
	c.registerDurationVar(defaultValueInTimescaleUnits, ptr, false, timeScale, func(v time.Duration) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableDurationVar registers a hot-reloadable duration config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableDurationVar(
	defaultValueInTimescaleUnits int64, timeScale time.Duration, orderedKeys ...string,
) *Reloadable[time.Duration] {
	ptr := getOrCreatePointer(
		c.vars, c.varsMisuses, &c.varsLock, time.Duration(defaultValueInTimescaleUnits)*timeScale, true, orderedKeys...,
	).(*Reloadable[time.Duration])
	c.registerDurationVar(defaultValueInTimescaleUnits, ptr, true, timeScale, func(v time.Duration) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerDurationVar(
	defaultValueInTimescaleUnits int64, ptr any, isHotReloadable bool, timeScale time.Duration,
	store func(time.Duration), orderedKeys ...string,
) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		multiplier:   timeScale,
		defaultValue: defaultValueInTimescaleUnits,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.GetDuration(key, defaultValueInTimescaleUnits, timeScale))
			return
		}
	}
	store(time.Duration(defaultValueInTimescaleUnits) * timeScale)
}

// RegisterStringConfigVariable registers string config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterStringVar or RegisterReloadableStringVar instead
func RegisterStringConfigVariable(defaultValue string, ptr *string, isHotReloadable bool, orderedKeys ...string) {
	Default.RegisterStringConfigVariable(defaultValue, ptr, isHotReloadable, orderedKeys...)
}

// RegisterStringVar registers a not hot-reloadable string config variable
// WARNING: a different order of keys will result in a different variable
func RegisterStringVar(defaultValue string, orderedKeys ...string) *string {
	return Default.RegisterStringVar(defaultValue, orderedKeys...)
}

// RegisterReloadableStringVar registers a hot-reloadable string config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableStringVar(defaultValue string, orderedKeys ...string) *Reloadable[string] {
	return Default.RegisterReloadableStringVar(defaultValue, orderedKeys...)
}

// RegisterStringConfigVariable registers string config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterStringVar or RegisterReloadableStringVar instead
func (c *Config) RegisterStringConfigVariable(defaultValue string, ptr *string, isHotReloadable bool, orderedKeys ...string) {
	c.registerStringVar(defaultValue, ptr, isHotReloadable, func(v string) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterStringVar registers a not hot-reloadable string config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterStringVar(defaultValue string, orderedKeys ...string) *string {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, false, orderedKeys...).(*string)
	c.registerStringVar(defaultValue, ptr, false, func(v string) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableStringVar registers a hot-reloadable string config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableStringVar(defaultValue string, orderedKeys ...string) *Reloadable[string] {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, true, orderedKeys...).(*Reloadable[string])
	c.registerStringVar(defaultValue, ptr, true, func(v string) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerStringVar(defaultValue string, ptr any, isHotReloadable bool, store func(string), orderedKeys ...string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.GetString(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// RegisterStringSliceConfigVariable registers string slice config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterStringSliceVar or RegisterReloadableStringSliceVar instead
func RegisterStringSliceConfigVariable(defaultValue []string, ptr *[]string, isHotReloadable bool, orderedKeys ...string) {
	Default.RegisterStringSliceConfigVariable(defaultValue, ptr, isHotReloadable, orderedKeys...)
}

// RegisterStringSliceVar registers a not hot-reloadable string slice config variable
// WARNING: a different order of keys will result in a different variable
func RegisterStringSliceVar(defaultValue []string, orderedKeys ...string) *[]string {
	return Default.RegisterStringSliceVar(defaultValue, orderedKeys...)
}

// RegisterReloadableStringSliceVar registers a hot-reloadable string slice config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableStringSliceVar(defaultValue []string, orderedKeys ...string) *Reloadable[[]string] {
	return Default.RegisterReloadableStringSliceVar(defaultValue, orderedKeys...)
}

// RegisterStringSliceConfigVariable registers string slice config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterStringSliceVar or RegisterReloadableStringSliceVar instead
func (c *Config) RegisterStringSliceConfigVariable(
	defaultValue []string, ptr *[]string, isHotReloadable bool, orderedKeys ...string,
) {
	c.registerStringSliceVar(defaultValue, ptr, isHotReloadable, func(v []string) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterStringSliceVar registers a not hot-reloadable string slice config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterStringSliceVar(defaultValue []string, orderedKeys ...string) *[]string {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, false, orderedKeys...).(*[]string)
	c.registerStringSliceVar(defaultValue, ptr, false, func(v []string) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableStringSliceVar registers a hot-reloadable string slice config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableStringSliceVar(defaultValue []string, orderedKeys ...string) *Reloadable[[]string] {
	ptr := getOrCreatePointer(c.vars, c.varsMisuses, &c.varsLock, defaultValue, true, orderedKeys...).(*Reloadable[[]string])
	c.registerStringSliceVar(defaultValue, ptr, true, func(v []string) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerStringSliceVar(
	defaultValue []string, ptr any, isHotReloadable bool, store func([]string), orderedKeys ...string,
) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.GetStringSlice(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

// RegisterStringMapConfigVariable registers string map config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterStringMapVar or RegisterReloadableStringMapVar instead
func RegisterStringMapConfigVariable(
	defaultValue map[string]interface{}, ptr *map[string]interface{}, isHotReloadable bool, orderedKeys ...string,
) {
	Default.RegisterStringMapConfigVariable(defaultValue, ptr, isHotReloadable, orderedKeys...)
}

// RegisterStringMapVar registers a not hot-reloadable string map config variable
// WARNING: a different order of keys will result in a different variable
func RegisterStringMapVar(
	defaultValue map[string]interface{}, orderedKeys ...string,
) *map[string]interface{} {
	return Default.RegisterStringMapVar(defaultValue, orderedKeys...)
}

// RegisterReloadableStringMapVar registers a hot-reloadable string map config variable
// WARNING: a different order of keys will result in a different variable
func RegisterReloadableStringMapVar(
	defaultValue map[string]interface{}, orderedKeys ...string,
) *Reloadable[map[string]interface{}] {
	return Default.RegisterReloadableStringMapVar(defaultValue, orderedKeys...)
}

// RegisterStringMapConfigVariable registers string map config variable
// WARNING: a different order of keys will result in a different variable
// Deprecated: use RegisterStringMapVar or RegisterReloadableStringMapVar instead
func (c *Config) RegisterStringMapConfigVariable(
	defaultValue map[string]interface{}, ptr *map[string]interface{}, isHotReloadable bool, orderedKeys ...string,
) {
	c.registerStringMapVar(defaultValue, ptr, isHotReloadable, func(v map[string]interface{}) {
		*ptr = v
	}, orderedKeys...)
}

// RegisterStringMapVar registers a not hot-reloadable string map config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterStringMapVar(defaultValue map[string]interface{}, orderedKeys ...string) *map[string]interface{} {
	ptr := getOrCreatePointer(
		c.vars, c.varsMisuses, &c.varsLock, defaultValue, false, orderedKeys...,
	).(*map[string]interface{})
	c.registerStringMapVar(defaultValue, ptr, false, func(v map[string]interface{}) {
		*ptr = v
	}, orderedKeys...)
	return ptr
}

// RegisterReloadableStringMapVar registers a hot-reloadable string map config variable
// WARNING: a different order of keys will result in a different variable
func (c *Config) RegisterReloadableStringMapVar(
	defaultValue map[string]interface{}, orderedKeys ...string,
) *Reloadable[map[string]interface{}] {
	ptr := getOrCreatePointer(
		c.vars, c.varsMisuses, &c.varsLock, defaultValue, true, orderedKeys...,
	).(*Reloadable[map[string]interface{}])
	c.registerStringMapVar(defaultValue, ptr, true, func(v map[string]interface{}) {
		ptr.store(v)
	}, orderedKeys...)
	return ptr
}

func (c *Config) registerStringMapVar(
	defaultValue map[string]interface{}, ptr any, isHotReloadable bool, store func(map[string]interface{}), orderedKeys ...string,
) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	c.hotReloadableConfigLock.Lock()
	defer c.hotReloadableConfigLock.Unlock()
	configVar := configValue{
		value:        ptr,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	}

	if isHotReloadable {
		c.appendVarToConfigMaps(orderedKeys, &configVar)
	}

	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.GetStringMap(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

func (c *Config) appendVarToConfigMaps(keys []string, configVar *configValue) {
	key := strings.Join(keys, ",")
	if _, ok := c.hotReloadableConfig[key]; !ok {
		c.hotReloadableConfig[key] = make([]*configValue, 0)
	}
	c.hotReloadableConfig[key] = append(c.hotReloadableConfig[key], configVar)
}

type configTypes interface {
	int | int64 | string | time.Duration | bool | float64 | []string | map[string]interface{}
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
