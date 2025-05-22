package jsonparser

import (
	"testing"

	"github.com/grafana/jsonparser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Benchmark raw gjson vs jsonparser Get operations
// cpu: Apple M2 Pro
// BenchmarkRawGetComparison
// BenchmarkRawGetComparison/Gjson_Get_Simple
// BenchmarkRawGetComparison/Gjson_Get_Simple-12         	26274926	        43.55 ns/op	       8 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_Get_Simple
// BenchmarkRawGetComparison/Jsonparser_Get_Simple-12    	70055266	        17.24 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_Get_Nested
// BenchmarkRawGetComparison/Gjson_Get_Nested-12         	 7047067	       167.7 ns/op	      16 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_Get_Nested
// BenchmarkRawGetComparison/Jsonparser_Get_Nested-12    	 8437245	       141.4 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_Get_Array
// BenchmarkRawGetComparison/Gjson_Get_Array-12          	11985482	        99.56 ns/op	       8 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_Get_Array
// BenchmarkRawGetComparison/Jsonparser_Get_Array-12     	 8237208	       143.6 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetString_Simple
// BenchmarkRawGetComparison/Gjson_GetString_Simple-12   	22952248	        52.15 ns/op	       8 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetString_Simple
// BenchmarkRawGetComparison/Jsonparser_GetString_Simple-12         	39349422	        29.81 ns/op	       4 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Gjson_GetString_Nested
// BenchmarkRawGetComparison/Gjson_GetString_Nested-12              	 7387480	       159.4 ns/op	      16 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetString_Nested
// BenchmarkRawGetComparison/Jsonparser_GetString_Nested-12         	 8652967	       135.2 ns/op	      16 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Gjson_GetString_Array
// BenchmarkRawGetComparison/Gjson_GetString_Array-12               	13925919	        85.78 ns/op	       8 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetString_Array
// BenchmarkRawGetComparison/Jsonparser_GetString_Array-12          	 7486746	       158.9 ns/op	       4 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Gjson_GetInt_Simple
// BenchmarkRawGetComparison/Gjson_GetInt_Simple-12                 	16314820	        72.17 ns/op	       2 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetInt_Simple
// BenchmarkRawGetComparison/Jsonparser_GetInt_Simple-12            	30300894	        40.56 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetInt_Nested
// BenchmarkRawGetComparison/Gjson_GetInt_Nested-12                 	12682888	        94.14 ns/op	       2 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetInt_Nested
// BenchmarkRawGetComparison/Jsonparser_GetInt_Nested-12            	20247824	        59.04 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetInt_Array
// BenchmarkRawGetComparison/Gjson_GetInt_Array-12                  	 7790226	       153.9 ns/op	       2 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetInt_Array
// BenchmarkRawGetComparison/Jsonparser_GetInt_Array-12             	 6108822	       197.0 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetBool_Simple
// BenchmarkRawGetComparison/Gjson_GetBool_Simple-12                	16506794	        73.28 ns/op	       4 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetBool_Simple
// BenchmarkRawGetComparison/Jsonparser_GetBool_Simple-12           	22443976	        52.36 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetBool_Nested
// BenchmarkRawGetComparison/Gjson_GetBool_Nested-12                	 6847804	       173.5 ns/op	       4 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetBool_Nested
// BenchmarkRawGetComparison/Jsonparser_GetBool_Nested-12           	 5173969	       227.1 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetBool_Nested2
// BenchmarkRawGetComparison/Gjson_GetBool_Nested2-12               	12065697	        98.12 ns/op	       4 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetBool_Nested2
// BenchmarkRawGetComparison/Jsonparser_GetBool_Nested2-12          	16470648	        72.62 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetFloat_Simple
// BenchmarkRawGetComparison/Gjson_GetFloat_Simple-12               	11554014	       104.1 ns/op	       4 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetFloat_Simple
// BenchmarkRawGetComparison/Jsonparser_GetFloat_Simple-12          	14588438	        82.30 ns/op	       0 B/op	       0 allocs/op
// BenchmarkRawGetComparison/Gjson_GetFloat_Nested
// BenchmarkRawGetComparison/Gjson_GetFloat_Nested-12               	 9448629	       126.9 ns/op	       4 B/op	       1 allocs/op
// BenchmarkRawGetComparison/Jsonparser_GetFloat_Nested
// BenchmarkRawGetComparison/Jsonparser_GetFloat_Nested-12          	11491615	       103.7 ns/op	       0 B/op	       0 allocs/op
func BenchmarkRawGetComparison(b *testing.B) {
	type testCase struct {
		name          string
		jsonData      []byte
		gjsonKey      string
		jsonparserKey []string
		op            string
	}

	// Define all test cases
	var benchmarks []testCase

	// Get operation test cases
	getTests := []testCase{
		{"Get_Simple", simpleJSON, "name", []string{"name"}, "Get"},
		{"Get_Nested", nestedJSON, "user.address.city", []string{"user", "address", "city"}, "Get"},
		{"Get_Array", arrayJSON, "users.1.name", []string{"users", "[1]", "name"}, "Get"},
	}
	benchmarks = append(benchmarks, getTests...)

	// GetString operation test cases
	getStringTests := []testCase{
		{"GetString_Simple", simpleJSON, "name", []string{"name"}, "GetString"},
		{"GetString_Nested", nestedJSON, "user.address.street", []string{"user", "address", "street"}, "GetString"},
		{"GetString_Array", arrayJSON, "users.0.name", []string{"users", "[0]", "name"}, "GetString"},
	}
	benchmarks = append(benchmarks, getStringTests...)

	// GetInt operation test cases
	getIntTests := []testCase{
		{"GetInt_Simple", simpleJSON, "age", []string{"age"}, "GetInt"},
		{"GetInt_Nested", nestedJSON, "user.age", []string{"user", "age"}, "GetInt"},
		{"GetInt_Array", arrayJSON, "scores.2", []string{"scores", "[2]"}, "GetInt"},
	}
	benchmarks = append(benchmarks, getIntTests...)

	// GetBool operation test cases
	getBoolTests := []testCase{
		{"GetBool_Simple", simpleJSON, "isActive", []string{"isActive"}, "GetBool"},
		{"GetBool_Nested", nestedJSON, "preferences.notifications", []string{"preferences", "notifications"}, "GetBool"},
		{"GetBool_Nested2", nestedJSON, "user.isActive", []string{"user", "isActive"}, "GetBool"},
	}
	benchmarks = append(benchmarks, getBoolTests...)

	// GetFloat operation test cases
	getFloatTests := []testCase{
		{"GetFloat_Simple", simpleJSON, "height", []string{"height"}, "GetFloat"},
		{"GetFloat_Nested", nestedJSON, "user.height", []string{"user", "height"}, "GetFloat"},
	}
	benchmarks = append(benchmarks, getFloatTests...)

	for _, bm := range benchmarks {
		// Benchmark gjson
		b.Run("Gjson_"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				switch bm.op {
				case "Get":
					_ = gjson.GetBytes(bm.jsonData, bm.gjsonKey)
				case "GetString":
					_ = gjson.GetBytes(bm.jsonData, bm.gjsonKey).String()
				case "GetInt":
					_ = gjson.GetBytes(bm.jsonData, bm.gjsonKey).Int()
				case "GetBool":
					_ = gjson.GetBytes(bm.jsonData, bm.gjsonKey).Bool()
				case "GetFloat":
					_ = gjson.GetBytes(bm.jsonData, bm.gjsonKey).Float()
				}
			}
			b.ReportAllocs()
		})

		// Benchmark jsonparser
		b.Run("Jsonparser_"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				switch bm.op {
				case "Get":
					_, _, _, _ = jsonparser.Get(bm.jsonData, bm.jsonparserKey...)
				case "GetString":
					_, _ = jsonparser.GetString(bm.jsonData, bm.jsonparserKey...)
				case "GetInt":
					_, _ = jsonparser.GetInt(bm.jsonData, bm.jsonparserKey...)
				case "GetBool":
					_, _ = jsonparser.GetBoolean(bm.jsonData, bm.jsonparserKey...)
				case "GetFloat":
					_, _ = jsonparser.GetFloat(bm.jsonData, bm.jsonparserKey...)
				}
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark raw sjson vs jsonparser Set operations
// cpu: Apple M2 Pro
// BenchmarkRawSetComparison
// BenchmarkRawSetComparison/Sjson_Set_Simple
// BenchmarkRawSetComparison/Sjson_Set_Simple-12         	 8134503	       138.8 ns/op	     224 B/op	       5 allocs/op
// BenchmarkRawSetComparison/Jsonparser_Set_Simple
// BenchmarkRawSetComparison/Jsonparser_Set_Simple-12    	22295629	        52.79 ns/op	     128 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_Set_Nested
// BenchmarkRawSetComparison/Sjson_Set_Nested-12         	 1846384	       661.6 ns/op	    1504 B/op	       9 allocs/op
// BenchmarkRawSetComparison/Jsonparser_Set_Nested
// BenchmarkRawSetComparison/Jsonparser_Set_Nested-12    	 5298859	       215.4 ns/op	     576 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_Set_Array
// BenchmarkRawSetComparison/Sjson_Set_Array-12          	 2247096	       519.6 ns/op	    1040 B/op	       9 allocs/op
// BenchmarkRawSetComparison/Jsonparser_Set_Array
// BenchmarkRawSetComparison/Jsonparser_Set_Array-12     	 5725411	       209.4 ns/op	     320 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetString_Simple
// BenchmarkRawSetComparison/Sjson_SetString_Simple-12   	 7974922	       140.1 ns/op	     224 B/op	       5 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetString_Simple
// BenchmarkRawSetComparison/Jsonparser_SetString_Simple-12         	22440060	        52.36 ns/op	     128 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetString_Nested
// BenchmarkRawSetComparison/Sjson_SetString_Nested-12              	 1824082	       677.0 ns/op	    1504 B/op	       9 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetString_Nested
// BenchmarkRawSetComparison/Jsonparser_SetString_Nested-12         	 5640355	       206.5 ns/op	     576 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetString_Array
// BenchmarkRawSetComparison/Sjson_SetString_Array-12               	 2395567	       499.2 ns/op	    1072 B/op	      10 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetString_Array
// BenchmarkRawSetComparison/Jsonparser_SetString_Array-12          	 6009832	       203.5 ns/op	     320 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetInt_Simple
// BenchmarkRawSetComparison/Sjson_SetInt_Simple-12                 	 6841435	       166.3 ns/op	     296 B/op	       5 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetInt_Simple
// BenchmarkRawSetComparison/Jsonparser_SetInt_Simple-12            	17185400	        69.15 ns/op	     128 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetInt_Nested
// BenchmarkRawSetComparison/Sjson_SetInt_Nested-12                 	 2913991	       405.7 ns/op	    1104 B/op	       7 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetInt_Nested
// BenchmarkRawSetComparison/Jsonparser_SetInt_Nested-12            	 9290992	       131.0 ns/op	     576 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetInt_Array
// BenchmarkRawSetComparison/Sjson_SetInt_Array-12                  	 3347336	       358.0 ns/op	     720 B/op	       5 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetInt_Array
// BenchmarkRawSetComparison/Jsonparser_SetInt_Array-12             	 4652617	       246.6 ns/op	     320 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetBool_Simple
// BenchmarkRawSetComparison/Sjson_SetBool_Simple-12                	 7364154	       161.4 ns/op	     272 B/op	       4 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetBool_Simple
// BenchmarkRawSetComparison/Jsonparser_SetBool_Simple-12           	14156611	        83.93 ns/op	     128 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetBool_Nested
// BenchmarkRawSetComparison/Sjson_SetBool_Nested-12                	 2852331	       422.3 ns/op	    1088 B/op	       5 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetBool_Nested
// BenchmarkRawSetComparison/Jsonparser_SetBool_Nested-12           	 3961231	       306.0 ns/op	     576 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetBool_Nested2
// BenchmarkRawSetComparison/Sjson_SetBool_Nested2-12               	 2514295	       456.8 ns/op	    1440 B/op	       8 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetBool_Nested2
// BenchmarkRawSetComparison/Jsonparser_SetBool_Nested2-12          	 8286582	       147.2 ns/op	     576 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetFloat_Simple
// BenchmarkRawSetComparison/Sjson_SetFloat_Simple-12               	 4744246	       247.8 ns/op	     232 B/op	       6 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetFloat_Simple
// BenchmarkRawSetComparison/Jsonparser_SetFloat_Simple-12          	10988809	       110.6 ns/op	     128 B/op	       2 allocs/op
// BenchmarkRawSetComparison/Sjson_SetFloat_Nested
// BenchmarkRawSetComparison/Sjson_SetFloat_Nested-12               	 2013194	       584.0 ns/op	    1192 B/op	      10 allocs/op
// BenchmarkRawSetComparison/Jsonparser_SetFloat_Nested
// BenchmarkRawSetComparison/Jsonparser_SetFloat_Nested-12          	 6297567	       177.9 ns/op	     576 B/op	       2 allocs/op
func BenchmarkRawSetComparison(b *testing.B) {
	type testCase struct {
		name          string
		jsonData      []byte
		sjsonKey      string
		jsonparserKey []string
		op            string
		strVal        string
		intVal        int64
		boolVal       bool
		floatVal      float64
		jsonparserVal []byte
	}

	// Define all test cases
	var benchmarks []testCase

	// Set operation test cases (generic interface{} value)
	setTests := []testCase{
		{"Set_Simple", simpleJSON, "name", []string{"name"}, "Set", "Jane", 0, false, 0, []byte(`"Jane"`)},
		{"Set_Nested", nestedJSON, "user.address.city", []string{"user", "address", "city"}, "Set", "Boston", 0, false, 0, []byte(`"Boston"`)},
		{"Set_Array", arrayJSON, "users.1.name", []string{"users", "[1]", "name"}, "Set", "Alice", 0, false, 0, []byte(`"Alice"`)},
	}
	benchmarks = append(benchmarks, setTests...)

	// SetString operation test cases
	setStringTests := []testCase{
		{"SetString_Simple", simpleJSON, "name", []string{"name"}, "SetString", "Jane", 0, false, 0, []byte(`"Jane"`)},
		{"SetString_Nested", nestedJSON, "user.address.street", []string{"user", "address", "street"}, "SetString", "456 Oak Ave", 0, false, 0, []byte(`"456 Oak Ave"`)},
		{"SetString_Array", arrayJSON, "users.0.name", []string{"users", "[0]", "name"}, "SetString", "Alice", 0, false, 0, []byte(`"Alice"`)},
	}
	benchmarks = append(benchmarks, setStringTests...)

	// SetInt operation test cases
	setIntTests := []testCase{
		{"SetInt_Simple", simpleJSON, "age", []string{"age"}, "SetInt", "", 31, false, 0, []byte(`31`)},
		{"SetInt_Nested", nestedJSON, "user.age", []string{"user", "age"}, "SetInt", "", 31, false, 0, []byte(`31`)},
		{"SetInt_Array", arrayJSON, "scores.2", []string{"scores", "[2]"}, "SetInt", "", 35, false, 0, []byte(`35`)},
	}
	benchmarks = append(benchmarks, setIntTests...)

	// SetBool operation test cases
	setBoolTests := []testCase{
		{"SetBool_Simple", simpleJSON, "isActive", []string{"isActive"}, "SetBool", "", 0, false, 0, []byte(`false`)},
		{"SetBool_Nested", nestedJSON, "preferences.notifications", []string{"preferences", "notifications"}, "SetBool", "", 0, false, 0, []byte(`false`)},
		{"SetBool_Nested2", nestedJSON, "user.isActive", []string{"user", "isActive"}, "SetBool", "", 0, false, 0, []byte(`false`)},
	}
	benchmarks = append(benchmarks, setBoolTests...)

	// SetFloat operation test cases
	setFloatTests := []testCase{
		{"SetFloat_Simple", simpleJSON, "height", []string{"height"}, "SetFloat", "", 0, false, 1.80, []byte(`1.8`)},
		{"SetFloat_Nested", nestedJSON, "user.height", []string{"user", "height"}, "SetFloat", "", 0, false, 1.80, []byte(`1.8`)},
	}
	benchmarks = append(benchmarks, setFloatTests...)

	for _, bm := range benchmarks {
		// Benchmark sjson
		b.Run("Sjson_"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)

				switch bm.op {
				case "Set", "SetString":
					_, _ = sjson.SetBytes(jsonCopy, bm.sjsonKey, bm.strVal)
				case "SetInt":
					_, _ = sjson.SetBytes(jsonCopy, bm.sjsonKey, bm.intVal)
				case "SetBool":
					_, _ = sjson.SetBytes(jsonCopy, bm.sjsonKey, bm.boolVal)
				case "SetFloat":
					_, _ = sjson.SetBytes(jsonCopy, bm.sjsonKey, bm.floatVal)
				}
			}
			b.ReportAllocs()
		})

		// Benchmark jsonparser
		b.Run("Jsonparser_"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)

				_, _ = jsonparser.Set(jsonCopy, bm.jsonparserVal, bm.jsonparserKey...)
			}
			b.ReportAllocs()
		})
	}
}

