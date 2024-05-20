package pulsar

import (
	"github.com/ory/dockertest/v3/docker"
)

type Option func(*config)

func WithTag(tag string) Option {
	return func(c *config) {
		c.tag = tag
	}
}

func WithNetwork(network *docker.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

type config struct {
	tag     string
	network *docker.Network
}
