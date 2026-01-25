package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/utils/logger"
)

var adapterLogger = logger.Named("PlatformAdapters")

// ShopeeOrderManager adapts Shopee API client to OrderManager interface
type ShopeeOrderManager struct {
	client   *platform.ShopeeAPIClient
	tenantID string
}

// NewShopeeOrderManager creates a Shopee order manager
func NewShopeeOrderManager(client *platform.ShopeeAPIClient, tenantID string) *ShopeeOrderManager {
	return &ShopeeOrderManager{
		client:   client,
		tenantID: tenantID,
	}
}

// GetOrderList fetches orders from Shopee
func (m *ShopeeOrderManager) GetOrderList(ctx context.Context, status string, days int) ([]Order, error) {
	if m.client == nil {
		adapterLogger.WithTenantID(m.tenantID).Warn("Shopee client is nil - cannot fetch orders")
		return nil, fmt.Errorf("shopee client not configured")
	}
	if !m.client.IsInitialized() {
		adapterLogger.WithTenantID(m.tenantID).Warn("Shopee client not initialized - missing credentials or tokens")
		return nil, fmt.Errorf("shopee client not initialized - check OAuth configuration")
	}

	adapterLogger.WithTenantID(m.tenantID).Info("Calling Shopee API: GetOrderList")
	rawOrders, err := m.client.GetOrderList(ctx, status, days)
	if err != nil {
		adapterLogger.WithTenantID(m.tenantID).Error("Shopee API error: " + err.Error())
		return nil, err
	}

	adapterLogger.WithFields(map[string]interface{}{
		"count":  len(rawOrders),
		"status": status,
	}).Info("Shopee API response received")

	orders := make([]Order, 0, len(rawOrders))
	for _, raw := range rawOrders {
		orderSN := getString(raw, "order_sn")
		order := Order{
			OrderSN:  orderSN,
			OrderNo:  orderSN, // Alias for frontend compatibility
			Platform: strings.ToUpper("shopee"),
			Status:   status,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderDetails fetches order details from Shopee
func (m *ShopeeOrderManager) GetOrderDetails(ctx context.Context, orderIDs []string) ([]Order, error) {
	if m.client == nil {
		return nil, fmt.Errorf("shopee client not configured")
	}
	if !m.client.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	rawOrders, err := m.client.GetOrderDetails(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	orders := make([]Order, 0, len(rawOrders))
	for _, raw := range rawOrders {
		orderSN := getString(raw, "order_sn")
		order := Order{
			OrderSN:       orderSN,
			OrderNo:       orderSN, // Alias for frontend compatibility
			Platform:      strings.ToUpper("shopee"),
			Status:        getString(raw, "status"),
			TotalAmount:   getFloat64(raw, "total_amount"),
			BuyerUsername: getString(raw, "buyer_username"),
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderItems fetches order items from Shopee
func (m *ShopeeOrderManager) GetOrderItems(ctx context.Context, orderIDs []string) (map[string][]OrderItem, error) {
	// Get order details which include items
	rawOrders, err := m.client.GetOrderDetails(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]OrderItem)
	for _, raw := range rawOrders {
		orderSN := getString(raw, "order_sn")
		if items, ok := raw["items"].([]interface{}); ok {
			orderItems := make([]OrderItem, 0, len(items))
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// Use model_sku first, fallback to item_sku
					sku := getString(itemMap, "model_sku")
					if sku == "" {
						sku = getString(itemMap, "item_sku")
					}

					orderItems = append(orderItems, OrderItem{
						OrderID:       orderSN,
						SKU:           sku,
						ProductName:   getString(itemMap, "item_name"),
						VariationName: getString(itemMap, "model_name"),
						Quantity:      getInt(itemMap, "quantity"),
						Price:         getFloat64(itemMap, "price"),
					})
				}
			}
			result[orderSN] = orderItems
		}
	}

	return result, nil
}

// LazadaOrderManager adapts Lazada API client to OrderManager interface
type LazadaOrderManager struct {
	client   *platform.LazadaAPIClient
	tenantID string
}

// NewLazadaOrderManager creates a Lazada order manager
func NewLazadaOrderManager(client *platform.LazadaAPIClient, tenantID string) *LazadaOrderManager {
	return &LazadaOrderManager{
		client:   client,
		tenantID: tenantID,
	}
}

// GetOrderList fetches orders from Lazada
func (m *LazadaOrderManager) GetOrderList(ctx context.Context, status string, days int) ([]Order, error) {
	if m.client == nil {
		adapterLogger.WithTenantID(m.tenantID).Warn("Lazada client is nil - cannot fetch orders")
		return nil, fmt.Errorf("lazada client not configured")
	}
	if !m.client.IsInitialized() {
		adapterLogger.WithTenantID(m.tenantID).Warn("Lazada client not initialized - missing credentials or tokens")
		return nil, fmt.Errorf("lazada client not initialized - check OAuth configuration")
	}

	adapterLogger.WithTenantID(m.tenantID).Info("Calling Lazada API: GetOrderList")
	rawOrders, err := m.client.GetOrderList(ctx, status, days)
	if err != nil {
		adapterLogger.WithTenantID(m.tenantID).Error("Lazada API error: " + err.Error())
		return nil, err
	}

	adapterLogger.WithFields(map[string]interface{}{
		"count":  len(rawOrders),
		"status": status,
	}).Info("Lazada API response received")

	orders := make([]Order, 0, len(rawOrders))
	for _, raw := range rawOrders {
		orderSN := getString(raw, "order_id")
		order := Order{
			OrderSN:  orderSN,
			OrderNo:  orderSN, // Alias for frontend compatibility
			Platform: strings.ToUpper("lazada"),
			Status:   status,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderDetails fetches order details from Lazada
func (m *LazadaOrderManager) GetOrderDetails(ctx context.Context, orderIDs []string) ([]Order, error) {
	if m.client == nil {
		return nil, fmt.Errorf("lazada client not configured")
	}
	if !m.client.IsInitialized() {
		return nil, fmt.Errorf("lazada client not initialized")
	}

	rawOrders, err := m.client.GetOrderDetails(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	orders := make([]Order, 0, len(rawOrders))
	for _, raw := range rawOrders {
		orderSN := getString(raw, "order_id")
		order := Order{
			OrderSN:       orderSN,
			OrderNo:       orderSN, // Alias for frontend compatibility
			Platform:      strings.ToUpper("lazada"),
			Status:        getString(raw, "statuses"),
			TotalAmount:   getFloat64(raw, "price"),
			BuyerUsername: getString(raw, "customer_name"),
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderItems fetches order items from Lazada
func (m *LazadaOrderManager) GetOrderItems(ctx context.Context, orderIDs []string) (map[string][]OrderItem, error) {
	return make(map[string][]OrderItem), nil
}
