package throttling

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"

	"github.com/rudderlabs/rudder-go-kit/cachettl"
)

const defaultMaxCASAttemptsLimit = 100

type gcra struct {
	mu    sync.Mutex
	store *cachettl.Cache[string, *throttled.GCRARateLimiterCtx]
}

func (g *gcra) limit(ctx context.Context, key string, cost, burst, rate, period int64) (
	bool, error,
) {
	rl, err := g.getLimiter(key, burst, rate, period)
	if err != nil {
		return false, err
	}

	limited, _, err := rl.RateLimitCtx(ctx, "key", int(cost))
	if err != nil {
		return false, fmt.Errorf("could not rate limit: %w", err)
	}

	return !limited, nil
}

func (g *gcra) getLimiter(key string, burst, rate, period int64) (*throttled.GCRARateLimiterCtx, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.store == nil {
		g.store = cachettl.New[string, *throttled.GCRARateLimiterCtx]()
	}

	rl := g.store.Get(key)
	if rl == nil {
		store, err := memstore.NewCtx(0)
		if err != nil {
			return nil, fmt.Errorf("could not create store: %w", err)
		}
		rl, err = throttled.NewGCRARateLimiterCtx(store, throttled.RateQuota{
			MaxRate:  throttled.PerDuration(int(rate), time.Duration(period)*time.Second),
			MaxBurst: int(burst),
		})
		if err != nil {
			return nil, fmt.Errorf("could not create rate limiter: %w", err)
		}
		rl.SetMaxCASAttemptsLimit(defaultMaxCASAttemptsLimit)
		g.store.Put(key, rl, time.Duration(period)*time.Second)
	}

	return rl, nil
}
