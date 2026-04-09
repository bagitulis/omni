package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	// Build allowed origins from environment + defaults
	allowedOrigins := make(map[string]bool)
	isProd := os.Getenv("GO_ENV") == "production"

	// Default origins (always allowed)
	defaults := []string{
		"http://localhost:5173",
		"http://localhost:5174",
		"http://localhost:3000",
		"http://localhost:80",
		"http://localhost",
		"https://yndigital.my.id",
		"https://www.yndigital.my.id",
		"http://yndigital.my.id",
		"http://www.yndigital.my.id",
	}
	for _, origin := range defaults {
		allowedOrigins[origin] = true
	}

	// Add origins from CORS_ORIGINS env variable
	corsEnv := os.Getenv("CORS_ORIGINS")
	if corsEnv != "" {
		for _, origin := range strings.Split(corsEnv, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins[origin] = true
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		isAllowed := allowedOrigins[origin] || origin == ""

		// In non-production: allow any localhost/127.0.0.1 origin (any port)
		// This prevents breakage when Vite picks a different dev port
		if !isAllowed && !isProd && origin != "" {
			isAllowed = strings.HasPrefix(origin, "http://localhost") ||
				strings.HasPrefix(origin, "http://127.0.0.1")
		}

		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, x-tenant-id, x-csrf-token, Cache-Control, Pragma")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Security headers (defense-in-depth, also set by nginx)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
