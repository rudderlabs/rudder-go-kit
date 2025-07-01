package config

// GetStringMap gets string map value from config
func GetStringMap(key string, defaultValue map[string]any) (value map[string]any) {
	return Default.GetStringMap(key, defaultValue)
}

// GetStringMap gets string map value from config
func (c *Config) GetStringMap(key string, defaultValue map[string]any) (value map[string]any) {
	return c.GetStringMapVar(defaultValue, key)
}

// GetStringMapVar registers a not hot-reloadable string map config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetStringMapVar(defaultValue map[string]any, orderedKeys ...string) map[string]any {
	return Default.GetStringMapVar(defaultValue, orderedKeys...)
}

// GetStringMapVar registers a not hot-reloadable string map config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetStringMapVar(
	defaultValue map[string]any, orderedKeys ...string,
) map[string]any {
	var ret map[string]any
	c.storeAndRegisterStringMapVar(defaultValue, &ret, func(v map[string]any) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getStringMapInternal gets string map value from config
func (c *Config) getStringMapInternal(key string, defaultValue map[string]any) (value map[string]any) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetStringMap(key)
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
		c.storeStringMapVar(defaultValue, func(v map[string]any) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeStringMapVar(
	defaultValue map[string]any, store func(map[string]any), orderedKeys ...string,
) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getStringMapInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

func (c *Config) storeAndRegisterStringMapVar(defaultValue map[string]any, ptr any, store func(map[string]any), orderedKeys ...string) {
	c.storeStringMapVar(defaultValue, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, defaultValue, &configValue{
		value:        *ptr.(*map[string]any),
		defaultValue: defaultValue,
		keys:         orderedKeys,
	})
}
