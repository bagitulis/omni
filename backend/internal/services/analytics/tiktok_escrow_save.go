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

// saveEscrowOrder saves order and item data with atomic upsert.
// It applies statement-month attribution and normalizes negative shipping fees.
func (s *TiktokEscrowSyncService) saveEscrowOrder(
	ctx context.Context, order tiktokPkg.TiktokOrder,
	tx *tiktokPkg.OrderTransactionResponse, month, year int,
) (int, error) {
	if tx == nil {
		return 0, nil
	}

	rawOrderData, _ := json.Marshal(order)
	rawTxData, _ := json.Marshal(tx.Data)

	shippingFee := parseFloat(order.PaymentInfo.ShippingFee)
	shippingPlatformDisc := parseFloat(order.PaymentInfo.ShippingFeePlatform)
	platformDiscount := parseFloat(order.PaymentInfo.PlatformDiscount)
	sellerDiscount := parseFloat(order.PaymentInfo.SellerDiscount)
	totalAmount := parseFloat(order.PaymentInfo.TotalAmount)
	subTotal := parseFloat(order.PaymentInfo.SubTotal)
	totalSettlementAmount := parseFloat(tx.Data.SettlementAmount)
	shippingFeeActual := normalizeShippingFee(parseFloat(tx.Data.ShippingCostAmount))

	if subTotal == 0 && len(tx.Data.SkuTransactions) > 0 {
		for _, sku := range tx.Data.SkuTransactions {
			subTotal += parseFloat(sku.RevenueAmount)
		}
	}
	if totalAmount == 0 && subTotal > 0 {
		totalAmount = subTotal + shippingFee
	}

	orderDate := parseTiktokTimestamp(order.CreateTime)
	orderStatus := order.Status
	now := time.Now()
	newID := uuid.New().String()

	itemTable := "tiktok_escrow_items"
	orderTable := "tiktok_escrow_orders"

	var escrowOrderID string
	var saved bool

	err := s.tenantDB.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		var statementTime *time.Time
		var transactionID string
		var aggSettlement, aggShipCustomer, aggShipActual, aggShipPlatformDisc float64
		var matchCount int

		for _, st := range tx.Data.StatementTransactions {
			t := parseTiktokTimestamp(st.StatementTime)
			if int(t.Month()) == month && t.Year() == year {
				matchCount++
				if matchCount == 1 {
					statementTime = &t
					transactionID = st.StatementID
				}
				if st.SettlementAmount != "" && st.SettlementAmount != "0" {
					aggSettlement += parseFloat(st.SettlementAmount)
				} else if st.Amount != "" && st.Amount != "0" {
					aggSettlement += parseFloat(st.Amount)
				}
				if st.CustomerPaidShippingFeeAmount != "" {
					aggShipCustomer += parseFloat(st.CustomerPaidShippingFeeAmount)
				}
				if st.PlatformShippingFeeDiscountAmount != "" {
					aggShipPlatformDisc += parseFloat(st.PlatformShippingFeeDiscountAmount)
				}
				if st.ActualShippingFeeAmount != "" {
					shipVal := normalizeShippingFee(parseFloat(st.ActualShippingFeeAmount))
					aggShipActual += shipVal
				}
			}
		}

		if matchCount > 0 {
			totalSettlementAmount = aggSettlement
			shippingFee = aggShipCustomer
			shippingPlatformDisc = aggShipPlatformDisc
			shippingFeeActual = aggShipActual
		}

		if statementTime == nil && len(tx.Data.StatementTransactions) > 0 {
			firstST := tx.Data.StatementTransactions[0]
			stmtT := parseTiktokTimestamp(firstST.StatementTime)
			if stmtT.Year() >= 2024 && (int(stmtT.Month()) != month || stmtT.Year() != year) {
				log.Debug().Str("order_id", order.ID).
					Int("target_month", month).Int("settlement_month", int(stmtT.Month())).
					Msg("[TiktokEscrowSync] Settlement in different month, skipping")
				return nil
			}
			createT := parseTiktokTimestamp(order.CreateTime)
			if int(createT.Month()) == month && createT.Year() == year {
				refT := stmtT
				statementTime = &refT
				transactionID = firstST.StatementID
				for _, st := range tx.Data.StatementTransactions {
					if st.SettlementAmount != "" && st.SettlementAmount != "0" {
						aggSettlement += parseFloat(st.SettlementAmount)
					} else if st.Amount != "" && st.Amount != "0" {
						aggSettlement += parseFloat(st.Amount)
					}
					if st.CustomerPaidShippingFeeAmount != "" {
						aggShipCustomer += parseFloat(st.CustomerPaidShippingFeeAmount)
					}
					if st.PlatformShippingFeeDiscountAmount != "" {
						aggShipPlatformDisc += parseFloat(st.PlatformShippingFeeDiscountAmount)
					}
					if st.ActualShippingFeeAmount != "" {
						shipVal := normalizeShippingFee(parseFloat(st.ActualShippingFeeAmount))
						aggShipActual += shipVal
					}
				}
				totalSettlementAmount = aggSettlement
				shippingFee = aggShipCustomer
				shippingPlatformDisc = aggShipPlatformDisc
				shippingFeeActual = aggShipActual
			}
		}

		if statementTime == nil {
			updateT := parseTiktokTimestamp(order.UpdateTime)
			topLevelSettlement := tx.Data.SettlementAmount
			if int(updateT.Month()) == month && updateT.Year() == year &&
				topLevelSettlement != "" && topLevelSettlement != "0" {
				statementTime = &updateT
				totalSettlementAmount = parseFloat(topLevelSettlement)
				if tx.Data.ShippingCostAmount != "" {
					shippingFeeActual = normalizeShippingFee(parseFloat(tx.Data.ShippingCostAmount))
				}
			} else {
				log.Debug().Str("order_id", order.ID).
					Int("month", month).Int("update_month", int(updateT.Month())).
					Msg("[TiktokEscrowSync] Skipping order - no settlement in target month")
				return nil
			}
		}

		buyerName := ""
		if order.RecipientAddress.Name != "" {
			buyerName = order.RecipientAddress.Name
		}

		currency := tx.Data.Currency
		if currency == "" {
			for _, st := range tx.Data.StatementTransactions {
				if st.Currency != "" {
					currency = st.Currency
					break
				}
			}
		}

		orderModel := models.TiktokEscrowOrder{
			ID:                          newID,
			TenantID:                    s.tenantID,
			OrderID:                     order.ID,
			Month:                       month,
			Year:                        year,
			OrderStatus:                 &orderStatus,
			OrderDate:                   &orderDate,
			BuyerName:                   &buyerName,
			TransactionID:               &transactionID,
			StatementTime:               statementTime,
			ProductRevenue:              subTotal,
			BuyerTotalAmount:            totalAmount,
			TotalSettlementAmount:       totalSettlementAmount,
			PlatformCommission:          platformDiscount,
			ShippingFeeCustomerPaid:     shippingFee,
			ShippingFeeActual:           shippingFeeActual,
			ShippingFeePlatformDiscount: shippingPlatformDisc,
			SellerShippingDiscount:      sellerDiscount,
			Currency:                    currency,
			RawOrderData:                stringPtr(string(rawOrderData)),
			RawTransactionData:          stringPtr(string(rawTxData)),
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		}
		if err := dbTx.Table(orderTable).Create(&orderModel).Error; err != nil {
			return fmt.Errorf("create order %s: %w", order.ID, err)
		}
		escrowOrderID = newID
		saved = true
		if err := dbTx.Exec(
			fmt.Sprintf("DELETE FROM %s WHERE tenant_id = ? AND escrow_order_id = ?", itemTable),
			s.tenantID, escrowOrderID,
		).Error; err != nil {
			return fmt.Errorf("delete items for order %s: %w", order.ID, err)
		}
		return s.saveEscrowItemsTx(ctx, dbTx, escrowOrderID, order, tx, itemTable)
	})

	if err != nil {
		return 0, err
	}
	if !saved {
		return 0, nil
	}
	return countEscrowItems(order, tx), nil
}

