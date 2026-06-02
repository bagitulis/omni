package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// GetTokenStatus returns token status for all platforms
// GET /api/oauth/status
func (h *OAuthHandler) GetTokenStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}

	repo := repositories.NewTenantPlatformConfigRepository(db)
	result := h.buildTokenStatusMap(c, repo)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// buildTokenStatusMap builds token status for all platforms
func (h *OAuthHandler) buildTokenStatusMap(c *gin.Context, repo *repositories.TenantPlatformConfigRepository) map[string]gin.H {
	nowMs := time.Now().UnixMilli()
	result := make(map[string]gin.H)
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}

	for _, platform := range platforms {
		configMap, err := repo.GetAllConfigByPlatform(c.Request.Context(), platform)
		if err != nil || configMap["accessToken"] == "" {
			result[platform] = buildNotConfiguredStatus()
			continue
		}

		result[platform] = buildPlatformTokenStatus(configMap, nowMs)
	}

	return result
}

// buildNotConfiguredStatus returns the default status for an unconfigured platform
func buildNotConfiguredStatus() gin.H {
	return gin.H{
		"isExpired":             true,
		"expiresAt":             nil,
		"refreshTokenExpiresAt": nil,
		"status":                "not_configured",
		"valid":                 false,
	}
}

// buildPlatformTokenStatus builds token status from config map
func buildPlatformTokenStatus(configMap map[string]string, nowMs int64) gin.H {
	tokenExpiryMs, err := parseTimestampMs(configMap["tokenExpiry"])
	if err != nil {
		return gin.H{
			"isExpired":             true,
			"expiresAt":             nil,
			"refreshTokenExpiresAt": nil,
			"status":                "invalid_config",
			"valid":                 false,
		}
	}

	refreshTokenExpiryMs, _ := parseTimestampMs(configMap["refreshTokenExpiry"])

	isExpired := tokenExpiryMs > 0 && tokenExpiryMs < nowMs
	expiresSoon := tokenExpiryMs > 0 && tokenExpiryMs < nowMs+(24*60*60*1000)

	status := "valid"
	if isExpired {
		status = "expired"
	} else if expiresSoon {
		status = "expiring"
	}

	return gin.H{
		"isExpired":             isExpired,
		"expiresAt":             formatTimestampMs(tokenExpiryMs),
		"refreshTokenExpiresAt": formatTimestampMs(refreshTokenExpiryMs),
		"status":                status,
		"valid":                 !isExpired,
		"shopId":                configMap["shopId"],
		"shopName":              configMap["shopName"],
	}
}

// parseTimestampMs parses a millisecond timestamp string
func parseTimestampMs(s string) (int64, error) {
	if s == "" {
		return 0, fmt.Errorf("invalid timestamp: empty string")
	}
	var ms int64
	n, _ := fmt.Sscanf(s, "%d", &ms)
	if n != 1 {
		return 0, fmt.Errorf("invalid timestamp: %s", s)
	}
	return ms, nil
}

// formatTimestampMs formats a millisecond timestamp to RFC3339 string
func formatTimestampMs(ms int64) string {
	if ms <= 0 {
		return ""
	}
	return time.UnixMilli(ms).Format(time.RFC3339)
}

// RefreshAllTokens refreshes tokens for all platforms
// POST /api/oauth/refresh-all
func (h *OAuthHandler) RefreshAllTokens(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	// Use tenant DB with adapter (NOT system DB + PlatformConfigRepository)
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}

	repo := repositories.NewPlatformConfigAdapter(db)
	configs, err := repo.FindByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch platform configs: " + err.Error(),
		})
		return
	}

	results := make(map[string]gin.H)
	nowMs := time.Now().UnixMilli()
	for _, cfg := range configs {
		// ExpiresAt from adapter is in milliseconds (from tokenExpiry key)
		if cfg.ExpiresAt > 0 && cfg.ExpiresAt < nowMs {
			results[cfg.Platform] = gin.H{
				"success": false,
				"error":   "Token expired - manual re-authorization required",
			}
		} else {
			results[cfg.Platform] = gin.H{
				"success": true,
				"message": "Token is still valid",
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}
