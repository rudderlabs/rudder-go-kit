package config

import "fmt"

// GetString gets string value from config
//
// Deprecated: Use GetStringVar(defaultValue, key) instead.
//
//go:fix inline
func GetString(key, defaultValue string) (value string) {
	return GetStringVar(defaultValue, key)
}

// GetString gets string value from config
//
// Deprecated: Use (*Config).GetStringVar(defaultValue, key) instead.
//
//go:fix inline
func (c *Config) GetString(key, defaultValue string) (value string) {
	return c.GetStringVar(defaultValue, key)
}

// GetStringVar registers a not hot-reloadable string config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func GetStringVar(defaultValue string, orderedKeys ...string) string {
	return Default.GetStringVar(defaultValue, orderedKeys...)
}

// GetStringVar registers a not hot-reloadable string config variable
//
// WARNING: keys are being looked up in requested order and the value of the first found key is returned,
// e.g. asking for the same keys but in a different order can result in a different value to be returned
func (c *Config) GetStringVar(defaultValue string, orderedKeys ...string) string {
	var ret string
	c.storeAndRegisterStringVar(defaultValue, &ret, func(v string) {
		ret = v
	}, orderedKeys...)
	return ret
}

// getStringInternal gets string value from config
func (c *Config) getStringInternal(key, defaultValue string) (value string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		return defaultValue
	}
	return c.v.GetString(key)
}

// MustGetString gets string value from config or panics if the config doesn't exist
func MustGetString(key string) (value string) {
	return Default.MustGetString(key)
}

// MustGetString gets string value from config or panics if the config doesn't exist
func (c *Config) MustGetString(key string) (value string) {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	if !c.isSetInternal(key) {
		panic(fmt.Errorf("config key %s not found", key))
	}
	return c.v.GetString(key)
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
		c.storeStringVar(defaultValue, func(v string) {
			ptr.store(v)
		}, orderedKeys...)
	}
	return ptr
}

func (c *Config) storeStringVar(defaultValue string, store func(string), orderedKeys ...string) {
	for _, key := range orderedKeys {
		if c.IsSet(key) {
			store(c.getStringInternal(key, defaultValue))
			return
		}
	}
	store(defaultValue)
}

func (c *Config) storeAndRegisterStringVar(defaultValue string, ptr any, store func(string), orderedKeys ...string) {
	c.storeStringVar(defaultValue, store, orderedKeys...) // store before registering non-reloadable keys
	registerNonReloadableConfigKeys(c, defaultValue, &configValue{
		value:        *ptr.(*string),
		defaultValue: defaultValue,
		keys:         orderedKeys,
	})
}
