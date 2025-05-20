package jsonparser

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGetValue(t *testing.T) {
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
			got, err := GetValue([]byte(tt.jsonData), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetValue(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"name": "John"`,
			key:      "age",
			value:    30,
			want:     nil,
			wantErr:  true,
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
			got, err := SetValue([]byte(tt.jsonData), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("SetValue() = %v, want %v", gotMap, tt.want)
			}
		})
	}
}

func TestArrayHandling(t *testing.T) {
	// Test more complex array scenarios
	jsonData := []byte(`{
		"users": [
			{"name": "John", "age": 30},
			{"name": "Jane", "age": 25}
		]
	}`)

	// Test getting a value from an array of objects
	value, err := GetValue(jsonData, "users.1.name")
	if err != nil {
		t.Errorf("GetValue() error = %v", err)
		return
	}
	if value != "Jane" {
		t.Errorf("GetValue() = %v, want %v", value, "Jane")
	}

	// Test setting a value in an array of objects
	updatedJSON, err := SetValue(jsonData, "users.0.age", 31)
	if err != nil {
		t.Errorf("SetValue() error = %v", err)
		return
	}

	// Verify the update
	value, err = GetValue(updatedJSON, "users.0.age")
	if err != nil {
		t.Errorf("GetValue() after update error = %v", err)
		return
	}
	if value != float64(31) {
		t.Errorf("GetValue() after update = %v, want %v", value, float64(31))
	}
}

func TestEdgeCases(t *testing.T) {
	// Test with a complex nested structure
	jsonData := []byte(`{
		"data": {
			"users": [
				{
					"profile": {
						"details": {
							"name": "John",
							"contact": {
								"email": "john@example.com"
							}
						}
					}
				}
			]
		}
	}`)

	// Test getting a deeply nested value
	value, err := GetValue(jsonData, "data.users.0.profile.details.contact.email")
	if err != nil {
		t.Errorf("GetValue() deep nesting error = %v", err)
		return
	}
	if value != "john@example.com" {
		t.Errorf("GetValue() deep nesting = %v, want %v", value, "john@example.com")
	}

	// Test setting a deeply nested value
	updatedJSON, err := SetValue(jsonData, "data.users.0.profile.details.contact.phone", "123-456-7890")
	if err != nil {
		t.Errorf("SetValue() deep nesting error = %v", err)
		return
	}

	// Verify the update
	value, err = GetValue(updatedJSON, "data.users.0.profile.details.contact.phone")
	if err != nil {
		t.Errorf("GetValue() after deep nesting update error = %v", err)
		return
	}
	if value != "123-456-7890" {
		t.Errorf("GetValue() after deep nesting update = %v, want %v", value, "123-456-7890")
	}
}

func TestGetBoolean(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"active": true`,
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
			got, err := GetBoolean([]byte(tt.jsonData), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBoolean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBoolean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"age": 30`,
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
			got, err := GetInt([]byte(tt.jsonData), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFloat(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"price": 29.99`,
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
			got, err := GetFloat([]byte(tt.jsonData), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetString(t *testing.T) {
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
			got, err := GetString([]byte(tt.jsonData), tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetBoolean(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"active": true`,
			key:      "active",
			value:    false,
			want:     nil,
			wantErr:  true,
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
			got, err := SetBoolean([]byte(tt.jsonData), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetBoolean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("SetBoolean() = %v, want %v", gotMap, tt.want)
			}
		})
	}
}

func TestSetInt(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"age": 25`,
			key:      "age",
			value:    30,
			want:     nil,
			wantErr:  true,
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
			got, err := SetInt([]byte(tt.jsonData), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetInt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("SetInt() = %v, want %v", gotMap, tt.want)
			}
		})
	}
}

func TestSetFloat(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"price": 19.99`,
			key:      "price",
			value:    29.99,
			want:     nil,
			wantErr:  true,
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
			got, err := SetFloat([]byte(tt.jsonData), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("SetFloat() = %v, want %v", gotMap, tt.want)
			}
		})
	}
}

func TestSetString(t *testing.T) {
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
			name:     "invalid json",
			jsonData: `{"name": "John"`,
			key:      "name",
			value:    "Jane",
			want:     nil,
			wantErr:  true,
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
			got, err := SetString([]byte(tt.jsonData), tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			var gotMap map[string]interface{}
			if err := json.Unmarshal(got, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, tt.want) {
				t.Errorf("SetString() = %v, want %v", gotMap, tt.want)
			}
		})
	}
}
