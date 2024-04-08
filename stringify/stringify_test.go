package stringify_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stringify"
)

type failOnJSONMarshal struct{}

func (f failOnJSONMarshal) MarshalJSON() ([]byte, error) {
	return nil, errors.New("failed to marshal")
}

func TestStringyData(t *testing.T) {
	testCases := []struct {
		name     string
		input    any
		expected string
	}{
		{
			name:     "Nil input",
			input:    nil,
			expected: "",
		},
		{
			name:     "String input",
			input:    "test string",
			expected: "test string",
		},
		{
			name: "Struct input",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "John", Age: 30},
			expected: `{"name":"John","age":30}`,
		},
		{
			name: "Slice input",
			input: []string{
				"apple", "banana", "cherry",
			},
			expected: `["apple","banana","cherry"]`,
		},
		{
			name:     "Fail on JSON marshal",
			input:    failOnJSONMarshal{},
			expected: "{}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stringify.Any(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
