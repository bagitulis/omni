// Package analytics provides Shopee escrow items saving logic
package analytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// saveEscrowOrder saves an escrow order and its items using upsert pattern
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

	// Prepare base record with new UUID (will be ignored if record exists)
	escrowOrder := models.ShopeeEscrowOrder{
		ID:        uuid.New().String(),
		TenantID:  s.tenantID,
		OrderSN:   order.OrderSN,
		CreatedAt: time.Now(),
	}

	// Data to update/assign (excludes ID and CreatedAt to preserve existing)
	updateData := models.ShopeeEscrowOrder{
		Month:                month,
		Year:                 year,
		BuyerUserName:        StringPtr(buyerUsername),
		EscrowAmount:         income.EscrowAmount,
		CommissionFee:        income.CommissionFee,
		ServiceFee:           income.ServiceFee,
		SellerProcessingFee:  income.SellerOrderProcessingFee,
		BuyerPaidShippingFee: income.BuyerPaidShippingFee,
		ActualShippingFee:    income.ActualShippingFee,
		ShopeeShippingRebate: income.ShopeeShippingRebate,
		EstimatedShippingFee: income.EstimatedShippingFee,
		BuyerTotalAmount:     income.BuyerTotalAmount,
		RawOrderIncome:       StringPtr(string(rawOrderIncome)),
		RawBuyerPaymentInfo:  StringPtr(string(rawBuyerPayment)),
		SyncedAt:             time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Upsert order - FirstOrCreate finds or creates, Assign updates fields
	tables := ShopeeEscrowTables()
	if err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.OrderTable)).
		Where("tenant_id = ? AND order_sn = ?", s.tenantID, order.OrderSN).
		Assign(updateData).
		FirstOrCreate(&escrowOrder).Error; err != nil {
		return 0, err
	}

	// Delete existing items
	s.base.DB.WithContext(ctx).Table(s.base.Table(tables.ItemTable)).
		Where("escrow_order_id = ?", escrowOrder.ID).
		Delete(&models.ShopeeEscrowItem{})

	// Save items
	for _, item := range income.Items {
		s.saveEscrowItem(ctx, escrowOrder.ID, order.OrderSN, item, month, year)
	}

	return len(income.Items), nil
}

// saveEscrowItem saves a single escrow item
func (s *ShopeeEscrowSyncService) saveEscrowItem(
	ctx context.Context,
	escrowOrderID, orderSN string,
	item shopeePkg.EscrowItemData,
	month, year int,
) {
	rawItemData, _ := json.Marshal(item)
	tables := ShopeeEscrowTables()

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
	s.base.DB.WithContext(ctx).Table(s.base.Table(tables.ItemTable)).Create(&escrowItem)
}
