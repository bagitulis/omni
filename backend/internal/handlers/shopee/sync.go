package shopee

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/cache"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// SyncHandler handles sync operations
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
		"analytics:ml:portfolio-health:shopee",
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

// SyncOrders handles POST /api/shopee/sync/orders
func (h *SyncHandler) SyncOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days < 1 || days > 30 {
		days = 7
	}

	// Get tenant database
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "shopee")
	if err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Failed to get Shopee credentials")
		c.JSON(http.StatusBadRequest, response.Error("Shopee not configured for this tenant"))
		return
	}
	if creds.ShopID == 0 || creds.AccessToken == "" {
		c.JSON(http.StatusBadRequest, response.Error("Shopee shop not connected - missing shopId or accessToken"))
		return
	}
	if creds.PartnerID == 0 || creds.PartnerKey == "" {
		c.JSON(http.StatusBadRequest, response.Error("Shopee global credentials not configured"))
		return
	}

	// Create API client with combined credentials
	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	// Sync orders
	syncService := shopeeService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncOrders(context.Background(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	// Invalidate analytics cache after successful sync
	h.invalidateAnalyticsCache(tenantID, "sync_orders")

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Orders synced successfully",
		"synced":  count,
		"days":    days,
	}))
}

// SyncProducts handles POST /api/shopee/sync/products
func (h *SyncHandler) SyncProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "shopee")
	if err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Failed to get Shopee credentials")
		c.JSON(http.StatusBadRequest, response.Error("Shopee not configured for this tenant"))
		return
	}
	if creds.ShopID == 0 || creds.AccessToken == "" {
		c.JSON(http.StatusBadRequest, response.Error("Shopee shop not connected - missing shopId or accessToken"))
		return
	}
	if creds.PartnerID == 0 || creds.PartnerKey == "" {
		c.JSON(http.StatusBadRequest, response.Error("Shopee global credentials not configured"))
		return
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	syncService := shopeeService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncProducts(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	// Invalidate analytics cache after successful sync
	h.invalidateAnalyticsCache(tenantID, "sync_products")

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Products synced successfully",
		"synced":  count,
	}))
}
