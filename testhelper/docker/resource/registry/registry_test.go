package registry

import (
	"os"
	"testing"

	dc "github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

func TestImagePath(t *testing.T) {
	t.Run("with docker registry mirror environment variable", func(t *testing.T) {
		// Set test environment variable
		originalMirror := os.Getenv("DOCKERHUB_REGISTRY_MIRROR")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR", originalMirror)
		})

		testMirror := "registry.example.com"
		os.Setenv("DOCKERHUB_REGISTRY_MIRROR", testMirror)

		result := ImagePath("mysql")
		require.Equal(t, "registry.example.com/library/mysql", result)

		resultWithOrg := ImagePath("organization/image")
		require.Equal(t, "registry.example.com/organization/image", resultWithOrg)

		resultWithRegistry := ImagePath("custom.registry.com/image")
		require.Equal(t, "custom.registry.com/image", resultWithRegistry)
	})

	t.Run("without mirror environment variable uses Docker Hub", func(t *testing.T) {
		// Unset environment variable
		originalMirror := os.Getenv("DOCKERHUB_REGISTRY_MIRROR")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR", originalMirror)
		})

		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR")

		result := ImagePath("postgres")
		require.Equal(t, "postgres", result)
	})

	t.Run("with empty mirror URL uses Docker Hub", func(t *testing.T) {
		// Set empty environment variable
		originalMirror := os.Getenv("DOCKERHUB_REGISTRY_MIRROR")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR", originalMirror)
		})

		os.Setenv("DOCKERHUB_REGISTRY_MIRROR", "")

		result := ImagePath("redis")
		require.Equal(t, "redis", result)
	})

	t.Run("with special characters in image name", func(t *testing.T) {
		originalMirror := os.Getenv("DOCKERHUB_REGISTRY_MIRROR")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR", originalMirror)
		})

		os.Setenv("DOCKERHUB_REGISTRY_MIRROR", "registry.example.com")

		result := ImagePath("organization/image-name")
		require.Equal(t, "registry.example.com/organization/image-name", result)
	})
}

func TestAuthConfiguration(t *testing.T) {
	t.Run("with both username and password", func(t *testing.T) {
		// Set test environment variables
		originalUser := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		testUser := "test-user"
		testPassword := "test-password"
		os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", testUser)
		os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", testPassword)

		result := AuthConfiguration()
		require.Equal(t, testUser, result.Username)
		require.Equal(t, testPassword, result.Password)
	})

	t.Run("without credentials returns empty auth", func(t *testing.T) {
		// Unset environment variables
		originalUser := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")

		result := AuthConfiguration()
		require.Equal(t, dc.AuthConfiguration{}, result)
	})

	t.Run("with only username", func(t *testing.T) {
		originalUser := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", "test-user")
		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")

		result := AuthConfiguration()
		require.Equal(t, "test-user", result.Username)
		require.Equal(t, "", result.Password)
	})

	t.Run("with only password", func(t *testing.T) {
		originalUser := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", "test-password")

		result := AuthConfiguration()
		require.Equal(t, "", result.Username)
		require.Equal(t, "test-password", result.Password)
	})
}

// TestIntegration tests both functions working together
func TestIntegration(t *testing.T) {
	t.Run("Registry mirror with authentication", func(t *testing.T) {
		// Set up environment variables
		originalMirror := os.Getenv("DOCKERHUB_REGISTRY_MIRROR")
		originalUser := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR", originalMirror)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Setenv("DOCKERHUB_REGISTRY_MIRROR", "registry.example.com")
		os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", "mirror-user")
		os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", "mirror-pass")

		// Test path generation
		imagePath := ImagePath("mysql")
		require.Equal(t, "registry.example.com/library/mysql", imagePath)

		// Test authentication
		auth := AuthConfiguration()
		require.Equal(t, "mirror-user", auth.Username)
		require.Equal(t, "mirror-pass", auth.Password)
	})

	t.Run("Docker Hub registry without authentication", func(t *testing.T) {
		// Unset environment variables to simulate Docker Hub usage
		originalMirror := os.Getenv("DOCKERHUB_REGISTRY_MIRROR")
		originalUser := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR", originalMirror)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR")
		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR_USERNAME")
		os.Unsetenv("DOCKERHUB_REGISTRY_MIRROR_PASSWORD")

		// Test path generation
		imagePath := ImagePath("postgres")
		require.Equal(t, "postgres", imagePath)

		// Test authentication (should be empty)
		auth := AuthConfiguration()
		require.Equal(t, dc.AuthConfiguration{}, auth)
	})
}
