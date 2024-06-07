package sanitize

import (
	"bytes"
	stdjson "encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/rudderlabs/rudder-go-kit/bytesize"
	kitrand "github.com/rudderlabs/rudder-go-kit/testhelper/rand"
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
			data, err := JSON(data)
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

func TestSanitize(t *testing.T) {
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
			input:    `{"pZwNSfv":[["Xsplf",0.21300102020231929,"VQeZct"],[1,2,3]]}`,
			expected: `{"pZwNSfv":[["Xsplf",0.21300102020231929,"VQeZct"],[1,2,3]]}`,
		},
		{
			input:    `{"EmpkTpX":60,"nMddqQbdlsh":10,"gZIlh":"dinyjdLFEgq","HcRtVmD":{"cUz":"pMqncfNL","jeMShhU":{"QFzsnVhYQcES":0.9932440939111381,"RuAOplNKpV":0.7897482040243524,"WwxzTx":"PUYGI"},"zoHMkiemcb":99},"fFlu":"VETBIeL","zHiVTMZ":"xcEQwJLMabRl","gxc":{"GUrvges":0.3808440487755879,"Geuv":"fwtQF","RVNKxLsy":[{"HKdrRi":null,"WDxPSdOubeVC":8,"wPaUDXFd":0.8633220142242786,"wcAMWgUXIYC":"bvOr"},"TPBgglYNRn",{"AskkKz":"zEoYOjekT","ZQHbSbvZy":false,"hAvDwz":true},84]},"NwxqqevwKmLH":0.6948128888176054,"TXlDGztev":97,"DPDccZd":82,"ayqiNrhDTSbq":false,"DsuttI":0.22825143944569815,"BxSRncZIcg":3,"EFDp":"qZCnKMFI","qWMWFS":0.00968211300380408,"SJJrmRgGykm":["Zwl",{"iKeIySC":[null,"yguM",false,null,0.7286897051192204]},{"BXBY":0.6263173559674706,"VEScYXqUljH":{"DxcqKAniO":0.7662477958047248,"KYTppiLmKZao":false,"fJb":82},"aeYiVc":"EWpjAfwSvZD","rffvG":{"VfgGcxmODZhy":67}}],"VwXezohstGb":[0.773808900934636,false,{"PpTnYVHsuIn":{"SGNyFA":false,"ibJVNQN":0.40290312272801454}},"yxtnwTMpEXo",[35,"hCCEsiz",[34],0.36730866001647755]]}`,
			expected: `{"EmpkTpX":60,"nMddqQbdlsh":10,"gZIlh":"dinyjdLFEgq","HcRtVmD":{"cUz":"pMqncfNL","jeMShhU":{"QFzsnVhYQcES":0.9932440939111381,"RuAOplNKpV":0.7897482040243524,"WwxzTx":"PUYGI"},"zoHMkiemcb":99},"fFlu":"VETBIeL","zHiVTMZ":"xcEQwJLMabRl","gxc":{"GUrvges":0.3808440487755879,"Geuv":"fwtQF","RVNKxLsy":[{"HKdrRi":null,"WDxPSdOubeVC":8,"wPaUDXFd":0.8633220142242786,"wcAMWgUXIYC":"bvOr"},"TPBgglYNRn",{"AskkKz":"zEoYOjekT","ZQHbSbvZy":false,"hAvDwz":true},84]},"NwxqqevwKmLH":0.6948128888176054,"TXlDGztev":97,"DPDccZd":82,"ayqiNrhDTSbq":false,"DsuttI":0.22825143944569815,"BxSRncZIcg":3,"EFDp":"qZCnKMFI","qWMWFS":0.00968211300380408,"SJJrmRgGykm":["Zwl",{"iKeIySC":[null,"yguM",false,null,0.7286897051192204]},{"BXBY":0.6263173559674706,"VEScYXqUljH":{"DxcqKAniO":0.7662477958047248,"KYTppiLmKZao":false,"fJb":82},"aeYiVc":"EWpjAfwSvZD","rffvG":{"VfgGcxmODZhy":67}}],"VwXezohstGb":[0.773808900934636,false,{"PpTnYVHsuIn":{"SGNyFA":false,"ibJVNQN":0.40290312272801454}},"yxtnwTMpEXo",[35,"hCCEsiz",[34],0.36730866001647755]]}`,
		},
		{
			input:    `[ { "key1": "value1" }, { "key2": "value2" } ]`,
			expected: `[{"key1":"value1"},{"key2":"value2"}]`,
		},
		{
			input:    `{ "escaped": "newline\n tab\t quote\" backslash\\ and unicode\u1234" }`,
			expected: `{"escaped":"newline\n tab\t quote\" backslash\\ and unicodeሴ"}`,
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
		// TODO add random brackets opening and not closing and then just closing without opening them
	}
	for index, tc := range testCases {
		t.Run(fmt.Sprintf("test-%d", index), func(t *testing.T) {
			data := []byte(tc.input)
			if tc.err != nil {
				require.False(t, stdjson.Valid(data))
			} else {
				require.True(t, stdjson.Valid(data))
			}

			var err error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("Recovered from panic: %v", r)
						t.Log("Data:", string(data))
						t.FailNow()
					}
				}()
				data, err = JSON(data)
			}()
			t.Logf("Produced output (err: %v): %s", err, data)

			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				require.False(t, stdjson.Valid(data))
				return
			}

			require.Truef(t, stdjson.Valid(data), "Invalid JSON: %s", data)
			require.NoError(t, err)
			require.Equal(t, tc.expected, string(data))
		})
	}
}

