package config

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkReloadable/mutex-24						135314967			8.845 ns/op
// BenchmarkReloadable/rw_mutex-24					132994274			8.715 ns/op
// BenchmarkReloadable/reloadable_value-24				1000000000			0.6007 ns/op
// BenchmarkReloadable/reloadable_custom_mutex-24		77852116			15.19 ns/op
func BenchmarkReloadable(b *testing.B) {
	b.Run("mutex", func(b *testing.B) {
		var v reloadableMutex[int]
		go func() {
			for {
				v.Store(1)
				time.Sleep(time.Millisecond)
			}
		}()
		for i := 0; i < b.N; i++ {
			_ = v.Load()
		}
	})
	b.Run("rw mutex", func(b *testing.B) {
		var v reloadableRWMutex[int]
		go func() {
			for {
				v.Store(1)
				time.Sleep(time.Millisecond)
			}
		}()
		for i := 0; i < b.N; i++ {
			_ = v.Load()
		}
	})
	b.Run("reloadable value", func(b *testing.B) {
		var v reloadableValue[int]
		go func() {
			for {
				v.Store(1)
				time.Sleep(time.Millisecond)
			}
		}()
		for i := 0; i < b.N; i++ {
			_ = v.Load()
		}
	})
	b.Run("reloadable custom mutex", func(b *testing.B) {
		var v reloadableCustomMutex[int]
		go func() {
			for {
				v.Store(1)
				time.Sleep(time.Millisecond)
			}
		}()
		for i := 0; i < b.N; i++ {
			_ = v.Load()
		}
	})
}

type reloadableMutex[T comparable] struct {
	value T
	lock  sync.Mutex
}

func (a *reloadableMutex[T]) Load() T {
	a.lock.Lock()
	v := a.value
	a.lock.Unlock()
	return v
}

func (a *reloadableMutex[T]) Store(v T) {
	a.lock.Lock()
	a.value = v
	a.lock.Unlock()
}

type reloadableRWMutex[T comparable] struct {
	value T
	lock  sync.RWMutex
}

func (a *reloadableRWMutex[T]) Load() T {
	a.lock.RLock()
	v := a.value
	a.lock.RUnlock()
	return v
}

func (a *reloadableRWMutex[T]) Store(v T) {
	a.lock.Lock()
	a.value = v
	a.lock.Unlock()
}

type reloadableValue[T comparable] struct {
	// Note: it would also be possible to use reloadable.Pointer to avoid the panics from
	// atomic.Value but we won't be able to do the "swapIfNotEqual" as a single transaction anyway
	atomic.Value
}

func (a *reloadableValue[T]) Load() (zero T) {
	v := a.Value.Load()
	if v == nil {
		return zero
	}
	return v.(T)
}

func (a *reloadableValue[T]) Store(v T) {
	a.Value.Store(v)
}

type reloadableCustomMutex[T comparable] struct {
	value T
	mutex int32
}

func (a *reloadableCustomMutex[T]) lock() {
	for atomic.CompareAndSwapInt32(&a.mutex, 0, 1) {
	}
}

func (a *reloadableCustomMutex[T]) unlock() {
	for atomic.CompareAndSwapInt32(&a.mutex, 1, 0) {
	}
}

func (a *reloadableCustomMutex[T]) Load() T {
	a.lock()
	v := a.value
	a.unlock()
	return v
}

func (a *reloadableCustomMutex[T]) Store(v T) {
	a.lock()
	a.value = v
	a.unlock()
}
