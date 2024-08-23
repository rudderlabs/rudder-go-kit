package compress

import (
	"bytes"
	"errors"
	"io"

	"github.com/klauspost/compress/zstd"
)

var ErrNotImplemented = errors.New("not implemented")

type CompressionAlgorithm interface {
	compressionAlgorithm() int
}

var CompressionAlgoZstd CompressionAlgorithm = compressionAlgorithm(0)

type CompressionLevel interface {
	compressionLevel() int
}

var (
	CompressionLevelZstdFastest CompressionLevel = compressionLevel(zstd.SpeedFastest)
	CompressionLevelZstdDefault CompressionLevel = compressionLevel(zstd.SpeedDefault) // "pretty fast" compression
	CompressionLevelZstdBetter  CompressionLevel = compressionLevel(zstd.SpeedBetterCompression)
	CompressionLevelZstdBest    CompressionLevel = compressionLevel(zstd.SpeedBestCompression)
)

func New(algo CompressionAlgorithm) (*Compressor, error) {
	if algo != CompressionAlgoZstd {
		return nil, ErrNotImplemented
	}
	return &Compressor{}, nil
}

type config struct {
	compressionLevel CompressionLevel
}

type Option func(*config)

func WithCompressionLevel(level CompressionLevel) Option {
	return func(o *config) {
		o.compressionLevel = level
	}
}

type Compressor struct{}

func (c *Compressor) Compress(reader io.Reader, opts ...Option) (io.Reader, error) {
	options := config{compressionLevel: CompressionLevelZstdDefault}
	for _, o := range opts {
		o(&options)
	}

	var buf bytes.Buffer
	encoder, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.EncoderLevel(options.compressionLevel.compressionLevel())))
	if err != nil {
		return nil, err
	}

	_, err = encoder.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	if err := encoder.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (c *Compressor) Decompress(reader io.Reader) (io.Reader, error) {
	decoder, err := zstd.NewReader(reader)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = decoder.WriteTo(&buf)
	if err != nil {
		return nil, err
	}

	decoder.Close()

	return &buf, nil
}

type compressionAlgorithm int

func (c compressionAlgorithm) compressionAlgorithm() int { return int(c) }

type compressionLevel int

func (c compressionLevel) compressionLevel() int { return int(c) }
