package jsonparser

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// tidwallJSONParser is the implementation of JSONParser using gjson/sjson libraries
type tidwallJSONParser struct{}

// GetValue retrieves the value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetValue(data []byte, keys ...string) (interface{}, error) {
	if len(data) == 0 {
		return nil, EmptyJSONError
	}

	// Handle empty keys or no keys - return the entire JSON object
	if len(keys) == 0 || (len(keys) == 1 && keys[0] == "") {
		var result interface{}
		if err := jsonrs.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return result, nil
	}

	// Use gjson to get the value
	result := gjson.GetBytes(data, getPath(keys...))
	if !result.Exists() {
		return nil, fmt.Errorf("key not found: %s", keys)
	}

	// Convert the gjson.Result to the appropriate Go type
	switch result.Type {
	case gjson.String:
		return result.String(), nil
	case gjson.Number:
		return result.Float(), nil
	case gjson.True:
		return true, nil
	case gjson.False:
		return false, nil
	case gjson.Null:
		// nolint: nilnil
		return nil, nil
	case gjson.JSON:
		if result.IsArray() {
			var arr []interface{}
			if err := jsonrs.Unmarshal([]byte(result.Raw), &arr); err != nil {
				return nil, fmt.Errorf("failed to unmarshal array: %w", err)
			}
			return arr, nil
		}
		var obj map[string]interface{}
		if err := jsonrs.Unmarshal([]byte(result.Raw), &obj); err != nil {
			return nil, fmt.Errorf("failed to unmarshal object: %w", err)
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("unknown value type")
	}
}

// GetBoolean retrieves a boolean value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetBoolean(data []byte, keys ...string) (bool, error) {
	if len(data) == 0 {
		return false, EmptyJSONError
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return false, KeyNotFoundError
	}

	// Check if the value is a boolean
	if result.Type != gjson.True && result.Type != gjson.False {
		return false, fmt.Errorf("value is not a boolean: %s", path)
	}

	return result.Bool(), nil
}

// GetInt retrieves an integer value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetInt(data []byte, keys ...string) (int64, error) {
	if len(data) == 0 {
		return 0, EmptyJSONError
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return 0, KeyNotFoundError
	}

	// Check if the value is a number
	if result.Type != gjson.Number {
		return 0, fmt.Errorf("value is not a number: %s", path)
	}

	return result.Int(), nil
}

// GetFloat retrieves a float value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetFloat(data []byte, keys ...string) (float64, error) {
	if len(data) == 0 {
		return 0, EmptyJSONError
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return 0, KeyNotFoundError
	}

	// Check if the value is a number
	if result.Type != gjson.Number {
		return 0, fmt.Errorf("value is not a number: %s", path)
	}

	return result.Float(), nil
}

// GetString retrieves a string value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetString(data []byte, keys ...string) (string, error) {
	if len(data) == 0 {
		return "", EmptyJSONError
	}

	// Join keys with dots to create a path
	path := getPath(keys...)

	// Use gjson to get the value
	result := gjson.GetBytes(data, path)
	if !result.Exists() {
		return "", KeyNotFoundError
	}

	// Check if the value is a string
	if result.Type != gjson.String {
		return "", fmt.Errorf("value is not a string: %s", path)
	}

	return result.String(), nil
}

// SetValue sets the value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetValue(data []byte, value interface{}, keys ...string) ([]byte, error) {
	if len(data) == 0 {
		// If data is empty, create a new JSON object
		data = []byte("{}")
	}

	if len(keys) == 0 {
		return nil, NoKeysProvidedError
	}

	if keys[0] == "" {
		return nil, EmptyKeyError
	}

	// Join keys with dots to create a path
	path := getPath(keys...)
	if path == "" {
		return nil, EmptyKeyError
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
		return nil, EmptyJSONError
	}

	if len(keys) == 0 {
		return nil, NoKeysProvidedError
	}

	if keys[0] == "" {
		return nil, EmptyKeyError
	}

	// Join keys with dots to create a path
	path := getPath(keys...)
	if path == "" {
		return nil, EmptyKeyError
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
