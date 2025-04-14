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
BenchmarkTrie/Insert/1000_items_10_chars-12                 1449            765763 ns/op         1571211 B/op      23404 allocs/op
BenchmarkTrie/Insert/10000_items_10_chars-12                 140           8791232 ns/op        14210344 B/op     212759 allocs/op
BenchmarkTrie/Insert/100000_items_10_chars-12                 13          88382346 ns/op        132672852 B/op   1975541 allocs/op
BenchmarkTrie/Insert/1000_items_50_chars-12                  248           4839381 ns/op         9587178 B/op     139020 allocs/op
BenchmarkTrie/Insert/10000_items_50_chars-12                  25          46529992 ns/op        94296304 B/op    1368181 allocs/op
BenchmarkTrie/Insert/100000_items_50_chars-12                  3         468312653 ns/op        933582800 B/op  13526962 allocs/op

BenchmarkTrie/Get/10000_items_10_chars_1.0_hit_rate-12              5466            211260 ns/op           46720 B/op       2710 allocs/op
BenchmarkTrie/Get/10000_items_10_chars_0.5_hit_rate-12              7622            147444 ns/op           29072 B/op       1955 allocs/op
BenchmarkTrie/Get/10000_items_10_chars_0.0_hit_rate-12             15283             78597 ns/op           11816 B/op       1217 allocs/op
BenchmarkTrie/Get/100000_items_50_chars_1.0_hit_rate-12             1315            865524 ns/op          248000 B/op       5000 allocs/op
BenchmarkTrie/Get/100000_items_50_chars_0.5_hit_rate-12             2481            472676 ns/op          130312 B/op       3126 allocs/op
BenchmarkTrie/Get/100000_items_50_chars_0.0_hit_rate-12            10333            123480 ns/op           12568 B/op       1262 allocs/op
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
			data := generateData(tc.numItems, tc.keyLength)
			searchKeys := make([]string, 0, tc.numItems)
			for _, key := range data {
				t.Insert(key)
				searchKeys = append(searchKeys, key)
			}

			// 2. Generate search keys based on hitRatio
			numSearchOps := 1000
			keysToSearch := make([]string, numSearchOps)
			numHits := int(float64(numSearchOps) * tc.hitRatio)
			var err error
			for i := 0; i < numSearchOps; i++ {
				if i < numHits {
					// Pick a random existing key
					keysToSearch[i] = searchKeys[mathrand.Intn(len(searchKeys))]
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
						_, _ = t.Get(key)
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
