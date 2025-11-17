package partmap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rudderlabs/rudder-go-kit/partmap"
)

func TestMurmur3Partition32(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		t.Run("should return valid partition index and value", func(t *testing.T) {
			key := "test-key"
			emptyKey := ""

			t.Run("2 partitions", func(t *testing.T) {
				numPartitions := uint32(2)
				partitionIdx, startsAt := partmap.Murmur3Partition32(key, numPartitions)
				assert.EqualValues(t, 1, int(partitionIdx), "invalid partition index")
				assert.EqualValues(t, 2147483648, int(startsAt), "invalid startsAt")

				t.Run("empty key", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32(emptyKey, numPartitions)
					assert.EqualValues(t, 0, int(partitionIdx), "partition index")
					assert.EqualValues(t, 0, int(startsAt), "invalid startsAt")
				})
				t.Run("upper range", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32("1", numPartitions)
					assert.EqualValues(t, 1, int(partitionIdx), "upper range key should map to highest partition")
					assert.EqualValues(t, 2147483648, int(startsAt), "invalid startsAt")
				})
			})

			t.Run("16 partitions", func(t *testing.T) {
				numPartitions := uint32(16)
				partitionIdx, startsAt := partmap.Murmur3Partition32(key, numPartitions)
				assert.EqualValues(t, 12, int(partitionIdx), "partition index")
				assert.EqualValues(t, 3221225472, int(startsAt), "invalid startsAt")
				t.Run("empty key", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32(emptyKey, numPartitions)
					assert.EqualValues(t, 0, int(partitionIdx), "partition index")
					assert.EqualValues(t, 0, int(startsAt), "invalid startsAt")
				})
				t.Run("upper range", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32("12", numPartitions)
					assert.EqualValues(t, 15, int(partitionIdx), "upper range key should map to highest partition")
					assert.EqualValues(t, 4026531840, int(startsAt), "invalid startsAt")
				})
			})

			t.Run("64 partitions", func(t *testing.T) {
				numPartitions := uint32(64)
				partitionIdx, startsAt := partmap.Murmur3Partition32(key, numPartitions)
				assert.EqualValues(t, 49, int(partitionIdx), "partition index")
				assert.EqualValues(t, 3288334336, int(startsAt), "invalid startsAt")
				t.Run("empty key", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32(emptyKey, numPartitions)
					assert.EqualValues(t, 0, int(partitionIdx), "partition index")
					assert.EqualValues(t, 0, int(startsAt), "invalid startsAt")
				})

				t.Run("upper range", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32("49", numPartitions)
					assert.EqualValues(t, 63, int(partitionIdx), "upper range key should map to highest partition")
					assert.EqualValues(t, 4227858432, int(startsAt), "invalid startsAt")
				})
			})
		})
	})
}
