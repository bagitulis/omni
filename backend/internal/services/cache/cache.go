package cache

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// CacheManager defines the interface for cache operations
type CacheManager interface {
	Get(tenantID, key string) (interface{}, bool)
	Set(tenantID, key string, value interface{}, ttl time.Duration) error
	Delete(tenantID, key string) error
	DeletePattern(tenantID, pattern string) error
	ClearTenant(tenantID string)
	Count() int
	Stop()
}

// Item represents a cached item with expiration
type Item struct {
	Value      interface{}
	Expiration int64
}

// Expired returns true if the item has expired
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// Cache is a thread-safe in-memory cache with TTL and multi-tenant support
type Cache struct {
	items             map[string]Item
	mu                sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan bool
}

// Ensure Cache implements CacheManager
var _ CacheManager = (*Cache)(nil)

// New creates a new cache with default expiration and cleanup interval
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:             make(map[string]Item),
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		stopCleanup:       make(chan bool),
	}

	if cleanupInterval > 0 {
		go c.startCleanup()
	}

	return c
}

// buildKey creates a tenant-scoped cache key
func buildKey(tenantID, key string) string {
	return fmt.Sprintf("tenant:%s:%s", tenantID, key)
}

// Get retrieves an item from the cache
func (c *Cache) Get(tenantID, key string) (interface{}, bool) {
	fullKey := buildKey(tenantID, key)

	c.mu.RLock()
	item, found := c.items[fullKey]
	c.mu.RUnlock()

	if !found {
		return nil, false
	}

	if item.Expired() {
		c.Delete(tenantID, key)
		return nil, false
	}

	return item.Value, true
}

// Set adds an item to the cache with TTL
func (c *Cache) Set(tenantID, key string, value interface{}, ttl time.Duration) error {
	fullKey := buildKey(tenantID, key)

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	} else if c.defaultExpiration > 0 {
		expiration = time.Now().Add(c.defaultExpiration).UnixNano()
	}

	c.mu.Lock()
	c.items[fullKey] = Item{
		Value:      value,
		Expiration: expiration,
	}
	c.mu.Unlock()

	return nil
}

// Delete removes an item from the cache
func (c *Cache) Delete(tenantID, key string) error {
	fullKey := buildKey(tenantID, key)

	c.mu.Lock()
	delete(c.items, fullKey)
	c.mu.Unlock()

	return nil
}

// DeletePattern removes all items matching a pattern (e.g., "products:*")
func (c *Cache) DeletePattern(tenantID, pattern string) error {
	prefix := buildKey(tenantID, "")

	// Handle wildcard pattern
	searchPattern := pattern
	if strings.HasSuffix(pattern, "*") {
		searchPattern = pattern[:len(pattern)-1]
	}
	fullPrefix := prefix + searchPattern

	c.mu.Lock()
	for key := range c.items {
		if strings.HasPrefix(key, fullPrefix) {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()

	return nil
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]Item)
	c.mu.Unlock()
}

// ClearTenant removes all items for a specific tenant
func (c *Cache) ClearTenant(tenantID string) {
	prefix := buildKey(tenantID, "")

	c.mu.Lock()
	for key := range c.items {
		if strings.HasPrefix(key, prefix) {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()
}

// Count returns the number of items in the cache
func (c *Cache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// startCleanup runs periodic cleanup of expired items
func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.deleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// deleteExpired removes all expired items from the cache
func (c *Cache) deleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	for key, item := range c.items {
		if item.Expiration > 0 && now > item.Expiration {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()
}

// Stop stops the cleanup goroutine
func (c *Cache) Stop() {
	close(c.stopCleanup)
}
