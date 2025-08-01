package registry

import (
	"fmt"
	"os"

	dc "github.com/ory/dockertest/v3/docker"
)

// ImagePath returns the full image path based on whether a registry mirror is configured
func ImagePath(image string) string {
	mirrorURL := os.Getenv("DOCKER_REGISTRY_MIRROR")
	if mirrorURL != "" {
		return fmt.Sprintf("%s/%s", mirrorURL, image)
	}
	return image
}

// AuthConfiguration returns the authentication configuration for the registry
func AuthConfiguration() dc.AuthConfiguration {
	username := os.Getenv("DOCKER_REGISTRY_MIRROR_USERNAME")
	password := os.Getenv("DOCKER_REGISTRY_MIRROR_PASSWORD")

	if username == "" && password == "" {
		return dc.AuthConfiguration{}
	}
	return dc.AuthConfiguration{
		Username: username,
		Password: password,
	}
}
