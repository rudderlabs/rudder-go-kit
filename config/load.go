package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

func (c *Config) load() {
	c.godotEnvErr = godotenv.Load()
	c.enableNonReloadableAdvancedDetection = getEnv("CONFIG_ADVANCED_DETECTION", "false") == "true"
	configPath := getEnv("CONFIG_PATH", "./config/config.yaml")

	v := viper.NewWithOptions(viper.EnvKeyReplacer(&envReplacer{c: c}))
	v.AutomaticEnv()
	bindLegacyEnv(v)

	v.SetConfigFile(configPath)

	// Find and read the config file
	// If config.yaml is not found or error with parsing. Use the default config values instead
	c.configPathErr = v.ReadInConfig()
	c.configPath = v.ConfigFileUsed()
	c.v = v

	c.currentSettings = c.getCurrentSettings()
	c.v.OnConfigChange(func(_ fsnotify.Event) {
		c.onConfigChange()
	})
	c.v.WatchConfig()
}

// getCurrentSettings retrieves the current configuration settings and flattens them into a map.
func (c *Config) getCurrentSettings() map[string]any {
	c.vLock.RLock()
	defer c.vLock.RUnlock()
	currentConfig := make(map[string]any)
	var flattenSettings func(envPrefix, prefix string, nested, flat map[string]any)
	flattenSettings = func(envPrefix, prefix string, nested, flat map[string]any) {
		for k, v := range nested {
			fullKey := k
			if prefix != "" {
				fullKey = prefix + "." + k
			}

			// Type switch for deeper nested maps
			switch child := v.(type) {
			case map[string]any:
				flattenSettings(envPrefix, fullKey, child, flat)
			case map[any]any: // in case of YAML decoding
				converted := make(map[string]any)
				for key, val := range child {
					if ks, ok := key.(string); ok {
						converted[ks] = val
					}
				}
				flattenSettings(envPrefix, fullKey, converted, flat)
			default:
				flat[strings.ToLower(fullKey)] = v
			}
		}
	}
	flattenSettings(c.envPrefix, "", c.v.AllSettings(), currentConfig)
	return currentConfig
}

// ConfigFileUsed returns the file used to load the config.
// If we failed to load the config file, it also returns an error.
func (c *Config) ConfigFileUsed() (string, error) {
	return c.configPath, c.configPathErr
}

// DotEnvLoaded returns an error if there was an error loading the .env file.
// It returns nil otherwise.
func (c *Config) DotEnvLoaded() error {
	return c.godotEnvErr
}

func (c *Config) onConfigChange() {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("error updating config variables: %v", r)
			fmt.Println(err)
		}
	}()
	func() { // wrap in a function to unlock the hotReloadableConfigLock in case of panic
		c.hotReloadableConfigLock.RLock()
		defer c.hotReloadableConfigLock.RUnlock()
		c.checkAndHotReloadConfig(c.hotReloadableConfig)
	}()

	if c.enableNonReloadableAdvancedDetection {
		func() { // wrap in a function to unlock the nonReloadableConfigLock in case of panic
			c.nonReloadableConfigLock.RLock()
			c.checkAndNotifyNonReloadConfigAdvanced(c.nonReloadableConfig)
			c.nonReloadableConfigLock.RUnlock()
		}()
	} else {
		c.checkAndNotifyNonReloadableConfig()
	}
}

// checkAndNotifyNonReloadableConfig checks for changes in non-reloadable config values
// and notifies subscribers if changes are detected
func (c *Config) checkAndNotifyNonReloadableConfig() {
	newConfig := c.getCurrentSettings()

	// Identify changed keys
	changedKeys := make(map[string]struct{})

	// Collect keys that were added or modified
	for key, newValue := range newConfig {
		oldValue, exists := c.currentSettings[key]
		if !exists {
			changedKeys[key] = struct{}{}
		} else {
			// first try to convert both values to string for comparison
			if old, new, err := func() (string, string, error) {
				old, err := cast.ToStringE(oldValue)
				if err != nil {
					return "", "", fmt.Errorf("error casting old value for key %s: %w", key, err)
				}
				new, err := cast.ToStringE(newValue)
				if err != nil {
					return "", "", fmt.Errorf("error casting new value for key %s: %w", key, err)
				}
				return old, new, nil
			}(); err == nil {
				if old != new {
					changedKeys[key] = struct{}{}
				}
			} else {
				// fallback to deep comparison for complex types that cannot be casted to string
				if !reflect.DeepEqual(oldValue, newValue) {
					changedKeys[key] = struct{}{}
				}
			}
		} // first try to cast both values to string for comparison

	}
	// Collect keys that were removed
	for key := range c.currentSettings {
		if _, exists := newConfig[key]; !exists {
			changedKeys[key] = struct{}{}
		}
	}

	func() { // wrap in a function to unlock the nonReloadableConfigLock in case of panic
		// Notify subscribers for non-reloadable config changes
		c.nonReloadableConfigLock.RLock()
		defer c.nonReloadableConfigLock.RUnlock()
		for key := range changedKeys {
			if originalKey, exists := c.nonReloadableKeys[key]; exists {
				c.notifier.notifyNonReloadableConfigChange(originalKey)
			}
		}
	}()

	// Update current config with new values
	c.currentSettings = newConfig
}

