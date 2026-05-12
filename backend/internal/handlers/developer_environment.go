package handlers

import (
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
)

// GetEnvironmentInfo returns non-sensitive environment information for the developer panel.
// GET /api/dev/environment
func (h *DeveloperHandler) GetEnvironmentInfo(c *gin.Context) {
	ctx := c.Request.Context()

	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "sqlite"
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	buildTime := os.Getenv("BUILD_TIME")
	if buildTime == "" {
		buildTime = "unknown"
	}

	nodeEnv := os.Getenv("NODE_ENV")
	if nodeEnv == "" {
		nodeEnv = "development"
	}

	apiBaseURL := c.Request.Host
	if apiBaseURL == "" {
		apiBaseURL = "localhost"
	}

	devLoginEnabled := "no"
	if environment == "development" || c.Request.Host == "localhost" {
		devLoginEnabled = "yes"
	}

	// Count active tenants
	var activeTenantsCount int
	tenants, err := h.tenantService.GetAvailableTenants(ctx)
	if err == nil {
		activeTenantsCount = len(tenants)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"go_version":           runtime.Version(),
			"environment":          environment,
			"db_driver":            dbDriver,
			"version":              version,
			"build_time":           buildTime,
			"node_env":             nodeEnv,
			"api_base_url":         apiBaseURL,
			"dev_login_enabled":    devLoginEnabled,
			"active_tenants_count": activeTenantsCount,
		},
	})
}
