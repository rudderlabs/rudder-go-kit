package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cast"
)

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
		c.checkConfigForChanges(c.hotReloadableConfig)
	}()
	if !c.enableNonReloadableAdvancedDetection {
		c.checkAndNotifyNonReloadableConfig()
		return
	}
	func() { // wrap in a function to unlock the nonReloadableConfigLock in case of panic
		c.nonReloadableConfigLock.RLock()
		defer c.nonReloadableConfigLock.RUnlock()
		c.checkConfigForChanges(c.nonReloadableConfig)
	}()
}

func swapHotReloadableConfig[T configTypes](key string, reloadableValue *Reloadable[T], newValue T, compare func(T, T) bool, notifier *notifier) {
	if oldValue, swapped := reloadableValue.swapIfNotEqual(newValue, compare); swapped {
		notifier.notifyReloadableConfigChange(key, oldValue, newValue)
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
				// fallback to deep comparison for complex types that cannot be cast to string
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

func (c *Config) checkConfigForChanges(configMap map[string]*configValue) {
	for _, configVal := range configMap {
		key := strings.Join(configVal.keys, ",")
		switch value := configVal.value.(type) {
		case int, *Reloadable[int]:
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
			switch value := value.(type) {
			case int: // non-reloadable int
				if value != _value {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[int]: // reloadable int
				swapHotReloadableConfig(key, value, _value, compare[int](), c.notifier)
			}
		case int64, *Reloadable[int64]:
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
			switch value := value.(type) {
			case int64: // non-reloadable int64
				if value != _value {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[int64]: // reloadable int64
				swapHotReloadableConfig(key, value, _value, compare[int64](), c.notifier)
			}
		case string, *Reloadable[string]:
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
			switch value := value.(type) {
			case string: // non-reloadable string
				if value != _value {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[string]: // reloadable string
				swapHotReloadableConfig(key, value, _value, compare[string](), c.notifier)
			}
		case time.Duration, *Reloadable[time.Duration]:
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
			switch value := value.(type) {
			case time.Duration: // non-reloadable duration
				if value != _value {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[time.Duration]: // reloadable duration
				swapHotReloadableConfig(key, value, _value, compare[time.Duration](), c.notifier)
			}
		case bool, *Reloadable[bool]:
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
			switch value := value.(type) {
			case bool: // non-reloadable bool
				if configVal.value != _value {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[bool]: // reloadable bool
				swapHotReloadableConfig(key, value, _value, compare[bool](), c.notifier)
			}
		case float64, *Reloadable[float64]:
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
			switch value := value.(type) {
			case float64: // non-reloadable float64
				if configVal.value != _value {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[float64]: // reloadable float64
				swapHotReloadableConfig(key, value, _value, compare[float64](), c.notifier)
			}

		case []string, *Reloadable[[]string]:
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
			switch value := value.(type) {
			case []string: // non-reloadable slice
				if slices.Compare(value, _value) != 0 {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[[]string]: // reloadable slice
				swapHotReloadableConfig(key, value, _value, func(a, b []string) bool {
					return slices.Compare(a, b) == 0
				}, c.notifier)
			}
		case map[string]any, *Reloadable[map[string]any]:
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
			switch value := value.(type) {
			case map[string]any: // non-reloadable map
				if !mapDeepEqual(value, _value) {
					configVal.value = _value
					c.notifier.notifyNonReloadableConfigChange(key)
				}
			case *Reloadable[map[string]any]: // reloadable map
				swapHotReloadableConfig(key, value, _value, func(a, b map[string]any) bool {
					return mapDeepEqual(a, b)
				}, c.notifier)
			}
		}
	}
}
