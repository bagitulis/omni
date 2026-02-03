package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

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
			OrderSN:       orderSN,
			OrderNo:       orderSN, // Alias for frontend compatibility
			Platform:      strings.ToUpper("lazada"),
			Status:        status,
			TotalAmount:   getFloat64(raw, "price"),
			BuyerUsername: getString(raw, "customer_first_name"),
			Currency:      "IDR", // Lazada default
		}

		// Extract promised_shipping_times and convert to ship_by_date Unix timestamp
		if promisedShip := getString(raw, "promised_shipping_times"); promisedShip != "" {
			// Lazada format: "2024-01-15" (ISO date) - convert to Unix timestamp
			if timestamp := parseLazadaDateToUnix(promisedShip); timestamp > 0 {
				order.ShipByDate = timestamp
			}
		}

		// Extract shipping info from delivery_info
		if deliveryInfo := getString(raw, "delivery_info"); deliveryInfo != "" {
			order.ShippingCarrier = deliveryInfo
		}

		orders = append(orders, order)
	}

	// Enrich orders with items (to get detailed info and images)
	if len(orders) > 0 {
		orders = m.enrichLazadaOrdersWithItems(ctx, orders)
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
		// - item_id -> item ID for product cache lookup
		// - product_main_image -> product image URL
		trackingCode := getString(raw, "tracking_code")
		shippingCarrier := cleanLazadaShippingCarrier(getString(raw, "shipment_provider"))

		// Extract item_id for product cache lookup
		itemID := getInt64(raw, "item_id")
		if itemID == 0 {
			itemID = getInt64(raw, "id")
		}

		// Extract product image URL
		productImage := getString(raw, "product_main_image")
		if productImage == "" {
			productImage = getString(raw, "image")
		}
		if productImage == "" {
			// Try product_detail_url as fallback (though not ideal)
			if imgs, ok := raw["images"].([]interface{}); ok && len(imgs) > 0 {
				if imgStr, ok := imgs[0].(string); ok {
					productImage = imgStr
				}
			}
		}

		item := OrderItem{
			OrderID:         orderID,
			ItemID:          itemID,                      // For product cache lookup
			SKU:             getString(raw, "sku"),       // Lazada uses 'sku' field
			ProductName:     getString(raw, "name"),      // Lazada uses 'name' field
			VariationName:   getString(raw, "variation"), // Lazada uses 'variation' field
			Quantity:        1,                           // Lazada items are separate rows
			Price:           getFloat64(raw, "paid_price"),
			TrackingNumber:  trackingCode,
			ShippingCarrier: shippingCarrier,
			ProductImage:    productImage,
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

// enrichLazadaOrdersWithItems fetches order items and enriches orders
func (m *LazadaOrderManager) enrichLazadaOrdersWithItems(ctx context.Context, orders []Order) []Order {
	orderIDs := make([]string, len(orders))
	for i, o := range orders {
		orderIDs[i] = o.OrderSN
	}

	// Fetch items for all orders
	itemsMap, err := m.GetOrderItems(ctx, orderIDs)
	if err != nil {
		adapterLogger.Warn("Failed to fetch Lazada order items: " + err.Error())
		return orders
	}

	// Merge items into orders
	for i := range orders {
		if items, ok := itemsMap[orders[i].OrderSN]; ok && len(items) > 0 {
			orders[i].Items = items

			// Extract shipping carrier from first item (all items in order have same carrier)
			if orders[i].ShippingCarrier == "" && items[0].ShippingCarrier != "" {
				orders[i].ShippingCarrier = items[0].ShippingCarrier
			}
		}
	}

	return orders
}

// parseLazadaDateToUnix converts Lazada ISO date to Unix timestamp
// Input: "2024-01-15" or "2024-01-15 10:30:00" -> Output: Unix timestamp
func parseLazadaDateToUnix(dateStr string) int64 {
	if dateStr == "" {
		return 0
	}

	// Try parsing common formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t.Unix()
		}
	}

	return 0
}
