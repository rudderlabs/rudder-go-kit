package throttling

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/rudderlabs/rudder-go-kit/stats"
)

/*
TODOs:
* guard against concurrency? according to benchmarks, Redis performs better if we have no more than 16 routines
  * see https://github.com/rudderlabs/redis-throttling-playground/blob/main/Benchmarks.md#best-concurrency-setting-with-sortedset---save-1-1-and---appendonly-yes
*/

var (
	//go:embed lua/gcra.lua
	gcraLua         string
	gcraRedisScript *redis.Script
	//go:embed lua/sortedset.lua
	sortedSetLua    string
	sortedSetScript *redis.Script
)

func init() {
	gcraRedisScript = redis.NewScript(gcraLua)
	sortedSetScript = redis.NewScript(sortedSetLua)
}

type redisSpeaker interface {
	redis.Scripter
	redisSortedSetRemover
}

type statsCollector interface {
	NewTaggedStat(name, statType string, tags stats.Tags) stats.Measurement
}

type Limiter struct {
	// for Redis configurations
	// a default redisSpeaker should always be provided for Redis configurations
	redisSpeaker redisSpeaker

	// for in-memory configurations
	gcra *gcra

	// other flags
	useGCRA   bool
	gcraBurst int64

	// metrics
	statsCollector statsCollector
}

func New(options ...Option) (*Limiter, error) {
	rl := &Limiter{}
	for i := range options {
		options[i](rl)
	}
	if rl.statsCollector == nil {
		rl.statsCollector = stats.Default
	}
	if rl.redisSpeaker != nil {
		return rl, nil
	}
	// Default to in-memory GCRA
	rl.gcra = &gcra{}
	rl.useGCRA = true
	return rl, nil
}

// Allow returns true if the limit is not exceeded, false otherwise.
func (l *Limiter) Allow(ctx context.Context, cost, rate, window int64, key string) (
	bool, func(context.Context) error, error,
) {
	allowed, _, tr, err := l.allow(ctx, cost, rate, window, key)
	return allowed, tr, err
}

// AllowAfter returns true if the limit is not exceeded, false otherwise.
// Additionally, it returns the time.Duration until the next allowed request.
func (l *Limiter) AllowAfter(ctx context.Context, cost, rate, window int64, key string) (
	bool, time.Duration, func(context.Context) error, error,
) {
	return l.allow(ctx, cost, rate, window, key)
}

func (l *Limiter) allow(ctx context.Context, cost, rate, window int64, key string) (
	bool, time.Duration, func(context.Context) error, error,
) {
	if cost < 1 {
		return false, 0, nil, fmt.Errorf("cost must be greater than 0")
	}
	if rate < 1 {
		return false, 0, nil, fmt.Errorf("rate must be greater than 0")
	}
	if window < 1 {
		return false, 0, nil, fmt.Errorf("window must be greater than 0")
	}
	if key == "" {
		return false, 0, nil, fmt.Errorf("key must not be empty")
	}

	if l.redisSpeaker != nil {
		if l.useGCRA {
			defer l.getTimer(key, "redis-gcra", rate, window)()
			_, allowed, retryAfter, tr, err := l.redisGCRA(ctx, cost, rate, window, key)
			return allowed, retryAfter, tr, err
		}

		defer l.getTimer(key, "redis-sorted-set", rate, window)()
		_, allowed, retryAfter, tr, err := l.redisSortedSet(ctx, cost, rate, window, key)
		return allowed, retryAfter, tr, err
	}

	defer l.getTimer(key, "gcra", rate, window)()
	allowed, retryAfter, tr, err := l.gcraLimit(ctx, cost, rate, window, key)
	return allowed, retryAfter, tr, err
}

