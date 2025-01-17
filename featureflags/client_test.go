package featureflags_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/featureflags"
	"github.com/rudderlabs/rudder-go-kit/featureflags/cache"
	"github.com/rudderlabs/rudder-go-kit/featureflags/provider"
)

func setupFeatureFlagClient() {
	os.Setenv("FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY", "test-key")
	// call getFeatureFlagClient to initialize the client
	// we will throw away the initialized client and set a custom client in the tests
	featureflags.GetFeatureFlagClient()
	os.Unsetenv("FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY")
}

func TestGetFeatureFlagClient(t *testing.T) {

	t.Run("panics when api key not set", func(t *testing.T) {
		os.Unsetenv("FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY")
		defer func() {
			r := recover()
			require.NotNil(t, r)
			require.Contains(t, r.(string), "FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY is not set")
			featureflags.ResetFeatureFlagClient()
		}()
		featureflags.GetFeatureFlagClient()

	})

	t.Run("creates singleton client", func(t *testing.T) {
		os.Setenv("FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY", "test-key")

		// Get client first time
		client1 := featureflags.GetFeatureFlagClient()
		require.NotNil(t, client1)

		// Get client second time - should be same instance
		client2 := featureflags.GetFeatureFlagClient()
		require.NotNil(t, client2)
		require.Equal(t, client1, client2)

		featureflags.ResetFeatureFlagClient()

	})
}

type mockProvider struct {
	features map[string]*provider.FeatureValue
	err      error
}

func (m *mockProvider) GetFeatureFlags(params provider.ProviderParams) (map[string]*provider.FeatureValue, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.features, nil
}

func (m *mockProvider) Name() string {
	return "mock"
}

func TestClientImpl_IsFeatureEnabled(t *testing.T) {

	setupFeatureFlagClient()
	currentTime := time.Now()
	mockFeatures := map[string]*provider.FeatureValue{
		"feature1": {
			Enabled:       true,
			Value:         "value1",
			LastUpdatedAt: &currentTime,
		},
		"feature2": {
			Enabled:       false,
			Value:         "value2",
			LastUpdatedAt: &currentTime,
		},
	}

	mockCache := cache.NewMemoryCache(cache.CacheConfig{
		Enabled:      true,
		TTLInSeconds: 60,
	})

	client := &featureflags.ClientImpl{
		Provider: &mockProvider{features: mockFeatures},
		Cache:    mockCache,
	}
	featureflags.SetFeatureFlagClient(client)

	t.Run("returns true for enabled feature", func(t *testing.T) {
		enabled, err := featureflags.IsFeatureEnabled("workspace1", "feature1")
		require.NoError(t, err)
		require.True(t, enabled)
	})

	t.Run("returns false for disabled feature", func(t *testing.T) {
		enabled, err := featureflags.IsFeatureEnabled("workspace1", "feature2")
		require.NoError(t, err)
		require.False(t, enabled)
	})

	t.Run("returns error for non-existent feature", func(t *testing.T) {
		enabled, err := featureflags.IsFeatureEnabled("workspace1", "non-existent")
		require.Error(t, err)
		require.False(t, enabled)
	})
}

func TestClientImpl_GetFeatureValue(t *testing.T) {
	setupFeatureFlagClient()
	currentTime := time.Now()
	mockFeatures := map[string]*provider.FeatureValue{
		"feature1": {
			Enabled:       true,
			Value:         "value1",
			LastUpdatedAt: &currentTime,
		},
	}

	mockCache := cache.NewMemoryCache(cache.CacheConfig{
		Enabled:      true,
		TTLInSeconds: 60,
	}) 

	client := &featureflags.ClientImpl{
		Provider: &mockProvider{features: mockFeatures},
		Cache:    mockCache,
	}
	featureflags.SetFeatureFlagClient(client)
	t.Run("returns feature value", func(t *testing.T) {
		value, err := featureflags.GetFeatureValue("workspace1", "feature1")
		require.NoError(t, err)
		require.Equal(t, "value1", value.Value)
		require.True(t, value.Enabled)
	})

	t.Run("returns error for non-existent feature", func(t *testing.T) {
		value, err := featureflags.GetFeatureValue("workspace1", "non-existent")
		require.Error(t, err)
		require.Empty(t, value)
	})
}

type mockProviderWithGetFeatureFlagsFunc struct {
	getFeatureFlagsFunc func(params provider.ProviderParams) (map[string]*provider.FeatureValue, error)
}

func (m *mockProviderWithGetFeatureFlagsFunc) GetFeatureFlags(params provider.ProviderParams) (map[string]*provider.FeatureValue, error) {
	return m.getFeatureFlagsFunc(params)
}

func (m *mockProviderWithGetFeatureFlagsFunc) Name() string {
	return "mock"
}

func TestClientImpl_SetDefaultTraits(t *testing.T) {
	setupFeatureFlagClient()
	var capturedParams provider.ProviderParams
	mockProv := &mockProviderWithGetFeatureFlagsFunc{
		getFeatureFlagsFunc: func(params provider.ProviderParams) (map[string]*provider.FeatureValue, error) {
			capturedParams = params
			return map[string]*provider.FeatureValue{
				"feature1": {
					Enabled: true,
					Value:   "value1",
				},
			}, nil
		},
	}
	mockCache := cache.NewMemoryCache(cache.CacheConfig{
		Enabled:      false,
		TTLInSeconds: 60,
	})

	client := &featureflags.ClientImpl{
		Provider: mockProv,
		Cache:    mockCache,
	}
	featureflags.SetFeatureFlagClient(client)

	traits := map[string]string{
		"trait1": "value1",
		"trait2": "value2",
	}

	// Set default traits
	featureflags.SetDefaultTraits(traits)

	// Call a method that should use the traits
	_, err := featureflags.GetFeatureValue("workspace1", "feature1")
	require.NoError(t, err)

	// Verify the provider was called with the correct traits
	require.Equal(t, traits, capturedParams.Traits)
}
