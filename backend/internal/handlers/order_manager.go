package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/sync"
	"github.com/omni/backend/internal/utils/logger"
)

var orderManagerLogger = logger.Named("OrderManagerHandler")

// OrderManagerHandler handles order manager endpoints for frontend compatibility
// SRP: Frontend-facing endpoints for Order Manager UI
type OrderManagerHandler struct{}

// NewOrderManagerHandler creates a new order manager handler
func NewOrderManagerHandler() *OrderManagerHandler {
	return &OrderManagerHandler{}
}

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
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"items":   []interface{}{},
			"data":    []interface{}{},
			"count":   0,
			"message": "Order sync service not available: " + err.Error(),
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
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"items":   []interface{}{},
			"data":    []interface{}{},
			"count":   0,
			"message": "Failed to get orders: " + err.Error(),
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

// GetLockedTodayOrders gets locked orders for today (used by OrderManager) - POST endpoint
// @Summary Get locked orders today
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/locked-today [post]
func (h *OrderManagerHandler) GetLockedTodayOrders(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	// Return empty for now - this will be populated from locked_orders table
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   []interface{}{},
		"count":   0,
	})
}

// GetSavedLockedOrders retrieves previously saved locked orders - GET endpoint
// @Summary Get saved locked orders
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/locked-today [get]
func (h *OrderManagerHandler) GetSavedLockedOrders(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	// Return empty for now - will be implemented with LockedOrder table
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"items":    []interface{}{},
		"count":    0,
		"totalQty": 0,
		"tenantId": tenantID,
		"days":     7,
	})
}

// GetOrdersToday gets today's orders
// @Summary Get today's orders
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/today [post]
func (h *OrderManagerHandler) GetOrdersToday(c *gin.Context) {
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
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"items":   []interface{}{},
			"count":   0,
			"message": "Order sync service not available",
		})
		return
	}

	// Get orders from all categories for today
	var allOrders []interface{}
	categories := []sync.OrderStatusCategory{"unpaid", "unprocess", "processed"}

	for _, cat := range categories {
		orders, err := service.GetOrdersByCategory(c.Request.Context(), cat, nil)
		if err == nil {
			for _, o := range orders {
				allOrders = append(allOrders, o)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"items":   allOrders,
		"count":   len(allOrders),
	})
}

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
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    map[string]interface{}{},
			"message": "Order sync service not available",
		})
		return
	}

	if !service.IsInitialized() {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    map[string]interface{}{},
			"message": "Order sync service not initialized",
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
