// Package analytics provides TikTok escrow order/item saving logic
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// saveEscrowOrder saves order and item data to database using atomic upsert
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
	shippingFeeActual := ParseFloat(tx.Data.ShippingCostAmount)

	orderDate := time.Unix(order.CreateTime, 0)
	orderStatus := order.Status
	now := time.Now()
	newID := uuid.New().String()

	tables := TiktokEscrowTables()
	orderTable := s.base.Table(tables.OrderTable)
	itemTable := s.base.Table(tables.ItemTable)

	var escrowOrderID string

	// Use a DB transaction for atomicity (order upsert + item replace)
	err := s.base.DB.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		// Atomic upsert using raw SQL with ON CONFLICT + RETURNING id
		row := dbTx.Raw(fmt.Sprintf(`
			INSERT INTO %s (
				id, tenant_id, order_id, month, year, order_status, order_date,
				product_revenue, buyer_total_amount, total_settlement_amount,
				platform_commission, shipping_fee_customer_paid, shipping_fee_actual,
				shipping_fee_platform_discount,
				seller_shipping_discount, currency,
				raw_order_data, raw_transaction_data,
				synced_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (tenant_id, order_id) DO UPDATE SET
				month = EXCLUDED.month,
				year = EXCLUDED.year,
				order_status = EXCLUDED.order_status,
				order_date = EXCLUDED.order_date,
				product_revenue = EXCLUDED.product_revenue,
				buyer_total_amount = EXCLUDED.buyer_total_amount,
				total_settlement_amount = EXCLUDED.total_settlement_amount,
				platform_commission = EXCLUDED.platform_commission,
				shipping_fee_customer_paid = EXCLUDED.shipping_fee_customer_paid,
				shipping_fee_actual = EXCLUDED.shipping_fee_actual,
				shipping_fee_platform_discount = EXCLUDED.shipping_fee_platform_discount,
				seller_shipping_discount = EXCLUDED.seller_shipping_discount,
				currency = EXCLUDED.currency,
				raw_order_data = EXCLUDED.raw_order_data,
				raw_transaction_data = EXCLUDED.raw_transaction_data,
				synced_at = EXCLUDED.synced_at,
				updated_at = EXCLUDED.updated_at
			RETURNING id
		`, orderTable),
			newID, s.tenantID, order.ID, month, year, orderStatus, orderDate,
			subTotal, totalAmount, totalSettlementAmount,
			platformDiscount, shippingFee, shippingFeeActual,
			shippingPlatformDisc,
			sellerDiscount, tx.Data.Currency,
			string(rawOrderData), string(rawTxData),
			now, now, now,
		).Row()

		if err := row.Scan(&escrowOrderID); err != nil {
			return fmt.Errorf("upsert order %s: %w", order.ID, err)
		}

		// Delete existing items for this order
		if err := dbTx.Exec(
			fmt.Sprintf("DELETE FROM %s WHERE tenant_id = ? AND escrow_order_id = ?", itemTable),
			s.tenantID, escrowOrderID,
		).Error; err != nil {
			return fmt.Errorf("delete items for order %s: %w", order.ID, err)
		}

		// Save items within the same transaction
		return s.saveEscrowItemsTx(ctx, dbTx, escrowOrderID, order, tx, itemTable)
	})

	if err != nil {
		return 0, err
	}

	return countEscrowItems(order, tx), nil
}

// countEscrowItems returns the expected item count
func countEscrowItems(order tiktokPkg.TiktokOrder, tx *tiktokPkg.OrderTransactionResponse) int {
	if len(tx.Data.SkuTransactions) > 0 {
		return len(tx.Data.SkuTransactions)
	}
	return len(order.LineItems)
}

// saveEscrowItemsTx saves escrow items within an existing DB transaction
func (s *TiktokEscrowSyncService) saveEscrowItemsTx(
	ctx context.Context,
	dbTx *gorm.DB,
	escrowOrderID string,
	order tiktokPkg.TiktokOrder,
	tx *tiktokPkg.OrderTransactionResponse,
	itemTable string,
) error {
	if len(tx.Data.SkuTransactions) > 0 {
		return s.saveSkuTransactionsTx(ctx, dbTx, escrowOrderID, order.ID, tx.Data.SkuTransactions, order.LineItems, itemTable)
	}

	if len(order.LineItems) > 0 {
		log.Info().Msgf("[TiktokEscrowSync] Using order line_items for order %s (no sku_transactions)", order.ID)
		return s.saveLineItemsTx(ctx, dbTx, escrowOrderID, order.ID, order.LineItems, itemTable)
	}

	return nil
}

