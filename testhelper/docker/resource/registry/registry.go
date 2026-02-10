package registry

import (
	"fmt"
	"os"
	"strings"

	dc "github.com/ory/dockertest/v3/docker"
)

// ImagePath returns the full image path based on whether a dockerhub registry mirror is configured
func ImagePath(image string) string {
	mirrorURL := os.Getenv("DOCKERHUB_REGISTRY_MIRROR")

	hasRegistry := func() bool { // If the image already has a registry, we don't need to prepend the mirror URL
		parts := strings.Split(image, "/")
		if len(parts) == 0 {
			return false
		}
		host := parts[0]
		return strings.Contains(host, ".") ||
			strings.Contains(host, ":") ||
			host == "localhost"
	}()
	if mirrorURL != "" && !hasRegistry {
		needsLibrary := !strings.Contains(image, "/")
		if needsLibrary {
			image = "library/" + image
		}
		return fmt.Sprintf("%s/%s", mirrorURL, image)
	}
	return image
}

// AuthConfiguration returns the authentication configuration for the registry
func AuthConfiguration() dc.AuthConfiguration {
	username := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
	password := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")

	if username == "" && password == "" {
		return dc.AuthConfiguration{}
	}
	return dc.AuthConfiguration{
		Username: username,
		Password: password,
	}
}
