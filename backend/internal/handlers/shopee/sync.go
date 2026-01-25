package shopee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// SyncHandler handles sync operations
type SyncHandler struct {
	basePath string
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(basePath string) *SyncHandler {
	return &SyncHandler{basePath: basePath}
}

// SyncOrders handles POST /api/shopee/sync/orders
func (h *SyncHandler) SyncOrders(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
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

	// Get system database for GlobalConfig
	systemDB, err := config.GetSystemDB(h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("System database failed"))
		return
	}

	// Get credentials from GlobalConfig
	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	creds, err := configRepo.GetShopeeCredentials(c.Request.Context())
	if err != nil || creds.PartnerID == 0 {
		c.JSON(http.StatusBadRequest, response.Error("Shopee credentials not configured"))
		return
	}

	// Create API client with credentials
	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, true)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	// Sync orders
	syncService := shopeeService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncOrders(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Sync failed: "+err.Error()))
		return
	}

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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	systemDB, err := config.GetSystemDB(h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("System database failed"))
		return
	}

	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	creds, err := configRepo.GetShopeeCredentials(c.Request.Context())
	if err != nil || creds.PartnerID == 0 {
		c.JSON(http.StatusBadRequest, response.Error("Shopee credentials not configured"))
		return
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, true)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	syncService := shopeeService.NewSyncServiceWithTenant(client, db, tenantID)
	count, err := syncService.SyncProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Sync failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Products synced successfully",
		"synced":  count,
	}))
}
