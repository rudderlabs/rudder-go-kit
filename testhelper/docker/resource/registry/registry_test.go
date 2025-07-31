package registry

import (
	"os"
	"testing"

	dc "github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	t.Run("with docker registry mirror environment variables", func(t *testing.T) {
		// Set test environment variables
		originalMirror := os.Getenv("DOCKER_REGISTRY_MIRROR")
		originalUser := os.Getenv("DOCKER_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKER_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKER_REGISTRY_MIRROR", originalMirror)
			os.Setenv("DOCKER_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKER_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		testMirror := "registry.example.com"
		testUser := "test-user"
		testPassword := "test-password"
		os.Setenv("DOCKER_REGISTRY_MIRROR", testMirror)
		os.Setenv("DOCKER_REGISTRY_MIRROR_USERNAME", testUser)
		os.Setenv("DOCKER_REGISTRY_MIRROR_PASSWORD", testPassword)

		config := NewRegistry()

		require.Equal(t, testMirror, config.URL)
		require.Equal(t, testUser, config.Username)
		require.Equal(t, testPassword, config.Password)
	})

	t.Run("without mirror environment variables uses Docker Hub", func(t *testing.T) {
		// Unset environment variables
		originalMirror := os.Getenv("DOCKER_REGISTRY_MIRROR")
		originalUser := os.Getenv("DOCKER_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKER_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKER_REGISTRY_MIRROR", originalMirror)
			os.Setenv("DOCKER_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKER_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Unsetenv("DOCKER_REGISTRY_MIRROR")
		os.Unsetenv("DOCKER_REGISTRY_MIRROR_USERNAME")
		os.Unsetenv("DOCKER_REGISTRY_MIRROR_PASSWORD")

		config := NewRegistry()

		require.Empty(t, config.URL)
		require.Empty(t, config.Username)
		require.Empty(t, config.Password)
	})
}

func TestRegistryConfig_GetRegistryPath(t *testing.T) {
	testCases := []struct {
		name           string
		registryConfig *RegistryConfig
		image          string
		expectedPath   string
	}{
		{
			name: "Registry mirror with URL",
			registryConfig: &RegistryConfig{
				URL: "registry.example.com",
			},
			image:        "mysql",
			expectedPath: "registry.example.com/mysql",
		},
		{
			name: "Docker Hub registry (no URL)",
			registryConfig: &RegistryConfig{
				URL: "",
			},
			image:        "postgres",
			expectedPath: "postgres",
		},
		{
			name: "Custom registry with URL",
			registryConfig: &RegistryConfig{
				URL: "custom.registry.com",
			},
			image:        "redis",
			expectedPath: "custom.registry.com/redis",
		},
		{
			name: "Empty image name with mirror URL",
			registryConfig: &RegistryConfig{
				URL: "registry.example.com",
			},
			image:        "",
			expectedPath: "registry.example.com/",
		},
		{
			name: "Image with special characters",
			registryConfig: &RegistryConfig{
				URL: "registry.example.com",
			},
			image:        "organization/image-name",
			expectedPath: "registry.example.com/organization/image-name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.registryConfig.GetRegistryPath(tc.image)
			require.Equal(t, tc.expectedPath, result)
		})
	}
}

func TestRegistryConfig_GetAuth(t *testing.T) {
	testCases := []struct {
		name           string
		registryConfig *RegistryConfig
		expectedAuth   dc.AuthConfiguration
	}{
		{
			name: "with username and password",
			registryConfig: &RegistryConfig{
				Username: "test-user",
				Password: "test-password",
			},
			expectedAuth: dc.AuthConfiguration{
				Username: "test-user",
				Password: "test-password",
			},
		},
		{
			name: "with empty username and password",
			registryConfig: &RegistryConfig{
				Username: "",
				Password: "",
			},
			expectedAuth: dc.AuthConfiguration{},
		},
		{
			name: "with only username",
			registryConfig: &RegistryConfig{
				Username: "test-user",
				Password: "",
			},
			expectedAuth: dc.AuthConfiguration{
				Username: "test-user",
				Password: "",
			},
		},
		{
			name: "with only password",
			registryConfig: &RegistryConfig{
				Username: "",
				Password: "test-password",
			},
			expectedAuth: dc.AuthConfiguration{
				Username: "",
				Password: "test-password",
			},
		},
		{
			name: "with whitespace credentials",
			registryConfig: &RegistryConfig{
				Username: "   ",
				Password: "   ",
			},
			expectedAuth: dc.AuthConfiguration{
				Username: "   ",
				Password: "   ",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.registryConfig.GetAuth()
			require.Equal(t, tc.expectedAuth, result)
		})
	}
}

// TestRegistryConfig_Integration tests the integration between different methods
func TestRegistryConfig_Integration(t *testing.T) {
	t.Run("Registry mirror with authentication", func(t *testing.T) {
		// Set up environment variables
		originalMirror := os.Getenv("DOCKER_REGISTRY_MIRROR")
		originalUser := os.Getenv("DOCKER_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKER_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKER_REGISTRY_MIRROR", originalMirror)
			os.Setenv("DOCKER_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKER_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Setenv("DOCKER_REGISTRY_MIRROR", "registry.example.com")
		os.Setenv("DOCKER_REGISTRY_MIRROR_USERNAME", "mirror-user")
		os.Setenv("DOCKER_REGISTRY_MIRROR_PASSWORD", "mirror-pass")

		config := NewRegistry()

		// Test path generation
		imagePath := config.GetRegistryPath("mysql")
		require.Equal(t, "registry.example.com/mysql", imagePath)

		// Test authentication
		auth := config.GetAuth()
		require.Equal(t, "mirror-user", auth.Username)
		require.Equal(t, "mirror-pass", auth.Password)
	})

	t.Run("Docker Hub registry without authentication", func(t *testing.T) {
		// Unset environment variables to simulate Docker Hub usage
		originalMirror := os.Getenv("DOCKER_REGISTRY_MIRROR")
		originalUser := os.Getenv("DOCKER_REGISTRY_MIRROR_USERNAME")
		originalPassword := os.Getenv("DOCKER_REGISTRY_MIRROR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("DOCKER_REGISTRY_MIRROR", originalMirror)
			os.Setenv("DOCKER_REGISTRY_MIRROR_USERNAME", originalUser)
			os.Setenv("DOCKER_REGISTRY_MIRROR_PASSWORD", originalPassword)
		})

		os.Unsetenv("DOCKER_REGISTRY_MIRROR")
		os.Unsetenv("DOCKER_REGISTRY_MIRROR_USERNAME")
		os.Unsetenv("DOCKER_REGISTRY_MIRROR_PASSWORD")

		config := NewRegistry()

		// Test path generation
		imagePath := config.GetRegistryPath("postgres")
		require.Equal(t, "postgres", imagePath)

		// Test authentication (should be empty)
		auth := config.GetAuth()
		require.Equal(t, dc.AuthConfiguration{}, auth)
	})
}

// Benchmark tests for performance-critical operations
func BenchmarkGetRegistryPath(b *testing.B) {
	config := &RegistryConfig{
		URL: "registry.example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetRegistryPath("mysql")
	}
}

func BenchmarkGetAuth(b *testing.B) {
	config := &RegistryConfig{
		Username: "test-user",
		Password: "test-password",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetAuth()
	}
}
