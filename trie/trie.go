package trie

import "strings"

// Trie is a prefix tree data structure.
// It stores words in a way that allows for efficient prefix-based searching.

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
// Note: If the input string is empty, the function will not insert anything.
func (t *Trie) Insert(word string) {
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

// Get checks if any prefix of the input string is registered as a complete word
// in the trie.
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
// Note: If the input string is empty, the function returns (false, "") immediately,
// regardless of whether an empty string was inserted via Insert(""). The root node's
// isEndOfWord flag is not considered in this specific scenario.
func (t *Trie) Get(input string) (bool, string) {
	node := t.root
	var matched strings.Builder
	for _, r := range input {
		node = node.children[r]
		if node == nil {
			// No further matching prefix found in the trie.
			break
		}

		matched.WriteRune(r)
		if node.isEndOfWord {
			// Found a prefix that is a word.
			return true, matched.String()
		}
	}

	// Reached the end of the input or a dead end in the trie
	// without finding a node marked as the end of a word.
	return false, ""
}
