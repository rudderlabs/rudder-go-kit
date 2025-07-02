package jsonparser

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// tidwallJSONParser is the implementation of JSONParser using gjson/sjson libraries
type tidwallJSONParser struct{}

// GetValue retrieves the value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetValue(data []byte, keys ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyJSON
	}

	if len(keys) == 0 {
		return nil, ErrNoKeysProvided
	}

	key := keys[0]
	if key == "" {
		return nil, ErrEmptyKey
	}

	// Use gjson to get the value
	result := gjson.GetBytes(data, getPath(keys...))
	if !result.Exists() {
		return nil, ErrKeyNotFound
	}

	return []byte(result.Raw), nil
}

// GetValueOrEmpty retrieves the raw value for a given key from JSON bytes.
// If the key does not exist or json is invalid, it returns an empty byte slice.
func (p *tidwallJSONParser) GetValueOrEmpty(data []byte, keys ...string) []byte {
	if len(data) == 0 || len(keys) == 0 {
		return []byte{}
	}

	key := keys[0]
	if key == "" {
		return []byte{}
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return []byte{}
	}

	return []byte(result.Raw)
}

// GetBoolean retrieves a boolean value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetBoolean(data []byte, keys ...string) (bool, error) {
	if len(data) == 0 {
		return false, ErrEmptyJSON
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return false, ErrKeyNotFound
	}

	// Check if the value is a boolean
	if result.Type != gjson.True && result.Type != gjson.False {
		return false, ErrNotOfExpectedType
	}

	return result.Bool(), nil
}

// GetBooleanOrFalse retrieves a boolean value for a given path from JSON bytes.
// If the key does not exist, json is invalid or value is not parsable to a boolean, it returns false.
// e.g `GetBooleanOrFalse({"key": true}, "key1")` returns false.
//
//		   `GetBooleanOrFalse({"key": true}, "key")` returns true.
//	    `GetBooleanOrFalse({"key": "val"}, "key")` returns false as `val` is not parsable to boolean.
//	    `GetBooleanOrFalse({"key": 1}, "key")` returns true.
func (p *tidwallJSONParser) GetBooleanOrFalse(data []byte, keys ...string) bool {
	if len(data) == 0 || len(keys) == 0 {
		return false
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return false
	}

	return result.Bool()
}

// GetInt retrieves an integer value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetInt(data []byte, keys ...string) (int64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return 0, ErrKeyNotFound
	}

	// Check if the value is a number
	if result.Type != gjson.Number {
		return 0, ErrNotOfExpectedType
	}

	return result.Int(), nil
}

// GetIntOrZero retrieves an integer value for a given key from JSON bytes.
// If the key does not exist, json is invalid or value is not parsable to an integer, it returns 0.
// e.g `GetIntOrZero({"key": 123}, "key1")` returns 0.
//
//	`GetIntOrZero({"key": 123}, "key")` returns 123.
//	`GetIntOrZero({"key": "val"}, "key")` returns 0 as `val` is not parsable to integer.
//	`GetIntOrZero({"key": true}, "key")` returns 1 as `true` is parsed to 1.
//	`GetIntOrZero({"key": "123"}, "key")` returns 123 as `"123"` is parsed to 123.
func (p *tidwallJSONParser) GetIntOrZero(data []byte, keys ...string) int64 {
	if len(data) == 0 || len(keys) == 0 {
		return 0
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return 0
	}

	return result.Int()
}

// GetFloat retrieves a float value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetFloat(data []byte, keys ...string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return 0, ErrKeyNotFound
	}

	// Check if the value is a number
	if result.Type != gjson.Number {
		return 0, ErrNotOfExpectedType
	}

	return result.Float(), nil
}

// GetFloatOrZero retrieves a float value for a given key from JSON bytes.
// If the key does not exist, json is invalid or value is not parsable to a float, it returns 0.
// e.g `GetFloatOrZero({"key": 123.456}, "key1") returns 0.
//
//	`GetFloatOrZero({"key": 123.456}, "key")` returns 123.456.
//	`GetFloatOrZero({"key": "123.456"}, "key")` returns 123.456 as `"123.456"` is parsed to 123.456.
//	`GetFloatOrZero({"key": "val"}, "key")` returns 0 as `val` is not parsable to float.
//	`GetFloatOrZero({"key": true}, "key")` returns 1 as `true` is parsed to 1.
func (p *tidwallJSONParser) GetFloatOrZero(data []byte, keys ...string) float64 {
	if len(data) == 0 || len(keys) == 0 {
		return 0
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return 0
	}

	return result.Float()
}

// GetString retrieves a string value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetString(data []byte, keys ...string) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyJSON
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return "", ErrKeyNotFound
	}

	// Check if the value is a string
	if result.Type != gjson.String {
		return "", ErrNotOfExpectedType
	}

	return result.String(), nil
}

// GetStringOrEmpty retrieves a string value for a given key from JSON bytes.
// If the key does not exist, json is invalid or value is not parsable to a string, it returns an empty string.
// e.g `GetStringOrEmpty({"key": "val"}, "key1")` returns "".
//
//	`GetStringOrEmpty({"key": "val"}, "key")` returns "val".
//	`GetStringOrEmpty({"key": 123}, "key")` returns "123" as 123 is parsed to string.
//	`GetStringOrEmpty({"key": true}, "key")` returns "true" as `true` is parsed to string.
func (p *tidwallJSONParser) GetStringOrEmpty(data []byte, keys ...string) string {
	if len(data) == 0 || len(keys) == 0 {
		return ""
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return ""
	}

	return result.String()
}

// SetValue sets the value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetValue(data []byte, value interface{}, keys ...string) ([]byte, error) {
	if len(data) == 0 {
		// If data is empty, create a new JSON object
		data = []byte("{}")
	}

	if len(keys) == 0 {
		return nil, ErrNoKeysProvided
	}

	if keys[0] == "" {
		return nil, ErrEmptyKey
	}

	// Join keys with dots to create a path
	path := getPath(keys...)
	if path == "" {
		return nil, ErrEmptyKey
	}

	// Use sjson to set the value
	result, err := sjson.SetBytes(data, path, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set value: %w", err)
	}

	return result, nil
}

// SetBoolean sets a boolean value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetBoolean(data []byte, value bool, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// SetInt sets an integer value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetInt(data []byte, value int64, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// SetFloat sets a float value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetFloat(data []byte, value float64, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// SetString sets a string value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetString(data []byte, value string, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// DeleteKey deletes a key from JSON bytes using sjson
func (p *tidwallJSONParser) DeleteKey(data []byte, keys ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyJSON
	}

	if len(keys) == 0 {
		return nil, ErrNoKeysProvided
	}

	if keys[0] == "" {
		return nil, ErrEmptyKey
	}

	// Join keys with dots to create a path
	path := getPath(keys...)
	if path == "" {
		return nil, ErrEmptyKey
	}

	// Use sjson to delete the key
	resultData, err := sjson.DeleteBytes(data, path)
	if err != nil {
		return nil, fmt.Errorf("failed to delete key: %w", err)
	}

	return resultData, nil
}

func getPath(keys ...string) string {
	if len(keys) == 0 {
		return ""
	}
	// Join keys with dots to create a path
	path := strings.Join(lo.Map(keys, func(key string, _ int) string {
		if key[0] == '[' {
			return key[1 : len(key)-1]
		}
		return key
	}), ".")
	return path
}
