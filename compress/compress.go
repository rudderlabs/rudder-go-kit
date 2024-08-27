package compress

import (
	"fmt"

	"github.com/klauspost/compress/zstd"
)

// CompressionAlgorithm is the interface that wraps the compression algorithm method.
type CompressionAlgorithm int

func (c CompressionAlgorithm) String() string {
	switch c {
	case CompressionAlgoZstd:
		return "zstd"
	default:
		return ""
	}
}

func (c CompressionAlgorithm) isValid() bool {
	return c == CompressionAlgoZstd
}

func NewCompressionAlgorithm(s string) (CompressionAlgorithm, error) {
	switch s {
	case "zstd":
		return CompressionAlgoZstd, nil
	default:
		return 0, fmt.Errorf("unknown compression algorithm: %s", s)
	}
}

// CompressionLevel is the interface that wraps the compression level method.
type CompressionLevel int

func (c CompressionLevel) String() string {
	switch c {
	case CompressionLevelZstdFastest:
		return "fastest"
	case CompressionLevelZstdDefault:
		return "default"
	case CompressionLevelZstdBetter:
		return "better"
	case CompressionLevelZstdBest:
		return "best"
	default:
		return ""
	}
}

func (c CompressionLevel) isValid() bool {
	switch c {
	case CompressionLevelZstdFastest,
		CompressionLevelZstdDefault,
		CompressionLevelZstdBetter,
		CompressionLevelZstdBest:
		return true
	default:
		return false
	}
}

func NewCompressionLevel(s string) (CompressionLevel, error) {
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
	CompressionAlgoZstd = CompressionAlgorithm(1)

	CompressionLevelZstdFastest = CompressionLevel(zstd.SpeedFastest)
	CompressionLevelZstdDefault = CompressionLevel(zstd.SpeedDefault) // "pretty fast" compression
	CompressionLevelZstdBetter  = CompressionLevel(zstd.SpeedBetterCompression)
	CompressionLevelZstdBest    = CompressionLevel(zstd.SpeedBestCompression)
)

func New(algo CompressionAlgorithm, level CompressionLevel) (*Compressor, error) {
	if !algo.isValid() {
		return nil, fmt.Errorf("invalid compression algorithm: %d", algo)
	}
	if !level.isValid() {
		return nil, fmt.Errorf("invalid compression level: %d", level)
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

// SerializeSettings serializes the compression settings.
func SerializeSettings(algo CompressionAlgorithm, level CompressionLevel) string {
	return fmt.Sprintf("%d:%d", algo, level)
}

// DeserializeSettings deserializes the compression settings.
func DeserializeSettings(s string) (CompressionAlgorithm, CompressionLevel, error) {
	var algoInt, levelInt int
	_, err := fmt.Sscanf(s, "%d:%d", &algoInt, &levelInt)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot deserialize settings: %w", err)
	}

	algo := CompressionAlgorithm(algoInt)
	if !algo.isValid() {
		return 0, 0, fmt.Errorf("invalid compression algorithm: %d", algoInt)
	}

	level := CompressionLevel(levelInt)
	if !level.isValid() {
		return 0, 0, fmt.Errorf("invalid compression level: %d", levelInt)
	}

	return algo, level, nil
}
