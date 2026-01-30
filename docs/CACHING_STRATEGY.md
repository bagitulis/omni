# 🚀 OMNI Caching Strategy - Complete Guide

**Version**: 1.0  
**Date**: 2026-01-30  
**Status**: Active  
**Author**: OMNI Backend Team

---

## 📋 Table of Contents

1. [Executive Summary](#executive-summary)
2. [Core Principles](#core-principles)
3. [Data Classification](#data-classification)
4. [Cache TTL Matrix](#cache-ttl-matrix)
5. [Recommended Architecture](#recommended-architecture)
6. [Implementation Patterns](#implementation-patterns)
7. [Code Examples](#code-examples)
8. [Cache Invalidation Strategy](#cache-invalidation-strategy)
9. [Testing Checklist](#testing-checklist)
10. [Monitoring & Metrics](#monitoring--metrics)
11. [AI Agent Workflow Rules](#ai-agent-workflow-rules)
12. [Migration to Redis](#migration-to-redis)

---

## Executive Summary

### 🎯 Objectives

1. **Preventive Optimization**: Improve performance before bottlenecks occur
2. **Smart Caching**: Balance speed with data freshness
3. **User-Centric**: Never cache user-triggered real-time operations
4. **Multi-Tenant Safe**: Prevent cross-tenant data leaks
5. **Future-Proof**: Easy upgrade path to Redis when scaling

### ✅ Success Criteria

| Metric                        | Target                        |
| ----------------------------- | ----------------------------- |
| Analytics dashboard load time | < 500ms (after cache hit)     |
| Cache hit rate                | > 70% for analytics endpoints |
| Cross-tenant data leaks       | Zero incidents                |
| Test coverage for cache logic | 100%                          |

### 🚫 Non-Goals (What We Do NOT Cache)

- ❌ User-triggered actions (order CRUD, manual sync buttons)
- ❌ Real-time data (webhooks, live stock levels)
- ❌ Sensitive data (passwords, API tokens)
- ❌ Data that must be 100% consistent (financial transactions)

---

## Core Principles

### 1️⃣ Real-time First for User Actions

```
RULE: If user clicks a button expecting immediate action → NO CACHE
```

**Examples of NO CACHE scenarios:**

| User Action                | Why No Cache                                             |
| -------------------------- | -------------------------------------------------------- |
| Click "Sync Orders" button | User expects fresh data from platform                    |
| Create new product         | User expects immediate confirmation                      |
| Update inventory           | Stock must be accurate to prevent overselling            |
| Webhook receives order     | Platform push notification must be processed immediately |

### 2️⃣ Cache for Display/Background Data Only

```
RULE: If data is for display and can be 1 minute old → CACHE IT
```

**Examples of CACHEABLE scenarios:**

| Display Data                 | Why Cacheable                         |
| ---------------------------- | ------------------------------------- |
| Analytics dashboard          | 1-minute delay acceptable for reports |
| Product listings in dropdown | Metadata rarely changes               |
| Global config settings       | Admin-only updates, very rare         |
| Historical analytics         | Past data doesn't change              |

### 3️⃣ Multi-Tenant Isolation (CRITICAL)

**Cache Key Format:**

```
tenant:{tenant_id}:{resource}:{id}
```

**Examples:**

```go
// ✅ CORRECT: Tenant-aware cache key
cacheKey := fmt.Sprintf("tenant:%s:analytics:summary", tenantID)

// ✅ CORRECT: Tenant-aware with resource ID
cacheKey := fmt.Sprintf("tenant:%s:product:%s", tenantID, productID)

// ❌ WRONG: Shared key (DATA LEAK RISK!)
cacheKey := "analytics:summary"  // NEVER DO THIS!
```

### 4️⃣ Graceful Degradation

```
RULE: Cache failure should NEVER break the application
```

**Pattern:**

1. Try cache → Success → Return cached data
2. Cache miss/error → Fallback to database → Still works!
3. Cache save fails → Log warning → Continue (don't fail request)

---

## Data Classification

### 🔴 Tier 1: Real-time (NO CACHE)

**Characteristics:**

- User-triggered actions
- Webhook events
- Live external API responses
- Financial/critical operations

| Data Type            | Reason                          | Handler Pattern           |
| -------------------- | ------------------------------- | ------------------------- |
| Order CRUD           | User expects immediate feedback | Direct to DB, no cache    |
| Inventory Stock      | Overselling risk if stale       | Direct to DB, no cache    |
| Webhook Events       | Platform push notifications     | Process immediately       |
| Manual Sync          | User-triggered API calls        | Bypass cache entirely     |
| Create/Update/Delete | Mutation operations             | Direct to DB + invalidate |

### 🟡 Tier 2: Semi-Static (SHORT CACHE: 1-5 minutes)

**Characteristics:**

- Displayed to users (dashboards, listings)
- Updated occasionally via background jobs
- Acceptable 1-minute delay

| Data Type           | TTL             | Use Case            |
| ------------------- | --------------- | ------------------- |
| Analytics Summary   | **60 seconds**  | Dashboard metrics   |
| ML Portfolio Health | **60 seconds**  | Health scores       |
| Ads Performance     | **60 seconds**  | Marketing metrics   |
| Product Listings    | **300 seconds** | Dropdown selections |

### 🟢 Tier 3: Static (LONG CACHE: 1-24 hours)

**Characteristics:**

- Configuration data
- Reference data (categories, variants)
- Rarely changes
- Updated via admin UI

| Data Type            | TTL                    | Invalidation Trigger       |
| -------------------- | ---------------------- | -------------------------- |
| Global Config        | **1 hour**             | On admin update            |
| Platform Credentials | **Until token expiry** | On token refresh           |
| Tenant Settings      | **15 minutes**         | On settings save           |
| Valid Tenant List    | **24 hours**           | On new tenant registration |

---

## Cache TTL Matrix

### Complete Reference Table

| Data Type           | Category | TTL          | Cache Key Pattern                        | Invalidation Events          |
| ------------------- | -------- | ------------ | ---------------------------------------- | ---------------------------- |
| Analytics Summary   | Tier 2   | 60s          | `tenant:{id}:analytics:summary`          | Sync complete, webhook order |
| ML Portfolio Health | Tier 2   | 60s          | `tenant:{id}:analytics:ml:health`        | Sync complete, ML recalc     |
| Ads Performance     | Tier 2   | 60s          | `tenant:{id}:ads:{platform}:performance` | Sync complete                |
| Product List        | Tier 2   | 300s         | `tenant:{id}:products:list:page:{n}`     | Product update webhook       |
| Product Detail      | Tier 2   | 300s         | `tenant:{id}:product:{id}`               | Product update               |
| Global Config       | Tier 3   | 3600s        | `system:config:{platform}:{key}`         | Config save                  |
| Tenant Settings     | Tier 3   | 900s         | `tenant:{id}:settings`                   | Settings save                |
| Platform Creds      | Tier 3   | Token expiry | `tenant:{id}:creds:{platform}`           | Token refresh                |

### TTL Selection Logic

```go
const (
    CacheTTLAnalytics = 60 * time.Second   // Tier 2: Dashboard data
    CacheTTLProducts  = 5 * time.Minute    // Tier 2: Product metadata
    CacheTTLConfig    = 1 * time.Hour      // Tier 3: System config
    CacheTTLSettings  = 15 * time.Minute   // Tier 3: User settings
)

func GetTTL(dataType string) time.Duration {
    switch dataType {
    case "analytics", "ml", "ads":
        return CacheTTLAnalytics
    case "products":
        return CacheTTLProducts
    case "config":
        return CacheTTLConfig
    case "settings":
        return CacheTTLSettings
    default:
        return 5 * time.Minute  // Safe default
    }
}
```

---

## Recommended Architecture

### Phase 1: In-Memory Cache (Current)

**Location:** `backend/internal/services/cache/cache.go`

**Stack:**

- ✅ Thread-safe `map[string]Item` with mutex
- ✅ TTL support with auto-cleanup goroutine
- ✅ Zero external dependencies
- ✅ Already implemented (needs multi-tenant refactor)

**Pros:**

- No external dependencies (Redis, etc.)
- Fast (in-process memory access)
- Simple to debug and maintain

**Cons:**

- Lost on server restart (acceptable for analytics)
- Not shared across multiple server instances
- Memory limited to single process

**When to Use:**

- Single server deployment
- Cache hit rate optimization
- Preventive performance improvement

### Phase 2: Redis Cache (Future Scaling)

**When to Upgrade:**

- Deploying multiple backend instances
- Need cache sharing across servers
- Need cache persistence across restarts
- Cache hit rate > 80% (cache becomes critical)

**Stack:**

- `github.com/redis/go-redis/v9`
- Separate Redis container in Docker
- Same interface as in-memory cache

**Migration Path:**

1. Create `CacheManager` interface
2. Implement both `InMemoryCache` and `RedisCache`
3. Switch via configuration flag

---

## Implementation Patterns

### Pattern 1: Cache-Aside (Read-Through)

```go
func (s *AnalyticsService) GetSummary(ctx context.Context, tenantID string) (*Summary, error) {
    cacheKey := "analytics:summary"

    // 1. Try cache first
    if cached, found := s.cache.Get(tenantID, cacheKey); found {
        return cached.(*Summary), nil
    }

    // 2. Cache miss → Query database
    summary, err := s.repo.GetSummary(ctx, tenantID)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache for next request
    s.cache.Set(tenantID, cacheKey, summary, CacheTTLAnalytics)

    return summary, nil
}
```

### Pattern 2: Singleflight (Stampede Prevention)

**Problem:** When cache expires, 100 concurrent requests all hit the database simultaneously.

**Solution:** Use `golang.org/x/sync/singleflight` - only ONE request queries DB, others wait for result.

```go
import "golang.org/x/sync/singleflight"

type AnalyticsService struct {
    cache        *cache.Cache
    repo         *AnalyticsRepository
    singleflight singleflight.Group
}

func (s *AnalyticsService) GetSummary(ctx context.Context, tenantID string) (*Summary, error) {
    cacheKey := "analytics:summary"

    // 1. Try cache
    if cached, found := s.cache.Get(tenantID, cacheKey); found {
        return cached.(*Summary), nil
    }

    // 2. Cache miss → Use singleflight to prevent stampede
    sfKey := fmt.Sprintf("%s:%s", tenantID, cacheKey)
    result, err, _ := s.singleflight.Do(sfKey, func() (interface{}, error) {
        // Only ONE goroutine executes this, others wait
        summary, err := s.repo.GetSummary(ctx, tenantID)
        if err != nil {
            return nil, err
        }

        // Cache the result
        s.cache.Set(tenantID, cacheKey, summary, CacheTTLAnalytics)

        return summary, nil
    })

    if err != nil {
        return nil, err
    }

    return result.(*Summary), nil
}
```

**Impact:**

- Before: 100 requests → 100 DB queries
- After: 100 requests → 1 DB query (99 wait for result)
- Reduction: 99% fewer database hits during cache miss

### Pattern 3: Write-Through Invalidation

```go
func (s *ProductService) UpdateProduct(ctx context.Context, tenantID string, product *Product) error {
    // 1. Update database (source of truth)
    if err := s.repo.Update(ctx, product); err != nil {
        return err
    }

    // 2. Invalidate cache immediately
    productKey := fmt.Sprintf("product:%d", product.ID)
    s.cache.Delete(tenantID, productKey)

    // 3. Invalidate related caches
    s.cache.DeletePattern(tenantID, "products:list:*")

    return nil
}
```

### Pattern 4: Graceful Degradation

```go
func (s *AnalyticsService) GetSummary(ctx context.Context, tenantID string) (*Summary, error) {
    cacheKey := "analytics:summary"

    // Try cache, but don't fail if cache is broken
    if cached, found := s.cache.Get(tenantID, cacheKey); found {
        return cached.(*Summary), nil
    }

    // Cache miss or error - always fallback to database
    summary, err := s.repo.GetSummary(ctx, tenantID)
    if err != nil {
        return nil, err  // DB error is real error
    }

    // Try to cache, but don't fail request if caching fails
    if err := s.cache.Set(tenantID, cacheKey, summary, CacheTTLAnalytics); err != nil {
        log.Warn().
            Err(err).
            Str("tenant_id", tenantID).
            Str("cache_key", cacheKey).
            Msg("Failed to cache data, continuing without cache")
        // Don't return error - request succeeded!
    }

    return summary, nil
}
```

---

## Code Examples

### Example 1: Multi-Tenant Cache Service

```go
// backend/internal/services/cache/cache.go

package cache

import (
    "fmt"
    "sync"
    "time"
)

// CacheManager defines the interface for cache operations
type CacheManager interface {
    Get(tenantID, key string) (interface{}, bool)
    Set(tenantID, key string, value interface{}, ttl time.Duration) error
    Delete(tenantID, key string) error
    DeletePattern(tenantID, pattern string) error
}

// Cache is a thread-safe in-memory cache with TTL and multi-tenant support
type Cache struct {
    items             map[string]Item
    mu                sync.RWMutex
    defaultExpiration time.Duration
    cleanupInterval   time.Duration
    stopCleanup       chan bool
}

// Item represents a cached item with expiration
type Item struct {
    Value      interface{}
    Expiration int64
}

// New creates a new cache instance
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
func (c *Cache) buildKey(tenantID, key string) string {
    return fmt.Sprintf("tenant:%s:%s", tenantID, key)
}

// Get retrieves an item from the cache
func (c *Cache) Get(tenantID, key string) (interface{}, bool) {
    fullKey := c.buildKey(tenantID, key)

    c.mu.RLock()
    item, found := c.items[fullKey]
    c.mu.RUnlock()

    if !found {
        return nil, false
    }

    // Check expiration
    if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
        c.Delete(tenantID, key)
        return nil, false
    }

    return item.Value, true
}

// Set adds an item to the cache with TTL
func (c *Cache) Set(tenantID, key string, value interface{}, ttl time.Duration) error {
    fullKey := c.buildKey(tenantID, key)

    var expiration int64
    if ttl > 0 {
        expiration = time.Now().Add(ttl).UnixNano()
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
    fullKey := c.buildKey(tenantID, key)

    c.mu.Lock()
    delete(c.items, fullKey)
    c.mu.Unlock()

    return nil
}

// DeletePattern removes all items matching a pattern (e.g., "products:*")
func (c *Cache) DeletePattern(tenantID, pattern string) error {
    prefix := fmt.Sprintf("tenant:%s:%s", tenantID, pattern[:len(pattern)-1]) // Remove *

    c.mu.Lock()
    for key := range c.items {
        if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
            delete(c.items, key)
        }
    }
    c.mu.Unlock()

    return nil
}

// startCleanup periodically removes expired items
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

// deleteExpired removes all expired items
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
```

### Example 2: Analytics Service with Cache

```go
// backend/internal/services/analytics/unified_service.go

package analytics

import (
    "context"
    "fmt"
    "time"

    "github.com/omni/backend/internal/models"
    "github.com/omni/backend/internal/repositories"
    "github.com/omni/backend/internal/services/cache"
    "github.com/rs/zerolog/log"
    "golang.org/x/sync/singleflight"
)

const CacheTTLAnalytics = 60 * time.Second

type UnifiedAnalyticsService struct {
    repo         *repositories.AnalyticsRepository
    cache        cache.CacheManager
    singleflight singleflight.Group
}

func NewUnifiedAnalyticsService(
    repo *repositories.AnalyticsRepository,
    cache cache.CacheManager,
) *UnifiedAnalyticsService {
    return &UnifiedAnalyticsService{
        repo:  repo,
        cache: cache,
    }
}

func (s *UnifiedAnalyticsService) GetSummary(ctx context.Context, tenantID string) (*models.UnifiedSummary, error) {
    cacheKey := "analytics:unified:summary"
    start := time.Now()

    // 1. Try cache first
    if cached, found := s.cache.Get(tenantID, cacheKey); found {
        log.Debug().
            Str("tenant_id", tenantID).
            Str("cache_key", cacheKey).
            Bool("cache_hit", true).
            Dur("duration", time.Since(start)).
            Msg("Analytics summary cache hit")

        return cached.(*models.UnifiedSummary), nil
    }

    // 2. Cache miss → Use singleflight
    log.Debug().
        Str("tenant_id", tenantID).
        Str("cache_key", cacheKey).
        Bool("cache_hit", false).
        Msg("Analytics summary cache miss, querying database")

    sfKey := fmt.Sprintf("%s:%s", tenantID, cacheKey)
    result, err, shared := s.singleflight.Do(sfKey, func() (interface{}, error) {
        // Query database/materialized view
        summary, err := s.repo.GetUnifiedSummary(ctx, tenantID)
        if err != nil {
            return nil, err
        }

        // Cache result
        if err := s.cache.Set(tenantID, cacheKey, summary, CacheTTLAnalytics); err != nil {
            log.Warn().
                Err(err).
                Str("tenant_id", tenantID).
                Msg("Failed to cache analytics summary")
        }

        return summary, nil
    })

    if shared {
        log.Debug().
            Str("tenant_id", tenantID).
            Msg("Request deduplicated via singleflight")
    }

    if err != nil {
        return nil, err
    }

    log.Debug().
        Str("tenant_id", tenantID).
        Dur("duration", time.Since(start)).
        Msg("Analytics summary fetched from database")

    return result.(*models.UnifiedSummary), nil
}

// InvalidateCache clears all analytics caches for a tenant
func (s *UnifiedAnalyticsService) InvalidateCache(tenantID string) {
    patterns := []string{
        "analytics:unified:*",
        "analytics:ml:*",
        "analytics:ads:*",
    }

    for _, pattern := range patterns {
        if err := s.cache.DeletePattern(tenantID, pattern); err != nil {
            log.Error().
                Err(err).
                Str("tenant_id", tenantID).
                Str("pattern", pattern).
                Msg("Failed to invalidate cache pattern")
        }
    }

    log.Info().
        Str("tenant_id", tenantID).
        Msg("Analytics cache invalidated")
}
```

### Example 3: Sync Handler with Invalidation

```go
// backend/internal/handlers/shopee/sync_handler.go

package shopee

import (
    "github.com/gofiber/fiber/v2"
    "github.com/omni/backend/internal/services/analytics"
    "github.com/omni/backend/internal/services/shopee"
)

type SyncHandler struct {
    syncService      *shopee.SyncService
    analyticsService *analytics.UnifiedAnalyticsService
}

func (h *SyncHandler) SyncOrders(c *fiber.Ctx) error {
    tenantID, ok := c.Locals("tenant_id").(string)
    if !ok || tenantID == "" {
        return c.Status(401).JSON(fiber.Map{
            "success": false,
            "error":   "Missing tenant_id - authentication required",
        })
    }

    // 1. Perform sync (NO CACHE - user action)
    result, err := h.syncService.SyncOrders(c.Context(), tenantID)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "success": false,
            "error":   err.Error(),
        })
    }

    // 2. Invalidate analytics cache (new orders = stale analytics)
    h.analyticsService.InvalidateCache(tenantID)

    // 3. Return fresh result
    return c.JSON(fiber.Map{
        "success": true,
        "data":    result,
        "message": "Orders synced successfully, analytics cache refreshed",
    })
}
```

---

## Cache Invalidation Strategy

### Strategy 1: Time-Based (TTL)

```go
// Automatic expiration - no manual cleanup needed
cache.Set(tenantID, "analytics:summary", data, 60*time.Second)

// Item expires automatically after 60 seconds
// Next request triggers fresh fetch
```

**Pros:** Simple, automatic, predictable  
**Cons:** May serve stale data until expiry

### Strategy 2: Event-Based (Manual Invalidation)

```go
// Invalidate on UPDATE/DELETE operations
func (s *ProductService) UpdateProduct(ctx context.Context, tenantID string, product *Product) error {
    // 1. Update database
    if err := s.repo.Update(ctx, product); err != nil {
        return err
    }

    // 2. Invalidate specific cache
    s.cache.Delete(tenantID, fmt.Sprintf("product:%d", product.ID))

    // 3. Invalidate related caches
    s.cache.DeletePattern(tenantID, "products:list:*")

    return nil
}
```

**Pros:** Immediate consistency  
**Cons:** Must remember to invalidate all related caches

### Strategy 3: Pattern-Based (Bulk Invalidation)

```go
// After sync, invalidate all analytics caches
func InvalidateAnalyticsCache(cache cache.CacheManager, tenantID string) {
    patterns := []string{
        "analytics:*",
        "ml:*",
        "ads:*",
    }

    for _, pattern := range patterns {
        cache.DeletePattern(tenantID, pattern)
    }
}
```

### Invalidation Event Matrix

| Event                  | Caches to Invalidate              |
| ---------------------- | --------------------------------- |
| Order Sync Complete    | `analytics:*`, `ml:*`             |
| Product Sync Complete  | `products:*`, `analytics:*`       |
| Product Update Webhook | `product:{id}`, `products:list:*` |
| Order Update Webhook   | `analytics:*`                     |
| Settings Save          | `settings`                        |
| Config Save            | `config:{platform}:{key}`         |

---

## Testing Checklist

### Unit Tests Required

```go
// Test cache hit
func TestAnalyticsService_GetSummary_CacheHit(t *testing.T) {
    // Setup
    mockRepo := new(MockAnalyticsRepository)
    cache := cache.New(5*time.Minute, 1*time.Minute)
    service := NewUnifiedAnalyticsService(mockRepo, cache)

    // Pre-populate cache
    cache.Set("test_tenant", "analytics:unified:summary", &expectedSummary, 60*time.Second)

    // Execute
    result, err := service.GetSummary(context.Background(), "test_tenant")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, &expectedSummary, result)
    mockRepo.AssertNotCalled(t, "GetUnifiedSummary")  // DB should NOT be called
}

// Test cache miss
func TestAnalyticsService_GetSummary_CacheMiss(t *testing.T) {
    mockRepo := new(MockAnalyticsRepository)
    mockRepo.On("GetUnifiedSummary", mock.Anything, "test_tenant").Return(&expectedSummary, nil)

    cache := cache.New(5*time.Minute, 1*time.Minute)
    service := NewUnifiedAnalyticsService(mockRepo, cache)

    result, err := service.GetSummary(context.Background(), "test_tenant")

    assert.NoError(t, err)
    assert.Equal(t, &expectedSummary, result)
    mockRepo.AssertCalled(t, "GetUnifiedSummary", mock.Anything, "test_tenant")
}

// Test multi-tenant isolation
func TestCache_MultiTenantIsolation(t *testing.T) {
    cache := cache.New(5*time.Minute, 1*time.Minute)

    // Tenant A sets cache
    cache.Set("tenant_a", "data", "value_a", 60*time.Second)

    // Tenant B should NOT see Tenant A's cache
    _, found := cache.Get("tenant_b", "data")
    assert.False(t, found, "Tenant B should not see Tenant A's cache")

    // Tenant A should see own cache
    value, found := cache.Get("tenant_a", "data")
    assert.True(t, found)
    assert.Equal(t, "value_a", value)
}

// Test cache invalidation
func TestAnalyticsService_InvalidateCache(t *testing.T) {
    mockRepo := new(MockAnalyticsRepository)
    cache := cache.New(5*time.Minute, 1*time.Minute)
    service := NewUnifiedAnalyticsService(mockRepo, cache)

    // Pre-populate cache
    cache.Set("test_tenant", "analytics:unified:summary", &summary1, 60*time.Second)

    // Invalidate
    service.InvalidateCache("test_tenant")

    // Cache should be empty
    _, found := cache.Get("test_tenant", "analytics:unified:summary")
    assert.False(t, found, "Cache should be invalidated")
}
```

### Manual QA Checklist

- [ ] Analytics dashboard loads faster on second request
- [ ] Cache hit rate visible in logs (`cache_hit: true/false`)
- [ ] User clicks "Sync" → Data refreshes (cache invalidated)
- [ ] Tenant A cannot see Tenant B's cached data
- [ ] Server restart → Cache regenerates (no stale data served)
- [ ] 100 concurrent requests → Only 1 DB query (singleflight works)

---

## Monitoring & Metrics

### Logging Pattern

```go
log.Debug().
    Str("tenant_id", tenantID).
    Str("cache_key", cacheKey).
    Bool("cache_hit", true).
    Dur("duration", time.Since(start)).
    Msg("Cache access")
```

**Expected Log Output:**

```json
{
  "level": "debug",
  "tenant_id": "yumna_bertigamart",
  "cache_key": "analytics:unified:summary",
  "cache_hit": true,
  "duration": "2.5ms",
  "message": "Cache access"
}
```

### Metrics to Track

| Metric             | Target  | Alert Threshold |
| ------------------ | ------- | --------------- |
| Cache Hit Rate     | > 70%   | < 50%           |
| Cache Miss Latency | < 500ms | > 2000ms        |
| Cache Error Rate   | < 1%    | > 5%            |
| Memory Usage       | < 100MB | > 200MB         |

### Metrics Endpoint (Optional)

```go
func (h *MetricsHandler) GetCacheMetrics(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "cache_hit_rate": h.cache.GetHitRate(),
        "total_items":    h.cache.Count(),
        "memory_bytes":   h.cache.MemoryUsage(),
    })
}
```

---

## AI Agent Workflow Rules

### ✅ MUST DO

1. **Always include `tenant_id` in cache keys**

   ```go
   cacheKey := fmt.Sprintf("tenant:%s:resource:%s", tenantID, resourceID)
   ```

2. **Validate tenant_id before cache operations**

   ```go
   if tenantID == "" {
       return c.Status(401).JSON(fiber.Map{"error": "Missing tenant_id"})
   }
   ```

3. **Use graceful degradation**

   ```go
   if err := cache.Set(...); err != nil {
       log.Warn().Err(err).Msg("Cache set failed")
       // Continue - don't fail request
   }
   ```

4. **Invalidate cache on mutations**

   ```go
   // After UPDATE/DELETE, always invalidate related caches
   cache.Delete(tenantID, key)
   cache.DeletePattern(tenantID, "related:*")
   ```

5. **Use `context.Context` in all operations**

   ```go
   func GetData(ctx context.Context, tenantID string) (*Data, error)
   ```

6. **Log cache hits/misses**

   ```go
   log.Debug().Bool("cache_hit", found).Msg("Cache access")
   ```

7. **Use singleflight for expensive queries**
   ```go
   result, err, _ := s.singleflight.Do(key, fetchFunc)
   ```

### ❌ MUST NOT DO

1. **Never cache without tenant isolation**

   ```go
   // ❌ WRONG
   cache.Set("global_key", data)

   // ✅ CORRECT
   cache.Set(tenantID, "key", data)
   ```

2. **Never cache user-triggered actions**

   ```go
   // ❌ WRONG: Caching CREATE operation
   func CreateOrder(...) {
       if cached := cache.Get(...); cached != nil {
           return cached  // NO! User expects fresh creation
       }
   }
   ```

3. **Never fail requests due to cache errors**

   ```go
   // ❌ WRONG
   if err := cache.Set(...); err != nil {
       return err  // Request fails!
   }

   // ✅ CORRECT
   if err := cache.Set(...); err != nil {
       log.Warn().Err(err).Msg("Cache failed")
       // Continue without cache
   }
   ```

4. **Never cache sensitive data**

   ```go
   // ❌ NEVER
   cache.Set(tenantID, "password", password)
   cache.Set(tenantID, "api_token", token)
   ```

5. **Never use global cache keys**
   ```go
   // ❌ WRONG - data leak risk
   cache.Set("", "analytics:summary", data)
   ```

### Code Review Checklist

Before merging cache-related code:

- [ ] All cache keys include `tenant_id`
- [ ] No user actions are cached
- [ ] Graceful degradation implemented
- [ ] Cache invalidation on UPDATE/DELETE
- [ ] `context.Context` used everywhere
- [ ] Structured logging added
- [ ] Tests cover cache hit/miss scenarios
- [ ] Tests cover multi-tenant isolation
- [ ] No sensitive data in cache
- [ ] TTL matches data classification
- [ ] Singleflight used for expensive queries

---

## Migration to Redis

### When to Migrate

**Trigger Conditions:**

- Deploying to multiple backend servers
- Need cache persistence across restarts
- Cache hit rate > 80% (cache is critical)
- Need distributed cache invalidation

### Step 1: Add Dependency

```bash
cd backend
go get github.com/redis/go-redis/v9
```

### Step 2: Create Redis Implementation

```go
// backend/internal/services/cache/redis_cache.go

package cache

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
    return &RedisCache{
        client: redis.NewClient(&redis.Options{
            Addr:            addr,
            Password:        password,
            DB:              db,
            ReadTimeout:     3 * time.Second,
            WriteTimeout:    3 * time.Second,
            PoolSize:        100,
            MinIdleConns:    10,
        }),
    }
}

func (c *RedisCache) Get(tenantID, key string) (interface{}, bool) {
    ctx := context.Background()
    fullKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)

    val, err := c.client.Get(ctx, fullKey).Result()
    if err == redis.Nil {
        return nil, false
    }
    if err != nil {
        return nil, false
    }

    var result interface{}
    if err := json.Unmarshal([]byte(val), &result); err != nil {
        return nil, false
    }

    return result, true
}

func (c *RedisCache) Set(tenantID, key string, value interface{}, ttl time.Duration) error {
    ctx := context.Background()
    fullKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)

    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return c.client.Set(ctx, fullKey, data, ttl).Err()
}

func (c *RedisCache) Delete(tenantID, key string) error {
    ctx := context.Background()
    fullKey := fmt.Sprintf("tenant:%s:%s", tenantID, key)
    return c.client.Del(ctx, fullKey).Err()
}

func (c *RedisCache) DeletePattern(tenantID, pattern string) error {
    ctx := context.Background()
    fullPattern := fmt.Sprintf("tenant:%s:%s", tenantID, pattern)

    iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()
    for iter.Next(ctx) {
        c.client.Del(ctx, iter.Val())
    }

    return iter.Err()
}
```

### Step 3: Update Docker Compose

```yaml
# docker-compose.yml
services:
  redis:
    image: redis:7-alpine
    container_name: omni_redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    networks:
      - omni_network

  backend:
    environment:
      - USE_REDIS=true
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis

volumes:
  redis_data:
```

### Step 4: Switch via Configuration

```go
// backend/internal/app/app.go

func initCache(config *Config) cache.CacheManager {
    if config.UseRedis {
        return cache.NewRedisCache(
            config.RedisAddr,
            config.RedisPassword,
            config.RedisDB,
        )
    }

    return cache.New(
        5 * time.Minute,   // default TTL
        10 * time.Minute,  // cleanup interval
    )
}
```

---

## Conclusion

### Summary

**What We Cache:**

- ✅ Analytics dashboards (60s TTL)
- ✅ Product listings (5min TTL)
- ✅ Global configs (1h TTL)
- ✅ Tenant settings (15min TTL)

**What We DON'T Cache:**

- ❌ User-triggered actions (order CRUD, sync buttons)
- ❌ Real-time data (webhooks, live stock)
- ❌ Sensitive data (passwords, tokens)

**Key Patterns:**

- Cache-Aside with graceful degradation
- Singleflight for stampede prevention
- Event-based invalidation on mutations
- Multi-tenant isolation via key prefixes

### Success Metrics

| Metric              | Target  |
| ------------------- | ------- |
| Cache Hit Rate      | > 70%   |
| Dashboard Load Time | < 500ms |
| DB Query Reduction  | 60-70%  |
| Cache-Related Bugs  | Zero    |

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-30  
**Next Review**: After Phase 1 implementation complete
