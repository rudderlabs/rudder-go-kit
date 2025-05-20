package jsonparser

import (
	"testing"
)

// Sample JSON data for benchmarks
var (
	simpleJSON = []byte(`{"name": "John", "age": 30, "isActive": true, "height": 1.75}`)
	nestedJSON = []byte(`{
		"user": {
			"name": "John",
			"age": 30,
			"isActive": true,
			"height": 1.75,
			"address": {
				"street": "123 Main St",
				"city": "New York",
				"zipcode": "10001"
			}
		},
		"preferences": {
			"theme": "dark",
			"notifications": true
		}
	}`)
	arrayJSON = []byte(`{
		"users": [
			{"name": "John", "age": 30},
			{"name": "Jane", "age": 25},
			{"name": "Bob", "age": 40}
		],
		"scores": [10, 20, 30, 40, 50]
	}`)
)

// Benchmark GetValue for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetValue
// BenchmarkGetValue/Tidwall_Simple
// BenchmarkGetValue/Tidwall_Simple-12         	17859669	        66.14 ns/op
// BenchmarkGetValue/Grafana_Simple
// BenchmarkGetValue/Grafana_Simple-12         	13822084	        83.20 ns/op
// BenchmarkGetValue/Tidwall_Nested
// BenchmarkGetValue/Tidwall_Nested-12         	 6316160	       190.9 ns/op
// BenchmarkGetValue/Grafana_Nested
// BenchmarkGetValue/Grafana_Nested-12         	 3875821	       305.8 ns/op
// BenchmarkGetValue/Tidwall_Array
// BenchmarkGetValue/Tidwall_Array-12          	 9728286	       121.1 ns/op
// BenchmarkGetValue/Grafana_Array
// BenchmarkGetValue/Grafana_Array-12          	 2843503	       414.4 ns/op
func BenchmarkGetValue(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "name"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "name"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.address.city"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.address.city"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, "users.1.name"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, "users.1.name"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetValue(bm.jsonData, bm.key)
			}
		})
	}
}

// Benchmark GetBoolean for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetBoolean
// BenchmarkGetBoolean/Tidwall_Simple
// BenchmarkGetBoolean/Tidwall_Simple-12         	15352449	        75.00 ns/op
// BenchmarkGetBoolean/Grafana_Simple
// BenchmarkGetBoolean/Grafana_Simple-12         	11581790	        99.19 ns/op
// BenchmarkGetBoolean/Tidwall_Nested
// BenchmarkGetBoolean/Tidwall_Nested-12         	 6822956	       174.5 ns/op
// BenchmarkGetBoolean/Grafana_Nested
// BenchmarkGetBoolean/Grafana_Nested-12         	 3517758	       338.2 ns/op
func BenchmarkGetBoolean(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "isActive"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "isActive"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "preferences.notifications"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "preferences.notifications"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetBoolean(bm.jsonData, bm.key)
			}
		})
	}
}

// Benchmark GetInt for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetInt
// BenchmarkGetInt/Tidwall_Simple
// BenchmarkGetInt/Tidwall_Simple-12         	14309932	        75.12 ns/op
// BenchmarkGetInt/Grafana_Simple
// BenchmarkGetInt/Grafana_Simple-12         	 1611382	       737.2 ns/op
// BenchmarkGetInt/Tidwall_Nested
// BenchmarkGetInt/Tidwall_Nested-12         	12943046	        92.59 ns/op
// BenchmarkGetInt/Grafana_Nested
// BenchmarkGetInt/Grafana_Nested-12         	  494947	      2425 ns/op
// BenchmarkGetInt/Tidwall_Array
// BenchmarkGetInt/Tidwall_Array-12          	 7754406	       152.7 ns/op
// BenchmarkGetInt/Grafana_Array
// BenchmarkGetInt/Grafana_Array-12          	  440175	      2667 ns/op
func BenchmarkGetInt(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "age"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "age"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.age"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.age"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, "scores.2"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, "scores.2"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetInt(bm.jsonData, bm.key)
			}
		})
	}
}

