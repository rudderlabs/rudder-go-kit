package compress

import (
	"fmt"
	"time"

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

type settings struct {
	timeout        time.Duration
	panicOnTimeout bool
}

// Option is a function that configures a Compressor.
type Option func(*settings)

// WithTimeout sets the timeout for compression and decompression operations.
// If the timeout is exceeded, the operation will return an error.
// If panicOnTimeout is enabled via WithPanicOnTimeout, it will panic instead.
//
// IMPORTANT: When a timeout occurs, the underlying compression/decompression
// goroutine will leak because the underlying libraries (both klauspost/compress/zstd
// and DataDog/zstd) do not support context cancellation. The goroutine will continue
// running until the operation completes or the process terminates. This is an
// unavoidable trade-off to prevent indefinite blocking of the caller.
//
// Consider the implications of goroutine leaks in high-throughput scenarios and
// set appropriate timeout values to balance between preventing indefinite hangs
// and minimizing leaked goroutines.
func WithTimeout(timeout time.Duration) Option {
	return func(s *settings) { s.timeout = timeout }
}

// WithPanicOnTimeout configures the compressor to panic when a timeout occurs
// instead of returning an error. This should be used with WithTimeout.
func WithPanicOnTimeout() Option {
	return func(s *settings) { s.panicOnTimeout = true }
}

func New(algo CompressionAlgorithm, level CompressionLevel, opts ...Option) (*Compressor, error) {
	customSettings := &settings{
		timeout:        0, // no timeout by default
		panicOnTimeout: false,
	}
	for _, opt := range opts {
		opt(customSettings)
	}

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

		return &Compressor{
			compressorZstd: &compressorZstd{
				encoder: encoder,
				decoder: decoder,
			},
			settings: customSettings,
		}, nil
	case CompressionAlgoZstdCgo:
		return &Compressor{
			compressorZstdCgo: &compressorZstdCgo{level: int(level)},
			settings:          customSettings,
		}, nil
	default:
		return nil, fmt.Errorf("unknown compression algorithm: %d", algo)
	}
}

// Compressor provides compression and decompression operations with optional timeout support.
//
// By default, compression and decompression operations have no timeout and may block
// indefinitely if the underlying library encounters issues. Use WithTimeout option
// during construction to add timeout protection.
//
// Note: Both supported algorithms (klauspost/compress/zstd and DataDog/zstd) do not
// support context cancellation. If a timeout occurs, the goroutine performing the
// operation will leak. See WithTimeout documentation for details.
type Compressor struct {
	*compressorZstd
	*compressorZstdCgo
	*settings
}

func (c *Compressor) Compress(src []byte) ([]byte, error) {
	return c.withTimeout("compress", func() ([]byte, error) {
		if c.compressorZstdCgo != nil {
			return c.compressorZstdCgo.Compress(src)
		}
		return c.compressorZstd.Compress(src)
	})
}

func (c *Compressor) Decompress(src []byte) ([]byte, error) {
	return c.withTimeout("decompress", func() ([]byte, error) {
		if c.compressorZstdCgo != nil {
			return c.compressorZstdCgo.Decompress(src)
		}
		return c.compressorZstd.Decompress(src)
	})
}

// withTimeout wraps a compression/decompression operation with timeout protection.
//
// If no timeout is configured (c.settings.timeout == 0), the function executes directly
// without any overhead. If a timeout is configured, the operation runs in a separate
// goroutine with timeout monitoring.
//
// On timeout, this method returns an error (or panics if panicOnTimeout is enabled).
// The goroutine executing the operation will continue running in the background and
// cannot be cancelled due to lack of context support in the underlying compression
// libraries. This is a known limitation - see WithTimeout for details.
func (c *Compressor) withTimeout(operation string, fn func() ([]byte, error)) ([]byte, error) {
	// If no timeout configured, call function directly
	if c.settings.timeout == 0 {
		return fn()
	}

	type result struct {
		data []byte
		err  error
	}

	resultCh := make(chan result, 1)
	go func() {
		data, err := fn()
		resultCh <- result{data: data, err: err}
		close(resultCh)
	}()

	select {
	case r := <-resultCh:
		return r.data, r.err
	case <-time.After(c.settings.timeout):
		// Goroutine leak: the goroutine above will continue running until fn() completes.
		// This is unavoidable without context cancellation support in the underlying libraries.
		timeoutErr := fmt.Errorf("%s operation timeout after %v", operation, c.settings.timeout)
		if c.settings.panicOnTimeout {
			panic(timeoutErr)
		}
		return nil, timeoutErr
	}
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
