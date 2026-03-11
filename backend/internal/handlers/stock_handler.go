package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/stock"
	"gorm.io/gorm"
)

// StockHandler handles stock management endpoints
type StockHandler struct {
	db *gorm.DB
}

// NewStockHandler creates a new stock handler
func NewStockHandler(db *gorm.DB) *StockHandler {
	return &StockHandler{db: db}
}

// List handles GET /api/stock
func (h *StockHandler) List(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	lowStockOnly := c.Query("low_stock") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	svc := stock.NewStockService(h.db, tenantID)
	stocks, total, err := svc.GetAllStock(c.Request.Context(), lowStockOnly, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":    total,
			"returned": len(stocks),
			"offset":   offset,
			"items":    stocks,
		},
	})
}

// GetBySKU handles GET /api/stock/:sku
func (h *StockHandler) GetBySKU(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "SKU required"})
		return
	}

	svc := stock.NewStockService(h.db, tenantID)
	stockInfo, err := svc.GetStock(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	if stockInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "SKU not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stockInfo})
}

// Update handles PUT /api/stock/:sku
func (h *StockHandler) Update(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "SKU required"})
		return
	}

	var req struct {
		Quantity  int      `json:"quantity" binding:"required"`
		Platforms []string `json:"platforms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := stock.NewStockService(h.db, tenantID)
	result, err := svc.UpdateStock(c.Request.Context(), models.StockUpdateRequest{
		SKU:       sku,
		Quantity:  req.Quantity,
		Platforms: req.Platforms,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// BulkUpdate handles POST /api/stock/bulk
func (h *StockHandler) BulkUpdate(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	var req models.BulkStockUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	svc := stock.NewStockService(h.db, tenantID)
	result, err := svc.BulkUpdateStock(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// GetAlerts handles GET /api/stock/alerts
func (h *StockHandler) GetAlerts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	svc := stock.NewStockService(h.db, tenantID)
	alerts, err := svc.GetLowStockAlerts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": alerts})
}
