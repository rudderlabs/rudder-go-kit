package jsonparser

import (
	"errors"

	"github.com/rudderlabs/rudder-go-kit/config"
)

// Getter is the interface that wraps the basic operations for retrieving values from JSON bytes.
//
// All methods accept:
//   - data: the raw JSON bytes to query. Must not be empty, otherwise returns ErrEmptyJSON.
//   - path: one or more path segments identifying the target value. At least one segment must be
//     provided (otherwise returns ErrNoKeysProvided), and no segment may be empty
//     (otherwise returns ErrEmptyKey). Use separate segments for nested access (e.g. "user", "name")
//     and bracket notation for array indices (e.g. "users", "[0]").
//     Each segment is treated as a literal key name; special characters (dots, wildcards, pipes, etc.)
//     have no special meaning.
type Getter interface {
	// GetValue retrieves the raw value for a given key from JSON bytes.
	// Returns the raw value bytes (including JSON type wrappers, e.g. quoted strings).
	//
	// Errors:
	//   - ErrEmptyJSON: data is empty.
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: a path segment is empty.
	//   - ErrKeyNotFound: the path does not exist in the JSON data.
	//
	// Examples:
	//   GetValue(data, "key1", "key2")  // retrieves the value at data.key1.key2
	//   GetValue(data, "key1", "[0]")   // retrieves the value at data.key1[0]
	//   GetValue({"key":"val"}, "key")  // returns []byte(`"val"`)
	GetValue(data []byte, path ...string) ([]byte, error)

	// GetBoolean retrieves a boolean value for a given key from JSON bytes.
	// The value must be a JSON boolean (true/false). Non-boolean values including null
	// return ErrNotOfExpectedType.
	//
	// Errors:
	//   - ErrEmptyJSON: data is empty.
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: a path segment is empty.
	//   - ErrKeyNotFound: the path does not exist in the JSON data.
	//   - ErrNotOfExpectedType: the value at the path is not a boolean (or is null).
	//
	// Examples:
	//   GetBoolean({"active": true}, "active")   // returns true
	//   GetBoolean({"active": "true"}, "active")  // returns ErrNotOfExpectedType
	GetBoolean(data []byte, path ...string) (bool, error)

	// GetInt retrieves an integer value for a given key from JSON bytes.
	// The value must be a JSON number. Float values are truncated to int64.
	// Non-numeric values including null return ErrNotOfExpectedType.
	//
	// Errors:
	//   - ErrEmptyJSON: data is empty.
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: a path segment is empty.
	//   - ErrKeyNotFound: the path does not exist in the JSON data.
	//   - ErrNotOfExpectedType: the value at the path is not a number (or is null).
	//
	// Examples:
	//   GetInt({"age": 30}, "age")     // returns 30
	//   GetInt({"age": 2.7}, "age")    // returns 2 (truncated)
	//   GetInt({"age": "30"}, "age")   // returns ErrNotOfExpectedType
	GetInt(data []byte, path ...string) (int64, error)

	// GetFloat retrieves a float value for a given key from JSON bytes.
	// The value must be a JSON number. Non-numeric values including null
	// return ErrNotOfExpectedType.
	//
	// Errors:
	//   - ErrEmptyJSON: data is empty.
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: a path segment is empty.
	//   - ErrKeyNotFound: the path does not exist in the JSON data.
	//   - ErrNotOfExpectedType: the value at the path is not a number (or is null).
	//
	// Examples:
	//   GetFloat({"price": 29.99}, "price")    // returns 29.99
	//   GetFloat({"price": "29.99"}, "price")  // returns ErrNotOfExpectedType
	GetFloat(data []byte, path ...string) (float64, error)

	// GetString retrieves a string value for a given key from JSON bytes.
	// The value must be a JSON string. Non-string values including null
	// return ErrNotOfExpectedType.
	//
	// Errors:
	//   - ErrEmptyJSON: data is empty.
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: a path segment is empty.
	//   - ErrKeyNotFound: the path does not exist in the JSON data.
	//   - ErrNotOfExpectedType: the value at the path is not a string (or is null).
	//
	// Examples:
	//   GetString({"name": "John"}, "name")  // returns "John"
	//   GetString({"name": 123}, "name")     // returns ErrNotOfExpectedType
	GetString(data []byte, path ...string) (string, error)
}

