package sync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	imageService "github.com/omni/backend/internal/services/image"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

var shopeeOrderRepoLogger = logger.Named("ShopeeOrderRepository")

// saveShopeeOrders saves Shopee orders using models package (upsert pattern)
func (r *GormOrderRepository) saveShopeeOrders(ctx context.Context, db *gorm.DB, orders []Order) error {
	for _, order := range orders {
		var totalAmount *float64
		if order.TotalAmount > 0 {
			totalAmount = &order.TotalAmount
		}

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
			if order.TrackingNumber != "" {
				updates["tracking_number"] = order.TrackingNumber
			}
			if order.ShipByDate > 0 {
				updates["ship_by_date"] = order.ShipByDate
			}

			if err := db.WithContext(ctx).Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			// Record doesn't exist - CREATE it
			var shipByDate *int64
			if order.ShipByDate > 0 {
				shipByDate = &order.ShipByDate
			}
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
				TrackingNumber:  order.TrackingNumber,
				ShipByDate:      shipByDate,
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
	cacheService := imageService.NewProductImageCacheService()
	urlCache := make(map[string]string)

	for i := range items {
		items[i].ProductImage = r.cacheShopeeProductImage(ctx, db, cacheService, urlCache, items[i])
	}

	// Delete existing items first
	if err := db.WithContext(ctx).Where("order_sn = ?", orderSN).Delete(&models.ShopeeOrderItem{}).Error; err != nil {
		return err
	}

	// Insert new items
	for _, item := range items {
		qty := item.Quantity
		price := item.Price
		itemModel := models.ShopeeOrderItem{
			TenantID:     r.tenantID,
			OrderSN:      orderSN,
			ItemID:       item.ItemID,
			ItemName:     item.ProductName,
			ModelName:    item.VariationName,
			ModelSku:     item.SKU,
			Quantity:     &qty,
			Price:        &price,
			ProductImage: item.ProductImage,
		}
		if err := db.WithContext(ctx).Create(&itemModel).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *GormOrderRepository) cacheShopeeProductImage(
	ctx context.Context,
	db *gorm.DB,
	cacheService *imageService.ProductImageCacheService,
	urlCache map[string]string,
	item OrderItem,
) string {
	if item.ProductImage == "" {
		return ""
	}
	if isLocalImageURL(item.ProductImage) {
		return item.ProductImage
	}
	if cached, ok := urlCache[item.ProductImage]; ok {
		return cached
	}

	allowedHosts := []string{"shopee.co.id", "susercontent.com"}
	localPath, err := cacheService.CacheRemoteImage(
		ctx,
		r.tenantID,
		item.ProductImage,
		fmt.Sprintf("shopee_%d", item.ItemID),
		allowedHosts,
	)
	if err != nil || localPath == "" {
		return item.ProductImage
	}

	urlCache[item.ProductImage] = localPath
	r.upsertShopeeProductImageCache(ctx, db, item, localPath)
	return localPath
}

func (r *GormOrderRepository) upsertShopeeProductImageCache(
	ctx context.Context,
	db *gorm.DB,
	item OrderItem,
	localPath string,
) {
	if item.ItemID == 0 || localPath == "" {
		return
	}

	var product models.ShopeeProduct
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND item_id = ?", r.tenantID, item.ItemID).
		First(&product).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newProduct := models.ShopeeProduct{
				TenantID:    r.tenantID,
				ItemID:      item.ItemID,
				Name:        item.ProductName,
				Image:       item.ProductImage,
				LocalImages: models.JSONArray{localPath},
			}
			if createErr := db.WithContext(ctx).Create(&newProduct).Error; createErr != nil {
				shopeeOrderRepoLogger.WithFields(map[string]interface{}{
					"tenant_id": r.tenantID,
					"item_id":   item.ItemID,
				}).Warn("Failed to create shopee product cache: " + createErr.Error())
			}
		}
		return
	}

	updatedImages := appendUniqueLocalImage(product.LocalImages, localPath)
	updates := map[string]interface{}{
		"local_images": updatedImages,
	}
	if product.Image == "" && item.ProductImage != "" {
		updates["image"] = item.ProductImage
	}
	if product.Name == "" && item.ProductName != "" {
		updates["name"] = item.ProductName
	}

	if updateErr := db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Where("id = ?", product.ID).
		Updates(updates).Error; updateErr != nil {
		shopeeOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id": r.tenantID,
			"item_id":   item.ItemID,
		}).Warn("Failed to update shopee product image cache: " + updateErr.Error())
	}
}

