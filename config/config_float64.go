package config

// GetFloat64 gets float64 value from config
//
// Deprecated: Use GetFloat64Var(defaultValue, key) instead.
//
//go:fix inline
func GetFloat64(key string, defaultValue float64) (value float64) {
	return GetFloat64Var(defaultValue, key)
}

// GetFloat64 gets float64 value from config
//
// Deprecated: Use (*Config).GetFloat64Var(defaultValue, key) instead.
//
//go:fix inline
func (c *Config) GetFloat64(key string, defaultValue float64) (value float64) {
	return c.GetFloat64Var(defaultValue, key)
}

// GetFloat64Var registers a not hot-reloadable float64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetFloat64Var(defaultValue float64, orderedKeys ...string) float64 {
	return Default.GetFloat64Var(defaultValue, orderedKeys...)
}

// GetFloat64Var registers a not hot-reloadable float64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetFloat64Var(defaultValue float64, orderedKeys ...string) float64 {
	var ret float64
	c.storeAndRegisterFloat64Var(defaultValue, &ret, func(v float64) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getFloat64Internal gets float64 value from config
func (c *Config) getFloat64Internal(key string, defaultValue float64) (value float64) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetFloat64(key)
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
		c.storeFloat64Var(defaultValue, func(v float64) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeFloat64Var(defaultValue float64, store func(float64), orderedKeys ...string) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getFloat64Internal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

func (c *Config) storeAndRegisterFloat64Var(defaultValue float64, ptr any, store func(float64), orderedKeys ...string) {
	c.storeFloat64Var(defaultValue, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, defaultValue, &configValue{
		value:        *ptr.(*float64),
		multiplier:   1.0,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	})
}
