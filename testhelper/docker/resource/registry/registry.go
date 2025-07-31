package registry

import (
	"fmt"
	"os"

	dc "github.com/ory/dockertest/v3/docker"
)

type RegistryConfig struct {
	URL      string
	Username string
	Password string
}

// NewRegistry creates a registry configuration that uses a mirror if configured, otherwise Docker Hub
func NewRegistry() *RegistryConfig {
	return &RegistryConfig{
		URL:      os.Getenv("DOCKER_REGISTRY_MIRROR"),
		Username: os.Getenv("DOCKER_REGISTRY_MIRROR_USERNAME"),
		Password: os.Getenv("DOCKER_REGISTRY_MIRROR_PASSWORD"),
	}
}

// GetRegistryPath returns the full image path based on whether a registry mirror is configured
func (r *RegistryConfig) GetRegistryPath(image string) string {
	if r.URL != "" {
		res := fmt.Sprintf("%s/%s", r.URL, image)
		fmt.Println(res)
		return res
	}
	return image
}

// GetAuth returns the authentication configuration for the registry
func (r *RegistryConfig) GetAuth() dc.AuthConfiguration {
	if r.Username == "" && r.Password == "" {
		return dc.AuthConfiguration{}
	}
	return dc.AuthConfiguration{
		Username: r.Username,
		Password: r.Password,
	}
}
