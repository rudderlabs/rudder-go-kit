package registry

import (
	"fmt"
	"os"

	dc "github.com/ory/dockertest/v3/docker"
)

type RegistryType uint8

const (
	RegistryHarbor RegistryType = iota
	RegistryDockerHub
	RegistryCustom
)

type RegistryConfig struct {
	Type     RegistryType
	URL      string
	Username string
	Password string
}

// NewHarborRegistry creates a new Harbor registry configuration
func NewHarborRegistry() *RegistryConfig {
	return &RegistryConfig{
		Type:     RegistryHarbor,
		URL:      "hub.dev-rudder.rudderlabs.com/dockerhub-proxy",
		Username: os.Getenv("HARBOR_USER_NAME"),
		Password: os.Getenv("HARBOR_PASSWORD"),
	}
}

// NewDockerHubRegistry creates a new Docker Hub registry configuration
func NewDockerHubRegistry() *RegistryConfig {
	return &RegistryConfig{
		Type: RegistryDockerHub,
	}
}

// NewCustomRegistry creates a new custom registry configuration
func NewCustomRegistry(url, username, password string) *RegistryConfig {
	return &RegistryConfig{
		Type:     RegistryCustom,
		URL:      url,
		Username: username,
		Password: password,
	}
}

// GetImagePath returns the full image path based on registry type
func (r *RegistryConfig) GetRegistryPath(image string) string {
	switch r.Type {
	case RegistryHarbor:
		return fmt.Sprintf("%s/%s", r.URL, image)
	case RegistryDockerHub:
		return image
	case RegistryCustom:
		return fmt.Sprintf("%s/%s", r.URL, image)
	default:
		return image
	}
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
