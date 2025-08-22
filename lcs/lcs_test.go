package lcs_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/lcs"
)

func TestSimilarity(t *testing.T) {
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
			expected:    0.933, // Very high similarity due to common structure (word-based)
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
			similarity := lcs.Similarity(tc.str1, tc.str2)
			require.InDelta(t, tc.expected, similarity, 0.001, "Similarity score mismatch: "+tc.description)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	// Test with very long strings
	longStr1 := string(make([]byte, 2000)) // 2000 bytes
	longStr2 := string(make([]byte, 2000)) // 2000 bytes

	// Should not panic and should handle length limits
	similarity := lcs.Similarity(longStr1, longStr2)
	require.GreaterOrEqual(t, similarity, 0.0)
	require.LessOrEqual(t, similarity, 1.0)

	// Test with special characters
	specialStr1 := "Hello\n\t\rWorld!@#$%^&*()"
	specialStr2 := "Hello\n\t\rWorld!@#$%^&*()"
	similarity = lcs.Similarity(specialStr1, specialStr2)
	require.Equal(t, 1.0, similarity)

	// Test with unicode characters
	unicodeStr1 := "Hello 世界"
	unicodeStr2 := "Hello 世界"
	similarity = lcs.Similarity(unicodeStr1, unicodeStr2)
	require.Equal(t, 1.0, similarity)
}

// TestDPTableBranches tests all branches in the DP table filling logic
func TestDPTableBranches(t *testing.T) {
	// Test case where dp[i-1][j] > dp[i][j-1]
	// This should trigger the "else if dp[i-1][j] >= dp[i][j-1]" branch
	str1 := "a b c d"
	str2 := "a x c y"
	similarity := lcs.Similarity(str1, str2)
	require.Greater(t, similarity, 0.0)
	require.Less(t, similarity, 1.0)

	// Test case where dp[i][j-1] > dp[i-1][j]
	// This should trigger the "else" branch
	str3 := "a b c d e"
	str4 := "a b x d e"
	similarity2 := lcs.Similarity(str3, str4)
	require.Greater(t, similarity2, 0.0)
	require.Less(t, similarity2, 1.0)

	// Test case where dp[i-1][j] == dp[i][j-1]
	// This should trigger the "else if dp[i-1][j] >= dp[i][j-1]" branch
	str5 := "a b c"
	str6 := "a x c"
	similarity3 := lcs.Similarity(str5, str6)
	require.Greater(t, similarity3, 0.0)
	require.Less(t, similarity3, 1.0)
}

// TestEmptyAndWhitespaceStrings tests edge cases with empty and whitespace strings
func TestEmptyAndWhitespaceStrings(t *testing.T) {
	// Test with whitespace-only strings
	whitespace1 := "   \t\n\r   "
	whitespace2 := "   \t\n\r   "
	similarity := lcs.Similarity(whitespace1, whitespace2)
	require.Equal(t, 1.0, similarity) // Both are empty after strings.Fields()

	// Test with one whitespace string and one normal string
	normal := "hello world"
	similarity2 := lcs.Similarity(whitespace1, normal)
	require.Equal(t, 0.0, similarity2) // One empty after strings.Fields()

	// Test with mixed whitespace and content
	mixed1 := "  hello  world  "
	mixed2 := "hello world"
	similarity3 := lcs.Similarity(mixed1, mixed2)
	require.Equal(t, 1.0, similarity3) // Should be identical after strings.Fields()
}
