package jsonparser

import (
	"testing"
)

// Sample JSON data for benchmarks
var (
	simpleJSON = []byte(`{"name": "John", "age": 30, "isActive": true, "height": 1.75, "strBoolean": "true", "intString": "42"}`)
	nestedJSON = []byte(`{
		"user": {
			"name": "John",
			"age": 30,
			"isActive": true,
			"height": 1.75,
			"address": {
				"street": "123 Main St",
				"city": "New York",
				"zipcode": "10001",
				"house": "14.12",
				"available": "false"
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
// BenchmarkGetValue/Tidwall_Simple-12         	16043620	        72.13 ns/op	      32 B/op	       3 allocs/op
// BenchmarkGetValue/Grafana_Simple
// BenchmarkGetValue/Grafana_Simple-12         	29369833	        41.58 ns/op	       8 B/op	       1 allocs/op
// BenchmarkGetValue/Tidwall_Nested
// BenchmarkGetValue/Tidwall_Nested-12         	 4781750	       245.4 ns/op	     104 B/op	       4 allocs/op
// BenchmarkGetValue/Grafana_Nested
// BenchmarkGetValue/Grafana_Nested-12         	 7085900	       167.7 ns/op	      16 B/op	       1 allocs/op
// BenchmarkGetValue/Tidwall_Array
// BenchmarkGetValue/Tidwall_Array-12          	 7087020	       169.3 ns/op	      80 B/op	       4 allocs/op
// BenchmarkGetValue/Grafana_Array
// BenchmarkGetValue/Grafana_Array-12          	 7049780	       170.2 ns/op	       8 B/op	       1 allocs/op
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

// cpu: Apple M2 Pro
// BenchmarkGetValueOrEmpty
// BenchmarkGetValueOrEmpty/Tidwall_Simple
// BenchmarkGetValueOrEmpty/Tidwall_Simple-12         	14146834	        71.27 ns/op	      32 B/op	       3 allocs/op
// BenchmarkGetValueOrEmpty/Grafana_Simple
// BenchmarkGetValueOrEmpty/Grafana_Simple-12         	26678152	        44.22 ns/op	       8 B/op	       1 allocs/op
// BenchmarkGetValueOrEmpty/Tidwall_Nested
// BenchmarkGetValueOrEmpty/Tidwall_Nested-12         	 4897113	       246.2 ns/op	     104 B/op	       4 allocs/op
// BenchmarkGetValueOrEmpty/Grafana_Nested
// BenchmarkGetValueOrEmpty/Grafana_Nested-12         	 7006022	       168.6 ns/op	      16 B/op	       1 allocs/op
// BenchmarkGetValueOrEmpty/Tidwall_Array
// BenchmarkGetValueOrEmpty/Tidwall_Array-12          	 8032423	       144.1 ns/op	      80 B/op	       4 allocs/op
// BenchmarkGetValueOrEmpty/Grafana_Array
// BenchmarkGetValueOrEmpty/Grafana_Array-12          	 6939673	       170.8 ns/op	       8 B/op	       1 allocs/op
func BenchmarkGetValueOrEmpty(b *testing.B) {
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
		{"Tidwall_Array", NewWithLibrary(TidwallLib), arrayJSON, []string{"users", "[0]", "name"}},
		{"Grafana_Array", NewWithLibrary(GrafanaLib), arrayJSON, []string{"users", "[0]", "name"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.parser.GetValueOrEmpty(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetBoolean for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetBoolean
// BenchmarkGetBoolean/Tidwall_Simple
// BenchmarkGetBoolean/Tidwall_Simple-12         	12143515	        96.84 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetBoolean/Grafana_Simple
// BenchmarkGetBoolean/Grafana_Simple-12         	21886494	        53.35 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetBoolean/Tidwall_Nested
// BenchmarkGetBoolean/Tidwall_Nested-12         	 4989630	       240.1 ns/op	      68 B/op	       3 allocs/op
// BenchmarkGetBoolean/Grafana_Nested
// BenchmarkGetBoolean/Grafana_Nested-12         	 4426470	       267.5 ns/op	       0 B/op	       0 allocs/op
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

// Benchmark GetBooleanOrFalse for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetBooleanOrFalse
// BenchmarkGetBooleanOrFalse/Tidwall_Simple
// BenchmarkGetBooleanOrFalse/Tidwall_Simple-12         	12475767	        94.91 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetBooleanOrFalse/Grafana_Simple
// BenchmarkGetBooleanOrFalse/Grafana_Simple-12         	20433153	        59.70 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetBooleanOrFalse/Tidwall_Nested
// BenchmarkGetBooleanOrFalse/Tidwall_Nested-12         	 5053959	       240.3 ns/op	      68 B/op	       3 allocs/op
// BenchmarkGetBooleanOrFalse/Grafana_Nested
// BenchmarkGetBooleanOrFalse/Grafana_Nested-12         	 4297123	       272.4 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetBooleanOrFalse/Tidwall__SimpleStrBool
// BenchmarkGetBooleanOrFalse/Tidwall__SimpleStrBool-12 	 9240284	       129.1 ns/op	      24 B/op	       2 allocs/op
// BenchmarkGetBooleanOrFalse/Grafana__SimpleStrBool
// BenchmarkGetBooleanOrFalse/Grafana__SimpleStrBool-12 	13376919	        88.05 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetBooleanOrFalse/Tidwall__NestedStrBool
// BenchmarkGetBooleanOrFalse/Tidwall__NestedStrBool-12 	 4111076	       295.1 ns/op	      80 B/op	       3 allocs/op
// BenchmarkGetBooleanOrFalse/Grafana__NestedStrBool
// BenchmarkGetBooleanOrFalse/Grafana__NestedStrBool-12 	 5608251	       212.0 ns/op	       0 B/op	       0 allocs/op
func BenchmarkGetBooleanOrFalse(b *testing.B) {
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
		{"Tidwall__SimpleStrBool", NewWithLibrary(TidwallLib), simpleJSON, []string{"strBoolean"}},
		{"Grafana__SimpleStrBool", NewWithLibrary(GrafanaLib), simpleJSON, []string{"strBoolean"}},
		{"Tidwall__NestedStrBool", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "available"}},
		{"Grafana__NestedStrBool", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "available"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.parser.GetBooleanOrFalse(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetInt for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetInt
// BenchmarkGetInt/Tidwall_Simple
// BenchmarkGetInt/Tidwall_Simple-12         	12105106	        94.73 ns/op	      18 B/op	       2 allocs/op
// BenchmarkGetInt/Grafana_Simple
// BenchmarkGetInt/Grafana_Simple-12         	24470610	        49.18 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetInt/Tidwall_Nested
// BenchmarkGetInt/Tidwall_Nested-12         	 8460069	       137.2 ns/op	      48 B/op	       3 allocs/op
// BenchmarkGetInt/Grafana_Nested
// BenchmarkGetInt/Grafana_Nested-12         	17596414	        66.91 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetInt/Tidwall_Array
// BenchmarkGetInt/Tidwall_Array-12          	 6082740	       199.9 ns/op	      48 B/op	       3 allocs/op
// BenchmarkGetInt/Grafana_Array
// BenchmarkGetInt/Grafana_Array-12          	 5860048	       207.8 ns/op	       0 B/op	       0 allocs/op
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

// Benchmark GetIntOrZero for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetIntOrZero
// BenchmarkGetIntOrZero/Tidwall_Simple
// BenchmarkGetIntOrZero/Tidwall_Simple-12         	11955668	        90.32 ns/op	      18 B/op	       2 allocs/op
// BenchmarkGetIntOrZero/Grafana_Simple
// BenchmarkGetIntOrZero/Grafana_Simple-12         	29360461	        40.46 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetIntOrZero/Tidwall_Nested
// BenchmarkGetIntOrZero/Tidwall_Nested-12         	 8742595	       138.0 ns/op	      48 B/op	       3 allocs/op
// BenchmarkGetIntOrZero/Grafana_Nested
// BenchmarkGetIntOrZero/Grafana_Nested-12         	20687350	        58.29 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetIntOrZero/Tidwall_Array
// BenchmarkGetIntOrZero/Tidwall_Array-12          	 6063993	       196.2 ns/op	      48 B/op	       3 allocs/op
// BenchmarkGetIntOrZero/Grafana_Array
// BenchmarkGetIntOrZero/Grafana_Array-12          	 6158414	       193.7 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetIntOrZero/Tidwall__SimpleStrInt
// BenchmarkGetIntOrZero/Tidwall__SimpleStrInt-12  	 8573775	       140.4 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetIntOrZero/Grafana__SimpleStrInt
// BenchmarkGetIntOrZero/Grafana__SimpleStrInt-12  	10829937	       111.1 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetIntOrZero/Tidwall__NestedStrInt
// BenchmarkGetIntOrZero/Tidwall__NestedStrInt-12  	 4671104	       271.0 ns/op	      80 B/op	       3 allocs/op
// BenchmarkGetIntOrZero/Grafana__NestedStrInt
// BenchmarkGetIntOrZero/Grafana__NestedStrInt-12  	 6516843	       182.3 ns/op	       0 B/op	       0 allocs/op
func BenchmarkGetIntOrZero(b *testing.B) {
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
		{"Tidwall__SimpleStrInt", NewWithLibrary(TidwallLib), simpleJSON, []string{"intString"}},
		{"Grafana__SimpleStrInt", NewWithLibrary(GrafanaLib), simpleJSON, []string{"intString"}},
		{"Tidwall__NestedStrInt", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "zipcode"}},
		{"Grafana__NestedStrInt", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "zipcode"}},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.parser.GetIntOrZero(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetFloat for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetFloat
// BenchmarkGetFloat/Tidwall_Simple
// BenchmarkGetFloat/Tidwall_Simple-12         	 8643547	       123.6 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetFloat/Grafana_Simple
// BenchmarkGetFloat/Grafana_Simple-12         	14293030	        82.96 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetFloat/Tidwall_Nested
// BenchmarkGetFloat/Tidwall_Nested-12         	 6796100	       176.2 ns/op	      52 B/op	       3 allocs/op
// BenchmarkGetFloat/Grafana_Nested
// BenchmarkGetFloat/Grafana_Nested-12         	11421079	       105.0 ns/op	       0 B/op	       0 allocs/op
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

// Benchmark GetFloatOrZero for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetFloatOrZero
// BenchmarkGetFloatOrZero/Tidwall_Simple
// BenchmarkGetFloatOrZero/Tidwall_Simple-12         	 9558728	       124.5 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetFloatOrZero/Grafana_Simple
// BenchmarkGetFloatOrZero/Grafana_Simple-12         	14099806	        83.53 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetFloatOrZero/Tidwall_Nested
// BenchmarkGetFloatOrZero/Tidwall_Nested-12         	 6808833	       176.8 ns/op	      52 B/op	       3 allocs/op
// BenchmarkGetFloatOrZero/Grafana_Nested
// BenchmarkGetFloatOrZero/Grafana_Nested-12         	11301868	       104.4 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetFloatOrZero/Tidwall__SimpleStrInt
// BenchmarkGetFloatOrZero/Tidwall__SimpleStrInt-12  	 8035160	       149.1 ns/op	      20 B/op	       2 allocs/op
// BenchmarkGetFloatOrZero/Grafana__SimpleStrInt
// BenchmarkGetFloatOrZero/Grafana__SimpleStrInt-12  	10788976	       118.2 ns/op	       0 B/op	       0 allocs/op
// BenchmarkGetFloatOrZero/Tidwall__NestedStrInt
// BenchmarkGetFloatOrZero/Tidwall__NestedStrInt-12  	 4257519	       283.0 ns/op	      80 B/op	       3 allocs/op
// BenchmarkGetFloatOrZero/Grafana__NestedStrInt
// BenchmarkGetFloatOrZero/Grafana__NestedStrInt-12  	 5840556	       200.6 ns/op	       0 B/op	       0 allocs/op
func BenchmarkGetFloatOrZero(b *testing.B) {
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
		{"Tidwall__SimpleStrInt", NewWithLibrary(TidwallLib), simpleJSON, []string{"intString"}},
		{"Grafana__SimpleStrInt", NewWithLibrary(GrafanaLib), simpleJSON, []string{"intString"}},
		{"Tidwall__NestedStrInt", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "address", "house"}},
		{"Grafana__NestedStrInt", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "address", "house"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.parser.GetFloatOrZero(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark GetString for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetString
// BenchmarkGetString/Tidwall_Simple
// BenchmarkGetString/Tidwall_Simple-12         	15520761	        71.72 ns/op	      24 B/op	       2 allocs/op
// BenchmarkGetString/Grafana_Simple
// BenchmarkGetString/Grafana_Simple-12         	38759308	        30.44 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetString/Tidwall_Nested
// BenchmarkGetString/Tidwall_Nested-12         	 5460160	       231.1 ns/op	      88 B/op	       3 allocs/op
// BenchmarkGetString/Grafana_Nested
// BenchmarkGetString/Grafana_Nested-12         	 8209220	       144.8 ns/op	      16 B/op	       1 allocs/op
// BenchmarkGetString/Tidwall_Array
// BenchmarkGetString/Tidwall_Array-12          	 8374755	       143.6 ns/op	      72 B/op	       3 allocs/op
// BenchmarkGetString/Grafana_Array
// BenchmarkGetString/Grafana_Array-12          	 7470157	       159.0 ns/op	       4 B/op	       1 allocs/op
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

// Benchmark GetStringOrEmpty for both implementations
// cpu: Apple M2 Pro
// BenchmarkGetStringOrEmpty
// BenchmarkGetStringOrEmpty/Tidwall_Simple
// BenchmarkGetStringOrEmpty/Tidwall_Simple-12         	14352879	        72.72 ns/op	      24 B/op	       2 allocs/op
// BenchmarkGetStringOrEmpty/Grafana_Simple
// BenchmarkGetStringOrEmpty/Grafana_Simple-12         	36492742	        33.14 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetStringOrEmpty/Tidwall_Nested
// BenchmarkGetStringOrEmpty/Tidwall_Nested-12         	 5333312	       220.4 ns/op	      88 B/op	       3 allocs/op
// BenchmarkGetStringOrEmpty/Grafana_Nested
// BenchmarkGetStringOrEmpty/Grafana_Nested-12         	 8620735	       139.7 ns/op	      16 B/op	       1 allocs/op
// BenchmarkGetStringOrEmpty/Tidwall_Array
// BenchmarkGetStringOrEmpty/Tidwall_Array-12          	 8223603	       144.6 ns/op	      72 B/op	       3 allocs/op
// BenchmarkGetStringOrEmpty/Grafana_Array
// BenchmarkGetStringOrEmpty/Grafana_Array-12          	 7379373	       160.2 ns/op	       4 B/op	       1 allocs/op
// BenchmarkGetStringOrEmpty/Tidwall__SimpleStrInt
// BenchmarkGetStringOrEmpty/Tidwall__SimpleStrInt-12  	12876675	        91.34 ns/op	      18 B/op	       2 allocs/op
// BenchmarkGetStringOrEmpty/Grafana__SimpleStrInt
// BenchmarkGetStringOrEmpty/Grafana__SimpleStrInt-12  	25136000	        48.07 ns/op	       2 B/op	       1 allocs/op
// BenchmarkGetStringOrEmpty/Tidwall__NestedStrInt
// BenchmarkGetStringOrEmpty/Tidwall__NestedStrInt-12  	 8624607	       138.2 ns/op	      48 B/op	       3 allocs/op
// BenchmarkGetStringOrEmpty/Grafana__NestedStrInt
// BenchmarkGetStringOrEmpty/Grafana__NestedStrInt-12  	17998604	        66.83 ns/op	       2 B/op	       1 allocs/op
func BenchmarkGetStringOrEmpty(b *testing.B) {
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
		{"Tidwall__SimpleStrInt", NewWithLibrary(TidwallLib), simpleJSON, []string{"age"}},
		{"Grafana__SimpleStrInt", NewWithLibrary(GrafanaLib), simpleJSON, []string{"age"}},
		{"Tidwall__NestedStrInt", NewWithLibrary(TidwallLib), nestedJSON, []string{"user", "age"}},
		{"Grafana__NestedStrInt", NewWithLibrary(GrafanaLib), nestedJSON, []string{"user", "age"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.parser.GetStringOrEmpty(bm.jsonData, bm.keys...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark SetValue for both implementations
// cpu: Apple M2 Pro
// BenchmarkSetValue
// BenchmarkSetValue/Tidwall_Simple
// BenchmarkSetValue/Tidwall_Simple-12         	 7641900	       153.4 ns/op	     320 B/op	       5 allocs/op
// BenchmarkSetValue/Grafana_Simple
// BenchmarkSetValue/Grafana_Simple-12         	13798303	        82.53 ns/op	     224 B/op	       2 allocs/op
// BenchmarkSetValue/Tidwall_Nested
// BenchmarkSetValue/Tidwall_Nested-12         	 1660099	       715.8 ns/op	    1592 B/op	      10 allocs/op
// BenchmarkSetValue/Grafana_Nested
// BenchmarkSetValue/Grafana_Nested-12         	 4780231	       252.8 ns/op	     640 B/op	       2 allocs/op
// BenchmarkSetValue/Tidwall_Array
// BenchmarkSetValue/Tidwall_Array-12          	 2147253	       577.0 ns/op	    1088 B/op	      10 allocs/op
// BenchmarkSetValue/Grafana_Array
// BenchmarkSetValue/Grafana_Array-12          	 5353195	       231.3 ns/op	     320 B/op	       2 allocs/op
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
// BenchmarkSetBoolean/Tidwall_Simple-12         	 6724695	       179.8 ns/op	     352 B/op	       5 allocs/op
// BenchmarkSetBoolean/Grafana_Simple
// BenchmarkSetBoolean/Grafana_Simple-12         	12107061	        96.64 ns/op	     224 B/op	       2 allocs/op
// BenchmarkSetBoolean/Tidwall_Nested
// BenchmarkSetBoolean/Tidwall_Nested-12         	 2260214	       511.1 ns/op	    1328 B/op	       7 allocs/op
// BenchmarkSetBoolean/Grafana_Nested
// BenchmarkSetBoolean/Grafana_Nested-12         	 3423878	       350.5 ns/op	     640 B/op	       2 allocs/op
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
// BenchmarkSetInt
// BenchmarkSetInt/Tidwall_Simple
// BenchmarkSetInt/Tidwall_Simple-12         	 5715664	       221.5 ns/op	     376 B/op	       6 allocs/op
// BenchmarkSetInt/Grafana_Simple
// BenchmarkSetInt/Grafana_Simple-12         	12599146	        91.53 ns/op	     224 B/op	       2 allocs/op
// BenchmarkSetInt/Tidwall_Nested
// BenchmarkSetInt/Tidwall_Nested-12         	 2319024	       508.8 ns/op	    1320 B/op	       9 allocs/op
// BenchmarkSetInt/Grafana_Nested
// BenchmarkSetInt/Grafana_Nested-12         	 8070174	       150.8 ns/op	     640 B/op	       2 allocs/op
// BenchmarkSetInt/Tidwall_Array
// BenchmarkSetInt/Tidwall_Array-12          	 2832908	       428.6 ns/op	     760 B/op	       7 allocs/op
// BenchmarkSetInt/Grafana_Array
// BenchmarkSetInt/Grafana_Array-12          	 4777756	       253.5 ns/op	     320 B/op	       2 allocs/op
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
// BenchmarkSetFloat/Tidwall_Simple-12         	 3532658	       331.3 ns/op	     424 B/op	       8 allocs/op
// BenchmarkSetFloat/Grafana_Simple
// BenchmarkSetFloat/Grafana_Simple-12         	 5967984	       197.3 ns/op	     256 B/op	       4 allocs/op
// BenchmarkSetFloat/Tidwall_Nested
// BenchmarkSetFloat/Tidwall_Nested-12         	 1852850	       632.0 ns/op	    1416 B/op	      12 allocs/op
// BenchmarkSetFloat/Grafana_Nested
// BenchmarkSetFloat/Grafana_Nested-12         	 4654510	       258.1 ns/op	     672 B/op	       4 allocs/op
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
// BenchmarkSetString/Tidwall_Simple-12         	 6697904	       169.8 ns/op	     336 B/op	       6 allocs/op
// BenchmarkSetString/Grafana_Simple
// BenchmarkSetString/Grafana_Simple-12         	12084946	        98.76 ns/op	     240 B/op	       3 allocs/op
// BenchmarkSetString/Tidwall_Nested
// BenchmarkSetString/Tidwall_Nested-12         	 1620744	       729.9 ns/op	    1608 B/op	      11 allocs/op
// BenchmarkSetString/Grafana_Nested
// BenchmarkSetString/Grafana_Nested-12         	 4853798	       249.7 ns/op	     656 B/op	       3 allocs/op
// BenchmarkSetString/Tidwall_Array
// BenchmarkSetString/Tidwall_Array-12          	 2076668	       554.5 ns/op	    1136 B/op	      12 allocs/op
// BenchmarkSetString/Grafana_Array
// BenchmarkSetString/Grafana_Array-12          	 4925148	       241.5 ns/op	     336 B/op	       3 allocs/op
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
// BenchmarkDeleteKey/Tidwall_Simple-12         	 7168490	       151.9 ns/op	     304 B/op	       5 allocs/op
// BenchmarkDeleteKey/Grafana_Simple
// BenchmarkDeleteKey/Grafana_Simple-12         	12347451	        99.32 ns/op	     288 B/op	       3 allocs/op
// BenchmarkDeleteKey/Tidwall_Nested
// BenchmarkDeleteKey/Tidwall_Nested-12         	 1633711	       726.4 ns/op	    1592 B/op	      10 allocs/op
// BenchmarkDeleteKey/Grafana_Nested
// BenchmarkDeleteKey/Grafana_Nested-12         	 2902386	       406.6 ns/op	     704 B/op	       3 allocs/op
// BenchmarkDeleteKey/Tidwall_Array
// BenchmarkDeleteKey/Tidwall_Array-12          	 3008270	       446.0 ns/op	     584 B/op	       8 allocs/op
// BenchmarkDeleteKey/Grafana_Array
// BenchmarkDeleteKey/Grafana_Array-12          	 5273028	       221.2 ns/op	     320 B/op	       2 allocs/op
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
