package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// saveLazadaOrders saves Lazada orders using models package
func (r *GormOrderRepository) saveLazadaOrders(ctx context.Context, db *gorm.DB, orders []Order) error {
	for _, order := range orders {
		model := models.LazadaOrder{
			TenantID:    r.tenantID,
			OrderSN:     order.OrderSN,
			OrderStatus: order.Status,
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
			if err := r.saveLazadaOrderItems(ctx, db, order.OrderSN, order.Items); err != nil {
				return err
			}
		}
	}
	return nil
}

// saveLazadaOrderItems saves Lazada order items
func (r *GormOrderRepository) saveLazadaOrderItems(ctx context.Context, db *gorm.DB, orderSN string, items []OrderItem) error {
	// Delete existing items first
	if err := db.WithContext(ctx).Where("order_sn = ?", orderSN).Delete(&models.LazadaOrderItem{}).Error; err != nil {
		return err
	}

	// Insert new items
	for _, item := range items {
		qty := item.Quantity
		price := item.Price
		itemModel := models.LazadaOrderItem{
			TenantID:    r.tenantID,
			OrderSN:     orderSN,
			SellerSku:   item.SKU,
			ProductName: item.ProductName,
			Quantity:    &qty,
			Price:       &price,
		}
		if err := db.WithContext(ctx).Create(&itemModel).Error; err != nil {
			return err
		}
	}
	return nil
}

// getLazadaOrdersByStatus gets Lazada orders by status WITH ITEMS
// Returns flattened items format for frontend (same as Node.js OrderFormatterService)
func (r *GormOrderRepository) getLazadaOrdersByStatus(ctx context.Context, db *gorm.DB, status string, limit int) ([]Order, error) {
	var orderModels []models.LazadaOrder
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
	var items []models.LazadaOrderItem
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).Where("order_sn IN ?", orderSNs).Find(&items).Error; err != nil {
			fmt.Printf("Warning: failed to fetch lazada items: %v\n", err)
		}
	}

	// Group items by order_sn
	itemsByOrder := make(map[string][]models.LazadaOrderItem)
	for _, item := range items {
		itemsByOrder[item.OrderSN] = append(itemsByOrder[item.OrderSN], item)
	}

	// Build flattened orders with items
	return r.flattenLazadaOrders(orderModels, itemsByOrder), nil
}

// flattenLazadaOrders creates flattened order list for frontend
func (r *GormOrderRepository) flattenLazadaOrders(orderModels []models.LazadaOrder, itemsByOrder map[string][]models.LazadaOrderItem) []Order {
	orders := make([]Order, 0)

	for _, m := range orderModels {
		orderItems := itemsByOrder[m.OrderSN]

		if len(orderItems) == 0 {
			orders = append(orders, Order{
				ID:        fmt.Sprintf("%d", m.ID),
				OrderSN:   m.OrderSN,
				OrderNo:   m.OrderSN,
				Platform:  strings.ToUpper("lazada"),
				Status:    m.OrderStatus,
				CreatedAt: m.CreatedAt,
				UpdatedAt: m.UpdatedAt,
			})
		} else {
			for _, item := range orderItems {
				qty := 0
				if item.Quantity != nil {
					qty = *item.Quantity
				}

				orders = append(orders, Order{
					ID:          fmt.Sprintf("%d", m.ID),
					OrderSN:     m.OrderSN,
					OrderNo:     m.OrderSN,
					Platform:    strings.ToUpper("lazada"),
					Status:      m.OrderStatus,
					SKU:         item.SellerSku,
					ProductName: item.ProductName,
					Quantity:    qty,
					CreatedAt:   m.CreatedAt,
					UpdatedAt:   m.UpdatedAt,
				})
			}
		}
	}

	return orders
}
