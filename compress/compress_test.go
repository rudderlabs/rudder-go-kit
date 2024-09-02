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
	type testCase struct {
		algo  CompressionAlgorithm
		level CompressionLevel
	}
	testCases := []testCase{
		{CompressionAlgoZstd, CompressionLevelZstdFastest},
		{CompressionAlgoZstd, CompressionLevelZstdDefault},
		{CompressionAlgoZstd, CompressionLevelZstdBetter},
		{CompressionAlgoZstd, CompressionLevelZstdBest},

		{CompressionAlgoZstdCgo, CompressionLevelZstdCgoFastest},
		{CompressionAlgoZstdCgo, CompressionLevelZstdCgoDefault},
		{CompressionAlgoZstdCgo, CompressionLevelZstdCgoBest},
	}

	for _, tc := range testCases {
		t.Run(tc.algo.String()+"-"+tc.level.String(), func(t *testing.T) {
			c, err := New(tc.algo, tc.level)
			require.NoError(t, err)

			t.Cleanup(func() { _ = c.Close() })

			compressed, err := c.Compress(loremIpsumDolor)
			require.NoError(t, err)
			require.Less(t, len(compressed), len(loremIpsumDolor))

			decompressed, err := c.Decompress(compressed)
			require.NoError(t, err)
			require.Equal(t, string(loremIpsumDolor), string(decompressed))
		})
	}
}

func TestSerialization(t *testing.T) {
	algo, err := NewCompressionAlgorithm("zstd")
	require.NoError(t, err)
	require.Equal(t, CompressionAlgoZstd, algo)

	level, err := NewCompressionLevel("best")
	require.NoError(t, err)
	require.Equal(t, CompressionLevelZstdBest, level)

	serialized := SerializeSettings(algo, level)
	require.Equal(t, "1:4", serialized)

	algo, level, err = DeserializeSettings(serialized)
	require.NoError(t, err)
	require.Equal(t, CompressionAlgoZstd, algo)
	require.Equal(t, CompressionLevelZstdBest, level)
}

func TestDeserializationError(t *testing.T) {
	// valid algo is 1, 2
	// valid level is 1-4 for algo 1
	// valid level is 1, 5, 20 for algo 2
	testCases := []string{
		"0:0", "0:1", "0:2", "0:3", "0:4", "0:5", "0:20",

		"1:0", "1:5", "1:20",

		"2:0", "2:2", "2:3", "2:4", "2:6", "2:7", "2:8", "2:9", "2:10", "2:11",
		"2:12", "2:13", "2:14", "2:15", "2:16", "2:17", "2:18", "2:19", "2:21",
	}
	for _, tc := range testCases {
		_, _, err := DeserializeSettings(tc)
		require.Error(t, err)
	}
}

func TestNewError(t *testing.T) {
	c, err := New(CompressionAlgorithm(0), CompressionLevelZstdDefault)
	require.Nil(t, c)
	require.Error(t, err)

	c, err = New(CompressionAlgoZstd, CompressionLevel(0))
	require.Nil(t, c)
	require.Error(t, err)
}
