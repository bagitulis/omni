package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

// RepopulateItems restores Shopee escrow item rows from raw_order_income JSON.
func (s *ShopeeAnalyticsService) RepopulateItems(ctx context.Context, tenantID string, period string) error {
	month, year, err := parseRepopulatePeriod(period)
	if err != nil {
		return err
	}
	if err := validateMonthYear(month, year); err != nil {
		return err
	}

	var orders []models.ShopeeEscrowOrder
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ? AND raw_order_income IS NOT NULL AND raw_order_income != ''", tenantID, month, year).
		Find(&orders).Error; err != nil {
		return fmt.Errorf("fetch Shopee escrow orders for repopulate: %w", err)
	}

	return s.tenantDB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, order := range orders {
			items, err := parseShopeeRawItems(order)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				continue
			}

			if err := tx.Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).
				Delete(&models.ShopeeEscrowItem{}).Error; err != nil {
				return fmt.Errorf("delete Shopee escrow items for order %s: %w", order.OrderSN, err)
			}

			for _, item := range items {
				if err := createShopeeEscrowItemFromRaw(tx, tenantID, order, item, month, year); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func parseRepopulatePeriod(period string) (int, int, error) {
	parts := strings.Split(period, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid period %q (expected YYYY-MM)", period)
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid period year %q: %w", parts[0], err)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid period month %q: %w", parts[1], err)
	}
	return month, year, nil
}

func parseShopeeRawItems(order models.ShopeeEscrowOrder) ([]shopeePkg.EscrowItemData, error) {
	if order.RawOrderIncome == nil || strings.TrimSpace(*order.RawOrderIncome) == "" {
		return nil, nil
	}

	var income shopeePkg.EscrowOrderData
	if err := json.Unmarshal([]byte(*order.RawOrderIncome), &income); err != nil {
		return nil, fmt.Errorf("parse raw_order_income for Shopee order %s: malformed JSON: %w", order.OrderSN, err)
	}
	return income.Items, nil
}

func createShopeeEscrowItemFromRaw(
	tx *gorm.DB,
	tenantID string,
	order models.ShopeeEscrowOrder,
	item shopeePkg.EscrowItemData,
	month, year int,
) error {
	rawItemData, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal Shopee raw item for order %s item %d: %w", order.OrderSN, item.ItemID, err)
	}

	now := time.Now()
	escrowItem := models.ShopeeEscrowItem{
		ID:                        uuid.New().String(),
		TenantID:                  tenantID,
		EscrowOrderID:             order.ID,
		OrderSN:                   &order.OrderSN,
		Month:                     &month,
		Year:                      &year,
		ItemID:                    &item.ItemID,
		ModelID:                   &item.ModelID,
		Sku:                       &item.ItemSKU,
		ModelSku:                  &item.ModelSKU,
		ItemName:                  &item.ItemName,
		ModelName:                 &item.ModelName,
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
		RawItemData:               stringPtr(string(rawItemData)),
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}

	if err := tx.Create(&escrowItem).Error; err != nil {
		return fmt.Errorf("create Shopee escrow item for order %s item %d: %w", order.OrderSN, item.ItemID, err)
	}
	return nil
}
