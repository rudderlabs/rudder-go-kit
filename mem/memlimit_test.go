package mem

import (
	"context"
	"runtime/debug"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

func TestSetMemoryLimit(t *testing.T) {
	t.Run("uses default 90 percent", func(t *testing.T) {
		ctx := context.Background()

		SetMemoryLimit(ctx)

		currentLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, currentLimit, int64(0), "memory limit should be set to a positive value")

		stat, err := Get()
		require.NoError(t, err)
		expectedLimit := int64(stat.Total * 90 / 100)
		require.Equal(t, expectedLimit, currentLimit, "should use default 90% of total memory")
	})

	t.Run("sets memory limit with custom percentage", func(t *testing.T) {
		ctx := context.Background()

		SetMemoryLimit(ctx, SetWithPercentage(75))

		currentLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, currentLimit, int64(0), "memory limit should be set to a positive value")

		stat, err := Get()
		require.NoError(t, err)
		expectedLimit := int64(stat.Total * 75 / 100)
		require.Equal(t, expectedLimit, currentLimit, "should use 75% of total memory")
	})

	t.Run("sets memory limit with percentage loader", func(t *testing.T) {
		ctx := context.Background()
		limitPercent := config.GetReloadableIntVar(60, 1, "memoryLimitPercent")

		SetMemoryLimit(ctx, SetWithPercentageLoader(limitPercent))

		currentLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, currentLimit, int64(0), "memory limit should be set to a positive value")

		stat, err := Get()
		require.NoError(t, err)
		expectedLimit := int64(stat.Total * 60 / 100)
		require.Equal(t, expectedLimit, currentLimit, "should use 60% of total memory")
	})

	t.Run("handles zero percentage", func(t *testing.T) {
		ctx := context.Background()

		SetMemoryLimit(ctx, SetWithPercentage(0))

		currentLimit := debug.SetMemoryLimit(-1)
		require.GreaterOrEqual(t, currentLimit, int64(0), "memory limit should be set")
	})

	t.Run("handles 100 percentage", func(t *testing.T) {
		ctx := context.Background()

		SetMemoryLimit(ctx, SetWithPercentage(100))

		currentLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, currentLimit, int64(0), "memory limit should be set to a positive value")

		stat, err := Get()
		require.NoError(t, err)
		expectedLimit := int64(stat.Total)
		require.Equal(t, expectedLimit, currentLimit, "should use 100% of total memory")
	})

	t.Run("uses custom logger", func(t *testing.T) {
		ctx := context.Background()
		customLog := logger.NOP

		SetMemoryLimit(ctx, SetWithPercentage(50), SetWithLogger(customLog))

		currentLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, currentLimit, int64(0), "memory limit should be set with custom logger")
	})
}

