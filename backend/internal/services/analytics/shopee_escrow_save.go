// Package analytics provides Shopee escrow items saving logic
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// saveEscrowOrder saves an escrow order and its items using atomic upsert
func (s *ShopeeEscrowSyncService) saveEscrowOrder(
	ctx context.Context,
	order shopeePkg.EscrowOrder,
	month, year int,
) (int, error) {
	// Marshal raw data
	rawOrderIncome, _ := json.Marshal(order.OrderIncome)
	rawBuyerPayment, _ := json.Marshal(order.BuyerPaymentInfo)

	// Get order income data
	income := order.OrderIncome
	buyerUsername := order.BuyerUsername
	now := time.Now()
	newID := uuid.New().String()

	tables := ShopeeEscrowTables()
	orderTable := s.base.Table(tables.OrderTable)
	itemTable := s.base.Table(tables.ItemTable)

	var escrowOrderID string

	// Use a DB transaction for atomicity (order upsert + item replace)
	err := s.base.DB.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		// Atomic upsert using raw SQL with ON CONFLICT + RETURNING id
		row := dbTx.Raw(fmt.Sprintf(`
			INSERT INTO %s (
				id, tenant_id, order_sn, month, year,
				buyer_user_name, escrow_amount, commission_fee, service_fee,
				seller_processing_fee, buyer_paid_shipping_fee, actual_shipping_fee,
				shopee_shipping_rebate, estimated_shipping_fee, buyer_total_amount,
				raw_order_income, raw_buyer_payment_info,
				synced_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (tenant_id, order_sn) DO UPDATE SET
				month = EXCLUDED.month,
				year = EXCLUDED.year,
				buyer_user_name = EXCLUDED.buyer_user_name,
				escrow_amount = EXCLUDED.escrow_amount,
				commission_fee = EXCLUDED.commission_fee,
				service_fee = EXCLUDED.service_fee,
				seller_processing_fee = EXCLUDED.seller_processing_fee,
				buyer_paid_shipping_fee = EXCLUDED.buyer_paid_shipping_fee,
				actual_shipping_fee = EXCLUDED.actual_shipping_fee,
				shopee_shipping_rebate = EXCLUDED.shopee_shipping_rebate,
				estimated_shipping_fee = EXCLUDED.estimated_shipping_fee,
				buyer_total_amount = EXCLUDED.buyer_total_amount,
				raw_order_income = EXCLUDED.raw_order_income,
				raw_buyer_payment_info = EXCLUDED.raw_buyer_payment_info,
				synced_at = EXCLUDED.synced_at,
				updated_at = EXCLUDED.updated_at
			RETURNING id
		`, orderTable),
			newID, s.tenantID, order.OrderSN, month, year,
			buyerUsername, income.EscrowAmount, income.CommissionFee, income.ServiceFee,
			income.SellerOrderProcessingFee, income.BuyerPaidShippingFee, income.ActualShippingFee,
			income.ShopeeShippingRebate, income.EstimatedShippingFee, income.BuyerTotalAmount,
			string(rawOrderIncome), string(rawBuyerPayment),
			now, now, now,
		).Row()

		if err := row.Scan(&escrowOrderID); err != nil {
			return fmt.Errorf("upsert order %s: %w", order.OrderSN, err)
		}

		// Delete existing items for this order
		if err := dbTx.Exec(
			fmt.Sprintf("DELETE FROM %s WHERE tenant_id = ? AND escrow_order_id = ?", itemTable),
			s.tenantID, escrowOrderID,
		).Error; err != nil {
			return fmt.Errorf("delete items for order %s: %w", order.OrderSN, err)
		}

		// Save items within the same transaction
		for _, item := range income.Items {
			s.saveEscrowItemTx(ctx, dbTx, escrowOrderID, order.OrderSN, item, month, year, itemTable)
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return len(income.Items), nil
}

// saveEscrowItemTx saves a single escrow item within an existing DB transaction
func (s *ShopeeEscrowSyncService) saveEscrowItemTx(
	ctx context.Context,
	dbTx *gorm.DB,
	escrowOrderID, orderSN string,
	item shopeePkg.EscrowItemData,
	month, year int,
	itemTable string,
) {
	rawItemData, _ := json.Marshal(item)

	escrowItem := models.ShopeeEscrowItem{
		ID:                        uuid.New().String(),
		TenantID:                  s.tenantID,
		EscrowOrderID:             escrowOrderID,
		OrderSN:                   StringPtr(orderSN),
		Month:                     IntPtr(month),
		Year:                      IntPtr(year),
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
	if err := dbTx.Table(itemTable).Create(&escrowItem).Error; err != nil {
		log.Warn().Str("order_sn", orderSN).Int64("item_id", item.ItemID).
			Err(err).Msg("[ShopeeEscrowSync] Failed to save item")
	}
}
