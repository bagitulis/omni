package sync

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

var shopeeOrderRepoLogger = logger.Named("ShopeeOrderRepository")

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

	orderSNs := make([]string, 0, len(orderModels))
	for _, m := range orderModels {
		orderSNs = append(orderSNs, m.OrderSN)
	}

	var items []models.ShopeeOrderItem
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).Where("order_sn IN ?", orderSNs).Find(&items).Error; err != nil {
			shopeeOrderRepoLogger.WithFields(map[string]interface{}{
				"tenant_id":   r.tenantID,
				"order_count": len(orderSNs),
			}).Warn("Failed to fetch shopee items: " + err.Error())
		}
	}

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

	imageMap := r.getProductImagesFromCache(ctx, db, itemIDs)
	masterImageMap := r.getShopeeMasterImagesByIDs(ctx, db, itemIDs)

	itemsByOrder := make(map[string][]models.ShopeeOrderItem)
	for i := range items {
		if items[i].ItemID > 0 {
			if masterImage, ok := masterImageMap[items[i].ItemID]; ok && masterImage != "" {
				items[i].ProductImage = masterImage
			}
		}
		if items[i].ProductImage == "" && items[i].ItemID > 0 {
			if imgURL, ok := imageMap[items[i].ItemID]; ok {
				items[i].ProductImage = imgURL
			}
		}
		itemsByOrder[items[i].OrderSN] = append(itemsByOrder[items[i].OrderSN], items[i])
	}

	return r.flattenShopeeOrders(orderModels, itemsByOrder), nil
}

func (r *GormOrderRepository) getShopeeMasterImages(ctx context.Context, db *gorm.DB, items []OrderItem) map[int64]string {
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
	return r.getShopeeMasterImagesByIDs(ctx, db, itemIDs)
}

func (r *GormOrderRepository) getShopeeMasterImagesByIDs(ctx context.Context, db *gorm.DB, itemIDs []int64) map[int64]string {
	result := make(map[int64]string)
	if len(itemIDs) == 0 {
		return result
	}
	itemIDStrings := make([]string, 0, len(itemIDs))
	for _, id := range itemIDs {
		if id > 0 {
			itemIDStrings = append(itemIDStrings, fmt.Sprintf("%d", id))
		}
	}

	masterMap := r.getMasterProductImagesByItemIDs(ctx, db, models.PlatformShopee, itemIDStrings)
	for _, id := range itemIDs {
		key := fmt.Sprintf("%d", id)
		if image, ok := masterMap[key]; ok && image != "" {
			result[id] = image
		}
	}
	return result
}

// getProductImagesFromCache fetches product images from shopee_products table
func (r *GormOrderRepository) getProductImagesFromCache(ctx context.Context, db *gorm.DB, itemIDs []int64) map[int64]string {
	result := make(map[int64]string)
	if len(itemIDs) == 0 {
		return result
	}

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
		if len(p.LocalImages) > 0 {
			if path, ok := p.LocalImages[0].(string); ok && path != "" {
				imgURL = path
			}
		}
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

		totalAmount := float64(0)
		if m.TotalAmount != nil {
			totalAmount = *m.TotalAmount
		}

		currency := m.Currency
		if currency == "" {
			currency = "IDR"
		}

		var shipByDate int64 = 0
		countdown := ""
		if m.ShipByDate != nil && *m.ShipByDate > 0 {
			shipByDate = *m.ShipByDate
			countdown = formatShipByDateFromDB(shipByDate)
		}

		if len(orderItems) == 0 {
			orders = append(orders, Order{
				ID:              fmt.Sprintf("%d", m.ID),
				OrderSN:         m.OrderSN,
				OrderNo:         m.OrderSN,
				Platform:        "SHOPEE",
				Status:          m.OrderStatus,
				TotalAmount:     totalAmount,
				Currency:        currency,
				BuyerUsername:    m.BuyerUsername,
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
			orders = append(orders, r.flattenShopeeOrderItems(m, orderItems, totalAmount, currency, shipByDate, countdown)...)
		}
	}

	return orders
}

// flattenShopeeOrderItems groups Shopee order items by SKU and creates flattened Order entries
func (r *GormOrderRepository) flattenShopeeOrderItems(
	m models.ShopeeOrder,
	orderItems []models.ShopeeOrderItem,
	totalAmount float64,
	currency string,
	shipByDate int64,
	countdown string,
) []Order {
	type groupedItem struct {
		item models.ShopeeOrderItem
		qty  int
	}
	groupKey := func(item models.ShopeeOrderItem) string {
		if item.ModelSku != "" {
			return item.ModelSku
		}
		if item.ItemSku != "" {
			return item.ItemSku
		}
		return item.ItemName + "|" + item.ModelName
	}

	seen := make(map[string]*groupedItem)
	var keyOrder []string
	for _, item := range orderItems {
		key := groupKey(item)
		itemQty := 0
		if item.Quantity != nil {
			itemQty = *item.Quantity
		}
		if itemQty == 0 {
			itemQty = 1
		}

		if existing, ok := seen[key]; ok {
			existing.qty += itemQty
		} else {
			seen[key] = &groupedItem{item: item, qty: itemQty}
			keyOrder = append(keyOrder, key)
		}
	}

	orders := make([]Order, 0, len(keyOrder))
	for _, key := range keyOrder {
		g := seen[key]
		price := float64(0)
		if g.item.Price != nil {
			price = *g.item.Price
		}

		displaySku := g.item.ModelSku
		if displaySku == "" {
			displaySku = g.item.ItemSku
		}

		orders = append(orders, Order{
			ID:              fmt.Sprintf("%d", m.ID),
			OrderSN:         m.OrderSN,
			OrderNo:         m.OrderSN,
			Platform:        "SHOPEE",
			Status:          m.OrderStatus,
			TotalAmount:     totalAmount,
			Currency:        currency,
			BuyerUsername:    m.BuyerUsername,
			PaymentMethod:   m.PaymentMethod,
			TrackingNumber:  m.TrackingNumber,
			ShippingCarrier: m.ShippingCarrier,
			BuyerMessage:    m.BuyerMessage,
			SKU:             displaySku,
			ProductName:     g.item.ItemName,
			VariationName:   g.item.ModelName,
			Quantity:        g.qty,
			Price:           price,
			ProductImage:    g.item.ProductImage,
			ShipByDate:      shipByDate,
			Countdown:       countdown,
			CreatedAt:       m.CreatedAt,
			UpdatedAt:       m.UpdatedAt,
		})
	}
	return orders
}
