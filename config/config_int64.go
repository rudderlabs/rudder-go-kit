package config

// GetInt64 gets int64 value from config
func GetInt64(key string, defaultValue int64) (value int64) {
	return Default.GetInt64(key, defaultValue)
}

// GetInt64 gets int64 value from config
func (c *Config) GetInt64(key string, defaultValue int64) (value int64) {
	return c.GetInt64Var(defaultValue, 1, key)
}

// GetInt64Var registers a not hot-reloadable int64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetInt64Var(defaultValue, valueScale int64, orderedKeys ...string) int64 {
	return Default.GetInt64Var(defaultValue, valueScale, orderedKeys...)
}

// GetInt64Var registers a not hot-reloadable int64 config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetInt64Var(defaultValue, valueScale int64, orderedKeys ...string) int64 {
	var ret int64
	c.storeAndRegisterInt64Var(defaultValue, &ret, valueScale, func(v int64) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getInt64Internal gets int64 value from config
func (c *Config) getInt64Internal(key string, defaultValue int64) (value int64) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetInt64(key)
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
		c.storeInt64Var(defaultValue, valueScale, func(v int64) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeInt64Var(defaultValue, valueScale int64, store func(int64), orderedKeys ...string) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getInt64Internal(key, defaultValue) * valueScale)
			return
		}
	}
	store(defaultValue * valueScale)
}

func (c *Config) storeAndRegisterInt64Var(defaultValue int64, ptr any, valueScale int64, store func(int64), orderedKeys ...string) {
	c.storeInt64Var(defaultValue, valueScale, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, defaultValue*valueScale, &configValue{
		value:        *ptr.(*int64),
		multiplier:   valueScale,
		defaultValue: defaultValue,
		keys:         orderedKeys,
	})
}
