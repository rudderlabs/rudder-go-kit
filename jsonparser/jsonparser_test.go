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
		key      string
		want     interface{}
		wantErr  bool
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			key:      "name",
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			key:      "user.name",
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "array index",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			key:      "users.1",
			want:     "Jane",
			wantErr:  false,
		},
		{
			name:     "nested array",
			jsonData: `{"data": {"users": [{"name": "John"}, {"name": "Jane"}]}}`,
			key:      "data.users.1.name",
			want:     "Jane",
			wantErr:  false,
		},
		{
			name:     "numeric value",
			jsonData: `{"user": {"age": 30}}`,
			key:      "user.age",
			want:     float64(30),
			wantErr:  false,
		},
		{
			name:     "boolean value",
			jsonData: `{"user": {"active": true}}`,
			key:      "user.active",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "null value",
			jsonData: `{"user": {"middleName": null}}`,
			key:      "user.middleName",
			want:     nil,
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			key:      "",
			want:     map[string]interface{}{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			key:      "age",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid array index",
			jsonData: `{"users": ["John", "Jane"]}`,
			key:      "users.2",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "invalid json",
			jsonData: `{"name": "John"`,
			key:      "name",
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "name",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetValue([]byte(tt.jsonData), tt.key)
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
		key      string
		value    interface{}
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			key:      "name",
			value:    "Jane",
			want:     map[string]interface{}{"name": "Jane", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			key:      "user.name",
			value:    "Jane",
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "Jane", "age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "new key",
			jsonData: `{"name": "John"}`,
			key:      "age",
			value:    30,
			want:     map[string]interface{}{"name": "John", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "new nested key",
			jsonData: `{"user": {"name": "John"}}`,
			key:      "user.age",
			value:    30,
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "John", "age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "create nested structure",
			jsonData: `{}`,
			key:      "user.name",
			value:    "John",
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
			wantErr:  false,
		},
		{
			name:     "array element",
			jsonData: `{"users": ["John", "Jane"]}`,
			key:      "users.1",
			value:    "Bob",
			want:     map[string]interface{}{"users": []interface{}{"John", "Bob"}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "name",
			value:    "John",
			want:     map[string]interface{}{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			key:      "",
			value:    "value",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetValue([]byte(tt.jsonData), tt.key, tt.value)
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
	value, err := jsonParser.GetValue(jsonData, "users.1.name")
	require.NoError(t, err)
	require.Equal(t, "Jane", value)

	// Test setting a value in an array of objects
	updatedJSON, err := jsonParser.SetValue(jsonData, "users.0.age", 31)
	require.NoError(t, err)

	// Verify the update
	value, err = jsonParser.GetValue(updatedJSON, "users.0.age")
	require.NoError(t, err)
	require.Equal(t, float64(31), value)
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
	value, err := jsonParser.GetValue(jsonData, "data.users.0.profile.details.contact.email")
	require.NoError(t, err)
	require.Equal(t, "john@example.com", value)

	// Test setting a deeply nested value
	updatedJSON, err := jsonParser.SetValue(jsonData, "data.users.0.profile.details.contact.phone", "123-456-7890")
	require.NoError(t, err)

	// Verify the update
	value, err = jsonParser.GetValue(updatedJSON, "data.users.0.profile.details.contact.phone")
	require.NoError(t, err)
	require.Equal(t, "123-456-7890", value)
}

func suiteGetBoolean(t *testing.T, jsonParser JSONParser) {
	tests := []struct {
		name     string
		jsonData string
		key      string
		want     bool
		wantErr  bool
	}{
		{
			name:     "simple boolean true",
			jsonData: `{"active": true}`,
			key:      "active",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "simple boolean false",
			jsonData: `{"active": false}`,
			key:      "active",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "nested boolean",
			jsonData: `{"user": {"active": true}}`,
			key:      "user.active",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "array boolean",
			jsonData: `{"settings": [true, false, true]}`,
			key:      "settings.0",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "non-boolean value",
			jsonData: `{"active": "true"}`,
			key:      "active",
			want:     false,
			wantErr:  true,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			key:      "active",
			want:     false,
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "active",
			want:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetBoolean([]byte(tt.jsonData), tt.key)
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
		key      string
		want     int64
		wantErr  bool
	}{
		{
			name:     "simple integer",
			jsonData: `{"age": 30}`,
			key:      "age",
			want:     30,
			wantErr:  false,
		},
		{
			name:     "nested integer",
			jsonData: `{"user": {"age": 30}}`,
			key:      "user.age",
			want:     30,
			wantErr:  false,
		},
		{
			name:     "array integer",
			jsonData: `{"ages": [10, 20, 30]}`,
			key:      "ages.2",
			want:     30,
			wantErr:  false,
		},
		{
			name:     "non-integer value",
			jsonData: `{"age": "30"}`,
			key:      "age",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			key:      "age",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "age",
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetInt([]byte(tt.jsonData), tt.key)
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
		key      string
		want     float64
		wantErr  bool
	}{
		{
			name:     "simple float",
			jsonData: `{"price": 29.99}`,
			key:      "price",
			want:     29.99,
			wantErr:  false,
		},
		{
			name:     "integer as float",
			jsonData: `{"price": 30}`,
			key:      "price",
			want:     30.0,
			wantErr:  false,
		},
		{
			name:     "nested float",
			jsonData: `{"product": {"price": 29.99}}`,
			key:      "product.price",
			want:     29.99,
			wantErr:  false,
		},
		{
			name:     "array float",
			jsonData: `{"prices": [10.5, 20.75, 30.99]}`,
			key:      "prices.1",
			want:     20.75,
			wantErr:  false,
		},
		{
			name:     "non-float value",
			jsonData: `{"price": "29.99"}`,
			key:      "price",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "Product"}`,
			key:      "price",
			want:     0,
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "price",
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetFloat([]byte(tt.jsonData), tt.key)
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
		key      string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple string",
			jsonData: `{"name": "John"}`,
			key:      "name",
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "nested string",
			jsonData: `{"user": {"name": "John"}}`,
			key:      "user.name",
			want:     "John",
			wantErr:  false,
		},
		{
			name:     "array string",
			jsonData: `{"names": ["John", "Jane", "Bob"]}`,
			key:      "names.1",
			want:     "Jane",
			wantErr:  false,
		},
		{
			name:     "non-string value",
			jsonData: `{"name": 123}`,
			key:      "name",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "key not found",
			jsonData: `{"age": 30}`,
			key:      "name",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "invalid json",
			jsonData: `{"name": "John"`,
			key:      "name",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "name",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.GetString([]byte(tt.jsonData), tt.key)
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
		key      string
		value    bool
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple boolean true",
			jsonData: `{"active": false}`,
			key:      "active",
			value:    true,
			want:     map[string]interface{}{"active": true},
			wantErr:  false,
		},
		{
			name:     "simple boolean false",
			jsonData: `{"active": true}`,
			key:      "active",
			value:    false,
			want:     map[string]interface{}{"active": false},
			wantErr:  false,
		},
		{
			name:     "nested boolean",
			jsonData: `{"user": {"active": false}}`,
			key:      "user.active",
			value:    true,
			want:     map[string]interface{}{"user": map[string]interface{}{"active": true}},
			wantErr:  false,
		},
		{
			name:     "new boolean key",
			jsonData: `{"name": "John"}`,
			key:      "active",
			value:    true,
			want:     map[string]interface{}{"name": "John", "active": true},
			wantErr:  false,
		},
		{
			name:     "array boolean",
			jsonData: `{"settings": [false, false]}`,
			key:      "settings.0",
			value:    true,
			want:     map[string]interface{}{"settings": []interface{}{true, false}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "active",
			value:    true,
			want:     map[string]interface{}{"active": true},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"active": true}`,
			key:      "",
			value:    false,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetBoolean([]byte(tt.jsonData), tt.key, tt.value)
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
		key      string
		value    int64
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple integer",
			jsonData: `{"age": 25}`,
			key:      "age",
			value:    30,
			want:     map[string]interface{}{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested integer",
			jsonData: `{"user": {"age": 25}}`,
			key:      "user.age",
			value:    30,
			want:     map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "new integer key",
			jsonData: `{"name": "John"}`,
			key:      "age",
			value:    30,
			want:     map[string]interface{}{"name": "John", "age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "array integer",
			jsonData: `{"ages": [10, 20]}`,
			key:      "ages.1",
			value:    30,
			want:     map[string]interface{}{"ages": []interface{}{float64(10), float64(30)}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "age",
			value:    30,
			want:     map[string]interface{}{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"age": 25}`,
			key:      "",
			value:    30,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetInt([]byte(tt.jsonData), tt.key, tt.value)
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
		key      string
		value    float64
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple float",
			jsonData: `{"price": 19.99}`,
			key:      "price",
			value:    29.99,
			want:     map[string]interface{}{"price": 29.99},
			wantErr:  false,
		},
		{
			name:     "nested float",
			jsonData: `{"product": {"price": 19.99}}`,
			key:      "product.price",
			value:    29.99,
			want:     map[string]interface{}{"product": map[string]interface{}{"price": 29.99}},
			wantErr:  false,
		},
		{
			name:     "new float key",
			jsonData: `{"name": "Product"}`,
			key:      "price",
			value:    29.99,
			want:     map[string]interface{}{"name": "Product", "price": 29.99},
			wantErr:  false,
		},
		{
			name:     "array float",
			jsonData: `{"prices": [10.5, 20.75]}`,
			key:      "prices.1",
			value:    29.99,
			want:     map[string]interface{}{"prices": []interface{}{10.5, 29.99}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "price",
			value:    29.99,
			want:     map[string]interface{}{"price": 29.99},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"price": 19.99}`,
			key:      "",
			value:    29.99,
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetFloat([]byte(tt.jsonData), tt.key, tt.value)
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
		key      string
		value    string
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple string",
			jsonData: `{"name": "John"}`,
			key:      "name",
			value:    "Jane",
			want:     map[string]interface{}{"name": "Jane"},
			wantErr:  false,
		},
		{
			name:     "nested string",
			jsonData: `{"user": {"name": "John"}}`,
			key:      "user.name",
			value:    "Jane",
			want:     map[string]interface{}{"user": map[string]interface{}{"name": "Jane"}},
			wantErr:  false,
		},
		{
			name:     "new string key",
			jsonData: `{"age": 30}`,
			key:      "name",
			value:    "John",
			want:     map[string]interface{}{"age": float64(30), "name": "John"},
			wantErr:  false,
		},
		{
			name:     "array string",
			jsonData: `{"names": ["John", "Bob"]}`,
			key:      "names.1",
			value:    "Jane",
			want:     map[string]interface{}{"names": []interface{}{"John", "Jane"}},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "name",
			value:    "John",
			want:     map[string]interface{}{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			key:      "",
			value:    "Jane",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.SetString([]byte(tt.jsonData), tt.key, tt.value)
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
		key      string
		want     map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple key",
			jsonData: `{"name": "John", "age": 30}`,
			key:      "name",
			want:     map[string]interface{}{"age": float64(30)},
			wantErr:  false,
		},
		{
			name:     "nested key",
			jsonData: `{"user": {"name": "John", "age": 30}}`,
			key:      "user.name",
			want:     map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}},
			wantErr:  false,
		},
		{
			name:     "array element",
			jsonData: `{"users": ["John", "Jane", "Bob"]}`,
			key:      "users.1",
			want:     map[string]interface{}{"users": []interface{}{"John", "Bob"}},
			wantErr:  false,
		},
		{
			name:     "key not found",
			jsonData: `{"name": "John"}`,
			key:      "age",
			want:     map[string]interface{}{"name": "John"},
			wantErr:  false,
		},
		{
			name:     "empty json",
			jsonData: ``,
			key:      "name",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "empty key",
			jsonData: `{"name": "John"}`,
			key:      "",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonParser.DeleteKey([]byte(tt.jsonData), tt.key)
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
		})
	}
}
