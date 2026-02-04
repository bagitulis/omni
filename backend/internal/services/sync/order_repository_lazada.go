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

var lazadaOrderRepoLogger = logger.Named("LazadaOrderRepository")

// saveLazadaOrders saves Lazada orders using models package
func (r *GormOrderRepository) saveLazadaOrders(ctx context.Context, db *gorm.DB, orders []Order) error {
	for _, order := range orders {
		var totalAmount *float64
		if order.TotalAmount > 0 {
			totalAmount = &order.TotalAmount
		}

		var shipByDate *int64
		if order.ShipByDate > 0 {
			shipByDate = &order.ShipByDate
		}

		var existing models.LazadaOrder
		result := db.WithContext(ctx).Where("order_sn = ?", order.OrderSN).First(&existing)
		if result.Error == nil {
			updates := map[string]interface{}{
				"order_status": order.Status,
			}
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
			model := models.LazadaOrder{
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
			if err := r.saveLazadaOrderItems(ctx, db, order.OrderSN, order.Items); err != nil {
				return err
			}
		}
	}
	return nil
}

// saveLazadaOrderItems saves Lazada order items
func (r *GormOrderRepository) saveLazadaOrderItems(ctx context.Context, db *gorm.DB, orderSN string, items []OrderItem) error {
	cacheService := imageService.NewProductImageCacheService()
	urlCache := make(map[string]string)
	masterImageMap := r.getLazadaMasterImages(ctx, db, items)
	for i := range items {
		itemID := fmt.Sprintf("%d", items[i].ItemID)
		if masterImage, ok := masterImageMap[itemID]; ok && masterImage != "" {
			items[i].ProductImage = masterImage
			continue
		}
		items[i].ProductImage = r.cacheLazadaProductImage(ctx, db, cacheService, urlCache, items[i])
	}

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

func (r *GormOrderRepository) cacheLazadaProductImage(
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

	localPath, err := cacheService.CacheRemoteImage(
		ctx,
		r.tenantID,
		item.ProductImage,
		fmt.Sprintf("lazada_%d", item.ItemID),
		nil,
	)
	if err != nil || localPath == "" {
		return item.ProductImage
	}

	urlCache[item.ProductImage] = localPath
	r.upsertLazadaProductImageCache(ctx, db, item, localPath)
	return localPath
}

func (r *GormOrderRepository) upsertLazadaProductImageCache(
	ctx context.Context,
	db *gorm.DB,
	item OrderItem,
	localPath string,
) {
	if item.ItemID == 0 || localPath == "" {
		return
	}
	itemIDStr := fmt.Sprintf("%d", item.ItemID)

	var product models.LazadaProduct
	err := db.WithContext(ctx).
		Where("tenant_id = ? AND item_id = ?", r.tenantID, itemIDStr).
		First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newProduct := models.LazadaProduct{
				TenantID:    r.tenantID,
				ItemID:      itemIDStr,
				Name:        item.ProductName,
				Image:       item.ProductImage,
				LocalImages: models.JSONArray{localPath},
			}
			if createErr := db.WithContext(ctx).Create(&newProduct).Error; createErr != nil {
				lazadaOrderRepoLogger.WithFields(map[string]interface{}{
					"tenant_id": r.tenantID,
					"item_id":   itemIDStr,
				}).Warn("Failed to create lazada product cache: " + createErr.Error())
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
		Model(&models.LazadaProduct{}).
		Where("id = ?", product.ID).
		Updates(updates).Error; updateErr != nil {
		lazadaOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id": r.tenantID,
			"item_id":   itemIDStr,
		}).Warn("Failed to update lazada product image cache: " + updateErr.Error())
	}
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
			lazadaOrderRepoLogger.WithFields(map[string]interface{}{
				"tenant_id":   r.tenantID,
				"order_count": len(orderSNs),
			}).Warn("Failed to fetch lazada items: " + err.Error())
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

	// Fetch master product images first, then lazada_products cache
	masterImageMap := r.getMasterProductImagesByItemIDs(ctx, db, models.PlatformLazada, itemIDs)
	imageMap := r.getLazadaProductImagesFromCache(ctx, db, itemIDs)

	// Group items by order_sn and enrich with images
	itemsByOrder := make(map[string][]models.LazadaOrderItem)
	for i := range items {
		itemIDStr := fmt.Sprintf("%d", items[i].ItemID)
		if masterImage, ok := masterImageMap[itemIDStr]; ok && masterImage != "" {
			items[i].ProductImage = masterImage
		}
		// Enrich item with cached product image if empty
		if items[i].ProductImage == "" && items[i].ItemID > 0 {
			if imgURL, ok := imageMap[itemIDStr]; ok {
				items[i].ProductImage = imgURL
			}
		}
		itemsByOrder[items[i].OrderSN] = append(itemsByOrder[items[i].OrderSN], items[i])
	}

	// Build flattened orders with items
	return r.flattenLazadaOrders(orderModels, itemsByOrder), nil
}

func (r *GormOrderRepository) getLazadaMasterImages(ctx context.Context, db *gorm.DB, items []OrderItem) map[string]string {
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
	if len(itemIDs) == 0 {
		return map[string]string{}
	}
	return r.getMasterProductImagesByItemIDs(ctx, db, models.PlatformLazada, itemIDs)
}

// getLazadaProductImagesFromCache fetches product images from lazada_products table
func (r *GormOrderRepository) getLazadaProductImagesFromCache(ctx context.Context, db *gorm.DB, itemIDs []string) map[string]string {
	result := make(map[string]string)
	if len(itemIDs) == 0 {
		return result
	}

	// Use temporary struct to handle LocalImages as JSONArray
	type ProductImage struct {
		ItemID      string           `gorm:"column:item_id"`
		Image       string           `gorm:"column:image"`
		LocalImages models.JSONArray `gorm:"column:local_images"`
	}

	var products []ProductImage
	err := db.WithContext(ctx).
		Model(&models.LazadaProduct{}).
		Select("item_id, image, local_images").
		Where("item_id IN ?", itemIDs).
		Find(&products).Error

	if err != nil {
		lazadaOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id":  r.tenantID,
			"item_count": len(itemIDs),
		}).Warn("Failed to fetch lazada product images: " + err.Error())
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

// flattenLazadaOrders creates flattened order list for frontend
func (r *GormOrderRepository) flattenLazadaOrders(orderModels []models.LazadaOrder, itemsByOrder map[string][]models.LazadaOrderItem) []Order {
	orders := make([]Order, 0)

	for _, m := range orderModels {
		orderItems := itemsByOrder[m.OrderSN]

		// Get total amount, currency, buyer username from order model
		totalAmount := float64(0)
		if m.TotalAmount != nil {
			totalAmount = *m.TotalAmount
		}

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
			orders = append(orders, Order{
				ID:              fmt.Sprintf("%d", m.ID),
				OrderSN:         m.OrderSN,
				OrderNo:         m.OrderSN,
				Platform:        strings.ToUpper("lazada"),
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
			for _, item := range orderItems {
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
					Platform:        strings.ToUpper("lazada"),
					Status:          m.OrderStatus,
					TotalAmount:     totalAmount,
					Currency:        currency,
					BuyerUsername:   m.BuyerUsername,
					PaymentMethod:   m.PaymentMethod,
					TrackingNumber:  m.TrackingNumber,
					ShippingCarrier: m.ShippingCarrier, // From order header
					ShipByDate:      shipByDate,        // Unix timestamp for frontend
					SKU:             item.SellerSku,
					ProductName:     item.ProductName,
					VariationName:   item.VariationName,
					Quantity:        qty,
					Price:           price,
					OrderItemID:     item.ItemID,
					ProductImage:    item.ProductImage,
					CreatedAt:       m.CreatedAt,
					UpdatedAt:       m.UpdatedAt,
				})
			}
		}
	}

	return orders
}
