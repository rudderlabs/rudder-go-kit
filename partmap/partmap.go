package partmap

import (
	"github.com/twmb/murmur3"
)

// Murmur3Partition32 computes the partition index and beginning of the partition range in a 32-bit space for a given key and number of partitions.
//
// Please note that numPartitions must be a power of 2 (e.g. 1, 2, 4, 8, 16, ...), otherwise the distribution may not be uniform and could lead to unexpected results, including panics
//
// Implementation details:
//
//   - It uses the Murmur3 hashing algorithm to generate a 32-bit hash of the key, then calculates the partition index by dividing the hash value by the size of each partition window.
//   - It returns the partition index and the corresponding partition value (the start of the partition range).
func Murmur3Partition32(key string, numPartitions uint32) (uint32, uint32) {
	if numPartitions == 1 {
		return 0, 0 // Return zero values if numPartitions is 1 (as there's only one partition)
	}
	window := uint32((1 << 32) / int(numPartitions))    // Size of each partition window (dividing the full 32-bit range by number of partitions)
	partitionIdx := murmur3.Sum32([]byte(key)) / window // Determine partition index by dividing hash by window size
	partition := partitionIdx * window                  // Calculate the start of the partition range
	return partitionIdx, partition
}
