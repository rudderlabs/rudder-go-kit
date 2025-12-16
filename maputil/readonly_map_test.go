package maputil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadonlyMap(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		t.Run("should return value and true when key exists", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2})
			val, ok := m.Get("a")
			require.True(t, ok)
			require.Equal(t, 1, val)
		})

		t.Run("should return zero value and false when key does not exist", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1})
			val, ok := m.Get("b")
			require.False(t, ok)
			require.Equal(t, 0, val)
		})

		t.Run("should work with empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			_, ok := m.Get("a")
			require.False(t, ok)
		})
	})

	t.Run("Has", func(t *testing.T) {
		t.Run("should return true when key exists", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2})
			require.True(t, m.Has("a"))
		})

		t.Run("should return false when key does not exist", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1})
			require.False(t, m.Has("b"))
		})

		t.Run("should work with empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			require.False(t, m.Has("a"))
		})
	})

	t.Run("Len", func(t *testing.T) {
		t.Run("should return correct length", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2, "c": 3})
			require.Equal(t, 3, m.Len())
		})

		t.Run("should return 0 for empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			require.Equal(t, 0, m.Len())
		})
	})

	t.Run("Keys", func(t *testing.T) {
		t.Run("should return all keys", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2, "c": 3})
			keys := make(map[string]bool)
			for k := range m.Keys() {
				keys[k] = true
			}
			require.Equal(t, 3, len(keys))
			require.True(t, keys["a"])
			require.True(t, keys["b"])
			require.True(t, keys["c"])
		})

		t.Run("should return empty iterator for empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			count := 0
			for range m.Keys() {
				count++
			}
			require.Equal(t, 0, count)
		})
	})

	t.Run("Values", func(t *testing.T) {
		t.Run("should return all values", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2, "c": 3})
			values := make(map[int]bool)
			for v := range m.Values() {
				values[v] = true
			}
			require.Equal(t, 3, len(values))
			require.True(t, values[1])
			require.True(t, values[2])
			require.True(t, values[3])
		})

		t.Run("should return empty iterator for empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			count := 0
			for range m.Values() {
				count++
			}
			require.Equal(t, 0, count)
		})
	})

	t.Run("ForEach", func(t *testing.T) {
		t.Run("should iterate over all key-value pairs", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2, "c": 3})
			pairs := make(map[string]int)
			m.ForEach(func(k string, v int) {
				pairs[k] = v
			})
			require.Equal(t, 3, len(pairs))
			require.Equal(t, 1, pairs["a"])
			require.Equal(t, 2, pairs["b"])
			require.Equal(t, 3, pairs["c"])
		})

		t.Run("should not call function for empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			called := false
			m.ForEach(func(k string, v int) {
				called = true
			})
			require.False(t, called)
		})
	})

	t.Run("Append", func(t *testing.T) {
		t.Run("should return new map with added entries", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2})
			newM := m.Append(map[string]int{"c": 3, "d": 4})

			require.Equal(t, 4, newM.Len())
			val, ok := newM.Get("c")
			require.True(t, ok)
			require.Equal(t, 3, val)
			val, ok = newM.Get("d")
			require.True(t, ok)
			require.Equal(t, 4, val)
		})

		t.Run("should not modify original map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2})
			_ = m.Append(map[string]int{"c": 3})

			require.Equal(t, 2, m.Len())
			require.False(t, m.Has("c"))
		})

		t.Run("should overwrite existing keys", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2})
			newM := m.Append(map[string]int{"a": 10, "c": 3})

			val, ok := newM.Get("a")
			require.True(t, ok)
			require.Equal(t, 10, val)
			require.Equal(t, 3, newM.Len())
		})

		t.Run("should work with empty entries", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1})
			newM := m.Append(map[string]int{})

			require.Equal(t, 1, newM.Len())
		})

		t.Run("should work when appending to empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			newM := m.Append(map[string]int{"a": 1, "b": 2})

			require.Equal(t, 2, newM.Len())
		})
	})

	t.Run("Remove", func(t *testing.T) {
		t.Run("should return new map with removed entries", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2})
			newM := m.Remove([]string{"a"})

			require.Equal(t, 1, newM.Len())
			require.True(t, newM.Has("b"))
		})

		t.Run("should not modify original map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1, "b": 2})
			_ = m.Remove([]string{"a"})

			require.Equal(t, 2, m.Len())
		})

		t.Run("should work with empty entries", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{"a": 1})
			newM := m.Remove([]string{})
			require.Equal(t, 1, newM.Len())
		})

		t.Run("should work when removing from an empty map", func(t *testing.T) {
			m := NewReadOnlyMap(map[string]int{})
			newM := m.Remove([]string{"a", "b"})
			require.Equal(t, 0, newM.Len())
		})
	})

	t.Run("newReadOnlyMap", func(t *testing.T) {
		t.Run("should create map with provided data", func(t *testing.T) {
			data := map[string]int{"a": 1, "b": 2}
			m := NewReadOnlyMap(data)

			require.Equal(t, 2, m.Len())
			val, ok := m.Get("a")
			require.True(t, ok)
			require.Equal(t, 1, val)
		})

		t.Run("should work with nil map", func(t *testing.T) {
			m := NewReadOnlyMap[string, int](nil)

			require.Equal(t, 0, m.Len())
		})
	})
}
