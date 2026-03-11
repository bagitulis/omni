package tiktok

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	tiktokSvc "github.com/omni/backend/internal/services/tiktok"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// ProductSearchHandler handles TikTok product search/sync operations
type ProductSearchHandler struct {
	basePath string
}

// NewProductSearchHandler creates a new product search handler
func NewProductSearchHandler(basePath string) *ProductSearchHandler {
	return &ProductSearchHandler{basePath: basePath}
}

// SearchProducts handles POST /api/tiktok/products/search
// Delegates sync flow to SyncService
func (h *ProductSearchHandler) SearchProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	svc := tiktokSvc.NewSyncServiceWithTenant(client, db, tenantID)
	totalProducts, err := svc.SyncProducts(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"total_products": totalProducts,
	})
}

// getTiktokClient creates TikTok API client for tenant
func (h *ProductSearchHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	return NewTiktokClient(tenantID, h.basePath)
}
