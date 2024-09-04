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
