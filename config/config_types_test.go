package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetTypeName(t *testing.T) {
	require.Equal(t, stringType, getTypeName("test"), "string type check")
	require.Equal(t, intType, getTypeName(42), "int type check")
	require.Equal(t, int64Type, getTypeName(int64(42)), "int64 type check")
	require.Equal(t, float64Type, getTypeName(3.14), "float64 type check")
	require.Equal(t, boolType, getTypeName(true), "bool type check")
	require.Equal(t, durationType, getTypeName(time.Hour), "duration type check")
	require.Equal(t, stringSliceType, getTypeName([]string{"a", "b"}), "string slice type check")
	require.Equal(t, stringMapType, getTypeName(map[string]any{"key": "value"}), "string map type check")
}

func TestGetStringValue(t *testing.T) {
	require.Equal(t, "test", getStringValue("test"), "string value check")
	require.Equal(t, "42", getStringValue(42), "int value check")
	require.Equal(t, "42", getStringValue(int64(42)), "int64 value check")
	require.Equal(t, "3.14", getStringValue(3.14), "float64 value check")
	require.Equal(t, "true", getStringValue(true), "bool true value check")
	require.Equal(t, "false", getStringValue(false), "bool false value check")
	require.Equal(t, "1h0m0s", getStringValue(time.Hour), "duration value check")
	require.Equal(t, "a,b,c", getStringValue([]string{"a", "b", "c"}), "string slice value check")
	require.Equal(t, "", getStringValue([]string{}), "empty string slice value check")
	require.Equal(t, "map[key:value]", getStringValue(map[string]any{"key": "value"}), "string map value check")
}

func TestMapDeepEqual(t *testing.T) {
	tests := []struct {
		name     string
		map1     map[string]any
		map2     map[string]any
		expected bool
	}{
		{
			"empty maps",
			map[string]any{},
			map[string]any{},
			true,
		},
		{
			"same maps",
			map[string]any{"a": 1, "b": "test"},
			map[string]any{"a": 1, "b": "test"},
			true,
		},
		{
			"different values",
			map[string]any{"a": 1, "b": "test"},
			map[string]any{"a": 2, "b": "test"},
			false,
		},
		{
			"different keys",
			map[string]any{"a": 1, "b": "test"},
			map[string]any{"a": 1, "c": "test"},
			false,
		},
		{
			"different sizes",
			map[string]any{"a": 1, "b": "test"},
			map[string]any{"a": 1},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapDeepEqual(tt.map1, tt.map2)
			if result != tt.expected {
				t.Errorf("mapDeepEqual(%v, %v) = %v, want %v", tt.map1, tt.map2, result, tt.expected)
			}
		})
	}
}
