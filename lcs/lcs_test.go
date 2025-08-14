package lcs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/lcs"
)

func TestCalculateSimilarity(t *testing.T) {
	testCases := []struct {
		name        string
		str1        string
		str2        string
		expected    float64
		description string
	}{
		{
			name:        "Exact match",
			str1:        "hello world",
			str2:        "hello world",
			expected:    1.0,
			description: "Identical strings should have 100% similarity",
		},
		{
			name:        "Case insensitive exact match",
			str1:        "Hello World",
			str2:        "hello world",
			expected:    1.0,
			description: "Case-insensitive identical strings should have 100% similarity",
		},
		{
			name:        "Empty strings",
			str1:        "",
			str2:        "",
			expected:    1.0,
			description: "Empty strings should have 100% similarity (both identical)",
		},
		{
			name:        "One empty string",
			str1:        "hello",
			str2:        "",
			expected:    0.0,
			description: "One empty string should have 0% similarity",
		},
		{
			name:        "No common subsequence",
			str1:        "abc",
			str2:        "xyz",
			expected:    0.0,
			description: "No common characters should have 0% similarity",
		},
		{
			name:        "Partial match",
			str1:        "hello world",
			str2:        "hello there",
			expected:    0.5, // (2 * 1) / (2 + 2) = 2/4 = 0.5 (word-based)
			description: "Partial match should have intermediate similarity",
		},
		{
			name:        "Subsequence match",
			str1:        "programming",
			str2:        "grammar",
			expected:    0.0, // (2 * 0) / (1 + 1) = 0/2 = 0.0 (word-based, no common words)
			description: "Subsequence match should calculate correct similarity",
		},
		{
			name:        "TIKTOK_ADS similar events",
			str1:        "Event name login is not valid must be mapped to one of standard events",
			str2:        "Event name logout is not valid must be mapped to one of standard events",
			expected:    0.929, // Very high similarity due to common structure (word-based)
			description: "Similar TIKTOK_ADS error messages should have high similarity",
		},
		{
			name:        "GA4 validation errors",
			str1:        "Validation Server Response Handler Validation Error for of field path events params NAME_INVALID Event param string_value hometogo has invalid name utm_campaign last touch Only alphanumeric characters and underscores are allowed",
			str2:        "Validation Server Response Handler Validation Error for of field path events params NAME_INVALID Event param string_value housinganywhere has invalid name utm_term last touch Only alphanumeric characters and underscores are allowed",
			expected:    0.9, // Very high similarity due to common structure (word-based, truncated to 150 chars)
			description: "Similar GA4 validation error messages should have high similarity",
		},
		{
			name:        "ITERABLE userId errors",
			str1:        "userId error bcdf in invalidUserIds failedUpdates notFoundUserIds",
			str2:        "userId error in invalidUserIds invalidUserIds failedUpdates notFoundUserIds failedUpdates notFoundUserIds",
			expected:    0.75, // Moderate similarity due to common patterns (word-based)
			description: "Similar ITERABLE userId error messages should have moderate similarity",
		},
		{
			name:        "Different error types",
			str1:        "Event name login is not valid must be mapped to one of standard events",
			str2:        "Message type not supported",
			expected:    0.111, // Low similarity due to few common words (word-based)
			description: "Different error types should have low similarity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lcs.CalculateSimilarity(tc.str1, tc.str2)
			require.InDelta(t, tc.expected, result, 0.001, tc.description)
		})
	}
}

func TestCalculateSimilarityWithOptions(t *testing.T) {
	testCases := []struct {
		name     string
		str1     string
		str2     string
		opts     lcs.Options
		expected float64
	}{
		{
			name: "Case sensitive",
			str1: "Hello World",
			str2: "hello world",
			opts: lcs.Options{
				CaseSensitive: true,
				MaxLength:     1000,
				Threshold:     0.7,
			},
			expected: 0.818, // Different case, but still some similarity
		},
		{
			name: "Case insensitive",
			str1: "Hello World",
			str2: "hello world",
			opts: lcs.Options{
				CaseSensitive: false,
				MaxLength:     1000,
				Threshold:     0.7,
			},
			expected: 1.0, // Same after case conversion
		},
		{
			name: "Length limit applied",
			str1: "very long string that exceeds the limit",
			str2: "very long string that exceeds the limit",
			opts: lcs.Options{
				CaseSensitive: false,
				MaxLength:     10,
				Threshold:     0.7,
			},
			expected: 1.0, // Truncated to same string
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := lcs.CalculateSimilarityWithOptions(tc.str1, tc.str2, tc.opts)
			require.InDelta(t, tc.expected, result, 0.001)
		})
	}
}

func TestSimilarMessageExists(t *testing.T) {
	messages := []string{
		"Event name login is not valid must be mapped to one of standard events",
		"Event name logout is not valid must be mapped to one of standard events",
		"Message type not supported",
		"Validation error occurred",
	}

	target := "Event name login is not valid must be mapped to one of standard events"

	// Should find exact match (using default threshold of 0.75)
	exists := lcs.SimilarMessageExists(target, messages)
	require.True(t, exists, "Should find exact match")

	// Should find similar message (using default threshold of 0.75)
	similarTarget := "Event name logout is not valid must be mapped to one of standard events"
	exists = lcs.SimilarMessageExists(similarTarget, messages)
	require.True(t, exists, "Should find similar message")

	// Should not find completely different message (using default threshold of 0.75)
	differentTarget := "Completely different error message"
	exists = lcs.SimilarMessageExists(differentTarget, messages)
	require.False(t, exists, "Should not find completely different message")
}