func (c *Config) checkAndHotReloadConfig(configMap map[string]*configValue) {
	for _, configVal := range configMap {
		key := strings.Join(configVal.keys, ",")
		value := configVal.value
		switch value := value.(type) {
		case *Reloadable[int]:
			var _value int
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getIntInternal(key, configVal.defaultValue.(int))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(int)
			}
			_value = _value * configVal.multiplier.(int)
			swapHotReloadableConfig(key, value, _value, compare[int](), c.notifier)
		case *Reloadable[int64]:
			var _value int64
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getInt64Internal(key, configVal.defaultValue.(int64))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(int64)
			}
			_value = _value * configVal.multiplier.(int64)
			swapHotReloadableConfig(key, value, _value, compare[int64](), c.notifier)
		case *Reloadable[string]:
			var _value string
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getStringInternal(key, configVal.defaultValue.(string))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(string)
			}
			swapHotReloadableConfig(key, value, _value, compare[string](), c.notifier)
		case *Reloadable[time.Duration]:
			var _value time.Duration
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getDurationInternal(key, configVal.defaultValue.(int64), configVal.multiplier.(time.Duration))
					break
				}
			}
			if !isSet {
				_value = time.Duration(configVal.defaultValue.(int64)) * configVal.multiplier.(time.Duration)
			}
			swapHotReloadableConfig(key, value, _value, compare[time.Duration](), c.notifier)
		case *Reloadable[bool]:
			var _value bool
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getBoolInternal(key, configVal.defaultValue.(bool))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(bool)
			}
			swapHotReloadableConfig(key, value, _value, compare[bool](), c.notifier)
		case *Reloadable[float64]:
			var _value float64
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getFloat64Internal(key, configVal.defaultValue.(float64))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(float64)
			}
			_value = _value * configVal.multiplier.(float64)
			swapHotReloadableConfig(key, value, _value, compare[float64](), c.notifier)
		case *Reloadable[[]string]:
			var _value []string
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getStringSliceInternal(key, configVal.defaultValue.([]string))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.([]string)
			}
			swapHotReloadableConfig(key, value, _value, func(a, b []string) bool {
				return slices.Compare(a, b) == 0
			}, c.notifier)
		case *Reloadable[map[string]any]:
			var _value map[string]any
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getStringMapInternal(key, configVal.defaultValue.(map[string]any))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(map[string]any)
			}
			swapHotReloadableConfig(key, value, _value, func(a, b map[string]any) bool {
				return mapDeepEqual(a, b)
			}, c.notifier)
		}
	}
}

func (c *Config) checkAndNotifyNonReloadConfigAdvanced(configMap map[string]*configValue) {
	for _, configVal := range configMap {
		key := strings.Join(configVal.keys, ",")
		switch value := configVal.value.(type) {
		case *int:
			var _value int
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getIntInternal(key, configVal.defaultValue.(int))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(int)
			}
			_value = _value * configVal.multiplier.(int)
			if *value != _value {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		case *int64:
			var _value int64
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getInt64Internal(key, configVal.defaultValue.(int64))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(int64)
			}
			_value = _value * configVal.multiplier.(int64)
			if *value != _value {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		case *string:
			var _value string
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getStringInternal(key, configVal.defaultValue.(string))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(string)
			}
			if *value != _value {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		case *time.Duration:
			var _value time.Duration
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true

					_value = c.getDurationInternal(key, configVal.defaultValue.(int64), configVal.multiplier.(time.Duration))
					break
				}
			}
			if !isSet {
				_value = time.Duration(configVal.defaultValue.(int64)) * configVal.multiplier.(time.Duration)
			}
			if *value != _value {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		case *bool:
			var _value bool
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true
					_value = c.getBoolInternal(key, configVal.defaultValue.(bool))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(bool)
			}
			if *value != _value {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		case *float64:
			var _value float64
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true

					_value = c.getFloat64Internal(key, configVal.defaultValue.(float64))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(float64)
			}
			_value = _value * configVal.multiplier.(float64)
			if *value != _value {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		case *[]string:
			var _value []string
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true

					_value = c.getStringSliceInternal(key, configVal.defaultValue.([]string))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.([]string)
			}
			if slices.Compare(*value, _value) != 0 {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		case *map[string]any:
			var _value map[string]any
			var isSet bool
			for _, key := range configVal.keys {
				if c.IsSet(key) {
					isSet = true

					_value = c.getStringMapInternal(key, configVal.defaultValue.(map[string]any))
					break
				}
			}
			if !isSet {
				_value = configVal.defaultValue.(map[string]any)
			}
			if !mapDeepEqual(*value, _value) {
				*value = _value
				c.notifier.notifyNonReloadableConfigChange(key)
			}
		}
	}
}

func swapHotReloadableConfig[T configTypes](key string, reloadableValue *Reloadable[T], newValue T, compare func(T, T) bool, notifier *notifier) {
	if oldValue, swapped := reloadableValue.swapIfNotEqual(newValue, compare); swapped {
		notifier.notifyReloadableConfigChange(key, oldValue, newValue)
	}
}
