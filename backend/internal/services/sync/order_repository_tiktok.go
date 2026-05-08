package sync

import (
	"context"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

var tiktokOrderRepoLogger = logger.Named("TiktokOrderRepository")

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

	// Convert int64 product IDs to strings for product_id column query
	productIDStrings := make([]string, 0, len(productIDs))
	for _, id := range productIDs {
		if id > 0 {
			productIDStrings = append(productIDStrings, strconv.FormatInt(id, 10))
		}
	}
	if len(productIDStrings) == 0 {
		return result
	}

	// Query master product images using TikTok API product_id strings
	masterMap := r.getMasterProductImagesByItemIDs(ctx, db, models.PlatformTiktok, productIDStrings)
	for _, id := range productIDs {
		if id <= 0 {
			continue
		}
		lookupID := strconv.FormatInt(id, 10)
		if image, ok := masterMap[lookupID]; ok && image != "" {
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
// NOTE: productIDs are TikTok API product_id values (large 64-bit integers), queried via product_id column
func (r *GormOrderRepository) getTiktokProductImagesFromCache(ctx context.Context, db *gorm.DB, productIDs []int64) map[int64]string {
	result := make(map[int64]string)
	if len(productIDs) == 0 {
		return result
	}

	// Convert int64 IDs to strings for product_id column query
	productIDStrings := make([]string, 0, len(productIDs))
	for _, id := range productIDs {
		if id > 0 {
			productIDStrings = append(productIDStrings, strconv.FormatInt(id, 10))
		}
	}
	if len(productIDStrings) == 0 {
		return result
	}

	// Use temporary struct to handle LocalImages as JSONArray
	// Query by product_id (TikTok API string), not id (internal primary key)
	type ProductImage struct {
		ProductID   string           `gorm:"column:product_id"`
		Image       string           `gorm:"column:image"`
		LocalImages models.JSONArray `gorm:"column:local_images"`
	}

	var products []ProductImage
	err := db.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Select("product_id, image, local_images").
		Where("product_id IN ?", productIDStrings).
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
			// Convert product_id string back to int64 for result map
			if pid, err := strconv.ParseInt(p.ProductID, 10, 64); err == nil {
				result[pid] = imgURL
			}
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
				ID:              m.ID,
				OrderSN:         m.OrderSN,
				OrderNo:         m.OrderSN,
				Platform:        "TIKTOK",
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
			orders = append(orders, r.flattenTiktokOrderItems(m, orderItems, totalAmount, currency, shipByDate, countdown)...)
		}
	}

	return orders
}

// flattenTiktokOrderItems groups TikTok order items by SKU and creates flattened Order entries
func (r *GormOrderRepository) flattenTiktokOrderItems(
	m models.TiktokOrder,
	orderItems []models.TiktokOrderItem,
	totalAmount float64,
	currency string,
	shipByDate int64,
	countdown string,
) []Order {
	// Group items by SkuID (model number) to accumulate quantities
	// TikTok returns one line_item per unit, so 10x same SKU = 10 line_items with qty=1
	type groupedItem struct {
		item models.TiktokOrderItem
		qty  int
	}
	groupKey := func(item models.TiktokOrderItem) string {
		if item.SkuID != "" {
			return item.SkuID
		}
		if item.SellerSku != "" {
			return item.SellerSku
		}
		// Last resort: product_name + variation_name
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

	orders := make([]Order, 0, len(keyOrder))
	for _, key := range keyOrder {
		g := seen[key]
		price := float64(0)
		if g.item.Price != nil {
			price = *g.item.Price
		}

		// Use SellerSku for display; fallback to SkuID
		displaySku := g.item.SellerSku
		if displaySku == "" {
			displaySku = g.item.SkuID
		}

		orders = append(orders, Order{
			ID:              m.ID,
			OrderSN:         m.OrderSN,
			OrderNo:         m.OrderSN,
			Platform:        "TIKTOK",
			Status:          m.OrderStatus,
			TotalAmount:     totalAmount,
			Currency:        currency,
			BuyerUsername:    m.BuyerUsername,
			PaymentMethod:   m.PaymentMethod,
			TrackingNumber:  m.TrackingNumber,
			ShippingCarrier: m.ShippingCarrier,
			BuyerMessage:    m.BuyerMessage,
			SKU:             displaySku,
			ProductName:     g.item.ProductName,
			VariationName:   g.item.VariationName,
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
