// Package analytics provides concurrent worker functions for TikTok escrow sync
package analytics

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// fullSyncStatementFirst attempts statement-based sync with progress tracking.
// Returns nil if statements are unavailable (caller should fallback to order-centric).
func (s *TiktokEscrowSyncService) fullSyncStatementFirst(
	ctx context.Context, client *tiktokPkg.Client,
	tables EscrowSyncTables, month, year int,
	forceResync bool, onProgress ProgressCallback,
) *SyncResultWithProgress {
	stmtIDs, err := s.fetchStatementIDs(ctx, client, month, year)
	if err != nil {
		log.Warn().Err(err).Msg("[TiktokEscrowSync] Statement ID fetch failed, will fallback")
		return nil
	}
	if len(stmtIDs) == 0 {
		log.Info().Int("month", month).Int("year", year).
			Msg("[TiktokEscrowSync] No statements found for month, will fallback")
		return nil
	}

	if onProgress != nil {
		onProgress(12, 0, 0, fmt.Sprintf("Processing %d statements...", len(stmtIDs)))
	}
	orderData := s.collectOrdersFromStatements(ctx, client, stmtIDs)
	if len(orderData) == 0 {
		log.Warn().Int("stmts", len(stmtIDs)).
			Msg("[TiktokEscrowSync] Statements exist but no order data (tx API may be unsupported), will fallback")
		return nil
	}

	log.Info().Int("orders", len(orderData)).
		Msg("[TiktokEscrowSync] Statement-first: orders with settlement found")

	if onProgress != nil {
		onProgress(18, 0, len(orderData), fmt.Sprintf("Fetching details for %d orders...", len(orderData)))
	}
	orderIDs := orderDataKeys(orderData)
	orders := s.buildOrdersFromIDs(ctx, client, orderIDs)

	if forceResync {
		if onProgress != nil {
			onProgress(20, 0, len(orders), "Deleting existing data...")
		}
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			log.Warn().Err(err).Msg("[TiktokEscrowSync] Failed to delete month data")
		}
	}

	if onProgress != nil {
		onProgress(22, 0, len(orders), fmt.Sprintf("Saving %d statement orders...", len(orders)))
	}
	totalItems, processed, failed := s.processStatementOrdersWithProgress(
		ctx, client, orders, orderData, month, year, onProgress,
	)

	if onProgress != nil {
		onProgress(95, processed, len(orders), "Creating sync record...")
	}
	if err := s.saveSyncRecord(ctx, tables, month, year, processed, failed, nil); err != nil {
		log.Warn().Err(err).Msg("[TiktokEscrowSync] Failed to save sync record")
	}

	if onProgress != nil {
		onProgress(100, processed, processed,
			fmt.Sprintf("Statement sync selesai: %d order tersettlement bulan ini", processed))
	}

	return &SyncResultWithProgress{
		TotalOrders: len(orders), ProcessedOrders: processed,
		FailedOrders: failed, TotalItems: totalItems,
		Message: fmt.Sprintf("[Statement-First] Synced %d orders (%d failed), %d items for %d/%02d",
			processed, failed, totalItems, year, month),
	}
}

// processStatementOrdersWithProgress saves statement-sourced orders with concurrent workers.
func (s *TiktokEscrowSyncService) processStatementOrdersWithProgress(
	ctx context.Context, client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder, orderData map[string]*OrderStatementData,
	month, year int, onProgress ProgressCallback,
) (totalItems, processed, failed int) {
	totalOrders := len(orders)
	var atomicItems, atomicProcessed, atomicFailed, atomicDone int64

	progressDone := make(chan struct{})
	if onProgress != nil {
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-progressDone:
					return
				case <-ticker.C:
					done := int(atomic.LoadInt64(&atomicDone))
					proc := int(atomic.LoadInt64(&atomicProcessed))
					percent := 22 + (73 * done / max(totalOrders, 1))
					onProgress(percent, proc, totalOrders,
						fmt.Sprintf("Processing statement orders %d/%d (%d saved)...", done, totalOrders, proc))
				}
			}
		}()
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	var ctxDone bool
	for _, order := range orders {
		if ctxDone {
			break
		}

		sd := orderData[order.ID]
		if sd == nil {
			atomic.AddInt64(&atomicDone, 1)
			continue
		}

		select {
		case <-ctx.Done():
			ctxDone = true
			continue
		default:
		}

		wg.Add(1)
		go func(o tiktokPkg.TiktokOrder, settlement *OrderStatementData) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			skuTx := s.fetchSkuTransactions(ctx, client, o.ID)

			items, err := s.saveOrderFromStatement(ctx, o, settlement, skuTx, month, year)
			if err != nil {
				log.Warn().Str("order_id", o.ID).Err(err).
					Msg("[TiktokEscrowSync] Failed to save statement order")
				atomic.AddInt64(&atomicFailed, 1)
			} else if items > 0 {
				atomic.AddInt64(&atomicItems, int64(items))
				atomic.AddInt64(&atomicProcessed, 1)
			}
			atomic.AddInt64(&atomicDone, 1)
		}(order, sd)
	}

	wg.Wait()
	close(progressDone)

	totalItems = int(atomicItems)
	processed = int(atomicProcessed)
	failed = int(atomicFailed)

	log.Info().Int("processed", processed).Int("failed", failed).Int("items", totalItems).
		Msg("[TiktokEscrowSync] Statement orders saved")
	return
}

// processOrdersWithProgress processes orders concurrently with progress reporting.
// Returns failedOrderIDs containing only orders that errored (not legitimately skipped orders).
func (s *TiktokEscrowSyncService) processOrdersWithProgress(
	ctx context.Context, client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder, month, year int, onProgress ProgressCallback,
) (totalItems, processedOrders, failedOrders int, cancelled bool, failedOrderIDs []string) {
	totalOrders := len(orders)
	var atomicItems, atomicProcessed, atomicFailed, atomicDone int64

	var failedIDsMu sync.Mutex
	var failedIDsList []string

	progressDone := make(chan struct{})
	if onProgress != nil {
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-progressDone:
					return
				case <-ticker.C:
					done := int(atomic.LoadInt64(&atomicDone))
					proc := int(atomic.LoadInt64(&atomicProcessed))
					percent := 20 + (75 * done / max(totalOrders, 1))
					onProgress(percent, proc, totalOrders,
						fmt.Sprintf("Processing orders %d/%d (%d settled this month)...", done, totalOrders, proc))
				}
			}
		}()
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, order := range orders {
		select {
		case <-ctx.Done():
			cancelled = true
			break
		default:
		}
		if cancelled {
			break
		}

		wg.Add(1)
		go func(o tiktokPkg.TiktokOrder) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			items, err := s.processSingleOrder(ctx, client, o, month, year)
			if err != nil {
				atomic.AddInt64(&atomicFailed, 1)
				failedIDsMu.Lock()
				failedIDsList = append(failedIDsList, o.ID)
				failedIDsMu.Unlock()
			} else if items > 0 {
				atomic.AddInt64(&atomicItems, int64(items))
				atomic.AddInt64(&atomicProcessed, 1)
			}
			atomic.AddInt64(&atomicDone, 1)
		}(order)
	}

	wg.Wait()
	close(progressDone)

	totalItems = int(atomicItems)
	processedOrders = int(atomicProcessed)
	failedOrders = int(atomicFailed)
	failedOrderIDs = failedIDsList
	return totalItems, processedOrders, failedOrders, cancelled, failedOrderIDs
}
