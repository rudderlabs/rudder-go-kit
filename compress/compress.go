package compress

import (
	"errors"
	"fmt"

	"github.com/klauspost/compress/zstd"
)

// ErrNotImplemented is returned when a feature is not implemented.
var ErrNotImplemented = errors.New("not implemented")

// CompressionAlgorithm is the interface that wraps the compression algorithm method.
type CompressionAlgorithm int

func (c CompressionAlgorithm) FromString(s string) (CompressionAlgorithm, error) {
	switch s {
	case "zstd":
		return CompressionAlgoZstd, nil
	default:
		return 0, fmt.Errorf("unknown compression algorithm: %s", s)
	}
}

// CompressionLevel is the interface that wraps the compression level method.
type CompressionLevel int

func (c CompressionLevel) FromString(s string) (CompressionLevel, error) {
	switch s {
	case "fastest":
		return CompressionLevelZstdFastest, nil
	case "default":
		return CompressionLevelZstdDefault, nil
	case "better":
		return CompressionLevelZstdBetter, nil
	case "best":
		return CompressionLevelZstdBest, nil
	default:
		return 0, fmt.Errorf("unknown compression level: %s", s)
	}
}

var (
	CompressionAlgoZstd = CompressionAlgorithm(0)

	CompressionLevelZstdFastest = CompressionLevel(zstd.SpeedFastest)
	CompressionLevelZstdDefault = CompressionLevel(zstd.SpeedDefault) // "pretty fast" compression
	CompressionLevelZstdBetter  = CompressionLevel(zstd.SpeedBetterCompression)
	CompressionLevelZstdBest    = CompressionLevel(zstd.SpeedBestCompression)
)

func New(algo CompressionAlgorithm, level CompressionLevel) (*Compressor, error) {
	if algo != CompressionAlgoZstd {
		return nil, ErrNotImplemented
	}

	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevel(level)))
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
