package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/orders"
	"github.com/omni/backend/internal/services/sync"
)

// SyncOrdersToday syncs processed orders with tracking info and saves to order_today_items
// POST /api/orders/today - Syncs fresh data from all platforms
// @Summary Sync and save today's orders with tracking info
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/today [post]
func (h *OrderManagerHandler) SyncOrdersToday(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
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

	service, err := sync.GetOrderSyncService(tenantID)
	if err != nil {
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

	orderManagerLogger.WithTenantID(tenantID).Info("Starting Order Today sync")

	// Sync processed orders from all platforms
	// IMPORTANT: SyncByCategory returns SyncResult which contains Orders WITH tracking info
	syncResults, err := service.SyncByCategory(c.Request.Context(), sync.OrderStatusCategory("processed"), 7, nil)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Warn("Failed to sync processed: " + err.Error())
	}

	// Collect all orders from sync results (these have tracking info)
	orderTodayItems := make([]orders.OrderTodayData, 0)

	for platform, result := range syncResults {
		if !result.Success || len(result.Orders) == 0 {
			continue
		}

		orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
			"platform":    platform,
			"order_count": len(result.Orders),
		}).Info("Processing orders from sync result")

		for _, order := range result.Orders {
			// Each order may have items - flatten them
			if len(order.Items) > 0 {
				for _, item := range order.Items {
					qty := item.Quantity
					if qty <= 0 {
						qty = 1
					}

					// Use item-level tracking if available (Lazada), fallback to order-level (Shopee/TikTok)
					trackingNo := item.TrackingNumber
					if trackingNo == "" {
						trackingNo = order.TrackingNumber
					}
					carrier := item.ShippingCarrier
					if carrier == "" {
						carrier = order.ShippingCarrier
					}

					orderTodayItems = append(orderTodayItems, orders.OrderTodayData{
						Platform:      order.Platform,
						OrderSN:       order.OrderNo,
						TrackingNo:    trackingNo,
						Courier:       carrier,
						SellerSku:     item.SKU,
						ProductName:   item.ProductName,
						VariationName: item.VariationName,
						Quantity:      qty,
					})
				}
			} else {
				// Flattened format - order fields contain item data
				qty := order.Quantity
				if qty <= 0 {
					qty = 1
				}
				orderTodayItems = append(orderTodayItems, orders.OrderTodayData{
					Platform:      order.Platform,
					OrderSN:       order.OrderNo,
					TrackingNo:    order.TrackingNumber,
					Courier:       order.ShippingCarrier,
					SellerSku:     order.SKU,
					ProductName:   order.ProductName,
					VariationName: order.VariationName,
					Quantity:      qty,
				})
			}
		}
	}

	// Save to order_today_items table
	orderTodayService := orders.NewOrderTodayService(db)
	savedCount, err := orderTodayService.SaveOrderTodayItems(c.Request.Context(), tenantID, orderTodayItems)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to save order today items: " + err.Error())
	}

	// Get the saved items from database - these have proper JSON tags (camelCase)
	savedItems, err := orderTodayService.GetOrderTodayItems(c.Request.Context(), tenantID)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to retrieve saved items: " + err.Error())
		savedItems = []orders.OrderTodayItem{}
	}

	// Get platform counts
	platformCounts, _ := orderTodayService.GetPlatformCounts(c.Request.Context(), tenantID)

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"total_items":     len(orderTodayItems),
		"saved_count":     savedCount,
		"platform_counts": platformCounts,
	}).Info("Order Today sync completed")

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"items":           savedItems,
		"data":            savedItems,
		"count":           savedCount,
		"saved_count":     savedCount,
		"platform_counts": platformCounts,
		"message":         "Order Today synced successfully",
	})
}

// GetOrdersToday retrieves saved order today items from database
// GET /api/orders/today - Returns previously synced data
// @Summary Get saved today's orders
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/today [get]
func (h *OrderManagerHandler) GetOrdersToday(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant ID",
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

	orderTodayService := orders.NewOrderTodayService(db)
	items, err := orderTodayService.GetOrderTodayItems(c.Request.Context(), tenantID)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to get order today items: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get order today items: " + err.Error(),
			"code":    "DATABASE_ERROR",
			"items":   []interface{}{},
			"data":    []interface{}{},
			"count":   0,
		})
		return
	}

	// Get platform counts
	platformCounts, _ := orderTodayService.GetPlatformCounts(c.Request.Context(), tenantID)

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"count": len(items),
	}).Info("Retrieved order today items")

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"items":           items,
		"data":            items,
		"count":           len(items),
		"platform_counts": platformCounts,
	})
}
