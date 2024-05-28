package sanitize

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/rangetable"
)

// invisibleRunes unicode.IsPrint does not include all invisible characters,
// so I got this list from https://invisible-characters.com/
var invisibleRunes = []rune{
	'\u0000', // NULL
	'\u0009', // CHARACTER TABULATION
	'\u00A0', // NO-BREAK SPACE
	'\u00AD', // SOFT HYPHEN
	'\u034F', // COMBINING GRAPHEME JOINER
	'\u061C', // ARABIC LETTER MARK
	'\u115F', // HANGUL CHOSEONG FILLER
	'\u1160', // HANGUL JUNGSEONG FILLER
	'\u17B4', // KHMER VOWEL INHERENT AQ
	'\u17B5', // KHMER VOWEL INHERENT AA
	'\u180E', // MONGOLIAN VOWEL SEPARATOR
	'\u2000', // EN QUAD
	'\u2001', // EM QUAD
	'\u2002', // EN SPACE
	'\u2003', // EM SPACE
	'\u2004', // THREE-PER-EM SPACE
	'\u2005', // FOUR-PER-EM SPACE
	'\u2006', // SIX-PER-EM SPACE
	'\u2007', // FIGURE SPACE
	'\u2008', // PUNCTUATION SPACE
	'\u2009', // THIN SPACE
	'\u200A', // HAIR SPACE
	'\u200B', // ZERO WIDTH SPACE
	'\u200C', // ZERO WIDTH NON-JOINER
	'\u200D', // ZERO WIDTH JOINER
	'\u200E', // LEFT-TO-RIGHT MARK
	'\u200F', // RIGHT-TO-LEFT MARK
	'\u202F', // NARROW NO-BREAK SPACE
	'\u205F', // MEDIUM MATHEMATICAL SPACE
	'\u2060', // WORD JOINER
	'\u2061', // FUNCTION APPLICATION
	'\u2062', // INVISIBLE TIMES
	'\u2063', // INVISIBLE SEPARATOR
	'\u2064', // INVISIBLE PLUS
	'\u206A', // INHIBIT SYMMETRIC SWAPPING
	'\u206B', // ACTIVATE SYMMETRIC SWAPPING
	'\u206C', // INHIBIT ARABIC FORM SHAPING
	'\u206D', // ACTIVATE ARABIC FORM SHAPING
	'\u206E', // NATIONAL DIGIT SHAPES
	'\u206F', // NOMINAL DIGIT SHAPES
	'\u3000', // IDEOGRAPHIC SPACE
	'\u2800', // BRAILLE PATTERN BLANK
	'\u3164', // HANGUL FILLER
	'\uFEFF', // ZERO WIDTH NO-BREAK SPACE
	'\uFFA0', // HALF WIDTH HANGUL FILLER
}

var (
	invisibleRangeTable *unicode.RangeTable
	invisibleRunesMap   = map[rune]struct{}{}

	replacementInvalidCharByte   = byte(63) // 63 -> '?'
	replacementInvisibleCharByte = byte(32) // 32 -> ' '
)

func init() {
	invisibleRangeTable = rangetable.New(invisibleRunes...)
	for _, r := range invisibleRunes {
		invisibleRunesMap[r] = struct{}{}
	}
}

// Unicode removes irregularly invisible characters from a string.
//
// Irregularly invisible characters are defined as:
//   - Non-printable characters according to Go's unicode package (unicode.IsPrint).
//   - Characters in the invisibleRunes list (https://invisible-characters.com/).
//
// Note: Regular ASCII space (0x20) is not removed.
func Unicode(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(invisibleRangeTable, r) || !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, str)
}

// ReplaceInvalid detects invalid UTF-8 byte sequences and replaces them with the replacement character.
// It also detects invisible characters and replaces them with a space character.
// The slice length remains unchanged and no extra allocations are made.
// This is obtained by modifying the input slice in place and by making sure that the replacement character is a
// single-byte character. The side effect is that an invalid byte sequence is going to be replaced with multiple
// occurrences of the replacement characters e.g. "\xE0\x80\xAF" -> "???".
func ReplaceInvalid(data []byte) {
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size == 1 {
			// Replace the invalid byte with the replacement character
			data[i] = replacementInvalidCharByte
		} else if _, ok := invisibleRunesMap[r]; ok {
			// Replace the invisible character with a space
			for j := 0; j < size; j++ {
				data[i+j] = replacementInvisibleCharByte
			}
		}
		i += size
	}
}
