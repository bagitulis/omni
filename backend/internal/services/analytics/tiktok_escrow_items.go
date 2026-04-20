// Package analytics provides TikTok escrow item saving logic (SKU transactions + line items)
package analytics

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// saveEscrowItemsTx saves escrow items within an existing DB transaction.
// Prefers SKU transactions (Finance API, has settlement_amount) over line items.
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
	// Build sku_id → line_item map to resolve correct seller_sku
	lineItemMap := make(map[string]tiktokPkg.TiktokOrderItem, len(lineItems))
	for _, li := range lineItems {
		lineItemMap[li.SkuID] = li
	}

	for _, skuTx := range skuTxs {
		rawSkuData, _ := json.Marshal(skuTx)
		qty, _ := strconv.Atoi(skuTx.Quantity)
		if qty == 0 {
			qty = 1
		}

		settlementAmt := ParseFloat(skuTx.SettlementAmount)
		if settlementAmt == 0 {
			settlementAmt = ParseFloat(skuTx.SkuNetPayout)
		}

		salePrice := ParseFloat(skuTx.RevenueAmount)
		if salePrice == 0 {
			salePrice = ParseFloat(skuTx.SkuSubtotalAfterDisc)
		}

		// Resolve seller_sku from order line_items (Finance API sku_name = variation name)
		sellerSku := skuTx.SkuName
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

// saveLineItemsTx saves line items within a DB transaction.
// Used as fallback when no SKU transaction data is available from Finance API.
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
			qty = 1
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
