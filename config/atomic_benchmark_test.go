package config

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

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
	b.Run("atomic", func(b *testing.B) {
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
