package config

import "fmt"

// GetInt gets int value from config
//
// Deprecated: Use GetIntVar(defaultValue, 1, key) instead.
//
//go:fix inline
func GetInt(key string, defaultValue int) (value int) {
	return GetIntVar(defaultValue, 1, key)
}

// GetInt gets int value from config
//
// Deprecated: Use (*Config).GetIntVar(defaultValue, 1, key) instead.
//
//go:fix inline
func (c *Config) GetInt(key string, defaultValue int) (value int) {
	return c.GetIntVar(defaultValue, 1, key)
}

// GetIntVar registers a not hot-reloadable int config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetIntVar(defaultValue, valueScale int, orderedKeys ...string) int {
	return Default.GetIntVar(defaultValue, valueScale, orderedKeys...)
}

// GetIntVar registers a not hot-reloadable int config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetIntVar(defaultValue, valueScale int, orderedKeys ...string) int {
	var ret int
	c.storeAndRegisterIntVar(defaultValue, &ret, valueScale, func(v int) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getIntInternal gets int value from config
func (c *Config) getIntInternal(key string, defaultValue int) (value int) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetInt(key)
}

// MustGetInt gets int value from config or panics if the config doesn't exist
func MustGetInt(key string) (value int) {
	return Default.MustGetInt(key)
}

// MustGetInt gets int value from config or panics if the config doesn't exist
func (c *Config) MustGetInt(key string) (value int) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		panic(fmt.Errorf("config key %s not found", key))
	}
	return c.v.GetInt(key)
}

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
		c.storeIntVar(defaultValue, valueScale, func(v int) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeIntVar(defaultValue, valueScale int, store func(int), orderedKeys ...string) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getIntInternal(key, defaultValue) * valueScale)
			return
		}
	}
	store(defaultValue * valueScale)
}

func (c *Config) storeAndRegisterIntVar(defaultValue int, ptr any, valueScale int, store func(int), orderedKeys ...string) {
	c.storeIntVar(defaultValue, valueScale, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, defaultValue*valueScale, &configValue{
		value:        *ptr.(*int),
		multiplier:   valueScale,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	})
}
