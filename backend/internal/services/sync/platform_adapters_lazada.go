package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/services/platform"
)

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
		// lazada_client_api.go maps order_id to order_sn at line 72
		orderSN := getString(raw, "order_sn")
		if orderSN == "" {
			// Fallback to order_id if order_sn not present
			orderSN = getString(raw, "order_id")
		}
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

	// Lazada doesn't have a batch GetOrderDetails for headers only easily.
	// But we can use the data from GetOrderList or fetch each.
	// For now, let's assume we need to return basic Order structs.
	orders := make([]Order, 0, len(orderIDs))
	for _, id := range orderIDs {
		orders = append(orders, Order{
			OrderSN:  id,
			OrderNo:  id,
			Platform: strings.ToUpper("lazada"),
		})
	}

	return orders, nil
}

// GetOrderItems fetches order items from Lazada
// Uses /order/items/get endpoint which returns items with sku, name, variation, tracking fields
func (m *LazadaOrderManager) GetOrderItems(ctx context.Context, orderIDs []string) (map[string][]OrderItem, error) {
	if m.client == nil {
		return nil, fmt.Errorf("lazada client not configured")
	}

	// Use GetOrderDetails from client which calls /order/items/get
	rawItems, err := m.client.GetOrderDetails(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]OrderItem)
	for _, raw := range rawItems {
		orderID := getString(raw, "order_id")

		// Lazada field mappings (from Node.js reference):
		// - sku -> seller_sku (SKU)
		// - name -> product_name
		// - variation -> variation_name
		// - tracking_code -> tracking_number
		// - shipment_provider -> shipping_carrier (needs parsing)
		// - paid_price -> price
		trackingCode := getString(raw, "tracking_code")
		shippingCarrier := cleanLazadaShippingCarrier(getString(raw, "shipment_provider"))

		item := OrderItem{
			OrderID:         orderID,
			SKU:             getString(raw, "sku"),       // Lazada uses 'sku' field
			ProductName:     getString(raw, "name"),      // Lazada uses 'name' field
			VariationName:   getString(raw, "variation"), // Lazada uses 'variation' field
			Quantity:        1,                           // Lazada items are separate rows
			Price:           getFloat64(raw, "paid_price"),
			TrackingNumber:  trackingCode,
			ShippingCarrier: shippingCarrier,
		}

		result[orderID] = append(result[orderID], item)
	}

	adapterLogger.WithTenantID(m.tenantID).WithFields(map[string]interface{}{
		"order_count": len(orderIDs),
		"item_count":  len(rawItems),
	}).Info("Lazada GetOrderItems completed")

	return result, nil
}

// cleanLazadaShippingCarrier extracts clean carrier name from Lazada format
// Input: "Pickup: LEX ID, Delivery: LEX ID" -> Output: "LEX ID"
func cleanLazadaShippingCarrier(provider string) string {
	if provider == "" {
		return ""
	}

	// Check for "Pickup: XXX" or "Delivery: XXX" format
	if strings.Contains(provider, ":") {
		// Extract the part after colon, before comma
		parts := strings.SplitN(provider, ":", 2)
		if len(parts) > 1 {
			carrier := strings.TrimSpace(parts[1])
			// Remove everything after comma if present
			if idx := strings.Index(carrier, ","); idx > 0 {
				carrier = strings.TrimSpace(carrier[:idx])
			}
			return carrier
		}
	}

	return provider
}
