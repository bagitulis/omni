package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// GetTokenStatus returns token status for all platforms
// GET /api/oauth/status
func (h *OAuthHandler) GetTokenStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
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
	tokenExpiryMs := parseTimestampMs(configMap["tokenExpiry"])
	refreshTokenExpiryMs := parseTimestampMs(configMap["refreshTokenExpiry"])

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
func parseTimestampMs(s string) int64 {
	var ms int64
	if s != "" {
		fmt.Sscanf(s, "%d", &ms)
	}
	return ms
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
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	configs, err := h.platformRepo.FindByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch platform configs: " + err.Error(),
		})
		return
	}

	results := make(map[string]gin.H)
	now := time.Now().Unix()
	for _, cfg := range configs {
		if cfg.ExpiresAt < now {
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
