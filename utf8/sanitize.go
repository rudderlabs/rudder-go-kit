package utf8

import (
	"unicode/utf8"
)

var replacementCharByte = byte(63) // 63 -> '?'

// Sanitize detects invalid UTF-8 byte sequences and replaces them with the replacement character.
// The slice length remains unchanged and no extra allocations are made.
// This is obtained by modifying the input slice in place and by making sure that the replacement character is a
// single-byte character. The side effect is that an invalid byte sequence is going to be replaced with multiple
// occurrences of the replacement characters e.g. "\xE0\x80\xAF" -> "???".
func Sanitize(data []byte) {
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			// Replace the invalid byte with the replacement character
			data[i] = replacementCharByte
		}
		i += size
	}
}
