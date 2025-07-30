package registry

import (
	"os"
	"testing"

	dc "github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

func TestNewHarborRegistry(t *testing.T) {
	t.Run("with environment variables", func(t *testing.T) {
		// Set test environment variables
		originalUser := os.Getenv("HARBOR_USER_NAME")
		originalPassword := os.Getenv("HARBOR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("HARBOR_USER_NAME", originalUser)
			os.Setenv("HARBOR_PASSWORD", originalPassword)
		})

		testUser := "test-user"
		testPassword := "test-password"
		os.Setenv("HARBOR_USER_NAME", testUser)
		os.Setenv("HARBOR_PASSWORD", testPassword)

		config := NewHarborRegistry()

		require.Equal(t, RegistryHarbor, config.Type)
		require.Equal(t, "hub.dev-rudder.rudderlabs.com/dockerhub-proxy", config.URL)
		require.Equal(t, testUser, config.Username)
		require.Equal(t, testPassword, config.Password)
	})

	t.Run("without environment variables", func(t *testing.T) {
		// Unset environment variables
		originalUser := os.Getenv("HARBOR_USER_NAME")
		originalPassword := os.Getenv("HARBOR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("HARBOR_USER_NAME", originalUser)
			os.Setenv("HARBOR_PASSWORD", originalPassword)
		})

		os.Unsetenv("HARBOR_USER_NAME")
		os.Unsetenv("HARBOR_PASSWORD")

		config := NewHarborRegistry()

		require.Equal(t, RegistryHarbor, config.Type)
		require.Equal(t, "hub.dev-rudder.rudderlabs.com/dockerhub-proxy", config.URL)
		require.Empty(t, config.Username)
		require.Empty(t, config.Password)
	})
}

func TestNewDockerHubRegistry(t *testing.T) {
	config := NewDockerHubRegistry()

	require.Equal(t, RegistryDockerHub, config.Type)
	require.Empty(t, config.URL)
	require.Empty(t, config.Username)
	require.Empty(t, config.Password)
}

func TestNewCustomRegistry(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		username string
		password string
	}{
		{
			name:     "with all parameters",
			url:      "custom.registry.com",
			username: "custom-user",
			password: "custom-password",
		},
		{
			name:     "with empty credentials",
			url:      "registry.example.com",
			username: "",
			password: "",
		},
		{
			name:     "with empty URL",
			url:      "",
			username: "user",
			password: "pass",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewCustomRegistry(tc.url, tc.username, tc.password)

			require.Equal(t, RegistryCustom, config.Type)
			require.Equal(t, tc.url, config.URL)
			require.Equal(t, tc.username, config.Username)
			require.Equal(t, tc.password, config.Password)
		})
	}
}

func TestRegistryConfig_GetRegistryPath(t *testing.T) {
	testCases := []struct {
		name           string
		registryConfig *RegistryConfig
		image          string
		expectedPath   string
	}{
		{
			name: "Harbor registry",
			registryConfig: &RegistryConfig{
				Type: RegistryHarbor,
				URL:  "hub.dev-rudder.rudderlabs.com/dockerhub-proxy",
			},
			image:        "mysql",
			expectedPath: "hub.dev-rudder.rudderlabs.com/dockerhub-proxy/mysql",
		},
		{
			name: "Docker Hub registry",
			registryConfig: &RegistryConfig{
				Type: RegistryDockerHub,
			},
			image:        "postgres",
			expectedPath: "postgres",
		},
		{
			name: "Custom registry",
			registryConfig: &RegistryConfig{
				Type: RegistryCustom,
				URL:  "custom.registry.com",
			},
			image:        "redis",
			expectedPath: "custom.registry.com/redis",
		},
		{
			name: "Invalid/default registry type",
			registryConfig: &RegistryConfig{
				Type: RegistryType(99), // Invalid type
				URL:  "invalid.registry.com",
			},
			image:        "nginx",
			expectedPath: "nginx",
		},
		{
			name: "Empty image name with Harbor",
			registryConfig: &RegistryConfig{
				Type: RegistryHarbor,
				URL:  "hub.dev-rudder.rudderlabs.com/dockerhub-proxy",
			},
			image:        "",
			expectedPath: "hub.dev-rudder.rudderlabs.com/dockerhub-proxy/",
		},
		{
			name: "Image with special characters",
			registryConfig: &RegistryConfig{
				Type: RegistryCustom,
				URL:  "registry.example.com",
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

func TestRegistryTypes(t *testing.T) {
	// Test that registry type constants have expected values
	require.Equal(t, RegistryType(0), RegistryHarbor)
	require.Equal(t, RegistryType(1), RegistryDockerHub)
	require.Equal(t, RegistryType(2), RegistryCustom)
}

// TestRegistryConfig_Integration tests the integration between different methods
func TestRegistryConfig_Integration(t *testing.T) {
	t.Run("Harbor registry with authentication", func(t *testing.T) {
		// Set up environment variables
		originalUser := os.Getenv("HARBOR_USER_NAME")
		originalPassword := os.Getenv("HARBOR_PASSWORD")
		t.Cleanup(func() {
			os.Setenv("HARBOR_USER_NAME", originalUser)
			os.Setenv("HARBOR_PASSWORD", originalPassword)
		})

		os.Setenv("HARBOR_USER_NAME", "harbor-user")
		os.Setenv("HARBOR_PASSWORD", "harbor-pass")

		config := NewHarborRegistry()

		// Test path generation
		imagePath := config.GetRegistryPath("mysql")
		require.Equal(t, "hub.dev-rudder.rudderlabs.com/dockerhub-proxy/mysql", imagePath)

		// Test authentication
		auth := config.GetAuth()
		require.Equal(t, "harbor-user", auth.Username)
		require.Equal(t, "harbor-pass", auth.Password)
	})

	t.Run("Docker Hub registry without authentication", func(t *testing.T) {
		config := NewDockerHubRegistry()

		// Test path generation
		imagePath := config.GetRegistryPath("postgres")
		require.Equal(t, "postgres", imagePath)

		// Test authentication (should be empty)
		auth := config.GetAuth()
		require.Equal(t, dc.AuthConfiguration{}, auth)
	})

	t.Run("Custom registry with authentication", func(t *testing.T) {
		config := NewCustomRegistry("my-registry.com", "my-user", "my-pass")

		// Test path generation
		imagePath := config.GetRegistryPath("redis")
		require.Equal(t, "my-registry.com/redis", imagePath)

		// Test authentication
		auth := config.GetAuth()
		require.Equal(t, "my-user", auth.Username)
		require.Equal(t, "my-pass", auth.Password)
	})
}

// Benchmark tests for performance-critical operations
func BenchmarkGetRegistryPath(b *testing.B) {
	config := &RegistryConfig{
		Type: RegistryHarbor,
		URL:  "hub.dev-rudder.rudderlabs.com/dockerhub-proxy",
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
