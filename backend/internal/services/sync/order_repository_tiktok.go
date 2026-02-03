package sync

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// saveTiktokOrders saves TikTok orders using models package (upsert pattern)
func (r *GormOrderRepository) saveTiktokOrders(ctx context.Context, db *gorm.DB, orders []Order) error {
	for _, order := range orders {
		var totalAmount *float64
		if order.TotalAmount > 0 {
			totalAmount = &order.TotalAmount
		}

		// First, try to find existing record
		var existing models.TiktokOrder
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

			if err := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			// Record doesn't exist - CREATE it
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
			if err := db.WithContext(ctx).Create(&model).Error; err != nil {
				return err
			}
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

		// Fix: If product_id is 0, try to find it from tiktok_skus using SKU (SellerSku or SkuID)
		productID := item.ItemID
		if productID == 0 && item.SKU != "" {
			var skuModel models.TiktokSku
			// Try matching SKU against seller_sku or sku_id
			if err := db.WithContext(ctx).Where("seller_sku = ? OR sku_id = ?", item.SKU, item.SKU).First(&skuModel).Error; err == nil {
				productID = int64(skuModel.ProductID)
			}
		}

		itemModel := models.TiktokOrderItem{
			TenantID:      r.tenantID,
			OrderSN:       orderSN,
			ProductID:     productID, // ItemID contains product_id from TikTok API (or recovered from DB)
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

	// Collect unique product_ids for image lookup
	productIDSet := make(map[int64]bool)
	for _, item := range items {
		if item.ProductID > 0 {
			productIDSet[item.ProductID] = true
		}
	}
	productIDs := make([]int64, 0, len(productIDSet))
	for id := range productIDSet {
		productIDs = append(productIDs, id)
	}

	// Fetch product images from tiktok_products cache
	imageMap := r.getTiktokProductImagesFromCache(ctx, db, productIDs)

	// Group items by order_sn and enrich with images
	itemsByOrder := make(map[string][]models.TiktokOrderItem)
	for i := range items {
		// Enrich item with cached product image if empty
		if items[i].ProductImage == "" && items[i].ProductID > 0 {
			if imgURL, ok := imageMap[items[i].ProductID]; ok {
				items[i].ProductImage = imgURL
			}
		}
		itemsByOrder[items[i].OrderSN] = append(itemsByOrder[items[i].OrderSN], items[i])
	}

	// Build flattened orders with items and deduplication
	return r.flattenTiktokOrders(orderModels, itemsByOrder), nil
}

// getTiktokProductImagesFromCache fetches product images from tiktok_products table
// NOTE: productIDs are INTERNAL IDs (tiktok_products.id), not TikTok API product_id strings
func (r *GormOrderRepository) getTiktokProductImagesFromCache(ctx context.Context, db *gorm.DB, productIDs []int64) map[int64]string {
	result := make(map[int64]string)
	if len(productIDs) == 0 {
		return result
	}

	// Use temporary struct to handle LocalImages as JSONArray
	// Query by internal ID (tiktok_products.id), not product_id (TikTok API string)
	type ProductImage struct {
		ID          int64            `gorm:"column:id"`
		Image       string           `gorm:"column:image"`
		LocalImages models.JSONArray `gorm:"column:local_images"`
	}

	var products []ProductImage
	err := db.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Select("id, image, local_images").
		Where("id IN ?", productIDs).
		Find(&products).Error

	if err != nil {
		fmt.Printf("Warning: failed to fetch tiktok product images: %v\n", err)
		return result
	}

	for _, p := range products {
		imgURL := ""

		// Try local_images first (JSONArray is []interface{})
		if len(p.LocalImages) > 0 {
			if path, ok := p.LocalImages[0].(string); ok && path != "" {
				imgURL = path
			}
		}

		// Fallback to image field
		if imgURL == "" && p.Image != "" {
			imgURL = p.Image
		}

		if imgURL != "" {
			result[p.ID] = imgURL
		}
	}
	return result
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

		// Format countdown from ship_by_date
		countdown := ""
		shipByDate := int64(0)
		if m.ShipByDate != nil && *m.ShipByDate > 0 {
			shipByDate = *m.ShipByDate
			countdown = formatShipByDateFromDB(shipByDate)
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
				ShipByDate:      shipByDate,
				Countdown:       countdown,
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
					ProductImage:    item.ProductImage,
					ShipByDate:      shipByDate,
					Countdown:       countdown,
					CreatedAt:       m.CreatedAt,
					UpdatedAt:       m.UpdatedAt,
				})
			}
		}
	}

	return orders
}
