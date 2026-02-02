package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
)

// TokenHandler handles token management endpoints
type TokenHandler struct {
	tokenManager *services.TokenManager
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(tokenManager *services.TokenManager) *TokenHandler {
	return &TokenHandler{tokenManager: tokenManager}
}

// GetTokenStatus gets token status for a platform
func (h *TokenHandler) GetTokenStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantID"})
		return
	}

	platform := c.Param("platform")
	if platform == "" || !ValidatePlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid platform"})
		return
	}

	status, err := h.tokenManager.GetTokenStatus(c.Request.Context(), tenantID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"platform":     status.Platform,
			"shopID":       status.ShopID,
			"isValid":      status.IsValid,
			"needsRefresh": status.NeedsRefresh,
			"expiresAt":    status.ExpiresAt,
		},
	})
}

// GetAllTokenStatus gets token status for all platforms
func (h *TokenHandler) GetAllTokenStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantID"})
		return
	}

	statuses := h.buildAllPlatformStatuses(c, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"data":      statuses,
		"timestamp": FormatISOTimestamp(time.Now()),
	})
}

// RefreshToken refreshes token for a platform
func (h *TokenHandler) RefreshToken(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantID"})
		return
	}

	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Platform is required"})
		return
	}

	newToken, err := h.refreshPlatformToken(c, tenantID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if newToken == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid platform"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"platform": newToken.Platform, "isValid": newToken.IsValid, "expiresAt": newToken.ExpiresAt},
		"message": "Token refreshed successfully",
	})
}

// GetPlatformTokenStatus gets token status for a specific platform (alias route)
func (h *TokenHandler) GetPlatformTokenStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantID"})
		return
	}

	platform := c.Param("platform")
	if !ValidatePlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid platform"})
		return
	}

	status, err := h.tokenManager.GetTokenStatus(c.Request.Context(), tenantID, platform)
	if err != nil {
		// Return success:false when token status retrieval fails
		c.JSON(http.StatusOK, gin.H{
			"success":   false,
			"platform":  platform,
			"data":      gin.H{"platform": platform, "isExpired": true},
			"error":     "Token not found or expired",
			"timestamp": FormatISOTimestamp(time.Now()),
		})
		return
	}

	data := h.buildPlatformStatusData(platform, status)
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"platform":  platform,
		"data":      data,
		"timestamp": FormatISOTimestamp(time.Now()),
	})
}

// RefreshAllTokens refreshes tokens for all platforms
func (h *TokenHandler) RefreshAllTokens(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantID"})
		return
	}

	force := c.Query("force") == "true"
	results := h.refreshAllPlatforms(c, tenantID, force)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": results, "message": "Token refresh completed"})
}

// GetStatusWithTokens handles GET /api/status
func (h *TokenHandler) GetStatusWithTokens(c *gin.Context) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000Z")
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		tenantID = c.GetHeader("x-tenant-id")
	}

	if tenantID == "" {
		c.JSON(http.StatusOK, gin.H{
			"connection_status": "connected",
			"data":              map[string]string{"message": "Backend is running. Provide x-tenant-id header for token status."},
			"success":           true,
			"timestamp":         timestamp,
		})
		return
	}

	data := h.buildStatusWithTokensData(c, tenantID)
	c.JSON(http.StatusOK, gin.H{
		"connection_status": "connected", "data": data, "success": true, "timestamp": timestamp,
	})
}

// Helper methods

func (h *TokenHandler) buildAllPlatformStatuses(c *gin.Context, tenantID string) gin.H {
	platforms := GetAllPlatforms()
	statuses := make(gin.H, len(platforms))

	for _, platform := range platforms {
		status, _ := h.tokenManager.GetTokenStatus(c.Request.Context(), tenantID, platform)
		statuses[platform] = h.buildPlatformStatusData(platform, status)
	}
	return statuses
}

func (h *TokenHandler) buildPlatformStatusData(platform string, status *services.TokenInfo) gin.H {
	if status == nil {
		return gin.H{"platform": platform, "isExpired": true}
	}
	data := gin.H{"platform": platform, "isExpired": !status.IsValid || status.NeedsRefresh}
	if !status.ExpiresAt.IsZero() {
		data["expiresAt"] = FormatISOTimestamp(status.ExpiresAt)
	}
	if !status.RefreshTokenExpires.IsZero() {
		data["refreshTokenExpiresAt"] = FormatISOTimestamp(status.RefreshTokenExpires)
	}
	return data
}

func (h *TokenHandler) refreshPlatformToken(c *gin.Context, tenantID, platform string) (*services.TokenInfo, error) {
	switch platform {
	case models.PlatformShopee:
		return h.tokenManager.RefreshShopeeToken(c.Request.Context(), tenantID)
	case models.PlatformLazada:
		return h.tokenManager.RefreshLazadaToken(c.Request.Context(), tenantID)
	case models.PlatformTiktok:
		return h.tokenManager.RefreshTiktokToken(c.Request.Context(), tenantID)
	default:
		return nil, nil
	}
}

func (h *TokenHandler) refreshAllPlatforms(c *gin.Context, tenantID string, force bool) map[string]interface{} {
	results := make(map[string]interface{})
	for _, platform := range GetAllPlatforms() {
		if !force {
			status, _ := h.tokenManager.GetTokenStatus(c.Request.Context(), tenantID, platform)
			if status != nil && status.IsValid && !status.NeedsRefresh {
				results[platform] = gin.H{"success": true, "skipped": true, "reason": "Token still valid"}
				continue
			}
		}
		newToken, err := h.refreshPlatformToken(c, tenantID, platform)
		if err != nil {
			results[platform] = gin.H{"success": false, "error": err.Error()}
		} else if newToken != nil {
			results[platform] = gin.H{"success": true, "isValid": newToken.IsValid, "expiresAt": newToken.ExpiresAt}
		} else {
			results[platform] = gin.H{"success": false, "configured": false, "error": "Platform not configured"}
		}
	}
	return results
}

func (h *TokenHandler) buildStatusWithTokensData(c *gin.Context, tenantID string) map[string]interface{} {
	data := make(map[string]interface{})
	for _, platform := range GetAllPlatforms() {
		status, err := h.tokenManager.GetTokenStatus(c.Request.Context(), tenantID, platform)
		if err != nil || status == nil {
			data[platform] = gin.H{"platform": platform, "connected": false, "isExpired": true}
			continue
		}
		statusObj := gin.H{"platform": platform, "connected": status.IsValid, "isExpired": !status.IsValid || status.NeedsRefresh}
		if !status.ExpiresAt.IsZero() {
			statusObj["expiresAt"] = FormatISOTimestamp(status.ExpiresAt)
		}
		if !status.RefreshTokenExpires.IsZero() {
			statusObj["refreshTokenExpiresAt"] = FormatISOTimestamp(status.RefreshTokenExpires)
		}
		data[platform] = statusObj
	}
	return data
}