// saveSkuTransactionsTx saves SKU transactions within a DB transaction.
// lineItems from order.LineItems are used to resolve the correct seller_sku,
// since the Finance API's sku_name is the variant name, NOT the seller SKU.
func (s *TiktokEscrowSyncService) saveSkuTransactionsTx(
	ctx context.Context,
	dbTx *gorm.DB,
	escrowOrderID, orderID string,
	skuTxs []tiktokPkg.SkuTransaction,
	lineItems []tiktokPkg.TiktokOrderItem,
	itemTable string,
) error {
	// Build sku_id -> line_item map to resolve correct seller_sku
	lineItemMap := make(map[string]tiktokPkg.TiktokOrderItem, len(lineItems))
	for _, li := range lineItems {
		lineItemMap[li.SkuID] = li
	}

	for _, skuTx := range skuTxs {
		rawSkuData, _ := json.Marshal(skuTx)
		qty, _ := strconv.Atoi(skuTx.Quantity)
		if qty == 0 {
			qty = 1 // Default to 1 (matches Node.js logic)
		}

		// Parse settlement amount - prefer SettlementAmount, fallback to SkuNetPayout
		settlementAmt := ParseFloat(skuTx.SettlementAmount)
		if settlementAmt == 0 {
			settlementAmt = ParseFloat(skuTx.SkuNetPayout)
		}

		// Parse sale price from RevenueAmount (matches Node.js: sale_price = revenue_amount)
		salePrice := ParseFloat(skuTx.RevenueAmount)
		if salePrice == 0 {
			salePrice = ParseFloat(skuTx.SkuSubtotalAfterDisc)
		}

		// Resolve seller_sku: use order line_items (has correct seller_sku),
		// NOT sku_name from Finance API (which is the variation name like "Hitam")
		sellerSku := skuTx.SkuName // fallback to sku_name if line_item not found
		var productID string
		if li, ok := lineItemMap[skuTx.SkuID]; ok && li.SellerSku != "" {
			sellerSku = li.SellerSku
			productID = li.ProductID
		}

		escrowItem := models.TiktokEscrowItem{
			ID:                          uuid.New().String(),
			TenantID:                    s.tenantID,
			EscrowOrderID:               escrowOrderID,
			OrderID:                     orderID,
			ProductID:                   StringPtr(productID),
			SkuID:                       StringPtr(skuTx.SkuID),
			SellerSku:                   StringPtr(sellerSku),
			ProductName:                 StringPtr(skuTx.ProductName),
			Quantity:                    qty,
			OriginalPrice:               ParseFloat(skuTx.SkuSubtotalBeforeDisc),
			SalePrice:                   salePrice,
			PlatformDiscount:            ParseFloat(skuTx.SkuPlatformDiscount),
			SellerDiscount:              ParseFloat(skuTx.SkuSellerDiscount),
			SubtotalAfterSellerDiscount: ParseFloat(skuTx.SkuSubtotalAfterDisc),
			TransactionFeeItem:          ParseFloat(skuTx.TransactionFee),
			Commission:                  ParseFloat(skuTx.ReferralFee),
			SettlementAmount:            settlementAmt,
			RawItemData:                 StringPtr(string(rawSkuData)),
			SyncedAt:                    time.Now(),
			CreatedAt:                   time.Now(),
			UpdatedAt:                   time.Now(),
		}
		if err := dbTx.Table(itemTable).Create(&escrowItem).Error; err != nil {
			log.Warn().Str("order_id", orderID).Str("sku_id", skuTx.SkuID).
				Err(err).Msg("[TiktokEscrowSync] Failed to save sku_tx item")
		}
	}
	return nil
}

// saveLineItemsTx saves line items within a DB transaction
func (s *TiktokEscrowSyncService) saveLineItemsTx(
	ctx context.Context,
	dbTx *gorm.DB,
	escrowOrderID, orderID string,
	lineItems []tiktokPkg.TiktokOrderItem,
	itemTable string,
) error {
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
		if err := dbTx.Table(itemTable).Create(&escrowItem).Error; err != nil {
			log.Warn().Str("order_id", orderID).Str("seller_sku", item.SellerSku).
				Err(err).Msg("[TiktokEscrowSync] Failed to save line_item")
		}
	}
	return nil
}
