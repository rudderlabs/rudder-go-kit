package trie

import "strings"

// Trie is a prefix tree data structure.
// It stores words in a way that allows for efficient prefix-based searching.
// This implementation is case sensitive; "word" and "Word" are treated as different entries.

// This implementation is not concurrent safe.
// It should not be used from multiple goroutines concurrently without additional synchronization.

// Node represents a single node within the trie.
type Node struct {
	children    map[rune]*Node
	isEndOfWord bool
}

type Trie struct {
	root *Node
}

// NewTrie creates and initializes a new, empty Trie.
func NewTrie() *Trie {
	return &Trie{
		root: &Node{
			children:    make(map[rune]*Node),
			isEndOfWord: false,
		},
	}
}

// Insert adds a word to the trie. If the word already exists,
// its end marker might be updated, but the structure remains the same.
// Intermediate nodes are created as needed.
//
// Note: trailing and leading spaces are trimmed before inserting.
// If the input string is empty, the function will not insert anything.
func (t *Trie) Insert(word string) {
	word = strings.TrimSpace(word)
	if word == "" {
		return
	}

	node := t.root
	for _, r := range word {
		child := node.children[r]
		if child == nil {
			child = &Node{
				children:    make(map[rune]*Node),
				isEndOfWord: false,
			}
			node.children[r] = child
		}
		node = child
	}
	node.isEndOfWord = true
}

// GetMatchedPrefixWord checks if any prefix of the input string is registered as a complete word
// in the trie. Use HasMatchedPrefixWords for performance when only the existence of a matching
// prefix is needed.
//
// It traverses the trie character by character based on the input string.
// If it encounters a node marked as the end of a word during this traversal,
// it means a prefix of the input is a complete word in the trie.
//
// Return values:
//   - bool: true if a prefix matching a complete word is found, false otherwise.
//   - string: The shortest matching prefix word found. Returns an empty string
//     if no matching prefix word is found.
//
// Note: If the input string is empty, the function always returns (false, "")
func (t *Trie) GetMatchedPrefixWord(input string) (bool, string) {
	return t.traverse(input, true)
}

// HasMatchedPrefixWords checks if any prefix of the input string is registered as a complete word
// in the trie. Unlike GetMatchedPrefixWord, it does not construct the matching prefix string.
//
// It traverses the trie character by character based on the input string.
// If it encounters a node marked as the end of a word during this traversal,
// it means a prefix of the input is a complete word in the trie.
//
// Return value:
//   - bool: true if a prefix matching a complete word is found, false otherwise.
//
// Note: This method is optimized for performance when only the existence of a matching
// prefix is needed, as it avoids string construction operations.
func (t *Trie) HasMatchedPrefixWords(input string) bool {
	matched, _ := t.traverse(input, false)
	return matched
}

func (t *Trie) traverse(input string, buildString bool) (bool, string) {
	node := t.root
	var matched strings.Builder
	for _, r := range input {
		node = node.children[r]
		if node == nil {
			// No further matching prefix found in the trie.
			break
		}

		if buildString {
			_, err := matched.WriteRune(r)
			if err != nil {
				panic(err)
			}
		}

		if node.isEndOfWord {
			// Found a prefix that is a word.
			if buildString {
				return true, matched.String()
			}
			return true, ""
		}
	}

	// Reached the end of the input or a dead end in the trie
	// without finding a node marked as the end of a word.
	return false, ""
}
