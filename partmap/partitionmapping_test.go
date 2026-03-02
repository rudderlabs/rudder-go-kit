package partmap_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/partmap"
)

func TestPartitionMappingMarshalUnmarshalJSON(t *testing.T) {
	t.Run("basic marshal/unmarshal roundtrip", func(t *testing.T) {
		original := partmap.PartitionMapping{
			0:     0,  // PartitionRangeStart 0 -> NodeIndex 0
			100:   1,  // PartitionRangeStart 100 -> NodeIndex 1
			2000:  5,  // PartitionRangeStart 2000 -> NodeIndex 5
			40000: 10, // PartitionRangeStart 40000 -> NodeIndex 10
		}

		data, err := original.Marshal()
		require.NoError(t, err)
		require.NotEmpty(t, data)

		// UnmarshalToPartitionMapping back to PartitionMapping
		var unmarshaled partmap.PartitionMapping
		err = unmarshaled.Unmarshal(data)
		require.NoError(t, err)

		// Verify the mapping is preserved
		require.Equal(t, original, unmarshaled)
	})

	t.Run("empty mapping", func(t *testing.T) {
		original := partmap.PartitionMapping{}

		data, err := original.Marshal()
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var unmarshaled partmap.PartitionMapping
		err = unmarshaled.Unmarshal(data)
		require.NoError(t, err)
		require.Equal(t, original, unmarshaled)
	})

	t.Run("single entry", func(t *testing.T) {
		original := partmap.PartitionMapping{0: 0}

		data, err := original.Marshal()
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var unmarshaled partmap.PartitionMapping
		err = unmarshaled.Unmarshal(data)
		require.NoError(t, err)
		require.Equal(t, original, unmarshaled)
	})

	t.Run("large numbers", func(t *testing.T) {
		original := partmap.PartitionMapping{
			partmap.PartitionRangeStart(^uint32(0)): partmap.NodeIndex(^uint16(0)), // Max values
		}

		data, err := original.Marshal()
		require.NoError(t, err)
		require.NotEmpty(t, data)

		var unmarshaled partmap.PartitionMapping
		err = unmarshaled.Unmarshal(data)
		require.NoError(t, err)
		require.Equal(t, original, unmarshaled)
	})

	t.Run("error handling - invalid compressed data", func(t *testing.T) {
		invalidData := []byte("invalid compressed data")

		var unmarshaled partmap.PartitionMapping
		err := unmarshaled.Unmarshal(invalidData)
		require.Error(t, err)
	})

	t.Run("error handling - corrupted binary format", func(t *testing.T) {
		// Create valid compressed data with incomplete binary format
		// This would be a valid gzip header but corrupted content
		corruptedData := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0x01, 0x02, 0x03} // partial data

		var unmarshaled partmap.PartitionMapping
		err := unmarshaled.Unmarshal(corruptedData)
		require.Error(t, err)
	})

	t.Run("test unmarshalling valid bytes directly", func(t *testing.T) {
		originalByteData := []byte{
			31, 139, 8, 0, 155, 155, 172, 104, 2, 255, 29, 205, 29, 120, 130, 1, 24, 134, 209,
			175, 31, 24, 4, 65, 48, 8, 130, 96, 16, 4, 65, 16, 4, 131, 32, 8, 130, 32, 8, 6, 131, 32, 8, 130, 32, 8,
			130, 32, 8, 6, 131, 193, 32, 8, 130, 32, 8, 130, 32, 8, 6, 65, 16, 4, 131, 32, 8, 6, 131, 160, 235, 59, 47,
			220, 240, 30, 120, 130, 32, 40, 7, 46, 30, 38, 242, 20, 54, 154, 8, 27, 75, 122, 167, 232, 51, 77, 211, 12,
			205, 210, 23, 154, 163, 121, 90, 160, 69, 90, 162, 6, 34, 175, 180, 66, 171, 180, 70, 235, 180, 65, 155,
			180, 69, 223, 232, 59, 109, 211, 14, 237, 210, 30, 237, 211, 1, 29, 210, 17, 29, 211, 9, 157, 210, 15, 250,
			73, 191, 232, 55, 157, 209, 57, 93, 208, 37, 93, 209, 53, 221, 208, 45, 221, 209, 61, 253, 161, 7, 122, 164,
			39, 122, 166, 191, 244, 66, 175, 244, 70, 255, 232, 63, 189, 211, 7, 211, 147, 151, 90, 132, 1, 0, 0,
		}
		var unmarshaled partmap.PartitionMapping
		err := unmarshaled.Unmarshal(originalByteData)
		require.NoError(t, err)

		expected := partmap.PartitionMapping{
			3154116608: 2, 4227858432: 3, 469762048: 2, 2617245696: 4, 3087007744: 1, 3892314112: 3, 67108864: 1,
			335544320: 0, 872415232: 3, 2147483648: 2, 3489660928: 2, 3825205248: 2, 3959422976: 4, 1140850688: 2,
			1275068416: 4, 1476395008: 2, 0: 0, 603979776: 4, 738197504: 1, 1744830464: 1, 2281701376: 4, 2348810240: 0,
			402653184: 1, 1006632960: 0, 1207959552: 3, 1811939328: 2, 1879048192: 3, 3758096384: 1, 2415919104: 1,
			2818572288: 2, 134217728: 2, 805306368: 2, 2013265920: 0, 2080374784: 1, 2684354560: 0, 2885681152: 3,
			3556769792: 3, 4026531840: 0, 1073741824: 1, 1342177280: 0, 1543503872: 3, 2550136832: 3, 3221225472: 3,
			3422552064: 1, 536870912: 3, 1409286144: 1, 2483027968: 2, 2952790016: 4, 268435456: 4, 671088640: 0,
			939524096: 4, 1946157056: 4, 3623878656: 4, 3690987520: 0, 1677721600: 0, 3355443200: 0, 4093640704: 1,
			1610612736: 4, 2214592512: 3, 2751463424: 1, 201326592: 3, 3019898880: 0, 3288334336: 4, 4160749568: 2,
		}

		require.Equal(t, expected, unmarshaled)
	})
}

