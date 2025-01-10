package cache_test

import (
	"testing"
	"time"

	"github.com/rudderlabs/rudder-go-kit/featureflags/cache"
	"github.com/rudderlabs/rudder-go-kit/featureflags/provider"
)

func TestMemoryCache(t *testing.T) {
	config := cache.CacheConfig{
		Enabled:      true,
		TTLInSeconds: 1,
	}

	c := cache.NewMemoryCache(config)

	testValue := map[string]*provider.FeatureValue{
		"key1": {
			Value: "value1",
		},
	}

	// Test Set and Get
	setTime, err := c.Set("key1", testValue)
	if err != nil {
		t.Errorf("Failed to set cache: %v", err)
	}

	result, ok := c.Get("key1")
	if !ok {
		t.Error("Expected cache hit, got miss")
	}
	if result.LastUpdated != setTime {
		t.Error("LastUpdated time mismatch")
	}

	// Test staleness
	time.Sleep(2 * time.Second)
	result, ok = c.Get("key1")
	if !ok {
		t.Error("Expected stale cache entry")
	}

	// Test Delete
	err = c.Delete("key1")
	if err != nil {
		t.Errorf("Failed to delete cache: %v", err)
	}

	result, ok = c.Get("key1")
	if ok {
		t.Error("Expected cache miss after delete")
	}

	// Test Clear
	c.Set("key1", testValue)
	c.Set("key2", testValue)
	
	err = c.Clear()
	if err != nil {
		t.Errorf("Failed to clear cache: %v", err)
	}

	result, ok = c.Get("key1")
	if ok {
		t.Error("Expected cache miss after clear")
	}
}
