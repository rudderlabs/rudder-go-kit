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
// BenchmarkGetValue/Tidwall_Simple-12         	15010687	        67.33 ns/op	      24 B/op	       2 allocs/op
// BenchmarkGetValue/Grafana_Simple
// BenchmarkGetValue/Grafana_Simple-12         	13216773	        87.18 ns/op	      88 B/op	       5 allocs/op
// BenchmarkGetValue/Tidwall_Nested
// BenchmarkGetValue/Tidwall_Nested-12         	 6226434	       190.8 ns/op	      32 B/op	       2 allocs/op
// BenchmarkGetValue/Grafana_Nested
// BenchmarkGetValue/Grafana_Nested-12         	 3806776	       308.7 ns/op	     288 B/op	      10 allocs/op
// BenchmarkGetValue/Tidwall_Array
// BenchmarkGetValue/Tidwall_Array-12          	 9697383	       122.8 ns/op	      24 B/op	       2 allocs/op
// BenchmarkGetValue/Grafana_Array
// BenchmarkGetValue/Grafana_Array-12          	 2846985	       419.1 ns/op	     224 B/op	       9 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark GetBoolean for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetBoolean
// BenchmarkGetBoolean/Tidwall_Simple
// BenchmarkGetBoolean/Tidwall_Simple-12         	14620082	        74.68 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetBoolean/Grafana_Simple
// BenchmarkGetBoolean/Grafana_Simple-12         	11895601	       100.4 ns/op	      72 B/op	       3 allocs/op
// BenchmarkGetBoolean/Tidwall_Nested
// BenchmarkGetBoolean/Tidwall_Nested-12         	 6825517	       176.2 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetBoolean/Grafana_Nested
// BenchmarkGetBoolean/Grafana_Nested-12         	 3443974	       354.7 ns/op	     192 B/op	       6 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark GetInt for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetInt
// BenchmarkGetInt/Tidwall_Simple
// BenchmarkGetInt/Tidwall_Simple-12         	15730716	        74.17 ns/op	       2 B/op	       1 allocs/op
// BenchmarkGetInt/Grafana_Simple
// BenchmarkGetInt/Grafana_Simple-12         	 1603052	       739.5 ns/op	     648 B/op	      18 allocs/op
// BenchmarkGetInt/Tidwall_Nested
// BenchmarkGetInt/Tidwall_Nested-12         	13113523	        91.55 ns/op	       2 B/op	       1 allocs/op
// BenchmarkGetInt/Grafana_Nested
// BenchmarkGetInt/Grafana_Nested-12         	  498670	      2260 ns/op	    1968 B/op	      45 allocs/op
// BenchmarkGetInt/Tidwall_Array
// BenchmarkGetInt/Tidwall_Array-12          	 7950166	       153.6 ns/op	       2 B/op	       1 allocs/op
// BenchmarkGetInt/Grafana_Array
// BenchmarkGetInt/Grafana_Array-12          	  435844	      2797 ns/op	    2296 B/op	      57 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark GetFloat for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetFloat
// BenchmarkGetFloat/Tidwall_Simple
// BenchmarkGetFloat/Tidwall_Simple-12         	 9430736	       109.5 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetFloat/Grafana_Simple
// BenchmarkGetFloat/Grafana_Simple-12         	 9044486	       131.0 ns/op	      72 B/op	       3 allocs/op
// BenchmarkGetFloat/Tidwall_Nested
// BenchmarkGetFloat/Tidwall_Nested-12         	 9041659	       129.9 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetFloat/Grafana_Nested
// BenchmarkGetFloat/Grafana_Nested-12         	 5617030	       207.9 ns/op	     176 B/op	       6 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark GetString for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetString
// BenchmarkGetString/Tidwall_Simple
// BenchmarkGetString/Tidwall_Simple-12         	10134784	       116.4 ns/op	       8 B/op	       1 allocs/op
// BenchmarkGetString/Grafana_Simple
// BenchmarkGetString/Grafana_Simple-12         	 1619782	       725.5 ns/op	     648 B/op	      19 allocs/op
// BenchmarkGetString/Tidwall_Nested
// BenchmarkGetString/Tidwall_Nested-12         	 3004423	       397.9 ns/op	      16 B/op	       1 allocs/op
// BenchmarkGetString/Grafana_Nested
// BenchmarkGetString/Grafana_Nested-12         	  487449	      2386 ns/op	    2080 B/op	      48 allocs/op
// BenchmarkGetString/Tidwall_Array
// BenchmarkGetString/Tidwall_Array-12          	 4215164	       276.9 ns/op	       8 B/op	       1 allocs/op
// BenchmarkGetString/Grafana_Array
// BenchmarkGetString/Grafana_Array-12          	  447976	      2533 ns/op	    2376 B/op	      60 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark SetValue for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetValue
// BenchmarkSetValue/Tidwall_Simple
// BenchmarkSetValue/Tidwall_Simple-12         	 8493249	       128.8 ns/op	     208 B/op	       4 allocs/op
// BenchmarkSetValue/Grafana_Simple
// BenchmarkSetValue/Grafana_Simple-12         	 3760342	       305.2 ns/op	     552 B/op	      10 allocs/op
// BenchmarkSetValue/Tidwall_Nested
// BenchmarkSetValue/Tidwall_Nested-12         	 1944204	       612.7 ns/op	    1488 B/op	       8 allocs/op
// BenchmarkSetValue/Grafana_Nested
// BenchmarkSetValue/Grafana_Nested-12         	 2071608	       575.3 ns/op	    1192 B/op	      15 allocs/op
// BenchmarkSetValue/Tidwall_Array
// BenchmarkSetValue/Tidwall_Array-12          	 2470988	       477.3 ns/op	    1024 B/op	       8 allocs/op
// BenchmarkSetValue/Grafana_Array
// BenchmarkSetValue/Grafana_Array-12          	 1745088	       680.2 ns/op	     880 B/op	      14 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark SetBoolean for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetBoolean
// BenchmarkSetBoolean/Tidwall_Simple
// BenchmarkSetBoolean/Tidwall_Simple-12         	 7126352	       159.2 ns/op	     272 B/op	       4 allocs/op
// BenchmarkSetBoolean/Grafana_Simple
// BenchmarkSetBoolean/Grafana_Simple-12         	 3462522	       331.8 ns/op	     552 B/op	      10 allocs/op
// BenchmarkSetBoolean/Tidwall_Nested
// BenchmarkSetBoolean/Tidwall_Nested-12         	 2860891	       413.7 ns/op	    1088 B/op	       5 allocs/op
// BenchmarkSetBoolean/Grafana_Nested
// BenchmarkSetBoolean/Grafana_Nested-12         	 1877862	       636.1 ns/op	    1120 B/op	      13 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark SetInt for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetInt
// BenchmarkSetInt/Tidwall_Simple
// BenchmarkSetInt/Tidwall_Simple-12         	 6425004	       180.7 ns/op	     296 B/op	       5 allocs/op
// BenchmarkSetInt/Grafana_Simple
// BenchmarkSetInt/Grafana_Simple-12         	 3547204	       326.5 ns/op	     542 B/op	      10 allocs/op
// BenchmarkSetInt/Tidwall_Nested
// BenchmarkSetInt/Tidwall_Nested-12         	 2873828	       411.3 ns/op	    1104 B/op	       7 allocs/op
// BenchmarkSetInt/Grafana_Nested
// BenchmarkSetInt/Grafana_Nested-12         	 2600205	       450.5 ns/op	    1092 B/op	      13 allocs/op
// BenchmarkSetInt/Tidwall_Array
// BenchmarkSetInt/Tidwall_Array-12          	 3320079	       354.9 ns/op	     720 B/op	       5 allocs/op
// BenchmarkSetInt/Grafana_Array
// BenchmarkSetInt/Grafana_Array-12          	 1652884	       721.3 ns/op	     792 B/op	      12 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark SetFloat for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetFloat
// BenchmarkSetFloat/Tidwall_Simple
// BenchmarkSetFloat/Tidwall_Simple-12         	 4580779	       256.8 ns/op	     232 B/op	       6 allocs/op
// BenchmarkSetFloat/Grafana_Simple
// BenchmarkSetFloat/Grafana_Simple-12         	 2768779	       424.1 ns/op	     560 B/op	      11 allocs/op
// BenchmarkSetFloat/Tidwall_Nested
// BenchmarkSetFloat/Tidwall_Nested-12         	 2241553	       528.3 ns/op	    1192 B/op	      10 allocs/op
// BenchmarkSetFloat/Grafana_Nested
// BenchmarkSetFloat/Grafana_Nested-12         	 2134204	       553.9 ns/op	    1104 B/op	      14 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark SetString for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetString
// BenchmarkSetString/Tidwall_Simple
// BenchmarkSetString/Tidwall_Simple-12         	 8158161	       140.1 ns/op	     224 B/op	       5 allocs/op
// BenchmarkSetString/Grafana_Simple
// BenchmarkSetString/Grafana_Simple-12         	 3729015	       319.7 ns/op	     568 B/op	      11 allocs/op
// BenchmarkSetString/Tidwall_Nested
// BenchmarkSetString/Tidwall_Nested-12         	 1928187	       618.6 ns/op	    1504 B/op	       9 allocs/op
// BenchmarkSetString/Grafana_Nested
// BenchmarkSetString/Grafana_Nested-12         	 2042232	       603.6 ns/op	    1216 B/op	      16 allocs/op
// BenchmarkSetString/Tidwall_Array
// BenchmarkSetString/Tidwall_Array-12          	 2372666	       502.6 ns/op	    1072 B/op	      10 allocs/op
// BenchmarkSetString/Grafana_Array
// BenchmarkSetString/Grafana_Array-12          	 1610833	       734.5 ns/op	     896 B/op	      15 allocs/op
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
			b.ReportAllocs()
		})
	}
}

