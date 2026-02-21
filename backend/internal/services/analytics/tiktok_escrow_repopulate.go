// Package analytics provides re-population of TikTok escrow items from raw stored JSON
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
)

// tiktokRawOrderData mirrors the line_items structure stored in raw_order_data
type tiktokRawOrderData struct {
	ID        string              `json:"id"`
	Status    string              `json:"status"`
	LineItems []tiktokRawLineItem `json:"line_items"`
}

type tiktokRawLineItem struct {
	ID               string `json:"id"`
	ProductID        string `json:"product_id"`
	ProductName      string `json:"product_name"`
	SkuID            string `json:"sku_id"`
	SkuName          string `json:"sku_name"`
	SellerSku        string `json:"seller_sku"`
	Quantity         int    `json:"quantity"`
	OriginalPrice    string `json:"original_price"`
	SalePrice        string `json:"sale_price"`
	PlatformDiscount string `json:"platform_discount"`
	SellerDiscount   string `json:"seller_discount"`
}

// RepopulateEscrowItems re-populates tiktok_escrow_items by parsing raw_order_data
// stored in tiktok_escrow_orders. Fixes cases where items were not saved during sync.
func (s *TiktokEscrowSyncService) RepopulateEscrowItems(ctx context.Context, month, year int) (int, int, error) {
	tables := TiktokEscrowTables()
	ordersTable := s.base.Table(tables.OrderTable)
	itemsTable := s.base.Table(tables.ItemTable)

	log.Printf("[TiktokRepopulate] Starting re-populate for %d-%02d, tenant: %s", year, month, s.tenantID)

	// Fetch all escrow orders for this month that have raw_order_data
	var orders []models.TiktokEscrowOrder
	err := s.base.DB.WithContext(ctx).Table(ordersTable).
		Where("tenant_id = ? AND month = ? AND year = ? AND raw_order_data IS NOT NULL", s.tenantID, month, year).
		Find(&orders).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch orders: %w", err)
	}

	log.Printf("[TiktokRepopulate] Found %d orders to process", len(orders))

	processedOrders := 0
	totalItems := 0

	for _, order := range orders {
		if order.RawOrderData == nil || *order.RawOrderData == "" {
			continue
		}

		// Parse raw_order_data JSON
		var rawOrder tiktokRawOrderData
		if err := json.Unmarshal([]byte(*order.RawOrderData), &rawOrder); err != nil {
			log.Printf("[TiktokRepopulate] Failed to parse raw_order_data for order %s: %v", order.OrderID, err)
			continue
		}

		if len(rawOrder.LineItems) == 0 {
			log.Printf("[TiktokRepopulate] No line_items in raw_order_data for order %s", order.OrderID)
			continue
		}

		// Delete existing items for this order
		s.base.DB.WithContext(ctx).Table(itemsTable).
			Where("tenant_id = ? AND escrow_order_id = ?", s.tenantID, order.ID).
			Delete(nil)

		// Re-insert items from line_items
		for _, item := range rawOrder.LineItems {
			rawItemData, _ := json.Marshal(item)

			qty := item.Quantity
			if qty == 0 {
				qty = 1
			}

			escrowItem := models.TiktokEscrowItem{
				ID:               uuid.New().String(),
				TenantID:         s.tenantID,
				EscrowOrderID:    order.ID,
				OrderID:          order.OrderID,
				ProductID:        StringPtr(item.ProductID),
				ProductName:      StringPtr(item.ProductName),
				SkuID:            StringPtr(item.SkuID),
				SellerSku:        StringPtr(item.SellerSku),
				Quantity:         qty,
				OriginalPrice:    ParseFloat(item.OriginalPrice),
				SalePrice:        ParseFloat(item.SalePrice),
				PlatformDiscount: ParseFloat(item.PlatformDiscount),
				SellerDiscount:   ParseFloat(item.SellerDiscount),
				RawItemData:      StringPtr(string(rawItemData)),
				SyncedAt:         time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			if err := s.base.DB.WithContext(ctx).Table(itemsTable).Create(&escrowItem).Error; err != nil {
				log.Printf("[TiktokRepopulate] Failed to insert item for order %s (sku=%s): %v", order.OrderID, item.SellerSku, err)
				continue
			}
			totalItems++
		}

		processedOrders++
	}

	log.Printf("[TiktokRepopulate] Done. Processed %d orders, inserted %d items", processedOrders, totalItems)
	return processedOrders, totalItems, nil
}
