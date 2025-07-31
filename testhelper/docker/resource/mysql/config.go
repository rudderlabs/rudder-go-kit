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

type Config struct {
	Tag            string
	ShmSize        int64
	RegistryConfig *registry.RegistryConfig
}
