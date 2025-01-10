package cache

import (
	"sync"
	"time"

	"github.com/rudderlabs/rudder-go-kit/featureflags/provider"
)

type MemoryCache struct {
	entries map[string]CacheEntry
	config  CacheConfig
	mu      sync.RWMutex
}

func NewMemoryCache(config CacheConfig) *MemoryCache {
	return &MemoryCache{
		entries: make(map[string]CacheEntry),
		config:  config,
	}
}

func (c *MemoryCache) Get(key string) (*CacheGetResult, bool) {
	if !c.IsEnabled() {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	isStale := false
	if c.config.TTLInSeconds > 0 {
		expirationTime := entry.LastUpdated.Add(time.Duration(c.config.TTLInSeconds) * time.Second)
		isStale = time.Now().After(expirationTime)
	}

	return &CacheGetResult{
		CacheEntry: &entry,
		IsStale:    isStale,
	}, true
}

func (c *MemoryCache) Set(key string, value map[string]*provider.FeatureValue) (time.Time, error) {
	if !c.IsEnabled() {
		return time.Time{}, &CacheError{Message: "Cache is disabled"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.entries[key] = CacheEntry{
		Value:       value,	
		LastUpdated: now,
	}

	return now, nil
}

func (c *MemoryCache) Delete(key string) error {
	if !c.IsEnabled() {
		return &CacheError{Message: "Cache is disabled"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	return nil
}

func (c *MemoryCache) Clear() error {
	if !c.IsEnabled() {
		return &CacheError{Message: "Cache is disabled"}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]CacheEntry)
	return nil
}

func (c *MemoryCache) IsEnabled() bool {
	return c.config.Enabled
}