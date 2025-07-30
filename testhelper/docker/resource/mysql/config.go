package mysql

import "github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"

type Opt func(*Config)

func WithTag(tag string) Opt {
	return func(c *Config) {
		c.Tag = tag
	}
}

func WithShmSize(shmSize int64) Opt {
	return func(c *Config) {
		c.ShmSize = shmSize
	}
}

// WithRegistry allows to configure a custom registry
func WithRegistry(registryConfig *registry.RegistryConfig) Opt {
	return func(c *Config) {
		c.RegistryConfig = registryConfig
	}
}

// WithDockerHub allows to use Docker Hub registry
func WithDockerHub() Opt {
	return func(c *Config) {
		c.RegistryConfig = registry.NewDockerHubRegistry()
	}
}

type Config struct {
	Tag            string
	ShmSize        int64
	RegistryConfig *registry.RegistryConfig
}
