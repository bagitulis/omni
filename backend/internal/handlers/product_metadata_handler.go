package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
)

// ProductMetadataHandler handles product creation metadata endpoints
type ProductMetadataHandler struct {
	basePath string
}

// NewProductMetadataHandler creates a new product metadata handler
func NewProductMetadataHandler(basePath string) *ProductMetadataHandler {
	return &ProductMetadataHandler{basePath: basePath}
}

// GetShopeeCategories handles GET /api/products/create/shopee/categories
func (h *ProductMetadataHandler) GetShopeeCategories(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	categories := []CategoryItem{
		{ID: "100001", Name: "Women Clothes", Level: 1, IsLeaf: false},
		{ID: "100002", Name: "Men Clothes", Level: 1, IsLeaf: false},
		{ID: "100003", Name: "Mobile & Gadgets", Level: 1, IsLeaf: false},
		{ID: "100004", Name: "Home & Living", Level: 1, IsLeaf: false},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categories": categories}))
}

// GetShopeeAttributes handles GET /api/products/create/shopee/attributes/:catId
func (h *ProductMetadataHandler) GetShopeeAttributes(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	catID := c.Param("catId")
	attributes := []AttributeItem{
		{ID: "brand", Name: "Brand", Type: "string", Required: true, InputType: "dropdown"},
		{ID: "color", Name: "Color", Type: "string", Required: false, Options: []string{"Red", "Blue", "Green", "Black", "White"}, InputType: "dropdown"},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categoryId": catID, "attributes": attributes}))
}

// GetShopeeBrands handles GET /api/products/create/shopee/brands/:catId
func (h *ProductMetadataHandler) GetShopeeBrands(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	catID := c.Param("catId")
	brands := []BrandItem{{ID: "0", Name: "No Brand"}, {ID: "1", Name: "Brand A"}, {ID: "2", Name: "Brand B"}}

	c.JSON(http.StatusOK, response.Success(gin.H{"categoryId": catID, "brands": brands}))
}

// GetShopeeLogistics handles GET /api/products/create/shopee/logistics
func (h *ProductMetadataHandler) GetShopeeLogistics(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	logistics := []LogisticsItem{
		{ID: "1", Name: "Standard Delivery", Enabled: true, Fee: 0, MaxWeight: 30},
		{ID: "2", Name: "Express Delivery", Enabled: true, Fee: 5000, MaxWeight: 10},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"logistics": logistics}))
}

// GetLazadaCategories handles GET /api/products/create/lazada/categories
func (h *ProductMetadataHandler) GetLazadaCategories(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	categories := []CategoryItem{
		{ID: "3", Name: "Cameras", Level: 1, IsLeaf: false},
		{ID: "4", Name: "Mobiles & Tablets", Level: 1, IsLeaf: false},
		{ID: "5", Name: "Consumer Electronics", Level: 1, IsLeaf: false},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categories": categories}))
}

// GetLazadaAttributes handles GET /api/products/create/lazada/attributes/:catId
func (h *ProductMetadataHandler) GetLazadaAttributes(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	catID := c.Param("catId")
	attributes := []AttributeItem{
		{ID: "brand", Name: "Brand", Type: "string", Required: true, InputType: "dropdown"},
		{ID: "warranty_type", Name: "Warranty Type", Type: "string", Required: true, InputType: "dropdown"},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categoryId": catID, "attributes": attributes}))
}

// GetLazadaBrands handles GET /api/products/create/lazada/brands/:catId
func (h *ProductMetadataHandler) GetLazadaBrands(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	catID := c.Param("catId")
	brands := []BrandItem{{ID: "0", Name: "No Brand"}, {ID: "1", Name: "Samsung"}, {ID: "2", Name: "Apple"}}

	c.JSON(http.StatusOK, response.Success(gin.H{"categoryId": catID, "brands": brands}))
}

// GetTiktokCategories handles GET /api/products/create/tiktok/categories
func (h *ProductMetadataHandler) GetTiktokCategories(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	categories := []CategoryItem{
		{ID: "1", Name: "Womenswear & Underwear", Level: 1, IsLeaf: false},
		{ID: "2", Name: "Menswear & Underwear", Level: 1, IsLeaf: false},
		{ID: "3", Name: "Phones & Electronics", Level: 1, IsLeaf: false},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categories": categories}))
}

// GetTiktokAttributes handles GET /api/products/create/tiktok/attributes/:catId
func (h *ProductMetadataHandler) GetTiktokAttributes(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	catID := c.Param("catId")
	attributes := []AttributeItem{
		{ID: "brand", Name: "Brand", Type: "string", Required: true, InputType: "dropdown"},
		{ID: "material", Name: "Material", Type: "string", Required: false, InputType: "text"},
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"categoryId": catID, "attributes": attributes}))
}

// GetTiktokBrands handles GET /api/products/create/tiktok/brands
func (h *ProductMetadataHandler) GetTiktokBrands(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	brands := []BrandItem{{ID: "0", Name: "No Brand"}, {ID: "1", Name: "Brand X"}, {ID: "2", Name: "Brand Y"}}
	c.JSON(http.StatusOK, response.Success(gin.H{"brands": brands}))
}

// GetTiktokWarehouses handles GET /api/products/create/tiktok/warehouses
func (h *ProductMetadataHandler) GetTiktokWarehouses(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	warehouses := []WarehouseItem{{ID: "wh1", Name: "Main Warehouse", Address: "Jakarta", IsDefault: true}}
	c.JSON(http.StatusOK, response.Success(gin.H{"warehouses": warehouses}))
}

// UploadImage handles POST /api/products/create/upload-image
func (h *ProductMetadataHandler) UploadImage(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	var req UploadImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"platform": req.Platform,
		"imageUrl": "https://example.com/uploaded/image.jpg",
		"imageId":  "img_12345",
		"success":  true,
	}))
}

// ValidateProduct handles POST /api/products/create/validate
func (h *ProductMetadataHandler) ValidateProduct(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	var req ValidateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	result := ValidationResult{Valid: true, Errors: []ValidationError{}, Warnings: []string{}}

	if req.Product["title"] == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{Field: "title", Message: "Title is required"})
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetTemplates handles GET /api/products/create/templates
func (h *ProductMetadataHandler) GetTemplates(c *gin.Context) {
	if !h.validateTenant(c) {
		return
	}

	platform := c.Query("platform")
	templates := []ProductTemplate{
		{ID: "1", Name: "Basic Product", Platform: "all", Description: "Simple product template"},
		{ID: "2", Name: "Fashion Item", Platform: "shopee", CategoryID: "100001", Description: "Template for fashion products"},
	}

	if platform != "" {
		filtered := make([]ProductTemplate, 0)
		for _, t := range templates {
			if t.Platform == platform || t.Platform == "all" {
				filtered = append(filtered, t)
			}
		}
		templates = filtered
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"templates": templates}))
}

// Helper: validateTenant checks tenant ID and returns false if invalid
func (h *ProductMetadataHandler) validateTenant(c *gin.Context) bool {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return false
	}
	return true
}
