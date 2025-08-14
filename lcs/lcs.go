package lcs

import (
	"strings"
)

// Options for LCS calculation
type Options struct {
	MaxLength     int     // Maximum character length to process (default: 150)
	CaseSensitive bool    // Whether to consider case (default: false)
	Threshold     float64 // Default similarity threshold (default: 0.7)
	WordBased     bool    // Whether to use word-based comparison (default: true)
}

// DefaultOptions returns default configuration options
func DefaultOptions() Options {
	return Options{
		MaxLength:     150,
		CaseSensitive: false,
		Threshold:     0.75,
		WordBased:     true,
	}
}

// CalculateSimilarity calculates similarity between two strings using LCS
// Returns similarity score between 0.0-1.0
// Formula: (2 * LCS_length) / (length1 + length2)
func CalculateSimilarity(str1, str2 string) float64 {
	return CalculateSimilarityWithOptions(str1, str2, DefaultOptions())
}

// CalculateSimilarityWithOptions calculates similarity with custom options
func CalculateSimilarityWithOptions(str1, str2 string, opts Options) float64 {
	str1 = strings.TrimSpace(str1)
	str2 = strings.TrimSpace(str2)

	if str1 == str2 {
		return 1.0
	}

	if len(str1) == 0 || len(str2) == 0 {
		return 0.0 // One empty sequence means no similarity
	}

	// Apply length limits first (always character-based)
	if len(str1) > opts.MaxLength {
		str1 = str1[:opts.MaxLength]
	}
	if len(str2) > opts.MaxLength {
		str2 = str2[:opts.MaxLength]
	}

	// Apply case sensitivity
	if !opts.CaseSensitive {
		str1 = strings.ToLower(str1)
		str2 = strings.ToLower(str2)
	}

	if opts.WordBased {
		// Split strings into words
		words1 := strings.Fields(str1)
		words2 := strings.Fields(str2)
		return calculateSimilarityFromSequences(words1, words2)
	} else {
		// Character-based comparison - convert strings to []rune for proper Unicode handling
		runes1 := []rune(str1)
		runes2 := []rune(str2)
		return calculateSimilarityFromSequences(runes1, runes2)
	}
}

// calculateSimilarityFromSequences calculates similarity between two sequences using LCS
func calculateSimilarityFromSequences[T comparable](seq1, seq2 []T) float64 {
	lcsLength := longestCommonSubsequence(seq1, seq2)
	return float64(2*lcsLength) / float64(len(seq1)+len(seq2))
}

// SimilarMessageExists checks if a similar message already exists in the given set
func SimilarMessageExists(target string, messages []string) bool {
	return SimilarMessageExistsWithOptions(target, messages, DefaultOptions())
}

// SimilarMessageExistsWithOptions checks if a similar message exists with custom options
func SimilarMessageExistsWithOptions(target string, messages []string, opts Options) bool {
	if len(messages) == 0 {
		return false
	}

	// Check each message and return true as soon as we find one that meets the threshold
	for _, message := range messages {
		similarity := CalculateSimilarityWithOptions(target, message, opts)
		if similarity >= opts.Threshold {
			return true
		}
	}

	return false
}

// longestCommonSubsequence finds the length of the longest common subsequence
// Uses dynamic programming approach for any comparable slice type
func longestCommonSubsequence[T comparable](seq1, seq2 []T) int {
	m, n := len(seq1), len(seq2)

	// Create DP table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// Fill DP table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if seq1[i-1] == seq2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
