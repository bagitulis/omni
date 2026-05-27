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

// saveEscrowOrder upserts an escrow order and replaces its items atomically.
func saveEscrowOrder(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
	order shopeePkg.EscrowOrder,
	month, year int,
) (int, error) {
	rawOrderIncome, _ := json.Marshal(order.OrderIncome)
	rawBuyerPayment, _ := json.Marshal(order.BuyerPaymentInfo)
	income := order.OrderIncome
	now := time.Now()

	var savedItemCount int

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.ShopeeEscrowOrder
		err := tx.Where("tenant_id = ? AND order_sn = ?", tenantID, order.OrderSN).
			First(&existing).Error

		var escrowOrder models.ShopeeEscrowOrder
		buyerUsername := order.BuyerUsername

		if err == gorm.ErrRecordNotFound {
			escrowOrder = models.ShopeeEscrowOrder{
				ID:                   uuid.New().String(),
				TenantID:             tenantID,
				OrderSN:              order.OrderSN,
				Month:                month,
				Year:                 year,
				BuyerUserName:        &buyerUsername,
				EscrowAmount:         income.EscrowAmount,
				CommissionFee:        income.CommissionFee,
				ServiceFee:           income.ServiceFee,
				SellerProcessingFee:  income.SellerOrderProcessingFee,
				BuyerPaidShippingFee: income.BuyerPaidShippingFee,
				ActualShippingFee:    income.ActualShippingFee,
				ShopeeShippingRebate: income.ShopeeShippingRebate,
				EstimatedShippingFee: income.EstimatedShippingFee,
				BuyerTotalAmount:     income.BuyerTotalAmount,
				BuyerPaymentMethod:   &income.BuyerPaymentMethod,
				RawOrderIncome:       stringPtr(string(rawOrderIncome)),
				RawBuyerPaymentInfo:  stringPtr(string(rawBuyerPayment)),
				SyncedAt:             now,
				CreatedAt:            now,
				UpdatedAt:            now,
			}
			if err := tx.Create(&escrowOrder).Error; err != nil {
				return fmt.Errorf("create order %s: %w", order.OrderSN, err)
			}
		} else if err != nil {
			return fmt.Errorf("query order %s: %w", order.OrderSN, err)
		} else {
			existing.Month = month
			existing.Year = year
			existing.BuyerUserName = &buyerUsername
			existing.EscrowAmount = income.EscrowAmount
			existing.CommissionFee = income.CommissionFee
			existing.ServiceFee = income.ServiceFee
			existing.SellerProcessingFee = income.SellerOrderProcessingFee
			existing.BuyerPaidShippingFee = income.BuyerPaidShippingFee
			existing.ActualShippingFee = income.ActualShippingFee
			existing.ShopeeShippingRebate = income.ShopeeShippingRebate
			existing.EstimatedShippingFee = income.EstimatedShippingFee
			existing.BuyerTotalAmount = income.BuyerTotalAmount
			existing.BuyerPaymentMethod = &income.BuyerPaymentMethod
			existing.RawOrderIncome = stringPtr(string(rawOrderIncome))
			existing.RawBuyerPaymentInfo = stringPtr(string(rawBuyerPayment))
			existing.SyncedAt = now
			existing.UpdatedAt = now
			if err := tx.Save(&existing).Error; err != nil {
				return fmt.Errorf("update order %s: %w", order.OrderSN, err)
			}
			escrowOrder = existing
		}

		if err := tx.Where("tenant_id = ? AND escrow_order_id = ?", tenantID, escrowOrder.ID).
			Delete(&models.ShopeeEscrowItem{}).Error; err != nil {
			return fmt.Errorf("delete items for order %s: %w", order.OrderSN, err)
		}

		for _, item := range income.Items {
			if err := saveEscrowItemTx(ctx, tx, escrowOrder.ID, order.OrderSN, tenantID, item, month, year); err != nil {
				log.Warn().Err(err).Str("order_sn", order.OrderSN).Int64("item_id", item.ItemID).
					Msg("Failed to save shopee escrow item")
				continue
			}
			savedItemCount++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return savedItemCount, nil
}

// saveEscrowItemTx creates a single escrow item within an existing transaction.
func saveEscrowItemTx(
	_ context.Context,
	tx *gorm.DB,
	escrowOrderID, orderSN, tenantID string,
	item shopeePkg.EscrowItemData,
	month, year int,
) error {
	rawItemData, _ := json.Marshal(item)
	now := time.Now()

	escrowItem := models.ShopeeEscrowItem{
		ID:                        uuid.New().String(),
		TenantID:                  tenantID,
		EscrowOrderID:             escrowOrderID,
		OrderSN:                   &orderSN,
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
	return tx.Create(&escrowItem).Error
}