// Benchmark GetFloat for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetFloat
// BenchmarkGetFloat/Tidwall_Simple
// BenchmarkGetFloat/Tidwall_Simple-12         	10967910	       109.0 ns/op
// BenchmarkGetFloat/Grafana_Simple
// BenchmarkGetFloat/Grafana_Simple-12         	 8816583	       132.0 ns/op
// BenchmarkGetFloat/Tidwall_Nested
// BenchmarkGetFloat/Tidwall_Nested-12         	 9089606	       130.5 ns/op
// BenchmarkGetFloat/Grafana_Nested
// BenchmarkGetFloat/Grafana_Nested-12         	 5602970	       206.5 ns/op
func BenchmarkGetFloat(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "height"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "height"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.height"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.height"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetFloat(bm.jsonData, bm.key)
			}
		})
	}
}

// Benchmark GetString for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetString
// BenchmarkGetString/Tidwall_Simple
// BenchmarkGetString/Tidwall_Simple-12         	 9427159	       118.3 ns/op
// BenchmarkGetString/Grafana_Simple
// BenchmarkGetString/Grafana_Simple-12         	 1416777	       739.4 ns/op
// BenchmarkGetString/Tidwall_Nested
// BenchmarkGetString/Tidwall_Nested-12         	 2994652	       396.6 ns/op
// BenchmarkGetString/Grafana_Nested
// BenchmarkGetString/Grafana_Nested-12         	  498884	      2352 ns/op
// BenchmarkGetString/Tidwall_Array
// BenchmarkGetString/Tidwall_Array-12          	 4306902	       277.0 ns/op
// BenchmarkGetString/Grafana_Array
// BenchmarkGetString/Grafana_Array-12          	  461388	      2513 ns/op
func BenchmarkGetString(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "name"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "name"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.address.street"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.address.street"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, "users.0.name"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, "users.0.name"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetString(bm.jsonData, bm.key)
			}
		})
	}
}

// Benchmark SetValue for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetValue
// BenchmarkSetValue/Tidwall_Simple
// BenchmarkSetValue/Tidwall_Simple-12         	 8829393	       121.4 ns/op
// BenchmarkSetValue/Grafana_Simple
// BenchmarkSetValue/Grafana_Simple-12         	 3739717	       320.6 ns/op
// BenchmarkSetValue/Tidwall_Nested
// BenchmarkSetValue/Tidwall_Nested-12         	 1932092	       616.4 ns/op
// BenchmarkSetValue/Grafana_Nested
// BenchmarkSetValue/Grafana_Nested-12         	 2012409	       592.3 ns/op
// BenchmarkSetValue/Tidwall_Array
// BenchmarkSetValue/Tidwall_Array-12          	 2502354	       479.1 ns/op
// BenchmarkSetValue/Grafana_Array
// BenchmarkSetValue/Grafana_Array-12          	 1668434	       681.2 ns/op
func BenchmarkSetValue(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
		value    interface{}
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "name", "Jane"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "name", "Jane"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.address.city", "Boston"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.address.city", "Boston"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, "users.1.name", "Alice"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, "users.1.name", "Alice"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetValue(jsonCopy, bm.key, bm.value)
			}
		})
	}
}

// Benchmark SetBoolean for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetBoolean
// BenchmarkSetBoolean/Tidwall_Simple
// BenchmarkSetBoolean/Tidwall_Simple-12         	 7104781	       158.4 ns/op
// BenchmarkSetBoolean/Grafana_Simple
// BenchmarkSetBoolean/Grafana_Simple-12         	 3495379	       331.1 ns/op
// BenchmarkSetBoolean/Tidwall_Nested
// BenchmarkSetBoolean/Tidwall_Nested-12         	 2888811	       413.0 ns/op
// BenchmarkSetBoolean/Grafana_Nested
// BenchmarkSetBoolean/Grafana_Nested-12         	 1894135	       628.9 ns/op
func BenchmarkSetBoolean(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
		value    bool
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "isActive", false},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "isActive", false},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "preferences.notifications", false},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "preferences.notifications", false},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetBoolean(jsonCopy, bm.key, bm.value)
			}
		})
	}
}

