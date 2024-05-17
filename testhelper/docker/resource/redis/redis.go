package redis

import (
	"context"
	_ "encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

// WithTag is used to specify a custom tag that is used when pulling the Redis image from the container registry
func WithTag(tag string) Option {
	return func(c *redisConfig) {
		c.tag = tag
	}
}

// WithCmdArg is used to specify the save argument when running the container.
func WithCmdArg(key, value string) Option {
	return func(c *redisConfig) {
		c.cmdArgs = append(c.cmdArgs, key, value)
	}
}

// WithEnv is used to pass environment variables to the container.
func WithEnv(envs ...string) Option {
	return func(c *redisConfig) {
		c.envs = envs
	}
}

func WithRepository(repository string) Option {
	return func(rc *redisConfig) {
		rc.repository = repository
	}
}

type Resource struct {
	Addr string
}

type Option func(*redisConfig)

type redisConfig struct {
	repository string
	tag        string
	envs       []string
	cmdArgs    []string
}

func Setup(ctx context.Context, pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	conf := redisConfig{
		tag: "6",
	}
	for _, opt := range opts {
		opt(&conf)
	}
	repo := "redis"
	if conf.repository == "" {
		repo = conf.repository
	}
	runOptions := &dockertest.RunOptions{
		Repository: repo,
		Tag:        conf.tag,
		Env:        conf.envs,
		Cmd:        []string{"redis-server"},
	}
	if len(conf.cmdArgs) > 0 {
		runOptions.Cmd = append(runOptions.Cmd, conf.cmdArgs...)
	}

	// pulls a redis image, creates a container based on it and runs it
	container, err := pool.RunWithOptions(runOptions)
	if err != nil {
		return nil, err
	}
	d.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	addr := fmt.Sprintf("localhost:%s", container.GetPort("6379/tcp"))
	err = pool.Retry(func() error {
		redisClient := redis.NewClient(&redis.Options{
			Addr: addr,
		})
		defer func() { _ = redisClient.Close() }()
		_, err := redisClient.Ping(ctx).Result()
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Resource{Addr: addr}, nil
}
