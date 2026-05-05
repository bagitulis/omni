package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements token bucket algorithm for rate limiting
type RateLimiter struct {
	buckets    map[string]*bucket
	mu         sync.Mutex
	rate       int           // tokens per interval
	interval   time.Duration // refill interval
	bucketSize int           // max tokens in bucket
	stopCh     chan struct{}
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per interval
// interval: time period for rate calculation
// bucketSize: maximum burst size
func NewRateLimiter(rate int, interval time.Duration, bucketSize int) *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*bucket),
		rate:       rate,
		interval:   interval,
		bucketSize: bucketSize,
		stopCh:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// AllowWithInfo checks if request is allowed and returns remaining tokens and retry-after duration
func (rl *RateLimiter) AllowWithInfo(key string) (allowed bool, remaining int, retryAfter time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     rl.bucketSize,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = b
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed/rl.interval) * rl.rate

	if tokensToAdd > 0 {
		b.tokens = min(b.tokens+tokensToAdd, rl.bucketSize)
		b.lastRefill = now
	}

	// Check if request is allowed
	if b.tokens > 0 {
		b.tokens--
		return true, b.tokens, 0
	}

	// Calculate retry-after
	retryAfter = rl.interval - (now.Sub(b.lastRefill) % rl.interval)
	return false, 0, retryAfter
}

// Allow checks if request is allowed for given key (backward compatible)
func (rl *RateLimiter) Allow(key string) bool {
	allowed, _, _ := rl.AllowWithInfo(key)
	return allowed
}

// cleanup removes old buckets periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			threshold := time.Now().Add(-10 * time.Minute)
			for key, b := range rl.buckets {
				if b.lastRefill.Before(threshold) {
					delete(rl.buckets, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
			}

			// Stop stops the cleanup goroutine
		func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// RateLimitMiddleware creates Gin middleware for rate limiting
func RateLimitMiddleware(limiter *RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)

		if !limiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded",
				"message": "Too many requests, please try again later",
			})
			return
		}

		c.Next()
	}
}

// RateLimitMiddlewareWithHeaders creates Gin middleware with rate limit headers
func RateLimitMiddlewareWithHeaders(limiter *RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		allowed, remaining, retryAfter := limiter.AllowWithInfo(key)

		// Always set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.bucketSize))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"error":       "Rate limit exceeded",
				"message":     "Too many requests, please try again later",
				"retry_after": int(retryAfter.Seconds()) + 1,
			})
			return
		}

		c.Next()
	}
}

// IPRateLimitMiddleware creates rate limiter using client IP as key
func IPRateLimitMiddleware(rate int, interval time.Duration, bucketSize int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, interval, bucketSize)
	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// IPRateLimitMiddlewareWithHeaders creates rate limiter with headers using client IP as key
func IPRateLimitMiddlewareWithHeaders(rate int, interval time.Duration, bucketSize int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, interval, bucketSize)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// TenantRateLimitMiddleware creates rate limiter using tenant ID as key
func TenantRateLimitMiddleware(rate int, interval time.Duration, bucketSize int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, interval, bucketSize)
	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
		tenantID := GetTenantID(c)
		if tenantID == "" {
			return c.ClientIP()
		}
		return tenantID
	})
}

// UserRateLimitMiddleware creates rate limiter using user ID from context
func UserRateLimitMiddleware(rate int, interval time.Duration, bucketSize int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, interval, bucketSize)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		// Try to get user ID from context (set by auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(string); ok && id != "" {
				return id
			}
		}
		// Fallback to tenant + IP for unauthenticated or missing user
		tenantID := GetTenantID(c)
		return tenantID + ":" + c.ClientIP()
	})
}

