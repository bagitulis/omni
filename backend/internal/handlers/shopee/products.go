package shopee

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// ProductHandler handles Shopee product HTTP requests
type ProductHandler struct {
	basePath string
}

// NewProductHandler creates a new product handler
func NewProductHandler(basePath string) *ProductHandler {
	return &ProductHandler{basePath: basePath}
}

// GetProducts handles GET /api/shopee/products
// Fetches products from Shopee API and syncs to database, then returns the products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	itemStatus := c.DefaultQuery("itemStatus", "NORMAL")
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if limit < 1 || limit > 100 {
		limit = 50
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
		log.Error().
			Str("handler", "shopee_products").
			Str("tenant_id", tenantID).
			Err(err).
			Msg("Failed to get Shopee credentials")
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

	// Create API client with global + tenant credentials
	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	// Use SyncService to fetch and save products with tenant ID
	syncService := shopeeService.NewSyncServiceWithTenant(client, db, tenantID)
	products, savedCount, err := syncService.SyncProductsWithDetails(c.Request.Context(), itemStatus, offset, limit)
	if err != nil {
		log.Error().
			Str("handler", "shopee_products").
			Str("tenant_id", tenantID).
			Err(err).
			Msg("Sync products error")
		c.JSON(http.StatusInternalServerError, response.Error("Failed to sync products: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"products":     products,
		"total":        len(products),
		"detail_saved": savedCount,
	}))
}

// GetProductByID handles GET /api/shopee/products/:itemId
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	itemIDStr := c.Param("itemId")

	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid item ID"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	service := shopeeService.NewProductService(db)
	product, err := service.GetProductByItemID(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Product not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(product))
}
