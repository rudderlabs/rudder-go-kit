package config

// GetStringSlice gets string slice value from config
//
// Deprecated: Use GetStringSliceVar(defaultValue, key) instead.
//
//go:fix inline
func GetStringSlice(key string, defaultValue []string) (value []string) {
	return GetStringSliceVar(defaultValue, key)
}

// GetStringSlice gets string slice value from config
//
// Deprecated: Use (*Config).GetStringSliceVar(defaultValue, key) instead.
//
//go:fix inline
func (c *Config) GetStringSlice(key string, defaultValue []string) (value []string) {
	return c.GetStringSliceVar(defaultValue, key)
}

// GetStringSliceVar registers a not hot-reloadable string slice config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetStringSliceVar(defaultValue []string, orderedKeys ...string) []string {
	return Default.GetStringSliceVar(defaultValue, orderedKeys...)
}

// GetStringSliceVar registers a not hot-reloadable string slice config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetStringSliceVar(defaultValue []string, orderedKeys ...string) []string {
	var ret []string
	c.storeAndRegisterStringSliceVar(defaultValue, &ret, func(v []string) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getStringSliceInternal gets string slice value from config
func (c *Config) getStringSliceInternal(key string, defaultValue []string) (value []string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetStringSlice(key)
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
		c.storeStringSliceVar(defaultValue, func(v []string) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeStringSliceVar(defaultValue []string, store func([]string), orderedKeys ...string) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getStringSliceInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

func (c *Config) storeAndRegisterStringSliceVar(defaultValue []string, ptr any, store func([]string), orderedKeys ...string) {
	c.storeStringSliceVar(defaultValue, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, defaultValue, &configValue{
		value:        *ptr.(*[]string),
		defaultValue: defaultValue,
		keys:         orderedKeys,
	})
}
