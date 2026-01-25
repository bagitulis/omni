// Package analytics provides TikTok escrow order/item saving logic
package analytics

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// saveEscrowOrder saves order and item data to database using upsert pattern
func (s *TiktokEscrowSyncService) saveEscrowOrder(
	ctx context.Context,
	order tiktokPkg.TiktokOrder,
	tx *tiktokPkg.OrderTransactionResponse,
	month, year int,
) (int, error) {
	// Marshal raw data
	rawOrderData, _ := json.Marshal(order)
	rawTxData, _ := json.Marshal(tx.Data)

	// Parse amounts
	shippingFee := ParseFloat(order.PaymentInfo.ShippingFee)
	shippingPlatformDisc := ParseFloat(order.PaymentInfo.ShippingFeePlatform)
	platformDiscount := ParseFloat(order.PaymentInfo.PlatformDiscount)
	sellerDiscount := ParseFloat(order.PaymentInfo.SellerDiscount)
	totalAmount := ParseFloat(order.PaymentInfo.TotalAmount)
	subTotal := ParseFloat(order.PaymentInfo.SubTotal)
	totalSettlementAmount := ParseFloat(tx.Data.SettlementAmount)

	orderDate := time.Unix(order.CreateTime, 0)
	orderStatus := order.Status

	// Prepare base record with new UUID (will be ignored if record exists)
	escrowOrder := models.TiktokEscrowOrder{
		ID:        uuid.New().String(),
		TenantID:  s.tenantID,
		OrderID:   order.ID,
		CreatedAt: time.Now(),
	}

	// Data to update/assign (excludes ID and CreatedAt to preserve existing)
	updateData := models.TiktokEscrowOrder{
		Month:                       month,
		Year:                        year,
		OrderStatus:                 &orderStatus,
		OrderDate:                   &orderDate,
		ProductRevenue:              subTotal,
		BuyerTotalAmount:            totalAmount,
		TotalSettlementAmount:       totalSettlementAmount,
		PlatformCommission:          platformDiscount,
		ShippingFeeCustomerPaid:     shippingFee,
		ShippingFeePlatformDiscount: shippingPlatformDisc,
		SellerShippingDiscount:      sellerDiscount,
		Currency:                    tx.Data.Currency,
		RawOrderData:                StringPtr(string(rawOrderData)),
		RawTransactionData:          StringPtr(string(rawTxData)),
		SyncedAt:                    time.Now(),
		UpdatedAt:                   time.Now(),
	}

	// Upsert order
	tables := TiktokEscrowTables()
	if err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.OrderTable)).
		Where("tenant_id = ? AND order_id = ?", s.tenantID, order.ID).
		Assign(updateData).
		FirstOrCreate(&escrowOrder).Error; err != nil {
		return 0, err
	}

	// Delete existing items
	s.base.DB.WithContext(ctx).Table(s.base.Table(tables.ItemTable)).
		Where("escrow_order_id = ?", escrowOrder.ID).
		Delete(&models.TiktokEscrowItem{})

	// Save items - prefer SkuTransactions, fallback to LineItems
	itemCount := s.saveEscrowItems(ctx, escrowOrder.ID, order, tx)

	return itemCount, nil
}

// saveEscrowItems saves escrow items from SKU transactions or line items
func (s *TiktokEscrowSyncService) saveEscrowItems(
	ctx context.Context,
	escrowOrderID string,
	order tiktokPkg.TiktokOrder,
	tx *tiktokPkg.OrderTransactionResponse,
) int {
	tables := TiktokEscrowTables()

	if len(tx.Data.SkuTransactions) > 0 {
		return s.saveSkuTransactions(ctx, escrowOrderID, order.ID, tx.Data.SkuTransactions, tables)
	}

	if len(order.LineItems) > 0 {
		log.Printf("[TiktokEscrowSync] Using order line_items for order %s", order.ID)
		return s.saveLineItems(ctx, escrowOrderID, order.ID, order.LineItems, tables)
	}

	return 0
}

// saveSkuTransactions saves SKU transactions from Finance API
func (s *TiktokEscrowSyncService) saveSkuTransactions(
	ctx context.Context,
	escrowOrderID, orderID string,
	skuTxs []tiktokPkg.SkuTransaction,
	tables EscrowSyncTables,
) int {
	for _, skuTx := range skuTxs {
		rawSkuData, _ := json.Marshal(skuTx)
		qty := skuTx.Quantity
		if qty == 0 {
			qty = 1 // Default to 1 (matches Node.js logic)
		}
		escrowItem := models.TiktokEscrowItem{
			ID:                          uuid.New().String(),
			TenantID:                    s.tenantID,
			EscrowOrderID:               escrowOrderID,
			OrderID:                     orderID,
			SkuID:                       StringPtr(skuTx.SkuID),
			ProductName:                 StringPtr(skuTx.ProductName),
			Quantity:                    qty,
			OriginalPrice:               ParseFloat(skuTx.SkuSubtotalBeforeDisc),
			PlatformDiscount:            ParseFloat(skuTx.SkuPlatformDiscount),
			SellerDiscount:              ParseFloat(skuTx.SkuSellerDiscount),
			SubtotalAfterSellerDiscount: ParseFloat(skuTx.SkuSubtotalAfterDisc),
			TransactionFeeItem:          ParseFloat(skuTx.TransactionFee),
			Commission:                  ParseFloat(skuTx.ReferralFee),
			SettlementAmount:            ParseFloat(skuTx.SkuNetPayout),
			RawItemData:                 StringPtr(string(rawSkuData)),
			SyncedAt:                    time.Now(),
			CreatedAt:                   time.Now(),
			UpdatedAt:                   time.Now(),
		}
		s.base.DB.WithContext(ctx).Table(s.base.Table(tables.ItemTable)).Create(&escrowItem)
	}
	return len(skuTxs)
}

// saveLineItems saves line items from Order API
func (s *TiktokEscrowSyncService) saveLineItems(
	ctx context.Context,
	escrowOrderID, orderID string,
	lineItems []tiktokPkg.TiktokOrderItem,
	tables EscrowSyncTables,
) int {
	for _, item := range lineItems {
		rawItemData, _ := json.Marshal(item)
		qty := item.Quantity
		if qty == 0 {
			qty = 1 // Default to 1 (matches Node.js logic)
		}
		escrowItem := models.TiktokEscrowItem{
			ID:               uuid.New().String(),
			TenantID:         s.tenantID,
			EscrowOrderID:    escrowOrderID,
			OrderID:          orderID,
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
		s.base.DB.WithContext(ctx).Table(s.base.Table(tables.ItemTable)).Create(&escrowItem)
	}
	return len(lineItems)
}
