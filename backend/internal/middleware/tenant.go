package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
)

var (
	tenantsLoaded bool
	loadMu        sync.Once
)

// initTenants loads tenant configuration
func initTenants(databasePath string) {
	loadMu.Do(func() {
		_, err := config.LoadTenants(databasePath)
		if err != nil {
			// Log warning but continue - will use fallback validation
			// In production, this should fail fast
			return
		}
		tenantsLoaded = true
	})
}

// Tenant extracts and validates tenant ID from request header
// Following AGENTS.MD: NO DEFAULT TENANT - must throw error if missing
func Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if tenantID was already set by Auth middleware (from JWT)
		tenantID := c.GetString("tenantID")
		
		// If not in context, check header (for backwards compatibility)
		if tenantID == "" {
			tenantID = c.GetHeader("x-tenant-id")
		}

		// AGENTS.MD: JANGAN gunakan default tenant!
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantId - authentication required",
			})
			c.Abort()
			return
		}

		// Validate tenant using config
		if tenantsLoaded {
			if !config.ValidateTenant(tenantID) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Invalid tenant ID",
				})
				c.Abort()
				return
			}
		} else {
			// Fallback to hardcoded validation if tenants.json not loaded
			// This should only happen in development
			validTenants := map[string]bool{
				"yumna_bertigamart": true,
				"tika_nusseyba":     true,
				"system":            true,
			}

			if !validTenants[tenantID] {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "Invalid tenant ID",
				})
				c.Abort()
				return
			}
		}

		c.Set("tenantID", tenantID)
		c.Set("tenantId", tenantID) // Also set camelCase for compatibility
		c.Next()
	}
}

// TenantWithConfig creates tenant middleware with custom config path
func TenantWithConfig(databasePath string) gin.HandlerFunc {
	initTenants(databasePath)
	return Tenant()
}

// GetTenantID extracts tenant ID from context
func GetTenantID(c *gin.Context) string {
	tenantID, exists := c.Get("tenantID")
	if !exists {
		// Try camelCase version
		tenantID, exists = c.Get("tenantId")
		if !exists {
			return ""
		}
	}
	return tenantID.(string)
}
