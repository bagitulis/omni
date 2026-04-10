// Package analytics provides sync record management and failed order tracking
package analytics

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// saveSyncRecord creates sync record with failed order tracking
func (s *TiktokEscrowSyncService) saveSyncRecord(
	ctx context.Context, tables EscrowSyncTables,
	month, year, totalOrders, failedOrders int, failedIDs []string,
) error {
	syncTable := s.base.Table(tables.SyncTable)
	now := time.Now()

	// Delete old record first (idempotent)
	s.base.DB.WithContext(ctx).Table(syncTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Delete(nil)

	record := map[string]interface{}{
		"id":               uuid.New().String(),
		"tenant_id":        s.tenantID,
		"month":            month,
		"year":             year,
		"total_orders":     totalOrders,
		"failed_orders":    failedOrders,
		"failed_order_ids": marshalFailedIDs(failedIDs),
		"synced_at":        now,
		"created_at":       now,
		"updated_at":       now,
	}
	return s.base.DB.WithContext(ctx).Table(syncTable).Create(record).Error
}

// mergeSyncRecord updates existing sync record after a smart retry
func (s *TiktokEscrowSyncService) mergeSyncRecord(
	ctx context.Context, tables EscrowSyncTables,
	month, year, newlyProcessed int, remainingFailedIDs []string,
) error {
	syncTable := s.base.Table(tables.SyncTable)
	now := time.Now()

	updates := map[string]interface{}{
		"total_orders":     s.getCurrentTotalOrders(ctx, syncTable, month, year) + newlyProcessed,
		"failed_orders":    len(remainingFailedIDs),
		"failed_order_ids": marshalFailedIDs(remainingFailedIDs),
		"synced_at":        now,
		"updated_at":       now,
	}

	return s.base.DB.WithContext(ctx).Table(syncTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Updates(updates).Error
}

// updateSyncRecordClearFailures clears failure data from sync record
func (s *TiktokEscrowSyncService) updateSyncRecordClearFailures(
	ctx context.Context, tables EscrowSyncTables, month, year int,
) error {
	syncTable := s.base.Table(tables.SyncTable)
	return s.base.DB.WithContext(ctx).Table(syncTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Updates(map[string]interface{}{
			"failed_orders":    0,
			"failed_order_ids": nil,
			"updated_at":       time.Now(),
		}).Error
}

// getCurrentTotalOrders reads the current total_orders from sync record
func (s *TiktokEscrowSyncService) getCurrentTotalOrders(
	ctx context.Context, syncTable string, month, year int,
) int {
	var existing models.TiktokEscrowSync
	err := s.base.DB.WithContext(ctx).Table(syncTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		First(&existing).Error
	if err != nil {
		return 0
	}
	return existing.TotalOrders
}

// checkExistingSyncForRetry checks sync record and determines the appropriate action.
// Returns (result, nil) if fully synced, (nil, retryIDs) if failures exist, or (nil, nil) if not synced.
func (s *TiktokEscrowSyncService) checkExistingSyncForRetry(
	ctx context.Context, tables EscrowSyncTables, month, year int,
) (*SyncResultWithProgress, []string) {
	var existing models.TiktokEscrowSync
	err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		First(&existing).Error
	if err != nil {
		return nil, nil // No sync record → full sync needed
	}

	if existing.FailedOrders > 0 && existing.FailedOrderIDs != nil {
		var retryIDs []string
		if err := json.Unmarshal([]byte(*existing.FailedOrderIDs), &retryIDs); err == nil && len(retryIDs) > 0 {
			log.Info().Int("failed_count", existing.FailedOrders).
				Msg("[TiktokEscrowSync] Found previous sync with failures — will smart retry")
			return nil, retryIDs
		}
	}

	// Fully synced, no failures
	return &SyncResultWithProgress{
		TotalOrders:     existing.TotalOrders,
		ProcessedOrders: existing.TotalOrders,
		Message:         "Already synced. Use forceResync to update.",
	}, nil
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

// collectFailedOrderIDsFromAll identifies which orders failed during sync
// by comparing API orders against what was saved in the database
func (s *TiktokEscrowSyncService) collectFailedOrderIDsFromAll(
	ctx context.Context,
	_ *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
	month, year int,
) []string {
	tables := TiktokEscrowTables()
	orderTable := s.base.Table(tables.OrderTable)

	var savedOrderIDs []string
	s.base.DB.WithContext(ctx).Table(orderTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Pluck("order_id", &savedOrderIDs)

	savedSet := make(map[string]bool, len(savedOrderIDs))
	for _, id := range savedOrderIDs {
		savedSet[id] = true
	}

	var failedIDs []string
	for _, order := range orders {
		if !savedSet[order.ID] {
			failedIDs = append(failedIDs, order.ID)
		}
	}
	return failedIDs
}

// collectFailedOrderIDs identifies which retried orders still failed
func (s *TiktokEscrowSyncService) collectFailedOrderIDs(
	orders []tiktokPkg.TiktokOrder, _ map[string]bool, processed int,
) []string {
	if processed >= len(orders) {
		return nil
	}

	tables := TiktokEscrowTables()
	orderTable := s.base.Table(tables.OrderTable)

	orderIDs := make([]string, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}

	var savedOrderIDs []string
	s.base.DB.Table(orderTable).
		Where("tenant_id = ? AND order_id IN ?", s.tenantID, orderIDs).
		Pluck("order_id", &savedOrderIDs)

	savedSet := make(map[string]bool, len(savedOrderIDs))
	for _, id := range savedOrderIDs {
		savedSet[id] = true
	}

	var failedIDs []string
	for _, order := range orders {
		if !savedSet[order.ID] {
			failedIDs = append(failedIDs, order.ID)
		}
	}
	return failedIDs
}

// marshalFailedIDs converts failed IDs to JSON string pointer (nil if empty)
func marshalFailedIDs(ids []string) *string {
	if len(ids) == 0 {
		return nil
	}
	data, err := json.Marshal(ids)
	if err != nil {
		return nil
	}
	s := string(data)
	return &s
}
