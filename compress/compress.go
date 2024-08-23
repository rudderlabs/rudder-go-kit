package compress

import (
	"errors"
	"fmt"

	"github.com/klauspost/compress/zstd"
)

// ErrNotImplemented is returned when a feature is not implemented.
var ErrNotImplemented = errors.New("not implemented")

// CompressionAlgorithm is the interface that wraps the compression algorithm method.
type CompressionAlgorithm interface {
	compressionAlgorithm() int
}

var CompressionAlgoZstd CompressionAlgorithm = compressionAlgorithm(0)

// CompressionLevel is the interface that wraps the compression level method.
type CompressionLevel interface {
	compressionLevel() int
}

var (
	CompressionLevelZstdFastest CompressionLevel = compressionLevel(zstd.SpeedFastest)
	CompressionLevelZstdDefault CompressionLevel = compressionLevel(zstd.SpeedDefault) // "pretty fast" compression
	CompressionLevelZstdBetter  CompressionLevel = compressionLevel(zstd.SpeedBetterCompression)
	CompressionLevelZstdBest    CompressionLevel = compressionLevel(zstd.SpeedBestCompression)
)

func New(algo CompressionAlgorithm, level CompressionLevel) (*Compressor, error) {
	if algo != CompressionAlgoZstd {
		return nil, ErrNotImplemented
	}

	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevel(level.compressionLevel())))
	if err != nil {
		return nil, fmt.Errorf("cannot create zstd encoder: %w", err)
	}

	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create zstd decoder: %w", err)
	}

	return &Compressor{
		encoder: encoder,
		decoder: decoder,
	}, nil
}

type Compressor struct {
	encoder *zstd.Encoder
	decoder *zstd.Decoder
}

func (c *Compressor) Compress(src []byte) ([]byte, error) {
	return c.encoder.EncodeAll(src, nil), nil
}

func (c *Compressor) Decompress(src []byte) ([]byte, error) {
	return c.decoder.DecodeAll(src, nil)
}

func (c *Compressor) Close() error {
	c.decoder.Close()
	return c.encoder.Close()
}

type compressionAlgorithm int

func (c compressionAlgorithm) compressionAlgorithm() int { return int(c) }

type compressionLevel int

func (c compressionLevel) compressionLevel() int { return int(c) }
