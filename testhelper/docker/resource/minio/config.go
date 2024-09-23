package minio

import (
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
)

type Opt func(*Config)

func WithTag(tag string) Opt {
	return func(c *Config) {
		c.Tag = tag
	}
}

func WithNetwork(network *docker.Network) Opt {
	return func(c *Config) {
		c.Network = network
	}
}

func WithOptions(options ...string) Opt {
	return func(c *Config) {
		c.Options = options
	}
}

func WithNetworkBindingConfig(cfg resource.NetworkBindingConfig) Opt {
	return func(c *Config) {
		c.NetworkBindingConfig = cfg
	}
}

type Config struct {
	resource.NetworkBindingConfig
	Tag     string
	Network *docker.Network
	Options []string
}
