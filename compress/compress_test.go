package compress

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var loremIpsumDolor = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.
Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`

func TestCompress(t *testing.T) {
	c, err := New(CompressionAlgoZstd)
	require.NoError(t, err)

	reader, err := c.Compress(strings.NewReader(loremIpsumDolor), WithCompressionLevel(CompressionLevelZstdBest))
	require.NoError(t, err)

	compressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Less(t, len(compressed), len(loremIpsumDolor))

	reader, err = c.Decompress(bytes.NewReader(compressed))
	require.NoError(t, err)

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, loremIpsumDolor, string(decompressed))
}

func TestReader(t *testing.T) {
	c, err := New(CompressionAlgoZstd)
	require.NoError(t, err)

	reader, err := c.Compress(strings.NewReader(loremIpsumDolor), WithCompressionLevel(CompressionLevelZstdBest))
	require.NoError(t, err)

	buf, err := c.Decompress(reader)
	require.NoError(t, err)

	decompressed, err := io.ReadAll(buf)
	require.NoError(t, err)
	require.Equal(t, loremIpsumDolor, string(decompressed))
}
