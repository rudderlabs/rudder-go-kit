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
// BenchmarkGetValue/Tidwall_Simple-12         	12702640	        79.55 ns/op	      40 B/op	       3 allocs/op
// BenchmarkGetValue/Grafana_Simple
// BenchmarkGetValue/Grafana_Simple-12         	27977652	        41.51 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetValue/Tidwall_Nested
// BenchmarkGetValue/Tidwall_Nested-12         	 4783870	       250.3 ns/op	     104 B/op	       4 allocs/op
// BenchmarkGetValue/Grafana_Nested
// BenchmarkGetValue/Grafana_Nested-12         	 7068457	       168.2 ns/op	      24 B/op	       2 allocs/op
// BenchmarkGetValue/Tidwall_Array
// BenchmarkGetValue/Tidwall_Array-12          	 6799860	       176.4 ns/op	      88 B/op	       4 allocs/op
// BenchmarkGetValue/Grafana_Array
// BenchmarkGetValue/Grafana_Array-12          	 7053883	       168.9 ns/op	      20 B/op	       2 allocs/op
func BenchmarkGetValue(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"name"}},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"name"}},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "city"}},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "city"}},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"users", "[1]", "name"}},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"users", "[1]", "name"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetValue(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetBoolean for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetBoolean
// BenchmarkGetBoolean/Tidwall_Simple
// BenchmarkGetBoolean/Tidwall_Simple-12         	11014276	        91.06 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetBoolean/Grafana_Simple
// BenchmarkGetBoolean/Grafana_Simple-12         	22884204	        52.69 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetBoolean/Tidwall_Nested
// BenchmarkGetBoolean/Tidwall_Nested-12         	 5403700	       220.8 ns/op	      68 B/op	       3 allocs/op
// BenchmarkGetBoolean/Grafana_Nested
// BenchmarkGetBoolean/Grafana_Nested-12         	 5253258	       227.9 ns/op	       0 B/op	       0 allocs/op
func BenchmarkGetBoolean(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"isActive"}},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"isActive"}},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"preferences", "notifications"}},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"preferences", "notifications"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetBoolean(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetInt for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetInt
// BenchmarkGetInt/Tidwall_Simple
// BenchmarkGetInt/Tidwall_Simple-12         	12537536	        87.76 ns/op	      18 B/op	       2 allocs/op
// BenchmarkGetInt/Grafana_Simple
// BenchmarkGetInt/Grafana_Simple-12         	29973991	        40.46 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetInt/Tidwall_Nested
// BenchmarkGetInt/Tidwall_Nested-12         	 8741730	       135.2 ns/op	      48 B/op	       3 allocs/op
// BenchmarkGetInt/Grafana_Nested
// BenchmarkGetInt/Grafana_Nested-12         	20628495	        57.99 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetInt/Tidwall_Array
// BenchmarkGetInt/Tidwall_Array-12          	 6123844	       195.1 ns/op	      48 B/op	       3 allocs/op
// BenchmarkGetInt/Grafana_Array
// BenchmarkGetInt/Grafana_Array-12          	 6181316	       193.3 ns/op	       0 B/op	       0 allocs/op
func BenchmarkGetInt(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"age"}},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"age"}},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "age"}},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "age"}},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"scores", "[2]"}},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"scores", "[2]"}},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetInt(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetFloat for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetFloat
// BenchmarkGetFloat/Tidwall_Simple
// BenchmarkGetFloat/Tidwall_Simple-12         	 9523972	       122.6 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetFloat/Grafana_Simple
// BenchmarkGetFloat/Grafana_Simple-12         	14540716	        82.35 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetFloat/Tidwall_Nested
// BenchmarkGetFloat/Tidwall_Nested-12         	 6636676	       178.1 ns/op	      52 B/op	       3 allocs/op
// BenchmarkGetFloat/Grafana_Nested
// BenchmarkGetFloat/Grafana_Nested-12         	11360421	       104.5 ns/op	       0 B/op	       0 allocs/op
func BenchmarkGetFloat(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"height"}},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"height"}},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "height"}},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "height"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetFloat(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetString for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetString
// BenchmarkGetString/Tidwall_Simple
// BenchmarkGetString/Tidwall_Simple-12         	15042636	        69.35 ns/op	      24 B/op	       2 allocs/op
// BenchmarkGetString/Grafana_Simple
// BenchmarkGetString/Grafana_Simple-12         	38384338	        31.36 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetString/Tidwall_Nested
// BenchmarkGetString/Tidwall_Nested-12         	 5471368	       220.6 ns/op	      88 B/op	       3 allocs/op
// BenchmarkGetString/Grafana_Nested
// BenchmarkGetString/Grafana_Nested-12         	 8712806	       135.5 ns/op	      16 B/op	       1 allocs/op
// BenchmarkGetString/Tidwall_Array
// BenchmarkGetString/Tidwall_Array-12          	 8505716	       141.4 ns/op	      72 B/op	       3 allocs/op
// BenchmarkGetString/Grafana_Array
// BenchmarkGetString/Grafana_Array-12          	 7496886	       157.7 ns/op	       4 B/op	       1 allocs/op
func BenchmarkGetString(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"name"}},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"name"}},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "street"}},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "street"}},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"users", "[0]", "name"}},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"users", "[0]", "name"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = bm.parser.GetString(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark SetValue for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetValue
// BenchmarkSetValue/Tidwall_Simple
// BenchmarkSetValue/Tidwall_Simple-12         	 7339918	       142.6 ns/op	     224 B/op	       5 allocs/op
// BenchmarkSetValue/Grafana_Simple
// BenchmarkSetValue/Grafana_Simple-12         	16763170	        71.24 ns/op	     128 B/op	       2 allocs/op
// BenchmarkSetValue/Tidwall_Nested
// BenchmarkSetValue/Tidwall_Nested-12         	 1766529	       671.3 ns/op	    1560 B/op	      10 allocs/op
// BenchmarkSetValue/Grafana_Nested
// BenchmarkSetValue/Grafana_Nested-12         	 5006306	       240.5 ns/op	     576 B/op	       2 allocs/op
// BenchmarkSetValue/Tidwall_Array
// BenchmarkSetValue/Tidwall_Array-12          	 2236798	       528.9 ns/op	    1088 B/op	      10 allocs/op
// BenchmarkSetValue/Grafana_Array
// BenchmarkSetValue/Grafana_Array-12          	 5423719	       219.1 ns/op	     320 B/op	       2 allocs/op
func BenchmarkSetValue(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
		value    interface{}
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"name"}, "Jane"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"name"}, "Jane"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "city"}, "Boston"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "city"}, "Boston"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"users", "[1]", "name"}, "Alice"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"users", "[1]", "name"}, "Alice"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetValue(jsonCopy, bm.value, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark SetBoolean for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetBoolean
// BenchmarkSetBoolean/Tidwall_Simple
// BenchmarkSetBoolean/Tidwall_Simple-12         	 6539937	       176.4 ns/op	     288 B/op	       5 allocs/op
// BenchmarkSetBoolean/Grafana_Simple
// BenchmarkSetBoolean/Grafana_Simple-12         	13664455	        87.35 ns/op	     128 B/op	       2 allocs/op
// BenchmarkSetBoolean/Tidwall_Nested
// BenchmarkSetBoolean/Tidwall_Nested-12         	 2543312	       470.3 ns/op	    1152 B/op	       7 allocs/op
// BenchmarkSetBoolean/Grafana_Nested
// BenchmarkSetBoolean/Grafana_Nested-12         	 3956414	       302.1 ns/op	     576 B/op	       2 allocs/op
func BenchmarkSetBoolean(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
		value    bool
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"isActive"}, false},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"isActive"}, false},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"preferences", "notifications"}, false},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"preferences", "notifications"}, false},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetBoolean(jsonCopy, bm.value, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark SetInt for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetFloat
// BenchmarkSetFloat/Tidwall_Simple
// BenchmarkSetFloat/Tidwall_Simple-12         	 4202719	       267.1 ns/op	     248 B/op	       7 allocs/op
// BenchmarkSetFloat/Grafana_Simple
// BenchmarkSetFloat/Grafana_Simple-12         	 6857023	       174.4 ns/op	     160 B/op	       4 allocs/op
// BenchmarkSetFloat/Tidwall_Nested
// BenchmarkSetFloat/Tidwall_Nested-12         	 2051656	       579.8 ns/op	    1240 B/op	      12 allocs/op
// BenchmarkSetFloat/Grafana_Nested
// BenchmarkSetFloat/Grafana_Nested-12         	 4942752	       240.3 ns/op	     608 B/op	       4 allocs/op
func BenchmarkSetInt(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
		value    int64
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"age"}, 31},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"age"}, 31},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "age"}, 31},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "age"}, 31},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"scores", "[2]"}, 35},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"scores", "[2]"}, 35},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetInt(jsonCopy, bm.value, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark SetFloat for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetFloat
// BenchmarkSetFloat/Tidwall_Simple
// BenchmarkSetFloat/Tidwall_Simple-12         	 4297122	       271.1 ns/op	     248 B/op	       7 allocs/op
// BenchmarkSetFloat/Grafana_Simple
// BenchmarkSetFloat/Grafana_Simple-12         	 3120792	       387.4 ns/op	     488 B/op	       8 allocs/op
// BenchmarkSetFloat/Tidwall_Nested
// BenchmarkSetFloat/Tidwall_Nested-12         	 1961683	       613.0 ns/op	    1240 B/op	      12 allocs/op
// BenchmarkSetFloat/Grafana_Nested
// BenchmarkSetFloat/Grafana_Nested-12         	 2541321	       474.0 ns/op	     936 B/op	       8 allocs/op
func BenchmarkSetFloat(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
		value    float64
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"height"}, 1.80},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"height"}, 1.80},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "height"}, 1.80},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "height"}, 1.80},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetFloat(jsonCopy, bm.value, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark SetString for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetString
// BenchmarkSetString/Tidwall_Simple
// BenchmarkSetString/Tidwall_Simple-12         	 7182087	       155.2 ns/op	     240 B/op	       6 allocs/op
// BenchmarkSetString/Grafana_Simple
// BenchmarkSetString/Grafana_Simple-12         	13783906	        84.89 ns/op	     144 B/op	       3 allocs/op
// BenchmarkSetString/Tidwall_Nested
// BenchmarkSetString/Tidwall_Nested-12         	 1772337	       675.3 ns/op	    1576 B/op	      11 allocs/op
// BenchmarkSetString/Grafana_Nested
// BenchmarkSetString/Grafana_Nested-12         	 5017585	       233.9 ns/op	     592 B/op	       3 allocs/op
// BenchmarkSetString/Tidwall_Array
// BenchmarkSetString/Tidwall_Array-12          	 2224464	       536.6 ns/op	    1136 B/op	      12 allocs/op
// BenchmarkSetString/Grafana_Array
// BenchmarkSetString/Grafana_Array-12          	 5101358	       247.8 ns/op	     336 B/op	       3 allocs/op
func BenchmarkSetString(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
		value    string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"name"}, "Jane"},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"name"}, "Jane"},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "street"}, "456 Oak Ave"},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "street"}, "456 Oak Ave"},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"users", "[0]", "name"}, "Alice"},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"users", "[0]", "name"}, "Alice"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.SetString(jsonCopy, bm.value, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark DeleteKey for both implementations
// cpu: Apple M2 Pro
// BenchmarkDeleteKey
// BenchmarkDeleteKey/Tidwall_Simple
// BenchmarkDeleteKey/Tidwall_Simple-12         	 8254934	       142.3 ns/op	     208 B/op	       5 allocs/op
// BenchmarkDeleteKey/Grafana_Simple
// BenchmarkDeleteKey/Grafana_Simple-12         	14179249	        83.74 ns/op	     192 B/op	       3 allocs/op
// BenchmarkDeleteKey/Tidwall_Nested
// BenchmarkDeleteKey/Tidwall_Nested-12         	 1766205	       674.1 ns/op	    1560 B/op	      10 allocs/op
// BenchmarkDeleteKey/Grafana_Nested
// BenchmarkDeleteKey/Grafana_Nested-12         	 3318085	       359.3 ns/op	     640 B/op	       3 allocs/op
// BenchmarkDeleteKey/Tidwall_Array
// BenchmarkDeleteKey/Tidwall_Array-12          	 3133135	       380.3 ns/op	     584 B/op	       8 allocs/op
// BenchmarkDeleteKey/Grafana_Array
// BenchmarkDeleteKey/Grafana_Array-12          	 5465682	       218.9 ns/op	     320 B/op	       2 allocs/op
func BenchmarkDeleteKey(b *testing.B) {
	benchmarks := []struct {
		name     string
		parser   JSONParser
		jsonData []byte
		keys     []string
	}{
		{"Tidwall_Simple", NewWithLibrary(TidwallLib), simpleJSON, []string{"name"}},
		{"Grafana_Simple", NewWithLibrary(GrafanaLib), simpleJSON, []string{"name"}},
		{"Tidwall_Nested", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "city"}},
		{"Grafana_Nested", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "city"}},
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"users", "[1]"}},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"users", "[1]"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = bm.parser.DeleteKey(jsonCopy, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}
