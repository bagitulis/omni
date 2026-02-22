package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/omni/backend/internal/services/products"
	"gorm.io/gorm"
)

// ProductMasterHandler handles product master endpoints
type ProductMasterHandler struct {
	fallbackDB *gorm.DB
	systemDB   *gorm.DB
}

// NewProductMasterHandler creates a new product master handler
func NewProductMasterHandler(db *gorm.DB) *ProductMasterHandler {
	return &ProductMasterHandler{fallbackDB: db}
}

// SetSystemDB sets the system database for platform credential lookups
func (h *ProductMasterHandler) SetSystemDB(db *gorm.DB) {
	h.systemDB = db
}

// getDB returns the appropriate database for the current request
func (h *ProductMasterHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// GetMasterProductList returns master products
// @Summary Get master product list
// @Tags Products
// @Param platform query string false "Platform filter"
// @Param status query string false "Status filter"
// @Param sku query string false "SKU filter"
// @Param name query string false "Name filter"
// @Param limit query int false "Limit" default(100)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} map[string]interface{}
// @Router /api/products/master [get]
func (h *ProductMasterHandler) GetMasterProductList(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get tenant database: " + err.Error(),
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	params := products.ProductQueryParams{
		Platform: c.Query("platform"),
		Status:   products.ProductStatus(c.Query("status")),
		SKU:      c.Query("sku"),
		Name:     c.Query("name"),
		Limit:    limit,
		Offset:   offset,
	}

	service := products.NewProductMasterService(db)
	productList, total, err := service.GetProducts(c.Request.Context(), tenantID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"products": productList,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetMasterProductStats returns product statistics
// @Summary Get master product statistics
// @Tags Products
// @Success 200 {object} map[string]interface{}
// @Router /api/products/master/stats [get]
func (h *ProductMasterHandler) GetMasterProductStats(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get tenant database: " + err.Error(),
		})
		return
	}

	service := products.NewProductMasterService(db)
	stats, err := service.GetProductStats(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// GetProductByID returns a product by item ID
// @Summary Get product by item ID
// @Tags Products
// @Param platform path string true "Platform"
// @Param itemId path string true "Item ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/products/{platform}/{itemId} [get]
func (h *ProductMasterHandler) GetProductByID(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	platform := c.Param("platform")
	itemID := c.Param("itemId")

	if platform == "" || itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing platform or item ID",
		})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get tenant database: " + err.Error(),
		})
		return
	}

	service := products.NewProductMasterService(db)
	product, err := service.GetProductByItemID(c.Request.Context(), tenantID, platform, itemID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Product not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"product": product,
	})
}

// SyncSelected handles POST /api/products/master/sync-selected
// Syncs selected products from marketplace APIs to refresh price/stock.
func (h *ProductMasterHandler) SyncSelected(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant ID"})
		return
	}

	var req struct {
		ProductIDs []uint `json:"product_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "product_ids required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	systemDB := h.systemDB
	if systemDB == nil {
		systemDB = h.fallbackDB
	}

	svc := masterProductService.NewService(db)
	result, err := svc.SyncSelectedProducts(c.Request.Context(), tenantID, req.ProductIDs, systemDB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	log.Printf("[SyncSelected] Result: synced=%d, failed=%d, details=%v", result.Synced, result.Failed, result.Details)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
