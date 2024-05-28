package utf8

import "unicode/utf8"

const replacementChar = string('?')

var replacementCharByte = []byte(replacementChar)[0]

func init() {
	if len([]byte(replacementChar)) != 1 {
		panic("replacementChar must be a single-byte character")
	}
}

// Sanitize detects invalid UTF-8 byte sequences and replaces them with the replacement character.
// The slice length remains unchanged and no extra allocations are made.
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

// Invalid returns true if the input contains invalid UTF-8 byte sequences.
func Invalid(data []byte) bool {
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			return true
		}
		i += size
	}
	return false
}
