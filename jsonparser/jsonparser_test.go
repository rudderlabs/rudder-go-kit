package jsonparser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"
)

func suiteGetValue(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     any
		wantErr  bool
		err      error
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			keys:     []string{"name"},
			want:     []byte(`"John"`),
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user", "name"},
			want:     []byte(`"John"`),
			wantErr:  false,
		},
		{
			name:     "json bytes",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user"},
			want:     []byte(`{"name": "John", "age": 30}`),
			wantErr:  false,
		},
		{
			name:     "array index",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			keys:     []string{"users", "[1]"},
			want:     []byte(`"Jane"`),
			wantErr:  false,
		},
		{
			name:     "array bytes",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			keys:     []string{"users"},
			want:     []byte(`["John", "Jane", "Bob"]`),
			wantErr:  false,
		},
		{
			name:     "nested array",
			jsonData: `{"data": {"users": [{"name": "John"}, {"name": "Jane"}]}}`,
			keys:     []string{"data", "users", "[1]", "name"},
			want:     []byte(`"Jane"`),
			wantErr:  false,
		},
		{
			name:     "numeric value",
			jsonData: `{"user": {"age": 30}}`,
			keys:     []string{"user", "age"},
			want:     []byte("30"),
			wantErr:  false,
		},
		{
			name:     "boolean value",
			jsonData: `{"user": {"active": true}}`,
			keys:     []string{"user", "active"},
			want:     []byte("true"),
			wantErr:  false,
		},
		{
			name:     "null value",
			jsonData: `{"user": {"middleName": null}}`,
			keys:     []string{"user", "middleName"},
			want:     []byte("null"),
			wantErr:  false,
		},
		{
			name:     "string with escaped quote",
			jsonData: `{"key": "hello\"world"}`,
			keys:     []string{"key"},
			want:     []byte(`"hello\"world"`),
			wantErr:  false,
		},
		{
			name:     "string with escaped backslash",
			jsonData: `{"key": "hello\\world"}`,
			keys:     []string{"key"},
			want:     []byte(`"hello\\world"`),
			wantErr:  false,
		},
		{
			name:     "string with unicode escape",
			jsonData: `{"key": "\u0041\u0042"}`,
			keys:     []string{"key"},
			want:     []byte(`"\u0041\u0042"`),
			wantErr:  false,
		},
		{
			name:     "string with newline escape",
			jsonData: `{"key": "line1\nline2"}`,
			keys:     []string{"key"},
			want:     []byte(`"line1\nline2"`),
			wantErr:  false,
		},
		{
			name:     "string with tab escape",
			jsonData: `{"key": "col1\tcol2"}`,
			keys:     []string{"key"},
			want:     []byte(`"col1\tcol2"`),
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			want:     nil,
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
		{
			name:     "no keys provided",
			jsonData: `{"name": "John"}`,
			keys:     []string{},
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "invalid array index",
			jsonData: `{"users": ["John", "Jane"]}`,
			keys:     []string{"users", "[2]"},
			want:     nil,
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
		{
			name:     "invalid json",
			jsonData: `{"name": "John"`,
			keys:     []string{"name"},
			want:     []byte(`"John"`),
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyJSON,
		},
		{
			name:     "invalid path empty",
			jsonData: `{"user": {"active": true}}`,
			keys:     []string{"user", "", "active"},
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetValue([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, got, tt.want)
		})
	}
}

