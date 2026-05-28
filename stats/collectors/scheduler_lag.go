package collectors

import (
	"context"
	"sync"
	"time"

	"github.com/rudderlabs/rudder-go-kit/stats"
)

// SchedulerLagCollector measures Go runtime scheduler lag as a proxy for CPU pressure.
// It fires a timer every interval, records how late the actual firing was, and tracks
// the running maximum. The maximum resets to zero each time Collect is called.
type SchedulerLagCollector struct {
	interval   time.Duration
	maxLagMsMu sync.Mutex
	maxLagMs   int64
}

// NewSchedulerLagCollector creates a new SchedulerLagCollector.
func NewSchedulerLagCollector(interval time.Duration) *SchedulerLagCollector {
	return &SchedulerLagCollector{interval: interval}
}

// Run measures scheduler lag continuously until ctx is cancelled.
// It should be started in its own goroutine.
func (s *SchedulerLagCollector) Run(ctx context.Context) {
	for {
		start := time.Now()
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.interval):
		}
		if lagMs := max(time.Since(start)-s.interval, 0).Milliseconds(); lagMs > 0 {
			s.maxLagMsMu.Lock()
			if lagMs > s.maxLagMs {
				s.maxLagMs = lagMs
			}
			s.maxLagMsMu.Unlock()
		}
	}
}

func (s *SchedulerLagCollector) Collect(gaugeFunc func(key string, tags stats.Tags, val uint64)) {
	s.maxLagMsMu.Lock()
	v := s.maxLagMs
	s.maxLagMs = 0
	s.maxLagMsMu.Unlock()
	gaugeFunc("cpu.scheduler_lag_ms", nil, uint64(v))
}

func (s *SchedulerLagCollector) Zero(gaugeFunc func(key string, tags stats.Tags, val uint64)) {
	s.maxLagMsMu.Lock()
	s.maxLagMs = 0
	s.maxLagMsMu.Unlock()
	gaugeFunc("cpu.scheduler_lag_ms", nil, 0)
}

func (s *SchedulerLagCollector) ID() string {
	return "scheduler_lag"
}
