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
	client       *platform.ShopeeAPIClient
	tenantID     string
	imageService *ImageService
}

// NewShopeeOrderManager creates a Shopee order manager
func NewShopeeOrderManager(client *platform.ShopeeAPIClient, tenantID string, imageService *ImageService) *ShopeeOrderManager {
	return &ShopeeOrderManager{
		client:       client,
		tenantID:     tenantID,
		imageService: imageService,
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
				orders[i].ShipByDate = shipByDate
				orders[i].Countdown = formatShipByDate(shipByDate)
			}
			// Extract items with item_id for image fetching
			orders[i].Items = extractOrderItems(detail)
			mergedCount++
		}
	}

	// Fetch product images for order items
	if m.imageService != nil {
		m.enrichOrdersWithImages(ctx, orders)
	}

	adapterLogger.WithFields(map[string]interface{}{
		"merged_count": mergedCount,
		"total_orders": len(orders),
	}).Info("Enriched orders with details")

	return orders
}

// enrichOrdersWithImages fetches and populates product images for order items
func (m *ShopeeOrderManager) enrichOrdersWithImages(ctx context.Context, orders []Order) {
	itemIDs := collectUniqueItemIDs(orders)
	if len(itemIDs) == 0 {
		return
	}

	images, err := m.imageService.GetProductImages(ctx, itemIDs)
	if err != nil {
		adapterLogger.WithTenantID(m.tenantID).Warn("Failed to fetch product images: " + err.Error())
		return
	}

	// Populate images on order items
	for i := range orders {
		for j := range orders[i].Items {
			if url, ok := images[orders[i].Items[j].ItemID]; ok {
				orders[i].Items[j].ProductImage = url
			}
		}
		// Also set on flattened order fields if single item
		if len(orders[i].Items) == 1 {
			orders[i].ProductImage = orders[i].Items[0].ProductImage
		}
	}

	adapterLogger.WithFields(map[string]interface{}{
		"item_ids_requested": len(itemIDs),
		"images_found":       len(images),
	}).Info("Enriched orders with product images")
}

// collectUniqueItemIDs extracts unique item IDs from orders
func collectUniqueItemIDs(orders []Order) []int64 {
	seen := make(map[int64]bool)
	var ids []int64
	for _, o := range orders {
		for _, item := range o.Items {
			if item.ItemID > 0 && !seen[item.ItemID] {
				seen[item.ItemID] = true
				ids = append(ids, item.ItemID)
			}
		}
	}
	return ids
}

// extractOrderItems extracts items from order detail response
func extractOrderItems(detail map[string]interface{}) []OrderItem {
	items, ok := detail["items"].([]interface{})
	if !ok {
		return nil
	}

	orderItems := make([]OrderItem, 0, len(items))
	orderSN := getString(detail, "order_sn")

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		sku := getString(itemMap, "model_sku")
		if sku == "" {
			sku = getString(itemMap, "item_sku")
		}

		orderItems = append(orderItems, OrderItem{
			OrderID:       orderSN,
			ItemID:        getInt64(itemMap, "item_id"),
			SKU:           sku,
			ProductName:   getString(itemMap, "item_name"),
			VariationName: getString(itemMap, "model_name"),
			Quantity:      getInt(itemMap, "quantity"),
			Price:         getFloat64(itemMap, "price"),
		})
	}

	return orderItems
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
	rawOrders, err := m.client.GetOrderDetails(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]OrderItem)
	for _, raw := range rawOrders {
		orderSN := getString(raw, "order_sn")
		result[orderSN] = extractOrderItems(raw)
	}
	return result, nil
}

// formatShipByDate converts Unix timestamp to "Ship before DD/MM/YYYY" format
func formatShipByDate(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return fmt.Sprintf("Ship before %02d/%02d/%d", t.Day(), t.Month(), t.Year())
}
