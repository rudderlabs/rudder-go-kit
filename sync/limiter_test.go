package sync_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/memstats"
	miscsync "github.com/rudderlabs/rudder-go-kit/sync"
)

func TestLimiter(t *testing.T) {
	t.Run("without priority", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		ms, err := memstats.New()
		require.NoError(t, err)

		statsTriggerCh := make(chan time.Time)
		triggerFn := func() <-chan time.Time {
			return statsTriggerCh
		}

		limiter := miscsync.NewLimiter(ctx, &wg, "test", 1, ms, miscsync.WithLimiterStatsTriggerFunc(triggerFn))
		var counter int
		statsTriggerCh <- time.Now()
		statsTriggerCh <- time.Now()

		require.NotNil(t, ms.Get("test_limiter_active_routines", nil))
		require.EqualValues(t, 0, ms.Get("test_limiter_active_routines", nil).LastValue(), "shouldn't have any active")

		require.NotNil(t, ms.Get("test_limiter_availability", nil))
		require.EqualValues(t, 1, ms.Get("test_limiter_availability", nil).LastValue(), "should be available")

		require.Nil(t, ms.Get("test_limiter_waiting", nil))
		require.Nil(t, ms.Get("test_limiter_sleeping", nil))
		require.Nil(t, ms.Get("test_limiter_working", nil))

		for i := range 100 {
			wg.Add(1)
			key := strconv.Itoa(i)
			go func() {
				defer wg.Done()
				limiter.Do(key, func() {
					counter++ // since the limiter's limit is 1, we shouldn't need an atomic counter
				})
			}()
		}

		cancel()
		wg.Wait()

		require.EqualValues(t, 100, counter, "counter should be 100")

		select {
		case statsTriggerCh <- time.Now():
			require.Fail(t, "shouldn't be listening to triggerCh anymore")
		default:
		}
		for i := range 100 {
			wa := ms.Get("test_limiter_waiting", map[string]string{"key": strconv.Itoa(i)})
			require.NotNil(t, wa)
			require.Lenf(t, wa.Durations(), 1, "should have recorded 1 waiting timer duration for key %d", i)

			sl := ms.Get("test_limiter_sleeping", map[string]string{"key": strconv.Itoa(i)})
			require.NotNil(t, sl)
			require.Lenf(t, sl.Durations(), 1, "should have recorded 1 sleeping timer duration for key %d", i)
			require.EqualValues(t, 0, sl.LastValue(), "sleeping time should be 0 since we didn't sleep")

			wo := ms.Get("test_limiter_working", map[string]string{"key": strconv.Itoa(i)})
			require.NotNil(t, wo)
			require.Lenf(t, wo.Durations(), 1, "should have recorded 1 workingtimer duration for key %d", i)
		}
	})

	t.Run("with priority", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		limiter := miscsync.NewLimiter(ctx, &wg, "test", 1, stats.NOP)
		var counterLow int
		var counterHigh int
		sleepTime := 100 * time.Microsecond
		for i := range 1000 {
			wg.Add(1)
			key := strconv.Itoa(i)
			go func() {
				defer wg.Done()
				limiter.DoWithPriority(key, miscsync.LimiterPriorityValueHigh, func() {
					time.Sleep(sleepTime)
					counterHigh++ // since the limiter's limit is 1, we shouldn't need an atomic counter
					require.Equal(t, 0, counterLow, "counterLow should be 0")
				})
			}()
		}

		time.Sleep(10 * sleepTime)
		for i := range 1000 {
			wg.Add(1)
			key := strconv.Itoa(i)
			go func() {
				defer wg.Done()
				limiter.DoWithPriority(key, miscsync.LimiterPriorityValueLow, func() {
					counterLow++ // since the limiter's limit is 1, we shouldn't need an atomic counter
				})
			}()
		}

		cancel()
		wg.Wait()

		require.EqualValues(t, 1000, counterHigh, "high priority counter should be 1000")
		require.EqualValues(t, 1000, counterLow, "low priority counter should be 1000")
	})

	t.Run("with dynamic priority", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		sleepTime := 1 * time.Millisecond
		limiter := miscsync.NewLimiter(ctx, &wg, "test", 1, stats.NOP, miscsync.WithLimiterDynamicPeriod(sleepTime/100))
		var counterLow int
		var counterHigh int

		var dynamicPriorityVerified bool
		for i := range 1000 {
			wg.Add(1)
			key := strconv.Itoa(i)
			go func() {
				defer wg.Done()
				limiter.DoWithPriority(key, miscsync.LimiterPriorityValueHigh, func() {
					time.Sleep(sleepTime)
					counterHigh++ // since the limiter's limit is 1, we shouldn't need an atomic counter
					if counterLow > 0 {
						dynamicPriorityVerified = true
					}
				})
			}()
		}

		for i := range 10 {
			wg.Add(1)
			key := strconv.Itoa(i)
			go func() {
				defer wg.Done()
				limiter.DoWithPriority(key, miscsync.LimiterPriorityValueLow, func() {
					counterLow++ // since the limiter's limit is 1, we shouldn't need an atomic counter
				})
			}()
		}

		cancel()
		wg.Wait()

		require.True(t, dynamicPriorityVerified, "dynamic priority should have been verified")
		require.EqualValues(t, 1000, counterHigh, "high priority counter should be 1000")
		require.EqualValues(t, 10, counterLow, "low priority counter should be 10")
	})

	t.Run("with sleep", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		ms, err := memstats.New()
		require.NoError(t, err)

		sleepTime := 10 * time.Millisecond
		limiter := miscsync.NewLimiter(ctx, &wg, "test", 1, ms, miscsync.WithLimiterDynamicPeriod(sleepTime/100))
		var counter int
		var sleepVerified bool
		for i := range 1000 {
			wg.Add(1)
			key := strconv.Itoa(i)
			go func() {
				defer wg.Done()
				le := limiter.BeginWithSleep(key)
				defer le.End()
				if !sleepVerified && counter == 0 && i > 0 {
					sleepVerified = true
				}
				require.NoError(t, le.Sleep(context.Background(), sleepTime))
				counter++ // since the limiter's limit is 1, we shouldn't need an atomic counter
			}()
		}
		cancel()
		wg.Wait()
		require.True(t, sleepVerified, "sleep should have been verified")
		for i := range 1000 {
			key := strconv.Itoa(i)
			sl := ms.Get("test_limiter_sleeping", map[string]string{"key": key})
			require.NotNil(t, sl)
			require.Lenf(t, sl.Durations(), 1, "should have recorded 1 sleeping timer duration for key %d", i)
			require.GreaterOrEqual(t, sl.LastDuration(), sleepTime, "sleeping time should be greater than or equal to 1ms")
			require.Less(t, sl.LastDuration(), sleepTime*2, "sleeping time should be greater than or equal to 1ms")
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		require.Panics(t, func() {
			var wg sync.WaitGroup
			_ = miscsync.NewLimiter(context.Background(), &wg, "zerolimit", 0, stats.Default)
		})

		require.Panics(t, func() {
			var wg sync.WaitGroup
			_ = miscsync.NewLimiter(context.Background(), &wg, "negativelimit", -1, stats.Default)
		})
	})
}
