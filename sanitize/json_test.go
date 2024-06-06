package sanitize

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type unmarshaller func([]byte, any) error

func TestInvalidJSON(t *testing.T) {
	t.Run("single unpaired UTF-16 surrogate", func(t *testing.T) {
		unmarshallers := []unmarshaller{
			stdjson.Unmarshal,
			jsoniter.ConfigFastest.Unmarshal,
			// validateJSONWithoutAlloc,
		}

		data := []byte(`{"key": "value\uDEAD"}`)
		require.True(t, stdjson.Valid(data))
		require.True(t, gjson.ValidBytes(data))
		require.True(t, jsoniter.ConfigFastest.Valid(data))

		for _, unmarshal := range unmarshallers {
			var a any
			err := unmarshal(data, &a)
			require.NoError(t, err)
			t.Log(a)
			// require.Equal(t, "value�", a.(map[string]any)["key"])
		}
	})

	t.Run("emoji", func(t *testing.T) {
		unmarshallers := []unmarshaller{
			stdjson.Unmarshal,
			jsoniter.ConfigFastest.Unmarshal,
			// validateJSONWithoutAlloc,
		}

		data := []byte(`{"key": "value\U0001f64f"}`)
		require.False(t, stdjson.Valid(data))
		require.False(t, gjson.ValidBytes(data))
		require.False(t, jsoniter.ConfigFastest.Valid(data))

		for _, unmarshal := range unmarshallers {
			var a any
			err := unmarshal(data, &a)
			require.Error(t, err)
		}
	})
}

func TestScenarios(t *testing.T) {
	tests := []struct {
		in  string
		err error
	}{
		{`\u0000`, nil},
		{`\u0000☺\u0000b☺`, nil},
		// NOTE: we are not handling the following:
		// {"\u0000", ""},
		// {"\u0000☺\u0000b☺", "☺b☺"},

		{"abc", nil},
		{"\uFDDD", nil},
		{"a\xffb", nil},
		{"a\xffb\uFFFD", nil},
		{"a☺\xffb☺\xC0\xAFc☺\xff", nil},
		{"\xC0\xAF", nil},
		{"\xE0\x80\xAF", nil},
		{"\xed\xa0\x80", nil},
		{"\xed\xbf\xbf", nil},
		{"\xF0\x80\x80\xaf", nil},
		{"\xF8\x80\x80\x80\xAF", nil},
		{"\xFC\x80\x80\x80\x80\xAF", nil},

		// {"\ud800", ""},
		{`\ud800`, nil},
		{`\uDEAD`, nil},

		{`\uD83D\ub000`, nil},
		{`\uD83D\ude04`, nil},

		{`\u4e2d\u6587`, nil},
		{`\ud83d\udc4a`, nil},

		{`\U0001f64f`, errors.New(`readEscapedChar: invalid escape char after`)},
		{`\uD83D\u00`, errors.New(`readU4: expects 0~9 or a~f, but found`)},
	}
	for index, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			data := []byte(tt.in)
			valid := stdjson.Valid(data)
			// err := validateJSONWithoutAlloc(data, nil)
			data, err := sanitizeJSON(data)
			if valid {
				require.NoErrorf(t, err, "[index %d] Payload was valid but got an error: %s: %v", index, tt.in, err)
			} else {
				require.Errorf(t, err, "[index %d] Payload was invalid but didn't get an error: %s", index, tt.in)
			}
			if err != nil {
				require.Falsef(t, valid, "[index %d] Got an error but payload is valid: %s: %v", index, tt.in, err)
			} else {
				require.True(t, valid, "[index %d] Got no error but payload is not valid: %s: %v", index, tt.in, err)
			}
			if tt.err != nil {
				require.Errorf(t, err, "[index %d] Expected error: %s: %v", index, tt.in, tt.err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	testCases := []testCase{
		{
			input:    `{ "key": "value" }`,
			expected: `{"key":"value"}`,
		},
		{
			input:    `{ "key": "value\uDEAD", "array": [1, 2, {"a": "b"}] }`,
			expected: `{"key":"value�","array":[1,2,{"a":"b"}]}`,
		},
		{
			input:    `{ "key1": "value1", "key2": 123, "key3": true, "key4": null }`,
			expected: `{"key1":"value1","key2":123,"key3":true,"key4":null}`,
		},
		{
			input:    `[ 1, 2, 3, 4, 5 ]`,
			expected: `[1,2,3,4,5]`,
		},
		{
			input:    `{ "nested": { "innerKey": "innerValue" } }`,
			expected: `{"nested":{"innerKey":"innerValue"}}`,
		},
		{
			input:    `{ "array": [ { "key": "value" }, { "key": 123 }, { "key": true } ] }`,
			expected: `{"array":[{"key":"value"},{"key":123},{"key":true}]}`,
		},
		{
			input:    `[ { "key1": "value1" }, { "key2": "value2" } ]`,
			expected: `[{"key1":"value1"},{"key2":"value2"}]`,
		},
		{
			input:    `{ "escaped": "newline\n tab\t quote\" backslash\\ and unicode\u1234" }`,
			expected: `{"escaped":"newline\n tab\t quote\" backslash\\ and unicode\u1234"}`,
		},
		{
			input:    `{"emptyObj":{},"emptyArray":[]}`,
			expected: `{"emptyObj":{},"emptyArray":[]}`,
		},
		{
			input:    `{}`,
			expected: `{}`,
		},
		{
			input:    `[]`,
			expected: `[]`,
		},
	}
	for index, tc := range testCases {
		t.Run(fmt.Sprintf("test-%d", index), func(t *testing.T) {
			data := []byte(tc.input)
			data, err := sanitizeJSON(data)
			require.NoError(t, err)
			// require.True(t, stdjson.Valid(data))
			require.Equal(t, tc.expected, string(data))
		})
	}
}

func sanitizeJSON(data []byte) ([]byte, error) {
	var (
		decoder         = stdjson.NewDecoder(bytes.NewReader(data))
		writePos        int
		isKey, inObject bool
	)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch v := token.(type) {
		case stdjson.Delim:
			if v == '{' {
				isKey = true // The next string token is a key
				inObject = true
			} else if v == '}' {
				inObject = false
			}
			data[writePos] = byte(v)
			writePos++
		case string:
			data[writePos] = '"'
			writePos++
			writePos += copy(data[writePos:], v)
			data[writePos] = '"'
			writePos++
			if !inObject {
				continue
			}
			if isKey {
				data[writePos] = ':'
				writePos++
				isKey = false
			} else if decoder.More() {
				data[writePos] = ','
				writePos++
				isKey = true
			}
		case float64:
			n := copy(data[writePos:], strconv.FormatFloat(v, 'f', -1, 64))
			writePos += n
			if decoder.More() {
				data[writePos] = ','
				writePos++
				isKey = inObject
			}
		case bool:
			if v {
				n := copy(data[writePos:], "true")
				writePos += n
			} else {
				n := copy(data[writePos:], "false")
				writePos += n
			}
			if decoder.More() {
				data[writePos] = ','
				writePos++
				isKey = inObject
			}
		case nil:
			n := copy(data[writePos:], "null")
			writePos += n
			if decoder.More() {
				data[writePos] = ','
				writePos++
				isKey = inObject
			}
		}
	}

	// Return the compacted slice
	return data[:writePos], nil
}

type testCase struct {
	input    string
	expected string
}
