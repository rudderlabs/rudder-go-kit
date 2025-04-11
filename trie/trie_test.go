package trie

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrie(t *testing.T) {
	t.Run("Insert and Get", func(t *testing.T) {
		trie := NewTrie()
		trie.Insert("hello")

		found, matched := trie.Get("hello world")
		require.True(t, found)
		require.Equal(t, "hello", matched)

		found, matched = trie.Get("world")
		require.False(t, found)
		require.Equal(t, "", matched)

		// non-ascii
		trie.Insert("你好")
		trie.Insert("世界 ABCD")

		found, matched = trie.Get("你好世界")
		require.True(t, found)
		require.Equal(t, "你好", matched)

		found, matched = trie.Get("你")
		require.False(t, found)
		require.Equal(t, "", matched)

		found, matched = trie.Get("世界 ABCD")
		require.True(t, found)
		require.Equal(t, "世界 ABCD", matched)

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
}