func (l *Limiter) redisSortedSet(ctx context.Context, cost, rate, window int64, key string) (
	time.Duration, bool, time.Duration, func(context.Context) error, error,
) {
	res, err := sortedSetScript.Run(ctx, l.redisSpeaker, []string{key}, cost, rate, window).Result()
	if err != nil {
		return 0, false, 0, nil, fmt.Errorf("could not run SortedSet Redis script: %v", err)
	}

	result, ok := res.([]interface{})
	if !ok {
		return 0, false, 0, nil, fmt.Errorf("unexpected result from SortedSet Redis script of type %T: %v", res, res)
	}
	if len(result) != 3 {
		return 0, false, 0, nil, fmt.Errorf("unexpected result from SortedSet Redis script of length %d: %+v", len(result), result)
	}

	t, ok := result[0].(int64)
	if !ok {
		return 0, false, 0, nil, fmt.Errorf("unexpected result[0] from SortedSet Redis script of type %T: %v", result[0], result[0])
	}
	redisTime := time.Duration(t) * time.Microsecond

	members, ok := result[1].(string)
	if !ok {
		return redisTime, false, 0, nil, fmt.Errorf("unexpected result[1] from SortedSet Redis script of type %T: %v", result[1], result[1])
	}
	if members == "" { // limit exceeded
		retryAfter, ok := result[2].(int64)
		if !ok {
			return redisTime, false, 0, nil, fmt.Errorf("unexpected result[2] from SortedSet Redis script of type %T: %v", result[2], result[2])
		}
		return redisTime, false, time.Duration(retryAfter) * time.Microsecond, nil, nil
	}

	r := &sortedSetRedisReturn{
		key:     key,
		members: strings.Split(members, ","),
		remover: l.redisSpeaker,
	}
	return redisTime, true, 0, r.Return, nil
}

func (l *Limiter) redisGCRA(ctx context.Context, cost, rate, window int64, key string) (
	time.Duration, bool, time.Duration, func(context.Context) error, error,
) {
	burst := rate
	if l.gcraBurst > 0 {
		burst = l.gcraBurst
	}
	res, err := gcraRedisScript.Run(ctx, l.redisSpeaker, []string{key}, burst, rate, window, cost).Result()
	if err != nil {
		return 0, false, 0, nil, fmt.Errorf("could not run GCRA Redis script: %v", err)
	}

	result, ok := res.([]any)
	if !ok {
		return 0, false, 0, nil, fmt.Errorf("unexpected result from GCRA Redis script of type %T: %v", res, res)
	}
	if len(result) != 5 {
		return 0, false, 0, nil, fmt.Errorf("unexpected result from GCRA Redis scrip of length %d: %+v", len(result), result)
	}

	t, ok := result[0].(int64)
	if !ok {
		return 0, false, 0, nil, fmt.Errorf("unexpected result[0] from GCRA Redis script of type %T: %v", result[0], result[0])
	}
	redisTime := time.Duration(t) * time.Microsecond

	allowed, ok := result[1].(int64)
	if !ok {
		return redisTime, false, 0, nil, fmt.Errorf("unexpected result[1] from GCRA Redis script of type %T: %v", result[1], result[1])
	}
	if allowed < 1 { // limit exceeded
		retryAfter, ok := result[3].(int64)
		if !ok {
			return redisTime, false, 0, nil, fmt.Errorf("unexpected result[3] from GCRA Redis script of type %T: %v", result[3], result[3])
		}
		return redisTime, false, time.Duration(retryAfter) * time.Microsecond, nil, nil
	}

	r := &unsupportedReturn{}
	return redisTime, true, 0, r.Return, nil
}

func (l *Limiter) gcraLimit(ctx context.Context, cost, rate, window int64, key string) (
	bool, time.Duration, func(context.Context) error, error,
) {
	burst := rate
	if l.gcraBurst > 0 {
		burst = l.gcraBurst
	}
	allowed, retryAfter, err := l.gcra.limit(ctx, key, cost, burst, rate, window)
	if err != nil {
		return false, 0, nil, fmt.Errorf("could not limit: %w", err)
	}
	if !allowed {
		return false, retryAfter, nil, nil // limit exceeded
	}
	r := &unsupportedReturn{}
	return true, 0, r.Return, nil
}

func (l *Limiter) getTimer(key, algo string, rate, window int64) func() {
	m := l.statsCollector.NewTaggedStat("throttling", stats.TimerType, stats.Tags{
		"key":    key,
		"algo":   algo,
		"rate":   strconv.FormatInt(rate, 10),
		"window": strconv.FormatInt(window, 10),
	})
	start := time.Now()
	return func() {
		m.Since(start)
	}
}
