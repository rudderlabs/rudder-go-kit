package jsonparser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/grafana/jsonparser"
)

// jsonparserJSONParser is the implementation of JSONParser using jsonparser library
type jsonparserJSONParser struct{}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func getKeys(key string) []string {
	keys := []string{key}
	if strings.Contains(key, ".") {
		keys = strings.Split(key, ".")
		if len(keys) > 1 {
			for i, k := range keys {
				if isNumeric(k) {
					keys[i] = "[" + k + "]"
				}
			}
		}
	}
	return keys
}

// GetValue retrieves the value for a given key from JSON bytes using jsonparser
func (p *jsonparserJSONParser) GetValue(data []byte, key string) (interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	// Handle empty key - return the entire JSON object
	if key == "" {
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return result, nil
	}

	keys := getKeys(key)
	value, dataType, _, err := jsonparser.Get(data, keys...)
	if err != nil {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	// Convert the value to the appropriate Go type based on the data type
	switch dataType {
	case jsonparser.String:
		return string(value), nil
	case jsonparser.Number:
		// Try to parse as float64
		f, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse number: %w", err)
		}
		return f, nil
	case jsonparser.Boolean:
		return string(value) == "true", nil
	case jsonparser.Null:
		return nil, nil
	case jsonparser.Array, jsonparser.Object:
		var result interface{}
		if err := json.Unmarshal(value, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal complex value: %w", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown value type")
	}
}

// GetBoolean retrieves a boolean value for a given key from JSON bytes
func (p *jsonparserJSONParser) GetBoolean(data []byte, key string) (bool, error) {
	if len(data) == 0 {
		return false, fmt.Errorf("empty JSON data")
	}

	keys := getKeys(key)
	value, dataType, _, err := jsonparser.Get(data, keys...)
	if err != nil {
		return false, fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a boolean
	if dataType != jsonparser.Boolean {
		return false, fmt.Errorf("value is not a boolean: %s", key)
	}

	return string(value) == "true", nil
}

// GetInt retrieves an integer value for a given key from JSON bytes
func (p *jsonparserJSONParser) GetInt(data []byte, key string) (int64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty JSON data")
	}

	// Validate JSON
	var dummy interface{}
	if err := json.Unmarshal(data, &dummy); err != nil {
		return 0, fmt.Errorf("invalid JSON data")
	}

	keys := getKeys(key)
	value, dataType, _, err := jsonparser.Get(data, keys...)
	if err != nil {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a number
	if dataType != jsonparser.Number {
		return 0, fmt.Errorf("value is not a number: %s", key)
	}

	// Parse the number as int64
	i, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse integer: %w", err)
	}

	return i, nil
}

// GetFloat retrieves a float value for a given key from JSON bytes
func (p *jsonparserJSONParser) GetFloat(data []byte, key string) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty JSON data")
	}

	// Validate JSON
	var dummy interface{}
	if err := json.Unmarshal(data, &dummy); err != nil {
		return 0, fmt.Errorf("invalid JSON data")
	}

	keys := getKeys(key)
	value, dataType, _, err := jsonparser.Get(data, keys...)
	if err != nil {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a number
	if dataType != jsonparser.Number {
		return 0, fmt.Errorf("value is not a number: %s", key)
	}

	// Parse the number as float64
	f, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float: %w", err)
	}

	return f, nil
}

// GetString retrieves a string value for a given key from JSON bytes
func (p *jsonparserJSONParser) GetString(data []byte, key string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty JSON data")
	}

	// Validate JSON
	var dummy interface{}
	if err := json.Unmarshal(data, &dummy); err != nil {
		return "", fmt.Errorf("invalid JSON data")
	}

	keys := getKeys(key)
	value, dataType, _, err := jsonparser.Get(data, keys...)
	if err != nil {
		return "", fmt.Errorf("key not found: %s", key)
	}

	// Check if the value is a string
	if dataType != jsonparser.String {
		return "", fmt.Errorf("value is not a string: %s", key)
	}

	return string(value), nil
}

// SetValue sets the value for a given key in JSON bytes using jsonparser
func (p *jsonparserJSONParser) SetValue(data []byte, key string, value interface{}) ([]byte, error) {
	if len(data) == 0 {
		// If data is empty, create a new JSON object
		data = []byte("{}")
	}

	if key == "" {
		return nil, fmt.Errorf("empty key")
	}

	valueBytes, err := jsonrs.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	keys := getKeys(key)
	resultData, err := jsonparser.Set(data, valueBytes, keys...)
	if err != nil {
		return nil, fmt.Errorf("failed to set value: %w", err)
	}

	return resultData, nil
}

// SetBoolean sets a boolean value for a given key in JSON bytes
func (p *jsonparserJSONParser) SetBoolean(data []byte, key string, value bool) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetInt sets an integer value for a given key in JSON bytes
func (p *jsonparserJSONParser) SetInt(data []byte, key string, value int64) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetFloat sets a float value for a given key in JSON bytes
func (p *jsonparserJSONParser) SetFloat(data []byte, key string, value float64) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetString sets a string value for a given key in JSON bytes
func (p *jsonparserJSONParser) SetString(data []byte, key string, value string) ([]byte, error) {
	return p.SetValue(data, key, value)
}
