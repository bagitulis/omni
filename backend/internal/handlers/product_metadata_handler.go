package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
)

// ProductMetadataHandler handles product creation metadata endpoints.
// All endpoints return 501 until real platform SDK integration is implemented.
// Real implementations exist in platform-specific handlers:
//   - Lazada categories: lazada/product_extended_handler.go:GetCategories
//   - TikTok warehouses: tiktok/product_create_handler.go:GetWarehouses
//   - TikTok attributes: tiktok/product_create_handler.go (uses pkg/tiktok/category.go)
type ProductMetadataHandler struct {
	basePath string
}

// NewProductMetadataHandler creates a new product metadata handler
func NewProductMetadataHandler(basePath string) *ProductMetadataHandler {
	return &ProductMetadataHandler{basePath: basePath}
}

// notImplemented returns a standard 501 response
func (h *ProductMetadataHandler) notImplemented(c *gin.Context, feature string) {
	c.JSON(http.StatusNotImplemented, response.Error(feature+" not yet implemented — use platform-specific endpoints"))
}

// GetShopeeCategories handles GET /api/products/create/shopee/categories
func (h *ProductMetadataHandler) GetShopeeCategories(c *gin.Context) {
	h.notImplemented(c, "Shopee categories")
}

// GetShopeeAttributes handles GET /api/products/create/shopee/attributes/:catId
func (h *ProductMetadataHandler) GetShopeeAttributes(c *gin.Context) {
	h.notImplemented(c, "Shopee attributes")
}

// GetShopeeBrands handles GET /api/products/create/shopee/brands/:catId
func (h *ProductMetadataHandler) GetShopeeBrands(c *gin.Context) {
	h.notImplemented(c, "Shopee brands")
}

// GetShopeeLogistics handles GET /api/products/create/shopee/logistics
func (h *ProductMetadataHandler) GetShopeeLogistics(c *gin.Context) {
	h.notImplemented(c, "Shopee logistics")
}

// GetLazadaCategories handles GET /api/products/create/lazada/categories
func (h *ProductMetadataHandler) GetLazadaCategories(c *gin.Context) {
	h.notImplemented(c, "Lazada categories (use /api/lazada/categories)")
}

// GetLazadaAttributes handles GET /api/products/create/lazada/attributes/:catId
func (h *ProductMetadataHandler) GetLazadaAttributes(c *gin.Context) {
	h.notImplemented(c, "Lazada attributes (use /api/lazada/categories/:id/attributes)")
}

// GetLazadaBrands handles GET /api/products/create/lazada/brands/:catId
func (h *ProductMetadataHandler) GetLazadaBrands(c *gin.Context) {
	h.notImplemented(c, "Lazada brands")
}

// GetTiktokCategories handles GET /api/products/create/tiktok/categories
func (h *ProductMetadataHandler) GetTiktokCategories(c *gin.Context) {
	h.notImplemented(c, "TikTok categories")
}

// GetTiktokAttributes handles GET /api/products/create/tiktok/attributes/:catId
func (h *ProductMetadataHandler) GetTiktokAttributes(c *gin.Context) {
	h.notImplemented(c, "TikTok attributes (use /api/tiktok/products/categories/:id/attributes)")
}

// GetTiktokBrands handles GET /api/products/create/tiktok/brands
func (h *ProductMetadataHandler) GetTiktokBrands(c *gin.Context) {
	h.notImplemented(c, "TikTok brands")
}

// GetTiktokWarehouses handles GET /api/products/create/tiktok/warehouses
func (h *ProductMetadataHandler) GetTiktokWarehouses(c *gin.Context) {
	h.notImplemented(c, "TikTok warehouses (use /api/tiktok/products/warehouses)")
}

// UploadImage handles POST /api/products/create/upload-image
func (h *ProductMetadataHandler) UploadImage(c *gin.Context) {
	h.notImplemented(c, "Unified image upload (use platform-specific upload endpoints)")
}

// ValidateProduct handles POST /api/products/create/validate
func (h *ProductMetadataHandler) ValidateProduct(c *gin.Context) {
	h.notImplemented(c, "Product validation")
}

// GetTemplates handles GET /api/products/create/templates
func (h *ProductMetadataHandler) GetTemplates(c *gin.Context) {
	h.notImplemented(c, "Product templates")
}

// Helper: validateTenant checks tenant ID and returns false if invalid
func (h *ProductMetadataHandler) validateTenant(c *gin.Context) bool {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return false
	}
	return true
}
