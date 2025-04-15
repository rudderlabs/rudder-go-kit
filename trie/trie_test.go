package trie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrie(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("hello")

		found, matched := trie.GetMatchedPrefixWord("hello world")
		require.True(t, found)
		require.Equal(t, "hello", matched)
		require.True(t, trie.HasMatchedPrefixWords("hello world"))

		found, matched = trie.GetMatchedPrefixWord("world")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("world"))

		// number and special characters
		trie.Insert("1234567890")
		trie.Insert("!@#$%^&*()")

		found, matched = trie.GetMatchedPrefixWord("1234567890")
		require.True(t, found)
		require.Equal(t, "1234567890", matched)
		require.True(t, trie.HasMatchedPrefixWords("1234567890"))

		found, matched = trie.GetMatchedPrefixWord("!@#$%^&*()1234567890")
		require.True(t, found)
		require.Equal(t, "!@#$%^&*()", matched)
		require.True(t, trie.HasMatchedPrefixWords("!@#$%^&*()1234567890"))

		found, matched = trie.GetMatchedPrefixWord("1234567890!@#$%^&*()")
		require.True(t, found)
		require.Equal(t, "1234567890", matched)
		require.True(t, trie.HasMatchedPrefixWords("1234567890!@#$%^&*()"))
	})

	t.Run("only complete words can match", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("Message in a bottle")

		found, matched := trie.GetMatchedPrefixWord("bot")
		require.False(t, found)
		require.Empty(t, matched)
		require.False(t, trie.HasMatchedPrefixWords("bot"))
	})

	t.Run("non-ascii", func(t *testing.T) {
		trie := NewTrie()

		trie.Insert("你好")
		trie.Insert("世界 ABCD")

		found, matched := trie.GetMatchedPrefixWord("你好世界")
		require.True(t, found)
		require.Equal(t, "你好", matched)
		require.True(t, trie.HasMatchedPrefixWords("你好世界"))

		found, matched = trie.GetMatchedPrefixWord("你")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("你"))

		found, matched = trie.GetMatchedPrefixWord("世界 ABCD")
		require.True(t, found)
		require.Equal(t, "世界 ABCD", matched)
		require.True(t, trie.HasMatchedPrefixWords("世界 ABCD"))
	})

	t.Run("empty string", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("") // word is not added to the trie as it's empty after trimming

		found, matched := trie.GetMatchedPrefixWord("")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords(""))
	})

	t.Run("spaces and tabs", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert(" ") // word is not added to the trie as it's empty after trimming
		trie.Insert("    abc   ")

		found, matched := trie.GetMatchedPrefixWord(" abc")
		require.False(t, found)
		require.Empty(t, matched)
		require.False(t, trie.HasMatchedPrefixWords(" abc"))

		found, matched = trie.GetMatchedPrefixWord("abc   ")
		require.True(t, found)
		require.Equal(t, "abc", matched)
		require.True(t, trie.HasMatchedPrefixWords("abc   "))

		found, matched = trie.GetMatchedPrefixWord("  ")
		require.False(t, found)
		require.Empty(t, matched)
		require.False(t, trie.HasMatchedPrefixWords("  "))

		found, matched = trie.GetMatchedPrefixWord("abc")
		require.True(t, found)
		require.Equal(t, "abc", matched)
		require.True(t, trie.HasMatchedPrefixWords("abc"))

		// tab
		trie.Insert("	") // word is not added to the trie as it's empty after trimming
		trie.Insert("	 xyz	")

		found, matched = trie.GetMatchedPrefixWord("	 xyz")
		require.False(t, found)
		require.Empty(t, matched)
		require.False(t, trie.HasMatchedPrefixWords("	 xyz"))

		found, matched = trie.GetMatchedPrefixWord("xyz		")
		require.True(t, found)
		require.Equal(t, "xyz", matched)
		require.True(t, trie.HasMatchedPrefixWords("xyz		"))

		found, matched = trie.GetMatchedPrefixWord("	")
		require.False(t, found)
		require.Empty(t, matched)
		require.False(t, trie.HasMatchedPrefixWords("	"))

		found, matched = trie.GetMatchedPrefixWord("xyz")
		require.True(t, found)
		require.Equal(t, "xyz", matched)
		require.True(t, trie.HasMatchedPrefixWords("xyz"))
	})

	t.Run("multiple word insertions with common prefixes", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("ab")
		trie.Insert("abc")

		found, matched := trie.GetMatchedPrefixWord("abcd")
		require.True(t, found)
		require.Equal(t, "ab", matched)
		require.True(t, trie.HasMatchedPrefixWords("abcd"))

		found, matched = trie.GetMatchedPrefixWord("ab")
		require.True(t, found)
		require.Equal(t, "ab", matched)
		require.True(t, trie.HasMatchedPrefixWords("ab"))

		trie = NewTrie()
		trie.Insert("a")
		trie.Insert("ab")
		trie.Insert("abc")

		found, matched = trie.GetMatchedPrefixWord("abcd")
		require.True(t, found)
		require.Equal(t, "a", matched)
		require.True(t, trie.HasMatchedPrefixWords("abcd"))

		found, matched = trie.GetMatchedPrefixWord("abd")
		require.True(t, found)
		require.Equal(t, "a", matched)
		require.True(t, trie.HasMatchedPrefixWords("abd"))

		found, matched = trie.GetMatchedPrefixWord("ad")
		require.True(t, found)
		require.Equal(t, "a", matched)
		require.True(t, trie.HasMatchedPrefixWords("ad"))

		found, matched = trie.GetMatchedPrefixWord("b")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("b"))
	})

	t.Run("case sensitivity", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("Hello")

		found, matched := trie.GetMatchedPrefixWord("Hello")
		require.True(t, found)
		require.Equal(t, "Hello", matched)
		require.True(t, trie.HasMatchedPrefixWords("Hello"))

		found, matched = trie.GetMatchedPrefixWord("hello")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("hello"))

		trie.Insert("hello")
		found, matched = trie.GetMatchedPrefixWord("hello")
		require.True(t, found)
		require.Equal(t, "hello", matched)
		require.True(t, trie.HasMatchedPrefixWords("hello"))
	})

	t.Run("edge cases - unusual unicode", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("😀")
		trie.Insert("👩‍👧‍👦") // Family: woman, girl, boy ZWJ sequence

		found, matched := trie.GetMatchedPrefixWord("😀 Smiling Face With Open Mouth")
		require.True(t, found)
		require.Equal(t, "😀", matched)
		require.True(t, trie.HasMatchedPrefixWords("😀 Smiling Face With Open Mouth"))

		found, matched = trie.GetMatchedPrefixWord("👩‍👧‍👦 Family")
		require.True(t, found)
		require.Equal(t, "👩‍👧‍👦", matched)
		require.True(t, trie.HasMatchedPrefixWords("👩‍👧‍👦 Family"))

		found, matched = trie.GetMatchedPrefixWord("👩‍👧‍") // Part of the ZWJ sequence
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("👩‍👧‍"))
	})

	t.Run("edge cases - long string", func(t *testing.T) {
		trie := NewTrie()
		longString := "This is a very long string that we are inserting into the trie to test its behavior with lengthy inputs, potentially exceeding buffer sizes or hitting performance limits if not implemented efficiently."
		trie.Insert(longString)

		found, matched := trie.GetMatchedPrefixWord(longString + " And some extra text.")
		require.True(t, found)
		require.Equal(t, longString, matched)
		require.True(t, trie.HasMatchedPrefixWords(longString+" And some extra text."))

		found, matched = trie.GetMatchedPrefixWord(longString[:len(longString)-1]) // Slightly shorter
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords(longString[:len(longString)-1]))
	})

	t.Run("empty trie", func(t *testing.T) {
		trie := NewTrie()
		found, matched := trie.GetMatchedPrefixWord("anything")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("anything"))

		found, matched = trie.GetMatchedPrefixWord("")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("anything"))
	})

	t.Run("multiple inserts of the same word", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("test")
		trie.Insert("test")
		trie.Insert("test")

		found, matched := trie.GetMatchedPrefixWord("testing")
		require.True(t, found)
		require.Equal(t, "test", matched)
		require.True(t, trie.HasMatchedPrefixWords("testing"))

		found, matched = trie.GetMatchedPrefixWord("tes")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("tes"))
	})

	t.Run("deeper trie structure", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("a")
		trie.Insert("ab")
		trie.Insert("abc")
		trie.Insert("abcd")
		trie.Insert("b")
		trie.Insert("bc")
		trie.Insert("bcd")
		trie.Insert("abd") // Branching off 'ab'
		trie.Insert("cdef")
		trie.Insert("cdefg")

		// Test deepest matches
		found, matched := trie.GetMatchedPrefixWord("abcde")
		require.True(t, found)
		require.Equal(t, "a", matched)
		require.True(t, trie.HasMatchedPrefixWords("abcde"))

		found, matched = trie.GetMatchedPrefixWord("bcde")
		require.True(t, found)
		require.Equal(t, "b", matched)
		require.True(t, trie.HasMatchedPrefixWords("bcde"))

		found, matched = trie.GetMatchedPrefixWord("abde")
		require.True(t, found)
		require.Equal(t, "a", matched)
		require.True(t, trie.HasMatchedPrefixWords("abde"))

		found, matched = trie.GetMatchedPrefixWord("cdef123456")
		require.True(t, found)
		require.Equal(t, "cdef", matched)
		require.True(t, trie.HasMatchedPrefixWords("cdef123456"))

		found, matched = trie.GetMatchedPrefixWord("cdefg")
		require.True(t, found)
		require.Equal(t, "cdef", matched)
		require.True(t, trie.HasMatchedPrefixWords("cdefg"))

		// Test non-matches
		found, matched = trie.GetMatchedPrefixWord("c")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("c"))

		found, matched = trie.GetMatchedPrefixWord("cde1")
		require.False(t, found)
		require.Equal(t, "", matched)
		require.False(t, trie.HasMatchedPrefixWords("cde1"))
	})
}
