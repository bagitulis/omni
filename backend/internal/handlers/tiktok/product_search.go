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
	totalProducts, err := svc.SyncProducts(c.Request.Context())
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
	ctx := context.Background()
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	// Get platform credentials from tenant's key-value storage
	// NOTE: PostgreSQL uses schema isolation, NOT tenant_id column
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
		// Fall back to global credentials
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
