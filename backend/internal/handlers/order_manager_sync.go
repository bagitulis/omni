package handlers

import (
	"fmt"
	"net/http"
	"strings"

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
			"error":   "Missing tenantId",
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
	hasFailure := false

	for _, cat := range categories {
		catResults, err := service.SyncByCategory(
			c.Request.Context(),
			sync.OrderStatusCategory(cat),
			7,
			nil,
		)
		if err != nil {
			hasFailure = true
			results[cat] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
		} else {
			categorySuccess := true
			categoryErrors := make([]string, 0)
			for platform, result := range catResults {
				if !result.Success {
					categorySuccess = false
					errMsg := result.Error
					if errMsg == "" {
						errMsg = "unknown platform sync error"
					}
					categoryErrors = append(categoryErrors, fmt.Sprintf("%s: %s", platform, errMsg))
				}
			}

			if !categorySuccess {
				hasFailure = true
				results[cat] = map[string]interface{}{
					"success": false,
					"error":   strings.Join(categoryErrors, "; "),
					"data":    catResults,
				}
				continue
			}

			results[cat] = map[string]interface{}{
				"success": true,
				"data":    catResults,
			}
		}
	}

	if hasFailure {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"code":    "PARTIAL_SYNC_FAILURE",
			"error":   "One or more categories failed to sync",
			"data":    results,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}
