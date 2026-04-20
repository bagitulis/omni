// Package analytics provides TikTok escrow sync with progress tracking
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
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
	totalItems, processed, failed, cancelled := s.processOrdersWithProgress(ctx, client, ordersToRetry, month, year, onProgress)

	if cancelled {
		return &SyncResultWithProgress{
			TotalOrders: len(ordersToRetry), ProcessedOrders: processed,
			FailedOrders: failed, TotalItems: totalItems, Cancelled: true,
			Message: fmt.Sprintf("Retry cancelled after processing %d/%d orders", processed, len(ordersToRetry)),
		}, ctx.Err()
	}

	// Collect still-failing order IDs
	failedIDs := s.collectFailedOrderIDs(ordersToRetry, nil, processed)

	if onProgress != nil {
		onProgress(95, processed, len(ordersToRetry), "Updating sync record...")
	}
	if err := s.mergeSyncRecord(ctx, tables, month, year, processed, failedIDs); err != nil {
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

// fullSync performs a complete sync (first time or force resync)
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

	if onProgress != nil {
		onProgress(10, 0, 0, "Fetching completed orders...")
	}
	orders, err := s.fetchOrdersForSettlementWindow(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		return &SyncResultWithProgress{TotalOrders: 0, Message: "No completed orders found"}, nil
	}

	log.Info().Msgf("[TiktokEscrowSync] Found %d completed orders", len(orders))

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
	totalItems, processed, failed, cancelled := s.processOrdersWithProgress(ctx, client, orders, month, year, onProgress)

	if cancelled {
		return &SyncResultWithProgress{
			TotalOrders: len(orders), ProcessedOrders: processed,
			FailedOrders: failed, TotalItems: totalItems, Cancelled: true,
			Message: fmt.Sprintf("Cancelled after processing %d/%d orders", processed, len(orders)),
		}, ctx.Err()
	}

	failedIDs := s.collectFailedOrderIDsFromAll(ctx, client, orders, month, year)

	if onProgress != nil {
		onProgress(95, processed, len(orders), "Creating sync record...")
	}
	if err := s.saveSyncRecord(ctx, tables, month, year, processed, failed, failedIDs); err != nil {
		return nil, err
	}

	if onProgress != nil {
		onProgress(100, processed, len(orders), "Sync completed")
	}

	return &SyncResultWithProgress{
		TotalOrders: len(orders), ProcessedOrders: processed,
		FailedOrders: failed, TotalItems: totalItems,
		Message: fmt.Sprintf("Synced %d/%d orders (%d failed), %d items", processed, len(orders), failed, totalItems),
	}, nil
}

// validateMonth checks if month can be synced
func (s *TiktokEscrowSyncService) validateMonth(month, year int) error {
	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return fmt.Errorf("cannot sync current month, wait until month ends")
	}
	return nil
}

// processOrdersWithProgress processes orders with progress callback
func (s *TiktokEscrowSyncService) processOrdersWithProgress(
	ctx context.Context, client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder, month, year int, onProgress ProgressCallback,
) (totalItems, processedOrders, failedOrders int, cancelled bool) {
	totalOrders := len(orders)
	for i, order := range orders {
		select {
		case <-ctx.Done():
			return totalItems, processedOrders, failedOrders, true
		default:
		}

		if onProgress != nil && i%5 == 0 {
			percent := 20 + (75 * (i + 1) / totalOrders)
			onProgress(percent, processedOrders, totalOrders, fmt.Sprintf("Processing order %d/%d...", i+1, totalOrders))
		}

		items, err := s.processSingleOrder(ctx, client, order, month, year)
		if err != nil {
			failedOrders++
			continue
		}
		if items > 0 {
			totalItems += items
			processedOrders++
		}
		time.Sleep(150 * time.Millisecond)
	}
	return totalItems, processedOrders, failedOrders, false
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
