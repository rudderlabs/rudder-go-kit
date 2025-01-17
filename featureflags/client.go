package featureflags

import (
	"fmt"
	"os"
	"sync"

	"github.com/rudderlabs/rudder-go-kit/featureflags/cache"
	"github.com/rudderlabs/rudder-go-kit/featureflags/provider"
)

// client represents the feature flag client interface
type client interface {
	IsFeatureEnabled(workspaceID string, feature string) (bool, error)
	IsFeatureEnabledLatest(workspaceID string, feature string) (bool, error)
	GetFeatureValue(workspaceID string, feature string) (provider.FeatureValue, error)
	GetFeatureValueLatest(workspaceID string, feature string) (provider.FeatureValue, error)
	SetDefaultTraits(traits map[string]string)
}

var ffclient client

// getFeatureFlagClient returns the singleton feature flag client instance
func getFeatureFlagClient() client {
	initFeatureFlagClientOnce()
	return ffclient
}

var initFeatureFlagClientOnce = sync.OnceFunc(initFeatureFlagClient)

func initFeatureFlagClient() {
	// read the api key from env vars and create the default cache config
	apiKey := os.Getenv("FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY")
	if apiKey == "" {
		panic("FLAGSMITH_SERVER_SIDE_ENVIRONMENT_KEY is not set")
	}
	defaultCacheConfig := cache.CacheConfig{
		Enabled:      true,
		TTLInSeconds: 60,
	}

	// create the provider
	provider, err := provider.NewProvider(provider.ProviderConfig{
		Type:   "flagsmith",
		ApiKey: apiKey,
	})
	if err != nil {
		panic(err)
	}
	ffclient = &clientImpl{
		provider: provider,
		cache:    cache.NewMemoryCache(defaultCacheConfig),
	}
}

type clientImpl struct {
	provider      provider.Provider
	cache         cache.Cache
	defaultTraits map[string]string
}

// IsFeatureEnabled checks if a feature is enabled for a workspace
// Note: Result may be stale if returned from cache. Use IsFeatureEnabledLatest if stale values are not acceptable.
func (c *clientImpl) IsFeatureEnabled(workspaceID string, feature string) (bool, error) {
	featureval, err := c.getFeatureValue(workspaceID, feature, false)
	if err != nil {
		return false, err
	}
	return featureval.Enabled, nil
}

// IsFeatureEnabledLatest checks if a feature is enabled for a workspace, bypassing the cache
// Note: This method always fetches fresh values from the provider(bypassing the cache), which may impact performance.
func (c *clientImpl) IsFeatureEnabledLatest(workspaceID string, feature string) (bool, error) {
	featureval, err := c.getFeatureValue(workspaceID, feature, true)
	if err != nil {
		return false, err
	}
	return featureval.Enabled, nil
}

// GetFeatureValue gets the value of a feature for a workspace
// Note: Result may be stale if returned from cache. Use GetFeatureValueLatest if stale values are not acceptable.
func (c *clientImpl) GetFeatureValue(workspaceID string, feature string) (provider.FeatureValue, error) {
	featureval, err := c.getFeatureValue(workspaceID, feature, false)
	if err != nil {
		return provider.FeatureValue{}, err
	}
	return featureval, nil
}

// GetFeatureValueLatest gets the value of a feature for a workspace, bypassing the cache
// Note: This method always fetches fresh values from the provider(bypassing the cache), which may impact performance.
func (c *clientImpl) GetFeatureValueLatest(workspaceID string, feature string) (provider.FeatureValue, error) {
	featureval, err := c.getFeatureValue(workspaceID, feature, true)
	if err != nil {
		return provider.FeatureValue{}, err
	}
	return featureval, nil
}

// SetDefaultTraits sets the default traits for the feature flag client
// These traits will always be used when fetching feature flags
// This function is only meant to be called once in the application lifecycle.
func (c *clientImpl) SetDefaultTraits(traits map[string]string) {
	c.defaultTraits = traits
}

func (c *clientImpl) getAllFeatures(workspaceID string, traits map[string]string, skipCache bool) (map[string]*provider.FeatureValue, error) {
	if !skipCache && c.cache.IsEnabled() {
		if val, ok := c.cache.Get(workspaceID); ok {
			if val.IsStale {
				// refresh the feature flags asynchronously if the cache is stale
				go c.refreshFeatureFlags(workspaceID)
			}
			return val.Value, nil
		}
	}

	// Fetch from provider, if cache is disabled or cache miss
	ff, err := c.provider.GetFeatureFlags(provider.ProviderParams{
		WorkspaceID: workspaceID,
		Traits:      traits,
	})
	if err != nil {
		return nil, err
	}

	// Cache the fetched feature flags if cache is enabled
	if c.cache.IsEnabled() {
		if _, err := c.cache.Set(workspaceID, ff); err != nil {
			return nil, err
		}
	}
	return ff, nil
}

func (c *clientImpl) getFeatureValue(workspaceID string, feature string, skipCache bool) (provider.FeatureValue, error) {
	ff, err := c.getAllFeatures(workspaceID, c.defaultTraits, skipCache)
	if err != nil {
		return provider.FeatureValue{}, err
	}
	featureval, ok := ff[feature]
	if !ok {
		return provider.FeatureValue{}, newFeatureError(fmt.Sprintf("feature %s does not exist", feature))
	}
	// create a copy of the feature value and return it
	// return a copy since the feature value might be stored in a cache and should be immutable
	featurevalCopy := *featureval
	return featurevalCopy, nil
}

func (c *clientImpl) refreshFeatureFlags(workspaceID string) error {
	// fetch the feature flags from the provider
	ff, err := c.provider.GetFeatureFlags(provider.ProviderParams{
		WorkspaceID: workspaceID,
		Traits:      c.defaultTraits,
	})
	if err != nil {
		return err
	}

	// set the feature flags in the cache
	if _, err := c.cache.Set(workspaceID, ff); err != nil {
		return err
	}

	return nil
}
