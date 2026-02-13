package sync_test

import (
	gsync "sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/sync"
)

func TestPartitionLocker(t *testing.T) {
	t.Run("Lock and Unlock different partitions at the same time", func(t *testing.T) {
		locker := sync.NewPartitionLocker()
		locker.Lock("id1")
		locker.Lock("id2")

		locker.Unlock("id1")
		locker.Unlock("id2")
	})

	t.Run("Concurrent locks", func(t *testing.T) {
		locker := sync.NewPartitionLocker()
		var wg gsync.WaitGroup
		var counter int
		goroutines := 1000
		for range goroutines {
			wg.Go(func() {
				locker.Lock("id")
				counter = counter + 1
				time.Sleep(1 * time.Millisecond)
				locker.Unlock("id")
			})
		}
		wg.Wait()
		require.Equalf(t, goroutines, counter, "it should have incremented the counter %d times", goroutines)
	})

	t.Run("Try to lock the same partition twice", func(t *testing.T) {
		type l struct {
			locker sync.PartitionLocker
		}
		var s l
		s.locker = *sync.NewPartitionLocker()
		var locks atomic.Int64
		const id = "id"
		s.locker.Lock(id)
		go func() {
			s.locker.Lock(id)
			locks.Store(locks.Add(1))
			s.locker.Unlock(id)
		}()
		require.Never(t, func() bool { return locks.Load() == 1 }, 100*time.Millisecond, 1*time.Millisecond)
		s.locker.Unlock(id)
		require.Eventually(t, func() bool { return locks.Load() == 1 }, 100*time.Millisecond, 1*time.Millisecond)
	})
}