// saveEscrowItemsTx saves escrow items within an existing DB transaction.
func (s *TiktokEscrowSyncService) saveEscrowItemsTx(
	ctx context.Context, dbTx *gorm.DB,
	escrowOrderID string, order tiktokPkg.TiktokOrder,
	tx *tiktokPkg.OrderTransactionResponse, itemTable string,
) error {
	if len(tx.Data.SkuTransactions) > 0 {
		return s.saveSkuTransactionsTx(ctx, dbTx, escrowOrderID, order.ID,
			tx.Data.SkuTransactions, order.LineItems, itemTable)
	}
	if len(order.LineItems) > 0 {
		log.Info().Msgf("[TiktokEscrowSync] Using order line_items for order %s (no sku_transactions)", order.ID)
		return s.saveLineItemsTx(ctx, dbTx, escrowOrderID, order.ID, order.LineItems, itemTable)
	}
	return nil
}

// saveSkuTransactionsTx saves SKU transactions as escrow items.
func (s *TiktokEscrowSyncService) saveSkuTransactionsTx(
	_ context.Context, dbTx *gorm.DB,
	escrowOrderID, orderID string,
	skuTxs []tiktokPkg.SkuTransaction,
	lineItems []tiktokPkg.TiktokOrderItem,
	itemTable string,
) error {
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

		settlementAmt := parseFloat(skuTx.SettlementAmount)
		if settlementAmt == 0 {
			settlementAmt = parseFloat(skuTx.SkuNetPayout)
		}

		salePrice := parseFloat(skuTx.RevenueAmount)
		if salePrice == 0 {
			salePrice = parseFloat(skuTx.SkuSubtotalAfterDisc)
		}

		originalPrice := parseFloat(skuTx.SkuSubtotalBeforeDisc)
		if originalPrice == 0 {
			originalPrice = salePrice
		}

		sellerSku := skuTx.SkuName
		var productID string
		if li, ok := lineItemMap[skuTx.SkuID]; ok && li.SellerSku != "" {
			sellerSku = li.SellerSku
			productID = li.ProductID
		}

		productName := skuTx.ProductName
		if productName == "" {
			if li, ok := lineItemMap[skuTx.SkuID]; ok {
				productName = li.ProductName
			}
		}

		escrowItem := models.TiktokEscrowItem{
			ID:                          uuid.New().String(),
			TenantID:                    s.tenantID,
			EscrowOrderID:               escrowOrderID,
			OrderID:                     orderID,
			ProductID:                   stringPtr(productID),
			SkuID:                       stringPtr(skuTx.SkuID),
			SellerSku:                   stringPtr(sellerSku),
			ProductName:                 stringPtr(productName),
			Quantity:                    qty,
			OriginalPrice:               originalPrice,
			SalePrice:                   salePrice,
			PlatformDiscount:            parseFloat(skuTx.SkuPlatformDiscount),
			SellerDiscount:              parseFloat(skuTx.SkuSellerDiscount),
			SubtotalAfterSellerDiscount: parseFloat(skuTx.SkuSubtotalAfterDisc),
			TransactionFeeItem:          parseFloat(skuTx.TransactionFee),
			Commission:                  parseFloat(skuTx.ReferralFee),
			SettlementAmount:            settlementAmt,
			RawItemData:                 stringPtr(string(rawSkuData)),
			SyncedAt:                    time.Now(),
			CreatedAt:                   time.Now(),
			UpdatedAt:                   time.Now(),
		}
		if err := dbTx.Table(itemTable).Create(&escrowItem).Error; err != nil {
			return fmt.Errorf("save sku_tx item for order %s sku %s: %w", orderID, skuTx.SkuID, err)
		}
	}
	return nil
}

