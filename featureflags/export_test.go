package featureflags

import (
	"sync"

	"github.com/rudderlabs/rudder-go-kit/featureflags/cache"
	"github.com/rudderlabs/rudder-go-kit/featureflags/provider"
)

// This file is used to expose internal functions/types for testing purposes.
// These will only be available during testing.

// ClientImpl exposes the internal clientImpl type for testing
type ClientImpl struct {
    Provider      provider.Provider
    Cache         cache.Cache
    DefaultTraits map[string]string
}

func SetFeatureFlagClient(c *ClientImpl) {
	ffclient = &clientImpl{
		provider: c.Provider,
		cache:    c.Cache,
		defaultTraits: c.DefaultTraits,
	}
}

// GetFeatureFlagClient exposes the internal getFeatureFlagClient function for testing
var GetFeatureFlagClient = getFeatureFlagClient 

func ResetFeatureFlagClient() {
	ffclient = nil
	initFeatureFlagClientOnce = sync.OnceFunc(initFeatureFlagClient)
}

