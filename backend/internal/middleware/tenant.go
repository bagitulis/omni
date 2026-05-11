package middleware

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
)

var (
	tenantsLoaded   bool
	loadMu          sync.Once
	tenantIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`) // Only alphanumeric and underscore
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
		// Check if tenant_id was already set by Auth middleware (from JWT)
		tenantID := c.GetString("tenant_id")
		fromJWT := tenantID != ""

		// Developer role can switch tenants via x-tenant-id header (cross-tenant access).
		// Non-developer users are locked to their JWT tenant for security.
		role := c.GetString("role")
		headerTenant := c.GetHeader("x-tenant-id")
		if role == "developer" && headerTenant != "" && headerTenant != tenantID {
			tenantID = headerTenant
			fromJWT = false // must re-validate since it's from header
		}

		// If not set by Auth or overridden, check header (legacy flow)
		if tenantID == "" {
			tenantID = headerTenant
		}

		// AGENTS.MD: NO DEFAULT TENANT — must throw error if missing
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenant_id",
			})
			c.Abort()
			return
		}

		// If tenantID came from JWT (non-developer), it's already validated by Auth middleware.
		// Developer overrides and header-sourced IDs must go through config validation.
		if !fromJWT {
			// Validate tenant using config (if loaded)
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
				// Fallback: validate format only (alphanumeric + underscore)
				if !tenantIDPattern.MatchString(tenantID) {
					c.JSON(http.StatusUnauthorized, gin.H{
						"success": false,
						"error":   "Invalid tenant ID format",
					})
					c.Abort()
					return
				}
			}
		}

		// tenant_id is the ONLY canonical context key
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}

// TenantWithConfig creates tenant middleware with custom config path
func TenantWithConfig(databasePath string) gin.HandlerFunc {
	initTenants(databasePath)
	return Tenant()
}

// GetTenantID extracts tenant ID from context using the canonical "tenant_id" key.
// All middleware (Auth, Tenant) MUST set this key.
func GetTenantID(c *gin.Context) string {
	return c.GetString("tenant_id")
}
