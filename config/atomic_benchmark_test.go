package config

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkAtomic/mutex-24						135314967			8.845 ns/op
// BenchmarkAtomic/rw_mutex-24					132994274			8.715 ns/op
// BenchmarkAtomic/atomic_value-24				1000000000			0.6007 ns/op
// BenchmarkAtomic/atomic_custom_mutex-24		77852116			15.19 ns/op
func BenchmarkAtomic(b *testing.B) {
	b.Run("mutex", func(b *testing.B) {
		var v atomicMutex[int]
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
		var v atomicRWMutex[int]
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
	b.Run("atomic value", func(b *testing.B) {
		var v atomicValue[int]
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
	b.Run("atomic custom mutex", func(b *testing.B) {
		var v atomicCustomMutex[int]
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

type atomicMutex[T comparable] struct {
	value T
	lock  sync.Mutex
}

func (a *atomicMutex[T]) Load() T {
	a.lock.Lock()
	v := a.value
	a.lock.Unlock()
	return v
}

func (a *atomicMutex[T]) Store(v T) {
	a.lock.Lock()
	a.value = v
	a.lock.Unlock()
}

type atomicRWMutex[T comparable] struct {
	value T
	lock  sync.RWMutex
}

func (a *atomicRWMutex[T]) Load() T {
	a.lock.RLock()
	v := a.value
	a.lock.RUnlock()
	return v
}

func (a *atomicRWMutex[T]) Store(v T) {
	a.lock.Lock()
	a.value = v
	a.lock.Unlock()
}

type atomicValue[T comparable] struct {
	// Note: it would also be possible to use atomic.Pointer to avoid the panics from
	// atomic.Value but we won't be able to do the "swapIfNotEqual" as a single transaction anyway
	atomic.Value
}

func (a *atomicValue[T]) Load() (zero T) {
	v := a.Value.Load()
	if v == nil {
		return zero
	}
	return v.(T)
}

func (a *atomicValue[T]) Store(v T) {
	a.Value.Store(v)
}

type atomicCustomMutex[T comparable] struct {
	value T
	mutex int32
}

func (a *atomicCustomMutex[T]) lock() {
	for atomic.CompareAndSwapInt32(&a.mutex, 0, 1) {
	}
}

func (a *atomicCustomMutex[T]) unlock() {
	for atomic.CompareAndSwapInt32(&a.mutex, 1, 0) {
	}
}

func (a *atomicCustomMutex[T]) Load() T {
	a.lock()
	v := a.value
	a.unlock()
	return v
}

func (a *atomicCustomMutex[T]) Store(v T) {
	a.lock()
	a.value = v
	a.unlock()
}