// Setter is the interface that wraps the basic operations for setting values in JSON bytes.
//
// All methods accept:
//   - data: the raw JSON bytes to modify. If empty, a new JSON object ("{}") is created.
//   - path: one or more path segments identifying the target key. At least one segment must be
//     provided (otherwise returns ErrNoKeysProvided), and no segment may be empty
//     (otherwise returns ErrEmptyKey). Use separate segments for nested access (e.g. "user", "name")
//     and bracket notation for array indices (e.g. "users", "[0]").
//     Each segment is treated as a literal key name; special characters (dots, wildcards, pipes, etc.)
//     have no special meaning. Intermediate objects are created automatically if they do not exist.
type Setter interface {
	// SetValue sets the value for a given key in JSON bytes.
	// The value is JSON-marshalled before being set. Supported types (string, bool, int64, float64)
	// are serialized directly; other types are marshalled via jsonrs.Marshal.
	//
	// Errors:
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: any path segment is empty.
	//
	// Examples:
	//   SetValue(data, "Jane", "user", "name")  // sets data.user.name to "Jane"
	//   SetValue(data, 30, "user", "age")        // sets data.user.age to 30
	//   SetValue([]byte{}, "v", "key")           // creates {"key":"v"}
	SetValue(data []byte, value any, path ...string) ([]byte, error)

	// SetBoolean sets a boolean value for a given key in JSON bytes.
	//
	// Errors:
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: any path segment is empty.
	SetBoolean(data []byte, value bool, path ...string) ([]byte, error)

	// SetInt sets an integer value for a given key in JSON bytes.
	//
	// Errors:
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: any path segment is empty.
	SetInt(data []byte, value int64, path ...string) ([]byte, error)

	// SetFloat sets a float value for a given key in JSON bytes.
	//
	// Errors:
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: any path segment is empty.
	SetFloat(data []byte, value float64, path ...string) ([]byte, error)

	// SetString sets a string value for a given key in JSON bytes.
	//
	// Errors:
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: any path segment is empty.
	SetString(data []byte, value string, path ...string) ([]byte, error)
}

// Deleter is the interface that wraps the basic operations for deleting keys from JSON bytes.
type Deleter interface {
	// DeleteKey deletes a key from JSON bytes.
	//
	// Parameters:
	//   - data: the raw JSON bytes to modify. Must not be empty, otherwise returns ErrEmptyJSON.
	//   - path: one or more path segments identifying the key to delete. At least one segment must be
	//     provided (otherwise returns ErrNoKeysProvided), and no segment may be empty
	//     (otherwise returns ErrEmptyKey). Each segment is treated as a literal key name.
	//
	// If the key does not exist, the original JSON is returned unchanged (no error).
	//
	// Errors:
	//   - ErrEmptyJSON: data is empty.
	//   - ErrNoKeysProvided: no path segments provided.
	//   - ErrEmptyKey: a path segment is empty.
	DeleteKey(data []byte, path ...string) ([]byte, error)
}

// SoftGetter is the interface that wraps lenient getter operations that never return errors.
// Unlike [Getter], these methods return zero values on any failure (empty data, missing keys,
// type mismatches, invalid JSON). They also perform best-effort type coercion (e.g. string "123"
// to int, boolean true to 1).
//
// All methods accept:
//   - data: the raw JSON bytes to query. If empty or nil, the zero value is returned.
//   - path: one or more path segments identifying the target value. If empty or if any
//     segment is empty, the zero value is returned. Each segment is treated as a literal key name.
type SoftGetter interface {
	// GetValueOrEmpty retrieves the raw value for a given key from JSON bytes.
	// Returns nil if the key does not exist, data is empty/nil, no path is given,
	// or the first path segment is empty.
	GetValueOrEmpty(data []byte, path ...string) []byte

	// GetBooleanOrFalse retrieves a boolean value for a given path from JSON bytes.
	// Returns false on any failure. Performs type coercion:
	//   - boolean: returned as-is
	//   - string: parsed via strconv.ParseBool ("true"/"false"/"1"/"0")
	//   - number: non-zero is true
	//   - null/other: returns false
	GetBooleanOrFalse(data []byte, path ...string) bool

	// GetIntOrZero retrieves an integer value for a given key from JSON bytes.
	// Returns 0 on any failure. Performs type coercion:
	//   - number: parsed as float64 then truncated to int64
	//   - boolean: true=1, false=0
	//   - string: parsed as float64 then truncated to int64
	//   - null/other: returns 0
	GetIntOrZero(data []byte, path ...string) int64

	// GetFloatOrZero retrieves a float value for a given key from JSON bytes.
	// Returns 0 on any failure. Performs type coercion:
	//   - number: parsed as float64
	//   - boolean: true=1, false=0
	//   - string: parsed as float64
	//   - null/other: returns 0
	GetFloatOrZero(data []byte, path ...string) float64

	// GetStringOrEmpty retrieves a string value for a given key from JSON bytes.
	// Returns "" on any failure. Performs type coercion:
	//   - string/boolean/number: returned as string representation
	//   - object/array: returned as raw JSON string
	//   - null: returns ""
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
func GetValue(data []byte, path ...string) (any, error) {
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
func SetValue(data []byte, value any, path ...string) ([]byte, error) {
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

// GetValueOrEmpty is a convenience function that uses the default JSONParser.
func GetValueOrEmpty(data []byte, path ...string) []byte {
	return Default.GetValueOrEmpty(data, path...)
}

// GetBooleanOrFalse is a convenience function that uses the default JSONParser.
func GetBooleanOrFalse(data []byte, path ...string) bool {
	return Default.GetBooleanOrFalse(data, path...)
}

// GetIntOrZero is a convenience function that uses the default JSONParser.
func GetIntOrZero(data []byte, path ...string) int64 {
	return Default.GetIntOrZero(data, path...)
}

// GetFloatOrZero is a convenience function that uses the default JSON
func GetFloatOrZero(data []byte, path ...string) float64 {
	return Default.GetFloatOrZero(data, path...)
}

// GetStringOrEmpty is a convenience function that uses the default JSONParser.
func GetStringOrEmpty(data []byte, path ...string) string {
	return Default.GetStringOrEmpty(data, path...)
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
