package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/sync"
	"github.com/omni/backend/internal/utils/logger"
)

var orderSyncLogger = logger.Named("OrderSyncHandler")

// OrderSyncHandler handles order sync endpoints
// SRP: Handles sync operations from platform APIs to database
type OrderSyncHandler struct{}

// NewOrderSyncHandler creates a new order sync handler
func NewOrderSyncHandler() *OrderSyncHandler {
	return &OrderSyncHandler{}
}

// SyncByCategoryBody represents the request body for sync by category
type SyncByCategoryBody struct {
	Days int `json:"days"`
}

// SyncByCategory syncs orders by category from platform APIs
// This is called by frontend BEFORE calling GET /orders/:category
// @Summary Sync orders by category
// @Tags Orders
// @Param category path string true "Order category"
// @Param days query int false "Days to sync" default(7)
// @Param platforms query []string false "Platforms to sync"
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/sync/{category} [post]
func (h *OrderSyncHandler) SyncByCategory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing category",
		})
		return
	}

	// Parse days from query or body
	days := 7
	if daysQuery := c.Query("days"); daysQuery != "" {
		if parsed, err := strconv.Atoi(daysQuery); err == nil {
			days = parsed
		}
	}
	// Also try to parse from body
	var body SyncByCategoryBody
	if err := c.ShouldBindJSON(&body); err == nil && body.Days > 0 {
		days = body.Days
	}

	platformsQuery := c.QueryArray("platforms")

	var platforms []sync.PlatformType
	if len(platformsQuery) > 0 {
		for _, p := range platformsQuery {
			platforms = append(platforms, sync.PlatformType(p))
		}
	}

	orderSyncLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"category":  category,
		"days":      days,
		"platforms": platforms,
	}).Info("Starting order sync by category")

	service, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		orderSyncLogger.WithTenantID(tenantID).Error("Failed to get order sync service: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// If service not initialized, return partial success indicating sync requires platform config
	if !service.IsInitialized() {
		orderSyncLogger.WithTenantID(tenantID).Warn("Order sync service not initialized - OAuth configuration required")
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"data":     map[string]interface{}{},
			"message":  "Platform OAuth configuration required - please complete OAuth setup in Settings",
			"category": category,
			"days":     days,
		})
		return
	}

	results, err := service.SyncByCategory(
		c.Request.Context(),
		sync.OrderStatusCategory(category),
		days,
		platforms,
	)
	if err != nil {
		orderSyncLogger.WithTenantID(tenantID).Error("Sync failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Log sync results
	for platform, result := range results {
		orderSyncLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
			"platform": platform,
			"success":  result.Success,
			"count":    result.Count,
			"error":    result.Error,
		}).Info("Platform sync result")
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data":     results,
		"category": category,
		"days":     days,
	})
}

// SyncPlatformOrders syncs orders for a specific platform
func (h *OrderSyncHandler) SyncPlatformOrders(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	platform := c.Param("platform")
	category := c.Query("category")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	if platform == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing platform or category",
		})
		return
	}

	service, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// If service not initialized, return response indicating platform needs OAuth config
	if !service.IsInitialized() {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
			"count":   0,
			"message": "Platform not configured - complete OAuth setup first",
		})
		return
	}

	orders, err := service.SyncPlatformOrders(
		c.Request.Context(),
		sync.PlatformType(platform),
		sync.OrderStatusCategory(category),
		days,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"count":   len(orders),
	})
}

// GetOrdersByCategory gets orders by category
func (h *OrderSyncHandler) GetOrdersByCategory(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	category := c.Param("category")
	platformStr := c.Query("platform")

	service, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	var platform *sync.PlatformType
	if platformStr != "" {
		p := sync.PlatformType(platformStr)
		platform = &p
	}

	orders, err := service.GetOrdersByCategory(
		c.Request.Context(),
		sync.OrderStatusCategory(category),
		platform,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"count":   len(orders),
	})
}

// GetOrderDetails gets detailed order information
func (h *OrderSyncHandler) GetOrderDetails(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
		})
		return
	}

	platform := c.Param("platform")
	orderIDs := c.QueryArray("orderIds")

	if platform == "" || len(orderIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing platform or order IDs",
		})
		return
	}

	service, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	orders, err := service.GetOrderDetails(
		c.Request.Context(),
		sync.PlatformType(platform),
		orderIDs,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
		"count":   len(orders),
	})
}
