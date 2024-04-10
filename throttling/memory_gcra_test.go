package throttling

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoryGCRA(t *testing.T) {
	t.Run("burst and cost greater than rate", func(t *testing.T) {
		l := &gcra{}

		burst := int64(5)
		rate := int64(1)
		period := int64(1)

		allowed, _, err := l.limit(context.Background(), "key", burst+rate, burst, rate, period)
		require.NoError(t, err)
		require.True(t, allowed, "it should be able to fill the bucket (burst)")

		// next request should be allowed after 5 seconds
		start := time.Now()

		require.Eventually(t, func() bool {
			allowed, _, err := l.limit(context.Background(), "key", burst, burst, rate, period)
			if err != nil {
				t.Logf("Memory GCRA error: %v", err)
				return false
			}
			return allowed
		}, 10*time.Second, 1*time.Second, "next request should be eventually allowed")

		require.GreaterOrEqual(t, time.Since(start).Seconds(), 5.0, "next request should be allowed after 5 seconds")
	})
}
