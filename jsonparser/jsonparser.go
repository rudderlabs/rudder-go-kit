package jsonparser

import (
	"errors"

	"github.com/rudderlabs/rudder-go-kit/config"
)

// Getter is the interface that wraps the basic operations for retrieving values from JSON bytes.
type Getter interface {
	// GetValue retrieves the raw value for a given key from JSON bytes.
	// Accept multiple keys to specify path to JSON value (in case of quering nested structures).
	// e.g. `GetValue(data, "key1", "key2")` retrieves the value at `data.key1.key2`.
	//      `GetValue(data, "key1", "[0]")` retrieves the value at `data.key1[0]`.
	//      `GetValue({"key":"val"}, "key1")` returns `ErrKeyNotFound`.
	// Returns the raw value bytes and an error if the key does not exist or if there is a JSON parsing error.
	GetValue(data []byte, path ...string) ([]byte, error)

	// GetBoolean retrieves a boolean value for a given key from JSON bytes.
	// Accept multiple keys to specify path to JSON value (in case of quering nested structures).
	// e.g. `GetBoolean(data, "key1", "key2")` retrieves the value at `data.key1.key2`.
	//      `GetBoolean(data, "key1", "[0]")` retrieves the value at `data.key1[0]`.
	//      `GetBoolean({"key":"val"}, "key1")` returns `ErrKeyNotFound`.
	// GetBoolean expects the value to be a boolean type if the value is not a boolean type it returns `ErrNotOfExpectedType`.
	//      `GetBoolean({"key":"val"}, "key")` returns `ErrNotOfExpectedType`.
	// Returns the boolean value and an error if the key does not exist or if there is a JSON parsing error.
	GetBoolean(data []byte, path ...string) (bool, error)

	// GetInt retrieves an integer value for a given key from JSON bytes.
	// Accept multiple keys to specify path to JSON value (in case of quering nested structures).
	// e.g. `GetInt(data, "key1", "key2")` retrieves the value at `data.key1.key2`.
	//      `GetInt(data, "key1", "[0]")` retrieves the value at `data.key1[0]`.
	//      `GetInt({"key":"val"}, "key1")` returns `ErrKeyNotFound`.
	// GetInt expects the value to be an integer type if the value is not an integer type, it returns `ErrNotOfExpectedType`.
	//      `GetInt({"key":"val"}, "key")` returns `ErrNotOfExpectedType`.
	// Returns the integer value and an error if the key does not exist or if there is a JSON parsing error.
	GetInt(data []byte, path ...string) (int64, error)

	// GetFloat retrieves a float value for a given key from JSON bytes.
	// Accept multiple keys to specify path to JSON value (in case of quering nested structures).
	// e.g. `GetFloat(data, "key1", "key2")` retrieves the value at `data.key1.key2`.
	//      `GetFloat(data, "key1", "[0]")` retrieves the value at `data.key1[0]`.
	//      `GetFloat({"key":"val"}, "key1")` returns `ErrKeyNotFound`.
	// GetFloat expects the value to be a float type if the value is not a float type, it returns `ErrNotOfExpectedType`.
	//      `GetFloat({"key":"val"}, "key")` returns `ErrNotOfExpectedType`.
	// Returns the float value and an error if the key does not exist or if there is a JSON parsing error.
	GetFloat(data []byte, path ...string) (float64, error)

	// GetString retrieves a string value for a given key from JSON bytes.
	// Accept multiple keys to specify path to JSON value (in case of quering nested structures).
	// e.g. `GetString(data, "key1", "key2")` retrieves the value at `data.key1.key2`.
	//      `GetString(data, "key1", "[0]")` retrieves the value at `data.key1[0]`.
	//      `GetString({"key":"val"}, "key1")` returns `ErrKeyNotFound`.
	// GetString expects the value to be a string type if the value is not a string type, it returns `ErrNotOfExpectedType`.
	//      `GetString({"key": 0}, "key")` returns `ErrNotOfExpectedType`.
	// Returns the string value and an error if the key does not exist or if there is a JSON parsing error.
	GetString(data []byte, path ...string) (string, error)
}

// Setter is the interface that wraps the basic operations for setting values in JSON bytes.
type Setter interface {
	// SetValue sets the value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetValue(data []byte, value interface{}, path ...string) ([]byte, error)

	// SetBoolean sets a boolean value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetBoolean(data []byte, value bool, path ...string) ([]byte, error)

	// SetInt sets an integer value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetInt(data []byte, value int64, path ...string) ([]byte, error)

	// SetFloat sets a float value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetFloat(data []byte, value float64, path ...string) ([]byte, error)

	// SetString sets a string value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetString(data []byte, value string, path ...string) ([]byte, error)
}

// Deleter is the interface that wraps the basic operations for deleting keys from JSON bytes.
type Deleter interface {
	// DeleteKey deletes a key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	DeleteKey(data []byte, path ...string) ([]byte, error)
}