func TestSimilarMessageExistsWithOptions(t *testing.T) {
	messages := []string{
		"Hello World",
		"hello world",
		"HELLO WORLD",
		"Goodbye World",
	}

	opts := lcs.Options{
		CaseSensitive: true,
		MaxLength:     1000,
		Threshold:     0.8,
	}

	target := "Hello World"

	// Should find exact case match
	exists := lcs.SimilarMessageExistsWithOptions(target, messages, opts)
	require.True(t, exists, "Should find exact case match")

	// Should find case-insensitive match with case-sensitive option and high threshold (actual similarity is 0.818)
	exists = lcs.SimilarMessageExistsWithOptions("hello world", messages, opts)
	require.True(t, exists, "Should find case-insensitive match with case-sensitive option and high threshold")

	// Should find exact match with case-sensitive option and very high threshold (similarity is 1.0 >= 0.9)
	opts.Threshold = 0.9
	exists = lcs.SimilarMessageExistsWithOptions("hello world", messages, opts)
	require.True(t, exists, "Should find exact match with case-sensitive option and very high threshold")
}

func TestSimilarMessageExistsEmptyList(t *testing.T) {
	exists := lcs.SimilarMessageExists("test", []string{})
	require.False(t, exists, "Should return false for empty message list")
}

func TestSimilarMessageExistsEdgeCases(t *testing.T) {
	messages := []string{
		"Event name login is not valid",
		"Event name logout is not valid",
	}

	// Test with exact same message (using default threshold of 0.75)
	exists := lcs.SimilarMessageExists(messages[0], messages)
	require.True(t, exists, "Should find exact same message")

	// Test with very low threshold (should find any message with similarity > 0.1)
	opts := lcs.DefaultOptions()
	opts.Threshold = 0.1
	exists = lcs.SimilarMessageExistsWithOptions("Event name signup is not valid", messages, opts)
	require.True(t, exists, "Should find similar message with very low threshold")

	// Test with very high threshold (should not find similar but not identical)
	opts = lcs.DefaultOptions()
	opts.Threshold = 0.99
	exists = lcs.SimilarMessageExistsWithOptions("Event name signup is not valid", messages, opts)
	require.False(t, exists, "Should not find similar message with very high threshold")

	// Test with moderate threshold (should find similar message)
	opts = lcs.DefaultOptions()
	opts.Threshold = 0.7
	exists = lcs.SimilarMessageExistsWithOptions("Event name signup is not valid", messages, opts)
	require.True(t, exists, "Should find similar message with moderate threshold")
}

func TestDefaultOptions(t *testing.T) {
	opts := lcs.DefaultOptions()

	require.Equal(t, 150, opts.MaxLength)
	require.False(t, opts.CaseSensitive)
	require.Equal(t, 0.75, opts.Threshold)
	require.True(t, opts.WordBased)
}

func TestEdgeCases(t *testing.T) {
	// Test with very long strings
	longStr1 := string(make([]byte, 2000)) // 2000 bytes
	longStr2 := string(make([]byte, 2000)) // 2000 bytes

	// Should not panic and should handle length limits
	similarity := lcs.CalculateSimilarity(longStr1, longStr2)
	require.GreaterOrEqual(t, similarity, 0.0)
	require.LessOrEqual(t, similarity, 1.0)

	// Test with special characters
	specialStr1 := "Hello\n\t\rWorld!@#$%^&*()"
	specialStr2 := "Hello\n\t\rWorld!@#$%^&*()"
	similarity = lcs.CalculateSimilarity(specialStr1, specialStr2)
	require.Equal(t, 1.0, similarity)

	// Test with unicode characters
	unicodeStr1 := "Hello 世界"
	unicodeStr2 := "Hello 世界"
	similarity = lcs.CalculateSimilarity(unicodeStr1, unicodeStr2)
	require.Equal(t, 1.0, similarity)
}

func TestWordBasedVsCharacterBased(t *testing.T) {
	str1 := "hello world"
	str2 := "hello there"

	// Word-based comparison (default)
	wordSimilarity := lcs.CalculateSimilarity(str1, str2)
	require.Equal(t, 0.5, wordSimilarity) // 1 common word out of 2 words each

	// Character-based comparison
	opts := lcs.Options{WordBased: false, MaxLength: 1000}
	charSimilarity := lcs.CalculateSimilarityWithOptions(str1, str2, opts)
	require.Greater(t, charSimilarity, wordSimilarity) // Character-based should be higher

	// Test with longer strings
	longStr1 := "Event name login is not valid must be mapped to one of standard events"
	longStr2 := "Event name logout is not valid must be mapped to one of standard events"

	wordSimilarityLong := lcs.CalculateSimilarity(longStr1, longStr2)
	charSimilarityLong := lcs.CalculateSimilarityWithOptions(longStr1, longStr2, opts)

	require.Greater(t, charSimilarityLong, wordSimilarityLong) // Character-based should be higher for longer strings
}
