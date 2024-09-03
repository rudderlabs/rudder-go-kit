package compress

import (
	"fmt"

	zstdcgo "github.com/DataDog/zstd"
	"github.com/klauspost/compress/zstd"
)

// CompressionAlgorithm is the interface that wraps the compression algorithm method.
type CompressionAlgorithm int

func (c CompressionAlgorithm) String() string {
	switch c {
	case CompressionAlgoZstd:
		return "zstd"
	case CompressionAlgoZstdCgo:
		return "zstd-cgo"
	default:
		return ""
	}
}

func NewCompressionAlgorithm(s string) (CompressionAlgorithm, error) {
	switch s {
	case "zstd":
		return CompressionAlgoZstd, nil
	case "zstd-cgo":
		return CompressionAlgoZstdCgo, nil
	default:
		return 0, fmt.Errorf("unknown compression algorithm: %s", s)
	}
}

// CompressionLevel is the interface that wraps the compression level method.
type CompressionLevel int

func (c CompressionLevel) String() string {
	switch c {
	case CompressionLevelZstdFastest, CompressionLevelZstdCgoFastest:
		return "fastest"
	case CompressionLevelZstdDefault, CompressionLevelZstdCgoDefault:
		return "default"
	case CompressionLevelZstdBetter:
		return "better"
	case CompressionLevelZstdBest, CompressionLevelZstdCgoBest:
		return "best"
	default:
		return ""
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
	CompressionAlgoZstd         = CompressionAlgorithm(1)
	CompressionLevelZstdFastest = CompressionLevel(zstd.SpeedFastest)
	CompressionLevelZstdDefault = CompressionLevel(zstd.SpeedDefault) // "pretty fast" compression
	CompressionLevelZstdBetter  = CompressionLevel(zstd.SpeedBetterCompression)
	CompressionLevelZstdBest    = CompressionLevel(zstd.SpeedBestCompression)

	CompressionAlgoZstdCgo         = CompressionAlgorithm(2)
	CompressionLevelZstdCgoFastest = CompressionLevel(zstdcgo.BestSpeed)          // 1
	CompressionLevelZstdCgoDefault = CompressionLevel(zstdcgo.DefaultCompression) // 5
	CompressionLevelZstdCgoBest    = CompressionLevel(zstdcgo.BestCompression)    // 20
)

func New(algo CompressionAlgorithm, level CompressionLevel) (*Compressor, error) {
	switch algo {
	case CompressionAlgoZstd:
		switch level {
		case CompressionLevelZstdFastest,
			CompressionLevelZstdDefault,
			CompressionLevelZstdBetter,
			CompressionLevelZstdBest:
		default:
			return nil, fmt.Errorf("invalid compression level for %q: %d", algo, level)
		}

		encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevel(level)))
		if err != nil {
			return nil, fmt.Errorf("cannot create zstd encoder: %w", err)
		}

		decoder, err := zstd.NewReader(nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create zstd decoder: %w", err)
		}

		return &Compressor{compressorZstd: &compressorZstd{
			encoder: encoder,
			decoder: decoder,
		}}, nil
	case CompressionAlgoZstdCgo:
		var cgoLevel int
		switch level {
		case CompressionLevelZstdCgoFastest:
			cgoLevel = zstdcgo.BestSpeed
		case CompressionLevelZstdCgoDefault:
			cgoLevel = zstdcgo.DefaultCompression
		case CompressionLevelZstdCgoBest:
			cgoLevel = zstdcgo.BestCompression
		default:
			return nil, fmt.Errorf("invalid compression level for %q: %d", algo, level)
		}

		return &Compressor{
			compressorZstdCgo: &compressorZstdCgo{level: cgoLevel},
		}, nil
	default:
		return nil, fmt.Errorf("unknown compression algorithm: %d", algo)
	}
}

type Compressor struct {
	*compressorZstd
	*compressorZstdCgo
}

func (c *Compressor) Compress(src []byte) ([]byte, error) {
	if c.compressorZstdCgo != nil {
		return c.compressorZstdCgo.Compress(src)
	}
	return c.compressorZstd.Compress(src)
}

func (c *Compressor) Decompress(src []byte) ([]byte, error) {
	if c.compressorZstdCgo != nil {
		return c.compressorZstdCgo.Decompress(src)
	}
	return c.compressorZstd.Decompress(src)
}

func (c *Compressor) Close() error {
	if c.compressorZstdCgo != nil {
		return nil
	}
	return c.compressorZstd.Close()
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
	level := CompressionLevel(levelInt)
	switch algo {
	case CompressionAlgoZstd:
		switch level {
		case CompressionLevelZstdFastest,
			CompressionLevelZstdDefault,
			CompressionLevelZstdBetter,
			CompressionLevelZstdBest:
		default:
			return 0, 0, fmt.Errorf("invalid compression level for %q: %d", algo, level)
		}
	case CompressionAlgoZstdCgo:
		switch level {
		case CompressionLevelZstdCgoFastest,
			CompressionLevelZstdCgoDefault,
			CompressionLevelZstdCgoBest:
		default:
			return 0, 0, fmt.Errorf("invalid compression level for %q: %d", algo, level)
		}
	default:
		return 0, 0, fmt.Errorf("invalid compression algorithm: %d", algoInt)
	}

	return algo, level, nil
}
