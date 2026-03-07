package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/wholesale"
)

// Preview handles POST /api/wholesale/preview
func (h *WholesaleBatchHandler) Preview(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	settings, err := service.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get settings"))
		return
	}

	var results []gin.H
	for _, sku := range req.SKUs {
		basePrice := float64(100000)
		if req.BasePrice > 0 {
			basePrice = req.BasePrice
		}
		tiers := service.CalculateTiersFromSettings(basePrice, settings)
		results = append(results, gin.H{
			"sku":        sku,
			"base_price": basePrice,
			"tiers":      tiers,
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"results":       results,
		"settings_used": settings,
	}))
}
