package resourcettl

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/rudderlabs/rudder-go-kit/cachettl"
	kitsync "github.com/rudderlabs/rudder-go-kit/sync"
)

// NewCache creates a new resource cache.
//
//   - ttl - is the time after which the resource is considered expired and cleaned up.
//   - opts - options for the cache.
//
// A resource's ttl is extended every time it is checked out.
//
// The cache keeps track of the resources' usage and makes sure that
// expired resources are not cleaned up while they are still in use
// and cleaned up only when they are no longer needed (zero checkouts).
//
// Resources with any of following methods can be cleaned up:
//   - Cleanup()
//   - Close()
//   - Close() error
//   - Stop()
//   - Stop() error
func NewCache[K comparable, R any](ttl time.Duration, opts ...cachettl.Opt) *Cache[K, R] {
	c := &Cache[K, R]{
		keyMu:     kitsync.NewPartitionLocker(),
		resources: make(map[string]R),
		checkouts: make(map[string]int),
		expiries:  make(map[string]struct{}),
		ttl:       ttl,
		ttlcache:  cachettl.New[K, string](opts...),
	}
	c.ttlcache.OnEvicted(c.onEvicted)
	return c
}

// Cache is a cache for resources that need to be closed/cleaned-up after expiration.
//
// The cache keeps track of the resources' usage and makes sure that
// expired resources are not cleaned up while they are still in use
// and cleaned up only when they are no longer needed (zero checkouts).
//
// Resources with any of following methods can be cleaned up:
//   - Cleanup()
//   - Close()
//   - Close() error
//   - Stop()
//   - Stop() error
type Cache[K comparable, R any] struct {
	// synchronizes access to the cache for a given key. This is to
	// allow multiple go-routines to access the cache concurrently for different keys, but still
	// avoid multiple go-routines creating multiple resources for the same key.
	keyMu *kitsync.PartitionLocker

	mu        sync.RWMutex        // protects the following maps
	resources map[string]R        // maps an resourceID to its resource
	checkouts map[string]int      // keeps track of how many checkouts are active for a given resourceID
	expiries  map[string]struct{} // keeps track of resources that are expired and need to be cleaned up after all checkouts are done

	ttl      time.Duration
	ttlcache *cachettl.Cache[K, string]
}

// Checkout returns a resource for the given key. If the resource is not available, it creates a new one, using the new function.
// The caller must call the returned checkin function when the resource is no longer needed, to release the resource.
// Multiple checkouts for the same key are allowed and they can all share the same resource. The resource is cleaned up
// only when all checkouts are checked-in and the resource's ttl has expired (or its key has been invalidated through [Invalidate]).
func (c *Cache[K, R]) Checkout(key K, new func() (R, error)) (resource R, checkin func(), err error) {
	defer c.lockKey(key)()

	if resourceID := c.ttlcache.Get(key); resourceID != "" {
		c.mu.Lock()
		defer c.mu.Unlock()
		r := c.resources[resourceID]
		c.checkouts[resourceID]++
		return r, c.checkinFunc(r, resourceID), nil
	}
	return c.newResource(key, new)
}

// Invalidate invalidates the resource for the given key.
func (c *Cache[K, R]) Invalidate(key K) {
	defer c.lockKey(key)()
	resourceID := c.ttlcache.Get(key)
	if resourceID != "" {
		c.ttlcache.Put(key, "", -1)
	}
	if resourceID != "" {
		c.onEvicted(key, resourceID)
	}
}

// newResource creates a new resource for the given key.
func (c *Cache[K, R]) newResource(key K, new func() (R, error)) (R, func(), error) {
	r, err := new()
	if err != nil {
		return r, nil, err
	}

	resourceID := uuid.NewString()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resources[resourceID] = r
	c.checkouts[resourceID]++
	c.ttlcache.Put(key, resourceID, c.ttl)
	return r, c.checkinFunc(r, resourceID), nil
}

// checkinFunc returns a function that decrements the checkout count and cleans up the resource if it is no longer needed.
func (c *Cache[K, R]) checkinFunc(r R, resourceID string) func() {
	var once sync.Once
	return func() {
		once.Do(func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			c.checkouts[resourceID]--
			if _, ok := c.expiries[resourceID]; ok && // resource is expired
				c.checkouts[resourceID] == 0 { // no more checkouts
				delete(c.expiries, resourceID)
				go c.cleanup(r)
			}
		})
	}
}

// onEvicted is called when a key is evicted from the cache. It cleans up the resource if it is not checked out.
func (c *Cache[K, R]) onEvicted(_ K, resourceID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	checkouts, ok := c.checkouts[resourceID]
	if !ok {
		return // already cleaned up through Invalidate
	}
	if checkouts == 0 {
		r := c.resources[resourceID]
		delete(c.resources, resourceID)
		delete(c.checkouts, resourceID)
		go c.cleanup(r)
	} else { // mark the resource for cleanup
		c.expiries[resourceID] = struct{}{}
	}
}

// cleanup cleans up the resource if it implements the cleanup interface or io.Closer.
func (c *Cache[K, R]) cleanup(r R) {
	cleanup := func() {}
	var v any = r
	switch v := v.(type) {
	case interface{ Cleanup() }:
		cleanup = v.Cleanup
	case interface{ Cleanup() error }:
		cleanup = func() { _ = v.Cleanup() }
	case interface{ Close() }:
		cleanup = v.Close
	case interface{ Close() error }:
		cleanup = func() { _ = v.Close() }
	case interface{ Stop() }:
		cleanup = v.Stop
	case interface{ Stop() error }:
		cleanup = func() { _ = v.Stop() }
	}
	cleanup()
}

// lockKey locks the key for exclusive access and returns a function to unlock it, which can be deferred.
func (c *Cache[K, R]) lockKey(key K) func() {
	k := fmt.Sprintf("%v", key)
	c.keyMu.Lock(k)
	return func() {
		c.keyMu.Unlock(k)
	}
}
