package sanitize

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

var out string

func BenchmarkMessageID(b *testing.B) {
	dirtyMessageID := "\u0000 Test foo_bar-baz \u034F 123-222 "
	properMessageID := "123e4567-e89b-12d3-a456-426614174000"

	b.Run("in-place for loop - dirty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out = sanitizeMessageIDForLoop(dirtyMessageID)
		}
	})

	b.Run("in-place for loop - proper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out = sanitizeMessageIDForLoop(properMessageID)
		}
	})

	b.Run("strings map - dirty", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out = Unicode(dirtyMessageID)
		}
	})

	b.Run("strings map - proper", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			out = Unicode(properMessageID)
		}
	})
}

// incorrect implementation of sanitizeMessageID, but used for benchmarking
func sanitizeMessageIDForLoop(messageID string) string {
	for i, r := range messageID {
		if unicode.IsPrint(r) {
			continue
		}
		if !unicode.Is(invisibleRangeTable, r) {
			continue
		}

		messageID = messageID[:i] + messageID[i+1:]
	}
	return messageID
}

func TestSanitizeMessageID(t *testing.T) {
	testcases := []struct {
		in  string
		out string
	}{
		{"\u0000 Test \u0000foo_bar-baz 123-222 \u0000", " Test foo_bar-baz 123-222 "},
		{"\u0000", ""},
		{"\u0000 ", " "},
		{"\u0000 \u0000", " "},
		{"\u00A0\t\n\r\u034F", ""},
		{"τυχαίο;", "τυχαίο;"},
	}

	for _, tc := range testcases {
		cleanMessageID := Unicode(tc.in)
		require.Equal(t, tc.out, cleanMessageID, fmt.Sprintf("%#v -> %#v", tc.in, tc.out))
	}

	for _, r := range invisibleRunes {
		cleanMessageID := Unicode(string(r))
		require.Empty(t, cleanMessageID, fmt.Sprintf("%U", r))
	}
}

func TestSanitize(t *testing.T) {
	var (
		rb      = replacementInvalidCharByte
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

			ReplaceInvalid(inputCopy)
			t.Logf("ReplaceInvalid(%s) = %s", tt.input, inputCopy)

			// After sanitization, the result should be valid
			require.True(t, utf8.Valid(inputCopy))
			require.Equal(t, tt.expected, inputCopy)
		})
	}
}

func TestSanitizeInOut(t *testing.T) {
	ch := func(n int) string { return strings.Repeat(string(replacementInvalidCharByte), n) }

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
		{"\u0000", " "},
	}

	for _, tt := range toValidUTF8Tests {
		t.Run(tt.in, func(t *testing.T) {
			inputCopy := make([]byte, len(tt.in)) // Copy to avoid modifying the original input
			copy(inputCopy, tt.in)

			ReplaceInvalid(inputCopy)
			require.Equal(t, tt.out, string(inputCopy))
			require.True(t, utf8.Valid(inputCopy))
		})
	}
}

func TestJSONUnmarshal(t *testing.T) {
	type test struct {
		Value string `json:"value"`
	}

	testCases := [][]byte{
		append([]byte(`{"value": "a`), append([]byte("\u0000"), []byte(`b"}`)...)...),
		append([]byte(`{"value": "a`), append([]byte("\u0000\uFFA0\u206B"), []byte(`b"}`)...)...),
		append([]byte(`{"value": "a`), append([]byte("\u0000OK\uFFA0\u206B"), []byte(`b"}`)...)...),
	}
	for i, tc := range testCases {
		t.Run(fmt.Sprintf("test-%d", i), func(t *testing.T) {
			ts := test{}

			// the first unmarshal is expected to fail
			require.Error(t, json.Unmarshal(tc, &ts))

			// sanitize the input and try again
			ReplaceInvalid(tc)

			// the second unmarshal is expected to succeed
			require.NoError(t, json.Unmarshal(tc, &ts))
		})
	}
}

func TestPostgres(t *testing.T) {
	t.Skip("TODO")
}
