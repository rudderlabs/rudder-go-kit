package valkey

import (
	"context"
	_ "encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

const servicePort = "6379"

// WithTag is used to specify a custom tag that is used when pulling the Valkey image from the container registry
func WithTag(tag string) Option {
	return func(c *valkeyConfig) {
		c.tag = tag
	}
}

// WithCmdArg is used to specify the save argument when running the container.
func WithCmdArg(key, value string) Option {
	return func(c *valkeyConfig) {
		c.cmdArgs = append(c.cmdArgs, key, value)
	}
}

// WithEnv is used to pass environment variables to the container.
func WithEnv(envs ...string) Option {
	return func(c *valkeyConfig) {
		c.envs = envs
	}
}

// WithRepository is used to specify a custom image that should be pulled from the container registry
func WithRepository(repository string) Option {
	return func(c *valkeyConfig) {
		c.repository = repository
	}
}

type Resource struct {
	Addr string
}

type Option func(*valkeyConfig)

type valkeyConfig struct {
	repository string
	tag        string
	envs       []string
	cmdArgs    []string
}

func Setup(ctx context.Context, pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	conf := valkeyConfig{
		tag:        "8",
		repository: "valkey/valkey",
	}
	for _, opt := range opts {
		opt(&conf)
	}
	runOptions := &dockertest.RunOptions{
		Repository:   registry.ImagePath(conf.repository),
		Tag:          conf.tag,
		Env:          conf.envs,
		Cmd:          []string{"valkey-server"},
		ExposedPorts: []string{servicePort + "/tcp"},
		PortBindings: internal.IPv4PortBindings([]string{servicePort}),
		Auth:         registry.AuthConfiguration(),
	}
	if len(conf.cmdArgs) > 0 {
		runOptions.Cmd = append(runOptions.Cmd, conf.cmdArgs...)
	}

	// pulls a valkey image, creates a container based on it and runs it
	container, err := pool.RunWithOptions(runOptions, internal.DefaultHostConfig)
	d.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("run valkey container: %w", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	addr := fmt.Sprintf("%s:%s", container.GetBoundIP(servicePort+"/tcp"), container.GetPort(servicePort+"/tcp"))
	err = pool.Retry(func() error {
		valkeyClient := redis.NewClient(&redis.Options{
			Addr: addr,
		})
		defer func() { _ = valkeyClient.Close() }()
		_, err := valkeyClient.Ping(ctx).Result()
		return err
	})
	if err != nil {
		return nil, err
	}

	return &Resource{Addr: addr}, nil
}