// Benchmark DeleteKey for both implementations
// cpu: Apple M2 Pro
// BenchmarkDeleteKey
// BenchmarkDeleteKey/Tidwall_Simple
// BenchmarkDeleteKey/Tidwall_Simple-12         	 8312287	       124.2 ns/op	     192 B/op	       4 allocs/op
// BenchmarkDeleteKey/Grafana_Simple
// BenchmarkDeleteKey/Grafana_Simple-12         	 9203091	       129.4 ns/op	     260 B/op	       6 allocs/op
// BenchmarkDeleteKey/Tidwall_Nested
// BenchmarkDeleteKey/Tidwall_Nested-12         	 1921638	       615.0 ns/op	    1488 B/op	       8 allocs/op
// BenchmarkDeleteKey/Grafana_Nested
// BenchmarkDeleteKey/Grafana_Nested-12         	 2281663	       501.5 ns/op	     896 B/op	      11 allocs/op
// BenchmarkDeleteKey/Tidwall_Array
// BenchmarkDeleteKey/Tidwall_Array-12          	 3481876	       340.1 ns/op	     544 B/op	       6 allocs/op
// BenchmarkDeleteKey/Grafana_Array
// BenchmarkDeleteKey/Grafana_Array-12          	 2611114	       455.1 ns/op	     440 B/op	       7 allocs/op
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
			b.ReportAllocs()
		})
	}
}
