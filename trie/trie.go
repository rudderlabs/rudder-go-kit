package trie

import "strings"

type Node struct {
	children    map[rune]*Node
	isEndOfWord bool
}

type Trie struct {
	root *Node
}

// NewTrie creates a new Trie
func NewTrie() *Trie {
	return &Trie{
		root: &Node{
			children:    make(map[rune]*Node),
			isEndOfWord: false,
		},
	}
}

// Insert adds a word to the trie
func (t *Trie) Insert(word string) {
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

// Get checks if any word in the trie matches the start of the input string
func (t *Trie) Get(input string) (bool, string) {
	node := t.root
	var matched strings.Builder
	for _, r := range input {
		node = node.children[r]
		if node == nil {
			break
		}

		matched.WriteRune(r)
		if node.isEndOfWord {
			return true, matched.String()
		}
	}
	return false, ""
}