// Benchmark SetInt for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetInt
// BenchmarkSetInt/Tidwall_Simple
// BenchmarkSetInt/Tidwall_Simple-12         	 6572295	       177.5 ns/op
// BenchmarkSetInt/Grafana_Simple
// BenchmarkSetInt/Grafana_Simple-12         	 3572002	       321.2 ns/op
// BenchmarkSetInt/Tidwall_Nested
// BenchmarkSetInt/Tidwall_Nested-12         	 2903062	       407.5 ns/op
// BenchmarkSetInt/Grafana_Nested
// BenchmarkSetInt/Grafana_Nested-12         	 2680000	       452.2 ns/op
// BenchmarkSetInt/Tidwall_Array
// BenchmarkSetInt/Tidwall_Array-12          	 3368642	       361.7 ns/op
// BenchmarkSetInt/Grafana_Array
// BenchmarkSetInt/Grafana_Array-12          	 1631982	       730.2 ns/op
func BenchmarkSetInt(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
		value    int64
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "age", 31},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "age", 31},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.age", 31},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.age", 31},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, "scores.2", 35},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, "scores.2", 35},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetInt(jsonCopy, bm.key, bm.value)
			}
		})
	}
}

// Benchmark SetFloat for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetFloat
// BenchmarkSetFloat/Tidwall_Simple
// BenchmarkSetFloat/Tidwall_Simple-12         	 4385934	       254.0 ns/op
// BenchmarkSetFloat/Grafana_Simple
// BenchmarkSetFloat/Grafana_Simple-12         	 2755562	       443.2 ns/op
// BenchmarkSetFloat/Tidwall_Nested
// BenchmarkSetFloat/Tidwall_Nested-12         	 2160414	       562.4 ns/op
// BenchmarkSetFloat/Grafana_Nested
// BenchmarkSetFloat/Grafana_Nested-12         	 2038359	       559.1 ns/op
func BenchmarkSetFloat(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
		value    float64
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "height", 1.80},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "height", 1.80},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.height", 1.80},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.height", 1.80},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetFloat(jsonCopy, bm.key, bm.value)
			}
		})
	}
}

// Benchmark SetString for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetString
// BenchmarkSetString/Tidwall_Simple
// BenchmarkSetString/Tidwall_Simple-12         	 8017136	       143.9 ns/op
// BenchmarkSetString/Grafana_Simple
// BenchmarkSetString/Grafana_Simple-12         	 3493243	       318.7 ns/op
// BenchmarkSetString/Tidwall_Nested
// BenchmarkSetString/Tidwall_Nested-12         	 1945220	       615.3 ns/op
// BenchmarkSetString/Grafana_Nested
// BenchmarkSetString/Grafana_Nested-12         	 2031193	       580.2 ns/op
// BenchmarkSetString/Tidwall_Array
// BenchmarkSetString/Tidwall_Array-12          	 2512034	       476.1 ns/op
// BenchmarkSetString/Grafana_Array
// BenchmarkSetString/Grafana_Array-12          	 1698864	       712.4 ns/op
func BenchmarkSetString(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
		value    string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "name", "Jane"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "name", "Jane"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.address.street", "456 Oak Ave"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.address.street", "456 Oak Ave"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, "users.0.name", "Alice"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, "users.0.name", "Alice"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetString(jsonCopy, bm.key, bm.value)
			}
		})
	}
}

// Benchmark DeleteKey for both implementations
// cpu: Apple M2 Pro
// BenchmarkDeleteKey
// BenchmarkDeleteKey/Tidwall_Simple
// BenchmarkDeleteKey/Tidwall_Simple-12         	 9172108	       123.1 ns/op
// BenchmarkDeleteKey/Grafana_Simple
// BenchmarkDeleteKey/Grafana_Simple-12         	 9044328	       131.8 ns/op
// BenchmarkDeleteKey/Tidwall_Nested
// BenchmarkDeleteKey/Tidwall_Nested-12         	 1925108	       621.0 ns/op
// BenchmarkDeleteKey/Grafana_Nested
// BenchmarkDeleteKey/Grafana_Nested-12         	 2343013	       529.6 ns/op
// BenchmarkDeleteKey/Tidwall_Array
// BenchmarkDeleteKey/Tidwall_Array-12          	 3457268	       345.3 ns/op
// BenchmarkDeleteKey/Grafana_Array
// BenchmarkDeleteKey/Grafana_Array-12          	 2537061	       476.3 ns/op
func BenchmarkDeleteKey(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		key      string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, "name"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, "name"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, "user.address.city"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, "user.address.city"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, "users.1"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, "users.1"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.DeleteKey(jsonCopy, bm.key)
			}
		})
	}
}