func TestWatchMemoryLimit(t *testing.T) {
	t.Run("uses default 90 percent", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		wg := &sync.WaitGroup{}

		WatchMemoryLimit(ctx, wg, WatchWithInterval(100*time.Millisecond))

		// Wait for initial value to be set
		time.Sleep(200 * time.Millisecond)
		limit := debug.SetMemoryLimit(-1)
		require.Greater(t, limit, int64(0), "memory limit should be set")

		stat, err := Get()
		require.NoError(t, err)
		expectedLimit := int64(stat.Total * 90 / 100)
		require.Equal(t, expectedLimit, limit, "should use default 90% of total memory")

		cancel()
		wg.Wait()
	})

	t.Run("updates memory limit when percentage changes", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		limitPercent := config.NewMockValueLoader(50)
		wg := &sync.WaitGroup{}

		WatchMemoryLimit(ctx, wg,
			WatchWithPercentageLoader(limitPercent),
			WatchWithInterval(500*time.Millisecond),
		)

		// Wait for initial value to be set
		time.Sleep(100 * time.Millisecond)
		firstLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, firstLimit, int64(0), "initial memory limit should be set")

		// Change the percentage
		limitPercent.Set(60)

		// Wait for the watcher to detect the change
		require.Eventually(t, func() bool {
			currentLimit := debug.SetMemoryLimit(-1)
			return currentLimit != firstLimit
		}, 2*time.Second, 100*time.Millisecond, "memory limit should update when percentage changes")

		cancel()
		wg.Wait()
	})

	t.Run("uses custom percentage", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		wg := &sync.WaitGroup{}

		WatchMemoryLimit(ctx, wg,
			WatchWithPercentage(75),
			WatchWithInterval(100*time.Millisecond),
		)

		// Wait for value to be set
		time.Sleep(200 * time.Millisecond)
		limit := debug.SetMemoryLimit(-1)

		stat, err := Get()
		require.NoError(t, err)
		expectedLimit := int64(stat.Total * 75 / 100)
		require.Equal(t, expectedLimit, limit, "should use custom 75% of total memory")

		cancel()
		wg.Wait()
	})

	t.Run("stops watching when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		wg := &sync.WaitGroup{}

		WatchMemoryLimit(ctx, wg,
			WatchWithPercentage(70),
			WatchWithInterval(100*time.Millisecond),
		)

		// Wait a bit for goroutines to start
		time.Sleep(50 * time.Millisecond)

		cancel()

		// Verify goroutines complete
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - goroutines completed
		case <-time.After(500 * time.Millisecond):
			t.Fatal("goroutines did not complete after context cancellation")
		}
	})

	t.Run("runs without errors when percentage unchanged", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		wg := &sync.WaitGroup{}

		WatchMemoryLimit(ctx, wg,
			WatchWithPercentage(75),
			WatchWithInterval(200*time.Millisecond),
		)

		// Wait for initial value to be set
		time.Sleep(100 * time.Millisecond)
		initialLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, initialLimit, int64(0), "initial memory limit should be set")

		// Wait through multiple polling cycles
		time.Sleep(1 * time.Second)

		// Verify it's still running and limit is reasonable
		currentLimit := debug.SetMemoryLimit(-1)
		require.Greater(t, currentLimit, int64(0), "memory limit should still be set")

		cancel()
		wg.Wait()
	})

	t.Run("uses custom logger", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()

		wg := &sync.WaitGroup{}
		customLog := logger.NOP

		WatchMemoryLimit(ctx, wg,
			WatchWithPercentage(60),
			WatchWithInterval(100*time.Millisecond),
			WatchWithLogger(customLog),
		)

		// Wait for value to be set
		time.Sleep(200 * time.Millisecond)
		limit := debug.SetMemoryLimit(-1)
		require.Greater(t, limit, int64(0), "memory limit should be set with custom logger")

		cancel()
		wg.Wait()
	})
}

func TestCalculateMemoryLimit(t *testing.T) {
	t.Run("calculates correct memory limit", func(t *testing.T) {
		limitPercent := config.NewMockValueLoader(50)

		memoryLimit, err := calculateMemoryLimit(limitPercent)
		require.NoError(t, err)
		require.Greater(t, memoryLimit, int64(0), "memory limit should be positive")

		stat, err := Get()
		require.NoError(t, err)
		expectedLimit := int64(stat.Total / 2)
		require.Equal(t, expectedLimit, memoryLimit, "should calculate 50% of total memory")
	})

	t.Run("handles different percentages", func(t *testing.T) {
		stat, err := Get()
		require.NoError(t, err)

		testCases := []struct {
			percent  int
			expected int64
		}{
			{percent: 10, expected: int64(stat.Total * 10 / 100)},
			{percent: 25, expected: int64(stat.Total * 25 / 100)},
			{percent: 90, expected: int64(stat.Total * 90 / 100)},
			{percent: 100, expected: int64(stat.Total)},
		}

		for _, tc := range testCases {
			limitPercent := config.NewMockValueLoader(tc.percent)
			memoryLimit, err := calculateMemoryLimit(limitPercent)
			require.NoError(t, err)
			require.Equal(t, tc.expected, memoryLimit)
		}
	})
}
