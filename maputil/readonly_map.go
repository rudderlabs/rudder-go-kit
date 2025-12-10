package maputil

import (
	"iter"
	"maps"
)

// NewReadOnlyMap creates a new read-only map from an existing map
// The original map is copied to prevent external modifications
func NewReadOnlyMap[K comparable, V any](data map[K]V) *ReadOnlyMap[K, V] {
	return &ReadOnlyMap[K, V]{
		data: maps.Clone(data),
	}
}

// ReadOnlyMap provides a read-only view of a map
type ReadOnlyMap[K comparable, V any] struct {
	data map[K]V
}

// Get retrieves a value by key
func (r *ReadOnlyMap[K, V]) Get(key K) (V, bool) {
	val, ok := r.data[key]
	return val, ok
}

// Has checks if a key exists
func (r *ReadOnlyMap[K, V]) Has(key K) bool {
	_, ok := r.data[key]
	return ok
}

// Len returns the number of elements
func (r *ReadOnlyMap[K, V]) Len() int {
	return len(r.data)
}

// Keys returns all keys
func (r *ReadOnlyMap[K, V]) Keys() iter.Seq[K] {
	return maps.Keys(r.data)
}

// Values returns all values
func (r *ReadOnlyMap[K, V]) Values() iter.Seq[V] {
	return maps.Values(r.data)
}

// ForEach iterates over all key-value pairs
func (r *ReadOnlyMap[K, V]) ForEach(fn func(K, V)) {
	for k, v := range r.data {
		fn(k, v)
	}
}

// Append returns a new read-only map with the provided entries added
func (r *ReadOnlyMap[K, V]) Append(entries map[K]V) *ReadOnlyMap[K, V] {
	newMap := maps.Clone(r.data)
	maps.Copy(newMap, entries)
	return &ReadOnlyMap[K, V]{data: newMap}
}

// Remove returns a new read-only map with the provided keys removed
func (r *ReadOnlyMap[K, V]) Remove(entries []K) *ReadOnlyMap[K, V] {
	newMap := maps.Clone(r.data)
	for _, k := range entries {
		delete(newMap, k)
	}
	return &ReadOnlyMap[K, V]{data: newMap}
}