func suiteSetValue(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		value    any
		want     map[string]any
		wantErr  bool
		err      error
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			keys:     []string{"name"},
			value:    "Jane",
			want:     map[string]any{"name": "Jane", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user", "name"},
			value:    "Jane",
			want:     map[string]any{"user": map[string]any{"name": "Jane", "age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "new key",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]any{"name": "John", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "new nested key",
			jsonData: `{"user": {"name": "John"}}`,
			keys:     []string{"user", "age"},
			value:    30,
			want:     map[string]any{"user": map[string]any{"name": "John", "age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "create nested structure",
			jsonData: `{}`,
			keys:     []string{"user", "name"},
			value:    "John",
			want:     map[string]any{"user": map[string]any{"name": "John"}},
			wantErr:  false,
		},
		{
			name:     "array element",
			jsonData: `{"users": ["John", "Jane"]}`,
			keys:     []string{"users", "[1]"},
			value:    "Bob",
			want:     map[string]any{"users": []any{"John", "Bob"}},
			wantErr:  false,
		},
		{
			name:     "string value with embedded quote",
			jsonData: `{}`,
			keys:     []string{"name"},
			value:    `Jane "JJ" Doe`,
			want:     map[string]any{"name": `Jane "JJ" Doe`},
			wantErr:  false,
		},
		{
			name:     "string value with embedded newline",
			jsonData: `{}`,
			keys:     []string{"name"},
			value:    "Jane\nDoe",
			want:     map[string]any{"name": "Jane\nDoe"},
			wantErr:  false,
		},
		{
			name:     "string value with embedded backslash",
			jsonData: `{}`,
			keys:     []string{"path"},
			value:    `C:\Users\file`,
			want:     map[string]any{"path": `C:\Users\file`},
			wantErr:  false,
		},
		{
			name:     "string value with embedded tab",
			jsonData: `{}`,
			keys:     []string{"text"},
			value:    "col1\tcol2",
			want:     map[string]any{"text": "col1\tcol2"},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			value:    "John",
			want:     map[string]any{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			value:    "value",
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"name": "John"}`,
			keys:     []string{},
			value:    "value",
			want:     nil,
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "invalid path empty",
			jsonData: `{"user": {"active": true}}`,
			keys:     []string{"user", "", "active"},
			value:    "cat",
			wantErr:  true,
			err:      ErrEmptyKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetValue([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			var gotMap map[string]any
			err = jsonrs.Unmarshal(got, &gotMap)
			require.NoError(t, err)
			require.Equal(t, tt.want, gotMap)
		})
	}
}

func suiteArrayHandling(t *testing.T, jsonParser JSONParser) {
	// Test more complex array scenarios
	jsonData := []byte(`{
		"users": [
			{"name": "John", "age": 30},
			{"name": "Jane", "age": 25}
		]
	}`)

	// Test getting a value from an array of objects
	value, err := jsonParser.GetValue(jsonData, "users", "[1]", "name")
	require.NoError(t, err)
	require.Equal(t, []byte(`"Jane"`), value)
	str, err := jsonParser.GetString(jsonData, "users", "[1]", "name")
	require.NoError(t, err)
	require.Equal(t, "Jane", str)

	// Test setting a value in an array of objects
	updatedJSON, err := jsonParser.SetValue(jsonData, 31, "users", "[0]", "age")
	require.NoError(t, err)

	// Verify the update
	value, err = jsonParser.GetValue(updatedJSON, "users", "[0]", "age")
	require.NoError(t, err)
	require.Equal(t, []byte("31"), value)
	age, err := jsonParser.GetInt(updatedJSON, "users", "[0]", "age")
	require.NoError(t, err)
	require.Equal(t, int64(31), age)
}

func suiteEdgeCases(t *testing.T, jsonParser JSONParser) {
	// Test with a complex nested structure
	jsonData := []byte(`{
		"data": {
			"users": {
				"0": {
					"profile": {
						"details": {
							"name": "John",
							"contact": {
								"email": "john@example.com"
							}
						}
					}
				}
			}
		}
	}`)

	// Test getting a deeply nested value
	value, err := jsonParser.GetValue(jsonData, "data", "users", "0", "profile", "details", "contact", "email")
	require.NoError(t, err)
	require.Equal(t, []byte(`"john@example.com"`), value)

	// Test setting a deeply nested value
	updatedJSON, err := jsonParser.SetValue(jsonData, "123-456-7890", "data", "users", "0", "profile", "details", "contact", "phone")
	require.NoError(t, err)

	// Verify the update
	value, err = jsonParser.GetValue(updatedJSON, "data", "users", "0", "profile", "details", "contact", "phone")
	require.NoError(t, err)
	require.Equal(t, []byte(`"123-456-7890"`), value)
}

func suiteGetBoolean(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     bool
		wantErr  bool
		err      error
	}{
		{
			name:     "simple boolean true",
			jsonData: `{"active": true}`,
			keys:     []string{"active"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "simple boolean false",
			jsonData: `{"active": false}`,
			keys:     []string{"active"},
			want:     false,
			wantErr:  false,
		},
		{
			name:     "nested boolean",
			jsonData: `{"user": {"active": true}}`,
			keys:     []string{"user", "active"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "array boolean",
			jsonData: `{"settings": [true, false, true]}`,
			keys:     []string{"settings", "[0]"},
			want:     true,
			wantErr:  false,
		},
		{
			name:     "non-boolean value",
			jsonData: `{"active": "true"}`,
			keys:     []string{"active"},
			want:     false,
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"active"},
			want:     false,
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"active"},
			want:     false,
			wantErr:  true,
			err:      ErrEmptyJSON,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     false,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     false,
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "null",
			jsonData: `{"active": null}`,
			keys:     []string{"active"},
			want:     false,
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetBoolean([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func suiteGetInt(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     int64
		wantErr  bool
		err      error
	}{
		{
			name:     "simple integer",
			jsonData: `{"age": 30}`,
			keys:     []string{"age"},
			want:     30,
			wantErr:  false,
		},
		{
			name:     "nested integer",
			jsonData: `{"user": {"age": 30}}`,
			keys:     []string{"user", "age"},
			want:     30,
			wantErr:  false,
		},
		{
			name:     "array integer",
			jsonData: `{"ages": [10, 20, 30]}`,
			keys:     []string{"ages", "[2]"},
			want:     30,
			wantErr:  false,
		},
		{
			name:     "non-integer value",
			jsonData: `{"age": "30"}`,
			keys:     []string{"age"},
			want:     0,
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			want:     0,
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"age"},
			want:     0,
			wantErr:  true,
			err:      ErrEmptyJSON,
		},
		{
			name:     "float value",
			jsonData: `{"age": 2.32}`,
			keys:     []string{"age"},
			want:     2,
			wantErr:  false,
		},
		{
			name:     "negative fractional number value truncated",
			jsonData: `{"age": -2.32}`,
			keys:     []string{"age"},
			want:     -2,
			wantErr:  false,
		},
		{
			name:     "scientific notation integer value",
			jsonData: `{"age": 1e3}`,
			keys:     []string{"age"},
			want:     1000,
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     0,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     0,
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "null",
			jsonData: `{"active": null}`,
			keys:     []string{"active"},
			want:     0,
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetInt([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func suiteGetFloat(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     float64
		wantErr  bool
		err      error
	}{
		{
			name:     "simple float",
			jsonData: `{"price": 29.99}`,
			keys:     []string{"price"},
			want:     29.99,
			wantErr:  false,
		},
		{
			name:     "integer as float",
			jsonData: `{"price": 30}`,
			keys:     []string{"price"},
			want:     30.0,
			wantErr:  false,
		},
		{
			name:     "nested float",
			jsonData: `{"product": {"price": 29.99}}`,
			keys:     []string{"product", "price"},
			want:     29.99,
			wantErr:  false,
		},
		{
			name:     "array float",
			jsonData: `{"prices": [10.5, 20.75, 30.99]}`,
			keys:     []string{"prices", "[1]"},
			want:     20.75,
			wantErr:  false,
		},
		{
			name:     "non-float value",
			jsonData: `{"price": "29.99"}`,
			keys:     []string{"price"},
			want:     0,
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "Product"}`,
			keys:     []string{"price"},
			want:     0,
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"price"},
			want:     0,
			wantErr:  true,
			err:      ErrEmptyJSON,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     0,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     0,
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "null",
			jsonData: `{"active": null}`,
			keys:     []string{"active"},
			want:     0,
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetFloat([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func suiteGetString(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     string
		wantErr  bool
		err      error
	}{
		{
			name:     "simple string",
			jsonData: `{"name": "John"}`,
			keys:     []string{"name"},
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "nested string",
			jsonData: `{"user": {"name": "John"}}`,
			keys:     []string{"user", "name"},
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "array string",
			jsonData: `{"names": ["John", "Jane", "Bob"]}`,
			keys:     []string{"names", "[1]"},
			want:     "Jane",
			wantErr:  false,
		},
		{
			name:     "non-string value",
			jsonData: `{"name": 123}`,
			keys:     []string{"name"},
			want:     "",
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
		{
			name:     "key not found",
			jsonData: `{"age": 30}`,
			keys:     []string{"name"},
			want:     "",
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			want:     "",
			wantErr:  true,
			err:      ErrEmptyJSON,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     "",
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     "",
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "null",
			jsonData: `{"active": null}`,
			keys:     []string{"active"},
			want:     "",
			wantErr:  true,
			err:      ErrNotOfExpectedType,
		},
		{
			name:     "literal special key star",
			jsonData: `{"obj": {"ab": "wildcard-match", "a*b": "literal-star"}}`,
			keys:     []string{"obj", "a*b"},
			want:     "literal-star",
			wantErr:  false,
		},
		{
			name:     "literal special key question",
			jsonData: `{"obj": {"acb": "question-match", "a?b": "literal-question"}}`,
			keys:     []string{"obj", "a?b"},
			want:     "literal-question",
			wantErr:  false,
		},
		{
			name:     "literal special key modifier",
			jsonData: `{"obj": {"@reverse": "literal-modifier"}}`,
			keys:     []string{"obj", "@reverse"},
			want:     "literal-modifier",
			wantErr:  false,
		},
		{
			name:     "literal special key pipe",
			jsonData: `{"obj": {"a|b": "literal-pipe"}}`,
			keys:     []string{"obj", "a|b"},
			want:     "literal-pipe",
			wantErr:  false,
		},
		{
			name:     "literal special key hash",
			jsonData: `{"obj": {"a#b": "literal-hash"}}`,
			keys:     []string{"obj", "a#b"},
			want:     "literal-hash",
			wantErr:  false,
		},
		{
			name:     "literal special key colon",
			jsonData: `{"obj": {"a:b": "literal-colon"}}`,
			keys:     []string{"obj", "a:b"},
			want:     "literal-colon",
			wantErr:  false,
		},
		{
			name:     "literal special key backslash",
			jsonData: `{"obj": {"a\\b": "literal-backslash"}}`,
			keys:     []string{"obj", `a\b`},
			want:     "literal-backslash",
			wantErr:  false,
		},
		{
			name:     "numeric object key",
			jsonData: `{"obj": {"0": {"x": "object-key"}}}`,
			keys:     []string{"obj", "0", "x"},
			want:     "object-key",
			wantErr:  false,
		},
		{
			name:     "numeric array segment requires brackets",
			jsonData: `{"arr": [{"x": "array-index"}]}`,
			keys:     []string{"arr", "0", "x"},
			want:     "",
			wantErr:  true,
			err:      ErrKeyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetString([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func suiteSetBoolean(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		value    bool
		want     map[string]any
		wantErr  bool
		err      error
	}{
		{
			name:     "simple boolean true",
			jsonData: `{"active": false}`,
			keys:     []string{"active"},
			value:    true,
			want:     map[string]any{"active": true},
			wantErr:  false,
		},
		{
			name:     "simple boolean false",
			jsonData: `{"active": true}`,
			keys:     []string{"active"},
			value:    false,
			want:     map[string]any{"active": false},
			wantErr:  false,
		},
		{
			name:     "nested boolean",
			jsonData: `{"user": {"active": false}}`,
			keys:     []string{"user", "active"},
			value:    true,
			want:     map[string]any{"user": map[string]any{"active": true}},
			wantErr:  false,
		},
		{
			name:     "new boolean key",
			jsonData: `{"name": "John"}`,
			keys:     []string{"active"},
			value:    true,
			want:     map[string]any{"name": "John", "active": true},
			wantErr:  false,
		},
		{
			name:     "array boolean",
			jsonData: `{"settings": [false, false]}`,
			keys:     []string{"settings", "[0]"},
			value:    true,
			want:     map[string]any{"settings": []any{true, false}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"active"},
			value:    true,
			want:     map[string]any{"active": true},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			value:    false,
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetBoolean([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			var gotMap map[string]any
			err = jsonrs.Unmarshal(got, &gotMap)
			require.NoError(t, err)
			require.Equal(t, tt.want, gotMap)
		})
	}
}

func suiteSetInt(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		value    int64
		want     map[string]any
		wantErr  bool
		err      error
	}{
		{
			name:     "simple integer",
			jsonData: `{"age": 25}`,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]any{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested integer",
			jsonData: `{"user": {"age": 25}}`,
			keys:     []string{"user", "age"},
			value:    30,
			want:     map[string]any{"user": map[string]any{"age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "new integer key",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]any{"name": "John", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "array integer",
			jsonData: `{"ages": [10, 20]}`,
			keys:     []string{"ages", "[1]"},
			value:    30,
			want:     map[string]any{"ages": []any{float64(10), float64(30)}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]any{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"age": 25}`,
			keys:     []string{""},
			value:    30,
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetInt([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			var gotMap map[string]any
			err = jsonrs.Unmarshal(got, &gotMap)
			require.NoError(t, err)
			require.Equal(t, tt.want, gotMap)
		})
	}
}

func suiteSetFloat(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		value    float64
		want     map[string]any
		wantErr  bool
		err      error
	}{
		{
			name:     "simple float",
			jsonData: `{"price": 19.99}`,
			keys:     []string{"price"},
			value:    29.99,
			want:     map[string]any{"price": 29.99},
			wantErr:  false,
		},
		{
			name:     "nested float",
			jsonData: `{"product": {"price": 19.99}}`,
			keys:     []string{"product", "price"},
			value:    29.99,
			want:     map[string]any{"product": map[string]any{"price": 29.99}},
			wantErr:  false,
		},
		{
			name:     "new float key",
			jsonData: `{"name": "Product"}`,
			keys:     []string{"price"},
			value:    29.99,
			want:     map[string]any{"name": "Product", "price": 29.99},
			wantErr:  false,
		},
		{
			name:     "array float",
			jsonData: `{"prices": [10.5, 20.75]}`,
			keys:     []string{"prices", "[1]"},
			value:    29.99,
			want:     map[string]any{"prices": []any{10.5, 29.99}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"price"},
			value:    29.99,
			want:     map[string]any{"price": 29.99},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"price": 19.99}`,
			keys:     []string{""},
			value:    29.99,
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			value:    29.99,
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetFloat([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			var gotMap map[string]any
			err = jsonrs.Unmarshal(got, &gotMap)
			require.NoError(t, err)
			require.Equal(t, tt.want, gotMap)
		})
	}
}

func suiteSetString(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		value    string
		want     map[string]any
		wantErr  bool
		err      error
	}{
		{
			name:     "simple string",
			jsonData: `{"name": "John"}`,
			keys:     []string{"name"},
			value:    "Jane",
			want:     map[string]any{"name": "Jane"},
			wantErr:  false,
		},
		{
			name:     "nested string",
			jsonData: `{"user": {"name": "John"}}`,
			keys:     []string{"user", "name"},
			value:    "Jane",
			want:     map[string]any{"user": map[string]any{"name": "Jane"}},
			wantErr:  false,
		},
		{
			name:     "new string key",
			jsonData: `{"age": 30}`,
			keys:     []string{"name"},
			value:    "John",
			want:     map[string]any{"age": float64(30), "name": "John"},
			wantErr:  false,
		},
		{
			name:     "array string",
			jsonData: `{"names": ["John", "Bob"]}`,
			keys:     []string{"names", "[1]"},
			value:    "Jane",
			want:     map[string]any{"names": []any{"John", "Jane"}},
			wantErr:  false,
		},
		{
			name:     "string with embedded quote",
			jsonData: `{}`,
			keys:     []string{"name"},
			value:    `Jane "JJ" Doe`,
			want:     map[string]any{"name": `Jane "JJ" Doe`},
			wantErr:  false,
		},
		{
			name:     "string with embedded newline",
			jsonData: `{}`,
			keys:     []string{"name"},
			value:    "Jane\nDoe",
			want:     map[string]any{"name": "Jane\nDoe"},
			wantErr:  false,
		},
		{
			name:     "string with embedded backslash",
			jsonData: `{}`,
			keys:     []string{"path"},
			value:    `C:\Users\file`,
			want:     map[string]any{"path": `C:\Users\file`},
			wantErr:  false,
		},
		{
			name:     "string with embedded tab",
			jsonData: `{}`,
			keys:     []string{"text"},
			value:    "col1\tcol2",
			want:     map[string]any{"text": "col1\tcol2"},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			value:    "John",
			want:     map[string]any{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			value:    "Jane",
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "literal special key star",
			jsonData: `{}`,
			keys:     []string{"obj", "a*b"},
			value:    "literal-star",
			want:     map[string]any{"obj": map[string]any{"a*b": "literal-star"}},
			wantErr:  false,
		},
		{
			name:     "literal special key question",
			jsonData: `{}`,
			keys:     []string{"obj", "a?b"},
			value:    "literal-question",
			want:     map[string]any{"obj": map[string]any{"a?b": "literal-question"}},
			wantErr:  false,
		},
		{
			name:     "literal special key modifier",
			jsonData: `{}`,
			keys:     []string{"obj", "@reverse"},
			value:    "literal-modifier",
			want:     map[string]any{"obj": map[string]any{"@reverse": "literal-modifier"}},
			wantErr:  false,
		},
		{
			name:     "literal special key pipe",
			jsonData: `{}`,
			keys:     []string{"obj", "a|b"},
			value:    "literal-pipe",
			want:     map[string]any{"obj": map[string]any{"a|b": "literal-pipe"}},
			wantErr:  false,
		},
		{
			name:     "literal special key hash",
			jsonData: `{}`,
			keys:     []string{"obj", "a#b"},
			value:    "literal-hash",
			want:     map[string]any{"obj": map[string]any{"a#b": "literal-hash"}},
			wantErr:  false,
		},
		{
			name:     "literal special key colon",
			jsonData: `{}`,
			keys:     []string{"obj", "a:b"},
			value:    "literal-colon",
			want:     map[string]any{"obj": map[string]any{"a:b": "literal-colon"}},
			wantErr:  false,
		},
		{
			name:     "numeric object key",
			jsonData: `{}`,
			keys:     []string{"obj", "0", "x"},
			value:    "object-key",
			want:     map[string]any{"obj": map[string]any{"0": map[string]any{"x": "object-key"}}},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetString([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			var gotMap map[string]any
			err = jsonrs.Unmarshal(got, &gotMap)
			require.NoError(t, err)
			require.Equal(t, tt.want, gotMap)
		})
	}
}

func suiteDeleteKey(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     map[string]any
		wantErr  bool
		err      error
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			keys:     []string{"name"},
			want:     map[string]any{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user", "name"},
			want:     map[string]any{"user": map[string]any{"age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "array element",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			keys:     []string{"users", "[1]"},
			want:     map[string]any{"users": []any{"John", "Bob"}},
			wantErr:  false,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			want:     map[string]any{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyJSON,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			want:     nil,
			wantErr:  true,
			err:      ErrEmptyKey,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
			err:      ErrNoKeysProvided,
		},
		{
			name:     "null",
			jsonData: `{"user": {"name": "John", "age": null}}`,
			keys:     []string{"user", "age"},
			want:     map[string]any{"user": map[string]any{"name": "John"}},
			wantErr:  false,
		},
		{
			name:     "literal special key",
			jsonData: `{"obj": {"@reverse": "literal-modifier", "a|b": "literal-pipe"}}`,
			keys:     []string{"obj", "@reverse"},
			want:     map[string]any{"obj": map[string]any{"a|b": "literal-pipe"}},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.DeleteKey([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.err != nil {
					require.ErrorIs(t, err, tt.err)
				}
				return
			}
			require.NoError(t, err)
			var gotMap map[string]any
			err = jsonrs.Unmarshal(got, &gotMap)
			require.NoError(t, err)
			require.Equal(t, tt.want, gotMap)
		})
	}
}

func suiteSoftGetter(t *testing.T, jsonParser JSONParser) {
	// Test GetValueOrEmpty
	value := jsonParser.GetValueOrEmpty([]byte(`{"name": "John"}`), "name")
	require.Equal(t, []byte(`"John"`), value)

	// Test GetBooleanOrFalse
	boolVal := jsonParser.GetBooleanOrFalse([]byte(`{"active": true}`), "active")
	require.True(t, boolVal)
	boolVal = jsonParser.GetBooleanOrFalse([]byte(`{"active": "true"}`), "active")
	require.True(t, boolVal)
	boolVal = jsonParser.GetBooleanOrFalse([]byte(`{"active": 30}`), "active")
	require.True(t, boolVal)
	boolVal = jsonParser.GetBooleanOrFalse([]byte(`{"active": "random"}`), "active")
	require.False(t, boolVal)
	boolVal = jsonParser.GetBooleanOrFalse([]byte(`{"active": null}`), "active")
	require.False(t, boolVal)

	// Test GetIntOrZero
	intVal := jsonParser.GetIntOrZero([]byte(`{"age": 30}`), "age")
	require.Equal(t, int64(30), intVal)
	intVal = jsonParser.GetIntOrZero([]byte(`{"age": "30"}`), "age")
	require.Equal(t, int64(30), intVal)
	intVal = jsonParser.GetIntOrZero([]byte(`{"age": true}`), "age")
	require.Equal(t, int64(1), intVal)
	intVal = jsonParser.GetIntOrZero([]byte(`{"age": "random"}`), "age")
	require.Equal(t, int64(0), intVal)
	intVal = jsonParser.GetIntOrZero([]byte(`{"age": null}`), "age")
	require.Equal(t, int64(0), intVal)
	intVal = jsonParser.GetIntOrZero([]byte(`{"age": 2.9}`), "age")
	require.Equal(t, int64(2), intVal)
	intVal = jsonParser.GetIntOrZero([]byte(`{"age": -2.9}`), "age")
	require.Equal(t, int64(-2), intVal)
	intVal = jsonParser.GetIntOrZero([]byte(`{"age": 1e3}`), "age")
	require.Equal(t, int64(1000), intVal)

	// Test GetFloatOrZero
	floatVal := jsonParser.GetFloatOrZero([]byte(`{"price": 29.99}`), "price")
	require.Equal(t, 29.99, floatVal)
	floatVal = jsonParser.GetFloatOrZero([]byte(`{"price": "29.99"}`), "price")
	require.Equal(t, 29.99, floatVal)
	floatVal = jsonParser.GetFloatOrZero([]byte(`{"price": true}`), "price")
	require.Equal(t, float64(1), floatVal)
	floatVal = jsonParser.GetFloatOrZero([]byte(`{"price": "random"}`), "price")
	require.Equal(t, float64(0), floatVal)
	floatVal = jsonParser.GetFloatOrZero([]byte(`{"price": null}`), "price")
	require.Equal(t, float64(0), floatVal)

	// Test GetStringOrEmpty
	strVal := jsonParser.GetStringOrEmpty([]byte(`{"city": "New York"}`), "city")
	require.Equal(t, "New York", strVal)
	strVal = jsonParser.GetStringOrEmpty([]byte(`{"city": 30}`), "city")
	require.Equal(t, "30", strVal)
	strVal = jsonParser.GetStringOrEmpty([]byte(`{"city": true}`), "city")
	require.Equal(t, "true", strVal)
	strVal = jsonParser.GetStringOrEmpty([]byte(`{"city": {"New": "York"}}`), "city")
	require.Equal(t, `{"New": "York"}`, strVal)
	strVal = jsonParser.GetStringOrEmpty([]byte(`{"city": null}`), "city")
	require.Equal(t, "", strVal)

	// Test that non-existent keys return zero values
	require.Nil(t, jsonParser.GetValueOrEmpty([]byte(`{}`), "missing"))
	require.False(t, jsonParser.GetBooleanOrFalse([]byte(`{}`), "missing"))
	require.Zero(t, jsonParser.GetIntOrZero([]byte(`{}`), "missing"))
	require.Zero(t, jsonParser.GetFloatOrZero([]byte(`{}`), "missing"))
	require.Empty(t, jsonParser.GetStringOrEmpty([]byte(`{}`), "missing"))

	// Test that empty data return zero values
	require.Nil(t, jsonParser.GetValueOrEmpty([]byte{}, "missing"))
	require.False(t, jsonParser.GetBooleanOrFalse([]byte{}, "missing"))
	require.Zero(t, jsonParser.GetIntOrZero([]byte{}, "missing"))
	require.Zero(t, jsonParser.GetFloatOrZero([]byte{}, "missing"))
	require.Empty(t, jsonParser.GetStringOrEmpty([]byte{}, "missing"))

	// Test that no keys provided return zero values
	require.Nil(t, jsonParser.GetValueOrEmpty([]byte(`{}`)))
	require.False(t, jsonParser.GetBooleanOrFalse([]byte(`{}`)))
	require.Zero(t, jsonParser.GetIntOrZero([]byte(`{}`)))
	require.Zero(t, jsonParser.GetFloatOrZero([]byte(`{}`)))
	require.Empty(t, jsonParser.GetStringOrEmpty([]byte(`{}`)))

	// Test that empty keys return zero values
	require.Nil(t, jsonParser.GetValueOrEmpty([]byte(`{"city": 30}`), ""))
	require.False(t, jsonParser.GetBooleanOrFalse([]byte(`{"city": 30}`), ""))
	require.Zero(t, jsonParser.GetIntOrZero([]byte(`{"city": 30}`), ""))
	require.Zero(t, jsonParser.GetFloatOrZero([]byte(`{"city": 30}`), ""))
	require.Empty(t, jsonParser.GetStringOrEmpty([]byte(`{"city": 30}`), ""))
}

// suiteGJSONFeaturesNotSupported verifies that advanced GJSON path syntax
// features are NOT interpreted as special syntax by our jsonparser library.
// All special characters in path segments must be treated as literal key names.
// Each test uses a JSON payload that contains the special-syntax string as a
// real object key so the parser must return the literal value, proving it
// never activates the GJSON feature.
func suiteGJSONFeaturesNotSupported(t *testing.T, jsonParser JSONParser) {
	t.Run("wildcards are not pattern matchers", func(t *testing.T) {
		data := []byte(`{
			"child*": {"2": "literal-child*"},
			"children": ["Sara", "Alex", "Jack"],
			"c?ildren": ["Literal-Sara"],
			"*": "literal-star",
			"ag?": "literal-ag?"
		}`)
		// In GJSON: "child*.2" would match "children" and return "Jack"
		// Our parser treats "child*" as a literal key
		got, err := jsonParser.GetString(data, "child*", "2")
		require.NoError(t, err)
		require.Equal(t, "literal-child*", got)

		// In GJSON: "c?ildren.[0]" would match "children" and return "Sara"
		got, err = jsonParser.GetString(data, "c?ildren", "[0]")
		require.NoError(t, err)
		require.Equal(t, "Literal-Sara", got)

		// In GJSON: "*" would match first key
		got, err = jsonParser.GetString(data, "*")
		require.NoError(t, err)
		require.Equal(t, "literal-star", got)

		// In GJSON: "ag?" would match "age"
		got, err = jsonParser.GetString(data, "ag?")
		require.NoError(t, err)
		require.Equal(t, "literal-ag?", got)
	})

	t.Run("hash does not return array length", func(t *testing.T) {
		data := []byte(`{
			"children": {"#": "literal-hash"},
			"friends": {"#": "literal-friends-hash"}
		}`)
		// In GJSON: "children.#" returns 3 (array length)
		got, err := jsonParser.GetString(data, "children", "#")
		require.NoError(t, err)
		require.Equal(t, "literal-hash", got)

		got, err = jsonParser.GetString(data, "friends", "#")
		require.NoError(t, err)
		require.Equal(t, "literal-friends-hash", got)
	})

	t.Run("hash does not iterate array members", func(t *testing.T) {
		data := []byte(`{
			"friends": {
				"#": {"age": "literal-hash-age", "first": "literal-hash-first"}
			}
		}`)
		// In GJSON: "friends.#.age" returns [44,68,47]
		got, err := jsonParser.GetString(data, "friends", "#", "age")
		require.NoError(t, err)
		require.Equal(t, "literal-hash-age", got)

		got, err = jsonParser.GetString(data, "friends", "#", "first")
		require.NoError(t, err)
		require.Equal(t, "literal-hash-first", got)
	})

	t.Run("queries are not supported", func(t *testing.T) {
		data := []byte(`{
			"friends": {
				"#(last==Murphy)": {"first": "literal-query-single"},
				"#(last==Murphy)#": {"first": "literal-query-multi"},
				"#(age>45)#": {"last": "literal-gt-query"},
				"#(first%D*)": {"last": "literal-like-query"},
				"#(first!%D*)": {"last": "literal-notlike-query"}
			}
		}`)
		got, err := jsonParser.GetString(data, "friends", "#(last==Murphy)", "first")
		require.NoError(t, err)
		require.Equal(t, "literal-query-single", got)

		got, err = jsonParser.GetString(data, "friends", "#(last==Murphy)#", "first")
		require.NoError(t, err)
		require.Equal(t, "literal-query-multi", got)

		got, err = jsonParser.GetString(data, "friends", "#(age>45)#", "last")
		require.NoError(t, err)
		require.Equal(t, "literal-gt-query", got)

		got, err = jsonParser.GetString(data, "friends", "#(first%D*)", "last")
		require.NoError(t, err)
		require.Equal(t, "literal-like-query", got)

		got, err = jsonParser.GetString(data, "friends", "#(first!%D*)", "last")
		require.NoError(t, err)
		require.Equal(t, "literal-notlike-query", got)
	})

	t.Run("nested queries are not supported", func(t *testing.T) {
		data := []byte(`{
			"friends": {
				"#(nets.#(==fb))#": {"first": "literal-nested-query"}
			}
		}`)
		got, err := jsonParser.GetString(data, "friends", "#(nets.#(==fb))#", "first")
		require.NoError(t, err)
		require.Equal(t, "literal-nested-query", got)
	})

	t.Run("tilde comparison operator not supported", func(t *testing.T) {
		data := []byte(`{
			"vals": {
				"#(b==~true)#": {"a": "literal-tilde-true"},
				"#(b==~false)#": {"a": "literal-tilde-false"},
				"#(b==~null)#": {"a": "literal-tilde-null"},
				"#(b==~*)#": {"a": "literal-tilde-star"}
			}
		}`)
		got, err := jsonParser.GetString(data, "vals", "#(b==~true)#", "a")
		require.NoError(t, err)
		require.Equal(t, "literal-tilde-true", got)

		got, err = jsonParser.GetString(data, "vals", "#(b==~false)#", "a")
		require.NoError(t, err)
		require.Equal(t, "literal-tilde-false", got)

		got, err = jsonParser.GetString(data, "vals", "#(b==~null)#", "a")
		require.NoError(t, err)
		require.Equal(t, "literal-tilde-null", got)

		got, err = jsonParser.GetString(data, "vals", "#(b==~*)#", "a")
		require.NoError(t, err)
		require.Equal(t, "literal-tilde-star", got)
	})

	t.Run("pipe operator does not work as path separator", func(t *testing.T) {
		data := []byte(`{
			"friends|0|first": "literal-pipe-path",
			"name|first": "literal-pipe-name",
			"friends": {
				"#(last=Murphy)#|0": "literal-pipe-query"
			}
		}`)
		// In GJSON: "friends|0|first" is equivalent to "friends.0.first"
		got, err := jsonParser.GetString(data, "friends|0|first")
		require.NoError(t, err)
		require.Equal(t, "literal-pipe-path", got)

		// In GJSON: "name|first" would traverse into name.first
		got, err = jsonParser.GetString(data, "name|first")
		require.NoError(t, err)
		require.Equal(t, "literal-pipe-name", got)

		got, err = jsonParser.GetString(data, "friends", "#(last=Murphy)#|0")
		require.NoError(t, err)
		require.Equal(t, "literal-pipe-query", got)
	})

	t.Run("modifiers are not supported", func(t *testing.T) {
		data := []byte(`{
			"children": {"@reverse": "literal-reverse"},
			"@ugly": "literal-ugly",
			"@pretty": "literal-pretty",
			"@this": "literal-this",
			"@valid": "literal-valid",
			"friends": {"@flatten": "literal-flatten", "@join": "literal-join", "@group": "literal-group"},
			"name": {"@keys": "literal-keys", "@values": "literal-values", "@tostr": "literal-tostr"},
			"@fromstr": "literal-fromstr",
			"@dig": {"first": "literal-dig"}
		}`)
		got, err := jsonParser.GetString(data, "children", "@reverse")
		require.NoError(t, err)
		require.Equal(t, "literal-reverse", got)

		got, err = jsonParser.GetString(data, "@ugly")
		require.NoError(t, err)
		require.Equal(t, "literal-ugly", got)

		got, err = jsonParser.GetString(data, "@pretty")
		require.NoError(t, err)
		require.Equal(t, "literal-pretty", got)

		// In GJSON: @this returns the root element
		got, err = jsonParser.GetString(data, "@this")
		require.NoError(t, err)
		require.Equal(t, "literal-this", got)

		got, err = jsonParser.GetString(data, "@valid")
		require.NoError(t, err)
		require.Equal(t, "literal-valid", got)

		got, err = jsonParser.GetString(data, "friends", "@flatten")
		require.NoError(t, err)
		require.Equal(t, "literal-flatten", got)

		got, err = jsonParser.GetString(data, "name", "@keys")
		require.NoError(t, err)
		require.Equal(t, "literal-keys", got)

		got, err = jsonParser.GetString(data, "name", "@values")
		require.NoError(t, err)
		require.Equal(t, "literal-values", got)

		got, err = jsonParser.GetString(data, "friends", "@join")
		require.NoError(t, err)
		require.Equal(t, "literal-join", got)

		got, err = jsonParser.GetString(data, "name", "@tostr")
		require.NoError(t, err)
		require.Equal(t, "literal-tostr", got)

		got, err = jsonParser.GetString(data, "@fromstr")
		require.NoError(t, err)
		require.Equal(t, "literal-fromstr", got)

		got, err = jsonParser.GetString(data, "friends", "@group")
		require.NoError(t, err)
		require.Equal(t, "literal-group", got)

		got, err = jsonParser.GetString(data, "@dig", "first")
		require.NoError(t, err)
		require.Equal(t, "literal-dig", got)
	})

	t.Run("modifier arguments are not supported", func(t *testing.T) {
		data := []byte(`{
			"@pretty:indent": "literal-pretty-arg",
			"children": {"@case:upper": "literal-case-arg"}
		}`)
		got, err := jsonParser.GetString(data, "@pretty:indent")
		require.NoError(t, err)
		require.Equal(t, "literal-pretty-arg", got)

		got, err = jsonParser.GetString(data, "children", "@case:upper")
		require.NoError(t, err)
		require.Equal(t, "literal-case-arg", got)
	})

	t.Run("multipaths are not supported", func(t *testing.T) {
		data := []byte(`{
			"{first,age}": "literal-multipath-obj",
			"{name.first,age}": "literal-multipath-dotted"
		}`)
		// In GJSON: {first,age} creates a new JSON array from multiple paths
		got, err := jsonParser.GetString(data, "{first,age}")
		require.NoError(t, err)
		require.Equal(t, "literal-multipath-obj", got)

		// In GJSON: {name.first,age} creates a new JSON object
		got, err = jsonParser.GetString(data, "{name.first,age}")
		require.NoError(t, err)
		require.Equal(t, "literal-multipath-dotted", got)
	})

	t.Run("literals are not supported", func(t *testing.T) {
		data := []byte(`{
			"!true": "literal-bang-true",
			"!false": "literal-bang-false"
		}`)
		// In GJSON: !true is a literal boolean value injected into the result
		got, err := jsonParser.GetString(data, "!true")
		require.NoError(t, err)
		require.Equal(t, "literal-bang-true", got)

		got, err = jsonParser.GetString(data, "!false")
		require.NoError(t, err)
		require.Equal(t, "literal-bang-false", got)
	})

	t.Run("dot in single path segment is literal not path separator", func(t *testing.T) {
		data := []byte(`{
			"fav.movie": "Deer Hunter",
			"fav": {"movie": "nested-movie"}
		}`)
		// A single path segment "fav.movie" matches the literal key "fav.movie",
		// it does NOT split on dots into separate path segments.
		// Use separate segments "fav", "movie" to navigate nested paths.
		got, err := jsonParser.GetString(data, "fav.movie")
		require.NoError(t, err)
		require.Equal(t, "Deer Hunter", got)

		// Separate segments navigate the nested path
		got, err = jsonParser.GetString(data, "fav", "movie")
		require.NoError(t, err)
		require.Equal(t, "nested-movie", got)
	})

	t.Run("SetValue treats special syntax as literal keys", func(t *testing.T) {
		t.Run("wildcard creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "a*b")
			require.NoError(t, err)
			// Should create a key literally named "a*b"
			got, err := jsonParser.GetString(result, "a*b")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("question mark creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "a?b")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "a?b")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("hash creates literal key not array length", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "items", "#")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "items", "#")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("modifier syntax creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "@reverse")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "@reverse")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("pipe creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "a|b")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "a|b")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("query syntax creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{"items": {}}`), "val", "items", "#(age>45)")
			require.NoError(t, err)
			// Should create a literal key, not query into any array
			got, err := jsonParser.GetString(result, "items", "#(age>45)")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("modifier with argument creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "@pretty:indent")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "@pretty:indent")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("exclamation mark creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "!true")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "!true")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("colon creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "a:b")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "a:b")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})

		t.Run("dot creates literal key", func(t *testing.T) {
			result, err := jsonParser.SetValue([]byte(`{}`), "val", "a.b")
			require.NoError(t, err)
			got, err := jsonParser.GetString(result, "a.b")
			require.NoError(t, err)
			require.Equal(t, "val", got)
		})
	})

	t.Run("DeleteKey treats special syntax as literal keys", func(t *testing.T) {
		t.Run("delete wildcard key", func(t *testing.T) {
			data := []byte(`{"a*b": "val", "other": "keep"}`)
			result, err := jsonParser.DeleteKey(data, "a*b")
			require.NoError(t, err)
			// "a*b" should be deleted, "other" should remain
			_, err = jsonParser.GetString(result, "a*b")
			require.ErrorIs(t, err, ErrKeyNotFound)
			got, err := jsonParser.GetString(result, "other")
			require.NoError(t, err)
			require.Equal(t, "keep", got)
		})

		t.Run("delete hash key", func(t *testing.T) {
			data := []byte(`{"items": {"#": "val", "other": "keep"}}`)
			result, err := jsonParser.DeleteKey(data, "items", "#")
			require.NoError(t, err)
			_, err = jsonParser.GetString(result, "items", "#")
			require.ErrorIs(t, err, ErrKeyNotFound)
			got, err := jsonParser.GetString(result, "items", "other")
			require.NoError(t, err)
			require.Equal(t, "keep", got)
		})

		t.Run("delete modifier key", func(t *testing.T) {
			data := []byte(`{"@reverse": "val", "other": "keep"}`)
			result, err := jsonParser.DeleteKey(data, "@reverse")
			require.NoError(t, err)
			_, err = jsonParser.GetString(result, "@reverse")
			require.ErrorIs(t, err, ErrKeyNotFound)
			got, err := jsonParser.GetString(result, "other")
			require.NoError(t, err)
			require.Equal(t, "keep", got)
		})

		t.Run("delete pipe key", func(t *testing.T) {
			data := []byte(`{"a|b": "val", "other": "keep"}`)
			result, err := jsonParser.DeleteKey(data, "a|b")
			require.NoError(t, err)
			_, err = jsonParser.GetString(result, "a|b")
			require.ErrorIs(t, err, ErrKeyNotFound)
			got, err := jsonParser.GetString(result, "other")
			require.NoError(t, err)
			require.Equal(t, "keep", got)
		})
	})
}

func TestJsonParser(t *testing.T) {
	libs := []string{TidwallLib, GrafanaLib}
	for _, lib := range libs {
		t.Run(lib, func(t *testing.T) {
			jsonParser := NewWithLibrary(lib)
			t.Run("GetValue", func(t *testing.T) {
				suiteGetValue(t, jsonParser)
			})
			t.Run("SetValue", func(t *testing.T) {
				suiteSetValue(t, jsonParser)
			})
			t.Run("GetBoolean", func(t *testing.T) {
				suiteGetBoolean(t, jsonParser)
			})
			t.Run("SetBoolean", func(t *testing.T) {
				suiteSetBoolean(t, jsonParser)
			})
			t.Run("GetInt", func(t *testing.T) {
				suiteGetInt(t, jsonParser)
			})
			t.Run("SetInt", func(t *testing.T) {
				suiteSetInt(t, jsonParser)
			})
			t.Run("GetFloat", func(t *testing.T) {
				suiteGetFloat(t, jsonParser)
			})
			t.Run("SetFloat", func(t *testing.T) {
				suiteSetFloat(t, jsonParser)
			})
			t.Run("GetString", func(t *testing.T) {
				suiteGetString(t, jsonParser)
			})
			t.Run("SetString", func(t *testing.T) {
				suiteSetString(t, jsonParser)
			})
			t.Run("DeleteKey", func(t *testing.T) {
				suiteDeleteKey(t, jsonParser)
			})
			t.Run("EdgeCases", func(t *testing.T) {
				suiteEdgeCases(t, jsonParser)
			})
			t.Run("ArrayHandling", func(t *testing.T) {
				suiteArrayHandling(t, jsonParser)
			})
			t.Run("SoftGetter", func(t *testing.T) {
				suiteSoftGetter(t, jsonParser)
			})
			t.Run("GJSONFeaturesNotSupported", func(t *testing.T) {
				suiteGJSONFeaturesNotSupported(t, jsonParser)
			})
		})
	}
}
