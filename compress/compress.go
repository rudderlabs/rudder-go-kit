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

func NewSettings(algo, level string) (CompressionAlgorithm, CompressionLevel, error) {
	switch algo {
	case "zstd":
		switch level {
		case "fastest":
			return CompressionAlgoZstd, CompressionLevelZstdFastest, nil
		case "default":
			return CompressionAlgoZstd, CompressionLevelZstdDefault, nil
		case "better":
			return CompressionAlgoZstd, CompressionLevelZstdBetter, nil
		case "best":
			return CompressionAlgoZstd, CompressionLevelZstdBest, nil
		default:
			return 0, 0, fmt.Errorf("unknown compression level for %s: %s", algo, level)
		}
	case "zstd-cgo":
		switch level {
		case "fastest":
			return CompressionAlgoZstdCgo, CompressionLevelZstdCgoFastest, nil
		case "default":
			return CompressionAlgoZstdCgo, CompressionLevelZstdCgoDefault, nil
		case "best":
			return CompressionAlgoZstdCgo, CompressionLevelZstdCgoBest, nil
		default:
			return 0, 0, fmt.Errorf("unknown compression level for %s: %s", algo, level)
		}
	default:
		return 0, 0, fmt.Errorf("unknown compression algorithm: %s", algo)
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
	var err error
	algo, level, err = NewSettings(algo.String(), level.String())
	if err != nil {
		return nil, err
	}

	switch algo {
	case CompressionAlgoZstd:
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
		return &Compressor{
			compressorZstdCgo: &compressorZstdCgo{level: int(level)},
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
