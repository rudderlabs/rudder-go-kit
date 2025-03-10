package provider_test

import (
	"context"
	"testing"

	flagsmith "github.com/Flagsmith/flagsmith-go-client/v3"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/featureflags/provider"
)

// MockFlagsmithFlags implements provider.FlagsmithFlags interface
type MockFlagsmithFlags struct {
	flags []flagsmith.Flag
}

func (m *MockFlagsmithFlags) AllFlags() []flagsmith.Flag {
	return m.flags
}

// MockFlagsmithClient implements provider.FlagsmithClient interface
type MockFlagsmithClient struct {
	flags []flagsmith.Flag
	err   error
}

func (m *MockFlagsmithClient) GetIdentityFlags(_ context.Context, _ string, _ []*flagsmith.Trait) (provider.FlagsmithFlags, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &MockFlagsmithFlags{flags: m.flags}, nil
}

func TestNewFlagsmithProvider(t *testing.T) {
	t.Run("successful initialization", func(t *testing.T) {
		config := provider.ProviderConfig{
			ApiKey:                  "test-api-key",
			TimeoutInSeconds:        10,
			RetryAttempts:          3,
			RetryWaitTimeInSeconds: 1,
		}

		p, err := provider.NewFlagsmithProvider(config)
		require.NoError(t, err)
		require.NotNil(t, p)
		require.Equal(t, "flagsmith", p.Name())
	})
}

func TestFlagsmithProvider_GetFeatureFlags(t *testing.T) {
	t.Run("successful flags retrieval", func(t *testing.T) {
		mockFlags := []flagsmith.Flag{
			{
				FeatureName: "feature1",
				Enabled:     true,
				Value:      "value1",
			},
			{
				FeatureName: "feature2",
				Enabled:     false,
				Value:      "value2",
			},
		}

		p := &provider.FlagsmithProvider{
			Client: &MockFlagsmithClient{flags: mockFlags},
		}

		params := provider.ProviderParams{
			WorkspaceID: "test-workspace",
			Traits: map[string]string{
				"trait1": "value1",
				"trait2": "value2",
			},
		}

		flags, err := p.GetFeatureFlags(params)
		require.NoError(t, err)
		require.Len(t, flags, 2)

		// Check feature1
		require.Contains(t, flags, "feature1")
		require.True(t, flags["feature1"].Enabled)
		require.Equal(t, "value1", flags["feature1"].Value)
		require.False(t, flags["feature1"].IsStale)
		require.NotNil(t, flags["feature1"].LastUpdatedAt)

		// Check feature2
		require.Contains(t, flags, "feature2")
		require.False(t, flags["feature2"].Enabled)
		require.Equal(t, "value2", flags["feature2"].Value)
		require.False(t, flags["feature2"].IsStale)
		require.NotNil(t, flags["feature2"].LastUpdatedAt)
	})

	t.Run("empty flags", func(t *testing.T) {
		p := &provider.FlagsmithProvider{
			Client: &MockFlagsmithClient{flags: []flagsmith.Flag{}},
		}

		flags, err := p.GetFeatureFlags(provider.ProviderParams{})
		require.NoError(t, err)
		require.Empty(t, flags)
	})

	t.Run("client error", func(t *testing.T) {
		p := &provider.FlagsmithProvider{
			Client: &MockFlagsmithClient{err: provider.ErrProviderInit},
		}

		flags, err := p.GetFeatureFlags(provider.ProviderParams{})
		require.Error(t, err)
		require.Nil(t, flags)
	})
}

