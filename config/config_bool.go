package config

// GetBool gets bool value from config
//
// Deprecated: Use GetBoolVar(defaultValue, key) instead.
//
//go:fix inline
func GetBool(key string, defaultValue bool) (value bool) {
	return GetBoolVar(defaultValue, key)
}

// GetBool gets bool value from config
//
// Deprecated: Use (*Config).GetBoolVar(defaultValue, key) instead.
//
//go:fix inline
func (c *Config) GetBool(key string, defaultValue bool) (value bool) {
	return c.GetBoolVar(defaultValue, key)
}

// GetBoolVar registers a not hot-reloadable bool config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetBoolVar(defaultValue bool, orderedKeys ...string) bool {
	return Default.GetBoolVar(defaultValue, orderedKeys...)
}

// GetBoolVar registers a not hot-reloadable bool config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetBoolVar(defaultValue bool, orderedKeys ...string) bool {
	var ret bool
	c.storeAndRegisterBoolVar(defaultValue, &ret, func(v bool) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getBoolInternal gets bool value from config
func (c *Config) getBoolInternal(key string, defaultValue bool) (value bool) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetBool(key)
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
		c.storeBoolVar(defaultValue, func(v bool) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeBoolVar(defaultValue bool, store func(bool), orderedKeys ...string) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getBoolInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

func (c *Config) storeAndRegisterBoolVar(defaultValue bool, ptr any, store func(bool), orderedKeys ...string) {
	c.storeBoolVar(defaultValue, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, defaultValue, &configValue{
		value:        *ptr.(*bool),
		defaultValue: defaultValue,
		keys:         orderedKeys,
	})
}
