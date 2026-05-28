package collectors_test

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/stats"
	"github.com/rudderlabs/rudder-go-kit/stats/collectors"
)

func TestSchedulerLagCollector(t *testing.T) {
	collect := func(c *collectors.SchedulerLagCollector) uint64 {
		var val uint64
		c.Collect(func(_ string, _ stats.Tags, v uint64) { val = v })
		return val
	}
	startRun := func(t *testing.T, c *collectors.SchedulerLagCollector) {
		ctx, cancel := context.WithCancel(t.Context())
		var wg sync.WaitGroup
		wg.Go(func() { c.Run(ctx) })
		t.Cleanup(func() { cancel(); wg.Wait() })
	}

	t.Run("id", func(t *testing.T) {
		c := collectors.NewSchedulerLagCollector(10 * time.Millisecond)
		require.Equal(t, "scheduler_lag", c.ID())
	})

	t.Run("zero before run", func(t *testing.T) {
		c := collectors.NewSchedulerLagCollector(10 * time.Millisecond)
		require.Equal(t, uint64(0), collect(c))
	})

	t.Run("collect resets to zero", func(t *testing.T) {
		c := collectors.NewSchedulerLagCollector(10 * time.Millisecond)
		startRun(t, c)

		require.Eventually(t, func() bool {
			collect(c) // drain whatever was accumulated since startup
			return collect(c) == 0
		}, 5*time.Second, 50*time.Millisecond, "collect should reset max to zero")
	})

	t.Run("zero clears accumulated lag", func(t *testing.T) {
		c := collectors.NewSchedulerLagCollector(10 * time.Millisecond)
		startRun(t, c)

		time.Sleep(50 * time.Millisecond)

		var zeroed uint64
		c.Zero(func(_ string, _ stats.Tags, v uint64) { zeroed = v })
		require.Equal(t, uint64(0), zeroed)
		require.Equal(t, uint64(0), collect(c), "collect after Zero should also be zero")
	})

	t.Run("captures non-zero lag", func(t *testing.T) {
		c := collectors.NewSchedulerLagCollector(1 * time.Millisecond)
		startRun(t, c)

		// Saturate all OS threads so the timer goroutine experiences real scheduling delay.
		var busyWg sync.WaitGroup
		for range runtime.NumCPU() {
			busyWg.Go(func() {
				ctx := t.Context()
				for ctx.Err() == nil {
					var n uint64
					for range 10_000 {
						n++
					}
					_ = n
				}
			})
		}
		t.Cleanup(busyWg.Wait)

		require.Eventually(t, func() bool {
			return collect(c) > 0
		}, 5*time.Second, 10*time.Millisecond, "expected at least one tick with measurable lag")
	})

	t.Run("lag under load with varying MAXPROCS", func(t *testing.T) {
		numCPU := runtime.NumCPU()

		dedup := func(candidates []int) []int {
			seen := make(map[int]struct{})
			var out []int
			for _, v := range candidates {
				if _, ok := seen[v]; !ok {
					seen[v] = struct{}{}
					out = append(out, v)
				}
			}
			return out
		}

		maxProcsValues := dedup([]int{1, 2, numCPU})
		busyCounts := dedup([]int{1, numCPU / 2, numCPU, numCPU * 2})

		for _, maxProcs := range maxProcsValues {
			t.Run(fmt.Sprintf("GOMAXPROCS_%d", maxProcs), func(t *testing.T) {
				prev := runtime.GOMAXPROCS(maxProcs)
				t.Cleanup(func() { runtime.GOMAXPROCS(prev) })

				for _, numBusy := range busyCounts {
					t.Run(fmt.Sprintf("busy_%d", numBusy), func(t *testing.T) {
						c := collectors.NewSchedulerLagCollector(1 * time.Millisecond)
						startRun(t, c)

						var busyWg sync.WaitGroup
						for range numBusy {
							busyWg.Go(func() {
								ctx := t.Context()
								for ctx.Err() == nil {
									var n uint64
									for range 10_000 {
										n++
									}
									_ = n
								}
							})
						}
						t.Cleanup(busyWg.Wait)

						time.Sleep(200 * time.Millisecond)
						lag := collect(c)
						t.Logf("GOMAXPROCS=%d busy=%d num_cpu=%d max_lag_ms=%d",
							maxProcs, numBusy, numCPU, lag)
					})
				}
			})
		}
	})

	t.Run("metric key", func(t *testing.T) {
		c := collectors.NewSchedulerLagCollector(10 * time.Millisecond)
		var key string
		c.Collect(func(k string, _ stats.Tags, _ uint64) { key = k })
		require.Equal(t, "cpu.scheduler_lag_ms", key)
	})
}
