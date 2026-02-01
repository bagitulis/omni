package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/sync"
)

// SyncAll syncs all orders across platforms
// @Summary Sync all orders
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/sync-all [post]
func (h *OrderManagerHandler) SyncAll(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	service, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Order sync service not available: " + err.Error(),
			"code":    "SERVICE_UNAVAILABLE",
			"data":    map[string]interface{}{},
		})
		return
	}

	if !service.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Order sync service not initialized",
			"code":    "SERVICE_NOT_INITIALIZED",
			"data":    map[string]interface{}{},
		})
		return
	}

	// Sync all categories
	results := make(map[string]interface{})
	categories := []string{"unpaid", "unprocess", "processed"}

	for _, cat := range categories {
		catResults, err := service.SyncByCategory(
			c.Request.Context(),
			sync.OrderStatusCategory(cat),
			7,
			nil,
		)
		if err != nil {
			results[cat] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		} else {
			results[cat] = map[string]interface{}{
				"success": true,
				"data":    catResults,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}
