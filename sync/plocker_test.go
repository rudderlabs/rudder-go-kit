package sync_test

import (
	gsync "sync"
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
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				locker.Lock("id")
				counter = counter + 1
				time.Sleep(1 * time.Millisecond)
				locker.Unlock("id")
			}()
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
		var locks int
		const id = "id"
		s.locker.Lock(id)
		go func() {
			s.locker.Lock(id)
			locks = locks + 1
			s.locker.Unlock(id)
		}()
		require.Never(t, func() bool { return locks == 1 }, 100*time.Millisecond, 1*time.Millisecond)
		s.locker.Unlock(id)
		require.Eventually(t, func() bool { return locks == 1 }, 100*time.Millisecond, 1*time.Millisecond)
	})
}
