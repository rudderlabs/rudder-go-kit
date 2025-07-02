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
func (p *grafanaJSONParser) GetValue(data []byte, path ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyJSON
	}

	if len(path) == 0 {
		return nil, ErrNoKeysProvided
	}

	key := path[0]
	if key == "" {
		return nil, ErrEmptyKey
	}

	value, dtype, _, err := jsonparser.Get(data, path...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get value for path %v: %w", path, err)
	}

	// If the value is a string, wrap it in quotes as the library is specifically stripping quotes from string value
	if dtype == jsonparser.String {
		return []byte("\"" + string(value) + "\""), nil
	}

	return value, nil
}

// GetValueOrEmpty retrieves the raw value for a given key from JSON bytes.
// If the key does not exist, json is invalid or value is not a string, it returns an empty byte slice.
func (p *grafanaJSONParser) GetValueOrEmpty(data []byte, path ...string) []byte {
	if len(data) == 0 || len(path) == 0 {
		return nil
	}

	key := path[0]
	if key == "" {
		return nil
	}

	value, dtype, _, err := jsonparser.Get(data, path...)
	if err != nil {
		return nil
	}

	// If the value is a string, wrap it in quotes as the library is specifically stripping quotes from string value
	if dtype == jsonparser.String {
		return []byte("\"" + string(value) + "\"")
	}

	return value
}

// GetBoolean retrieves a boolean value for a given key from JSON bytes
func (p *grafanaJSONParser) GetBoolean(data []byte, path ...string) (bool, error) {
	if len(data) == 0 {
		return false, ErrEmptyJSON
	}

	value, err := jsonparser.GetBoolean(data, path...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return false, ErrKeyNotFound
		}
		return false, fmt.Errorf("failed to get value for path %v, %w", path, err)
	}
	return value, nil
}

// GetBooleanOrFalse retrieves a boolean value for a given path from JSON bytes.
// If the key does not exist, json is invalid or value is not a boolean, it returns false.
func (p *grafanaJSONParser) GetBooleanOrFalse(data []byte, path ...string) bool {
	if len(data) == 0 || len(path) == 0 {
		return false
	}

	value, dtype, _, err := jsonparser.Get(data, path...)
	if err != nil {
		return false
	}

	boolVal, _ := parseBool(value, dtype)

	return boolVal
}

// GetInt retrieves an integer value for a given key from JSON bytes
func (p *grafanaJSONParser) GetInt(data []byte, path ...string) (int64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	floatVal, err := jsonparser.GetFloat(data, path...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, ErrKeyNotFound
		}
		return 0, fmt.Errorf("failed to get value for path %v, %w", path, err)
	}
	value := int64(floatVal)

	return value, nil
}

// GetIntOrZero retrieves an integer value for a given key from JSON bytes.
// If the key does not exist, json is invalid or value is not an integer, it returns 0.
func (p *grafanaJSONParser) GetIntOrZero(data []byte, path ...string) int64 {
	if len(data) == 0 || len(path) == 0 {
		return 0
	}

	val, dtype, _, err := jsonparser.Get(data, path...)
	if err != nil {
		return 0
	}
	intVal, _ := parseInt(val, dtype)

	return intVal
}

// GetFloat retrieves a float value for a given key from JSON bytes
func (p *grafanaJSONParser) GetFloat(data []byte, path ...string) (float64, error) {
	if len(data) == 0 {
		return 0, ErrEmptyJSON
	}

	value, err := jsonparser.GetFloat(data, path...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return 0, ErrKeyNotFound
		}
		return 0, fmt.Errorf("failed to get value for path %v, %w", path, err)
	}

	return value, nil
}

// GetFloatOrZero retrieves a float value for a given key from JSON bytes.
// If the key does not exist, json is invalid or value is not a float, it returns 0.
func (p *grafanaJSONParser) GetFloatOrZero(data []byte, path ...string) float64 {
	if len(data) == 0 || len(path) == 0 {
		return 0
	}

	val, dtype, _, err := jsonparser.Get(data, path...)
	if err != nil {
		return 0
	}
	value, _ := parseFloat(val, dtype)

	return value
}

// GetString retrieves a string value for a given key from JSON bytes
func (p *grafanaJSONParser) GetString(data []byte, path ...string) (string, error) {
	if len(data) == 0 {
		return "", ErrEmptyJSON
	}

	value, err := jsonparser.GetString(data, path...)
	if err != nil {
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("failed to get value for path %v, %w", path, err)
	}

	return value, nil
}

