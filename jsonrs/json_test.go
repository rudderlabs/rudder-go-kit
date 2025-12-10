package jsonrs_test

import (
	"bytes"
	stadjson "encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
)

func TestJSONCommonFunctionality(t *testing.T) {
	run := func(t *testing.T, name string) {
		t.Run(name, func(t *testing.T) {
			c := config.New()
			c.Set("Json.Library", name)
			j := jsonrs.New(c)

			t.Run("marshall", func(t *testing.T) {
				type test struct {
					A string `json:"a"`
				}
				data, err := j.Marshal(test{A: "a"})
				require.NoError(t, err)
				require.Equal(t, `{"a":"a"}`, string(data))

				t.Run("json.RawMessage", func(t *testing.T) {
					type test struct {
						A stadjson.RawMessage `json:"a"`
					}
					data, err := j.Marshal(test{A: stadjson.RawMessage(`{"a":"a"}`)})
					require.NoError(t, err)
					require.Equal(t, `{"a":{"a":"a"}}`, string(data))
				})
			})

			t.Run("unmarshall", func(t *testing.T) {
				type test struct {
					A string `json:"a"`
				}
				var v test
				err := j.Unmarshal([]byte(`{"a":"a"}`), &v)
				require.NoError(t, err)
				require.Equal(t, "a", v.A)

				t.Run("json.RawMessage", func(t *testing.T) {
					type test struct {
						A stadjson.RawMessage `json:"a"`
					}
					var v test
					err := j.Unmarshal([]byte(`{"a":{"a":"a"}}`), &v)
					require.NoError(t, err)
					require.Equal(t, `{"a":"a"}`, string(v.A))
				})
			})

			t.Run("marshalToString", func(t *testing.T) {
				type test struct {
					A string `json:"a"`
				}
				data, err := j.MarshalToString(test{A: "a"})
				require.NoError(t, err)
				require.Equal(t, `{"a":"a"}`, data)
			})

			t.Run("unmarshal and marshal unicode", func(t *testing.T) {
				expect := func(input, output string, expectedErr ...error) {
					t.Run(input, func(t *testing.T) {
						tpl := `{"batch":[{"anonymousId":"anon_id","sentAt":"2019-08-12T05:08:30.909Z","type":"here"}]}`
						var v any
						err := j.Unmarshal(bytes.Replace([]byte(tpl), []byte("here"), []byte(input), 1), &v)
						if len(expectedErr) > 0 && expectedErr[0] != nil {
							require.Error(t, err, "expected an error")
							require.ErrorContains(t, err, expectedErr[0].Error())
						} else {
							require.NoError(t, err)
							res, err := j.Marshal(v)
							require.NoError(t, err)
							require.Equal(t, string(bytes.Replace([]byte(tpl), []byte("here"), []byte(output), 1)), string(res))

						}
					})
				}
				escapeChar := "�"
				if name == jsonrs.JsoniterLib {
					escapeChar = `\ufffd` // unique behaviour: jsoniter uses the \ufffd instead of � that all other libraries use
				}
				ecTimes := func(i int) string {
					return strings.Repeat(escapeChar, i)
				}

				expect("☺b☺", "☺b☺")
				expect("", "")
				expect("abc", "abc")
				expect("\uFDDD", "\uFDDD")
				expect("a\xffb", "a"+ecTimes(1)+"b")
				expect("\xC0\xAF", ecTimes(2))
				expect("\xE0\x80\xAF", ecTimes(3))
				expect("\xed\xa0\x80", ecTimes(3))
				expect("\xFC\x80\x80\x80\x80\xAF", ecTimes(6))
				expect(`\uD83D\ub000`, string([]byte{239, 191, 189, 235, 128, 128}))
				expect(`\uD83D\ude04`, `😄`)
				expect(`\u4e2d\u6587`, `中文`)
				expect(`\ud83d\udc4a`, `👊`)
				expect(`\U0001f64f`, "", errors.New("invalid"))
				expect(`\uD83D\u00`, "", errors.New(``))
				expect(`\ud800`, "�")
				expect(`\uDEAD`, "�")
			})
		})
	}
	run(t, jsonrs.StdLib)
	run(t, jsonrs.SonnetLib)
	run(t, jsonrs.JsoniterLib)
}

func TestJSONValidFunctionality(t *testing.T) {
	run := func(t *testing.T, name string) {
		t.Run(name, func(t *testing.T) {
			c := config.New()
			c.Set("Json.Library", name)
			j := jsonrs.New(c)

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
				// Note: jsoniter's Valid method doesn't consider standalone primitives as valid JSON,
				// unlike the standard library and sonnet implementations
				if name != jsonrs.JsoniterLib {
					require.True(t, j.Valid([]byte(`"string"`)), "string should be valid")
					require.True(t, j.Valid([]byte(`123`)), "number should be valid")
					require.True(t, j.Valid([]byte(`true`)), "boolean true should be valid")
					require.True(t, j.Valid([]byte(`false`)), "boolean false should be valid")
					require.True(t, j.Valid([]byte(`null`)), "null should be valid")
				}
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
		})
	}

	run(t, jsonrs.StdLib)
	run(t, jsonrs.SonnetLib)
	run(t, jsonrs.JsoniterLib)
	run(t, jsonrs.TidwallLib)
}
