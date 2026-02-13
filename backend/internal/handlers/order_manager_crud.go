package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
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

	platformFilter, err := parseOrderPlatformFilter(c.Query("platform"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get orders from database (no sync - frontend already called /sync/:category)
	orders, err := service.GetOrdersByCategory(
		c.Request.Context(),
		sync.OrderStatusCategory(category),
		platformFilter,
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

func parseOrderPlatformFilter(platform string) (*sync.PlatformType, error) {
	if platform == "" {
		return nil, nil
	}

	normalized := strings.ToLower(strings.TrimSpace(platform))
	if normalized == "" || normalized == "all" {
		return nil, nil
	}

	switch normalized {
	case string(sync.PlatformShopee), string(sync.PlatformLazada), string(sync.PlatformTiktok):
		selected := sync.PlatformType(normalized)
		return &selected, nil
	default:
		return nil, fmt.Errorf("invalid platform filter: %s", platform)
	}
}

// GetOrderByOrderSn gets order detail by order_sn
// Searches across all platforms and returns the order with items
// @Summary Get order detail by order_sn
// @Tags Orders
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/:orderSn [get]
func (h *OrderManagerHandler) GetOrderByOrderSn(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	orderSn := c.Param("orderSn")
	if orderSn == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing order_sn",
		})
		return
	}

	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"order_sn": orderSn,
	}).Info("Fetching order detail by order_sn")

	// Get database connection
	db, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		orderManagerLogger.WithTenantID(tenantID).Error("Failed to get database connection: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}

	// Try to find order in each platform table
	// Search Shopee
	var shopeeOrder models.ShopeeOrder
	if err := db.WithContext(c.Request.Context()).Where("order_sn = ?", orderSn).First(&shopeeOrder).Error; err == nil {
		// Found in Shopee, get items
		var items []models.ShopeeOrderItem
		db.WithContext(c.Request.Context()).Where("order_sn = ?", orderSn).Find(&items)

		// Convert to frontend format
		orderDetail := h.convertShopeeOrderToDetail(&shopeeOrder, items)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    orderDetail,
		})
		return
	}

	// Search Lazada
	var lazadaOrder models.LazadaOrder
	if err := db.WithContext(c.Request.Context()).Where("order_sn = ?", orderSn).First(&lazadaOrder).Error; err == nil {
		// Found in Lazada, get items
		var items []models.LazadaOrderItem
		db.WithContext(c.Request.Context()).Where("order_sn = ?", orderSn).Find(&items)

		// Convert to frontend format
		orderDetail := h.convertLazadaOrderToDetail(&lazadaOrder, items)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    orderDetail,
		})
		return
	}

	// Search TikTok
	var tiktokOrder models.TiktokOrder
	if err := db.WithContext(c.Request.Context()).Where("order_sn = ?", orderSn).First(&tiktokOrder).Error; err == nil {
		// Found in TikTok, get items
		var items []models.TiktokOrderItem
		db.WithContext(c.Request.Context()).Where("order_sn = ?", orderSn).Find(&items)

		// Convert to frontend format
		orderDetail := h.convertTiktokOrderToDetail(&tiktokOrder, items)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    orderDetail,
		})
		return
	}

	// Order not found in any platform
	orderManagerLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"order_sn": orderSn,
	}).Warn("Order not found")
	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Order not found",
	})
}

// Helper functions to convert platform orders to frontend format
func (h *OrderManagerHandler) convertShopeeOrderToDetail(order *models.ShopeeOrder, items []models.ShopeeOrderItem) map[string]interface{} {
	orderItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		orderItems[i] = map[string]interface{}{
			"item_id":   item.ItemID,
			"item_name": item.ItemName,
			"item_sku":  item.ItemSku,
			"quantity":  item.Quantity,
			"price":     item.Price,
			"total":     float64(*item.Quantity) * *item.Price,
		}
	}

	return map[string]interface{}{
		"id":               order.ID,
		"order_sn":         order.OrderSN,
		"order_no":         order.OrderSN,
		"order_status":     order.OrderStatus,
		"status":           order.OrderStatus,
		"platform":         "shopee",
		"category":         "", // Category not stored in order model
		"buyer_username":   order.BuyerUsername,
		"total_amount":     order.TotalAmount,
		"currency":         order.Currency,
		"payment_method":   order.PaymentMethod,
		"shipping_carrier": order.ShippingCarrier,
		"tracking_number":  order.TrackingNumber,
		"ship_by_date":     order.ShipByDate,
		"buyer_message":    order.BuyerMessage,
		"created_at":       order.CreatedAt,
		"updated_at":       order.UpdatedAt,
		"items":            orderItems,
	}
}

func (h *OrderManagerHandler) convertLazadaOrderToDetail(order *models.LazadaOrder, items []models.LazadaOrderItem) map[string]interface{} {
	orderItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		orderItems[i] = map[string]interface{}{
			"item_id":   item.ItemID,
			"item_name": item.ProductName,
			"item_sku":  item.SellerSku,
			"quantity":  item.Quantity,
			"price":     item.Price,
			"total":     float64(*item.Quantity) * *item.Price,
		}
	}

	return map[string]interface{}{
		"id":               order.ID,
		"order_sn":         order.OrderSN,
		"order_no":         order.OrderSN,
		"order_status":     order.OrderStatus,
		"status":           order.OrderStatus,
		"platform":         "lazada",
		"category":         "",
		"buyer_username":   order.BuyerUsername,
		"total_amount":     order.TotalAmount,
		"currency":         order.Currency,
		"payment_method":   order.PaymentMethod,
		"shipping_carrier": order.ShippingCarrier,
		"tracking_number":  order.TrackingNumber,
		"ship_by_date":     order.ShipByDate,
		"buyer_message":    order.BuyerMessage,
		"created_at":       order.CreatedAt,
		"updated_at":       order.UpdatedAt,
		"items":            orderItems,
	}
}

func (h *OrderManagerHandler) convertTiktokOrderToDetail(order *models.TiktokOrder, items []models.TiktokOrderItem) map[string]interface{} {
	orderItems := make([]map[string]interface{}, len(items))
	for i, item := range items {
		orderItems[i] = map[string]interface{}{
			"item_id":   item.ProductID,
			"item_name": item.ProductName,
			"item_sku":  item.SellerSku,
			"quantity":  item.Quantity,
			"price":     item.Price,
			"total":     float64(*item.Quantity) * *item.Price,
		}
	}

	return map[string]interface{}{
		"id":               order.ID,
		"order_sn":         order.OrderSN,
		"order_no":         order.OrderSN,
		"order_status":     order.OrderStatus,
		"status":           order.OrderStatus,
		"platform":         "tiktok",
		"category":         "",
		"buyer_username":   order.BuyerUsername,
		"total_amount":     order.TotalAmount,
		"currency":         order.Currency,
		"payment_method":   order.PaymentMethod,
		"shipping_carrier": order.ShippingCarrier,
		"tracking_number":  order.TrackingNumber,
		"ship_by_date":     order.ShipByDate,
		"buyer_message":    order.BuyerMessage,
		"created_at":       order.CreatedAt,
		"updated_at":       order.UpdatedAt,
		"items":            orderItems,
	}
}
