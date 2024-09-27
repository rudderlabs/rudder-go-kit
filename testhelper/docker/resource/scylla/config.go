package scylla

import "github.com/ory/dockertest/v3/docker"

type Option func(*config)

func WithTag(tag string) Option {
	return func(c *config) {
		c.tag = tag
	}
}

func WithKeyspace(keyspace string) Option {
	return func(c *config) {
		c.keyspace = keyspace
	}
}

func WithDockerNetwork(network *docker.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

type config struct {
	tag      string
	keyspace string
	network  *docker.Network
}
