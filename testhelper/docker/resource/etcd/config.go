package etcd

import "github.com/ory/dockertest/v3/docker"

type config struct {
	network *docker.Network
	bindIP  string
}

type Option func(*config)

func WithNetwork(network *docker.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

func WithBindIP(bindIP string) Option {
	return func(c *config) {
		c.bindIP = bindIP
	}
}
