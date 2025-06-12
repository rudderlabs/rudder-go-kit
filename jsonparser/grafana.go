package jsonparser

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/grafana/jsonparser"
)

// grafanaJSONParser is the implementation of JSONParser using jsonparser library
type grafanaJSONParser struct{}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// GetValue retrieves the value for a given key from JSON bytes using jsonparser
func (p *grafanaJSONParser) GetValue(data []byte, keys ...string) (interface{}, error) {
	if len(data) == 0 {
		return nil, EmptyJSONError
	}

	// Handle empty key - return the entire JSON object
	if len(keys) == 0 || (len(keys) == 1 && keys[0] == "") {
		var result interface{}
		if err := jsonrs.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		return result, nil
	}

	value, dataType, _, err := jsonparser.Get(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return nil, KeyNotFoundError
		}
		return nil, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
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
		// nolint: nilnil
		return nil, nil
	case jsonparser.Array, jsonparser.Object:
		var result interface{}
		if err := jsonrs.Unmarshal(value, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal complex value: %w", err)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unknown value type")
	}
}

// GetBoolean retrieves a boolean value for a given key from JSON bytes
func (p *grafanaJSONParser) GetBoolean(data []byte, keys ...string) (bool, error) {
	if len(data) == 0 {
		return false, EmptyJSONError
	}

	value, err := jsonparser.GetBoolean(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return false, KeyNotFoundError
		}
		return false, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}
	return value, nil
}

// GetInt retrieves an integer value for a given key from JSON bytes
func (p *grafanaJSONParser) GetInt(data []byte, keys ...string) (int64, error) {
	if len(data) == 0 {
		return 0, EmptyJSONError
	}

	value, err := jsonparser.GetInt(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, KeyNotFoundError
		}
		return 0, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}

	return value, nil
}

// GetFloat retrieves a float value for a given key from JSON bytes
func (p *grafanaJSONParser) GetFloat(data []byte, keys ...string) (float64, error) {
	if len(data) == 0 {
		return 0, EmptyJSONError
	}

	value, err := jsonparser.GetFloat(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, KeyNotFoundError
		}
		return 0, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}

	return value, nil
}

// GetString retrieves a string value for a given key from JSON bytes
func (p *grafanaJSONParser) GetString(data []byte, keys ...string) (string, error) {
	if len(data) == 0 {
		return "", EmptyJSONError
	}

	value, err := jsonparser.GetString(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return "", KeyNotFoundError
		}
		return "", fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}

	return value, nil
}

// SetValue sets the value for a given key in JSON bytes using jsonparser
func (p *grafanaJSONParser) SetValue(data []byte, value interface{}, keys ...string) ([]byte, error) {

	if len(data) == 0 {
		// If data is empty, create a new JSON object
		data = []byte("{}")
	}

	if len(keys) == 0 {
		return nil, NoKeysProvidedError
	}

	key := keys[0]
	if key == "" {
		return nil, EmptyKeyError
	}

	var valueBytes []byte
	var err error

	switch v := value.(type) {
	default:
		valueBytes, err = jsonrs.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value: %w", err)
		}
	case string:
		valueBytes = []byte("\"" + v + "\"")
	case bool:
		if v {
			valueBytes = []byte("true")
		} else {
			valueBytes = []byte("false")
		}
	case int64:
		valueBytes = []byte(strconv.FormatInt(v, 10))
	case float64:
		valueBytes = []byte(strconv.FormatFloat(v, 'f', -1, 64))
	}

	resultData, err := jsonparser.Set(data, valueBytes, keys...)
	if err != nil {
		return nil, fmt.Errorf("failed to set value: %w", err)
	}

	return resultData, nil
}

// SetBoolean sets a boolean value for a given key in JSON bytes
func (p *grafanaJSONParser) SetBoolean(data []byte, value bool, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// SetInt sets an integer value for a given key in JSON bytes
func (p *grafanaJSONParser) SetInt(data []byte, value int64, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// SetFloat sets a float value for a given key in JSON bytes
func (p *grafanaJSONParser) SetFloat(data []byte, value float64, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// SetString sets a string value for a given key in JSON bytes
func (p *grafanaJSONParser) SetString(data []byte, value string, keys ...string) ([]byte, error) {
	return p.SetValue(data, value, keys...)
}

// DeleteKey deletes a key from JSON bytes
func (p *grafanaJSONParser) DeleteKey(data []byte, keys ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, EmptyJSONError
	}

	if len(keys) == 0 {
		return nil, NoKeysProvidedError
	}

	key := keys[0]
	if key == "" {
		return nil, EmptyKeyError
	}

	// Use jsonparser.Delete to delete the key
	resultData := jsonparser.Delete(data, keys...)

	return resultData, nil
}
