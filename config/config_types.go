package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

// Reloadable is used as a wrapper for hot-reloadable config variables
type Reloadable[T configTypes] struct {
	value T
	lock  sync.RWMutex
}

// Load should be used to read the underlying value without worrying about data races
func (a *Reloadable[T]) Load() T {
	a.lock.RLock()
	v := a.value
	a.lock.RUnlock()
	return v
}

func (a *Reloadable[T]) store(v T) {
	a.lock.Lock()
	a.value = v
	a.lock.Unlock()
}

// swapIfNotEqual is used internally to swap the value of a hot-reloadable config variable
// if the new value is not equal to the old value
func (a *Reloadable[T]) swapIfNotEqual(new T, compare func(old, new T) bool) (old T, swapped bool) {
	a.lock.Lock()
	defer a.lock.Unlock()
	if !compare(a.value, new) {
		old := a.value
		a.value = new
		return old, true
	}
	return a.value, false
}

type configValue struct {
	value        any
	multiplier   any
	defaultValue any
	keys         []string
}

func newConfigValue(value, multiplier, defaultValue any, keys []string) *configValue {
	return &configValue{
		value:        value,
		multiplier:   multiplier,
		defaultValue: defaultValue,
		keys:         keys,
	}
}

func compare[T comparable]() func(a, b T) bool {
	return func(a, b T) bool {
		return a == b
	}
}

func mapDeepEqual[K comparable, V any](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || !reflect.DeepEqual(v, w) {
			return false
		}
	}
	return true
}

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
