package lcs

import (
	"strings"
)

// Similarity calculates similarity between two sequences using LCS
func Similarity(str1, str2 string) float64 {
	if str1 == str2 {
		return 1.0
	}

	if len(str1) == 0 || len(str2) == 0 {
		return 0.0 // One empty sequence means no similarity
	}

	// Word-based comparison
	words1 := strings.Fields(str1)
	words2 := strings.Fields(str2)
	lcsLength := longestCommonSubsequence(words1, words2)

	return float64(2*lcsLength) / float64(len(words1)+len(words2))
}

// LongestCommonSubsequence finds the length of the longest common subsequence
// Uses dynamic programming approach for any comparable slice type
func longestCommonSubsequence[T comparable](seq1, seq2 []T) int {
	m, n := len(seq1), len(seq2)
	if m == 0 || n == 0 {
		return 0
	}

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
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	return dp[m][n]
}
