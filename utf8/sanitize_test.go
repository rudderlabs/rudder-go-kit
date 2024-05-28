package utf8

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestSanitize(t *testing.T) {
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

func TestSanitizeInOut(t *testing.T) {
	ch := func(n int) string { return strings.Repeat(string(replacementCharByte), n) }

	toValidUTF8Tests := []struct {
		in  string
		out string
	}{
		{"", ""},
		{"abc", "abc"},
		{"\uFDDD", "\uFDDD"},
		{"a\xffb", "a" + ch(1) + "b"},
		{"a\xffb\uFFFD", "a" + ch(1) + "b\uFFFD"},
		{"a☺\xffb☺\xC0\xAFc☺\xff", "a☺" + ch(1) + "b☺" + ch(2) + "c☺" + ch(1)},
		{"\xC0\xAF", ch(2)},
		{"\xE0\x80\xAF", ch(3)},
		{"\xed\xa0\x80", ch(3)},
		{"\xed\xbf\xbf", ch(3)},
		{"\xF0\x80\x80\xaf", ch(4)},
		{"\xF8\x80\x80\x80\xAF", ch(5)},
		{"\xFC\x80\x80\x80\x80\xAF", ch(6)},
	}

	for _, tt := range toValidUTF8Tests {
		t.Run(tt.in, func(t *testing.T) {
			inputCopy := make([]byte, len(tt.in)) // Copy to avoid modifying the original input
			copy(inputCopy, tt.in)

			Sanitize(inputCopy)
			require.Equal(t, tt.out, string(inputCopy))
			require.True(t, utf8.Valid(inputCopy))
		})
	}
}
