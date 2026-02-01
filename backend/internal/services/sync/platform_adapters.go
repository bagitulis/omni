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
// For PROCESSED status, returns orders with tracking info from Package API
func (m *ShopeeOrderManager) GetOrderList(ctx context.Context, status string, days int) ([]Order, error) {
	if m.client == nil {
		adapterLogger.WithTenantID(m.tenantID).Warn("Shopee client is nil - cannot fetch orders")
		return nil, fmt.Errorf("shopee client not configured")
	}
	if !m.client.IsInitialized() {
		adapterLogger.WithTenantID(m.tenantID).Warn("Shopee client not initialized - missing credentials or tokens")
		return nil, fmt.Errorf("shopee client not initialized - check OAuth configuration")
	}

	adapterLogger.WithTenantID(m.tenantID).Info("Calling Shopee API: GetOrderList for status=" + status)
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
			OrderNo:  orderSN,
			Platform: strings.ToUpper("shopee"),
			Status:   status,
		}

		// For PROCESSED orders, extract tracking info from Package API response
		if status == "PROCESSED" {
			order.TrackingNumber = getString(raw, "tracking_number")
			order.ShippingCarrier = getString(raw, "shipping_carrier")
		}

		orders = append(orders, order)
	}

	// Enrich orders with details (total_amount, buyer_username, payment_method)
	if len(orders) > 0 {
		orders = m.enrichOrdersWithDetails(ctx, orders)
	}

	return orders, nil
}

// enrichOrdersWithDetails fetches and merges order details into orders
// Returns the enriched orders slice
func (m *ShopeeOrderManager) enrichOrdersWithDetails(ctx context.Context, orders []Order) []Order {
	orderSNs := make([]string, len(orders))
	for i, o := range orders {
		orderSNs[i] = o.OrderSN
	}

	// Batch fetch details (50 per request per Shopee API limit)
	detailMap := make(map[string]map[string]interface{})
	for i := 0; i < len(orderSNs); i += 50 {
		end := i + 50
		if end > len(orderSNs) {
			end = len(orderSNs)
		}
		batch := orderSNs[i:end]

		details, err := m.client.GetOrderDetails(ctx, batch)
		if err != nil {
			adapterLogger.Warn("Failed to fetch order details for enrichment: " + err.Error())
			continue
		}

		adapterLogger.WithFields(map[string]interface{}{
			"batch_size":   len(batch),
			"details_size": len(details),
		}).Info("Fetched order details for enrichment")

		// Debug: print first detail
		if len(details) > 0 {
			d := details[0]
			fmt.Printf("[enrichOrdersWithDetails] Sample detail: order_sn=%v, total_amount=%v, buyer_username=%v, payment_method=%v\n",
				d["order_sn"], d["total_amount"], d["buyer_username"], d["payment_method"])
		}

		for _, d := range details {
			if sn, ok := d["order_sn"].(string); ok {
				detailMap[sn] = d
			}
		}
	}

	// Merge details into orders - create new slice with modified values
	mergedCount := 0
	for i := range orders {
		if detail, ok := detailMap[orders[i].OrderSN]; ok {
			orders[i].TotalAmount = getFloat64(detail, "total_amount")
			orders[i].BuyerUsername = getString(detail, "buyer_username")
			orders[i].PaymentMethod = getString(detail, "payment_method")
			orders[i].Currency = getString(detail, "currency")
			if orders[i].Currency == "" {
				orders[i].Currency = "IDR"
			}
			mergedCount++
		}
	}

	fmt.Printf("[enrichOrdersWithDetails] Merged %d orders with details\n", mergedCount)

	return orders
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
			PaymentMethod: getString(raw, "payment_method"),
			Currency:      getString(raw, "currency"),
		}
		if order.Currency == "" {
			order.Currency = "IDR"
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
