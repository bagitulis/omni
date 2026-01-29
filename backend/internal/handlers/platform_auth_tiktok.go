package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// TiktokShop represents a TikTok shop
type TiktokShop struct {
	ShopID     string `json:"shop_id"`
	ShopName   string `json:"shop_name"`
	Region     string `json:"region"`
	IsActive   bool   `json:"is_active"`
	SellerType string `json:"seller_type,omitempty"`
}

// GetTiktokShops handles GET /api/platform-auth/tiktok/shops
func (h *PlatformAuthHandler) GetTiktokShops(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	shops, err := h.fetchTiktokShops(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"shops": shops}))
}

// GetActiveTiktokShop handles GET /api/platform-auth/tiktok/active-shop
func (h *PlatformAuthHandler) GetActiveTiktokShop(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewPlatformConfigRepository(db)
	cfg, err := repo.FindByTenantAndPlatform(c.Request.Context(), tenantID, models.PlatformTiktok)
	if err != nil || cfg == nil {
		c.JSON(http.StatusNotFound, response.Error("TikTok not connected"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"shop": TiktokShop{
			ShopID:   cfg.ShopID,
			ShopName: cfg.ShopName,
			Region:   cfg.Region,
			IsActive: cfg.IsActive,
		},
	}))
}

// fetchTiktokShops fetches TikTok shops from API
func (h *PlatformAuthHandler) fetchTiktokShops(ctx context.Context, tenantID string) ([]TiktokShop, error) {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, fmt.Errorf("database connection failed")
	}

	repo := repositories.NewPlatformConfigRepository(db)
	cfg, err := repo.FindByTenantAndPlatform(ctx, tenantID, models.PlatformTiktok)
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("TikTok not connected")
	}

	// In production, this would call TikTok API to get authorized shops
	// For now, return the connected shop
	shops := []TiktokShop{
		{
			ShopID:   cfg.ShopID,
			ShopName: cfg.ShopName,
			Region:   cfg.Region,
			IsActive: cfg.IsActive,
		},
	}

	return shops, nil
}

// GetStatus handles GET /api/platform-auth/status
func (h *PlatformAuthHandler) GetStatus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Use TenantPlatformConfigRepository which uses key-value pattern
	repo := repositories.NewTenantPlatformConfigRepository(db)

	now := time.Now()
	statuses := make(map[string]ConnectionStatus)
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}

	// Get config for each platform
	for _, platform := range platforms {
		configMap, err := repo.GetAllConfigByPlatform(c.Request.Context(), platform)
		if err != nil {
			statuses[platform] = ConnectionStatus{
				Platform:    platform,
				Connected:   false,
				LastChecked: now,
			}
			continue
		}

		// Check if connected (has access token)
		accessToken := configMap["accessToken"]
		connected := accessToken != ""

		// Parse expiry timestamp
		var expiresAt int64
		if expStr, ok := configMap["tokenExpiry"]; ok && expStr != "" {
			fmt.Sscanf(expStr, "%d", &expiresAt)
		}

		// Normalize to milliseconds - if value is too small, it's in seconds
		// Timestamps after year 2001 in ms are > 1_000_000_000_000
		expiresAtMs := expiresAt
		if expiresAt > 0 && expiresAt < 1_000_000_000_000 {
			expiresAtMs = expiresAt * 1000 // Convert seconds to milliseconds
		}

		expiresSoon := expiresAtMs > 0 && expiresAtMs < now.Add(24*time.Hour).UnixMilli()
		expired := expiresAtMs > 0 && expiresAtMs < now.UnixMilli()

		statuses[platform] = ConnectionStatus{
			Platform:    platform,
			Connected:   connected,
			ShopID:      configMap["shopId"],
			ShopName:    configMap["shopName"],
			ExpiresAt:   expiresAtMs,
			ExpiresSoon: expiresSoon,
			Expired:     expired,
			LastChecked: now,
		}
	}

	c.JSON(http.StatusOK, response.Success(statuses))
}

// GetLogs handles GET /api/platform-auth/logs
func (h *PlatformAuthHandler) GetLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Get OAuth logs from database
	var logs []models.OAuthLog
	result := db.Order("created_at desc").Limit(100).Find(&logs)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get logs"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"logs":  logs,
		"total": len(logs),
	}))
}

// getBackendURL gets the backend URL from request
func (h *PlatformAuthHandler) getBackendURL(c *gin.Context) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}
