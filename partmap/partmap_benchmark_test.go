package partmap

import (
	"fmt"
	"math"
	"testing"

	"github.com/spaolacci/murmur3"

	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"
)

func BenchmarkMurmur3Partition32(b *testing.B) {
	partitionCounts := []uint32{2, 8, 16, 64, 256}
	keyLengths := []int{5, 20, 50}

	for _, numPartitions := range partitionCounts {
		for _, keyLength := range keyLengths {
			b.Run(fmt.Sprintf("Current/partitions_%d/keylen_%d", numPartitions, keyLength), func(b *testing.B) {
				// Pre-generate keys to avoid allocation overhead during benchmark
				keys := make([]string, b.N)
				for i := 0; i < b.N; i++ {
					keys[i] = rand.String(keyLength)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					Murmur3Partition32(keys[i], numPartitions)
				}
			})

			b.Run(fmt.Sprintf("Legacy/partitions_%d/keylen_%d", numPartitions, keyLength), func(b *testing.B) {
				// Pre-generate keys to avoid allocation overhead during benchmark
				keys := make([]string, b.N)
				for i := 0; i < b.N; i++ {
					keys[i] = rand.String(keyLength)
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					legacyMurmur3Partition32(keys[i], numPartitions)
				}
			})
		}
	}
}

func FuzzMurmur3Partition32(f *testing.F) {
	// Add seed corpus with various test cases
	f.Add("test-key", uint32(2))
	f.Add("", uint32(4))
	f.Add("1", uint32(8))
	f.Add("12", uint32(16))
	f.Add("49", uint32(64))
	f.Add("test-key-for-benchmarking", uint32(128))
	f.Add("very-long-key-with-many-characters-to-test-edge-cases", uint32(256))
	f.Add("short", uint32(32))

	f.Fuzz(func(t *testing.T, key string, numPartitions uint32) {
		// Skip invalid partition counts (must be > 0 and preferably power of 2)
		if numPartitions == 0 {
			t.Skip("skipping zero partitions")
		}
		if numPartitions == 1 || (numPartitions&(numPartitions-1)) != 0 {
			t.Skip("skipping non-power-of-two partitions")
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic occurred for key=%q, partitions=%d: %v", key, numPartitions, r)
			}
		}()
		// Get results from both implementations
		currentIdx, currentPartition := Murmur3Partition32(key, numPartitions)
		legacyIdx, legacyPartition := legacyMurmur3Partition32(key, numPartitions)

		// Compare partition indices
		if currentIdx != legacyIdx {
			t.Errorf("Partition index mismatch for key=%q, partitions=%d: current=%d, legacy=%d",
				key, numPartitions, currentIdx, legacyIdx)
		}

		// Compare partition values
		if currentPartition != legacyPartition {
			t.Errorf("Partition value mismatch for key=%q, partitions=%d: current=%d, legacy=%d",
				key, numPartitions, currentPartition, legacyPartition)
		}

		// Validate that partition index is within bounds
		if currentIdx >= numPartitions {
			t.Errorf("Partition index %d is out of bounds for %d partitions", currentIdx, numPartitions)
		}
	})
}

func legacyMurmur3Partition32(key string, numPartitions uint32) (uint32, uint32) {
	hash := murmur3.New32()
	if _, err := hash.Write([]byte(key)); err != nil {
		panic(err)
	}
	hashValue := hash.Sum32()
	partitionRange := (1 << 32) / int(numPartitions)

	partitionIdx := math.Floor(float64(hashValue) / float64(partitionRange))
	partition := int(math.Floor(partitionIdx * float64(partitionRange)))

	return uint32(partitionIdx), uint32(partition)
}
