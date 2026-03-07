package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/orders"
	"github.com/omni/backend/internal/services/sync"
)

// GetLockedTodayOrders syncs orders and aggregates locked orders for today - POST endpoint
// Time-based logic:
// - 00:00-14:00 (before 2pm Jakarta): unprocess + processed orders
// - 14:00-24:00 (after 2pm Jakarta): unprocess orders only
// @Summary Sync and save locked orders today
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/locked-today [post]
func (h *OrderManagerHandler) GetLockedTodayOrders(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	// Get tenant database
	db, err := GetTenantDB(c)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to get tenant DB: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}

	// Get order sync service
	syncService, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Warn("Order sync service not available: " + err.Error())
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Order sync service not available: " + err.Error(),
			"code":    "SERVICE_UNAVAILABLE",
			"items":   []interface{}{},
			"count":   0,
		})
		return
	}

	// Time-based logic: Check current hour (Jakarta timezone UTC+7)
	now := time.Now().UTC()
	jakartaHour := (now.Hour() + 7) % 24
	includeProcessed := jakartaHour < 14 // Before 2pm include processed

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"jakarta_hour":      jakartaHour,
		"include_processed": includeProcessed,
	}).Info("Processing locked orders")

	// Sync and get unprocess orders
	_, err = syncService.SyncByCategory(c.Request.Context(), sync.OrderStatusCategory("unprocess"), 7, nil)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Warn("Failed to sync unprocess: " + err.Error())
	}
	unprocessOrders, unprocessErr := syncService.GetOrdersByCategory(c.Request.Context(), sync.OrderStatusCategory("unprocess"), nil)
	if unprocessErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get unprocess orders: " + unprocessErr.Error(),
			"code":    "ORDER_FETCH_FAILED",
			"items":   []interface{}{},
			"count":   0,
		})
		return
	}

	// Sync and get processed orders (only before 2pm)
	var processedOrders []sync.Order
	if includeProcessed {
		_, err = syncService.SyncByCategory(c.Request.Context(), sync.OrderStatusCategory("processed"), 7, nil)
		if err != nil {
			orderManagerLogger.WithTenantID(tenantID).Warn("Failed to sync processed: " + err.Error())
		}
		processedOrders, err = syncService.GetOrdersByCategory(c.Request.Context(), sync.OrderStatusCategory("processed"), nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to get processed orders: " + err.Error(),
				"code":    "ORDER_FETCH_FAILED",
				"items":   []interface{}{},
				"count":   0,
			})
			return
		}
	}

	// Aggregate locked orders by SKU + ProductName + Variation
	lockedItems := aggregateLockedOrders(unprocessOrders, processedOrders)

	// Save to database
	lockedService := orders.NewLockedOrderService(db)
	savedItems := make([]orders.LockedOrderItem, len(lockedItems))
	for i, item := range lockedItems {
		savedItems[i] = orders.LockedOrderItem{
			SKU:           item.SKU,
			ProductName:   item.ProductName,
			VariationName: item.VariationName,
			Qty:           item.Qty,
		}
	}

	savedCount, err := lockedService.SaveLockedOrders(c.Request.Context(), tenantID, savedItems)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to save locked orders: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save locked orders: " + err.Error(),
			"code":    "DATABASE_ERROR",
			"items":   []interface{}{},
			"count":   0,
		})
		return
	}

	totalQty, err := lockedService.GetTotalQty(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get total locked quantity: " + err.Error(),
			"code":    "DATABASE_ERROR",
			"items":   []interface{}{},
			"count":   0,
		})
		return
	}

	// Recalculate Locked/Sellable in inventory_records using locked_orders table.
	// This reads the just-saved locked orders and updates Sellable = rawTotal - Locked.
	recalcResult, recalcErr := orders.RecalculateLockedSellable(c.Request.Context(), db, tenantID)
	if recalcErr != nil {
		orderManagerLogger.WithTenantID(tenantID).Warn("Failed to recalculate Locked/Sellable: " + recalcErr.Error())
	} else if recalcResult != nil {
		orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
			"updated_records": recalcResult.UpdatedRecords,
			"total_records":   recalcResult.TotalRecords,
			"locked_skus":     recalcResult.LockedSKUs,
			"total_column":    recalcResult.TotalColumn,
		}).Info("Updated inventory records with Locked/Sellable columns")
	}

	mode := "unprocess-only"
	message := "After 14:00 Jakarta time - counting unprocess orders only"
	if includeProcessed {
		mode = "unprocess+processed"
		message = "Before 14:00 Jakarta time - counting unprocess and processed orders"
	}

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"mode":        mode,
		"saved_count": savedCount,
		"total_qty":   totalQty,
	}).Info("Locked orders saved successfully")

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"items":        lockedItems,
		"count":        len(lockedItems),
		"saved_count":  savedCount,
		"total_qty":    totalQty,
		"mode":         mode,
		"jakarta_hour": jakartaHour,
		"message":      message,
		"tenant_id":    tenantID,
		"days":         7,
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
			"error":   "Missing tenantId",
		})
		return
	}

	// Get tenant database
	db, err := GetTenantDB(c)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to get tenant DB: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}

	// Get locked orders from database
	lockedService := orders.NewLockedOrderService(db)
	lockedOrders, err := lockedService.GetLockedOrders(c.Request.Context(), tenantID)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to get locked orders: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":   false,
			"error":     "Failed to get locked orders: " + err.Error(),
			"code":      "DATABASE_ERROR",
			"items":     []interface{}{},
			"count":     0,
			"total_qty": 0,
			"tenant_id": tenantID,
			"days":      7,
		})
		return
	}

	totalQty, err := lockedService.GetTotalQty(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get total locked quantity: " + err.Error(),
			"code":    "DATABASE_ERROR",
			"items":   []interface{}{},
			"count":   0,
		})
		return
	}

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"count":     len(lockedOrders),
		"total_qty": totalQty,
	}).Info("Retrieved locked orders from database")

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"items":     lockedOrders,
		"count":     len(lockedOrders),
		"total_qty": totalQty,
		"tenant_id": tenantID,
		"days":      7,
	})
}
