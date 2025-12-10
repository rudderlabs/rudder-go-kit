package jsonrs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTidwallJSONValidFunctionality(t *testing.T) {
	j := NewValidatorWithLibrary(TidwallLib)

	t.Run("valid json", func(t *testing.T) {
		// Test valid JSON objects
		require.True(t, j.Valid([]byte(`{}`)), "empty object should be valid")
		require.True(t, j.Valid([]byte(`{"a":"b"}`)), "simple object should be valid")
		require.True(t, j.Valid([]byte(`{"a":1,"b":2.3,"c":true,"d":null,"e":["f",false]}`)), "complex object should be valid")

		// Test valid JSON arrays
		require.True(t, j.Valid([]byte(`[]`)), "empty array should be valid")
		require.True(t, j.Valid([]byte(`[1,2,3]`)), "array of numbers should be valid")
		require.True(t, j.Valid([]byte(`["a","b","c"]`)), "array of strings should be valid")
		require.True(t, j.Valid([]byte(`[{"a":"b"},{"c":"d"}]`)), "array of objects should be valid")

		// Test valid JSON primitives
		require.True(t, j.Valid([]byte(`"string"`)), "string should be valid")
		require.True(t, j.Valid([]byte(`123`)), "number should be valid")
		require.True(t, j.Valid([]byte(`true`)), "boolean true should be valid")
		require.True(t, j.Valid([]byte(`false`)), "boolean false should be valid")
		require.True(t, j.Valid([]byte(`null`)), "null should be valid")
	})

	t.Run("invalid json", func(t *testing.T) {
		// Test invalid JSON syntax
		require.False(t, j.Valid([]byte(`{`)), "unclosed object should be invalid")
		require.False(t, j.Valid([]byte(`[`)), "unclosed array should be invalid")
		require.False(t, j.Valid([]byte(`"`)), "unclosed string should be invalid")
		require.False(t, j.Valid([]byte(`{"a":"b"`)), "missing closing brace should be invalid")
		require.False(t, j.Valid([]byte(`{"a":}`)), "missing value should be invalid")
		require.False(t, j.Valid([]byte(`{"a",1}`)), "comma instead of colon should be invalid")
		require.False(t, j.Valid([]byte(`{a:"b"}`)), "unquoted key should be invalid")

		// Test invalid JSON values
		require.False(t, j.Valid([]byte(`undefined`)), "undefined is not valid JSON")
		require.False(t, j.Valid([]byte(`NaN`)), "NaN is not valid JSON")
		require.False(t, j.Valid([]byte(`Infinity`)), "Infinity is not valid JSON")
		require.False(t, j.Valid([]byte(`{a}`)), "shorthand object notation is not valid JSON")

		// Test empty or non-JSON input
		require.False(t, j.Valid([]byte(``)), "empty input should be invalid")
		require.False(t, j.Valid([]byte(` `)), "whitespace should be invalid")
		require.False(t, j.Valid([]byte(`not json`)), "plain text should be invalid")
	})
}
