package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORS handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Allowed origins
		allowedOrigins := map[string]bool{
			"http://localhost:5173":      true,
			"http://localhost:80":        true,
			"https://yndigital.my.id":    true,
			"https://www.yndigital.my.id": true,
		}

		// Check if origin is allowed
		if allowedOrigins[origin] || origin == "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, x-tenant-id, x-csrf-token")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