// ===========================================
// RATE LIMIT CONFIGURATIONS FOR MARKETPLACE
// ===========================================
//
// Marketplace context considerations:
// - High-frequency API sync (Shopee/Lazada/TikTok) - hundreds of calls per sync
// - Multiple users from same IP (office, shared NAT) - up to 15-20 users per IP
// - Background sync every 5-15 minutes
// - Mobile + Desktop + Tablet simultaneous access
// - Flash sales: 200+ orders in 5 minutes across 3 platforms
//
// Rate limits are designed to:
// 1. Prevent brute force attacks on auth (but not block legitimate office users)
// 2. Allow high-throughput marketplace operations
// 3. Protect against runaway scripts/bugs
// 4. Handle flash sale webhook bursts

// AuthRateLimitMiddleware creates rate limiter for authentication endpoints
// Balanced to prevent brute force while allowing legitimate multi-device usage
//
// Config: 30 attempts per minute per IP, burst of 50
// - Office with 15 users on 1 IP can all login during morning rush
// - Still prevents automated brute force (need 30+ min for 1000 attempts)
// - Burst allows quick retries after password reset
func AuthRateLimitMiddleware() gin.HandlerFunc {
	// 30 attempts per minute per IP, burst of 50
	limiter := NewRateLimiter(30, time.Minute, 50)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		return "auth:" + c.ClientIP()
	})
}

// LoginRateLimitMiddleware creates rate limiter specifically for login endpoint
// Per-IP rate limiting with reasonable limits for shared office environments
//
// Config: 10 attempts per minute per IP, burst of 20
// - Prevents targeted brute force attacks
// - Allows office users (NAT) to login without blocking each other
// - Combined with account lockout (10 failed = 15min lock) provides layered defense
func LoginRateLimitMiddleware() gin.HandlerFunc {
	// 10 attempts per minute per IP
	limiter := NewRateLimiter(10, time.Minute, 20)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		return "login:" + c.ClientIP()
	})
}

// APIRateLimitMiddleware creates rate limiter for general API endpoints
// High limits suitable for marketplace sync operations
//
// Config: 1000 requests per minute per user, burst of 2000
// - Sync operations can make 100-500 calls per batch
// - Users may run multiple syncs (orders, products, inventory)
// - Burst allows initial sync after downtime
func APIRateLimitMiddleware() gin.HandlerFunc {
	// 1000 requests per minute for authenticated users, burst of 2000
	return UserRateLimitMiddleware(1000, time.Minute, 2000)
}

// SyncRateLimitMiddleware creates rate limiter for platform sync endpoints
// Very high limits for background sync operations
//
// Config: 5000 requests per minute per tenant
// - Platform syncs (Shopee/Lazada/TikTok) are API-intensive
// - Single sync can fetch hundreds of orders with pagination
// - Multiple platforms may sync simultaneously
func SyncRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(5000, time.Minute, 10000)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		// Rate limit by tenant to prevent one tenant from affecting others
		tenantID := GetTenantID(c)
		if tenantID == "" {
			return "sync:" + c.ClientIP()
		}
		return "sync:" + tenantID
	})
}

// WebhookRateLimitMiddleware creates rate limiter for incoming webhooks
// Platform webhooks can come in bursts during high-activity periods
//
// Config: 1000 requests per minute per source IP, burst of 2000
// Flash sale math:
// - 200 orders in 5 minutes
// - × 3 platforms (Shopee, Lazada, TikTok)
// - × 3 webhook types (order.create, payment.confirm, inventory.update)
// = 1,800 webhooks in 5 minutes = 360/min per platform
// 1000/min with burst 2000 handles this comfortably
func WebhookRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(1000, time.Minute, 2000)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		return "webhook:" + c.ClientIP()
	})
}

// PublicRateLimitMiddleware creates rate limiter for public/unauthenticated endpoints
// Stricter limits for endpoints that don't require auth
//
// Config: 30 requests per minute per IP
// - Health checks, public info endpoints
// - Prevents scraping and enumeration attacks
func PublicRateLimitMiddleware() gin.HandlerFunc {
	limiter := NewRateLimiter(30, time.Minute, 60)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		return "public:" + c.ClientIP()
	})
}
