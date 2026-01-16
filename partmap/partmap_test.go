package partmap_test

import (
	"testing"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/partmap"
)

func TestMurmur3Partition32(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		t.Run("should return valid partition index and value", func(t *testing.T) {
			key := "test-key"
			emptyKey := ""

			t.Run("0 partition", func(t *testing.T) {
				numPartitions := uint32(0)
				require.Panics(t, func() {
					partmap.Murmur3Partition32(key, numPartitions)
				})
			})

			t.Run("1 partition", func(t *testing.T) {
				numPartitions := uint32(1)
				partitionIdx, startsAt := partmap.Murmur3Partition32(key, numPartitions)
				require.EqualValues(t, 0, int(partitionIdx), "invalid partition index")
				require.EqualValues(t, 0, int(startsAt), "invalid startsAt")
			})

			t.Run("2 partitions", func(t *testing.T) {
				numPartitions := uint32(2)
				partitionIdx, startsAt := partmap.Murmur3Partition32(key, numPartitions)
				require.EqualValues(t, 1, int(partitionIdx), "invalid partition index")
				require.EqualValues(t, 2147483648, int(startsAt), "invalid startsAt")

				t.Run("empty key", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32(emptyKey, numPartitions)
					require.EqualValues(t, 0, int(partitionIdx), "partition index")
					require.EqualValues(t, 0, int(startsAt), "invalid startsAt")
				})
				t.Run("upper range", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32("1", numPartitions)
					require.EqualValues(t, 1, int(partitionIdx), "upper range key should map to highest partition")
					require.EqualValues(t, 2147483648, int(startsAt), "invalid startsAt")
				})
			})

			t.Run("16 partitions", func(t *testing.T) {
				numPartitions := uint32(16)
				partitionIdx, startsAt := partmap.Murmur3Partition32(key, numPartitions)
				require.EqualValues(t, 12, int(partitionIdx), "partition index")
				require.EqualValues(t, 3221225472, int(startsAt), "invalid startsAt")
				t.Run("empty key", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32(emptyKey, numPartitions)
					require.EqualValues(t, 0, int(partitionIdx), "partition index")
					require.EqualValues(t, 0, int(startsAt), "invalid startsAt")
				})
				t.Run("upper range", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32("12", numPartitions)
					require.EqualValues(t, 15, int(partitionIdx), "upper range key should map to highest partition")
					require.EqualValues(t, 4026531840, int(startsAt), "invalid startsAt")
				})
			})

			t.Run("64 partitions", func(t *testing.T) {
				numPartitions := uint32(64)
				partitionIdx, startsAt := partmap.Murmur3Partition32(key, numPartitions)
				require.EqualValues(t, 49, int(partitionIdx), "partition index")
				require.EqualValues(t, 3288334336, int(startsAt), "invalid startsAt")
				t.Run("empty key", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32(emptyKey, numPartitions)
					require.EqualValues(t, 0, int(partitionIdx), "partition index")
					require.EqualValues(t, 0, int(startsAt), "invalid startsAt")
				})

				t.Run("upper range", func(t *testing.T) {
					partitionIdx, startsAt := partmap.Murmur3Partition32("49", numPartitions)
					require.EqualValues(t, 63, int(partitionIdx), "upper range key should map to highest partition")
					require.EqualValues(t, 4227858432, int(startsAt), "invalid startsAt")
				})
			})
		})
	})
}

func TestPartitionMappingMarshalUnmarshalJSON(t *testing.T) {
	t.Run("basic marshal/unmarshal roundtrip", func(t *testing.T) {
		original := partmap.PartitionMapping{
			0:     0,  // PartitionRangeStart 0 -> ServerIndex 0
			100:   1,  // PartitionRangeStart 100 -> ServerIndex 1
			2000:  5,  // PartitionRangeStart 2000 -> ServerIndex 5
			40000: 10, // PartitionRangeStart 40000 -> ServerIndex 10
		}

		// Marshal to JSON (compressed bytes)
		data, err := jsonrs.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		// Unmarshal back to PartitionMapping
		var unmarshaled partmap.PartitionMapping
		err = jsonrs.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		// Verify the mapping is preserved
		require.Equal(t, original, unmarshaled)
	})

	t.Run("empty mapping", func(t *testing.T) {
		original := partmap.PartitionMapping{}

		data, err := jsonrs.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var unmarshaled partmap.PartitionMapping
		err = jsonrs.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		require.Equal(t, original, unmarshaled)
	})

	t.Run("single entry", func(t *testing.T) {
		original := partmap.PartitionMapping{0: 0}

		data, err := jsonrs.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var unmarshaled partmap.PartitionMapping
		err = jsonrs.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		require.Equal(t, original, unmarshaled)
	})

	t.Run("large numbers", func(t *testing.T) {
		original := partmap.PartitionMapping{
			partmap.PartitionRangeStart(^uint32(0)): partmap.ServerIndex(^uint16(0)), // Max values
		}

		data, err := jsonrs.Marshal(original)
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var unmarshaled partmap.PartitionMapping
		err = jsonrs.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)
		require.Equal(t, original, unmarshaled)
	})

	t.Run("error handling - invalid compressed data", func(t *testing.T) {
		invalidData := []byte("invalid compressed data")

		var pm partmap.PartitionMapping
		err := jsonrs.Unmarshal(invalidData, &pm)
		require.Error(t, err)
	})

	t.Run("error handling - corrupted binary format", func(t *testing.T) {
		// Create valid compressed data with incomplete binary format
		// This would be a valid gzip header but corrupted content
		corruptedData := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x01, 0x02, 0x03} // partial data

		var pm partmap.PartitionMapping
		err := jsonrs.Unmarshal(corruptedData, &pm)
		require.Error(t, err)
	})
}

func TestPartitionMappingClone(t *testing.T) {
	t.Run("clones non-nil mapping", func(t *testing.T) {
		original := partmap.PartitionMapping{
			0:     0,
			100:   1,
			2000:  5,
			40000: 10,
		}

		cloned := original.Clone()

		// Verify the clone is not the same pointer
		require.NotSame(t, &original, &cloned)

		// Verify the contents are equal
		require.Equal(t, original, cloned)

		// Modify original and verify clone is unchanged
		delete(original, 0)
		require.NotEqual(t, original, cloned)
		require.Len(t, original, 3)
		require.Len(t, cloned, 4)
	})

	t.Run("clones empty mapping", func(t *testing.T) {
		original := partmap.PartitionMapping{}

		cloned := original.Clone()

		require.NotSame(t, &original, &cloned)
		require.Equal(t, original, cloned)
		require.Empty(t, cloned)
	})

	t.Run("handles nil mapping", func(t *testing.T) {
		var original partmap.PartitionMapping

		cloned := original.Clone()

		require.Nil(t, cloned)
	})

	t.Run("independent modifications", func(t *testing.T) {
		original := partmap.PartitionMapping{
			100: 1,
			200: 2,
		}

		cloned := original.Clone()

		// Modify original
		original[300] = 3
		delete(original, 100)

		// Verify clone is unaffected
		require.Len(t, original, 2)
		require.Len(t, cloned, 2)
		_, exists := cloned[100]
		require.True(t, exists)
		_, exists = cloned[300]
		require.False(t, exists)

		// Modify clone
		cloned[400] = 4
		delete(cloned, 200)

		// Verify original is unaffected
		_, exists = original[200]
		require.True(t, exists)
		_, exists = original[400]
		require.False(t, exists)
	})
}
