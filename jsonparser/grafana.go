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

// GetValue retrieves the value for a given key from JSON bytes using jsonparser
func (p *grafanaJSONParser) GetValue(data []byte, keys ...string) ([]byte, error) {
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

	value, dtype, _, err := jsonparser.Get(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}

	// If the value is a string, wrap it in quotes as the library is specifically stripping quotes from string value
	if dtype == jsonparser.String {
		return []byte("\"" + string(value) + "\""), nil
	}

	return value, nil
}

// GetBoolean retrieves a boolean value for a given key from JSON bytes
func (p *grafanaJSONParser) GetBoolean(data []byte, keys ...string) (bool, error) {
	if len(data) == 0 {
		return false, ErrEmptyJSON
	}

	value, err := jsonparser.GetBoolean(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return false, ErrKeyNotFound
		}
		return false, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}
	return value, nil
}

// GetInt retrieves an integer value for a given key from JSON bytes
func (p *grafanaJSONParser) GetInt(data []byte, keys ...string) (int64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	floatVal, err := jsonparser.GetFloat(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, ErrKeyNotFound
		}
		return 0, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}
	value := int64(floatVal)

	return value, nil
}

// GetFloat retrieves a float value for a given key from JSON bytes
func (p *grafanaJSONParser) GetFloat(data []byte, keys ...string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	value, err := jsonparser.GetFloat(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, ErrKeyNotFound
		}
		return 0, fmt.Errorf("failed to get value for keys %v: %w", keys, err)
	}

	return value, nil
}

// GetString retrieves a string value for a given key from JSON bytes
func (p *grafanaJSONParser) GetString(data []byte, keys ...string) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyJSON
	}

	value, err := jsonparser.GetString(data, keys...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return "", ErrKeyNotFound
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
		return nil, ErrNoKeysProvided
	}

	key := keys[0]
	if key == "" {
		return nil, ErrEmptyKey
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
		return nil, ErrEmptyJSON
	}

	if len(keys) == 0 {
		return nil, ErrNoKeysProvided
	}

	key := keys[0]
	if key == "" {
		return nil, ErrEmptyKey
	}

	// Use jsonparser.Delete to delete the key
	resultData := jsonparser.Delete(data, keys...)

	return resultData, nil
}