// saveLineItemsTx saves line items as fallback when no SKU transactions exist.
func (s *TiktokEscrowSyncService) saveLineItemsTx(
	_ context.Context, dbTx *gorm.DB,
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
			ProductID:        stringPtr(item.ProductID),
			ProductName:      stringPtr(item.ProductName),
			SkuID:            stringPtr(item.SkuID),
			SellerSku:        stringPtr(item.SellerSku),
			Quantity:         qty,
			OriginalPrice:    parseFloat(item.OriginalPrice),
			SalePrice:        parseFloat(item.SalePrice),
			PlatformDiscount: parseFloat(item.PlatformDiscount),
			SellerDiscount:   parseFloat(item.SellerDiscount),
			RawItemData:      stringPtr(string(rawItemData)),
			SyncedAt:         time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := dbTx.Table(itemTable).Create(&escrowItem).Error; err != nil {
			return fmt.Errorf("save line_item for order %s sku %s: %w", orderID, item.SellerSku, err)
		}
	}
	return nil
}

// countEscrowItems returns the expected item count.
func countEscrowItems(order tiktokPkg.TiktokOrder, tx *tiktokPkg.OrderTransactionResponse) int {
	if tx != nil && len(tx.Data.SkuTransactions) > 0 {
		return len(tx.Data.SkuTransactions)
	}
	return len(order.LineItems)
}
