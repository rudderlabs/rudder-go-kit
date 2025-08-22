package lcs_test

import (
	"fmt"
	"testing"

	"github.com/rudderlabs/rudder-go-kit/lcs"
	trand "github.com/rudderlabs/rudder-go-kit/testhelper/rand"
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
			str1: trand.String(200),
			str2: trand.String(200),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lcs.Similarity(tc.str1, tc.str2)
			}
		})
	}
}

// Benchmark different string lengths to verify O(m*n) complexity
func BenchmarkComplexityScaling(b *testing.B) {
	lengths := []int{10, 50, 100, 200}

	for _, length := range lengths {
		str1 := trand.String(length)
		str2 := trand.String(length)

		b.Run(fmt.Sprintf("Length_%d", length), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				lcs.Similarity(str1, str2)
			}
		})
	}
}
