package config

import "sync"

// SingleValueLoader returns a ValueLoader that always returns the same value.
func SingleValueLoader[T any](v T) ValueLoader[T] {
	return &loader[T]{v}
}

// ValueLoader is an interface that can be used to load a value.
type ValueLoader[T any] interface {
	Load() T
}

// loader is a ValueLoader that always returns the same value.
type loader[T any] struct {
	v T
}

// Load returns the value.
func (l *loader[T]) Load() T {
	return l.v
}

// NewMockValueLoader creates a new MockValueLoader with the given initial value.
func NewMockValueLoader[T any](initialValue T) *MockValueLoader[T] {
	return &MockValueLoader[T]{value: initialValue}
}

// MockValueLoader is a simple mock for config.ValueLoader
type MockValueLoader[T any] struct {
	value T
	mu    sync.RWMutex
}

func (m *MockValueLoader[T]) Load() T {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

func (m *MockValueLoader[T]) Set(value T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = value
}
