package jsonparser

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test SetValue with gjson implementation
	updatedData, err := parser.SetValue(data, "age", 31)
	assert.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	assert.NoError(t, err)
	assert.Equal(t, float64(31), value)

	// Test with jsonparser implementation
	conf.Set("JSONParser.Library", JsonparserLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test SetValue with jsonparser implementation
	updatedData, err = parser.SetValue(data, "age", 32)
	assert.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	assert.NoError(t, err)
	assert.Equal(t, float64(32), value)

	// Test with separate getter and setter implementations
	conf.Set("JSONParser.Library.Getter", JsonparserLib)
	conf.Set("JSONParser.Library.Setter", GjsonLib)
	parser = NewWithConfig(conf)

	// Test GetValue with jsonparser implementation
	value, err = parser.GetValue(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test SetValue with gjson implementation
	updatedData, err = parser.SetValue(data, "age", 33)
	assert.NoError(t, err)
	value, err = parser.GetValue(updatedData, "age")
	assert.NoError(t, err)
	assert.Equal(t, float64(33), value)

	// Test NewWithLibrary
	parser = NewWithLibrary(JsonparserLib)
	value, err = parser.GetValue(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test Reset
	oldDefault := Default
	Reset()
	assert.NotSame(t, oldDefault, Default)
}

func TestGetterSetter(t *testing.T) {
	// Test the new getter and setter interfaces
	data := []byte(`{"name": "John", "age": 30}`)

	// Test with NewGetter
	getter := NewGetter()
	value, err := getter.GetValue(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test with NewSetter
	setter := NewSetter()
	updatedData, err := setter.SetValue(data, "age", 31)
	assert.NoError(t, err)

	// Verify the update using the getter
	value, err = getter.GetValue(updatedData, "age")
	assert.NoError(t, err)
	assert.Equal(t, float64(31), value)

	// Test with NewGetterWithConfig
	conf := config.New()
	conf.Set("JSONParser.Library.Getter", JsonparserLib)
	getter = NewGetterWithConfig(conf)
	value, err = getter.GetValue(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test with NewSetterWithConfig
	conf.Set("JSONParser.Library.Setter", GjsonLib)
	setter = NewSetterWithConfig(conf)
	updatedData, err = setter.SetValue(data, "age", 32)
	assert.NoError(t, err)

	// Verify the update using the getter
	value, err = getter.GetValue(updatedData, "age")
	assert.NoError(t, err)
	assert.Equal(t, float64(32), value)

	// Test with NewGetterWithLibrary
	getter = NewGetterWithLibrary(JsonparserLib)
	value, err = getter.GetValue(data, "name")
	assert.NoError(t, err)
	assert.Equal(t, "John", value)

	// Test with NewSetterWithLibrary
	setter = NewSetterWithLibrary(GjsonLib)
	updatedData, err = setter.SetValue(data, "age", 33)
	assert.NoError(t, err)

	// Verify the update using the getter
	value, err = getter.GetValue(updatedData, "age")
	assert.NoError(t, err)
	assert.Equal(t, float64(33), value)
}
