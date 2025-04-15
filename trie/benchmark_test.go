package trie

import (
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"testing"
)

/*
goos: darwin
goarch: arm64
pkg: github.com/rudderlabs/rudder-go-kit/trie
cpu: Apple M2 Pro
BenchmarkTrie/Insert/1000_items_10_chars-12                 1440            917506 ns/op         1563102 B/op      23287 allocs/op
BenchmarkTrie/Insert/10000_items_10_chars-12                 122           9826675 ns/op        14141638 B/op     212074 allocs/op
BenchmarkTrie/Insert/100000_items_10_chars-12                 13          91967837 ns/op        132748437 B/op   1976261 allocs/op
BenchmarkTrie/Insert/1000_items_50_chars-12                  247           5039432 ns/op         9564105 B/op     138687 allocs/op
BenchmarkTrie/Insert/10000_items_50_chars-12                  22          47210977 ns/op        94315532 B/op    1368365 allocs/op
BenchmarkTrie/Insert/100000_items_50_chars-12                  2         509287521 ns/op        933780416 B/op  13530038 allocs/op

BenchmarkTrie/Get/10000_items_10_chars_1.0_hit_rate-12              5078            220148 ns/op           46912 B/op       2716 allocs/op
BenchmarkTrie/Get/10000_items_10_chars_0.5_hit_rate-12              6721            164567 ns/op           29368 B/op       1960 allocs/op
BenchmarkTrie/Get/10000_items_10_chars_0.0_hit_rate-12             14371             86925 ns/op           11736 B/op       1200 allocs/op
BenchmarkTrie/Get/100000_items_50_chars_1.0_hit_rate-12             1318            868631 ns/op          248000 B/op       5000 allocs/op
BenchmarkTrie/Get/100000_items_50_chars_0.5_hit_rate-12             2562            475571 ns/op          130176 B/op       3120 allocs/op
BenchmarkTrie/Get/100000_items_50_chars_0.0_hit_rate-12            10000            108950 ns/op           12624 B/op       1260 allocs/op



memory profile:

Showing nodes accounting for 23638.81MB, 99.71% of 23707.34MB total
Dropped 45 nodes (cum <= 118.54MB)
Showing top 10 nodes out of 11

	flat  flat%   sum%        cum   cum%

22098.73MB 93.21% 93.21% 22098.73MB 93.21%  github.com/rudderlabs/rudder-go-kit/trie.(*Trie).Insert (inline)

	1105.05MB  4.66% 97.88%  1105.05MB  4.66%  unicode/utf8.appendRuneNonASCII
	 435.02MB  1.83% 99.71%  1540.07MB  6.50%  unicode/utf8.AppendRune (inline)
	        0     0% 99.71%  1540.07MB  6.50%  github.com/rudderlabs/rudder-go-kit/trie.(*Trie).Get
	        0     0% 99.71% 19397.43MB 81.82%  github.com/rudderlabs/rudder-go-kit/trie.BenchmarkTrie.func1.1
	        0     0% 99.71%  2744.61MB 11.58%  github.com/rudderlabs/rudder-go-kit/trie.BenchmarkTrie.func2
	        0     0% 99.71%  1540.07MB  6.50%  github.com/rudderlabs/rudder-go-kit/trie.BenchmarkTrie.func2.1
	        0     0% 99.71%  1540.07MB  6.50%  strings.(*Builder).WriteRune
	        0     0% 99.71% 19854.87MB 83.75%  testing.(*B).launch
	        0     0% 99.71%  3846.46MB 16.22%  testing.(*B).run1.func1
*/

func BenchmarkTrie(b *testing.B) {
	b.Run("Insert", func(b *testing.B) {
		testCases := []struct {
			numItems  int
			keyLength int
		}{
			{1000, 10},
			{10000, 10},
			{100000, 10}, // Potentially large memory usage
			{1000, 50},
			{10000, 50},
			{100000, 50}, // Potentially large memory usage
		}

		for _, tc := range testCases {
			name := fmt.Sprintf("%d_items_%d_chars", tc.numItems, tc.keyLength)
			dataToInsert := generateData(tc.numItems, tc.keyLength)

			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					t := NewTrie()

					for _, key := range dataToInsert {
						t.Insert(key)
					}
				}
			})
		}
	})

	b.Run("Get", func(b *testing.B) {
		searchTestCases := []struct {
			numItems  int
			keyLength int
			hitRatio  float64 // Percentage of searches for existing keys
		}{
			{10000, 10, 1.0}, // All hits
			{10000, 10, 0.5}, // 50% hits, 50% misses
			{10000, 10, 0.0}, // All misses
			{100000, 50, 1.0},
			{100000, 50, 0.5},
			{100000, 50, 0.0},
		}

		for _, tc := range searchTestCases {
			// 1. Pre-populate a trie
			t := NewTrie()
			dataToInsert := generateData(tc.numItems, tc.keyLength)
			for _, key := range dataToInsert {
				t.Insert(key)
			}

			// 2. Generate search keys based on hitRatio
			numSearchOps := 1000
			keysToSearch := make([]string, numSearchOps)
			numHits := int(float64(numSearchOps) * tc.hitRatio)
			var err error
			for i := 0; i < numSearchOps; i++ {
				if i < numHits {
					// Pick a random existing key
					keysToSearch[i] = dataToInsert[mathrand.Intn(len(dataToInsert))]
				} else {
					// Generate a random key likely not in the trie
					keysToSearch[i], err = generateRandomString(tc.keyLength)
					if err != nil {
						panic(err)
					}
				}
			}

			// 3. Run the benchmark
			name := fmt.Sprintf("%d_items_%d_chars_%.1f_hit_rate", tc.numItems, tc.keyLength, tc.hitRatio)
			b.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for _, key := range keysToSearch {
						_, _ = t.GetMatchedPrefixWord(key)
					}
				}
			})
		}
	})
}

func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return string(b), nil
}

func generateData(numItems, keyLength int) []string {
	data := make([]string, numItems)
	for i := 0; i < numItems; i++ {
		key, err := generateRandomString(keyLength)
		if err != nil {
			panic(err)
		}
		data[i] = key
	}
	return data
}
