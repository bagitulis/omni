package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/orders"
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
	unprocessOrders, _ := syncService.GetOrdersByCategory(c.Request.Context(), sync.OrderStatusCategory("unprocess"), nil)

	// Sync and get processed orders (only before 2pm)
	var processedOrders []sync.Order
	if includeProcessed {
		_, err = syncService.SyncByCategory(c.Request.Context(), sync.OrderStatusCategory("processed"), 7, nil)
		if err != nil {
			orderManagerLogger.WithTenantID(tenantID).Warn("Failed to sync processed: " + err.Error())
		}
		processedOrders, _ = syncService.GetOrdersByCategory(c.Request.Context(), sync.OrderStatusCategory("processed"), nil)
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
	}

	totalQty, _ := lockedService.GetTotalQty(c.Request.Context(), tenantID)

	mode := "unprocess-only"
	message := "Setelah jam 14:00 - hanya menghitung unprocess orders"
	if includeProcessed {
		mode = "unprocess+processed"
		message = "Sebelum jam 14:00 - menghitung unprocess + processed orders"
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

	totalQty, _ := lockedService.GetTotalQty(c.Request.Context(), tenantID)

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

// AggregatedLockedItem represents an aggregated locked order item
type AggregatedLockedItem struct {
	SKU           string   `json:"sku"`
	ProductName   string   `json:"product_name"`
	VariationName string   `json:"variation_name,omitempty"`
	Qty           int      `json:"qty"`
	Platforms     []string `json:"platforms"`
}

// aggregateLockedOrders aggregates orders by SKU + ProductName + Variation
// Groups items and sums quantities, tracking which platforms they came from
// NOTE: Orders are flattened (each row = one item), so we read from order fields directly
func aggregateLockedOrders(unprocessOrders, processedOrders []sync.Order) []AggregatedLockedItem {
	// Map to aggregate by key (SKU|ProductName|Variation)
	aggregated := make(map[string]*AggregatedLockedItem)
	platformSets := make(map[string]map[string]bool)

	// Process all orders (each order is a flattened item row)
	allOrders := append(unprocessOrders, processedOrders...)

	for _, order := range allOrders {
		// In flattened format, item data is directly on the order object
		sku := order.SKU
		productName := order.ProductName
		variationName := order.VariationName
		qty := order.Quantity // This is the "qty" field in flattened format
		platform := order.Platform

		if sku == "" && productName == "" {
			continue // Skip empty items
		}

		// Ensure minimum qty of 1 if item exists
		if qty <= 0 {
			qty = 1
		}

		key := sku + "|" + productName + "|" + variationName

		if _, exists := aggregated[key]; !exists {
			aggregated[key] = &AggregatedLockedItem{
				SKU:           sku,
				ProductName:   productName,
				VariationName: variationName,
				Qty:           0,
				Platforms:     []string{},
			}
			platformSets[key] = make(map[string]bool)
		}

		aggregated[key].Qty += qty
		platformSets[key][platform] = true
	}

	// Convert map to slice and add platforms
	result := make([]AggregatedLockedItem, 0, len(aggregated))
	for key, item := range aggregated {
		platforms := make([]string, 0, len(platformSets[key]))
		for p := range platformSets[key] {
			platforms = append(platforms, p)
		}
		item.Platforms = platforms
		result = append(result, *item)
	}

	// Sort by qty descending
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Qty > result[i].Qty {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
