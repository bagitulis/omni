package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	inventoryService "github.com/omni/backend/internal/services/inventory"
)

// PriceRecommendations handles GET /api/inventory/price-recommendations
// Returns per-platform prices for requested SKUs
func (h *InventoryHandler) PriceRecommendations(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	// Get SKUs from query param (comma-separated)
	skusParam := c.Query("skus")
	if skusParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing skus parameter"})
		return
	}

	skus := strings.Split(skusParam, ",")
	if len(skus) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "No SKUs provided"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Build result map: SKU -> per-platform prices
	result := make(map[string]map[string]float64)

	for _, sku := range skus {
		sku = strings.TrimSpace(sku)
		if sku == "" {
			continue
		}

		var record models.InventoryRecord
		err := db.WithContext(c.Request.Context()).
			Where("tenant_id = ? AND key_value = ?", tenantID, sku).
			First(&record).Error

		if err != nil {
			// SKU not found in inventory - skip it entirely
			// Frontend will use fallback (current product price)
			continue
		}

		// Get per-platform prices
		prices := inventoryService.GetPricePerPlatform(record)
		result[sku] = map[string]float64{
			"shopee":     prices["shopee"],
			"tiktok":     prices["tiktok"],
			"lazada":     prices["lazada"],
			"base_price": prices["base"],
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}
