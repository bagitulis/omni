package handlers

import "github.com/omni/backend/internal/models"

// convertShopeeOrderToDetail converts a Shopee order to frontend display format
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

	return buildOrderDetailMap(order.ID, order.OrderSN, order.OrderStatus, "shopee",
		order.BuyerUsername, order.TotalAmount, order.Currency, order.PaymentMethod,
		order.ShippingCarrier, order.TrackingNumber, order.ShipByDate, order.BuyerMessage,
		order.CreatedAt, order.UpdatedAt, orderItems)
}

// convertLazadaOrderToDetail converts a Lazada order to frontend display format
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

	return buildOrderDetailMap(order.ID, order.OrderSN, order.OrderStatus, "lazada",
		order.BuyerUsername, order.TotalAmount, order.Currency, order.PaymentMethod,
		order.ShippingCarrier, order.TrackingNumber, order.ShipByDate, order.BuyerMessage,
		order.CreatedAt, order.UpdatedAt, orderItems)
}

// convertTiktokOrderToDetail converts a TikTok order to frontend display format
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

	return buildOrderDetailMap(order.ID, order.OrderSN, order.OrderStatus, "tiktok",
		order.BuyerUsername, order.TotalAmount, order.Currency, order.PaymentMethod,
		order.ShippingCarrier, order.TrackingNumber, order.ShipByDate, order.BuyerMessage,
		order.CreatedAt, order.UpdatedAt, orderItems)
}

// buildOrderDetailMap builds the common order detail response map (DRY)
func buildOrderDetailMap(
	id interface{}, orderSN, orderStatus, platform string,
	buyerUsername, totalAmount, currency, paymentMethod interface{},
	shippingCarrier, trackingNumber, shipByDate, buyerMessage interface{},
	createdAt, updatedAt interface{}, items []map[string]interface{},
) map[string]interface{} {
	return map[string]interface{}{
		"id":               id,
		"order_sn":         orderSN,
		"order_no":         orderSN,
		"order_status":     orderStatus,
		"status":           orderStatus,
		"platform":         platform,
		"category":         "",
		"buyer_username":   buyerUsername,
		"total_amount":     totalAmount,
		"currency":         currency,
		"payment_method":   paymentMethod,
		"shipping_carrier": shippingCarrier,
		"tracking_number":  trackingNumber,
		"ship_by_date":     shipByDate,
		"buyer_message":    buyerMessage,
		"created_at":       createdAt,
		"updated_at":       updatedAt,
		"items":            items,
	}
}
