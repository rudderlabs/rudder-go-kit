package cachettl

import "time"

type Opt func(*cacheConfig)

// WithNoRefreshTTL disables the refresh of the TTL when the cache is accessed.
var WithNoRefreshTTL = func(c *cacheConfig) {
	c.refreshTTL = false
}

// WithNow sets the function to use to get the current time.
var WithNow = func(now func() time.Time) Opt {
	return func(c *cacheConfig) {
		c.now = now
	}
}

type cacheConfig struct {
	now        func() time.Time
	refreshTTL bool
}
