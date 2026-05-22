package throttling

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"

	redisdocker "github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/redis"
	valkeydocker "github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/valkey"
)

type tester interface {
	Context() context.Context
	Helper()
	Log(...any)
	Logf(string, ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Failed() bool
	FailNow()
	Cleanup(f func())
}

type testCase struct {
	rate,
	window int64
}

func newLimiter(t tester, opts ...Option) *Limiter {
	t.Helper()
	l, err := New(opts...)
	require.NoError(t, err)
	return l
}

func bootstrapRedis(t tester, pool *dockertest.Pool) *redis.Client {
	t.Helper()
	redisContainer, err := redisdocker.Setup(
		t.Context(), pool, t,
		redisdocker.WithTag("7.0-alpine"), // this is what is supported on AWS ElastiCache
	)
	require.NoError(t, err)

	rc := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    redisContainer.Addr,
	})
	t.Cleanup(func() { _ = rc.Close() })

	pong, err := rc.Ping(t.Context()).Result()
	require.NoError(t, err)
	require.Equal(t, "PONG", pong)

	return rc
}

func bootstrapValkey(t tester, pool *dockertest.Pool) *redis.Client {
	t.Helper()
	valkeyContainer, err := valkeydocker.Setup(
		t.Context(), pool, t,
		valkeydocker.WithTag("9.0-alpine"), // this is what is supported on AWS ElastiCache
	)
	require.NoError(t, err)

	rc := redis.NewClient(&redis.Options{
		Network: "tcp",
		Addr:    valkeyContainer.Addr,
	})
	t.Cleanup(func() { _ = rc.Close() })

	pong, err := rc.Ping(t.Context()).Result()
	require.NoError(t, err)
	require.Equal(t, "PONG", pong)

	return rc
}
