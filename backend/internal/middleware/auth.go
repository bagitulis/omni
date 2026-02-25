package middleware

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/utils"
)

var jwtService *utils.JWTService

func init() {
	secret := os.Getenv("JWT_SECRET")
	env := os.Getenv("GO_ENV")

	if secret == "" {
		if env == "production" {
			// SECURITY: Never allow empty JWT_SECRET in production
			log.Fatal().Msg("❌ FATAL: JWT_SECRET is required in production - server cannot start")
			os.Exit(1)
		}
		// Development only - use a long enough key for testing
		log.Warn().Msg("JWT_SECRET not set, using development key (NOT FOR PRODUCTION)")
		secret = "dev-secret-key-minimum-32-chars-for-security"
	} else {
		// SECURITY: Don't log any part of the secret
		log.Info().Msg("JWT_SECRET loaded successfully")
	}

	// Validate minimum key length
	if len(secret) < 32 {
		if env == "production" {
			log.Fatal().Msg("❌ FATAL: JWT_SECRET must be at least 32 characters")
			os.Exit(1)
		}
		log.Warn().Msg("JWT_SECRET should be at least 32 characters")
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
			// SECURITY: Never log tokens or their parts
			log.Warn().Err(err).Str("path", c.Request.URL.Path).Msg("JWT validation failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Unauthorized", // Generic message - don't expose internal details
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