func TestPartitionMappingToPartitionIndexMapping(t *testing.T) {
	for _, numPartitions := range []uint32{1, 2, 4, 8, 16} {
		t.Run("numPartitions="+strconv.Itoa(int(numPartitions)), func(t *testing.T) {
			window := uint32((1 << 32) / int(numPartitions))
			partitionMapping := partmap.PartitionMapping{}
			for partitionIdx := range numPartitions {
				partitionRangeStart := partitionIdx * window
				partitionMapping[partmap.PartitionRangeStart(partitionRangeStart)] = partmap.NodeIndex(partitionIdx)
			}
			partitionIndexMapping, err := partitionMapping.ToPartitionIndexMapping()
			require.NoError(t, err)
			require.Len(t, partitionIndexMapping, int(numPartitions))
			for partitionIdx := range numPartitions {
				nodeIdx, exists := partitionIndexMapping[partmap.PartitionIndex(partitionIdx)]
				require.True(t, exists, "partition index %d should exist in the mapping", partitionIdx)
				require.Equal(t, partmap.NodeIndex(partitionIdx), nodeIdx, "node index for partition index %d should be %d", partitionIdx, partitionIdx)
			}
		})
	}

	t.Run("invalid ranges", func(t *testing.T) {
		// prepare a map with invalid ranges
		partitionMapping := partmap.PartitionMapping{
			0:   0,
			100: 1,
		}
		_, err := partitionMapping.ToPartitionIndexMapping()
		require.Error(t, err)
	})

	t.Run("invalid number of partitions", func(t *testing.T) {
		partitionMapping := partmap.PartitionMapping{
			0: 0,
			1: 1,
		}
		_, err := partitionMapping.ToPartitionIndexMappingWithNumPartitions(0)
		require.Error(t, err)
		_, err = partitionMapping.ToPartitionIndexMappingWithNumPartitions(1)
		require.Error(t, err)
	})
}

