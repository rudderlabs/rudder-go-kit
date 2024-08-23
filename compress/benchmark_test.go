package compress

import (
	"strings"
	"testing"
)

// BenchmarkNew-24    	      85	  44396834 ns/op	909630999 B/op	     119 allocs/op
func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	c, _ := New(CompressionAlgoZstd)

	for i := 0; i < b.N; i++ {
		r, _ := c.Compress(strings.NewReader(loremIpsumDolor), WithCompressionLevel(CompressionLevelZstdBest))
		r, _ = c.Decompress(r)
	}
}
