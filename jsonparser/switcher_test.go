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
	require.Equal(t, "John", value)

	// Test SetValue with gjson implementation
	updatedData, err := parser.SetValue(data, "age", 31)
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, float64(31), value)

	// Test with jsonparser implementation
	conf.Set("JSONParser.Library", GrafanaLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, "John", value)

	// Test SetValue with jsonparser implementation
	updatedData, err = parser.SetValue(data, "age", 32)
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, float64(32), value)

	// Test with separate getter and setter implementations
	conf.Set("JSONParser.Library.Getter", GrafanaLib)
	conf.Set("JSONParser.Library.Setter", TidwallLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, "John", value)

	// Test SetValue with gjson implementation
	updatedData, err = parser.SetValue(data, "age", 33)
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, float64(33), value)

	// Test NewWithLibrary
	parser = NewWithLibrary(GrafanaLib)
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, "John", value)

	// Test Reset
	oldDefault := Default
	Reset()
	require.NotSame(t, oldDefault, Default)
}
