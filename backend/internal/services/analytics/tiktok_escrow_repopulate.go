package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// Internal types for parsing stored raw_order_data JSON.
// These mirror tiktokPkg.TiktokOrder / TiktokOrderItem structures
// so that repopulation does not depend on SDK types that may drift.
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

// RepopulateItems restores TikTok escrow item rows from raw_order_data JSON.
// Primary source is raw_order_data.line_items — the line_items array stored
// during TikTok sync (Task 5) from GetOrderDetail / SearchOrders responses.
//
// It follows the same transactional pattern as Shopee RepopulateItems:
// delete-all-then-recreate per order within a single transaction, aborting
// the whole batch on the first malformed raw JSON.
func (s *TiktokAnalyticsService) RepopulateItems(ctx context.Context, tenantID string, period string) error {
	month, year, err := parseRepopulatePeriod(period)
	if err != nil {
		return err
	}
	if err := validateMonthYear(month, year); err != nil {
		return err
	}

	var orders []models.TiktokEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ? AND raw_order_data IS NOT NULL AND raw_order_data != ''",
			tenantID, month, year).
		Find(&orders).Error; err != nil {
		return fmt.Errorf("fetch TikTok escrow orders for repopulate: %w", err)
	}

	return s.tenantDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			items, err := parseTiktokRawLineItems(order)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				continue
			}

			// Delete existing items for this order before recreating
			if err := tx.Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).
				Delete(&models.TiktokEscrowItem{}).Error; err != nil {
				return fmt.Errorf("delete TikTok escrow items for order %s: %w", order.OrderID, err)
			}

			for i := range items {
				items[i].TenantID = tenantID
				items[i].EscrowOrderID = order.ID
				items[i].OrderID = order.OrderID
				if err := tx.Create(&items[i]).Error; err != nil {
					return fmt.Errorf("create TikTok escrow item for order %s: %w", order.OrderID, err)
				}
			}
		}
		return nil
	})
}

// parseTiktokRawLineItems extracts item models from a TikTok order's raw_order_data.
// Returns nil slice (not error) when raw JSON has no line_items — this is a no-op case,
// not a malformed-data case. Malformed JSON returns an explicit, actionable error.
func parseTiktokRawLineItems(order models.TiktokEscrowOrder) ([]models.TiktokEscrowItem, error) {
	if order.RawOrderData == nil || *order.RawOrderData == "" {
		return nil, nil
	}

	var rawOrder tiktokRawOrderData
	if err := json.Unmarshal([]byte(*order.RawOrderData), &rawOrder); err != nil {
		return nil, fmt.Errorf("parse raw_order_data for TikTok order %s: malformed JSON: %w", order.OrderID, err)
	}

	if len(rawOrder.LineItems) == 0 {
		return nil, nil
	}

	now := time.Now()
	items := make([]models.TiktokEscrowItem, 0, len(rawOrder.LineItems))
	for _, li := range rawOrder.LineItems {
		qty := li.Quantity
		if qty == 0 {
			qty = 1
		}

		rawItemData, _ := json.Marshal(li)

		item := models.TiktokEscrowItem{
			ID:               uuid.New().String(),
			ProductID:        stringPtr(li.ProductID),
			ProductName:      stringPtr(li.ProductName),
			SkuID:            stringPtr(li.SkuID),
			SellerSku:        stringPtr(li.SellerSku),
			Quantity:         qty,
			OriginalPrice:    parseFloat(li.OriginalPrice),
			SalePrice:        parseFloat(li.SalePrice),
			PlatformDiscount: parseFloat(li.PlatformDiscount),
			SellerDiscount:   parseFloat(li.SellerDiscount),
			RawItemData:      stringPtr(string(rawItemData)),
			SyncedAt:         now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		items = append(items, item)
	}
	return items, nil
}