// Benchmark raw sjson vs jsonparser Delete operations
// cpu: Apple M2 Pro
// BenchmarkRawDeleteComparison
// BenchmarkRawDeleteComparison/Sjson_Delete_Simple
// BenchmarkRawDeleteComparison/Sjson_Delete_Simple-12         	 9040210	       121.0 ns/op	     192 B/op	       4 allocs/op
// BenchmarkRawDeleteComparison/Jsonparser_Delete_Simple
// BenchmarkRawDeleteComparison/Jsonparser_Delete_Simple-12    	14392932	        81.59 ns/op	     192 B/op	       3 allocs/op
// BenchmarkRawDeleteComparison/Sjson_Delete_Nested
// BenchmarkRawDeleteComparison/Sjson_Delete_Nested-12         	 1933540	       613.3 ns/op	    1488 B/op	       8 allocs/op
// BenchmarkRawDeleteComparison/Jsonparser_Delete_Nested
// BenchmarkRawDeleteComparison/Jsonparser_Delete_Nested-12    	 3353985	       358.9 ns/op	     640 B/op	       3 allocs/op
// BenchmarkRawDeleteComparison/Sjson_Delete_Array
// BenchmarkRawDeleteComparison/Sjson_Delete_Array-12          	 3591133	       338.9 ns/op	     544 B/op	       6 allocs/op
// BenchmarkRawDeleteComparison/Jsonparser_Delete_Array
// BenchmarkRawDeleteComparison/Jsonparser_Delete_Array-12     	 5487811	       219.0 ns/op	     320 B/op	       2 allocs/op
func BenchmarkRawDeleteComparison(b *testing.B) {
	type testCase struct {
		name          string
		jsonData      []byte
		sjsonKey      string
		jsonparserKey []string
	}

	// Define all test cases
	benchmarks := []testCase{
		{"Delete_Simple", simpleJSON, "name", []string{"name"}},
		{"Delete_Nested", nestedJSON, "user.address.city", []string{"user", "address", "city"}},
		{"Delete_Array", arrayJSON, "users.1", []string{"users", "[1]"}},
	}

	for _, bm := range benchmarks {
		// Benchmark sjson
		b.Run("Sjson_"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_, _ = sjson.DeleteBytes(jsonCopy, bm.sjsonKey)
			}
			b.ReportAllocs()
		})

		// Benchmark jsonparser
		b.Run("Jsonparser_"+bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// Create a copy of the JSON data for each iteration to avoid modifying the original
				jsonCopy := make([]byte, len(bm.jsonData))
				copy(jsonCopy, bm.jsonData)
				_ = jsonparser.Delete(jsonCopy, bm.jsonparserKey...)
			}
			b.ReportAllocs()
		})
	}
}
