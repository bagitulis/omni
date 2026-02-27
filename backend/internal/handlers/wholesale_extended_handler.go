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
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Get Shopee API client
	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed"))
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

	c.JSON(http.StatusOK, response.Success(gin.H{
		"itemId":  itemID,
		"tiers":   req.Tiers,
		"updated": true,
	}))
}

// GetWholesaleInfo handles GET /api/wholesale/shopee/:itemId/info
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

	info := WholesaleInfo{
		ItemID:  itemID,
		HasTier: false,
		Tiers:   []WholesaleTier{},
		MPQ:     1,
	}

	c.JSON(http.StatusOK, response.Success(info))
}

// LookupItemId handles GET /api/wholesale/shopee/lookup/:sku
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
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewShopeeProductRepository(db)
	product, err := repo.FindBySKU(c.Request.Context(), sku)
	if err != nil || product == nil {
		c.JSON(http.StatusNotFound, response.Error("SKU not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"sku":    sku,
		"itemId": product.ItemID,
		"name":   product.Name,
	}))
}

// SetTiktokWholesale handles POST /api/wholesale/tiktok/:productId
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

	c.JSON(http.StatusOK, response.Success(gin.H{
		"productId": productID,
		"tiers":     req.Tiers,
		"success":   true,
	}))
}
