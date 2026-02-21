// Package analytics provides re-population of escrow items from raw stored JSON
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// RepopulateEscrowItems re-populates shopee_escrow_items by parsing raw_order_income
// stored in shopee_escrow_orders. This fixes cases where items were not saved during sync
// because shopee_escrow_items is empty but raw JSON data exists.
func (s *ShopeeEscrowSyncService) RepopulateEscrowItems(ctx context.Context, month, year int) (int, int, error) {
	tables := ShopeeEscrowTables()
	ordersTable := s.base.Table(tables.OrderTable)
	itemsTable := s.base.Table(tables.ItemTable)

	log.Printf("[ShopeeRepopulate] Starting re-populate for %d-%02d, tenant: %s", year, month, s.tenantID)

	// Check current items count
	var currentCount int64
	s.base.DB.WithContext(ctx).Table(itemsTable).
		Where("tenant_id = ?", s.tenantID).
		Count(&currentCount)
	log.Printf("[ShopeeRepopulate] Current items count: %d", currentCount)

	// Fetch all escrow orders for this month that have raw_order_income
	var orders []models.ShopeeEscrowOrder
	err := s.base.DB.WithContext(ctx).Table(ordersTable).
		Where("tenant_id = ? AND month = ? AND year = ? AND raw_order_income IS NOT NULL", s.tenantID, month, year).
		Find(&orders).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch orders: %w", err)
	}

	log.Printf("[ShopeeRepopulate] Found %d orders to process", len(orders))

	processedOrders := 0
	totalItems := 0

	for _, order := range orders {
		if order.RawOrderIncome == nil || *order.RawOrderIncome == "" {
			continue
		}

		// Parse raw_order_income JSON
		var income shopeePkg.EscrowOrderData
		if err := json.Unmarshal([]byte(*order.RawOrderIncome), &income); err != nil {
			log.Printf("[ShopeeRepopulate] Failed to parse raw_order_income for order %s: %v", order.OrderSN, err)
			continue
		}

		if len(income.Items) == 0 {
			continue
		}

		// Delete existing items for this order (raw delete by escrow_order_id)
		s.base.DB.WithContext(ctx).Table(itemsTable).
			Where("tenant_id = ? AND escrow_order_id = ?", s.tenantID, order.ID).
			Delete(nil)

		// Re-insert items
		for _, item := range income.Items {
			rawItemData, _ := json.Marshal(item)

			escrowItem := models.ShopeeEscrowItem{
				ID:                        uuid.New().String(),
				TenantID:                  s.tenantID,
				EscrowOrderID:             order.ID,
				OrderSN:                   StringPtr(order.OrderSN),
				Month:                     &month,
				Year:                      &year,
				ItemID:                    Int64Ptr(item.ItemID),
				ModelID:                   Int64Ptr(item.ModelID),
				Sku:                       StringPtr(item.ItemSKU),
				ModelSku:                  StringPtr(item.ModelSKU),
				ItemName:                  StringPtr(item.ItemName),
				ModelName:                 StringPtr(item.ModelName),
				Quantity:                  item.QuantityPurchased,
				OriginalPrice:             item.OriginalPrice,
				SellingPrice:              item.SellingPrice,
				DiscountedPrice:           item.DiscountedPrice,
				SellerDiscount:            item.SellerDiscount,
				ShopeeDiscount:            item.ShopeeDiscount,
				DiscountFromCoin:          item.DiscountFromCoin,
				DiscountFromVoucherSeller: item.DiscountFromVoucherSeller,
				DiscountFromVoucherShopee: item.DiscountFromVoucherShopee,
				AmsCommissionFee:          item.AmsCommissionFee,
				SellerOrderProcessingFee:  item.SellerOrderProcessingFee,
				RawItemData:               StringPtr(string(rawItemData)),
				CreatedAt:                 time.Now(),
				UpdatedAt:                 time.Now(),
			}

			if err := s.base.DB.WithContext(ctx).Table(itemsTable).Create(&escrowItem).Error; err != nil {
				log.Printf("[ShopeeRepopulate] Failed to insert item for order %s: %v", order.OrderSN, err)
				continue
			}
			totalItems++
		}

		processedOrders++
	}

	log.Printf("[ShopeeRepopulate] Done. Processed %d orders, inserted %d items", processedOrders, totalItems)
	return processedOrders, totalItems, nil
}
