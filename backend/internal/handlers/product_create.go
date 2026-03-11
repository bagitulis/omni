package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/products"
	"gorm.io/gorm"
)

// ProductCreateHandler handles product creation endpoints
type ProductCreateHandler struct {
	db             *gorm.DB
	categoryMapper *products.CategoryMapper
	imageUploader  *products.ImageUploader
}

// NewProductCreateHandler creates a new product create handler
func NewProductCreateHandler(db *gorm.DB) *ProductCreateHandler {
	return &ProductCreateHandler{
		db:             db,
		categoryMapper: products.NewCategoryMapper(),
		imageUploader:  products.NewImageUploader(),
	}
}

// CreateOnShopee handles POST /api/products/create/shopee
func (h *ProductCreateHandler) CreateOnShopee(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenant_id"})
		return
	}

	var req products.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get API client from context (injected by middleware)
	api, exists := c.Get("shopeeAPI")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Shopee API not configured"})
		return
	}

	service := products.NewCreateService(h.db, tenantID)
	result, err := service.CreateOnShopee(c.Request.Context(), api.(products.PlatformAPI), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "item_id": result.ItemID})
}

// CreateOnLazada handles POST /api/products/create/lazada
func (h *ProductCreateHandler) CreateOnLazada(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenant_id"})
		return
	}

	var req products.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	api, exists := c.Get("lazadaAPI")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lazada API not configured"})
		return
	}

	service := products.NewCreateService(h.db, tenantID)
	result, err := service.CreateOnLazada(c.Request.Context(), api.(products.PlatformAPI), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "item_id": result.ItemID})
}

// CreateOnTiktok handles POST /api/products/create/tiktok
func (h *ProductCreateHandler) CreateOnTiktok(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenant_id"})
		return
	}

	var req products.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	api, exists := c.Get("tiktokAPI")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "TikTok API not configured"})
		return
	}

	service := products.NewCreateService(h.db, tenantID)
	result, err := service.CreateOnTiktok(c.Request.Context(), api.(products.PlatformAPI), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !result.Success {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": result.Error})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "product_id": result.Message})
}

// GetCategories handles GET /api/products/categories/:platform
func (h *ProductCreateHandler) GetCategories(c *gin.Context) {
	platform := c.Param("platform")
	if platform == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform required"})
		return
	}

	// Get parent_id from query
	parentID := int64(0)
	if pid := c.Query("parent_id"); pid != "" {
		// Parse parent_id if provided
	}

	// Get appropriate API client
	apiKey := platform + "API"
	api, exists := c.Get(apiKey)
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": platform + " API not configured"})
		return
	}

	cats, err := h.categoryMapper.GetCategories(c.Request.Context(), api.(products.PlatformAPI), platform, parentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "categories": cats})
}
