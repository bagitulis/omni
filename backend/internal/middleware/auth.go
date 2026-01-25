package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/utils"
)

var jwtService *utils.JWTService

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("⚠️ JWT_SECRET not set, using dev-secret-key")
		secret = "dev-secret-key" // Only for development
	} else {
		log.Printf("✅ JWT_SECRET loaded (first 8 chars: %s...)", secret[:8])
	}
	jwtService = utils.NewJWTService(secret)
}

// Auth validates JWT token from Authorization header
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing authorization header",
			})
			c.Abort()
			return
		}

		// Extract Bearer token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization format",
			})
			c.Abort()
			return
		}

		// Validate JWT
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			tokenPreview := token
			if len(token) > 20 {
				tokenPreview = token[:20]
			}
			log.Printf("❌ JWT validation failed: %v (token prefix: %s...)", err, tokenPreview)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Set user context from JWT claims
		c.Set("userID", claims.UserID)
		c.Set("tenantID", claims.TenantID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) string {
	if v, ok := c.Get("userID"); ok {
		return v.(string)
	}
	return ""
}

// GetRole extracts role from context
func GetRole(c *gin.Context) string {
	if v, ok := c.Get("role"); ok {
		return v.(string)
	}
	return ""
}
