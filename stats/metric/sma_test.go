package metric_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stats/metric"
)

func TestSimpleMovingAverage(t *testing.T) {
	t.Run("Constructor and Basic Operations", func(t *testing.T) {
		t.Run("creates simple moving average with correct initial state", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(5)
			require.Equal(t, 0.0, sa.Load())
		})

		t.Run("single observation", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(3)
			sa.Observe(10.0)
			require.Equal(t, 10.0, sa.Load())
		})

		t.Run("multiple observations within window", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(3)
			sa.Observe(10.0)
			sa.Observe(20.0)
			sa.Observe(30.0)

			// Average of [10, 20, 30] = 60/3 = 20
			require.Equal(t, 20.0, sa.Load())
		})
	})

	t.Run("Window Behavior", func(t *testing.T) {
		t.Run("sliding window replaces oldest values", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(3)
			sa.Observe(10.0)
			sa.Observe(20.0)
			sa.Observe(30.0)
			sa.Observe(40.0) // This should replace the 10.0

			// Average of [20, 30, 40] = 90/3 = 30
			require.Equal(t, 30.0, sa.Load())
		})

		t.Run("continues sliding after full rotation", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(2)
			sa.Observe(10.0)
			sa.Observe(20.0)
			sa.Observe(30.0) // replaces 10.0
			sa.Observe(40.0) // replaces 20.0
			sa.Observe(50.0) // replaces 30.0

			// Average of [40, 50] = 90/2 = 45
			require.Equal(t, 45.0, sa.Load())
		})

		t.Run("size 1 window", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(1)
			sa.Observe(10.0)
			require.Equal(t, 10.0, sa.Load())

			sa.Observe(20.0)
			require.Equal(t, 20.0, sa.Load())

			sa.Observe(30.0)
			require.Equal(t, 30.0, sa.Load())
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("zero values", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(3)
			sa.Observe(0.0)
			sa.Observe(0.0)
			sa.Observe(0.0)
			require.Equal(t, 0.0, sa.Load())
		})

		t.Run("negative values", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(2)
			sa.Observe(-10.0)
			sa.Observe(-20.0)
			require.Equal(t, -15.0, sa.Load())
		})

		t.Run("fractional values", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(3)
			sa.Observe(1.5)
			sa.Observe(2.5)
			sa.Observe(3.0)
			require.InDelta(t, 2.333333, sa.Load(), 0.000001)
		})

		t.Run("very large values", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(2)
			sa.Observe(1e9)
			sa.Observe(2e9)
			require.Equal(t, 1.5e9, sa.Load())
		})
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		t.Run("concurrent observations and loads", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(100)
			var wg sync.WaitGroup

			// Start multiple goroutines observing values
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func(val float64) {
					defer wg.Done()
					sa.Observe(val)
				}(float64(i))
			}

			// Start multiple goroutines reading values
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = sa.Load() // Just ensure it doesn't panic
				}()
			}

			wg.Wait()

			// Verify final state is valid
			avg := sa.Load()
			require.True(t, avg >= 0) // Should be positive since we added 0-49
		})
	})

	t.Run("Performance and Accuracy", func(t *testing.T) {
		t.Run("maintains accuracy over many operations", func(t *testing.T) {
			sa := metric.NewSimpleMovingAverage(10)

			// Add 10 values: 1, 2, 3, ..., 10
			for i := 1; i <= 10; i++ {
				sa.Observe(float64(i))
			}

			// Average should be (1+2+...+10)/10 = 55/10 = 5.5
			require.Equal(t, 5.5, sa.Load())

			// Add more values to test sliding
			sa.Observe(11.0)                 // replaces 1, new window: [2,3,4,5,6,7,8,9,10,11]
			require.Equal(t, 6.5, sa.Load()) // (2+3+...+11)/10 = 65/10 = 6.5
		})
	})
}
