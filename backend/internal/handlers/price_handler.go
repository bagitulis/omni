package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/price"
	"gorm.io/gorm"
)

// PriceHandler handles price management endpoints
type PriceHandler struct {
	db *gorm.DB
}

// NewPriceHandler creates a new price handler
func NewPriceHandler(db *gorm.DB) *PriceHandler {
	return &PriceHandler{db: db}
}

// List handles GET /api/price
func (h *PriceHandler) List(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	svc := price.NewPriceService(h.db, tenantID)
	prices, total, err := svc.GetAllPrices(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total":    total,
			"returned": len(prices),
			"offset":   offset,
			"items":    prices,
		},
	})
}

// GetBySKU handles GET /api/price/:sku
func (h *PriceHandler) GetBySKU(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU required"})
		return
	}

	svc := price.NewPriceService(h.db, tenantID)
	priceInfo, err := svc.GetPrice(c.Request.Context(), sku)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if priceInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SKU not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": priceInfo})
}

// Update handles PUT /api/price/:sku
func (h *PriceHandler) Update(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU required"})
		return
	}

	var req struct {
		Price     float64  `json:"price" binding:"required"`
		Platforms []string `json:"platforms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := price.NewPriceService(h.db, tenantID)
	result, err := svc.UpdatePrice(c.Request.Context(), price.PriceUpdateRequest{
		SKU:       sku,
		Price:     req.Price,
		Platforms: req.Platforms,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// BulkUpdate handles POST /api/price/bulk
func (h *PriceHandler) BulkUpdate(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	var req price.BulkPriceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svc := price.NewPriceService(h.db, tenantID)
	result, err := svc.BulkUpdatePrice(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
