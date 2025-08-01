package minio

import (
	"github.com/ory/dockertest/v3/docker"
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

func WithBindIP(ip string) Opt {
	return func(c *Config) {
		c.BindIP = ip
	}
}

type Config struct {
	Tag     string
	Network *docker.Network
	Options []string
	BindIP  string
}
