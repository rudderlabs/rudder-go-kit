package compress

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var loremIpsumDolor = []byte(`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`)

func TestCompress(t *testing.T) {
	compressionLevels := []CompressionLevel{
		CompressionLevelZstdFastest,
		CompressionLevelZstdDefault,
		CompressionLevelZstdBetter,
		CompressionLevelZstdBest,
	}
	for _, level := range compressionLevels {
		c, err := New(CompressionAlgoZstd, level)
		require.NoError(t, err)

		t.Cleanup(func() { _ = c.Close() })

		compressed, err := c.Compress(loremIpsumDolor)
		require.NoError(t, err)
		require.Less(t, len(compressed), len(loremIpsumDolor))

		decompressed, err := c.Decompress(compressed)
		require.NoError(t, err)
		require.Equal(t, string(loremIpsumDolor), string(decompressed))
	}
}

func TestSerialization(t *testing.T) {
	var (
		err   error
		algo  CompressionAlgorithm
		level CompressionLevel
	)

	algo, err = algo.FromString("zstd")
	require.NoError(t, err)
	require.Equal(t, CompressionAlgoZstd, algo)

	level, err = level.FromString("best")
	require.NoError(t, err)
	require.Equal(t, CompressionLevelZstdBest, level)

	serialized := SerializeSettings(algo, level)
	require.Equal(t, "1:4", serialized)

	algo, level, err = DeserializeSettings(serialized)
	require.NoError(t, err)
	require.Equal(t, CompressionAlgoZstd, algo)
	require.Equal(t, CompressionLevelZstdBest, level)
}