type testCase struct {
	input    string
	expected string
	err      error
}

func TestSanitizeRandom(t *testing.T) {
	t.Skip("TODO")
}

func BenchmarkSanitize(b *testing.B) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary
	data := []byte(generateJSON(10 * bytesize.MB))
	require.True(b, stdjson.Valid(data))
	b.ResetTimer()

	b.Run("marshal-unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cp := make([]byte, len(data))
			copy(cp, data)

			var a any
			err := json.Unmarshal(cp, &a)
			require.NoError(b, err)
			cp, err = json.Marshal(a)
			require.NoError(b, err)
		}
	})

	b.Run("sanitize", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cp := make([]byte, len(data))
			copy(cp, data)

			var err error
			cp, err = JSON(cp)
			require.NoError(b, err)
		}
	})
}

// Generates a JSON string of roughly the specified size in bytes.
func generateJSON(size int64) string {
	var buffer bytes.Buffer

	// Helper function to generate random keys
	randomKey := func() string { return kitrand.String(rand.Intn(10) + 3) }

	// Helper function to generate random values
	var generateValue func(depth int) any
	generateValue = func(depth int) any {
		switch rand.Intn(6) {
		case 0:
			return rand.Float64()
		case 1:
			return rand.Intn(100)
		case 2:
			return rand.Int()%2 == 0
		case 3:
			return randomKey()
		case 4:
			if depth > 0 {
				nestedObj := make(map[string]interface{})
				for i := 0; i < rand.Intn(5)+1; i++ {
					nestedObj[randomKey()] = generateValue(depth - 1)
				}
				return nestedObj
			}
		case 5:
			if depth > 0 {
				array := make([]interface{}, rand.Intn(5)+1)
				for i := range array {
					array[i] = generateValue(depth - 1)
				}
				return array
			}
		}
		return nil
	}

	// Start generating JSON
	buffer.WriteString("{")
	for int64(buffer.Len()) < size-10 { // Leave some room for closing brackets
		key := randomKey()
		value := generateValue(3) // Adjust depth for more complexity

		// Serialize key-value pair to JSON
		keyJSON, _ := stdjson.Marshal(key)
		valueJSON, _ := stdjson.Marshal(value)
		buffer.Write(keyJSON)
		buffer.WriteString(":")
		buffer.Write(valueJSON)
		buffer.WriteString(",")

		// Check if we've exceeded the desired size
		if int64(buffer.Len()) >= size {
			break
		}
	}
	// Remove last comma and close the JSON object
	if buffer.Len() > 1 && buffer.Bytes()[buffer.Len()-1] == ',' {
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteString("}")

	return buffer.String()
}
