package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/sync"
)

// GetUnpaidOrders gets unpaid orders from database
// NOTE: Frontend calls POST /sync/unpaid first, then this GET endpoint
// @Summary Get unpaid orders
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/unpaid [get]
func (h *OrderManagerHandler) GetUnpaidOrders(c *gin.Context) {
	h.getOrdersFromDatabase(c, "unpaid")
}

// GetUnprocessOrders gets unprocessed orders from database
// NOTE: Frontend calls POST /sync/unprocess first, then this GET endpoint
// @Summary Get unprocessed orders
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/unprocess [get]
func (h *OrderManagerHandler) GetUnprocessOrders(c *gin.Context) {
	h.getOrdersFromDatabase(c, "unprocess")
}

// GetProcessedOrders gets processed orders from database
// NOTE: Frontend calls POST /sync/processed first, then this GET endpoint
// @Summary Get processed orders
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/processed [get]
func (h *OrderManagerHandler) GetProcessedOrders(c *gin.Context) {
	h.getOrdersFromDatabase(c, "processed")
}

// getOrdersFromDatabase gets orders from database without syncing
// IMPORTANT: Does NOT sync from platform APIs - frontend should call /sync/:category first
// This separation allows proper control of when to sync vs when to just fetch
func (h *OrderManagerHandler) getOrdersFromDatabase(c *gin.Context, category string) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"category": category,
	}).Info("Fetching orders from database")

	service, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Warn("Order sync service not available: " + err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Order sync service not available: " + err.Error(),
			"code":    "SERVICE_UNAVAILABLE",
			"items":   []interface{}{},
			"data":    []interface{}{},
			"count":   0,
		})
		return
	}

	// Get orders from database (no sync - frontend already called /sync/:category)
	orders, err := service.GetOrdersByCategory(
		c.Request.Context(),
		sync.OrderStatusCategory(category),
		nil,
	)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to get orders: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get orders: " + err.Error(),
			"code":    "DATABASE_ERROR",
			"items":   []interface{}{},
			"data":    []interface{}{},
			"count":   0,
		})
		return
	}

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"category": category,
		"count":    len(orders),
	}).Info("Successfully fetched orders from database")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   orders,
		"data":    orders,
		"count":   len(orders),
	})
}
