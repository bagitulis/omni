package tiktok

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// SyncHandler handles TikTok sync operations
type SyncHandler struct {
	basePath string
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(basePath string) *SyncHandler {
	return &SyncHandler{basePath: basePath}
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

	systemDB, err := config.GetSystemDB(h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("System database failed"))
		return
	}

	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	creds, _ := configRepo.GetTiktokCredentials(c.Request.Context())
	if creds.AppKey == "" {
		c.JSON(http.StatusBadRequest, response.Error("TikTok credentials not configured"))
		return
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, "")

	syncService := tiktokService.NewSyncService(client, db)
	count, err := syncService.SyncOrders(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Sync failed: "+err.Error()))
		return
	}

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

	systemDB, err := config.GetSystemDB(h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("System database failed"))
		return
	}

	configRepo := repositories.NewGlobalConfigRepository(systemDB)
	creds, _ := configRepo.GetTiktokCredentials(c.Request.Context())
	if creds.AppKey == "" {
		c.JSON(http.StatusBadRequest, response.Error("TikTok credentials not configured"))
		return
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, "")

	syncService := tiktokService.NewSyncService(client, db)
	count, err := syncService.SyncProducts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Sync failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message": "Products synced",
		"synced":  count,
	}))
}
