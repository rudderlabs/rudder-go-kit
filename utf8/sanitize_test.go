package utf8

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestSanitizeInvalid(t *testing.T) {
	var (
		rb      = replacementCharByte
		hello   = []byte{72, 101, 108, 108, 111}       // Hello
		world   = []byte{228, 184, 150, 231, 149, 140} // 世界
		invalid = []byte{0xff, 0xfe, 0xfd}
	)

	tests := []struct {
		name     string
		valid    bool
		input    []byte
		expected []byte
	}{
		{"valid", true, []byte("Hello 世界"), []byte("Hello 世界")},
		{"invalid 1", false, invalid, []byte{rb, rb, rb}},
		{"invalid 2", false, []byte{0xd4, 0x6d}, []byte{rb, 0x6d}}, // pq: invalid byte sequence for encoding "UTF8": 0xd4 0x6d
		{"mixed", false, append(hello, append(invalid, world...)...), append(hello, append([]byte{rb, rb, rb}, world...)...)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.valid, utf8.Valid(tt.input))

			inputCopy := make([]byte, len(tt.input)) // Copy to avoid modifying the original input
			copy(inputCopy, tt.input)

			Sanitize(inputCopy)
			t.Logf("Sanitize(%s) = %s", tt.input, inputCopy)

			// After sanitization, the result should be valid
			require.True(t, utf8.Valid(inputCopy))
			require.Equal(t, tt.expected, inputCopy)
		})
	}
}
