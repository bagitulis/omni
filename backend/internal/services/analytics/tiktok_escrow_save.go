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

	// Fallback: derive product revenue from sku_transactions when PaymentInfo is empty.
	// Indonesian/SEA TikTok sellers: /order/202309/orders does not return sub_total.
	// sku_transactions[x].revenue_amount is the authoritative source for gross sales.
	if subTotal == 0 && len(tx.Data.SkuTransactions) > 0 {
		for _, sku := range tx.Data.SkuTransactions {
			subTotal += ParseFloat(sku.RevenueAmount)
		}
	}
	if totalAmount == 0 && subTotal > 0 {
		totalAmount = subTotal + shippingFee
	}

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
		// Strategy 1: Aggregate transactions whose statement_time falls in the target month
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

		// Strategy 2 (Opsi A — settlement-date based):
		// Only use create-time attribution when statement_transactions exist but have NO valid
		// settlement date (zero/epoch timestamp). If TikTok reports a valid settlement date
		// pointing to a DIFFERENT month, respect that — this order belongs to that other month.
		if statementTime == nil && len(tx.Data.StatementTransactions) > 0 {
			firstST := tx.Data.StatementTransactions[0]
			stmtT := ParseTiktokTimestamp(firstST.StatementTime)

			// If TikTok provides a valid settlement date for a DIFFERENT month → skip here.
			// This order will be correctly captured when that other month is synced.
			if stmtT.Year() >= 2024 && (int(stmtT.Month()) != month || stmtT.Year() != year) {
				log.Debug().Str("order_id", order.ID).
					Int("target_month", month).
					Int("settlement_month", int(stmtT.Month())).
					Int("settlement_year", stmtT.Year()).
					Msg("[TiktokEscrowSync] Settlement in different month, skipping for current period")
				return nil
			}

			// Settlement date is zero/invalid, or matches target month → use create-time attribution.
			createT := ParseTiktokTimestamp(order.CreateTime)
			if int(createT.Month()) == month && createT.Year() == year {
				// Order was created in the target month — attribute it here.
				// Use the first statement_transaction's time as the statement reference.
				refT := stmtT
				statementTime = &refT
				transactionID = firstST.StatementID

				// Aggregate ALL statement_transactions (not filtered by month)
				for _, st := range tx.Data.StatementTransactions {
					if st.SettlementAmount != "" && st.SettlementAmount != "0" {
						aggSettlement += ParseFloat(st.SettlementAmount)
					} else if st.Amount != "" && st.Amount != "0" {
						aggSettlement += ParseFloat(st.Amount)
					}
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
				totalSettlementAmount = aggSettlement
				shippingFee = aggShipCustomer
				shippingPlatformDisc = aggShipPlatformDisc
				shippingFeeActual = aggShipActual

				log.Debug().Str("order_id", order.ID).
					Int("month", month).
					Str("create_time", createT.String()).
					Float64("settlement", totalSettlementAmount).
					Int("stmt_tx_count", len(tx.Data.StatementTransactions)).
					Msg("[TiktokEscrowSync] Using create-time attribution with statement data")
			}
		}

		// Strategy 3: No statement_transactions at all — use UpdateTime + top-level SettlementAmount
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

		// Resolve currency: top-level is often empty for SEA sellers,
		// fallback to statement_transactions[].currency (e.g. "IDR")
		currency := tx.Data.Currency
		if currency == "" {
			for _, st := range tx.Data.StatementTransactions {
				if st.Currency != "" {
					currency = st.Currency
					break
				}
			}
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
			sellerDiscount, currency,
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