func appendUniqueLocalImage(existing models.JSONArray, localPath string) models.JSONArray {
	if localPath == "" {
		return existing
	}
	for _, entry := range existing {
		if path, ok := entry.(string); ok && path == localPath {
			return existing
		}
	}
	return append(existing, localPath)
}

func isLocalImageURL(url string) bool {
	return strings.HasPrefix(url, "/uploads/") || strings.Contains(url, "/uploads/")
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
			shopeeOrderRepoLogger.WithFields(map[string]interface{}{
				"tenant_id":   r.tenantID,
				"order_count": len(orderSNs),
			}).Warn("Failed to fetch shopee items: " + err.Error())
		}
	}

	// Collect unique item_ids for image lookup
	itemIDSet := make(map[int64]bool)
	for _, item := range items {
		if item.ItemID > 0 {
			itemIDSet[item.ItemID] = true
		}
	}
	itemIDs := make([]int64, 0, len(itemIDSet))
	for id := range itemIDSet {
		itemIDs = append(itemIDs, id)
	}

	// Fetch product images from shopee_products cache
	imageMap := r.getProductImagesFromCache(ctx, db, itemIDs)

	// Group items by order_sn and enrich with images
	itemsByOrder := make(map[string][]models.ShopeeOrderItem)
	for i := range items {
		// Enrich item with cached product image if empty
		if items[i].ProductImage == "" && items[i].ItemID > 0 {
			if imgURL, ok := imageMap[items[i].ItemID]; ok {
				items[i].ProductImage = imgURL
			}
		}
		itemsByOrder[items[i].OrderSN] = append(itemsByOrder[items[i].OrderSN], items[i])
	}

	// Build flattened orders with items (same format as Node.js)
	return r.flattenShopeeOrders(orderModels, itemsByOrder), nil
}

// getProductImagesFromCache fetches product images from shopee_products table
func (r *GormOrderRepository) getProductImagesFromCache(ctx context.Context, db *gorm.DB, itemIDs []int64) map[int64]string {
	result := make(map[int64]string)
	if len(itemIDs) == 0 {
		return result
	}

	// Use temporary struct to handle LocalImages as JSONArray
	type ProductImage struct {
		ItemID      int64            `gorm:"column:item_id"`
		Image       string           `gorm:"column:image"`
		LocalImages models.JSONArray `gorm:"column:local_images"`
	}

	var products []ProductImage
	err := db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Select("item_id, image, local_images").
		Where("item_id IN ?", itemIDs).
		Find(&products).Error

	if err != nil {
		shopeeOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id":  r.tenantID,
			"item_count": len(itemIDs),
		}).Warn("Failed to fetch shopee product images: " + err.Error())
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
			result[p.ItemID] = imgURL
		}
	}
	return result
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

		// Extract ship_by_date as Unix timestamp for frontend countdown calculation
		var shipByDate int64 = 0
		if m.ShipByDate != nil && *m.ShipByDate > 0 {
			shipByDate = *m.ShipByDate
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
				TrackingNumber:  m.TrackingNumber,
				ShippingCarrier: m.ShippingCarrier,
				BuyerMessage:    m.BuyerMessage,
				ShipByDate:      shipByDate, // Unix timestamp for frontend
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
					TrackingNumber:  m.TrackingNumber,
					ShippingCarrier: m.ShippingCarrier,
					BuyerMessage:    m.BuyerMessage,
					SKU:             sku,
					ProductName:     item.ItemName,
					VariationName:   item.ModelName,
					Quantity:        qty,
					Price:           price,
					ProductImage:    item.ProductImage,
					ShipByDate:      shipByDate, // Unix timestamp for frontend countdown
					CreatedAt:       m.CreatedAt,
					UpdatedAt:       m.UpdatedAt,
				})
			}
		}
	}

	return orders
}
