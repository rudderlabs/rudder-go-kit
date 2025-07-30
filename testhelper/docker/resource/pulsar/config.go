package pulsar

import (
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
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

// WithRegistry allows to configure a custom registry
func WithRegistry(registryConfig *registry.RegistryConfig) Option {
	return func(c *config) {
		c.registryConfig = registryConfig
	}
}

// WithDockerHub allows to use Docker Hub registry
func WithDockerHub() Option {
	return func(c *config) {
		c.registryConfig = registry.NewDockerHubRegistry()
	}
}

type config struct {
	tag            string
	network        *docker.Network
	registryConfig *registry.RegistryConfig
}
