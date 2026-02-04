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
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// ShippingHandler handles TikTok shipping document requests
type ShippingHandler struct {
	basePath string
}

// NewShippingHandler creates a new shipping handler
func NewShippingHandler(basePath string) *ShippingHandler {
	return &ShippingHandler{basePath: basePath}
}

// ShippingDocumentResponse represents the shipping document response
type ShippingDocumentResponse struct {
	DocURL         string `json:"doc_url"`
	TrackingNumber string `json:"tracking_number,omitempty"`
}

// getTiktokClient creates TikTok API client with tenant-specific credentials
func (h *ShippingHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	ctx := context.Background()
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	// Use PlatformCredentialsRepository for key-value based config
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

// GetShippingDocument handles GET /api/tiktok/shipping/document/:packageId
func (h *ShippingHandler) GetShippingDocument(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	packageID := c.Param("packageId")
	if packageID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing packageId"))
		return
	}

	documentType := c.DefaultQuery("document_type", "SHIPPING_LABEL")

	// Get TikTok client
	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok credentials: "+err.Error()))
		return
	}

	// Get shipping document from TikTok API
	docURL, err := client.GetShippingDocument(packageID, documentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get shipping document: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(ShippingDocumentResponse{
		DocURL: docURL,
	}))
}
