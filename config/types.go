package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// configTypes defines the types that can be used for configuration variables.
type configTypes interface {
	int | int64 | string | time.Duration | bool | float64 | []string | map[string]any
}

const (
	stringType      = "string"
	intType         = "int"
	int64Type       = "int64"
	float64Type     = "float64"
	boolType        = "bool"
	durationType    = "time.Duration"
	stringSliceType = "[]string"
	stringMapType   = "map[string]any"
)

// getTypeName returns the string representation of the type of a config variable.
func getTypeName[T configTypes](t T) string {
	switch v := any(t).(type) {
	case string:
		return stringType
	case int:
		return intType
	case int64:
		return int64Type
	case float64:
		return float64Type
	case bool:
		return boolType
	case []string:
		return stringSliceType
	case map[string]any:
		return stringMapType
	case time.Duration:
		return durationType
	default:
		panic(fmt.Errorf("unsupported type %T for config variable", v))
	}
}

// getStringValue converts a config variable of type T to its string representation.
func getStringValue[T configTypes](t T) string {
	switch v := any(t).(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case []string:
		return strings.Join(v, ",")
	case map[string]any:
		return fmt.Sprintf("%v", v) // or use a more specific format?
	case time.Duration:
		return v.String()
	default:
		panic(fmt.Errorf("unsupported type %T for config variable", v))
	}
}
