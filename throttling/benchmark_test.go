package throttling

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/rudderlabs/rudder-go-kit/cachettl"
	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"
)

/*
goos: linux, goarch: amd64
cpu: 12th Gen Intel(R) Core(TM) i9-12900K
BenchmarkLimiters/gcra_redis-24					58465			20173 ns/op
BenchmarkLimiters/sorted_sets_redis-24			60723			19385 ns/op
BenchmarkLimiters/gcra-24						9005494			129.9 ns/op
*/
func BenchmarkLimiters(b *testing.B) {
	pool, err := dockertest.NewPool("")
	require.NoError(b, err)

	var (
		rate     int64 = 10
		window   int64 = 1
		ctx            = context.Background()
		rc             = bootstrapRedis(ctx, b, pool)
		limiters       = map[string]*Limiter{
			"gcra":              newLimiter(b, WithInMemoryGCRA(0)),
			"gcra redis":        newLimiter(b, WithRedisGCRA(rc, 0)),
			"sorted sets redis": newLimiter(b, WithRedisSortedSet(rc)),
		}
	)

	for name, l := range limiters {
		l := l
		b.Run(name, func(b *testing.B) {
			key := rand.UniqueString(10)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = l.Allow(ctx, 1, rate, window, key)
			}
		})
	}
}

/*
goos: linux, goarch: amd64
cpu: 12th Gen Intel(R) Core(TM) i9-12900K
BenchmarkRedisSortedSetRemover/sortedSetRedisReturn-24		74870		14740 ns/op
*/
func BenchmarkRedisSortedSetRemover(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := dockertest.NewPool("")
	require.NoError(b, err)

	prepare := func(b *testing.B) (*redis.Client, string, []*redis.Z) {
		rc := bootstrapRedis(ctx, b, pool)

		key := rand.UniqueString(10)
		members := make([]*redis.Z, b.N*3)
		for i := range members {
			members[i] = &redis.Z{
				Score:  float64(i),
				Member: strconv.Itoa(i),
			}
		}
		_, err := rc.ZAdd(ctx, key, members...).Result()
		require.NoError(b, err)

		count, err := rc.ZCard(ctx, key).Result()
		require.NoError(b, err)
		require.EqualValues(b, b.N*3, count)

		return rc, key, members
	}

	b.Run("sortedSetRedisReturn", func(b *testing.B) {
		rc, key, members := prepare(b)
		rem := func(members ...string) *sortedSetRedisReturn {
			return &sortedSetRedisReturn{
				key:     key,
				remover: rc,
				members: members,
			}
		}

		b.ResetTimer()
		for i, j := 0, 0; i < b.N; i, j = i+1, j+3 {
			err = rem( // check error only once at the end to avoid altering benchmark results
				members[j].Member.(string),
				members[j+1].Member.(string),
				members[j+2].Member.(string),
			).Return(ctx)
		}

		require.NoError(b, err)

		b.StopTimer()
		count, err := rc.ZCard(ctx, key).Result()
		require.NoError(b, err)
		require.EqualValues(b, 0, count)
	})
}

func BenchmarkInMemoryGCRA(b *testing.B) {
	var (
		rate   = 10
		period = 100 * time.Millisecond
		burst  = rate
	)
	b.Run("one unlimited store per throttler", func(b *testing.B) {
		b.Run("single key", func(b *testing.B) {
			var (
				key   = "key"
				ctx   = context.Background()
				cache = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rl := cache.Get(key)
				if rl == nil {
					store, _ := memstore.NewCtx(0)
					rl, _ = throttled.NewGCRARateLimiterCtx(store, throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
		b.Run("multiple keys", func(b *testing.B) {
			var (
				div   = 10
				ctx   = context.Background()
				keys  = make([]string, b.N)
				cache = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)
			for i := 0; i < b.N; i++ {
				keys[i] = rand.UniqueString(10)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := keys[i/div] // don't always use a different key
				rl := cache.Get(key)
				if rl == nil {
					store, _ := memstore.NewCtx(0)
					rl, _ = throttled.NewGCRARateLimiterCtx(store, throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
	})
	b.Run("one unlimited store for all throttlers", func(b *testing.B) {
		b.Run("single key", func(b *testing.B) {
			var (
				key      = "key"
				ctx      = context.Background()
				store, _ = memstore.NewCtx(0)
				cache    = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rl := cache.Get(key)
				if rl == nil {
					rl, _ = throttled.NewGCRARateLimiterCtx(store, throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
		b.Run("multiple keys", func(b *testing.B) {
			var (
				div      = 10
				ctx      = context.Background()
				store, _ = memstore.NewCtx(0)
				keys     = make([]string, b.N)
				cache    = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)
			for i := 0; i < b.N; i++ {
				keys[i] = rand.UniqueString(10)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := keys[i/div] // don't always use a different key
				rl := cache.Get(key)
				if rl == nil {
					rl, _ = throttled.NewGCRARateLimiterCtx(store, throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
	})
	b.Run("one single key store per throttler", func(b *testing.B) {
		b.Run("single key", func(b *testing.B) {
			var (
				key   = "key"
				ctx   = context.Background()
				cache = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rl := cache.Get(key)
				if rl == nil {
					store, _ := memstore.NewCtx(1)
					rl, _ = throttled.NewGCRARateLimiterCtx(store, throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
		b.Run("multiple keys", func(b *testing.B) {
			var (
				div   = 10
				ctx   = context.Background()
				keys  = make([]string, b.N)
				cache = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)
			for i := 0; i < b.N; i++ {
				keys[i] = rand.UniqueString(10)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := keys[i/div]
				rl := cache.Get(key)
				if rl == nil {
					store, _ := memstore.NewCtx(1)
					rl, _ = throttled.NewGCRARateLimiterCtx(store, throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
	})
	b.Run("custom store per throttler", func(b *testing.B) {
		b.Run("single key", func(b *testing.B) {
			var (
				key   = "key"
				ctx   = context.Background()
				cache = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rl := cache.Get(key)
				if rl == nil {
					rl, _ = throttled.NewGCRARateLimiterCtx(newGCRAMemStore(), throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
		b.Run("multiple keys", func(b *testing.B) {
			var (
				div   = 10
				ctx   = context.Background()
				keys  = make([]string, b.N)
				cache = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
			)
			for i := 0; i < b.N; i++ {
				keys[i] = rand.UniqueString(10)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				key := keys[i/div] // don't always use a different key
				rl := cache.Get(key)
				if rl == nil {
					rl, _ = throttled.NewGCRARateLimiterCtx(newGCRAMemStore(), throttled.RateQuota{
						MaxRate:  throttled.PerDuration(rate, period),
						MaxBurst: burst,
					})
					rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
					cache.Put(key, rl, period)
				}
				_, _, _ = rl.RateLimitCtx(ctx, key, 1)
			}
		})
	})
}
