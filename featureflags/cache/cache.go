package cache

import (
	"time"

	"github.com/rudderlabs/rudder-go-kit/featureflags/provider"
)

// CacheEntry represents a cached item with its value and last updated timestamp
type CacheEntry struct {
	Value       map[string]*provider.FeatureValue
	LastUpdated time.Time
}

// CacheGetResult extends CacheEntry with staleness information
type CacheGetResult struct {
	*CacheEntry
	IsStale bool
}

// Cache defines the interface for cache operations
type Cache interface {
	Get(key string) (*CacheGetResult, bool)
	Set(key string, value map[string]*provider.FeatureValue) (time.Time, error)
	Delete(key string) error
	Clear() error
	IsEnabled() bool
}

// CacheError represents cache-specific errors
type CacheError struct {
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}

// CacheConfig holds configuration for the cache
type CacheConfig struct {
	Enabled      bool
	TTLInSeconds int64
}
