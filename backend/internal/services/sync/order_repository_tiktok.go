package sync

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// saveTiktokOrders saves TikTok orders using models package
func (r *GormOrderRepository) saveTiktokOrders(ctx context.Context, db *gorm.DB, orders []Order) error {
	for _, order := range orders {
		var totalAmount *float64
		if order.TotalAmount > 0 {
			totalAmount = &order.TotalAmount
		}

		model := models.TiktokOrder{
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

		result := db.WithContext(ctx).
			Where("order_sn = ?", order.OrderSN).
			Assign(model).
			FirstOrCreate(&model)
		if result.Error != nil {
			return result.Error
		}

		// Save order items if present
		if len(order.Items) > 0 {
			if err := r.saveTiktokOrderItems(ctx, db, order.OrderSN, order.Items); err != nil {
				return err
			}
		}
	}
	return nil
}

// saveTiktokOrderItems saves TikTok order items
func (r *GormOrderRepository) saveTiktokOrderItems(ctx context.Context, db *gorm.DB, orderSN string, items []OrderItem) error {
	// Delete existing items first
	if err := db.WithContext(ctx).Where("order_sn = ?", orderSN).Delete(&models.TiktokOrderItem{}).Error; err != nil {
		return err
	}

	// Insert new items
	for _, item := range items {
		qty := item.Quantity
		price := item.Price
		itemModel := models.TiktokOrderItem{
			TenantID:      r.tenantID,
			OrderSN:       orderSN,
			SellerSku:     item.SKU,
			ProductName:   item.ProductName,
			VariationName: item.VariationName,
			Quantity:      &qty,
			Price:         &price,
		}
		if err := db.WithContext(ctx).Create(&itemModel).Error; err != nil {
			return err
		}
	}
	return nil
}

// getTiktokOrdersByStatus gets TikTok orders by status WITH ITEMS
// Returns flattened items format for frontend (same as Node.js OrderFormatterService)
// TikTok has special logic: SellerSku for SKU, plus deduplication
func (r *GormOrderRepository) getTiktokOrdersByStatus(ctx context.Context, db *gorm.DB, status string, limit int) ([]Order, error) {
	var orderModels []models.TiktokOrder
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
	var items []models.TiktokOrderItem
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).Where("order_sn IN ?", orderSNs).Find(&items).Error; err != nil {
			fmt.Printf("Warning: failed to fetch tiktok items: %v\n", err)
		}
	}

	// Group items by order_sn
	itemsByOrder := make(map[string][]models.TiktokOrderItem)
	for _, item := range items {
		itemsByOrder[item.OrderSN] = append(itemsByOrder[item.OrderSN], item)
	}

	// Build flattened orders with items and deduplication
	return r.flattenTiktokOrders(orderModels, itemsByOrder), nil
}

// flattenTiktokOrders creates flattened order list for frontend with deduplication
func (r *GormOrderRepository) flattenTiktokOrders(orderModels []models.TiktokOrder, itemsByOrder map[string][]models.TiktokOrderItem) []Order {
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
			orders = append(orders, Order{
				ID:              fmt.Sprintf("%d", m.ID),
				OrderSN:         m.OrderSN,
				OrderNo:         m.OrderSN,
				Platform:        "TIKTOK",
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
			// Deduplicate by LineItemID
			seenLineItems := make(map[string]bool)
			for _, item := range orderItems {
				if item.LineItemID != "" && seenLineItems[item.LineItemID] {
					continue
				}
				if item.LineItemID != "" {
					seenLineItems[item.LineItemID] = true
				}

				qty := 0
				if item.Quantity != nil {
					qty = *item.Quantity
				}

				price := float64(0)
				if item.Price != nil {
					price = *item.Price
				}

				// TikTok uses SellerSku as the actual SKU
				orders = append(orders, Order{
					ID:              fmt.Sprintf("%d", m.ID),
					OrderSN:         m.OrderSN,
					OrderNo:         m.OrderSN,
					Platform:        "TIKTOK",
					Status:          m.OrderStatus,
					TotalAmount:     totalAmount,
					Currency:        currency,
					BuyerUsername:   m.BuyerUsername,
					PaymentMethod:   m.PaymentMethod,
					ShippingCarrier: m.ShippingCarrier,
					BuyerMessage:    m.BuyerMessage,
					SKU:             item.SellerSku,
					ProductName:     item.ProductName,
					VariationName:   item.VariationName,
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
