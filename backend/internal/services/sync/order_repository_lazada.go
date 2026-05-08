package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

var lazadaOrderRepoLogger = logger.Named("LazadaOrderRepository")

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

	orderSNs := make([]string, 0, len(orderModels))
	for _, m := range orderModels {
		orderSNs = append(orderSNs, m.OrderSN)
	}

	var items []models.LazadaOrderItem
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).Where("order_sn IN ?", orderSNs).Find(&items).Error; err != nil {
			lazadaOrderRepoLogger.WithFields(map[string]interface{}{
				"tenant_id":   r.tenantID,
				"order_count": len(orderSNs),
			}).Warn("Failed to fetch lazada items: " + err.Error())
		}
	}

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

	masterImageMap := r.getMasterProductImagesByItemIDs(ctx, db, models.PlatformLazada, itemIDs)
	imageMap := r.getLazadaProductImagesFromCache(ctx, db, itemIDs)

	itemsByOrder := make(map[string][]models.LazadaOrderItem)
	for i := range items {
		itemIDStr := fmt.Sprintf("%d", items[i].ItemID)
		if masterImage, ok := masterImageMap[itemIDStr]; ok && masterImage != "" {
			items[i].ProductImage = masterImage
		}
		if items[i].ProductImage == "" && items[i].ItemID > 0 {
			if imgURL, ok := imageMap[itemIDStr]; ok {
				items[i].ProductImage = imgURL
			}
		}
		itemsByOrder[items[i].OrderSN] = append(itemsByOrder[items[i].OrderSN], items[i])
	}

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

// flattenLazadaOrders creates flattened order list for frontend
func (r *GormOrderRepository) flattenLazadaOrders(orderModels []models.LazadaOrder, itemsByOrder map[string][]models.LazadaOrderItem) []Order {
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
				ID:              m.ID,
				OrderSN:         m.OrderSN,
				OrderNo:         m.OrderSN,
				Platform:        strings.ToUpper("lazada"),
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
			orders = append(orders, r.flattenLazadaOrderItems(m, orderItems, totalAmount, currency, shipByDate, countdown)...)
		}
	}

	return orders
}

// flattenLazadaOrderItems groups Lazada order items by SKU and creates flattened Order entries
func (r *GormOrderRepository) flattenLazadaOrderItems(
	m models.LazadaOrder,
	orderItems []models.LazadaOrderItem,
	totalAmount float64,
	currency string,
	shipByDate int64,
	countdown string,
) []Order {
	type groupedItem struct {
		item models.LazadaOrderItem
		qty  int
	}
	groupKey := func(item models.LazadaOrderItem) string {
		if item.SkuID != "" {
			return item.SkuID
		}
		if item.SellerSku != "" {
			return item.SellerSku
		}
		return item.ProductName + "|" + item.VariationName
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

	result := make([]Order, 0, len(keyOrder))
	for _, key := range keyOrder {
		g := seen[key]
		price := float64(0)
		if g.item.Price != nil {
			price = *g.item.Price
		}

		displaySku := g.item.SellerSku
		if displaySku == "" {
			displaySku = g.item.SkuID
		}

		result = append(result, Order{
			ID:              m.ID,
			OrderSN:         m.OrderSN,
			OrderNo:         m.OrderSN,
			Platform:        strings.ToUpper("lazada"),
			Status:          m.OrderStatus,
			TotalAmount:     totalAmount,
			Currency:        currency,
			BuyerUsername:    m.BuyerUsername,
			PaymentMethod:   m.PaymentMethod,
			TrackingNumber:  m.TrackingNumber,
			ShippingCarrier: m.ShippingCarrier,
			ShipByDate:      shipByDate,
			Countdown:       countdown,
			BuyerMessage:    m.BuyerMessage,
			SKU:             displaySku,
			ProductName:     g.item.ProductName,
			VariationName:   g.item.VariationName,
			Quantity:        g.qty,
			Price:           price,
			OrderItemID:     g.item.ItemID,
			ProductImage:    g.item.ProductImage,
			CreatedAt:       m.CreatedAt,
			UpdatedAt:       m.UpdatedAt,
		})
	}
	return result
}
