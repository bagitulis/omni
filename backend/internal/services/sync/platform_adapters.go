package sync

import (
	"context"
	"fmt"
	"strings"
	"time"

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
			// Populate countdown from ship_by_date (Unix timestamp)
			if shipByDate := getInt64(detail, "ship_by_date"); shipByDate > 0 {
				orders[i].Countdown = formatShipByDate(shipByDate)
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

// formatShipByDate converts Unix timestamp to "Ship before DD/MM/YYYY" format
func formatShipByDate(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return fmt.Sprintf("Ship before %02d/%02d/%d", t.Day(), t.Month(), t.Year())
}
