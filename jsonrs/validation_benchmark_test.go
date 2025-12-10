package jsonrs_test

import (
	"testing"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"
)

// Complex JSON payload for realistic benchmarking
var validComplexJSON = []byte(`{"users":[{"id":1,"name":"John Doe","email":"john.doe@example.com","age":30,"address":{"street":"123 Main St","city":"Anytown","state":"CA","zipcode":"12345"},"hobbies":["reading","swimming","coding"],"isActive":true,"balance":1234.56},{"id":2,"name":"Jane Smith","email":"jane.smith@example.com","age":25,"address":{"street":"456 Oak Ave","city":"Another City","state":"NY","zipcode":"67890"},"hobbies":["painting","cycling","photography"],"isActive":false,"balance":9876.54}],"metadata":{"created":"2023-01-01T00:00:00Z","updated":"2023-12-31T23:59:59Z","version":1}}`)

// Invalid JSON payload (missing closing brace)
var invalidComplexJSON = []byte(`{"users":[{"id":1,"name":"John Doe","email":"john.doe@example.com","age":30,"address":{"street":"123 Main St","city":"Anytown","state":"CA","zipcode":"12345"},"hobbies":["reading","swimming","coding"],"isActive":true,"balance":1234.56},{"id":2,"name":"Jane Smith","email":"jane.smith@example.com","age":25,"address":{"street":"456 Oak Ave","city":"Another City","state":"NY","zipcode":"67890"},"hobbies":["painting","cycling","photography"],"isActive":false,"balance":9876.54}],"metadata":{"created":"2023-01-01T00:00:00Z","updated":"2023-12-31T23:59:59Z","version":1}`)

// Simple JSON payload
var (
	validSimpleJSON   = []byte(`{"name":"John","age":30}`)
	invalidSimpleJSON = []byte(`{"name":"John","age":}`)
)

// runValidationBenchmarks runs benchmarks for all libraries with given data
func runValidationBenchmarks(b *testing.B, data []byte) {
	benchmarks := []struct {
		name string
		lib  string
	}{
		{"StdLib", jsonrs.StdLib},
		{"SonnetLib", jsonrs.SonnetLib},
		{"JsoniterLib", jsonrs.JsoniterLib},
		{"TidwallLib", jsonrs.TidwallLib},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			j := jsonrs.NewValidatorWithLibrary(bm.lib)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				j.Valid(data)
			}
		})
	}
}

// BenchmarkValidationComplexValid
// BenchmarkValidationComplexValid/StdLib
// BenchmarkValidationComplexValid/StdLib-12         	  717057	      1650 ns/op	       0 B/op	       0 allocs/op
// BenchmarkValidationComplexValid/SonnetLib
// BenchmarkValidationComplexValid/SonnetLib-12      	 1841227	       639.7 ns/op	       0 B/op	       0 allocs/op
// BenchmarkValidationComplexValid/JsoniterLib
// BenchmarkValidationComplexValid/JsoniterLib-12    	 1000000	      1008 ns/op	     192 B/op	      29 allocs/op
// BenchmarkValidationComplexValid/TidwallLib
// BenchmarkValidationComplexValid/TidwallLib-12     	 1899276	       625.2 ns/op	       0 B/op	       0 allocs/op
func BenchmarkValidationComplexValid(b *testing.B) {
	runValidationBenchmarks(b, validComplexJSON)
}

// BenchmarkValidationComplexInvalid
// BenchmarkValidationComplexInvalid/StdLib
// BenchmarkValidationComplexInvalid/StdLib-12         	  724369	      1619 ns/op	      24 B/op	       1 allocs/op
// BenchmarkValidationComplexInvalid/SonnetLib
// BenchmarkValidationComplexInvalid/SonnetLib-12      	 1845439	       657.5 ns/op	      24 B/op	       1 allocs/op
// BenchmarkValidationComplexInvalid/JsoniterLib
// BenchmarkValidationComplexInvalid/JsoniterLib-12    	  933415	      1239 ns/op	     512 B/op	      37 allocs/op
// BenchmarkValidationComplexInvalid/TidwallLib
// BenchmarkValidationComplexInvalid/TidwallLib-12     	 1934650	       618.8 ns/op	       0 B/op	       0 allocs/op
func BenchmarkValidationComplexInvalid(b *testing.B) {
	runValidationBenchmarks(b, invalidComplexJSON)
}

// BenchmarkValidationSimpleValid
// BenchmarkValidationSimpleValid/StdLib
// BenchmarkValidationSimpleValid/StdLib-12         	14405204	        84.00 ns/op	       0 B/op	       0 allocs/op
// BenchmarkValidationSimpleValid/SonnetLib
// BenchmarkValidationSimpleValid/SonnetLib-12      	35917478	        33.39 ns/op	       0 B/op	       0 allocs/op
// BenchmarkValidationSimpleValid/JsoniterLib
// BenchmarkValidationSimpleValid/JsoniterLib-12    	15764590	        74.25 ns/op	       8 B/op	       2 allocs/op
// BenchmarkValidationSimpleValid/TidwallLib
// BenchmarkValidationSimpleValid/TidwallLib-12     	36072866	        34.08 ns/op	       0 B/op	       0 allocs/op
func BenchmarkValidationSimpleValid(b *testing.B) {
	runValidationBenchmarks(b, validSimpleJSON)
}

// BenchmarkValidationSimpleInvalid
// BenchmarkValidationSimpleInvalid/StdLib
// BenchmarkValidationSimpleInvalid/StdLib-12         	 6881553	       173.3 ns/op	     104 B/op	       4 allocs/op
// BenchmarkValidationSimpleInvalid/SonnetLib
// BenchmarkValidationSimpleInvalid/SonnetLib-12      	12432151	        99.46 ns/op	      96 B/op	       3 allocs/op
// BenchmarkValidationSimpleInvalid/JsoniterLib
// BenchmarkValidationSimpleInvalid/JsoniterLib-12    	 3243615	       370.1 ns/op	     285 B/op	      11 allocs/op
// BenchmarkValidationSimpleInvalid/TidwallLib
// BenchmarkValidationSimpleInvalid/TidwallLib-12     	39805230	        29.96 ns/op	       0 B/op	       0 allocs/op
func BenchmarkValidationSimpleInvalid(b *testing.B) {
	runValidationBenchmarks(b, invalidSimpleJSON)
}
