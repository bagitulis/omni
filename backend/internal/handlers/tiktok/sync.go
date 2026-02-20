package tiktok

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/cache"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// SyncHandler handles TikTok sync operations
type SyncHandler struct {
	basePath     string
	cacheService cache.CacheManager
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(basePath string) *SyncHandler {
	return &SyncHandler{basePath: basePath}
}

// NewSyncHandlerWithCache creates a new sync handler with cache service
func NewSyncHandlerWithCache(basePath string, cacheService cache.CacheManager) *SyncHandler {
	return &SyncHandler{basePath: basePath, cacheService: cacheService}
}

// getTiktokClient creates TikTok API client with tenant-specific credentials
func (h *SyncHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	ctx := context.Background()
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	// Use PlatformCredentialsRepository for key-value based config (current schema)
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if tenantCreds.AccessToken == "" || tenantCreds.ShopCipher == "" {
		return nil, fmt.Errorf("missing TikTok credentials: accessToken or shopCipher not configured")
	}

	// Use tenant credentials for appKey/appSecret if available, otherwise fall back to global
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret

	if appKey == "" || appSecret == "" {
		systemDB, err := config.GetSystemDB(h.basePath)
		if err != nil {
			return nil, err
		}
		globalRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, err := globalRepo.GetTiktokCredentials(ctx)
		if err != nil {
			return nil, err
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
	}

	client := tiktokPkg.NewClient(appKey, appSecret)
	client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)
	return client, nil
}

// invalidateAnalyticsCache invalidates analytics cache after sync
func (h *SyncHandler) invalidateAnalyticsCache(tenantID, operation string) {
	if h.cacheService == nil {
		return
	}

	// Invalidate relevant cache keys
	keys := []string{
		"analytics:unified:summary",
		"analytics:ml:portfolio-health:tiktok",
	}

	for _, key := range keys {
		if err := h.cacheService.Delete(tenantID, key); err != nil {
			log.Warn().
				Err(err).
				Str("tenant_id", tenantID).
				Str("cache_key", key).
				Msg("Failed to invalidate cache")
		}
	}

	log.Debug().
		Str("tenant_id", tenantID).
		Str("operation", operation).
		Msg("Analytics cache invalidated after sync")
}

// SyncOrders handles POST /api/tiktok/sync/orders
func (h *SyncHandler) SyncOrders(c *gin.Context) {
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

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Failed to get TikTok client")
		c.JSON(http.StatusBadRequest, response.Error("TikTok credentials not configured: "+err.Error()))
		return
	}

	syncService := tiktokService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncOrders(context.Background(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Sync failed: "+err.Error()))
		return
	}

	// Invalidate analytics cache after successful sync
	h.invalidateAnalyticsCache(tenantID, "sync_orders")

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Orders synced",
		"synced":  count,
	}))
}

// SyncProducts handles POST /api/tiktok/sync/products
func (h *SyncHandler) SyncProducts(c *gin.Context) {
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

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Failed to get TikTok client")
		c.JSON(http.StatusBadRequest, response.Error("TikTok credentials not configured: "+err.Error()))
		return
	}

	syncService := tiktokService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncProducts(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Sync failed: "+err.Error()))
		return
	}

	// Invalidate analytics cache after successful sync
	h.invalidateAnalyticsCache(tenantID, "sync_products")

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Products synced",
		"synced":  count,
	}))
}
