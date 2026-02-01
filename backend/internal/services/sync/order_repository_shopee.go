package sync

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// saveShopeeOrders saves Shopee orders using models package (upsert pattern)
func (r *GormOrderRepository) saveShopeeOrders(ctx context.Context, db *gorm.DB, orders []Order) error {
	for _, order := range orders {
		var totalAmount *float64
		if order.TotalAmount > 0 {
			totalAmount = &order.TotalAmount
		}

		// Debug log
		fmt.Printf("[saveShopeeOrders] OrderSN=%s TotalAmount=%.2f Buyer=%s PaymentMethod=%s\n",
			order.OrderSN, order.TotalAmount, order.BuyerUsername, order.PaymentMethod)

		// First, try to find existing record
		var existing models.ShopeeOrder
		result := db.WithContext(ctx).Where("order_sn = ?", order.OrderSN).First(&existing)

		if result.Error == nil {
			// Record exists - UPDATE it with new data
			updates := map[string]interface{}{
				"order_status": order.Status,
			}
			// Only update non-empty fields - use actual value not pointer for GORM
			if order.TotalAmount > 0 {
				updates["total_amount"] = order.TotalAmount
			}
			if order.Currency != "" {
				updates["currency"] = order.Currency
			}
			if order.BuyerUsername != "" {
				updates["buyer_username"] = order.BuyerUsername
			}
			if order.PaymentMethod != "" {
				updates["payment_method"] = order.PaymentMethod
			}
			if order.ShippingCarrier != "" {
				updates["shipping_carrier"] = order.ShippingCarrier
			}
			if order.BuyerMessage != "" {
				updates["buyer_message"] = order.BuyerMessage
			}

			fmt.Printf("[saveShopeeOrders] Updating %s with updates=%+v\n", order.OrderSN, updates)

			if err := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			// Record doesn't exist - CREATE it
			model := models.ShopeeOrder{
				TenantID:        r.tenantID,
				OrderSN:         order.OrderSN,
				OrderStatus:     order.Status,
				TotalAmount:     totalAmount,
				Currency:        order.Currency,
				BuyerUsername:   order.BuyerUsername,
				PaymentMethod:   order.PaymentMethod,
				ShippingCarrier: order.ShippingCarrier,
				BuyerMessage:    order.BuyerMessage,
			}
			if err := db.WithContext(ctx).Create(&model).Error; err != nil {
				return err
			}
		}

		// Save order items if present
		if len(order.Items) > 0 {
			if err := r.saveShopeeOrderItems(ctx, db, order.OrderSN, order.Items); err != nil {
				return err
			}
		}
	}
	return nil
}

// saveShopeeOrderItems saves Shopee order items
func (r *GormOrderRepository) saveShopeeOrderItems(ctx context.Context, db *gorm.DB, orderSN string, items []OrderItem) error {
	// Delete existing items first
	if err := db.WithContext(ctx).Where("order_sn = ?", orderSN).Delete(&models.ShopeeOrderItem{}).Error; err != nil {
		return err
	}

	// Insert new items
	for _, item := range items {
		qty := item.Quantity
		price := item.Price
		itemModel := models.ShopeeOrderItem{
			TenantID:  r.tenantID,
			OrderSN:   orderSN,
			ItemName:  item.ProductName,
			ModelName: item.VariationName,
			ModelSku:  item.SKU,
			Quantity:  &qty,
			Price:     &price,
		}
		if err := db.WithContext(ctx).Create(&itemModel).Error; err != nil {
			return err
		}
	}
	return nil
}

// getShopeeOrdersByStatus gets Shopee orders by status WITH ITEMS
// Returns flattened items format for frontend (same as Node.js OrderFormatterService)
func (r *GormOrderRepository) getShopeeOrdersByStatus(ctx context.Context, db *gorm.DB, status string, limit int) ([]Order, error) {
	var orderModels []models.ShopeeOrder
	query := db.WithContext(ctx)

	if status != "" {
		query = query.Where("order_status = ?", status)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&orderModels).Error; err != nil {
		return nil, err
	}

	// Get order SNs for fetching items
	orderSNs := make([]string, 0, len(orderModels))
	for _, m := range orderModels {
		orderSNs = append(orderSNs, m.OrderSN)
	}

	// Fetch items for all orders
	var items []models.ShopeeOrderItem
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).Where("order_sn IN ?", orderSNs).Find(&items).Error; err != nil {
			fmt.Printf("Warning: failed to fetch shopee items: %v\n", err)
		}
	}

	// Group items by order_sn
	itemsByOrder := make(map[string][]models.ShopeeOrderItem)
	for _, item := range items {
		itemsByOrder[item.OrderSN] = append(itemsByOrder[item.OrderSN], item)
	}

	// Build flattened orders with items (same format as Node.js)
	return r.flattenShopeeOrders(orderModels, itemsByOrder), nil
}

// flattenShopeeOrders creates flattened order list for frontend
func (r *GormOrderRepository) flattenShopeeOrders(orderModels []models.ShopeeOrder, itemsByOrder map[string][]models.ShopeeOrderItem) []Order {
	orders := make([]Order, 0)

	for _, m := range orderModels {
		orderItems := itemsByOrder[m.OrderSN]

		// Get total amount
		totalAmount := float64(0)
		if m.TotalAmount != nil {
			totalAmount = *m.TotalAmount
		}

		// Default currency
		currency := m.Currency
		if currency == "" {
			currency = "IDR"
		}

		if len(orderItems) == 0 {
			// No items - still include order with empty item fields
			orders = append(orders, Order{
				ID:              fmt.Sprintf("%d", m.ID),
				OrderSN:         m.OrderSN,
				OrderNo:         m.OrderSN,
				Platform:        "SHOPEE",
				Status:          m.OrderStatus,
				TotalAmount:     totalAmount,
				Currency:        currency,
				BuyerUsername:   m.BuyerUsername,
				PaymentMethod:   m.PaymentMethod,
				ShippingCarrier: m.ShippingCarrier,
				BuyerMessage:    m.BuyerMessage,
				CreatedAt:       m.CreatedAt,
				UpdatedAt:       m.UpdatedAt,
			})
		} else {
			// Flatten items - each item becomes a separate "order" row for frontend
			for _, item := range orderItems {
				// Use model_sku first, fallback to item_sku (same as Node.js)
				sku := item.ModelSku
				if sku == "" {
					sku = item.ItemSku
				}

				qty := 0
				if item.Quantity != nil {
					qty = *item.Quantity
				}

				price := float64(0)
				if item.Price != nil {
					price = *item.Price
				}

				orders = append(orders, Order{
					ID:              fmt.Sprintf("%d", m.ID),
					OrderSN:         m.OrderSN,
					OrderNo:         m.OrderSN,
					Platform:        "SHOPEE",
					Status:          m.OrderStatus,
					TotalAmount:     totalAmount,
					Currency:        currency,
					BuyerUsername:   m.BuyerUsername,
					PaymentMethod:   m.PaymentMethod,
					ShippingCarrier: m.ShippingCarrier,
					BuyerMessage:    m.BuyerMessage,
					SKU:             sku,
					ProductName:     item.ItemName,
					VariationName:   item.ModelName,
					Quantity:        qty,
					Price:           price,
					CreatedAt:       m.CreatedAt,
					UpdatedAt:       m.UpdatedAt,
				})
			}
		}
	}

	return orders
}
