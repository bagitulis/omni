package sync

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/models"
	imageService "github.com/omni/backend/internal/services/image"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

var tiktokOrderRepoLogger = logger.Named("TiktokOrderRepository")

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
				TrackingNumber:  order.TrackingNumber,
				ShipByDate:      shipByDate,
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
	cacheService := imageService.NewProductImageCacheService()
	urlCache := make(map[string]string)
	masterImageMap := r.getTiktokMasterImages(ctx, db, items)
	for i := range items {
		if img, ok := masterImageMap[items[i].ItemID]; ok && img != "" {
			items[i].ProductImage = img
			continue
		}
		items[i].ProductImage = r.cacheTiktokProductImage(ctx, db, cacheService, urlCache, items[i])
	}

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

func (r *GormOrderRepository) cacheTiktokProductImage(
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

	allowedHosts := []string{"ibyteimg.com", "tiktokcdn.com"}
	localPath, err := cacheService.CacheRemoteImage(
		ctx,
		r.tenantID,
		item.ProductImage,
		fmt.Sprintf("tiktok_%d", item.ItemID),
		allowedHosts,
	)
	if err != nil || localPath == "" {
		return item.ProductImage
	}

	urlCache[item.ProductImage] = localPath
	r.upsertTiktokProductImageCache(ctx, db, item, localPath)
	return localPath
}

func (r *GormOrderRepository) upsertTiktokProductImageCache(
	ctx context.Context,
	db *gorm.DB,
	item OrderItem,
	localPath string,
) {
	if item.ItemID == 0 || localPath == "" {
		return
	}

	var product models.TiktokProduct
	err := db.WithContext(ctx).
		Where("id = ?", item.ItemID).
		First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		tiktokOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id": r.tenantID,
			"item_id":   item.ItemID,
		}).Warn("Failed to find tiktok product cache: " + err.Error())
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
		Model(&models.TiktokProduct{}).
		Where("id = ?", product.ID).
		Updates(updates).Error; updateErr != nil {
		tiktokOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id": r.tenantID,
			"item_id":   item.ItemID,
		}).Warn("Failed to update tiktok product image cache: " + updateErr.Error())
	}
}

func (r *GormOrderRepository) getTiktokMasterImages(ctx context.Context, db *gorm.DB, items []OrderItem) map[int64]string {
	idSet := make(map[int64]bool)
	for _, item := range items {
		if item.ItemID > 0 {
			idSet[item.ItemID] = true
		}
	}
	ids := make([]int64, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	return r.getTiktokMasterImagesByIDs(ctx, db, ids)
}

func (r *GormOrderRepository) getTiktokMasterImagesByIDs(ctx context.Context, db *gorm.DB, productIDs []int64) map[int64]string {
	result := make(map[int64]string)
	if len(productIDs) == 0 {
		return result
	}

	itemIDStrings := make([]string, 0, len(productIDs))
	for _, id := range productIDs {
		if id > 0 {
			itemIDStrings = append(itemIDStrings, strconv.FormatInt(id, 10))
		}
	}

	masterMap := r.getMasterProductImagesByItemIDs(ctx, db, models.PlatformTiktok, itemIDStrings)
	for _, id := range productIDs {
		key := strconv.FormatInt(id, 10)
		if image, ok := masterMap[key]; ok && image != "" {
			result[id] = image
		}
	}

	return result
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
			tiktokOrderRepoLogger.WithFields(map[string]interface{}{
				"tenant_id":   r.tenantID,
				"order_count": len(orderSNs),
			}).Warn("Failed to fetch tiktok items: " + err.Error())
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

	// Fetch product images from master product cache first, then tiktok_products cache
	masterImageMap := r.getTiktokMasterImagesByIDs(ctx, db, productIDs)
	imageMap := r.getTiktokProductImagesFromCache(ctx, db, productIDs)

	// Group items by order_sn and enrich with images
	itemsByOrder := make(map[string][]models.TiktokOrderItem)
	for i := range items {
		if items[i].ProductID > 0 {
			if masterImg, ok := masterImageMap[items[i].ProductID]; ok && masterImg != "" {
				items[i].ProductImage = masterImg
			}
		}
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
		tiktokOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id":  r.tenantID,
			"item_count": len(productIDs),
		}).Warn("Failed to fetch tiktok product images: " + err.Error())
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
				TrackingNumber:  m.TrackingNumber,
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
					TrackingNumber:  m.TrackingNumber,
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
