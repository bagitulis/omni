// Package analytics provides TikTok escrow sync with progress tracking
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// SyncMonthWithProgress syncs escrow data with progress callback for background jobs
func (s *TiktokEscrowSyncService) SyncMonthWithProgress(
	ctx context.Context,
	month, year int,
	forceResync bool,
	onProgress ProgressCallback,
) (*SyncResultWithProgress, error) {
	log.Info().Msgf("[TiktokEscrowSync] Starting sync with progress for %d-%02d, tenant: %s", year, month, s.tenantID)

	if onProgress != nil {
		onProgress(0, 0, 0, "Initializing sync...")
	}

	// Validate month
	if err := s.validateMonth(month, year); err != nil {
		return nil, err
	}

	tables := TiktokEscrowTables()

	// Check if already synced (smart retry if failures exist)
	if !forceResync {
		existing, retryIDs := s.checkExistingSyncForRetry(ctx, tables, month, year)
		if existing != nil && retryIDs == nil {
			return existing, nil // Fully synced, no failures
		}
		if retryIDs != nil {
			return s.smartRetrySync(ctx, tables, month, year, retryIDs, onProgress)
		}
	}

	// Full sync (first time or force resync)
	return s.fullSync(ctx, tables, month, year, forceResync, onProgress)
}

// smartRetrySync retries only the failed orders from a previous sync
func (s *TiktokEscrowSyncService) smartRetrySync(
	ctx context.Context,
	tables EscrowSyncTables,
	month, year int,
	retryIDs []string,
	onProgress ProgressCallback,
) (*SyncResultWithProgress, error) {
	log.Info().Int("retry_count", len(retryIDs)).
		Msg("[TiktokEscrowSync] Smart retry mode — retrying failed orders only")

	if onProgress != nil {
		onProgress(5, 0, 0, fmt.Sprintf("Retrying %d failed orders...", len(retryIDs)))
	}

	if cancelled := s.checkCancellation(ctx); cancelled != nil {
		return cancelled, ctx.Err()
	}

	client, err := s.getTiktokClient()
	if err != nil {
		return nil, fmt.Errorf("get tiktok client: %w", err)
	}

	// Fetch all orders to get latest data for the failed ones
	if onProgress != nil {
		onProgress(10, 0, 0, "Fetching order data for retry...")
	}
	allOrders, err := s.fetchOrdersForSettlementWindow(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders for retry: %w", err)
	}

	// Filter to only failed order IDs
	ordersToRetry := filterOrdersByIDs(allOrders, retryIDs)

	if len(ordersToRetry) == 0 {
		log.Info().Msg("[TiktokEscrowSync] No retry orders found in API — clearing failures")
		if err := s.updateSyncRecordClearFailures(ctx, tables, month, year); err != nil {
			return nil, err
		}
		return &SyncResultWithProgress{
			TotalOrders: len(allOrders),
			Message:     "Previously failed orders no longer exist in API. Sync record updated.",
		}, nil
	}

	log.Info().Msgf("[TiktokEscrowSync] Retrying %d/%d failed orders", len(ordersToRetry), len(retryIDs))

	if onProgress != nil {
		onProgress(18, 0, len(ordersToRetry), "Enriching order details (shipping fees)...")
	}
	s.enrichOrdersWithDetails(ctx, client, ordersToRetry)

	if onProgress != nil {
		onProgress(20, 0, len(ordersToRetry), fmt.Sprintf("Processing %d retry orders...", len(ordersToRetry)))
	}
	totalItems, processed, failed, cancelled, retryFailedIDs := s.processOrdersWithProgress(ctx, client, ordersToRetry, month, year, onProgress)

	if cancelled {
		return &SyncResultWithProgress{
			TotalOrders: len(ordersToRetry), ProcessedOrders: processed,
			FailedOrders: failed, TotalItems: totalItems, Cancelled: true,
			Message: fmt.Sprintf("Retry cancelled after processing %d/%d orders", processed, len(ordersToRetry)),
		}, ctx.Err()
	}

	// Use accurately tracked failed IDs (only real API errors, not skipped orders)
	if onProgress != nil {
		onProgress(95, processed, len(ordersToRetry), "Updating sync record...")
	}
	if err := s.mergeSyncRecord(ctx, tables, month, year, processed, retryFailedIDs); err != nil {
		return nil, err
	}

	if onProgress != nil {
		onProgress(100, processed, len(ordersToRetry), "Retry completed")
	}

	return &SyncResultWithProgress{
		TotalOrders: len(ordersToRetry), ProcessedOrders: processed,
		FailedOrders: failed, TotalItems: totalItems,
		Message: fmt.Sprintf("Retried %d orders: %d succeeded, %d still failing", len(ordersToRetry), processed, failed),
	}, nil
}

