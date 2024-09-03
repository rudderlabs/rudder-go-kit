package compress

import (
	"github.com/DataDog/zstd"
)

type compressorZstdCgo struct {
	level int
}

func (c *compressorZstdCgo) Compress(src []byte) ([]byte, error) {
	return zstd.CompressLevel(nil, src, c.level)
}

func (c *compressorZstdCgo) Decompress(src []byte) ([]byte, error) {
	return zstd.Decompress(nil, src)
}
