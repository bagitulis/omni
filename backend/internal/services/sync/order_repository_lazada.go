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
			TenantID:      r.tenantID,
			OrderSN:       orderSN,
			ItemID:        item.ItemID, // ItemID for product cache lookup
			SellerSku:     item.SKU,
			ProductName:   item.ProductName,
			VariationName: item.VariationName,
			Quantity:      &qty,
			Price:         &price,
			ProductImage:  item.ProductImage,
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

	// Collect unique item_ids for image lookup
	itemIDSet := make(map[string]bool)
	for _, item := range items {
		if item.ItemID > 0 {
			itemIDSet[fmt.Sprintf("%d", item.ItemID)] = true
		}
	}
	itemIDs := make([]string, 0, len(itemIDSet))
	for id := range itemIDSet {
		itemIDs = append(itemIDs, id)
	}

	// Fetch product images from lazada_products cache
	imageMap := r.getLazadaProductImagesFromCache(ctx, db, itemIDs)

	// Group items by order_sn and enrich with images
	itemsByOrder := make(map[string][]models.LazadaOrderItem)
	for i := range items {
		// Enrich item with cached product image if empty
		if items[i].ProductImage == "" && items[i].ItemID > 0 {
			itemIDStr := fmt.Sprintf("%d", items[i].ItemID)
			if imgURL, ok := imageMap[itemIDStr]; ok {
				items[i].ProductImage = imgURL
			}
		}
		itemsByOrder[items[i].OrderSN] = append(itemsByOrder[items[i].OrderSN], items[i])
	}

	// Build flattened orders with items
	return r.flattenLazadaOrders(orderModels, itemsByOrder), nil
}

// getLazadaProductImagesFromCache fetches product images from lazada_products table
func (r *GormOrderRepository) getLazadaProductImagesFromCache(ctx context.Context, db *gorm.DB, itemIDs []string) map[string]string {
	result := make(map[string]string)
	if len(itemIDs) == 0 {
		return result
	}

	var products []models.LazadaProduct
	err := db.WithContext(ctx).
		Select("item_id, image").
		Where("item_id IN ? AND image != ''", itemIDs).
		Find(&products).Error

	if err != nil {
		fmt.Printf("Warning: failed to fetch lazada product images: %v\n", err)
		return result
	}

	for _, p := range products {
		if p.Image != "" {
			result[p.ItemID] = p.Image
		}
	}
	return result
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
					ID:            fmt.Sprintf("%d", m.ID),
					OrderSN:       m.OrderSN,
					OrderNo:       m.OrderSN,
					Platform:      strings.ToUpper("lazada"),
					Status:        m.OrderStatus,
					SKU:           item.SellerSku,
					ProductName:   item.ProductName,
					VariationName: item.VariationName,
					Quantity:      qty,
					ProductImage:  item.ProductImage,
					CreatedAt:     m.CreatedAt,
					UpdatedAt:     m.UpdatedAt,
				})
			}
		}
	}

	return orders
}
