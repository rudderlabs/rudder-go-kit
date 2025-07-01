package jsonparser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
)

func TestSwitcher(t *testing.T) {
	// Create a test configuration
	conf := config.New()

	// Test with default configuration (should use gjson)
	parser := NewWithConfig(conf)

	// Test GetValue with gjson implementation
	data := []byte(`{"name": "John", "age": 30}`)
	value, err := parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, []byte("\"John\""), value)

	// Test GetString with gjson
	str, err := parser.GetString(data, "name")
	require.NoError(t, err)
	require.Equal(t, "John", str)

	// Test SetValue with gjson implementation
	updatedData, err := parser.SetValue(data, 31, "age")
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, []byte("31"), value)
	intVal, err := parser.GetInt(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, int64(31), intVal)

	// Test with jsonparser implementation
	conf.Set("JSONParser.Library", GrafanaLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, []byte("\"John\""), value)
	// Test GetString with gjson
	str, err = parser.GetString(data, "name")
	require.NoError(t, err)
	require.Equal(t, "John", str)

	// Test SetValue with jsonparser implementation
	updatedData, err = parser.SetValue(data, 32, "age")
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, []byte("32"), value)

	// Test with separate Getter and Setter implementations
	conf.Set("JSONParser.Library.Getter", GrafanaLib)
	conf.Set("JSONParser.Library.Setter", TidwallLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, []byte("\"John\""), value)

	// Test SetValue with gjson implementation
	updatedData, err = parser.SetValue(data, 33, "age")
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, []byte("33"), value)
	intVal, err = parser.GetInt(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, int64(33), intVal)

	// Test NewWithLibrary
	parser = NewWithLibrary(GrafanaLib)
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, []byte("\"John\""), value)

	// Test Reset
	oldDefault := Default
	Reset()
	require.NotSame(t, oldDefault, Default)
}