// GetStringOrEmpty retrieves a string value for a given key from JSON bytes.
// If the key does not exist, json is invalid or value is not a string, it returns an empty string.
func (p *grafanaJSONParser) GetStringOrEmpty(data []byte, path ...string) string {
	if len(data) == 0 || len(path) == 0 {
		return ""
	}

	val, dtype, _, err := jsonparser.Get(data, path...)
	if err != nil {
		return ""
	}

	value, _ := parseString(val, dtype)
	return value
}

// SetValue sets the value for a given key in JSON bytes using jsonparser
func (p *grafanaJSONParser) SetValue(data []byte, value interface{}, path ...string) ([]byte, error) {
	if len(data) == 0 {
		// If data is empty, create a new JSON object
		data = []byte("{}")
	}

	if len(path) == 0 {
		return nil, ErrNoKeysProvided
	}

	key := path[0]
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

	resultData, err := jsonparser.Set(data, valueBytes, path...)
	if err != nil {
		return nil, fmt.Errorf("failed to set value: %w", err)
	}

	return resultData, nil
}

// SetBoolean sets a boolean value for a given key in JSON bytes
func (p *grafanaJSONParser) SetBoolean(data []byte, value bool, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// SetInt sets an integer value for a given key in JSON bytes
func (p *grafanaJSONParser) SetInt(data []byte, value int64, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// SetFloat sets a float value for a given key in JSON bytes
func (p *grafanaJSONParser) SetFloat(data []byte, value float64, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// SetString sets a string value for a given key in JSON bytes
func (p *grafanaJSONParser) SetString(data []byte, value string, path ...string) ([]byte, error) {
	return p.SetValue(data, value, path...)
}

// DeleteKey deletes a key from JSON bytes
func (p *grafanaJSONParser) DeleteKey(data []byte, path ...string) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrEmptyJSON
	}

	if len(path) == 0 {
		return nil, ErrNoKeysProvided
	}

	key := path[0]
	if key == "" {
		return nil, ErrEmptyKey
	}

	// Use jsonparser.Delete to delete the key
	resultData := jsonparser.Delete(data, path...)

	return resultData, nil
}

func parseBool(value []byte, dataType jsonparser.ValueType) (bool, error) {
	switch dataType {
	default:
		return false, nil
	case jsonparser.Boolean, jsonparser.String:
		return jsonparser.ParseBoolean(value)
	case jsonparser.Number:
		num, err := jsonparser.ParseFloat(value)
		return num != 0, err
	}
}

func parseInt(value []byte, dataType jsonparser.ValueType) (int64, error) {
	switch dataType {
	default:
		return 0, nil
	case jsonparser.Number:
		return jsonparser.ParseInt(value)
	case jsonparser.Boolean:
		boolVal, err := jsonparser.ParseBoolean(value)
		if err != nil {
			return 0, err
		}
		if boolVal {
			return 1, nil
		}
		return 0, nil
	case jsonparser.String:
		num, err := jsonparser.ParseFloat(value)
		return int64(num), err
	}
}

func parseFloat(val []byte, dtype jsonparser.ValueType) (float64, error) {
	switch dtype {
	default:
		return 0, nil
	case jsonparser.Number:
		return jsonparser.ParseFloat(val)
	case jsonparser.Boolean:
		boolVal, err := jsonparser.ParseBoolean(val)
		if err != nil {
			return 0, err
		}
		if boolVal {
			return 1, nil
		}
		return 0, nil
	case jsonparser.String:
		num, err := jsonparser.ParseFloat(val)
		return num, err
	}
}

func parseString(val []byte, dtype jsonparser.ValueType) (string, error) {
	switch dtype {
	default:
		return "", nil
	case jsonparser.String:
		return jsonparser.ParseString(val)
	case jsonparser.Boolean:
		boolVal, err := jsonparser.ParseBoolean(val)
		if err != nil {
			return "false", nil
		}
		return strconv.FormatBool(boolVal), nil
	case jsonparser.Number:
		num, err := jsonparser.ParseFloat(val)
		if err != nil {
			return "0", nil
		}
		return strconv.FormatFloat(num, 'f', -1, 64), nil
	}
}
