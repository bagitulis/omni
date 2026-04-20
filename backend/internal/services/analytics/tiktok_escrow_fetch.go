// Package analytics provides TikTok statement-based fetch helpers for escrow sync
package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// OrderStatementData holds settlement data sourced from statement API
type OrderStatementData struct {
	OrderID          string
	SettlementAmount string
	RevenueAmount    string
	StatementID      string
}

// fetchStatementIDs lists all statement IDs for the target month.
// Statements are generated daily at 00:00 UTC by TikTok.
// Uses /finance/202309/statements (all-region compatible).
func (s *TiktokEscrowSyncService) fetchStatementIDs(
	ctx context.Context,
	client *tiktokPkg.Client,
	month, year int,
) ([]string, error) {
	startTime := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC).Unix()
	endTime := time.Date(year, time.Month(month)+1, 1, 0, 0, 0, 0, time.UTC).Unix()

	log.Info().Int("month", month).Int("year", year).
		Int64("start", startTime).Int64("end", endTime).
		Msg("[TiktokEscrowSync] Fetching statement IDs for month")

	var stmtIDs []string
	pageToken := ""

	for {
		resp, err := client.GetStatements(startTime, endTime, pageToken)
		if err != nil {
			return nil, fmt.Errorf("get statements: %w", err)
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("statements API code=%d: %s", resp.Code, resp.Message)
		}
		for _, st := range resp.Data.Statements {
			if st.StatementID != "" {
				stmtIDs = append(stmtIDs, st.StatementID)
			}
		}
		pageToken = resp.Data.NextPageToken
		if pageToken == "" {
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	log.Info().Int("count", len(stmtIDs)).Int("month", month).Int("year", year).
		Msg("[TiktokEscrowSync] Statement IDs found")
	return stmtIDs, nil
}

// collectOrdersFromStatements iterates all statements and collects order-level
// settlement data. Multiple entries per order (adjustments) are aggregated.
// Only transactions with type="ORDER" are included.
func (s *TiktokEscrowSyncService) collectOrdersFromStatements(
	ctx context.Context,
	client *tiktokPkg.Client,
	stmtIDs []string,
) map[string]*OrderStatementData {
	result := make(map[string]*OrderStatementData)
	var mu sync.Mutex

	for i, stmtID := range stmtIDs {
		pageToken := ""
		for {
			resp, err := client.GetStatementTransactions(stmtID, pageToken)
			if err != nil {
				log.Warn().Str("stmt_id", stmtID).Err(err).
					Msg("[TiktokEscrowSync] Statement tx fetch error, skipping")
				break
			}
			if resp.Code != 0 {
				log.Warn().Str("stmt_id", stmtID).Int("code", resp.Code).
					Str("msg", resp.Message).
					Msg("[TiktokEscrowSync] Statement tx API error, skipping")
				break
			}

			mu.Lock()
			for _, tx := range resp.Data.Transactions {
				if tx.OrderID == "" || tx.Type != "ORDER" {
					continue
				}
				if existing, ok := result[tx.OrderID]; ok {
					// Aggregate settlement amounts (e.g. adjustments)
					a := ParseFloat(existing.SettlementAmount)
					b := ParseFloat(tx.SettlementAmount)
					existing.SettlementAmount = fmt.Sprintf("%.2f", a+b)
				} else {
					result[tx.OrderID] = &OrderStatementData{
						OrderID:          tx.OrderID,
						SettlementAmount: tx.SettlementAmount,
						RevenueAmount:    tx.RevenueAmount,
						StatementID:      stmtID,
					}
				}
			}
			mu.Unlock()

			pageToken = resp.Data.NextPageToken
			if pageToken == "" {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		log.Debug().Msgf("[TiktokEscrowSync] Processed statement %d/%d: %s (%d orders so far)",
			i+1, len(stmtIDs), stmtID, len(result))
		time.Sleep(150 * time.Millisecond)
	}

	log.Info().Int("orders", len(result)).Msg("[TiktokEscrowSync] Orders collected from statements")
	return result
}

// buildOrdersFromIDs batch-fetches full order details in batches of 20.
// Returns []TiktokOrder enriched with buyer, status, items, and payment info.
func (s *TiktokEscrowSyncService) buildOrdersFromIDs(
	ctx context.Context,
	client *tiktokPkg.Client,
	orderIDs []string,
) []tiktokPkg.TiktokOrder {
	const batchSize = 20
	var orders []tiktokPkg.TiktokOrder

	for i := 0; i < len(orderIDs); i += batchSize {
		end := i + batchSize
		if end > len(orderIDs) {
			end = len(orderIDs)
		}
		batch := orderIDs[i:end]

		detail, err := client.GetOrderDetail(batch)
		if err != nil {
			log.Warn().Err(err).Msgf("[TiktokEscrowSync] Order detail batch %d failed", i/batchSize+1)
			continue
		}

		for _, d := range detail.Data.Orders {
			orders = append(orders, convertDetailToOrder(d))
		}

		log.Debug().Msgf("[TiktokEscrowSync] Order detail batch %d/%d: got %d orders",
			i/batchSize+1, (len(orderIDs)+batchSize-1)/batchSize, len(detail.Data.Orders))
		time.Sleep(200 * time.Millisecond)
	}

	log.Info().Int("fetched", len(orders)).Int("requested", len(orderIDs)).
		Msg("[TiktokEscrowSync] Order details batch-fetch complete")
	return orders
}

// convertDetailToOrder maps OrderDetailData (from GetOrderDetail) to TiktokOrder.
func convertDetailToOrder(d tiktokPkg.OrderDetailData) tiktokPkg.TiktokOrder {
	order := tiktokPkg.TiktokOrder{
		ID:         d.ID,
		Status:     d.Status,
		CreateTime: d.CreateTime,
		UpdateTime: d.UpdateTime,
	}
	if d.RecipientAddress != nil {
		order.RecipientAddress = *d.RecipientAddress
	}
	if d.PaymentInfo != nil {
		order.PaymentInfo.Currency    = d.PaymentInfo.Currency
		order.PaymentInfo.TotalAmount = d.PaymentInfo.TotalAmount
		order.PaymentInfo.SubTotal    = d.PaymentInfo.SubTotal
		order.PaymentInfo.ShippingFee = d.PaymentInfo.ShippingFee
	}
	for _, li := range d.LineItems {
		order.LineItems = append(order.LineItems, tiktokPkg.TiktokOrderItem{
			ID:            li.ID,
			SkuID:         li.SkuID,
			SkuName:       li.SkuName,
			ProductID:     li.ProductID,
			ProductName:   li.ProductName,
			SellerSku:     li.SellerSku,
			Quantity:      li.Quantity,
			SalePrice:     li.SalePrice,
			OriginalPrice: li.OriginalPrice,
		})
	}
	return order
}

// orderDataKeys returns a sorted slice of order IDs from the map
func orderDataKeys(m map[string]*OrderStatementData) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
