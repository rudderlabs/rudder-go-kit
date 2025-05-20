package jsonparser

import (
	"github.com/rudderlabs/rudder-go-kit/config"
)

// JSONGetter is the interface that wraps the basic operations for retrieving values from JSON bytes.
type JSONGetter interface {
	// GetValue retrieves the value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	GetValue(data []byte, key string) (interface{}, error)

	// GetBoolean retrieves a boolean value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	GetBoolean(data []byte, key string) (bool, error)

	// GetInt retrieves an integer value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	GetInt(data []byte, key string) (int64, error)

	// GetFloat retrieves a float value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	GetFloat(data []byte, key string) (float64, error)

	// GetString retrieves a string value for a given key from JSON bytes.
	// The key can be a dot-separated path to access nested values.
	GetString(data []byte, key string) (string, error)
}

// JSONSetter is the interface that wraps the basic operations for setting values in JSON bytes.
type JSONSetter interface {
	// SetValue sets the value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes.
	SetValue(data []byte, key string, value interface{}) ([]byte, error)

	// SetBoolean sets a boolean value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes.
	SetBoolean(data []byte, key string, value bool) ([]byte, error)

	// SetInt sets an integer value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes.
	SetInt(data []byte, key string, value int64) ([]byte, error)

	// SetFloat sets a float value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes.
	SetFloat(data []byte, key string, value float64) ([]byte, error)

	// SetString sets a string value for a given key in JSON bytes.
	// The key can be a dot-separated path to access nested values.
	// Returns the modified JSON bytes.
	SetString(data []byte, key, value string) ([]byte, error)
}

// JSONParser is the interface that combines both JSONGetter and JSONSetter interfaces.
type JSONParser interface {
	JSONGetter
	JSONSetter
}

// NewGetter returns a new JSONGetter implementation.
func NewGetter() JSONGetter {
	return &tidwallJSONParser{}
}

// NewSetter returns a new JSONSetter implementation.
func NewSetter() JSONSetter {
	return &tidwallJSONParser{}
}

// Default is the default JSONParser implementation.
var Default = NewWithConfig(config.Default)

// GetValue is a convenience function that uses the default JSONParser.
func GetValue(data []byte, key string) (interface{}, error) {
	return Default.GetValue(data, key)
}

// GetBoolean is a convenience function that uses the default JSONParser.
func GetBoolean(data []byte, key string) (bool, error) {
	return Default.GetBoolean(data, key)
}

// GetInt is a convenience function that uses the default JSONParser.
func GetInt(data []byte, key string) (int64, error) {
	return Default.GetInt(data, key)
}

// GetFloat is a convenience function that uses the default JSONParser.
func GetFloat(data []byte, key string) (float64, error) {
	return Default.GetFloat(data, key)
}

// GetString is a convenience function that uses the default JSONParser.
func GetString(data []byte, key string) (string, error) {
	return Default.GetString(data, key)
}

// SetValue is a convenience function that uses the default JSONParser.
func SetValue(data []byte, key string, value interface{}) ([]byte, error) {
	return Default.SetValue(data, key, value)
}

// SetBoolean is a convenience function that uses the default JSONParser.
func SetBoolean(data []byte, key string, value bool) ([]byte, error) {
	return Default.SetBoolean(data, key, value)
}

// SetInt is a convenience function that uses the default JSONParser.
func SetInt(data []byte, key string, value int64) ([]byte, error) {
	return Default.SetInt(data, key, value)
}

// SetFloat is a convenience function that uses the default JSONParser.
func SetFloat(data []byte, key string, value float64) ([]byte, error) {
	return Default.SetFloat(data, key, value)
}

// SetString is a convenience function that uses the default JSONParser.
func SetString(data []byte, key, value string) ([]byte, error) {
	return Default.SetString(data, key, value)
}

// Reset resets the default JSONParser implementation based on the default configuration.
func Reset() {
	Default = NewWithConfig(config.Default)
}
