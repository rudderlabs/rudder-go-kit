package compress

import "github.com/klauspost/compress/zstd"

type compressorZstd struct {
	encoder *zstd.Encoder
	decoder *zstd.Decoder
}

func (c *compressorZstd) Compress(src []byte) ([]byte, error) {
	return c.encoder.EncodeAll(src, nil), nil
}

func (c *compressorZstd) Decompress(src []byte) ([]byte, error) {
	return c.decoder.DecodeAll(src, nil)
}

func (c *compressorZstd) Close() error {
	c.decoder.Close()
	return c.encoder.Close()
}
