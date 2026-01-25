package middleware

import (
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

// Allow checks if request is allowed for given key
func (rl *RateLimiter) Allow(key string) bool {
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
	tokensToAdd := int(elapsed / rl.interval) * rl.rate

	if tokensToAdd > 0 {
		b.tokens = min(b.tokens+tokensToAdd, rl.bucketSize)
		b.lastRefill = now
	}

	// Check if request is allowed
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
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

// IPRateLimitMiddleware creates rate limiter using client IP as key
func IPRateLimitMiddleware(rate int, interval time.Duration, bucketSize int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, interval, bucketSize)
	return RateLimitMiddleware(limiter, func(c *gin.Context) string {
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
