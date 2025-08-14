# LCS (Longest Common Subsequence) Similarity Package

This package provides utilities for calculating string similarity using the Longest Common Subsequence (LCS) algorithm.

**Key Feature**: By default, this package uses **word-based comparison** rather than character-based comparison, which is more meaningful for structural comparison.

## Features

- **Word-based LCS similarity calculation** (default) - more meaningful for structural comparison
- **Character-based LCS similarity calculation** (optional) - for fine-grained comparison
- **Case-sensitive and case-insensitive** comparison modes
- **Configurable length limits** to prevent performance issues
- **Simple existence checking** - check if a similar message already exists
- **Performance optimized** with O(m*n) time complexity

## Installation

```bash
go get github.com/rudderlabs/rudder-go-kit/lcs
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/rudderlabs/rudder-go-kit/lcs"
)

func main() {
    // Basic similarity calculation
    similarity := lcs.CalculateSimilarity(
        "Event name login is not valid must be mapped to one of standard events",
        "Event name logout is not valid must be mapped to one of standard events",
    )
    fmt.Printf("Similarity: %.3f\n", similarity) // Output: Similarity: 0.929
    
    // Check if a similar message already exists
    messages := []string{
        "Event name login is not valid must be mapped to one of standard events",
        "Event name logout is not valid must be mapped to one of standard events",
        "Message type not supported",
    }
    
    newError := "Event name signup is not valid must be mapped to one of standard events"
    exists := lcs.SimilarMessageExists(newError, messages)
    fmt.Printf("Similar message exists: %t\n", exists) // Output: true
}
```

## API Reference

### Core Functions

#### `CalculateSimilarity(str1, str2 string) float64`
Calculates similarity between two strings using LCS algorithm.
Returns a value between 0.0 (no similarity) and 1.0 (identical).

#### `CalculateSimilarityWithOptions(str1, str2 string, opts Options) float64`
Calculates similarity with custom configuration options.

#### `SimilarMessageExists(target string, messages []string) bool`
Checks if a similar message already exists in the given set using default options.

#### `SimilarMessageExistsWithOptions(target string, messages []string, opts Options) bool`
Checks if a similar message exists with custom configuration options.

### Configuration Options

```go
type Options struct {
    MaxLength     int     // Maximum character length to process
    CaseSensitive bool    // Whether to consider case
    WordBased     bool    // Whether to use word-based comparison
}
```



## Algorithm Details

The similarity is calculated using the formula:
```
similarity = (2 * LCS_length) / (length1 + length2)
```

Where:
- `LCS_length` is the length of the longest common subsequence
- `length1` and `length2` are the lengths of the input (words or characters)

### Word-based vs Character-based Comparison

**Word-based (default)**: Splits strings into words using `strings.Fields()` and compares word sequences. This is more meaningful for structural comparison as it focuses on structural similarity rather than character-level differences.

**Character-based**: Compares individual characters. Useful for fine-grained similarity analysis.

**Example**:
- "Event name login is not valid" vs "Event name logout is not valid"
- Word-based: 5/6 words match = 0.91 similarity
- Character-based: 22/25 characters match = 0.88 similarity

This formula ensures that:
- Identical strings have similarity = 1.0
- Strings with no common subsequence have similarity = 0.0
- Partial matches have intermediate values

## Performance

- **Time Complexity**: O(m*n) where m and n are word counts (word-based) or string lengths (character-based)
- **Space Complexity**: O(m*n) for the dynamic programming table
- **Memory Efficient**: Configurable length limits prevent excessive memory usage
- **Optimized**: Uses dynamic programming for optimal LCS calculation
- **Word-based Advantage**: Typically faster for error messages as word count is usually much smaller than character count

## Use Cases

- **Error Message Deduplication**: Check if similar error messages already exist
- **Rate Limiting**: Limit unique error messages per time window
- **Log Analysis**: Identify patterns in log messages
- **Data Cleaning**: Find and merge similar records

## Examples

See the `lcs_test.go` file for comprehensive usage examples.

## Testing

Run the test suite:
```bash
go test ./lcs/...
```

Run benchmarks:
```bash
go test ./lcs/... -bench=. -benchmem
```

## License

This package is part of the rudder-go-kit library and follows the same license terms.
