// Package analytics provides Shopee wallet transaction helpers for escrow sync
package analytics

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// WalletTx represents a wallet transaction
type WalletTx struct {
	OrderSN   string
	Amount    float64
	BuyerName string
}

// GetWalletTransactions gets wallet transactions from Shopee API
// Shopee API has a 15-day limit, so we chunk requests by 14 days
func (s *ShopeeEscrowSyncService) getWalletTransactions(
	ctx context.Context,
	client *shopeePkg.Client,
	month, year int,
) ([]WalletTx, error) {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Second) // Last day of month

	log.Info().Msgf("[ShopeeEscrowSync] Fetching transactions from %v to %v", firstDay, lastDay)

	var allTx []WalletTx
	processedTx := make(map[string]bool) // Dedupe by transaction key
	currentStart := firstDay

	for currentStart.Before(lastDay) || currentStart.Equal(lastDay) {
		// Chunk by 14 days
		currentEnd := currentStart.AddDate(0, 0, 14)
		if currentEnd.After(lastDay) {
			currentEnd = lastDay
		}

		log.Info().Msgf("[ShopeeEscrowSync] Processing chunk: %v to %v", currentStart, currentEnd)

		chunkTx, err := s.fetchWalletChunk(ctx, client, currentStart, currentEnd)
		if err != nil {
			log.Info().Msgf("[ShopeeEscrowSync] Error fetching chunk: %v", err)
			// Continue with next chunk instead of failing
			currentStart = currentEnd.Add(time.Second)
			continue
		}

		// Dedupe
		for _, tx := range chunkTx {
			key := fmt.Sprintf("%s_%f", tx.OrderSN, tx.Amount)
			if !processedTx[key] {
				processedTx[key] = true
				allTx = append(allTx, tx)
			}
		}

		log.Info().Msgf("[ShopeeEscrowSync] Got %d unique transactions from chunk (total: %d)",
			len(chunkTx), len(allTx))

		// Move to next chunk
		currentStart = currentEnd.Add(time.Second)
	}

	log.Info().Msgf("[ShopeeEscrowSync] Total transactions for %d-%02d: %d", year, month, len(allTx))
	return allTx, nil
}

// fetchWalletChunk fetches wallet transactions for a date range (max 15 days)
func (s *ShopeeEscrowSyncService) fetchWalletChunk(
	ctx context.Context,
	client *shopeePkg.Client,
	startDate, endDate time.Time,
) ([]WalletTx, error) {
	var allTx []WalletTx
	pageNo := 1
	pageSize := 100

	for {
		log.Info().Msgf("[ShopeeEscrowSync] Calling GetWalletTransactions page %d (range: %v to %v)",
			pageNo, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

		resp, err := client.GetWalletTransactions(shopeePkg.GetWalletTransactionRequest{
			PageNo:             pageNo,
			PageSize:           pageSize,
			StartDate:          startDate.Unix(),
			EndDate:            endDate.Unix(),
			TransactionTabType: "wallet_order_income",
			MoneyFlow:          "MONEY_IN",
		})
		if err != nil {
			log.Info().Msgf("[ShopeeEscrowSync] API error: %v", err)
			return nil, err
		}

		// Log response
		log.Info().Msgf("[ShopeeEscrowSync] Response: Error=%s, Message=%s, More=%v, TxCount=%d",
			resp.Error, resp.Message, resp.Response.More, len(resp.Response.TransactionList))

		if resp.Error != "" {
			log.Info().Msgf("[ShopeeEscrowSync] Shopee API returned error: %s - %s",
				resp.Error, resp.Message)
			return nil, fmt.Errorf("shopee API error: %s - %s", resp.Error, resp.Message)
		}

		if len(resp.Response.TransactionList) == 0 {
			break
		}

		for _, tx := range resp.Response.TransactionList {
			allTx = append(allTx, WalletTx{
				OrderSN:   tx.OrderSN,
				Amount:    tx.Amount,
				BuyerName: tx.BuyerUsername,
			})
		}

		if !resp.Response.More {
			break
		}
		pageNo++
	}

	return allTx, nil
}
