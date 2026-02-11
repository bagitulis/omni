package sku

import (
	"sync"
	"time"
)

// SKUCache provides thread-safe caching for SKU status
type SKUCache struct {
	mu   sync.RWMutex
	data map[string]*cacheEntry
	ttl  time.Duration
}

type cacheEntry struct {
	value  *PlatformSKUStatus
	expiry time.Time
}

// NewSKUCache creates a new SKU cache
func NewSKUCache() *SKUCache {
	cache := &SKUCache{
		data: make(map[string]*cacheEntry),
		ttl:  5 * time.Minute,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves SKU status from cache
func (c *SKUCache) Get(sku string) *PlatformSKUStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[sku]
	if !ok || time.Now().After(entry.expiry) {
		return nil
	}

	return entry.value
}

// Set stores SKU status in cache
func (c *SKUCache) Set(sku string, status *PlatformSKUStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[sku] = &cacheEntry{
		value:  status,
		expiry: time.Now().Add(c.ttl),
	}
}

// Delete removes a SKU from cache
func (c *SKUCache) Delete(sku string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, sku)
}

// Clear clears all cache entries
func (c *SKUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheEntry)
}

// SetTTL sets the cache TTL
func (c *SKUCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
}

// cleanup periodically removes expired entries
func (c *SKUCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.removeExpired()
	}
}

// removeExpired removes expired entries
func (c *SKUCache) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for sku, entry := range c.data {
		if now.After(entry.expiry) {
			delete(c.data, sku)
		}
	}
}

// Stats returns cache statistics
func (c *SKUCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expired := 0
	valid := 0
	now := time.Now()

	for _, entry := range c.data {
		if now.After(entry.expiry) {
			expired++
		} else {
			valid++
		}
	}

	return map[string]interface{}{
		"total":   len(c.data),
		"valid":   valid,
		"expired": expired,
		"ttl":     c.ttl.String(),
	}
}
