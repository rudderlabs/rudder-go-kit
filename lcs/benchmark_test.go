package lcs_test

import (
	"fmt"
	"testing"

	"github.com/rudderlabs/rudder-go-kit/lcs"
)

func BenchmarkCalculateSimilarity(b *testing.B) {
	testCases := []struct {
		name string
		str1 string
		str2 string
	}{
		{
			name: "Short strings",
			str1: "hello world",
			str2: "hello there",
		},
		{
			name: "Medium strings",
			str1: "Event name login is not valid must be mapped to one of standard events",
			str2: "Event name logout is not valid must be mapped to one of standard events",
		},
		{
			name: "Long strings",
			str1: "Validation Server Response Handler Validation Error for of field path events params NAME_INVALID Event param string_value hometogo has invalid name utm_campaign last touch Only alphanumeric characters and underscores are allowed",
			str2: "Validation Server Response Handler Validation Error for of field path events params NAME_INVALID Event param string_value housinganywhere has invalid name utm_term last touch Only alphanumeric characters and underscores are allowed",
		},
		{
			name: "Very long strings",
			str1: string(make([]byte, 200)),
			str2: string(make([]byte, 200)),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lcs.CalculateSimilarity(tc.str1, tc.str2)
			}
		})
	}
}

func BenchmarkSimilarMessageExists(b *testing.B) {
	messages := []string{
		"Event name login is not valid must be mapped to one of standard events",
		"Event name logout is not valid must be mapped to one of standard events",
		"Event name signup is not valid must be mapped to one of standard events",
		"Message type not supported",
		"Validation error occurred",
		"Another validation error",
		"userId error bcdf in invalidUserIds failedUpdates notFoundUserIds",
		"userId error in invalidUserIds invalidUserIds failedUpdates notFoundUserIds",
	}

	target := "Event name signup is not valid must be mapped to one of standard events"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lcs.SimilarMessageExists(target, messages)
	}
}

func BenchmarkCalculateSimilarityWithOptions(b *testing.B) {
	str1 := "Event name login is not valid must be mapped to one of standard events"
	str2 := "Event name logout is not valid must be mapped to one of standard events"

	opts := lcs.Options{
		MaxLength:     200,
		CaseSensitive: false,
		Threshold:     0.7,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lcs.CalculateSimilarityWithOptions(str1, str2, opts)
	}
}

// Benchmark different string lengths to verify O(m*n) complexity
func BenchmarkComplexityScaling(b *testing.B) {
	lengths := []int{10, 50, 100, 200}

	for _, length := range lengths {
		str1 := string(make([]byte, length))
		str2 := string(make([]byte, length))

		b.Run(fmt.Sprintf("Length_%d", length), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lcs.CalculateSimilarity(str1, str2)
			}
		})
	}
}
