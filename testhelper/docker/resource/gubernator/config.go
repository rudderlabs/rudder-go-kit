package gubernator

import (
	"github.com/ory/dockertest/v3/docker"
)

type Option func(*config)

// WithTag overrides the default gubernator image tag.
func WithTag(tag string) Option {
	return func(c *config) {
		c.tag = tag
	}
}

// WithNetwork attaches the container to the given docker network so that it is
// reachable from other containers in the same network.
func WithNetwork(network *docker.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

type config struct {
	tag     string
	network *docker.Network
}
