package sync

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/queue"
	"github.com/rudderlabs/rudder-go-kit/stats"
)

// LimiterPriorityValue defines the priority values supported by Limiter.
// Greater priority value means higher priority
type LimiterPriorityValue int

const (
	_ LimiterPriorityValue = iota
	// LimiterPriorityValueLow Priority....
	LimiterPriorityValueLow
	// LimiterPriorityValueMedium Priority....
	LimiterPriorityValueMedium
	// LimiterPriorityValueMediumHigh Priority....
	LimiterPriorityValueMediumHigh
	// LimiterPriorityValueHigh Priority.....
	LimiterPriorityValueHigh
)

// Limiter limits the number of concurrent operations that can be performed
type Limiter interface {
	// Do executes the function f, but only if there are available slots.
	// Otherwise blocks until a slot becomes available
	Do(key string, f func())

	// DoWithPriority executes the function f, but only if there are available slots.
	// Otherwise blocks until a slot becomes available, respecting the priority
	DoWithPriority(key string, priority LimiterPriorityValue, f func())

	// Begin starts a new operation, blocking until a slot becomes available.
	// Caller is expected to call the returned function to end the operation, otherwise
	// the slot will be reserved indefinitely. End can be called multiple times without any side effects
	Begin(key string) (end func())

	// BeginWithPriority starts a new operation, blocking until a slot becomes available, respecting the priority.
	// Caller is expected to call the returned function to end the operation, otherwise
	// the slot will be reserved indefinitely. End can be called multiple times without any side effects
	BeginWithPriority(key string, priority LimiterPriorityValue) (end func())

	// BeginWithSleep starts a new operation with sleep capability, blocking until a slot becomes available.
	// Returns a LimiterExecutionHandle that can be used to sleep during the operation without affecting working time metrics or blocking other waiting operations
	BeginWithSleep(key string) LimiterExecution

	// BeginWithSleepAndPriority starts a new operation with sleep capability and priority, blocking until a slot becomes available.
	// Returns a LimiterExecutionHandle that can be used to sleep during the operation without affecting working time metrics or blocking other waiting operations
	BeginWithSleepAndPriority(key string, priority LimiterPriorityValue) LimiterExecution
}

// LimiterExecution represents a function that can end a limiter operation and optionally sleep
type LimiterExecution interface {
	// End the limiter operation (same as calling the function directly)
	End()
	// Sleep pauses the working timer for the specified duration
	Sleep(ctx context.Context, duration time.Duration) error
}

var WithLimiterStatsTriggerFunc = func(triggerFunc func() <-chan time.Time) func(*limiter) {
	return func(l *limiter) {
		l.stats.triggerFunc = triggerFunc
	}
}

var WithLimiterDynamicPeriod = func(dynamicPeriod time.Duration) func(*limiter) {
	return func(l *limiter) {
		l.dynamicPeriod = dynamicPeriod
	}
}

var WithLimiterTags = func(tags stats.Tags) func(*limiter) {
	return func(l *limiter) {
		l.tags = tags
	}
}

// NewLimiter creates a new limiter
func NewLimiter(ctx context.Context, wg *sync.WaitGroup, name string, limit int, statsf stats.Stats, opts ...func(*limiter)) Limiter {
	return NewReloadableLimiter(ctx, wg, name, config.SingleValueLoader(limit), statsf, opts...)
}

// NewReloadableLimiter creates a new limiter with a hot-reloadable limit, i.e. the limit can be changed at runtime
func NewReloadableLimiter(ctx context.Context, wg *sync.WaitGroup, name string, limit config.ValueLoader[int], statsf stats.Stats, opts ...func(*limiter)) Limiter {
	if limit.Load() <= 0 {
		panic(fmt.Errorf("limit for %q must be greater than 0", name))
	}
	l := &limiter{
		name:     name,
		limit:    limit,
		tags:     stats.Tags{},
		waitList: make(queue.PriorityQueue[chan struct{}], 0),
	}
	heap.Init(&l.waitList)
	l.stats.triggerFunc = func() <-chan time.Time {
		return time.After(15 * time.Second)
	}
	for _, opt := range opts {
		opt(l)
	}

	l.stats.stat = statsf
	l.stats.waitGauge = statsf.NewTaggedStat(name+"_limiter_waiting_routines", stats.GaugeType, l.tags)
	l.stats.activeGauge = statsf.NewTaggedStat(name+"_limiter_active_routines", stats.GaugeType, l.tags)
	l.stats.availabilityGauge = statsf.NewTaggedStat(name+"_limiter_availability", stats.GaugeType, l.tags)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-l.stats.triggerFunc():
			}
			l.mu.Lock()
			l.stats.activeGauge.Gauge(l.count)
			l.stats.waitGauge.Gauge(len(l.waitList))
			limit := l.limit.Load()
			availability := float64(limit-l.count) / float64(limit)
			l.stats.availabilityGauge.Gauge(availability)
			l.mu.Unlock()
		}
	}()
	return l
}

