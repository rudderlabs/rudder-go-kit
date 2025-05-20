package jsonparser

import (
	"encoding/json"
	"fmt"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// tidwallJSONParser is the implementation of JSONParser using gjson/sjson libraries
type tidwallJSONParser struct{}

// GetValue retrieves the value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetValue(data []byte, key string) (interface{}, error) {
	// Handle empty key - return the entire JSON object
	if key == "" {
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return result, nil
	}

	// Use gjson to get the value
	result := gjson.GetBytes(data, key)
	if !result.Exists() {
		return nil, fmt.Errorf("key not found: %s", key)
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
		//nolint(nilnil)
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
func (p *tidwallJSONParser) GetBoolean(data []byte, key string) (bool, error) {
	// Use gjson to get the value
	result := gjson.GetBytes(data, key)
	if !result.Exists() {
		return false, fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a boolean
	if result.Type != gjson.True && result.Type != gjson.False {
		return false, fmt.Errorf("value is not a boolean: %s", key)
	}

	return result.Bool(), nil
}

// GetInt retrieves an integer value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetInt(data []byte, key string) (int64, error) {
	// Use gjson to get the value
	result := gjson.GetBytes(data, key)
	if !result.Exists() {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a number
	if result.Type != gjson.Number {
		return 0, fmt.Errorf("value is not a number: %s", key)
	}

	return result.Int(), nil
}

// GetFloat retrieves a float value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetFloat(data []byte, key string) (float64, error) {
	// Use gjson to get the value
	result := gjson.GetBytes(data, key)
	if !result.Exists() {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a number
	if result.Type != gjson.Number {
		return 0, fmt.Errorf("value is not a number: %s", key)
	}

	return result.Float(), nil
}

// GetString retrieves a string value for a given key from JSON bytes using gjson
func (p *tidwallJSONParser) GetString(data []byte, key string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty JSON data")
	}

	// Check if the JSON is valid
	if !gjson.ValidBytes(data) {
		return "", fmt.Errorf("invalid JSON data")
	}

	// Use gjson to get the value
	result := gjson.GetBytes(data, key)
	if !result.Exists() {
		return "", fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a string
	if result.Type != gjson.String {
		return "", fmt.Errorf("value is not a string: %s", key)
	}

	return result.String(), nil
}

// SetValue sets the value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetValue(data []byte, key string, value interface{}) ([]byte, error) {
	if len(data) == 0 {
		// If data is empty, create a new JSON object
		data = []byte("{}")
	}

	if key == "" {
		return nil, fmt.Errorf("empty key")
	}

	// Use sjson to set the value
	result, err := sjson.SetBytes(data, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set value: %w", err)
	}

	return result, nil
}

// SetBoolean sets a boolean value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetBoolean(data []byte, key string, value bool) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetInt sets an integer value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetInt(data []byte, key string, value int64) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetFloat sets a float value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetFloat(data []byte, key string, value float64) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetString sets a string value for a given key in JSON bytes using sjson
func (p *tidwallJSONParser) SetString(data []byte, key, value string) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// DeleteKey deletes a key from JSON bytes using sjson
func (p *tidwallJSONParser) DeleteKey(data []byte, key string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	if key == "" {
		return nil, fmt.Errorf("empty key")
	}

	// Use sjson to delete the key
	resultData, err := sjson.DeleteBytes(data, key)
	if err != nil {
		return nil, fmt.Errorf("failed to delete key: %w", err)
	}

	return resultData, nil
}
