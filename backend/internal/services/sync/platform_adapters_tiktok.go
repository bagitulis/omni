package sync

import (
	"context"
	"fmt"
	"strconv"
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
// For AWAITING_COLLECTION status, also fetches order details for tracking info
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

		// Extract payment info from nested payment object
		totalAmount := float64(0)
		currency := "IDR"
		if payment, ok := raw["payment"].(map[string]interface{}); ok {
			if ta := getString(payment, "total_amount"); ta != "" {
				totalAmount, _ = strconv.ParseFloat(ta, 64)
			}
			if cur := getString(payment, "currency"); cur != "" {
				currency = cur
			}
		}

		// Extract buyer info
		buyerUsername := getString(raw, "buyer_nickname")
		if buyerUsername == "" {
			buyerUsername = getString(raw, "user_id")
		}

		// Extract payment method
		paymentMethod := getString(raw, "payment_method_name")
		if paymentMethod == "" {
			if isCod, ok := raw["is_cod"].(bool); ok && isCod {
				paymentMethod = "COD"
			}
		}

		// Extract shipping info
		shippingCarrier := getString(raw, "shipping_provider")
		if shippingCarrier == "" {
			shippingCarrier = getString(raw, "shipping_provider_name")
		}
		shippingType := getString(raw, "shipping_type")

		order := Order{
			OrderSN:         orderSN,
			OrderNo:         orderSN, // Alias for frontend compatibility
			Platform:        strings.ToUpper("tiktok"),
			Status:          status,
			TotalAmount:     totalAmount,
			Currency:        currency,
			BuyerUsername:   buyerUsername,
			PaymentMethod:   paymentMethod,
			ShippingCarrier: shippingCarrier,
			ShippingType:    shippingType,
			BuyerMessage:    getString(raw, "buyer_message"),
		}

		// For AWAITING_COLLECTION (processed), extract tracking info
		// TikTok API returns tracking in multiple places - try all of them
		if status == "AWAITING_COLLECTION" {
			// Method 1: Order-level tracking
			order.TrackingNumber = getString(raw, "tracking_number")
			if order.ShippingCarrier == "" {
				order.ShippingCarrier = getString(raw, "shipping_provider_name")
			}
			if order.ShippingCarrier == "" {
				order.ShippingCarrier = getString(raw, "shipping_provider")
			}

			// Method 2: From packages array (TikTok may return tracking in packages)
			if order.TrackingNumber == "" {
				if packages, ok := raw["packages"].([]interface{}); ok && len(packages) > 0 {
					if pkg, ok := packages[0].(map[string]interface{}); ok {
						if trackingNo := getString(pkg, "tracking_number"); trackingNo != "" {
							order.TrackingNumber = trackingNo
						}
						if carrier := getString(pkg, "shipping_provider_name"); carrier != "" {
							order.ShippingCarrier = carrier
						} else if carrier := getString(pkg, "shipping_provider"); carrier != "" {
							order.ShippingCarrier = carrier
						}
					}
				}
			}

			// Method 3: From line_items (each item may have its own tracking)
			if order.TrackingNumber == "" {
				if lineItems, ok := raw["line_items"].([]interface{}); ok && len(lineItems) > 0 {
					if item, ok := lineItems[0].(map[string]interface{}); ok {
						if trackingNo := getString(item, "tracking_number"); trackingNo != "" {
							order.TrackingNumber = trackingNo
						}
						if carrier := getString(item, "shipping_provider_name"); carrier != "" {
							order.ShippingCarrier = carrier
						}
					}
				}
			}
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
		orderSN := getString(raw, "id")
		if orderSN == "" {
			orderSN = getString(raw, "order_id")
		}

		order := Order{
			OrderSN:       orderSN,
			OrderNo:       orderSN,
			Platform:      strings.ToUpper("tiktok"),
			Status:        getString(raw, "order_status"),
			TotalAmount:   getFloat64(raw, "payment_info.total_amount"),
			BuyerUsername: getString(raw, "buyer_message"),
		}

		// Extract tracking info from order details
		order.TrackingNumber = getString(raw, "tracking_number")
		order.ShippingCarrier = getString(raw, "shipping_provider_name")
		if order.ShippingCarrier == "" {
			order.ShippingCarrier = getString(raw, "shipping_provider")
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

					// Extract product_id for image lookup from product cache
					productID := getInt64(itemMap, "product_id")

					// Extract product image URL if available from API
					productImage := getString(itemMap, "product_image")
					if productImage == "" {
						// TikTok may return image in sku_image or image field
						productImage = getString(itemMap, "sku_image")
					}
					if productImage == "" {
						productImage = getString(itemMap, "image")
					}
					// Try nested image object
					if productImage == "" {
						if imgObj, ok := itemMap["image"].(map[string]interface{}); ok {
							productImage = getString(imgObj, "url")
							if productImage == "" {
								productImage = getString(imgObj, "thumb_url")
							}
						}
					}

					orderItems = append(orderItems, OrderItem{
						OrderID:       orderID,
						ItemID:        productID, // Use ItemID field for product_id
						SKU:           sku,
						ProductName:   getString(itemMap, "product_name"),
						VariationName: getString(itemMap, "sku_name"),
						Quantity:      qty,
						Price:         getFloat64(itemMap, "original_price"),
						ProductImage:  productImage,
					})
				}
			}
			result[orderID] = orderItems
		}
	}

	return result, nil
}
