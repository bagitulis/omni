package lazada

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/cache"
	lazadaService "github.com/omni/backend/internal/services/lazada"
	"github.com/rs/zerolog/log"
)

// SyncHandler handles Lazada sync operations
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

// invalidateAnalyticsCache invalidates analytics cache after sync
func (h *SyncHandler) invalidateAnalyticsCache(tenantID, operation string) {
	if h.cacheService == nil {
		return
	}

	// Invalidate relevant cache keys
	keys := []string{
		"analytics:unified:summary",
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

// SyncOrders handles POST /api/lazada/sync/orders
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

	// Use helper to get client with proper tenant credentials
	client, err := GetLazadaClient(tenantID, h.basePath)
	if err != nil {
		if err == ErrMissingAccessToken {
			c.JSON(http.StatusBadRequest, response.Error("Lazada access token not configured"))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client: "+err.Error()))
		}
		return
	}

	syncService := lazadaService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncOrders(c.Request.Context(), "")
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

// SyncProducts handles POST /api/lazada/sync/products
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

	// Use helper to get client with proper tenant credentials
	client, err := GetLazadaClient(tenantID, h.basePath)
	if err != nil {
		if err == ErrMissingAccessToken {
			c.JSON(http.StatusBadRequest, response.Error("Lazada access token not configured"))
		} else {
			c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client: "+err.Error()))
		}
		return
	}

	syncService := lazadaService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncProducts(c.Request.Context())
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
