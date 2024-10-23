package compress

import (
	"testing"
)

/*
BenchmarkNew/zstd
BenchmarkNew/zstd-24         	   46011	     21925 ns/op	   28451 B/op	       2 allocs/op
BenchmarkNew/zstd-cgo
BenchmarkNew/zstd-cgo-24     	   47054	     25502 ns/op	     960 B/op	       2 allocs/op
*/
func BenchmarkNew(b *testing.B) {
	b.Run("zstd", func(b *testing.B) {
		b.ReportAllocs()
		c, _ := New(CompressionAlgoZstd, CompressionLevelZstdBest)
		defer func() { _ = c.Close() }()

		for i := 0; i < b.N; i++ {
			r, _ := c.Compress(loremIpsumDolor)
			_, _ = c.Decompress(r)
		}
	})
	b.Run("zstd-cgo", func(b *testing.B) {
		b.ReportAllocs()
		c, _ := New(CompressionAlgoZstdCgo, CompressionLevelZstdCgoBest)
		defer func() { _ = c.Close() }()

		for i := 0; i < b.N; i++ {
			r, _ := c.Compress(loremIpsumDolor)
			_, _ = c.Decompress(r)
		}
	})
}

/*
BenchmarkCompress/zstd-cgo/fastest-24 	  330330	      3634 ns/op	     512 B/op	       1 allocs/op
BenchmarkCompress/zstd-cgo/default-24 	  205960	      5673 ns/op	     512 B/op	       1 allocs/op
BenchmarkCompress/zstd-cgo/best-24    	   51088	     23294 ns/op	     512 B/op	       1 allocs/op
*/
func BenchmarkCompress(b *testing.B) {
	b.Run("zstd-cgo", func(b *testing.B) {
		b.Run("fastest", func(b *testing.B) {
			b.ReportAllocs()
			c, _ := New(CompressionAlgoZstdCgo, CompressionLevelZstdCgoFastest)
			defer func() { _ = c.Close() }()

			for i := 0; i < b.N; i++ {
				_, _ = c.Compress(loremIpsumDolor)
			}
		})
		b.Run("default", func(b *testing.B) {
			b.ReportAllocs()
			c, _ := New(CompressionAlgoZstdCgo, CompressionLevelZstdCgoDefault)
			defer func() { _ = c.Close() }()

			for i := 0; i < b.N; i++ {
				_, _ = c.Compress(loremIpsumDolor)
			}
		})
		b.Run("best", func(b *testing.B) {
			b.ReportAllocs()
			c, _ := New(CompressionAlgoZstdCgo, CompressionLevelZstdCgoBest)
			defer func() { _ = c.Close() }()

			for i := 0; i < b.N; i++ {
				_, _ = c.Compress(loremIpsumDolor)
			}
		})
	})
}

/*
Benchmark to prove that no matter the compression level, the decompression always takes the same time.

BenchmarkDecompress/zstd-cgo/fastest-24 	  562878	      2031 ns/op	     448 B/op	       1 allocs/op
BenchmarkDecompress/zstd-cgo/default-24 	  506319	      2210 ns/op	     448 B/op	       1 allocs/op
BenchmarkDecompress/zstd-cgo/best-24    	  526634	      2198 ns/op	     448 B/op	       1 allocs/op
*/
func BenchmarkDecompress(b *testing.B) {
	b.Run("zstd-cgo", func(b *testing.B) {
		b.Run("fastest", func(b *testing.B) {
			b.ReportAllocs()
			c, _ := New(CompressionAlgoZstdCgo, CompressionLevelZstdCgoFastest)
			r, _ := c.Compress(loremIpsumDolor)
			defer func() { _ = c.Close() }()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = c.Decompress(r)
			}
		})
		b.Run("default", func(b *testing.B) {
			b.ReportAllocs()
			c, _ := New(CompressionAlgoZstdCgo, CompressionLevelZstdCgoDefault)
			r, _ := c.Compress(loremIpsumDolor)
			defer func() { _ = c.Close() }()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = c.Decompress(r)
			}
		})
		b.Run("best", func(b *testing.B) {
			b.ReportAllocs()
			c, _ := New(CompressionAlgoZstdCgo, CompressionLevelZstdCgoBest)
			r, _ := c.Compress(loremIpsumDolor)
			defer func() { _ = c.Close() }()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = c.Decompress(r)
			}
		})
	})
}
