package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/orders"
	"github.com/omni/backend/internal/services/sync"
)

// AggregatedLockedItem represents an aggregated locked order item
type AggregatedLockedItem struct {
	SKU           string   `json:"sku"`
	ProductName   string   `json:"product_name"`
	VariationName string   `json:"variation_name,omitempty"`
	Qty           int      `json:"qty"`
	Platforms     []string `json:"platforms"`
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
			"error":   "Missing tenant_id",
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

	if len(unprocessOrders) == 0 && len(processedOrders) == 0 && err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   "Failed to sync locked orders: " + err.Error(),
			"code":    "SYNC_FAILED",
			"items":   []interface{}{},
			"count":   0,
		})
		return
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

	// Update inventory_records with Locked and Sellable columns
	// Build SKU→qty map from locked items
	lockedBySku := make(map[string]int, len(lockedItems))
	for _, item := range lockedItems {
		if item.SKU != "" {
			lockedBySku[item.SKU] += item.Qty
		}
	}

	// Get inventory settings to find the Total column name
	var invSettings models.InventorySettings
	totalColumnName := "Total" // default
	if err := db.WithContext(c.Request.Context()).
		Where("tenant_id = ?", tenantID).
		First(&invSettings).Error; err == nil && invSettings.KeyColumn != "" {
		totalColumnName = invSettings.KeyColumn
	}

	// Fetch all inventory records for this tenant
	var inventoryRecords []models.InventoryRecord
	if err := db.WithContext(c.Request.Context()).
		Where("tenant_id = ?", tenantID).
		Find(&inventoryRecords).Error; err != nil {
		orderManagerLogger.WithTenantID(tenantID).Warn("Failed to fetch inventory records for lock update: " + err.Error())
	} else {
		updatedCount := 0
		for _, record := range inventoryRecords {
			// Parse JSONB data
			var dataMap map[string]interface{}
			if err := json.Unmarshal([]byte(record.Data), &dataMap); err != nil {
				continue
			}

			// Get Total value from the configured column
			totalVal := 0.0
			if val, ok := dataMap[totalColumnName]; ok {
				switch v := val.(type) {
				case float64:
					totalVal = v
				case string:
					if parsed, e := strconv.ParseFloat(v, 64); e == nil {
						totalVal = parsed
					}
				}
			}

			// Set Locked and Sellable
			lockedQty := lockedBySku[record.KeyValue]
			sellable := int(totalVal) - lockedQty
			if sellable < 0 {
				sellable = 0
			}

			dataMap["Locked"] = lockedQty
			dataMap["Sellable"] = sellable

			// Save back
			updated, err := json.Marshal(dataMap)
			if err != nil {
				continue
			}

			if err := db.WithContext(c.Request.Context()).
				Model(&record).
				Update("data", string(updated)).Error; err != nil {
				orderManagerLogger.WithTenantID(tenantID).Warn("Failed to update inventory record " + record.KeyValue + ": " + err.Error())
				continue
			}
			updatedCount++
		}

		orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
			"updated_records": updatedCount,
			"total_records":   len(inventoryRecords),
			"locked_skus":     len(lockedBySku),
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
			"error":   "Missing tenant_id",
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
