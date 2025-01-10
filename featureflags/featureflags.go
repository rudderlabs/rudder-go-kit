package featureflags

import "github.com/rudderlabs/rudder-go-kit/featureflags/provider"

func IsFeatureEnabled(workspaceID string, feature string) (bool, error) {
	ff, err := getFeatureFlagClient().IsFeatureEnabled(workspaceID, feature)
	if err != nil {
		return false, err
	}
	return ff, nil
}

// IsFeatureEnabledLatest checks if a feature is enabled for a workspace, bypassing the cache
// Note: This method always fetches fresh values from the provider(bypassing the cache), which may impact performance.
func IsFeatureEnabledLatest(workspaceID string, feature string) (bool, error) {
	ff, err := getFeatureFlagClient().IsFeatureEnabledLatest(workspaceID, feature)
	if err != nil {
		return false, err
	}
	return ff, nil
}

// GetFeatureValue gets the value of a feature for a workspace
// Note: Result may be stale if returned from cache. Use GetFeatureValueLatest if stale values are not acceptable.
func GetFeatureValue(workspaceID string, feature string) (provider.FeatureValue, error) {
	ff, err := getFeatureFlagClient().GetFeatureValue(workspaceID, feature)
	if err != nil {
		return provider.FeatureValue{}, err
	}
	// create a copy of the feature value and return it
	return ff, nil
}

// GetFeatureValueLatest gets the value of a feature for a workspace, bypassing the cache
// Note: This method always fetches fresh values from the provider(bypassing the cache), which may impact performance.
func GetFeatureValueLatest(workspaceID string, feature string) (provider.FeatureValue, error) {
	ff, err := getFeatureFlagClient().GetFeatureValueLatest(workspaceID, feature)
	if err != nil {
		return provider.FeatureValue{}, err
	}
	return ff, nil
}

// SetDefaultTraits sets the default traits for the feature flag client
// These traits will always be used when fetching feature flags
// This function is only meant to be called once in the application lifecycle.
func SetDefaultTraits(traits map[string]string) {
	getFeatureFlagClient().SetDefaultTraits(traits)
}
