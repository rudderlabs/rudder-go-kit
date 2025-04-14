package trie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrie(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("hello")

		found, matched := trie.Get("hello world")
		require.True(t, found)
		require.Equal(t, "hello", matched)

		found, matched = trie.Get("world")
		require.False(t, found)
		require.Equal(t, "", matched)

		// number and special characters
		trie.Insert("1234567890")
		trie.Insert("!@#$%^&*()")

		found, matched = trie.Get("1234567890")
		require.True(t, found)
		require.Equal(t, "1234567890", matched)

		found, matched = trie.Get("!@#$%^&*()1234567890")
		require.True(t, found)
		require.Equal(t, "!@#$%^&*()", matched)

		found, matched = trie.Get("1234567890!@#$%^&*()")
		require.True(t, found)
		require.Equal(t, "1234567890", matched)
	})

	t.Run("non-ascii", func(t *testing.T) {
		trie := NewTrie()

		trie.Insert("你好")
		trie.Insert("世界 ABCD")

		found, matched := trie.Get("你好世界")
		require.True(t, found)
		require.Equal(t, "你好", matched)

		found, matched = trie.Get("你")
		require.False(t, found)
		require.Equal(t, "", matched)

		found, matched = trie.Get("世界 ABCD")
		require.True(t, found)
		require.Equal(t, "世界 ABCD", matched)
	})

	t.Run("empty string", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("")

		// Test querying with an empty string. Even if "" was inserted,
		// it shouldn't be considered a valid match.
		found, matched := trie.Get("")
		require.False(t, found)
		require.Equal(t, "", matched)
	})

	t.Run("spaces and tabs", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert(" ")

		found, matched := trie.Get(" abc")
		require.True(t, found)
		require.Equal(t, " ", matched)

		found, matched = trie.Get("  ")
		require.True(t, found)
		require.Equal(t, " ", matched)

		found, matched = trie.Get("abc")
		require.False(t, found)
		require.Equal(t, "", matched)

		// tab
		trie.Insert("	")

		found, matched = trie.Get("	 abc")
		require.True(t, found)
		require.Equal(t, "	", matched)

		found, matched = trie.Get("	")
		require.True(t, found)
		require.Equal(t, "	", matched)

		found, matched = trie.Get("abc")
		require.False(t, found)
		require.Equal(t, "", matched)
	})

	t.Run("multiple word insertions with common prefixes", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("ab")
		trie.Insert("abc")

		found, matched := trie.Get("abcd")
		require.True(t, found)
		require.Equal(t, "ab", matched)

		found, matched = trie.Get("ab")
		require.True(t, found)
		require.Equal(t, "ab", matched)

		trie = NewTrie()
		trie.Insert("a")
		trie.Insert("ab")
		trie.Insert("abc")

		found, matched = trie.Get("abcd")
		require.True(t, found)
		require.Equal(t, "a", matched)

		found, matched = trie.Get("abd")
		require.True(t, found)
		require.Equal(t, "a", matched)

		found, matched = trie.Get("ad")
		require.True(t, found)
		require.Equal(t, "a", matched)

		found, matched = trie.Get("b")
		require.False(t, found)
		require.Equal(t, "", matched)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("Hello")

		found, matched := trie.Get("Hello")
		require.True(t, found)
		require.Equal(t, "Hello", matched)

		found, matched = trie.Get("hello")
		require.False(t, found)
		require.Equal(t, "", matched)

		trie.Insert("hello")
		found, matched = trie.Get("hello")
		require.True(t, found)
		require.Equal(t, "hello", matched)
	})

	t.Run("edge cases - unusual unicode", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("😀")
		trie.Insert("👩‍👧‍👦") // Family: woman, girl, boy ZWJ sequence

		found, matched := trie.Get("😀 Smiling Face With Open Mouth")
		require.True(t, found)
		require.Equal(t, "😀", matched)

		found, matched = trie.Get("👩‍👧‍👦 Family")
		require.True(t, found)
		require.Equal(t, "👩‍👧‍👦", matched)

		found, matched = trie.Get("👩‍👧‍") // Part of the ZWJ sequence
		require.False(t, found)
		require.Equal(t, "", matched)
	})

	t.Run("edge cases - long string", func(t *testing.T) {
		trie := NewTrie()
		longString := "This is a very long string that we are inserting into the trie to test its behavior with lengthy inputs, potentially exceeding buffer sizes or hitting performance limits if not implemented efficiently."
		trie.Insert(longString)

		found, matched := trie.Get(longString + " And some extra text.")
		require.True(t, found)
		require.Equal(t, longString, matched)

		found, matched = trie.Get(longString[:len(longString)-1]) // Slightly shorter
		require.False(t, found)
		require.Equal(t, "", matched)
	})

	t.Run("empty trie", func(t *testing.T) {
		trie := NewTrie()
		found, matched := trie.Get("anything")
		require.False(t, found)
		require.Equal(t, "", matched)

		found, matched = trie.Get("")
		require.False(t, found)
		require.Equal(t, "", matched)
	})

	t.Run("multiple inserts of the same word", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("test")
		trie.Insert("test")
		trie.Insert("test")

		found, matched := trie.Get("testing")
		require.True(t, found)
		require.Equal(t, "test", matched)

		found, matched = trie.Get("tes")
		require.False(t, found)
		require.Equal(t, "", matched)
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
		found, matched := trie.Get("abcde")
		require.True(t, found)
		require.Equal(t, "a", matched)

		found, matched = trie.Get("bcde")
		require.True(t, found)
		require.Equal(t, "b", matched)

		found, matched = trie.Get("abde")
		require.True(t, found)
		require.Equal(t, "a", matched)

		found, matched = trie.Get("cdef123456")
		require.True(t, found)
		require.Equal(t, "cdef", matched)

		found, matched = trie.Get("cdefg")
		require.True(t, found)
		require.Equal(t, "cdef", matched)

		// Test non-matches
		found, matched = trie.Get("c")
		require.False(t, found)
		require.Equal(t, "", matched)

		found, matched = trie.Get("cde1")
		require.False(t, found)
		require.Equal(t, "", matched)
	})
}