func TestSwitcher_SoftGetter(t *testing.T) {
	// Create test data
	data := []byte(`{"name": "John", "age": 30, "active": true, "score": 9.5}`)
	emptyData := []byte(`{}`)
	invalidData := []byte(`{"invalid`)

	testCases := []struct {
		name     string
		library  string
		testFunc func(t *testing.T, parser JSONParser)
	}{
		{
			name:    "GrafanaLib SoftGetter",
			library: GrafanaLib,
			testFunc: func(t *testing.T, parser JSONParser) {
				// Test GetValueOrEmpty
				value := parser.GetValueOrEmpty(data, "name")
				require.Equal(t, []byte(`"John"`), value)
				require.Empty(t, parser.GetValueOrEmpty(data, "nonexistent"))
				require.Empty(t, parser.GetValueOrEmpty(emptyData, "name"))
				require.Empty(t, parser.GetValueOrEmpty(invalidData, "name"))

				// Test GetBooleanOrFalse
				require.True(t, parser.GetBooleanOrFalse(data, "active"))
				require.False(t, parser.GetBooleanOrFalse(data, "nonexistent"))
				require.False(t, parser.GetBooleanOrFalse(emptyData, "active"))
				require.False(t, parser.GetBooleanOrFalse(invalidData, "active"))
				require.False(t, parser.GetBooleanOrFalse(data, "name")) // Non-boolean returns false

				// Test GetIntOrZero
				require.Equal(t, int64(30), parser.GetIntOrZero(data, "age"))
				require.Equal(t, int64(0), parser.GetIntOrZero(data, "nonexistent"))
				require.Equal(t, int64(0), parser.GetIntOrZero(emptyData, "age"))
				require.Equal(t, int64(0), parser.GetIntOrZero(invalidData, "age"))
				require.Equal(t, int64(0), parser.GetIntOrZero(data, "name")) // Non-numeric returns 0

				// Test GetFloatOrZero
				require.Equal(t, 9.5, parser.GetFloatOrZero(data, "score"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(data, "nonexistent"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(emptyData, "score"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(invalidData, "score"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(data, "name")) // Non-numeric returns 0

				// Test GetStringOrEmpty
				require.Equal(t, "John", parser.GetStringOrEmpty(data, "name"))
				require.Equal(t, "", parser.GetStringOrEmpty(data, "nonexistent"))
				require.Equal(t, "", parser.GetStringOrEmpty(emptyData, "name"))
				require.Equal(t, "", parser.GetStringOrEmpty(invalidData, "name"))
			},
		},
		{
			name:    "TidwallLib SoftGetter",
			library: TidwallLib,
			testFunc: func(t *testing.T, parser JSONParser) {
				// Test GetValueOrEmpty
				value := parser.GetValueOrEmpty(data, "name")
				require.Equal(t, []byte(`"John"`), value)
				require.Empty(t, parser.GetValueOrEmpty(data, "nonexistent"))
				require.Empty(t, parser.GetValueOrEmpty(emptyData, "name"))
				require.Empty(t, parser.GetValueOrEmpty(invalidData, "name"))

				// Test GetBooleanOrFalse
				require.True(t, parser.GetBooleanOrFalse(data, "active"))
				require.False(t, parser.GetBooleanOrFalse(data, "nonexistent"))
				require.False(t, parser.GetBooleanOrFalse(emptyData, "active"))
				require.False(t, parser.GetBooleanOrFalse(invalidData, "active"))
				require.False(t, parser.GetBooleanOrFalse(data, "name")) // Non-boolean returns false

				// Test GetIntOrZero
				require.Equal(t, int64(30), parser.GetIntOrZero(data, "age"))
				require.Equal(t, int64(0), parser.GetIntOrZero(data, "nonexistent"))
				require.Equal(t, int64(0), parser.GetIntOrZero(emptyData, "age"))
				require.Equal(t, int64(0), parser.GetIntOrZero(invalidData, "age"))
				require.Equal(t, int64(0), parser.GetIntOrZero(data, "name")) // Non-numeric returns 0

				// Test GetFloatOrZero
				require.Equal(t, 9.5, parser.GetFloatOrZero(data, "score"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(data, "nonexistent"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(emptyData, "score"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(invalidData, "score"))
				require.Equal(t, 0.0, parser.GetFloatOrZero(data, "name")) // Non-numeric returns 0

				// Test GetStringOrEmpty
				require.Equal(t, "John", parser.GetStringOrEmpty(data, "name"))
				require.Equal(t, "", parser.GetStringOrEmpty(data, "nonexistent"))
				require.Equal(t, "", parser.GetStringOrEmpty(emptyData, "name"))
				require.Equal(t, "", parser.GetStringOrEmpty(invalidData, "name"))
			},
		},
		{
			name:    "Switcher with configurable SoftGetter",
			library: "",
			testFunc: func(t *testing.T, _ JSONParser) {
				// Create config and test with different SoftGetter implementations
				conf := config.New()

				// Test with GrafanaLib as SoftGetter
				conf.Set("JSONParser.Library.SoftGetter", GrafanaLib)
				parser := NewWithConfig(conf)

				require.Equal(t, "John", parser.GetStringOrEmpty(data, "name"))
				require.Equal(t, int64(30), parser.GetIntOrZero(data, "age"))
				require.True(t, parser.GetBooleanOrFalse(data, "active"))
				require.Equal(t, 9.5, parser.GetFloatOrZero(data, "score"))

				// Test with TidwallLib as SoftGetter
				conf.Set("JSONParser.Library.SoftGetter", TidwallLib)
				parser = NewWithConfig(conf)

				require.Equal(t, "John", parser.GetStringOrEmpty(data, "name"))
				require.Equal(t, int64(30), parser.GetIntOrZero(data, "age"))
				require.True(t, parser.GetBooleanOrFalse(data, "active"))
				require.Equal(t, 9.5, parser.GetFloatOrZero(data, "score"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var parser JSONParser
			if tc.library != "" {
				parser = NewWithLibrary(tc.library)
			}
			tc.testFunc(t, parser)
		})
	}
}
