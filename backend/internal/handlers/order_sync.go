package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
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
func (h *OrderSyncHandler) SyncByCategory(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
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

	days := 7
	if daysQuery := c.Query("days"); daysQuery != "" {
		if parsed, err := strconv.Atoi(daysQuery); err == nil {
			days = parsed
		}
	}
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

	if !service.IsInitialized() {
		orderSyncLogger.WithTenantID(tenantID).Warn("Order sync service not initialized - OAuth configuration required")
		c.JSON(http.StatusOK, gin.H{
			"success":  false,
			"code":     "SERVICE_NOT_INITIALIZED",
			"error":    "Platform OAuth configuration required - please complete OAuth setup in Settings",
			"data":     map[string]interface{}{},
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

	for platform, result := range results {
		orderSyncLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
			"platform": platform,
			"success":  result.Success,
			"count":    result.Count,
			"error":    result.Error,
		}).Info("Platform sync result")
	}

	hasPlatformFailure := false
	for _, result := range results {
		if !result.Success {
			hasPlatformFailure = true
			break
		}
	}

	if hasPlatformFailure {
		c.JSON(http.StatusOK, gin.H{
			"success":  false,
			"code":     "PARTIAL_SYNC_FAILURE",
			"error":    "One or more platforms failed to sync",
			"data":     results,
			"category": category,
			"days":     days,
		})
		return
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
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
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

	if !service.IsInitialized() {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"code":    "SERVICE_NOT_INITIALIZED",
			"error":   "Platform not configured - complete OAuth setup first",
			"data":    []interface{}{},
			"count":   0,
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
