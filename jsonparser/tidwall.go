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
func (p *tidwallJSONParser) GetValue(data []byte, path ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyJSON
	}

	if len(path) == 0 {
		return nil, ErrNoKeysProvided
	}

	if path[0] == "" {
		return nil, ErrEmptyKey
	}

	// Use gjson to get the value
	result := gjson.GetBytes(data, getPath(path...))
	if !result.Exists() {
		return nil, ErrKeyNotFound
	}

	return []byte(result.Raw), nil
}

// GetValueOrEmpty retrieves the raw value for a given key from JSON bytes.
// If the key does not exist or json is invalid, it returns an empty byte slice.
func (p *tidwallJSONParser) GetValueOrEmpty(data []byte, path ...string) []byte {
	if len(data) == 0 || len(path) == 0 {
		return nil
	}

	if path[0] == "" {
		return nil
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
	if !result.Exists() {
		return nil
	}

	return []byte(result.Raw)
}

// GetBoolean retrieves a boolean value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetBoolean(data []byte, path ...string) (bool, error) {
	if len(data) == 0 {
		return false, ErrEmptyJSON
	}

	if len(path) == 0 {
		return false, ErrNoKeysProvided
	}

	if path[0] == "" {
		return false, ErrEmptyKey
	}

	// Join path with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
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
func (p *tidwallJSONParser) GetBooleanOrFalse(data []byte, path ...string) bool {
	if len(data) == 0 || len(path) == 0 {
		return false
	}

	if path[0] == "" {
		return false
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
	if !result.Exists() {
		return false
	}

	return result.Bool()
}

// GetInt retrieves an integer value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetInt(data []byte, path ...string) (int64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	if len(path) == 0 {
		return 0, ErrNoKeysProvided
	}

	if path[0] == "" {
		return 0, ErrEmptyKey
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
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
func (p *tidwallJSONParser) GetIntOrZero(data []byte, path ...string) int64 {
	if len(data) == 0 || len(path) == 0 {
		return 0
	}

	if path[0] == "" {
		return 0
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
	if !result.Exists() {
		return 0
	}

	return result.Int()
}

// GetFloat retrieves a float value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetFloat(data []byte, path ...string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	if len(path) == 0 {
		return 0, ErrNoKeysProvided
	}

	if path[0] == "" {
		return 0, ErrEmptyKey
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
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
func (p *tidwallJSONParser) GetFloatOrZero(data []byte, path ...string) float64 {
	if len(data) == 0 || len(path) == 0 {
		return 0
	}

	if path[0] == "" {
		return 0
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
	if !result.Exists() {
		return 0
	}

	return result.Float()
}

// GetString retrieves a string value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetString(data []byte, path ...string) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyJSON
	}

	if len(path) == 0 {
		return "", ErrNoKeysProvided
	}

	key := path[0]
	if key == "" {
		return "", ErrEmptyKey
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
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
func (p *tidwallJSONParser) GetStringOrEmpty(data []byte, path ...string) string {
	if len(data) == 0 || len(path) == 0 {
		return ""
	}

	if path[0] == "" {
		return ""
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, dotSeperatedPath)
	if !result.Exists() {
		return ""
	}

	return result.String()
}

// SetValue sets the value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetValue(data []byte, value interface{}, path ...string) ([]byte, error) {
	if len(data) == 0 {
		// If data is empty, create a new JSON object
		data = []byte("{}")
	}

	if len(path) == 0 {
		return nil, ErrNoKeysProvided
	}

	if path[0] == "" {
		return nil, ErrEmptyKey
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use sjson to set the value
	result, err := sjson.SetBytes(data, dotSeperatedPath, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set value: %w", err)
	}

	return result, nil
}

// SetBoolean sets a boolean value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetBoolean(data []byte, value bool, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// SetInt sets an integer value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetInt(data []byte, value int64, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// SetFloat sets a float value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetFloat(data []byte, value float64, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// SetString sets a string value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetString(data []byte, value string, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// DeleteKey deletes a key from JSON bytes using sjson
func (p *tidwallJSONParser) DeleteKey(data []byte, path ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyJSON
	}

	if len(path) == 0 {
		return nil, ErrNoKeysProvided
	}

	if path[0] == "" {
		return nil, ErrEmptyKey
	}

	// Join keys with dots to create a path
	dotSeperatedPath := getPath(path...)

	// Use sjson to delete the key
	resultData, err := sjson.DeleteBytes(data, dotSeperatedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to delete key: %w", err)
	}

	return resultData, nil
}

func getPath(path ...string) string {
	if len(path) == 0 {
		return ""
	}
	// Join keys with dots to create a path
	return strings.Join(lo.Map(path, func(key string, _ int) string {
		if len(key) > 0 && key[0] == '[' {
			return key[1 : len(key)-1]
		}
		return key
	}), ".")
}
