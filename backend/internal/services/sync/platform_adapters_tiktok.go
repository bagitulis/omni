package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/services/platform"
)

// TiktokOrderManager adapts TikTok API client to OrderManager interface
type TiktokOrderManager struct {
	client   *platform.TiktokAPIClient
	tenantID string
}

// NewTiktokOrderManager creates a TikTok order manager
func NewTiktokOrderManager(client *platform.TiktokAPIClient, tenantID string) *TiktokOrderManager {
	return &TiktokOrderManager{
		client:   client,
		tenantID: tenantID,
	}
}

// GetOrderList fetches orders from TikTok
func (m *TiktokOrderManager) GetOrderList(ctx context.Context, status string, days int) ([]Order, error) {
	if m.client == nil {
		adapterLogger.WithTenantID(m.tenantID).Warn("TikTok client is nil - cannot fetch orders")
		return nil, fmt.Errorf("tiktok client not configured")
	}
	if !m.client.IsInitialized() {
		adapterLogger.WithTenantID(m.tenantID).Warn("TikTok client not initialized - missing credentials or tokens")
		return nil, fmt.Errorf("tiktok client not initialized - check OAuth configuration")
	}

	adapterLogger.WithTenantID(m.tenantID).Info("Calling TikTok API: GetOrderList")
	rawOrders, err := m.client.GetOrderList(ctx, status, days)
	if err != nil {
		adapterLogger.WithTenantID(m.tenantID).Error("TikTok API error: " + err.Error())
		return nil, err
	}

	adapterLogger.WithFields(map[string]interface{}{
		"count":  len(rawOrders),
		"status": status,
	}).Info("TikTok API response received")

	orders := make([]Order, 0, len(rawOrders))
	for _, raw := range rawOrders {
		// TikTok uses 'id' field, which tiktok_client maps to 'order_sn'
		orderSN := getString(raw, "order_sn")
		if orderSN == "" {
			orderSN = getString(raw, "id")
		}
		order := Order{
			OrderSN:  orderSN,
			OrderNo:  orderSN, // Alias for frontend compatibility
			Platform: strings.ToUpper("tiktok"),
			Status:   status,
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderDetails fetches order details from TikTok
func (m *TiktokOrderManager) GetOrderDetails(ctx context.Context, orderIDs []string) ([]Order, error) {
	if m.client == nil {
		return nil, fmt.Errorf("tiktok client not configured")
	}
	if !m.client.IsInitialized() {
		return nil, fmt.Errorf("tiktok client not initialized")
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
			Platform:      strings.ToUpper("tiktok"),
			Status:        getString(raw, "order_status"),
			TotalAmount:   getFloat64(raw, "payment_info.total_amount"),
			BuyerUsername: getString(raw, "buyer_message"),
		}
		orders = append(orders, order)
	}

	return orders, nil
}

// GetOrderItems fetches order items from TikTok
// TikTok includes line_items in order details response
func (m *TiktokOrderManager) GetOrderItems(ctx context.Context, orderIDs []string) (map[string][]OrderItem, error) {
	// Get order details which include line_items
	rawOrders, err := m.client.GetOrderDetails(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]OrderItem)
	for _, raw := range rawOrders {
		// TikTok uses 'id' for order identifier
		orderID := getString(raw, "id")
		if orderID == "" {
			orderID = getString(raw, "order_id")
		}

		// TikTok uses line_items instead of item_list
		if lineItems, ok := raw["line_items"].([]interface{}); ok {
			orderItems := make([]OrderItem, 0, len(lineItems))
			for _, item := range lineItems {
				if itemMap, ok := item.(map[string]interface{}); ok {
					// TikTok uses seller_sku or sku_id
					sku := getString(itemMap, "seller_sku")
					if sku == "" {
						sku = getString(itemMap, "sku_id")
					}

					// TikTok API doesn't always return quantity field
					// Try multiple possible keys, default to 1 if not found
					qty := getInt(itemMap, "quantity")
					if qty == 0 {
						qty = getInt(itemMap, "item_count")
					}
					if qty == 0 {
						qty = getInt(itemMap, "qty")
					}
					if qty == 0 {
						// Default to 1 - most TikTok orders have qty=1
						qty = 1
					}

					orderItems = append(orderItems, OrderItem{
						OrderID:       orderID,
						SKU:           sku,
						ProductName:   getString(itemMap, "product_name"),
						VariationName: getString(itemMap, "sku_name"),
						Quantity:      qty,
						Price:         getFloat64(itemMap, "original_price"),
					})
				}
			}
			result[orderID] = orderItems
		}
	}

	return result, nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

// getMapKeys returns all keys in a map (for debugging)
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