type SoftGetter interface {
	// GetValueOrEmpty retrieves the raw value for a given key from JSON bytes.
	// If the key does not exist or json is invalid, it returns an empty byte slice.
	GetValueOrEmpty(data []byte, path ...string) []byte

	// GetBooleanOrFalse retrieves a boolean value for a given path from JSON bytes.
	// If the key does not exist, json is invalid or value is not parsable to a boolean, it returns false.
	// e.g `GetBooleanOrFalse({"key": true}, "key1")` returns false.
	//	   `GetBooleanOrFalse({"key": true}, "key")` returns true.
	//     `GetBooleanOrFalse({"key": "val"}, "key")` returns false as `val` is not parsable to boolean.
	//     `GetBooleanOrFalse({"key": 1}, "key")` returns true.
	GetBooleanOrFalse(data []byte, path ...string) bool

	// GetIntOrZero retrieves an integer value for a given key from JSON bytes.
	// If the key does not exist, json is invalid or value is not parsable to an integer, it returns 0.
	// e.g `GetIntOrZero({"key": 123}, "key1")` returns 0.
	//     `GetIntOrZero({"key": 123}, "key")` returns 123.
	//     `GetIntOrZero({"key": "val"}, "key")` returns 0 as `val` is not parsable to integer.
	//     `GetIntOrZero({"key": true}, "key")` returns 1 as `true` is parsed to 1.
	//     `GetIntOrZero({"key": "123"}, "key")` returns 123 as `"123"` is parsed to 123.
	GetIntOrZero(data []byte, path ...string) int64

	// GetFloatOrZero retrieves a float value for a given key from JSON bytes.
	// If the key does not exist, json is invalid or value is not parsable to a float, it returns 0.
	// e.g `GetFloatOrZero({"key": 123.456}, "key1") returns 0.
	//     `GetFloatOrZero({"key": 123.456}, "key")` returns 123.456.
	//     `GetFloatOrZero({"key": "123.456"}, "key")` returns 123.456 as `"123.456"` is parsed to 123.456.
	//     `GetFloatOrZero({"key": "val"}, "key")` returns 0 as `val` is not parsable to float.
	//     `GetFloatOrZero({"key": true}, "key")` returns 1 as `true` is parsed to 1.
	GetFloatOrZero(data []byte, path ...string) float64

	// GetStringOrEmpty retrieves a string value for a given key from JSON bytes.
	// If the key does not exist, json is invalid or value is not parsable to a string, it returns an empty string.
	// e.g `GetStringOrEmpty({"key": "val"}, "key1")` returns "".
	//     `GetStringOrEmpty({"key": "val"}, "key")` returns "val".
	//     `GetStringOrEmpty({"key": 123}, "key")` returns "123" as 123 is parsed to string.
	//     `GetStringOrEmpty({"key": true}, "key")` returns "true" as `true` is parsed to string.
	GetStringOrEmpty(data []byte, path ...string) string
}

// JSONParser is the interface that combines Getter, Setter, and Deleter interfaces.
type JSONParser interface {
	Getter
	Setter
	Deleter
	SoftGetter
}

// Default is the default JSONParser implementation.
var Default = NewWithConfig(config.Default)

// GetValue is a convenience function that uses the default JSONParser.
func GetValue(data []byte, path ...string) (interface{}, error) {
	return Default.GetValue(data, path...)
}

// GetBoolean is a convenience function that uses the default JSONParser.
func GetBoolean(data []byte, path ...string) (bool, error) {
	return Default.GetBoolean(data, path...)
}

// GetInt is a convenience function that uses the default JSONParser.
func GetInt(data []byte, path ...string) (int64, error) {
	return Default.GetInt(data, path...)
}

// GetFloat is a convenience function that uses the default JSONParser.
func GetFloat(data []byte, path ...string) (float64, error) {
	return Default.GetFloat(data, path...)
}

// GetString is a convenience function that uses the default JSONParser.
func GetString(data []byte, path ...string) (string, error) {
	return Default.GetString(data, path...)
}

// SetValue is a convenience function that uses the default JSONParser.
func SetValue(data []byte, value interface{}, path ...string) ([]byte, error) {
	return Default.SetValue(data, value, path...)
}

// SetBoolean is a convenience function that uses the default JSONParser.
func SetBoolean(data []byte, value bool, path ...string) ([]byte, error) {
	return Default.SetBoolean(data, value, path...)
}

// SetInt is a convenience function that uses the default JSONParser.
func SetInt(data []byte, value int64, path ...string) ([]byte, error) {
	return Default.SetInt(data, value, path...)
}

// SetFloat is a convenience function that uses the default JSONParser.
func SetFloat(data []byte, value float64, path ...string) ([]byte, error) {
	return Default.SetFloat(data, value, path...)
}

// SetString is a convenience function that uses the default JSONParser.
func SetString(data []byte, value string, path ...string) ([]byte, error) {
	return Default.SetString(data, value, path...)
}

// DeleteKey is a convenience function that uses the default JSONParser.
func DeleteKey(data []byte, path ...string) ([]byte, error) {
	return Default.DeleteKey(data, path...)
}

// Reset resets the default JSONParser implementation based on the default configuration.
func Reset() {
	Default = NewWithConfig(config.Default)
}

var (
	ErrKeyNotFound       = errors.New("key not found in JSON data")
	ErrEmptyJSON         = errors.New("empty JSON data provided")
	ErrNoKeysProvided    = errors.New("no keys provided")
	ErrEmptyKey          = errors.New("empty key provided")
	ErrNotOfExpectedType = errors.New("value is not of expected type")
)
