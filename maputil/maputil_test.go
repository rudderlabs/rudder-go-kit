package maputil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapLookup(t *testing.T) {
	t.Run("simple lookups and edge cases", func(t *testing.T) {
		tests := []struct {
			name     string
			m        map[string]interface{}
			keys     []string
			expected interface{}
		}{
			{
				name:     "returns value if key is present",
				m:        map[string]interface{}{"hello": "world"},
				keys:     []string{"hello"},
				expected: "world",
			},
			{
				name:     "returns nil if key is not present",
				m:        map[string]interface{}{"foo": "bar"},
				keys:     []string{"baz"},
				expected: nil,
			},
			{
				name:     "returns nil for empty map",
				m:        map[string]interface{}{},
				keys:     []string{"foo"},
				expected: nil,
			},
			{
				name:     "returns value for empty string key",
				m:        map[string]interface{}{"": "empty"},
				keys:     []string{""},
				expected: "empty",
			},
			{
				name:     "returns nil if no keys are provided",
				m:        map[string]interface{}{"foo": "bar"},
				keys:     []string{},
				expected: nil,
			},
			{
				name:     "returns nil if mapToLookup is nil",
				m:        nil,
				keys:     []string{"foo"},
				expected: nil,
			},
			{
				name:     "returns nil if mapToLookup is not a map",
				m:        nil,
				keys:     nil,
				expected: nil,
			},
		}
		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.expected, MapLookup(tt.m, tt.keys...))
			})
		}
	})

	t.Run("nested map lookups", func(t *testing.T) {
		m := map[string]interface{}{
			"outer": map[string]interface{}{
				"inner": "value",
			},
		}
		require.Equal(t, "value", MapLookup(m, "outer", "inner"))
		require.Nil(t, MapLookup(m, "outer", "missing"))
		require.Nil(t, MapLookup(m, "missing", "inner"))
	})

	t.Run("deeply nested and non-ascii keys", func(t *testing.T) {
		m := map[string]interface{}{
			"héllo": map[string]interface{}{
				"nested": map[string]interface{}{
					"más": map[string]interface{}{
						"föö":     "bår",
						"emoji":   "😀",
						"русский": "язык",
						"中文":      "汉字",
					},
				},
			},
		}
		require.Equal(t, "bår", MapLookup(m, "héllo", "nested", "más", "föö"))
		require.Equal(t, "😀", MapLookup(m, "héllo", "nested", "más", "emoji"))
		require.Equal(t, "язык", MapLookup(m, "héllo", "nested", "más", "русский"))
		require.Equal(t, "汉字", MapLookup(m, "héllo", "nested", "más", "中文"))
	})

	t.Run("non-ascii characters in keys", func(t *testing.T) {
		m := map[string]interface{}{
			"héllo": map[string]interface{}{
				"föö":   "bår",
				"emoji": "😀",
				"中文":    "汉字",
			},
		}
		require.Equal(t, "bår", MapLookup(m, "héllo", "föö"))
		require.Equal(t, "😀", MapLookup(m, "héllo", "emoji"))
		require.Equal(t, "汉字", MapLookup(m, "héllo", "中文"))
	})

	t.Run("nil and non-map intermediate values", func(t *testing.T) {
		m1 := map[string]interface{}{"foo": nil}
		require.Nil(t, MapLookup(m1, "foo"))
		m2 := map[string]interface{}{"foo": "bar"}
		require.Nil(t, MapLookup(m2, "foo", "baz"))
		m3 := map[string]interface{}{"foo": map[string]interface{}{"bar": nil}}
		require.Nil(t, MapLookup(m3, "foo", "bar"))
	})

	t.Run("intermediate key missing", func(t *testing.T) {
		m := map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}
		require.Nil(t, MapLookup(m, "foo", "missing", "baz"))
	})

	t.Run("nested map with non-string keys", func(t *testing.T) {
		m := map[string]interface{}{
			"nested": map[interface{}]interface{}{
				1:     "one",
				"foo": "bar",
			},
		}
		require.Nil(t, MapLookup(m, "nested", "foo"))
	})

	t.Run("value is an empty interface nil", func(t *testing.T) {
		var nilValue interface{} = nil
		m := map[string]interface{}{
			"foo": nilValue,
		}
		require.Nil(t, MapLookup(m, "foo"))
	})

	t.Run("multiple keys, first key missing", func(t *testing.T) {
		m := map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": "baz",
			},
		}
		require.Nil(t, MapLookup(m, "missing", "bar"))
	})

	t.Run("nested nil value", func(t *testing.T) {
		m := map[string]interface{}{
			"foo": map[string]interface{}{
				"bar": nil,
			},
		}
		require.Nil(t, MapLookup(m, "foo", "bar"))
	})

	t.Run("nested non-map value", func(t *testing.T) {
		m := map[string]interface{}{
			"foo": "not a map",
		}
		require.Nil(t, MapLookup(m, "foo", "bar"))
	})
}
