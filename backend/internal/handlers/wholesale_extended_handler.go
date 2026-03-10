package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/wholesale"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WholesaleExtendedHandler handles extended wholesale endpoints
type WholesaleExtendedHandler struct {
	basePath   string
	fallbackDB *gorm.DB
}

// NewWholesaleExtendedHandler creates a new wholesale extended handler
func NewWholesaleExtendedHandler(basePath string, db *gorm.DB) *WholesaleExtendedHandler {
	return &WholesaleExtendedHandler{
		basePath:   basePath,
		fallbackDB: db,
	}
}

// getDB returns the appropriate database for the current request
func (h *WholesaleExtendedHandler) getDB(_ *gin.Context, tenantID string) (*gorm.DB, error) {
	if h.fallbackDB != nil {
		return h.fallbackDB, nil
	}
	return config.GetTenantDB(tenantID, h.basePath)
}

// DeleteWholesale handles DELETE /api/wholesale/shopee/:itemId
func (h *WholesaleExtendedHandler) DeleteWholesale(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid itemId"))
		return
	}

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed: "+err.Error()))
		return
	}

	// Get Shopee API client
	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed: "+err.Error()))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	service := wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI)
	err = service.DeleteWholesaleTiers(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"item_id": itemID,
		"deleted": true,
		"message": "Wholesale tiers deleted successfully",
	}))
}

// UpdateWholesale handles PUT /api/wholesale/shopee/:itemId
// REAL implementation: calls Shopee API to update wholesale tiers
func (h *WholesaleExtendedHandler) UpdateWholesale(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid itemId"))
		return
	}

	var req UpdateWholesaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed: "+err.Error()))
		return
	}

	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed: "+err.Error()))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	service := wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI)

	// Convert DTO tiers to service tiers
	tiers := make([]wholesale.WholesaleTier, len(req.Tiers))
	for i, t := range req.Tiers {
		tiers[i] = wholesale.WholesaleTier{
			MinCount:  t.MinCount,
			MaxCount:  t.MaxCount,
			UnitPrice: t.UnitPrice,
		}
	}

	err = service.UpdateWholesaleTiers(c.Request.Context(), itemID, tiers)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"item_id": itemID,
		"tiers":   req.Tiers,
		"updated": true,
		"message": "Wholesale tiers updated successfully",
	}))
}

// GetWholesaleInfo handles GET /api/wholesale/shopee/:itemId
// REAL implementation: gets wholesale info from Shopee API
func (h *WholesaleExtendedHandler) GetWholesaleInfo(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid itemId"))
		return
	}

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed: "+err.Error()))
		return
	}

	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed: "+err.Error()))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	service := wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI)

	tiers, err := service.GetWholesaleTiers(c.Request.Context(), itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get wholesale tiers: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"item_id":       itemID,
		"has_wholesale": len(tiers) > 0,
		"tiers":         tiers,
	}))
}

// LookupItemId handles GET /api/wholesale/shopee/lookup/:sku
// Response uses snake_case to match frontend expectations
func (h *WholesaleExtendedHandler) LookupItemId(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing SKU"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed: "+err.Error()))
		return
	}

	repo := repositories.NewShopeeProductRepository(db)
	product, err := repo.FindBySKU(c.Request.Context(), sku)
	if err != nil || product == nil {
		c.JSON(http.StatusNotFound, response.Error("SKU not found"))
		return
	}

	// Use snake_case keys matching frontend SkuLookupResult type
	c.JSON(http.StatusOK, response.Success(gin.H{
		"sku":     sku,
		"item_id": product.ItemID,
		"name":    product.Name,
	}))
}

// SetTiktokWholesale handles POST /api/wholesale/tiktok/:productId
// TODO: Implement real TikTok API call (requires TikTok API client)
func (h *WholesaleExtendedHandler) SetTiktokWholesale(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing productId"))
		return
	}

	var req TiktokWholesaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	log.Warn().
		Str("tenant_id", tenantID).
		Str("product_id", productID).
		Msg("TikTok wholesale not implemented - requires TikTok API client integration")

	c.JSON(http.StatusNotImplemented, response.Error("TikTok wholesale not yet implemented"))
}
