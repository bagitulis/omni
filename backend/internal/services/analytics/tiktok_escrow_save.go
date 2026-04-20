// Package analytics provides TikTok escrow order/item saving logic
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	var saved bool

	// Use a DB transaction for atomicity (order upsert + item replace)
	err := s.base.DB.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		// Aggregate ALL transactions that fall in the target month
		var statementTime *time.Time
		var transactionID string
		var aggSettlement, aggShipCustomer, aggShipActual, aggShipPlatformDisc float64
		var matchCount int

		for _, st := range tx.Data.StatementTransactions {
			t := ParseTiktokTimestamp(st.StatementTime)
			if int(t.Month()) == month && t.Year() == year {
				matchCount++
				if matchCount == 1 {
					statementTime = &t
					transactionID = st.StatementID
				}

				// Accumulate settlement amount
				if st.SettlementAmount != "" && st.SettlementAmount != "0" {
					aggSettlement += ParseFloat(st.SettlementAmount)
				} else if st.Amount != "" && st.Amount != "0" {
					aggSettlement += ParseFloat(st.Amount)
				}

				// Accumulate shipping fees
				if st.CustomerPaidShippingFeeAmount != "" {
					aggShipCustomer += ParseFloat(st.CustomerPaidShippingFeeAmount)
				}
				if st.PlatformShippingFeeDiscountAmount != "" {
					aggShipPlatformDisc += ParseFloat(st.PlatformShippingFeeDiscountAmount)
				}
				if st.ActualShippingFeeAmount != "" {
					shipVal := ParseFloat(st.ActualShippingFeeAmount)
					if shipVal < 0 {
						shipVal = -shipVal
					}
					aggShipActual += shipVal
				}
			}
		}

		// Apply aggregated values from statement_transactions
		if matchCount > 0 {
			totalSettlementAmount = aggSettlement
			shippingFee = aggShipCustomer
			shippingPlatformDisc = aggShipPlatformDisc
			shippingFeeActual = aggShipActual
		}

		// If no statement transaction matches the target month, try fallback:
		// Use order UpdateTime + top-level SettlementAmount from Finance API.
		// This handles newer TikTok orders (2026+) where v202309 Finance API
		// returns an empty statement_transactions array.
		if statementTime == nil {
			updateT := ParseTiktokTimestamp(order.UpdateTime)
			topLevelSettlement := tx.Data.SettlementAmount

			if int(updateT.Month()) == month && updateT.Year() == year && topLevelSettlement != "" && topLevelSettlement != "0" {
				log.Debug().Str("order_id", order.ID).
					Str("update_time", updateT.String()).
					Str("settlement", topLevelSettlement).
					Msg("[TiktokEscrowSync] Fallback: using UpdateTime + top-level SettlementAmount")
				statementTime = &updateT
				totalSettlementAmount = ParseFloat(topLevelSettlement)
				// Use ShippingCostAmount for actual shipping if available
				if tx.Data.ShippingCostAmount != "" {
					shippingFeeActual = ParseFloat(tx.Data.ShippingCostAmount)
				}
			} else {
				log.Debug().Str("order_id", order.ID).
					Int("month", month).
					Int("update_month", int(updateT.Month())).
					Int("update_year", updateT.Year()).
					Str("settlement", topLevelSettlement).
					Int("stmt_tx_count", len(tx.Data.StatementTransactions)).
					Msg("[TiktokEscrowSync] Skipping order - no settlement in target month")
				return nil
			}
		}


		buyerName := ""
		if order.RecipientAddress.Name != "" {
			buyerName = order.RecipientAddress.Name
		}

		// Atomic upsert using raw SQL with ON CONFLICT + RETURNING id
		row := dbTx.Raw(fmt.Sprintf(`
			INSERT INTO %s (
				id, tenant_id, order_id, month, year, order_status, order_date,
				buyer_name, transaction_id, statement_time,
				product_revenue, buyer_total_amount, total_settlement_amount,
				platform_commission, shipping_fee_customer_paid, shipping_fee_actual,
				shipping_fee_platform_discount,
				seller_shipping_discount, currency,
				raw_order_data, raw_transaction_data,
				synced_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (tenant_id, order_id) DO UPDATE SET
				month = EXCLUDED.month,
				year = EXCLUDED.year,
				order_status = EXCLUDED.order_status,
				order_date = EXCLUDED.order_date,
				buyer_name = EXCLUDED.buyer_name,
				transaction_id = EXCLUDED.transaction_id,
				statement_time = EXCLUDED.statement_time,
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
			buyerName, transactionID, statementTime,
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
		saved = true

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
	if !saved {
		return 0, nil
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

// saveOrderFromStatement saves an order whose settlement data comes directly from the
// TikTok Statement API. No settlement-date filtering is needed — the statement itself
// guarantees the order belongs to the target month.
func (s *TiktokEscrowSyncService) saveOrderFromStatement(
	ctx context.Context,
	order tiktokPkg.TiktokOrder,
	sd *OrderStatementData,
	skuTxs []tiktokPkg.SkuTransaction,
	month, year int,
) (int, error) {
	rawOrderData, _ := json.Marshal(order)
	rawTxData, _ := json.Marshal(sd)

	// Settlement date = 1st of target month (statements are monthly)
	statementTime := time.Date(year, time.Month(month), 1, 12, 0, 0, 0, time.UTC)
	orderDate := time.Unix(order.CreateTime, 0)

	totalSettlementAmount := ParseFloat(sd.SettlementAmount)
	shippingFee := ParseFloat(order.PaymentInfo.ShippingFee)
	totalAmount := ParseFloat(order.PaymentInfo.TotalAmount)
	subTotal := ParseFloat(order.PaymentInfo.SubTotal)
	buyerName := order.RecipientAddress.Name

	now := time.Now()
	newID := uuid.New().String()
	tables := TiktokEscrowTables()
	orderTable := s.base.Table(tables.OrderTable)
	itemTable := s.base.Table(tables.ItemTable)

	var escrowOrderID string
	var saved bool

	err := s.base.DB.WithContext(ctx).Transaction(func(dbTx *gorm.DB) error {
		row := dbTx.Raw(fmt.Sprintf(`
			INSERT INTO %s (
				id, tenant_id, order_id, month, year, order_status, order_date,
				buyer_name, transaction_id, statement_time,
				product_revenue, buyer_total_amount, total_settlement_amount,
				platform_commission, shipping_fee_customer_paid, shipping_fee_actual,
				shipping_fee_platform_discount,
				seller_shipping_discount, currency,
				raw_order_data, raw_transaction_data,
				synced_at, created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT (tenant_id, order_id) DO UPDATE SET
				month = EXCLUDED.month, year = EXCLUDED.year,
				order_status = EXCLUDED.order_status, order_date = EXCLUDED.order_date,
				buyer_name = EXCLUDED.buyer_name, transaction_id = EXCLUDED.transaction_id,
				statement_time = EXCLUDED.statement_time,
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
				synced_at = EXCLUDED.synced_at, updated_at = EXCLUDED.updated_at
			RETURNING id
		`, orderTable),
			newID, s.tenantID, order.ID, month, year, order.Status, orderDate,
			buyerName, sd.StatementID, statementTime,
			subTotal, totalAmount, totalSettlementAmount,
			0.0, shippingFee, 0.0, // platform_commission=0, shipping_fee_actual=0 (not in statement)
			0.0,
			0.0, order.PaymentInfo.Currency,
			string(rawOrderData), string(rawTxData),
			now, now, now,
		).Row()

		if err := row.Scan(&escrowOrderID); err != nil {
			return fmt.Errorf("upsert statement order %s: %w", order.ID, err)
		}
		saved = true

		if err := dbTx.Exec(
			fmt.Sprintf("DELETE FROM %s WHERE tenant_id = ? AND escrow_order_id = ?", itemTable),
			s.tenantID, escrowOrderID,
		).Error; err != nil {
			return fmt.Errorf("delete items for order %s: %w", order.ID, err)
		}

		// Build a synthetic OrderTransactionResponse so item-save helpers can be reused
		synthTx := &tiktokPkg.OrderTransactionResponse{}
		synthTx.Data.SkuTransactions = skuTxs
		synthTx.Data.Currency = order.PaymentInfo.Currency
		return s.saveEscrowItemsTx(ctx, dbTx, escrowOrderID, order, synthTx, itemTable)
	})

	if err != nil {
		return 0, err
	}
	if !saved {
		return 0, nil
	}

	if len(skuTxs) > 0 {
		return len(skuTxs), nil
	}
	return len(order.LineItems), nil
}
