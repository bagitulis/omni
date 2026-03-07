package lazada

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	lazadaService "github.com/omni/backend/internal/services/lazada"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog/log"
)

// ProductHandler handles Lazada product HTTP requests
type ProductHandler struct {
	basePath string
}

// NewProductHandler creates a new product handler
func NewProductHandler(basePath string) *ProductHandler {
	return &ProductHandler{basePath: basePath}
}

// GetProducts handles GET /api/lazada/products
// Fetches products from Lazada API and syncs to database, then returns the products
func (h *ProductHandler) GetProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

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

	// Get platform credentials from tenant's key-value storage
	// NOTE: PostgreSQL uses schema isolation, NOT tenant_id column
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	tenantCreds, err := credRepo.GetLazadaCredentials(c.Request.Context())
	if err != nil {
		log.Error().
			Str("handler", "lazada_products").
			Str("tenant_id", tenantID).
			Err(err).
			Msg("Failed to get Lazada credentials")
		c.JSON(http.StatusBadRequest, response.Error("Lazada not configured for this tenant"))
		return
	}
	if tenantCreds.AccessToken == "" {
		c.JSON(http.StatusBadRequest, response.Error("Lazada shop not connected - missing accessToken"))
		return
	}

	// Use tenant credentials for appKey/appSecret if available, otherwise fall back to global
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret
	region := tenantCreds.Region
	if region == "" {
		region = "ID" // Default to Indonesia
	}

	if appKey == "" || appSecret == "" {
		// Fall back to global credentials
		systemDB, err := config.GetSystemDB(h.basePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.Error("System database failed"))
			return
		}
		globalRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, err := globalRepo.GetLazadaCredentials(c.Request.Context())
		if err != nil || globalCreds.AppKey == "" {
			c.JSON(http.StatusBadRequest, response.Error("Lazada credentials not configured"))
			return
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
		if region == "" {
			region = globalCreds.Region
		}
	}

	// Create API client
	client := lazadaPkg.NewClient(appKey, appSecret, region)
	client.SetAccessToken(tenantCreds.AccessToken)

	// Use SyncService to fetch and save products with tenantID
	syncService := lazadaService.NewSyncServiceWithTenant(client, db, tenantID)
	products, savedCount, err := syncService.SyncProductsWithDetails(c.Request.Context(), offset, limit)
	if err != nil {
		log.Error().
			Str("handler", "lazada_products").
			Str("tenant_id", tenantID).
			Err(err).
			Msg("Sync products error")
		c.JSON(http.StatusInternalServerError, response.Error("Failed to sync products: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"products":     products,
		"total":        len(products),
		"detail_saved": savedCount,
	})
}

// GetProductByID handles GET /api/lazada/products/:itemId
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemID := c.Param("itemId")

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewLazadaProductRepository(db)
	product, err := repo.FindByItemID(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error("Product not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(product))
}

// CreateProduct handles POST /api/lazada/products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client"))
		return
	}

	createReq := lazadaPkg.CreateProductRequest{
		Name:            req.Name,
		Description:     req.Description,
		Brand:           req.Brand,
		PrimaryCategory: req.PrimaryCategory,
		Skus: []lazadaPkg.CreateProductSku{{
			SellerSku: req.SellerSku,
			Price:     req.Price,
			Quantity:  req.Quantity,
		}},
	}

	resp, err := client.CreateProduct(createReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to create product: "+err.Error()))
		return
	}

	c.JSON(http.StatusCreated, response.Success(gin.H{
		"item_id": resp.Data.ItemID,
		"skus":    resp.Data.SkuList,
	}))
}

// UpdateProduct handles PUT /api/lazada/products/:itemId
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemID := c.Param("itemId")

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client"))
		return
	}

	updateReq := lazadaPkg.UpdateProductRequest{
		ItemID:      itemID,
		Name:        req.Name,
		Description: req.Description,
	}

	resp, err := client.UpdateProduct(updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to update product: "+err.Error()))
		return
	}

	if resp.Code != "0" {
		c.JSON(http.StatusBadRequest, response.Error("Lazada update failed (code: "+resp.Code+")"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"updated": true,
	}))
}

// DeleteProduct handles DELETE /api/lazada/products/:itemId
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemID := c.Param("itemId")

	client, err := h.getLazadaClient(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get Lazada client"))
		return
	}

	resp, err := client.DeleteProduct(itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to delete product: "+err.Error()))
		return
	}

	if resp.Code != "0" {
		c.JSON(http.StatusBadRequest, response.Error("Lazada delete failed (code: "+resp.Code+")"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"deleted": true,
	}))
}

// getLazadaClient creates Lazada API client for tenant using shared helper
func (h *ProductHandler) getLazadaClient(tenantID string) (*lazadaPkg.Client, error) {
	return GetLazadaClient(tenantID, h.basePath)
}
