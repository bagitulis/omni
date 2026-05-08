package sku

import (
	"fmt"
	"sync"
	"time"
)

// SKUCache provides thread-safe caching for SKU status
type SKUCache struct {
	mu     sync.RWMutex
	data   map[string]*cacheEntry
	ttl    time.Duration
	stopCh chan struct{}
}

type cacheEntry struct {
	value  *PlatformSKUStatus
	expiry time.Time
}

// NewSKUCache creates a new SKU cache
func NewSKUCache() *SKUCache {
	cache := &SKUCache{
		data:   make(map[string]*cacheEntry),
		ttl:    5 * time.Minute,
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves SKU status from cache (tenant-scoped)
func (c *SKUCache) Get(tenantID, sku string) *PlatformSKUStatus {
	key := buildCacheKey(tenantID, sku)
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.data[key]
	if !ok || time.Now().After(entry.expiry) {
		return nil
	}

	return entry.value
}

// Set stores SKU status in cache (tenant-scoped)
func (c *SKUCache) Set(tenantID, sku string, status *PlatformSKUStatus) {
	key := buildCacheKey(tenantID, sku)
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheEntry{
		value:  status,
		expiry: time.Now().Add(c.ttl),
	}
}

// Delete removes a SKU from cache (tenant-scoped)
func (c *SKUCache) Delete(tenantID, sku string) {
	key := buildCacheKey(tenantID, sku)
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
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

	for {
		select {
		case <-ticker.C:
			c.removeExpired()
		case <-c.stopCh:
			return
		}
	}
}

			// Stop stops the cleanup goroutine
		func (c *SKUCache) Stop() {
	close(c.stopCh)
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

// buildCacheKey creates a tenant-scoped cache key to prevent cross-tenant leakage
func buildCacheKey(tenantID, sku string) string {
	return fmt.Sprintf("tenant:%s:%s", tenantID, sku)
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
