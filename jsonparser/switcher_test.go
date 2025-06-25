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
	require.Equal(t, []byte("John"), value)

	// Test SetValue with gjson implementation
	updatedData, err := parser.SetValue(data, 31, "age")
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, []byte("31"), value)

	// Test with jsonparser implementation
	conf.Set("JSONParser.Library", GrafanaLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, []byte("John"), value)

	// Test SetValue with jsonparser implementation
	updatedData, err = parser.SetValue(data, 32, "age")
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, []byte("32"), value)

	// Test with separate getter and setter implementations
	conf.Set("JSONParser.Library.Getter", GrafanaLib)
	conf.Set("JSONParser.Library.Setter", TidwallLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, []byte("John"), value)

	// Test SetValue with gjson implementation
	updatedData, err = parser.SetValue(data, 33, "age")
	require.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	require.NoError(t, err)
	require.Equal(t, []byte("33"), value)

	// Test NewWithLibrary
	parser = NewWithLibrary(GrafanaLib)
	value, err = parser.GetValue(data, "name")
	require.NoError(t, err)
	require.Equal(t, []byte("John"), value)

	// Test Reset
	oldDefault := Default
	Reset()
	require.NotSame(t, oldDefault, Default)
}
