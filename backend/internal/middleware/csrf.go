package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	csrfTokenLength = 32
	csrfCookieName  = "csrf_token"
	csrfHeaderName  = "x-csrf-token"
	csrfTokenTTL    = 24 * time.Hour
)

// CSRFExemptPaths contains paths that don't require CSRF validation
// Reference: AGENTS.MD Section 7.2
var CSRFExemptPaths = []string{
	"/api/webhooks/",
	"/api/platform-auth/",
	"/api/health",
	"/api/n8n/",
	"/api/auth/login",
	"/api/auth/register",
	"/api/csrf-token",
	"/api/status",
}

// CSRFTokenStore stores valid CSRF tokens with expiration
type CSRFTokenStore struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

var tokenStore = &CSRFTokenStore{
	tokens: make(map[string]time.Time),
}

// GenerateCSRFToken creates a new cryptographically secure token
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(bytes)

	// Store token with expiration
	tokenStore.mu.Lock()
	tokenStore.tokens[token] = time.Now().Add(csrfTokenTTL)
	tokenStore.mu.Unlock()

	return token, nil
}

// ValidateCSRFToken checks if token is valid and not expired
func ValidateCSRFToken(token string) bool {
	tokenStore.mu.RLock()
	expiry, exists := tokenStore.tokens[token]
	tokenStore.mu.RUnlock()

	if !exists {
		return false
	}

	return time.Now().Before(expiry)
}

// isExemptPath checks if path is exempt from CSRF validation
func isCSRFExemptPath(path string) bool {
	for _, exempt := range CSRFExemptPaths {
		if strings.HasPrefix(path, exempt) {
			return true
		}
	}
	return false
}

// isSafeMethod checks if HTTP method is safe (doesn't modify state)
func isSafeMethod(method string) bool {
	return method == http.MethodGet ||
		method == http.MethodHead ||
		method == http.MethodOptions
}

// CSRFProtection implements double-submit cookie pattern for CSRF protection
func CSRFProtection() gin.HandlerFunc {
	// Start cleanup goroutine for expired tokens
	go cleanupExpiredTokens()

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip CSRF for exempt paths
		if isCSRFExemptPath(path) {
			c.Next()
			return
		}

		// Skip CSRF for safe methods (GET, HEAD, OPTIONS)
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// For mutating methods (POST, PUT, DELETE, PATCH), validate token
		headerToken := c.GetHeader(csrfHeaderName)
		cookieToken, _ := c.Cookie(csrfCookieName)

		// Double-submit validation: both header and cookie must match
		if headerToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF token missing",
				"message": "CSRF token required in x-csrf-token header",
			})
			return
		}

		// Validate token exists in store and matches cookie
		if !ValidateCSRFToken(headerToken) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF token invalid",
				"message": "Invalid or expired CSRF token",
			})
			return
		}

		// For double-submit pattern, header and cookie should match
		if cookieToken != "" && headerToken != cookieToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF token mismatch",
				"message": "CSRF header token does not match cookie",
			})
			return
		}

		c.Next()
	}
}

// cleanupExpiredTokens removes expired tokens periodically
func cleanupExpiredTokens() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		tokenStore.mu.Lock()
		for token, expiry := range tokenStore.tokens {
			if now.After(expiry) {
				delete(tokenStore.tokens, token)
			}
		}
		tokenStore.mu.Unlock()
	}
}