// fullSync performs a complete sync (first time or force resync).
// Uses statement-first approach (accurate), falls back to order-centric if unavailable.
func (s *TiktokEscrowSyncService) fullSync(
	ctx context.Context, tables EscrowSyncTables,
	month, year int, forceResync bool, onProgress ProgressCallback,
) (*SyncResultWithProgress, error) {
	if cancelled := s.checkCancellation(ctx); cancelled != nil {
		return cancelled, ctx.Err()
	}

	if onProgress != nil {
		onProgress(5, 0, 0, "Getting TikTok credentials...")
	}
	client, err := s.getTiktokClient()
	if err != nil {
		return nil, fmt.Errorf("get tiktok client: %w", err)
	}

	// === PHASE 1: Statement-First (like Shopee wallet-tx pattern) ===
	if onProgress != nil {
		onProgress(8, 0, 0, "Fetching settlement statements...")
	}
	result := s.fullSyncStatementFirst(ctx, client, tables, month, year, forceResync, onProgress)
	if result != nil {
		return result, nil
	}

	// === PHASE 2: Order-Centric Fallback ===
	log.Info().Int("month", month).Int("year", year).
		Msg("[TiktokEscrowSync] Statement-first yielded no data, falling back to order-centric sync")

	if onProgress != nil {
		onProgress(10, 0, 0, "Fetching completed orders (fallback)...")
	}
	orders, err := s.fetchOrdersForSettlementWindow(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		return &SyncResultWithProgress{TotalOrders: 0, Message: "No completed orders found"}, nil
	}

	log.Info().Msgf("[TiktokEscrowSync] Found %d completed orders (fallback)", len(orders))

	if forceResync {
		if onProgress != nil {
			onProgress(15, 0, 0, "Deleting existing data...")
		}
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	if onProgress != nil {
		onProgress(18, 0, len(orders), "Enriching order details (shipping fees)...")
	}
	s.enrichOrdersWithDetails(ctx, client, orders)

	if onProgress != nil {
		onProgress(20, 0, len(orders), fmt.Sprintf("Processing %d orders...", len(orders)))
	}
	totalItems, processed, failed, cancelled, failedIDs := s.processOrdersWithProgress(ctx, client, orders, month, year, onProgress)

	if cancelled {
		return &SyncResultWithProgress{
			TotalOrders: len(orders), ProcessedOrders: processed,
			FailedOrders: failed, TotalItems: totalItems, Cancelled: true,
			Message: fmt.Sprintf("Cancelled after processing %d/%d orders", processed, len(orders)),
		}, ctx.Err()
	}

	// failedIDs contains only orders that had real API errors (not legitimately skipped orders)

	if onProgress != nil {
		onProgress(95, processed, len(orders), "Creating sync record...")
	}
	if err := s.saveSyncRecord(ctx, tables, month, year, processed, failed, failedIDs); err != nil {
		return nil, err
	}

	if onProgress != nil {
		// Show processed/processed (not processed/total) to avoid confusing "278/853" display.
		// The total candidates include orders from other months (future settlements).
		onProgress(100, processed, processed,
			fmt.Sprintf("Sync selesai: %d order tersettlement bulan ini (dari %d kandidat)", processed, len(orders)))
	}

	return &SyncResultWithProgress{
		TotalOrders: len(orders), ProcessedOrders: processed,
		FailedOrders: failed, TotalItems: totalItems,
		Message: fmt.Sprintf("%d orders settled this month (from %d candidates, %d failed), %d items saved",
			processed, len(orders), failed, totalItems),
	}, nil
}

// fullSyncStatementFirst attempts statement-based sync with progress tracking.
// Returns nil if statements are unavailable (caller should fallback to order-centric).
func (s *TiktokEscrowSyncService) fullSyncStatementFirst(
	ctx context.Context, client *tiktokPkg.Client,
	tables EscrowSyncTables, month, year int,
	forceResync bool, onProgress ProgressCallback,
) *SyncResultWithProgress {
	// Step 1: List statement IDs for target month
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

	// Step 2: Collect order settlements from all statements
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

	// Step 3: Batch fetch full order details
	if onProgress != nil {
		onProgress(18, 0, len(orderData), fmt.Sprintf("Fetching details for %d orders...", len(orderData)))
	}
	orderIDs := orderDataKeys(orderData)
	orders := s.buildOrdersFromIDs(ctx, client, orderIDs)

	// Step 4: Delete existing data if force resync
	if forceResync {
		if onProgress != nil {
			onProgress(20, 0, len(orders), "Deleting existing data...")
		}
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			log.Warn().Err(err).Msg("[TiktokEscrowSync] Failed to delete month data")
		}
	}

	// Step 5: Save each order with progress tracking
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

	// Progress ticker: report every 2 seconds
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


// validateMonth checks if month can be synced
func (s *TiktokEscrowSyncService) validateMonth(month, year int) error {
	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return fmt.Errorf("cannot sync current month, wait until month ends")
	}
	return nil
}

