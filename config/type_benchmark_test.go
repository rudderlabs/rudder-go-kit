package config

import (
	"fmt"
	"testing"
)

func BenchmarkGetType(b *testing.B) {
	b.Run("getTypeName", func(b *testing.B) {
		getTypeName("test")
	})

	b.Run("sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%T", "test")
		}
	})
}

func BenchmarkGetStringValue(b *testing.B) {
	b.Run("getStringValue", func(b *testing.B) {
		getStringValue("test")
	})

	b.Run("sprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = fmt.Sprintf("%v", "test")
		}
	})
}
