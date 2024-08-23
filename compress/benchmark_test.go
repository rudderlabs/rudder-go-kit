package compress

import (
	"testing"
)

// BenchmarkNew-24    	   55165	     22851 ns/op	   23884 B/op	       2 allocs/op
func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	c, _ := New(CompressionAlgoZstd, CompressionLevelZstdBest)
	defer func() { _ = c.Close() }()

	for i := 0; i < b.N; i++ {
		r, _ := c.Compress(loremIpsumDolor)
		r, _ = c.Decompress(r)
	}
}