// processOrdersWithProgress processes orders concurrently with progress reporting.
// Returns failedOrderIDs containing only orders that errored (not legitimately skipped orders).
func (s *TiktokEscrowSyncService) processOrdersWithProgress(
	ctx context.Context, client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder, month, year int, onProgress ProgressCallback,
) (totalItems, processedOrders, failedOrders int, cancelled bool, failedOrderIDs []string) {
	totalOrders := len(orders)
	var atomicItems, atomicProcessed, atomicFailed, atomicDone int64

	// Tracks order IDs that had real API/DB errors (excludes legitimately-skipped orders)
	var failedIDsMu sync.Mutex
	var failedIDsList []string

	// Progress ticker: report every 2 seconds for smooth UI updates
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
	sem := make(chan struct{}, 5) // 5 concurrent workers (matches processOrdersConcurrent)

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

// processSingleOrder processes one order and returns item count
func (s *TiktokEscrowSyncService) processSingleOrder(
	ctx context.Context, client *tiktokPkg.Client,
	order tiktokPkg.TiktokOrder, month, year int,
) (int, error) {
	transaction, err := s.fetchOrderTransaction(ctx, client, order.ID)
	if err != nil {
		log.Warn().Str("order_id", order.ID).Str("tenant_id", s.tenantID).
			Err(err).Msg("[TiktokEscrowSync] Failed to fetch transaction from API")
		return 0, err
	}
	if transaction == nil || transaction.Data.OrderID == "" {
		rawResp, _ := json.Marshal(transaction)
		log.Warn().Str("order_id", order.ID).Str("tenant_id", s.tenantID).
			Str("raw_response", string(rawResp)).
			Msg("[TiktokEscrowSync] Empty transaction data from API")
		return 0, fmt.Errorf("no transaction data for order %s", order.ID)
	}
	return s.saveEscrowOrder(ctx, order, transaction, month, year)
}

// filterOrdersByIDs filters orders to only include those with matching IDs
func filterOrdersByIDs(orders []tiktokPkg.TiktokOrder, ids []string) []tiktokPkg.TiktokOrder {
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	var filtered []tiktokPkg.TiktokOrder
	for _, order := range orders {
		if idSet[order.ID] {
			filtered = append(filtered, order)
		}
	}
	return filtered
}
