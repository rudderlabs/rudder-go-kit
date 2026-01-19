package partmap_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/partmap"
)

func TestPartitionMappingMarshalUnmarshalJSON(t *testing.T) {
	t.Run("basic marshal/unmarshal roundtrip", func(t *testing.T) {
		original := partmap.PartitionMapping{
			0:     0,  // PartitionRangeStart 0 -> ServerIndex 0
			100:   1,  // PartitionRangeStart 100 -> ServerIndex 1
			2000:  5,  // PartitionRangeStart 2000 -> ServerIndex 5
			40000: 10, // PartitionRangeStart 40000 -> ServerIndex 10
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
			partmap.PartitionRangeStart(^uint32(0)): partmap.ServerIndex(^uint16(0)), // Max values
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
}
