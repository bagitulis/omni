package sync

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	imageService "github.com/omni/backend/internal/services/image"
	"gorm.io/gorm"
)

// saveShopeeOrders saves Shopee orders using models package (upsert pattern)
func (r *GormOrderRepository) saveShopeeOrders(ctx context.Context, db *gorm.DB, orders []Order) error {
	for _, order := range orders {
		var totalAmount *float64
		if order.TotalAmount > 0 {
			totalAmount = &order.TotalAmount
		}

		var existing models.ShopeeOrder
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
	masterImageMap := r.getShopeeMasterImages(ctx, db, items)

	for i := range items {
		if masterImage, ok := masterImageMap[items[i].ItemID]; ok && masterImage != "" {
			items[i].ProductImage = masterImage
			continue
		}
		items[i].ProductImage = r.cacheShopeeProductImage(ctx, db, cacheService, urlCache, items[i])
	}

	if err := db.WithContext(ctx).Where("order_sn = ?", orderSN).Delete(&models.ShopeeOrderItem{}).Error; err != nil {
		return err
	}

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

	if existing := r.findExistingImageByURL(ctx, db, item.ProductImage); existing != "" {
		urlCache[item.ProductImage] = existing
		return existing
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
