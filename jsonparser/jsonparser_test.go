package jsonparser

import (
	"testing"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/stretchr/testify/require"
)

func suiteGetValue(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     interface{}
		wantErr  bool
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			keys:     []string{"name"},
			want:     []byte("\"John\""),
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user", "name"},
			want:     []byte("\"John\""),
			wantErr:  false,
		},
		{
			name:     "json bytes",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user"},
			want:     []byte("{\"name\": \"John\", \"age\": 30}"),
			wantErr:  false,
		},
		{
			name:     "array index",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			keys:     []string{"users", "[1]"},
			want:     []byte("\"Jane\""),
			wantErr:  false,
		},
		{
			name:     "array bytes",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			keys:     []string{"users"},
			want:     []byte("[\"John\", \"Jane\", \"Bob\"]"),
			wantErr:  false,
		},
		{
			name:     "nested array",
			jsonData: `{"data": {"users": [{"name": "John"}, {"name": "Jane"}]}}`,
			keys:     []string{"data", "users", "[1]", "name"},
			want:     []byte("\"Jane\""),
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
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			wantErr:  true,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "no keys provided",
			jsonData: `{"name": "John"}`,
			keys:     []string{},
			wantErr:  true,
		},
		{
			name:     "invalid array index",
			jsonData: `{"users": ["John", "Jane"]}`,
			keys:     []string{"users", "[2]"},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid json",
			jsonData: `{"name": "John"`,
			keys:     []string{"name"},
			want:     []byte("\"John\""),
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid path empty",
			jsonData: `{"user": {"active": true}}`,
			keys:     []string{"user", "", "active"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetValue([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
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
		value    interface{}
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			keys:     []string{"name"},
			value:    "Jane",
			want:     map[string]interface{}{"name": "Jane", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user", "name"},
			value:    "Jane",
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "Jane", "age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "new key",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]interface{}{"name": "John", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "new nested key",
			jsonData: `{"user": {"name": "John"}}`,
			keys:     []string{"user", "age"},
			value:    30,
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "John", "age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "create nested structure",
			jsonData: `{}`,
			keys:     []string{"user", "name"},
			value:    "John",
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
			wantErr:  false,
		},
		{
			name:     "array element",
			jsonData: `{"users": ["John", "Jane"]}`,
			keys:     []string{"users", "[1]"},
			value:    "Bob",
			want:     map[string]interface{}{"users": []interface{}{"John", "Bob"}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			value:    "John",
			want:     map[string]interface{}{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			value:    "value",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"name": "John"}`,
			keys:     []string{},
			value:    "value",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetValue([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var gotMap map[string]interface{}
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
	require.Equal(t, []byte("\"Jane\""), value)
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
	require.Equal(t, []byte("\"john@example.com\""), value)

	// Test setting a deeply nested value
	updatedJSON, err := jsonParser.SetValue(jsonData, "123-456-7890", "data", "users", "0", "profile", "details", "contact", "phone")
	require.NoError(t, err)

	// Verify the update
	value, err = jsonParser.GetValue(updatedJSON, "data", "users", "0", "profile", "details", "contact", "phone")
	require.NoError(t, err)
	require.Equal(t, []byte("\"123-456-7890\""), value)
}

func suiteGetBoolean(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		keys     []string
		want     bool
		wantErr  bool
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
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"active"},
			want:     false,
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"active"},
			want:     false,
			wantErr:  true,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     false,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetBoolean([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
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
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			want:     0,
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"age"},
			want:     0,
			wantErr:  true,
		},
		{
			name:     "float value",
			jsonData: `{"age": 2.32}`,
			keys:     []string{"age"},
			want:     2,
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     0,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetInt([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
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
		},
		{
			name:     "key not found",
			jsonData: `{"name": "Product"}`,
			keys:     []string{"price"},
			want:     0,
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"price"},
			want:     0,
			wantErr:  true,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     0,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetFloat([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
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
		},
		{
			name:     "key not found",
			jsonData: `{"age": 30}`,
			keys:     []string{"name"},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			want:     "",
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetString([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
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
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple boolean true",
			jsonData: `{"active": false}`,
			keys:     []string{"active"},
			value:    true,
			want:     map[string]interface{}{"active": true},
			wantErr:  false,
		},
		{
			name:     "simple boolean false",
			jsonData: `{"active": true}`,
			keys:     []string{"active"},
			value:    false,
			want:     map[string]interface{}{"active": false},
			wantErr:  false,
		},
		{
			name:     "nested boolean",
			jsonData: `{"user": {"active": false}}`,
			keys:     []string{"user", "active"},
			value:    true,
			want:     map[string]interface{}{"user": map[string]interface{}{"active": true}},
			wantErr:  false,
		},
		{
			name:     "new boolean key",
			jsonData: `{"name": "John"}`,
			keys:     []string{"active"},
			value:    true,
			want:     map[string]interface{}{"name": "John", "active": true},
			wantErr:  false,
		},
		{
			name:     "array boolean",
			jsonData: `{"settings": [false, false]}`,
			keys:     []string{"settings", "[0]"},
			value:    true,
			want:     map[string]interface{}{"settings": []interface{}{true, false}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"active"},
			value:    true,
			want:     map[string]interface{}{"active": true},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			value:    false,
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			keys:     []string{""},
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetBoolean([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var gotMap map[string]interface{}
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
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple integer",
			jsonData: `{"age": 25}`,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]interface{}{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested integer",
			jsonData: `{"user": {"age": 25}}`,
			keys:     []string{"user", "age"},
			value:    30,
			want:     map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "new integer key",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]interface{}{"name": "John", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "array integer",
			jsonData: `{"ages": [10, 20]}`,
			keys:     []string{"ages", "[1]"},
			value:    30,
			want:     map[string]interface{}{"ages": []interface{}{float64(10), float64(30)}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"age"},
			value:    30,
			want:     map[string]interface{}{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"age": 25}`,
			keys:     []string{""},
			value:    30,
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetInt([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var gotMap map[string]interface{}
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
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple float",
			jsonData: `{"price": 19.99}`,
			keys:     []string{"price"},
			value:    29.99,
			want:     map[string]interface{}{"price": 29.99},
			wantErr:  false,
		},
		{
			name:     "nested float",
			jsonData: `{"product": {"price": 19.99}}`,
			keys:     []string{"product", "price"},
			value:    29.99,
			want:     map[string]interface{}{"product": map[string]interface{}{"price": 29.99}},
			wantErr:  false,
		},
		{
			name:     "new float key",
			jsonData: `{"name": "Product"}`,
			keys:     []string{"price"},
			value:    29.99,
			want:     map[string]interface{}{"name": "Product", "price": 29.99},
			wantErr:  false,
		},
		{
			name:     "array float",
			jsonData: `{"prices": [10.5, 20.75]}`,
			keys:     []string{"prices", "[1]"},
			value:    29.99,
			want:     map[string]interface{}{"prices": []interface{}{10.5, 29.99}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"price"},
			value:    29.99,
			want:     map[string]interface{}{"price": 29.99},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"price": 19.99}`,
			keys:     []string{""},
			value:    29.99,
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			value:    29.99,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetFloat([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var gotMap map[string]interface{}
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
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple string",
			jsonData: `{"name": "John"}`,
			keys:     []string{"name"},
			value:    "Jane",
			want:     map[string]interface{}{"name": "Jane"},
			wantErr:  false,
		},
		{
			name:     "nested string",
			jsonData: `{"user": {"name": "John"}}`,
			keys:     []string{"user", "name"},
			value:    "Jane",
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "Jane"}},
			wantErr:  false,
		},
		{
			name:     "new string key",
			jsonData: `{"age": 30}`,
			keys:     []string{"name"},
			value:    "John",
			want:     map[string]interface{}{"age": float64(30), "name": "John"},
			wantErr:  false,
		},
		{
			name:     "array string",
			jsonData: `{"names": ["John", "Bob"]}`,
			keys:     []string{"names", "[1]"},
			value:    "Jane",
			want:     map[string]interface{}{"names": []interface{}{"John", "Jane"}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			value:    "John",
			want:     map[string]interface{}{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			value:    "Jane",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetString([]byte(tt.jsonData), tt.value, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var gotMap map[string]interface{}
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
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			keys:     []string{"name"},
			want:     map[string]interface{}{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			keys:     []string{"user", "name"},
			want:     map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "array element",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			keys:     []string{"users", "[1]"},
			want:     map[string]interface{}{"users": []interface{}{"John", "Bob"}},
			wantErr:  false,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			keys:     []string{"age"},
			want:     map[string]interface{}{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			keys:     []string{"name"},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			keys:     []string{""},
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "no key",
			jsonData: `{"active": true}`,
			keys:     []string{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.DeleteKey([]byte(tt.jsonData), tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var gotMap map[string]interface{}
			err = jsonrs.Unmarshal(got, &gotMap)
			require.NoError(t, err)
			require.Equal(t, tt.want, gotMap)
		})
	}
}

func suiteSoftGetter(t *testing.T, jsonParser JSONParser) {
	// Test GetValueOrEmpty
	value := jsonParser.GetValueOrEmpty([]byte(`{"name": "John"}`), "name")
	require.Equal(t, []byte("\"John\""), value)

	// Test GetBooleanOrFalse
	boolVal := jsonParser.GetBooleanOrFalse([]byte(`{"active": true}`), "active")
	require.True(t, boolVal)
	boolVal = jsonParser.GetBooleanOrFalse([]byte(`{"active": "true"}`), "active")
	require.True(t, boolVal)
	boolVal = jsonParser.GetBooleanOrFalse([]byte(`{"active": 30}`), "active")
	require.True(t, boolVal)
	boolVal = jsonParser.GetBooleanOrFalse([]byte(`{"active": "random"}`), "active")
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

	// Test GetFloatOrZero
	floatVal := jsonParser.GetFloatOrZero([]byte(`{"price": 29.99}`), "price")
	require.Equal(t, 29.99, floatVal)
	floatVal = jsonParser.GetFloatOrZero([]byte(`{"price": "29.99"}`), "price")
	require.Equal(t, 29.99, floatVal)
	floatVal = jsonParser.GetFloatOrZero([]byte(`{"price": true}`), "price")
	require.Equal(t, float64(1), floatVal)
	floatVal = jsonParser.GetFloatOrZero([]byte(`{"price": "random"}`), "price")
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
		})
	}
}
