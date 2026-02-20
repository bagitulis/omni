package sync

import (
	"context"
	"errors"
	"strconv"

	"github.com/omni/backend/internal/models"
	imageService "github.com/omni/backend/internal/services/image"
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
				BuyerUsername:    order.BuyerUsername,
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
	for i := range items {
		if items[i].ItemID == 0 && items[i].SKU != "" {
			var skuModel models.TiktokSku
			if err := db.WithContext(ctx).Where("seller_sku = ? OR sku_id = ?", items[i].SKU, items[i].SKU).First(&skuModel).Error; err == nil {
				items[i].ItemID = int64(skuModel.ProductID)
			}
		}
	}
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
		productID := item.ItemID

		itemModel := models.TiktokOrderItem{
			TenantID:      r.tenantID,
			OrderSN:       orderSN,
			LineItemID:    item.ID,
			ProductID:     productID,
			SkuID:         item.SkuID,
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

	// Check if image was already downloaded during product sync (dedup)
	if existing := r.findExistingImageByURL(ctx, db, item.ProductImage); existing != "" {
		urlCache[item.ProductImage] = existing
		return existing
	}

	allowedHosts := []string{"ibyteimg.com", "tiktokcdn.com"}
	localPath, err := cacheService.CacheRemoteImage(
		ctx,
		r.tenantID,
		item.ProductImage,
		"asset",
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

	// Query by product_id (TikTok API ID as string), not by internal id
	productIDStr := strconv.FormatInt(item.ItemID, 10)
	var product models.TiktokProduct
	err := db.WithContext(ctx).
		Where("product_id = ?", productIDStr).
		First(&product).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		tiktokOrderRepoLogger.WithFields(map[string]interface{}{
			"tenant_id":  r.tenantID,
			"product_id": productIDStr,
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
			"tenant_id":  r.tenantID,
			"product_id": productIDStr,
		}).Warn("Failed to update tiktok product image cache: " + updateErr.Error())
	}
}
