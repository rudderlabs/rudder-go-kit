package jsonparser

import (
	"errors"

	"github.com/rudderlabs/rudder-go-kit/config"
)

// getter is the interface that wraps the basic operations for retrieving values from JSON bytes.
type getter interface {
	// GetValue retrieves the value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the value as an interface{} and an error if the key does not exist or if there is a JSON parsing error.
	GetValue(data []byte, keys ...string) (interface{}, error)

	// GetBoolean retrieves a boolean value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the boolean value and an error if the key does not exist or if there is a JSON parsing error.
	GetBoolean(data []byte, keys ...string) (bool, error)

	// GetInt retrieves an integer value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the integer value and an error if the key does not exist or if there is a JSON parsing error.
	GetInt(data []byte, keys ...string) (int64, error)

	// GetFloat retrieves a float value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the float value and an error if the key does not exist or if there is a JSON parsing error.
	GetFloat(data []byte, keys ...string) (float64, error)

	// GetString retrieves a string value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the string value and an error if the key does not exist or if there is a JSON parsing error.
	GetString(data []byte, keys ...string) (string, error)
}

// setter is the interface that wraps the basic operations for setting values in JSON bytes.
type setter interface {
	// SetValue sets the value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetValue(data []byte, value interface{}, keys ...string) ([]byte, error)

	// SetBoolean sets a boolean value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetBoolean(data []byte, value bool, keys ...string) ([]byte, error)

	// SetInt sets an integer value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetInt(data []byte, value int64, keys ...string) ([]byte, error)

	// SetFloat sets a float value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetFloat(data []byte, value float64, keys ...string) ([]byte, error)

	// SetString sets a string value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	SetString(data []byte, value string, keys ...string) ([]byte, error)
}

// deleter is the interface that wraps the basic operations for deleting keys from JSON bytes.
type deleter interface {
	// DeleteKey deletes a key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes, and an error if no/empty key is passed, or if there is a JSON parsing error.
	DeleteKey(data []byte, keys ...string) ([]byte, error)
}

// JSONParser is the interface that combines getter, setter, and deleter interfaces.
type JSONParser interface {
	getter
	setter
	deleter
}

// Default is the default JSONParser implementation.
var Default = NewWithConfig(config.Default)

// GetValue is a convenience function that uses the default JSONParser.
func GetValue(data []byte, keys ...string) (interface{}, error) {
	return Default.GetValue(data, keys...)
}

// GetBoolean is a convenience function that uses the default JSONParser.
func GetBoolean(data []byte, keys ...string) (bool, error) {
	return Default.GetBoolean(data, keys...)
}

// GetInt is a convenience function that uses the default JSONParser.
func GetInt(data []byte, keys ...string) (int64, error) {
	return Default.GetInt(data, keys...)
}

// GetFloat is a convenience function that uses the default JSONParser.
func GetFloat(data []byte, keys ...string) (float64, error) {
	return Default.GetFloat(data, keys...)
}

// GetString is a convenience function that uses the default JSONParser.
func GetString(data []byte, keys ...string) (string, error) {
	return Default.GetString(data, keys...)
}

// SetValue is a convenience function that uses the default JSONParser.
func SetValue(data []byte, value interface{}, keys ...string) ([]byte, error) {
	return Default.SetValue(data, value, keys...)
}

// SetBoolean is a convenience function that uses the default JSONParser.
func SetBoolean(data []byte, value bool, keys ...string) ([]byte, error) {
	return Default.SetBoolean(data, value, keys...)
}

// SetInt is a convenience function that uses the default JSONParser.
func SetInt(data []byte, value int64, keys ...string) ([]byte, error) {
	return Default.SetInt(data, value, keys...)
}

// SetFloat is a convenience function that uses the default JSONParser.
func SetFloat(data []byte, value float64, keys ...string) ([]byte, error) {
	return Default.SetFloat(data, value, keys...)
}

// SetString is a convenience function that uses the default JSONParser.
func SetString(data []byte, value string, keys ...string) ([]byte, error) {
	return Default.SetString(data, value, keys...)
}

// DeleteKey is a convenience function that uses the default JSONParser.
func DeleteKey(data []byte, keys ...string) ([]byte, error) {
	return Default.DeleteKey(data, keys...)
}

// Reset resets the default JSONParser implementation based on the default configuration.
func Reset() {
	Default = NewWithConfig(config.Default)
}

var (
	ErrKeyNotFound    = errors.New("key not found in JSON data")
	ErrEmptyJSON      = errors.New("empty JSON data provided")
	ErrNoKeysProvided = errors.New("no keys provided")
	ErrEmptyKey       = errors.New("empty key provided")
)
