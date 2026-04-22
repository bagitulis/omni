// Package analytics provides statement-based order saving for TikTok escrow sync
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

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

	statementTime := time.Date(year, time.Month(month), 1, 12, 0, 0, 0, time.UTC)
	orderDate := time.Unix(order.CreateTime, 0)

	totalSettlementAmount := ParseFloat(sd.SettlementAmount)
	shippingFee := ParseFloat(order.PaymentInfo.ShippingFee)
	totalAmount := ParseFloat(order.PaymentInfo.TotalAmount)
	subTotal := ParseFloat(order.PaymentInfo.SubTotal)

	if subTotal == 0 && len(skuTxs) > 0 {
		for _, sku := range skuTxs {
			subTotal += ParseFloat(sku.RevenueAmount)
		}
	}
	if totalAmount == 0 && subTotal > 0 {
		totalAmount = subTotal + shippingFee
	}

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
			0.0, shippingFee, 0.0,
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