func TestPartitionIndexMappingToPartitionMapping(t *testing.T) {
	for _, numPartitions := range []uint32{1, 2, 4, 8, 16} {
		t.Run("numPartitions="+strconv.Itoa(int(numPartitions)), func(t *testing.T) {
			window := uint32((1 << 32) / int(numPartitions))
			partitionIndexMapping := partmap.PartitionIndexMapping{}
			for partitionIdx := range numPartitions {
				partitionIndexMapping[partmap.PartitionIndex(partitionIdx)] = partmap.NodeIndex(partitionIdx)
			}
			partitionMapping, err := partitionIndexMapping.ToPartitionMappingWithNumPartitions(numPartitions)
			require.NoError(t, err)
			require.Len(t, partitionMapping, int(numPartitions))
			for partitionIdx := range numPartitions {
				rangeStart := partmap.PartitionRangeStart(partitionIdx * window)
				nodeIdx, exists := partitionMapping[rangeStart]
				require.True(t, exists, "partition range start %d should exist in the mapping", rangeStart)
				require.Equal(t, partmap.NodeIndex(partitionIdx), nodeIdx, "node index for partition range start %d should be %d", rangeStart, partitionIdx)
			}
		})
	}

	t.Run("roundtrip with ToPartitionIndexMappingWithNumPartitions", func(t *testing.T) {
		for _, numPartitions := range []uint32{2, 4, 8, 16} {
			t.Run("numPartitions="+strconv.Itoa(int(numPartitions)), func(t *testing.T) {
				original := partmap.PartitionIndexMapping{}
				for partitionIdx := range numPartitions {
					original[partmap.PartitionIndex(partitionIdx)] = partmap.NodeIndex(partitionIdx % 3)
				}
				partitionMapping, err := original.ToPartitionMappingWithNumPartitions(numPartitions)
				require.NoError(t, err)

				result, err := partitionMapping.ToPartitionIndexMappingWithNumPartitions(numPartitions)
				require.NoError(t, err)
				require.Equal(t, original, result)
			})
		}
	})

	t.Run("invalid number of partitions", func(t *testing.T) {
		partitionIndexMapping := partmap.PartitionIndexMapping{
			0: 0,
			1: 1,
		}
		_, err := partitionIndexMapping.ToPartitionMappingWithNumPartitions(0)
		require.Error(t, err)
	})

	t.Run("numPartitions=1 with multiple entries", func(t *testing.T) {
		partitionIndexMapping := partmap.PartitionIndexMapping{
			0: 0,
			1: 1,
		}
		_, err := partitionIndexMapping.ToPartitionMappingWithNumPartitions(1)
		require.Error(t, err)
	})

	t.Run("numPartitions=1 with single entry", func(t *testing.T) {
		partitionIndexMapping := partmap.PartitionIndexMapping{
			0: 5,
		}
		partitionMapping, err := partitionIndexMapping.ToPartitionMappingWithNumPartitions(1)
		require.NoError(t, err)
		require.Len(t, partitionMapping, 1)
		require.Equal(t, partmap.NodeIndex(5), partitionMapping[0])
	})
}

func TestPartitionIdxToRangeStart(t *testing.T) {
	t.Run("valid conversions", func(t *testing.T) {
		for _, numPartitions := range []uint32{1, 2, 4, 8, 16, 64} {
			t.Run("numPartitions="+strconv.Itoa(int(numPartitions)), func(t *testing.T) {
				window := uint32((1 << 32) / int(numPartitions))
				for idx := range numPartitions {
					rangeStart, err := partmap.PartitionIndex(idx).ToPartitionRangeStart(numPartitions)
					require.NoError(t, err)
					require.Equal(t, partmap.PartitionRangeStart(idx*window), rangeStart)
				}
			})
		}
	})

	t.Run("zero numPartitions", func(t *testing.T) {
		_, err := partmap.PartitionIndex(0).ToPartitionRangeStart(0)
		require.Error(t, err)
	})

	t.Run("index out of range", func(t *testing.T) {
		_, err := partmap.PartitionIndex(4).ToPartitionRangeStart(4)
		require.Error(t, err)

		_, err = partmap.PartitionIndex(10).ToPartitionRangeStart(4)
		require.Error(t, err)
	})

	t.Run("boundary index", func(t *testing.T) {
		// last valid index
		rangeStart, err := partmap.PartitionIndex(3).ToPartitionRangeStart(4)
		require.NoError(t, err)
		expectedWindow := uint32((1 << 32) / 4)
		require.Equal(t, partmap.PartitionRangeStart(3*expectedWindow), rangeStart)
	})
}

func TestPartitionRangeStartToIdx(t *testing.T) {
	t.Run("valid conversions", func(t *testing.T) {
		for _, numPartitions := range []uint32{2, 4, 8, 16, 64} {
			t.Run("numPartitions="+strconv.Itoa(int(numPartitions)), func(t *testing.T) {
				window := uint32((1 << 32) / int(numPartitions))
				for idx := range numPartitions {
					rangeStart := partmap.PartitionRangeStart(idx * window)
					result, err := rangeStart.ToPartitionIndex(numPartitions)
					require.NoError(t, err)
					require.Equal(t, partmap.PartitionIndex(idx), result)
				}
			})
		}
	})

	t.Run("zero numPartitions", func(t *testing.T) {
		_, err := partmap.PartitionRangeStart(0).ToPartitionIndex(0)
		require.Error(t, err)
	})

	t.Run("roundtrip with PartitionIdxToRangeStart", func(t *testing.T) {
		for _, numPartitions := range []uint32{1, 2, 4, 8, 16, 64} {
			t.Run("numPartitions="+strconv.Itoa(int(numPartitions)), func(t *testing.T) {
				for idx := range numPartitions {
					originalIdx := partmap.PartitionIndex(idx)
					rangeStart, err := originalIdx.ToPartitionRangeStart(numPartitions)
					require.NoError(t, err)

					resultIdx, err := rangeStart.ToPartitionIndex(numPartitions)
					require.NoError(t, err)
					require.Equal(t, originalIdx, resultIdx)
				}
			})
		}
	})
}
