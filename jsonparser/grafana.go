package jsonparser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/grafana/jsonparser"
)

// grafanaJSONParser is the implementation of JSONParser using jsonparser library
type grafanaJSONParser struct{}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func getKeys(key string, data []byte) ([]string, error) {
	keys := []string{key}
	if strings.Contains(key, ".") {
		keys = strings.Split(key, ".")
	}
	keysTillNow := make([]string, 0, len(keys))
	for _, k := range keys {
		if isNumeric(k) {
			_, dataType, _, err := jsonparser.Get(data, keysTillNow...)
			if err != nil {
				return nil, fmt.Errorf("key not found: %s", key)
			}
			if dataType == jsonparser.Array {
				keysTillNow = append(keysTillNow, "["+k+"]")
			} else {
				keysTillNow = append(keysTillNow, k)
			}
		} else {
			keysTillNow = append(keysTillNow, k)
		}
	}
	return keysTillNow, nil
}

// GetValue retrieves the value for a given key from JSON bytes using jsonparser
func (p *grafanaJSONParser) GetValue(data []byte, key string) (interface{}, error) {
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

	keys, err := getKeys(key, data)
	if err != nil {
		return nil, err
	}

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
func (p *grafanaJSONParser) GetBoolean(data []byte, key string) (bool, error) {
	if len(data) == 0 {
		return false, fmt.Errorf("empty JSON data")
	}

	keys, err := getKeys(key, data)
	if err != nil {
		return false, err
	}
	value, err := jsonparser.GetBoolean(data, keys...)
	if err != nil {
		return false, fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

// GetInt retrieves an integer value for a given key from JSON bytes
func (p *grafanaJSONParser) GetInt(data []byte, key string) (int64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty JSON data")
	}

	// Validate JSON
	var dummy interface{}
	if err := json.Unmarshal(data, &dummy); err != nil {
		return 0, fmt.Errorf("invalid JSON data")
	}

	keys, err := getKeys(key, data)
	if err != nil {
		return 0, err
	}
	value, err := jsonparser.GetInt(data, keys...)
	if err != nil {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// GetFloat retrieves a float value for a given key from JSON bytes
func (p *grafanaJSONParser) GetFloat(data []byte, key string) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("empty JSON data")
	}

	keys, err := getKeys(key, data)
	if err != nil {
		return 0, err
	}
	value, err := jsonparser.GetFloat(data, keys...)
	if err != nil {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// GetString retrieves a string value for a given key from JSON bytes
func (p *grafanaJSONParser) GetString(data []byte, key string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty JSON data")
	}

	// Validate JSON
	var dummy interface{}
	if err := json.Unmarshal(data, &dummy); err != nil {
		return "", fmt.Errorf("invalid JSON data")
	}

	keys, err := getKeys(key, data)
	if err != nil {
		return "", err
	}
	value, err := jsonparser.GetString(data, keys...)
	if err != nil {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// SetValue sets the value for a given key in JSON bytes using jsonparser
func (p *grafanaJSONParser) SetValue(data []byte, key string, value interface{}) ([]byte, error) {
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

	keys, err := getKeys(key, data)
	if err != nil {
		return nil, err
	}
	resultData, err := jsonparser.Set(data, valueBytes, keys...)
	if err != nil {
		return nil, fmt.Errorf("failed to set value: %w", err)
	}

	return resultData, nil
}

// SetBoolean sets a boolean value for a given key in JSON bytes
func (p *grafanaJSONParser) SetBoolean(data []byte, key string, value bool) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetInt sets an integer value for a given key in JSON bytes
func (p *grafanaJSONParser) SetInt(data []byte, key string, value int64) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetFloat sets a float value for a given key in JSON bytes
func (p *grafanaJSONParser) SetFloat(data []byte, key string, value float64) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// SetString sets a string value for a given key in JSON bytes
func (p *grafanaJSONParser) SetString(data []byte, key, value string) ([]byte, error) {
	return p.SetValue(data, key, value)
}

// DeleteKey deletes a key from JSON bytes
func (p *grafanaJSONParser) DeleteKey(data []byte, key string) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty JSON data")
	}

	if key == "" {
		return nil, fmt.Errorf("empty key")
	}

	keys, err := getKeys(key, data)
	if err != nil {
		return nil, err
	}

	// Use jsonparser.Delete to delete the key
	resultData := jsonparser.Delete(data, keys...)

	return resultData, nil
}
