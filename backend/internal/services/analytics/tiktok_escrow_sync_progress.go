// Package analytics provides TikTok escrow sync with progress tracking
package analytics

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
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

	// Check if already synced
	tables := TiktokEscrowTables()
	if !forceResync {
		if result := s.checkExistingSync(ctx, tables, month, year); result != nil {
			return result, nil
		}
	}

	// Check cancellation
	if cancelled := s.checkCancellation(ctx); cancelled != nil {
		return cancelled, ctx.Err()
	}

	// Get client with progress
	if onProgress != nil {
		onProgress(5, 0, 0, "Getting TikTok credentials...")
	}
	client, err := s.getTiktokClient()
	if err != nil {
		return nil, fmt.Errorf("get tiktok client: %w", err)
	}

	// Fetch orders with progress
	if onProgress != nil {
		onProgress(10, 0, 0, "Fetching completed orders...")
	}
	orders, err := s.fetchOrdersByMonth(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		return &SyncResultWithProgress{TotalOrders: 0, Message: "No completed orders found"}, nil
	}

	log.Info().Msgf("[TiktokEscrowSync] Found %d completed orders", len(orders))

	// Delete existing if resync
	if forceResync {
		if onProgress != nil {
			onProgress(15, 0, 0, "Deleting existing data...")
		}
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	// Process orders
	if onProgress != nil {
		onProgress(20, 0, len(orders), fmt.Sprintf("Processing %d orders...", len(orders)))
	}
	totalItems, processed, failed, cancelled := s.processOrdersWithProgress(ctx, client, orders, month, year, onProgress)

	if cancelled {
		return &SyncResultWithProgress{
			TotalOrders:     len(orders),
			ProcessedOrders: processed,
			FailedOrders:    failed,
			TotalItems:      totalItems,
			Cancelled:       true,
			Message:         fmt.Sprintf("Cancelled after processing %d/%d orders", processed, len(orders)),
		}, ctx.Err()
	}

	// Create sync record
	if onProgress != nil {
		onProgress(95, processed, len(orders), "Creating sync record...")
	}
	if err := s.createSyncRecord(ctx, tables, month, year, processed); err != nil {
		return nil, err
	}

	if onProgress != nil {
		onProgress(100, processed, len(orders), "Sync completed")
	}

	return &SyncResultWithProgress{
		TotalOrders:     len(orders),
		ProcessedOrders: processed,
		FailedOrders:    failed,
		TotalItems:      totalItems,
		Message:         fmt.Sprintf("Synced %d/%d orders (%d failed), %d items", processed, len(orders), failed, totalItems),
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

// checkExistingSync returns result if already synced
func (s *TiktokEscrowSyncService) checkExistingSync(ctx context.Context, tables EscrowSyncTables, month, year int) *SyncResultWithProgress {
	var existing models.TiktokEscrowSync
	err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		First(&existing).Error
	if err == nil {
		return &SyncResultWithProgress{
			TotalOrders:     existing.TotalOrders,
			ProcessedOrders: existing.TotalOrders,
			Message:         "Already synced. Use forceResync to update.",
		}
	}
	return nil
}

// checkCancellation returns result if context is cancelled
func (s *TiktokEscrowSyncService) checkCancellation(ctx context.Context) *SyncResultWithProgress {
	select {
	case <-ctx.Done():
		return &SyncResultWithProgress{Cancelled: true, Message: "Cancelled before starting"}
	default:
		return nil
	}
}

// createSyncRecord creates sync record in database
func (s *TiktokEscrowSyncService) createSyncRecord(ctx context.Context, tables EscrowSyncTables, month, year, totalOrders int) error {
	syncRecord := models.TiktokEscrowSync{
		ID:          uuid.New().String(),
		TenantID:    s.tenantID,
		Month:       month,
		Year:        year,
		TotalOrders: totalOrders,
		SyncedAt:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).Create(&syncRecord).Error
}

// processOrdersWithProgress processes orders with progress callback
func (s *TiktokEscrowSyncService) processOrdersWithProgress(
	ctx context.Context,
	client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
	month, year int,
	onProgress ProgressCallback,
) (totalItems, processedOrders, failedOrders int, cancelled bool) {
	totalOrders := len(orders)

	for i, order := range orders {
		// Check cancellation
		select {
		case <-ctx.Done():
			log.Info().Msgf("[TiktokEscrowSync] Cancelled at order %d/%d", i+1, totalOrders)
			return totalItems, processedOrders, failedOrders, true
		default:
		}

		// Update progress every 5 orders
		if onProgress != nil && i%5 == 0 {
			percent := 20 + (75 * (i + 1) / totalOrders)
			onProgress(percent, processedOrders, totalOrders, fmt.Sprintf("Processing order %d/%d...", i+1, totalOrders))
		}

		// Process single order
		items, err := s.processSingleOrder(ctx, client, order, month, year)
		if err != nil {
			log.Info().Msgf("[TiktokEscrowSync] Error processing order %s: %v", order.ID, err)
			failedOrders++
			continue
		}

		totalItems += items
		processedOrders++
		time.Sleep(150 * time.Millisecond) // Rate limiting (reduced from 500ms)
	}
	return totalItems, processedOrders, failedOrders, false
}

// processSingleOrder processes one order and returns item count
func (s *TiktokEscrowSyncService) processSingleOrder(
	ctx context.Context,
	client *tiktokPkg.Client,
	order tiktokPkg.TiktokOrder,
	month, year int,
) (int, error) {
	transaction, err := s.fetchOrderTransaction(ctx, client, order.ID)
	if err != nil {
		return 0, err
	}
	if transaction == nil || transaction.Data.OrderID == "" {
		return 0, fmt.Errorf("no transaction data")
	}
	return s.saveEscrowOrder(ctx, order, transaction, month, year)
}
