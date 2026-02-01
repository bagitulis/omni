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

	for range ticker.C {
		rl.mu.Lock()
		threshold := time.Now().Add(-10 * time.Minute)
		for key, b := range rl.buckets {
			if b.lastRefill.Before(threshold) {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
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

// AuthRateLimitMiddleware creates strict rate limiter for authentication endpoints
// Uses IP as key with lower limits to prevent brute force attacks
func AuthRateLimitMiddleware() gin.HandlerFunc {
	// 5 attempts per minute per IP for auth endpoints
	limiter := NewRateLimiter(1, time.Minute/5, 5)
	return RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		return "auth:" + c.ClientIP()
	})
}

// APIRateLimitMiddleware creates rate limiter for general API endpoints
// Uses user ID with higher limits for authenticated users
func APIRateLimitMiddleware() gin.HandlerFunc {
	// 100 requests per minute for authenticated users
	return UserRateLimitMiddleware(100, time.Minute, 100)
}
