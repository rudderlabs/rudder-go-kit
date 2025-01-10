package provider_test

import (
    "testing"
    
    "github.com/rudderlabs/rudder-go-kit/featureflags/provider"
    "github.com/stretchr/testify/assert"
)

func TestFlagsmithProvider(t *testing.T) {
    p, err := provider.NewFlagsmithProvider(provider.ProviderConfig{
        ApiKey: "test-api-key",
    })
    assert.NoError(t, err)
    assert.NotNil(t, p)
    
    t.Run("Name", func(t *testing.T) {
        assert.Equal(t, "flagsmith", p.Name())
    })
    
    t.Run("GetFeatureFlags", func(t *testing.T) {
        params := provider.ProviderParams{
            WorkspaceID: "test-identity",
            Traits: map[string]string{
                "email": "test@example.com",
                "age":   "25",
            },
        }
        
        flags, err := p.GetFeatureFlags(params)
        assert.NoError(t, err)
        assert.IsType(t, map[string]provider.FeatureValue{}, flags)
        
        for _, flag := range flags {
            assert.NotEmpty(t, flag.Name)
            assert.NotNil(t, flag.LastUpdatedAt)
            // Value and Enabled fields will depend on your Flagsmith configuration
        }
    })
    
    t.Run("GetFeatureFlags_NoTraits", func(t *testing.T) {
        params := provider.ProviderParams{
            WorkspaceID: "test-identity",
            Traits:      map[string]string{},
        }
        
        flags, err := p.GetFeatureFlags(params)
        assert.NoError(t, err)
        assert.IsType(t, map[string]provider.FeatureValue{}, flags)
    })
}