type limiter struct {
	name          string
	limit         config.ValueLoader[int]
	tags          stats.Tags
	dynamicPeriod time.Duration

	mu       sync.Mutex // protects count and waitList below
	count    int
	waitList queue.PriorityQueue[chan struct{}]

	stats struct {
		triggerFunc       func() <-chan time.Time
		stat              stats.Stats
		waitGauge         stats.Measurement // gauge showing number of operations waiting in the queue
		activeGauge       stats.Measurement // gauge showing active number of operations
		availabilityGauge stats.Measurement // gauge showing availability percentage of limiter (0.0 to 1.0)
	}
}

func (l *limiter) Do(key string, f func()) {
	l.DoWithPriority(key, LimiterPriorityValueLow, f)
}

func (l *limiter) DoWithPriority(key string, priority LimiterPriorityValue, f func()) {
	defer l.BeginWithPriority(key, priority)()
	f()
}

func (l *limiter) Begin(key string) (end func()) {
	return l.BeginWithPriority(key, LimiterPriorityValueLow)
}

func (l *limiter) BeginWithPriority(key string, priority LimiterPriorityValue) (end func()) {
	return l.BeginWithSleepAndPriority(key, priority).End
}

func (l *limiter) BeginWithSleep(key string) LimiterExecution {
	return l.BeginWithSleepAndPriority(key, LimiterPriorityValueLow)
}

func (l *limiter) BeginWithSleepAndPriority(key string, priority LimiterPriorityValue) LimiterExecution {
	tags := lo.Assign(l.tags, stats.Tags{"key": key})
	start := time.Now()
	l.wait(priority)
	totalWaiting := time.Since(start)
	return &limiterExecution{
		p:            priority,
		l:            l,
		tags:         tags,
		workingStart: time.Now(),
		totalWaiting: totalWaiting,
	}
}

// wait until a slot becomes available
func (l *limiter) wait(priority LimiterPriorityValue) {
	l.mu.Lock()
	limit := l.limit.Load()
	if limit <= 0 {
		l.mu.Unlock()
		panic(fmt.Errorf("limit for %q must be greater than 0", l.name))
	}
	if l.count < limit {
		l.count++
		l.mu.Unlock()
		return
	}
	w := &queue.Item[chan struct{}]{
		Priority: int(priority),
		Value:    make(chan struct{}),
	}
	heap.Push(&l.waitList, w)
	l.mu.Unlock()

	// no dynamic priority
	if l.dynamicPeriod == 0 || priority == LimiterPriorityValueHigh {
		<-w.Value
		return
	}

	// dynamic priority (increment priority every dynamicPeriod)
	ticker := time.NewTicker(l.dynamicPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-w.Value:
			ticker.Stop()
			return
		case <-ticker.C:
			if w.Priority < int(LimiterPriorityValueHigh) {
				l.mu.Lock()
				l.waitList.Update(w, w.Priority+1)
				l.mu.Unlock()
			} else {
				ticker.Stop()
				<-w.Value
				return
			}
		}
	}
}

func (l *limiter) end() {
	l.mu.Lock()
	l.count--
	if len(l.waitList) == 0 {
		l.mu.Unlock()
		return
	}
	next := heap.Pop(&l.waitList).(*queue.Item[chan struct{}])
	l.count++
	l.mu.Unlock()
	next.Value <- struct{}{}
	close(next.Value)
}

type limiterExecution struct {
	p             LimiterPriorityValue
	l             *limiter
	tags          stats.Tags
	workingStart  time.Time
	totalWorking  time.Duration
	totalWaiting  time.Duration
	totalSleeping time.Duration
	ended         bool
	endOnce       sync.Once
}

func (le *limiterExecution) End() {
	le.endOnce.Do(func() {
		le.l.end()
		le.totalWorking += time.Since(le.workingStart)
		le.l.stats.stat.NewTaggedStat(le.l.name+"_limiter_waiting", stats.TimerType, le.tags).SendTiming(le.totalWaiting)
		le.l.stats.stat.NewTaggedStat(le.l.name+"_limiter_sleeping", stats.TimerType, le.tags).SendTiming(le.totalSleeping)
		le.l.stats.stat.NewTaggedStat(le.l.name+"_limiter_working", stats.TimerType, le.tags).SendTiming(le.totalWorking)
		le.ended = true
	})
}

func (le *limiterExecution) Sleep(ctx context.Context, duration time.Duration) error {
	if le.ended {
		return fmt.Errorf("limiter execution has ended, sleep not allowed")
	}
	start := time.Now()
	// before sleeping we need to release the slot
	// so that other waiting operations can proceed
	le.l.end()
	le.totalWorking += time.Since(le.workingStart)

	defer func() {
		le.totalSleeping += time.Since(start)
		// reacquire the slot after sleeping
		waitStart := time.Now()
		le.l.wait(le.p)
		le.totalWaiting += time.Since(waitStart)
		le.workingStart = time.Now()
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return nil
	}
}
