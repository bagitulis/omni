package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TiktokEscrowSyncService implements jobs.EscrowSyncService for TikTok.
// It handles month-level escrow synchronization with progress reporting.
type TiktokEscrowSyncService struct {
	systemDB *gorm.DB
	tenantDB *gorm.DB
	tenantID string
}

// NewTiktokEscrowSyncService creates a new TiktokEscrowSyncService.
func NewTiktokEscrowSyncService(systemDB *gorm.DB, tenantDB *gorm.DB, tenantID string) *TiktokEscrowSyncService {
	return &TiktokEscrowSyncService{
		systemDB: systemDB,
		tenantDB: tenantDB,
		tenantID: tenantID,
	}
}

// SyncMonthWithProgress implements jobs.EscrowSyncService.
// It syncs TikTok escrow data for the given month/year with progress reporting.
func (s *TiktokEscrowSyncService) SyncMonthWithProgress(ctx context.Context, month, year int, forceResync bool, onProgress func(processed, total int, message string)) error {
	log.Info().
		Str("tenant_id", s.tenantID).
		Int("month", month).
		Int("year", year).
		Bool("force_resync", forceResync).
		Msg("TikTok escrow sync started")

	onProgress(0, 100, "Starting sync...")

	// If not force resync, check if already synced
	if !forceResync {
		var existing models.TiktokEscrowSync
		err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			First(&existing).Error
		if err == nil {
			log.Info().Str("tenant_id", s.tenantID).Int("month", month).Int("year", year).
				Msg("TikTok escrow already synced, skipping")
			onProgress(100, 100, "Already synced")
			return nil
		}
	}

	onProgress(30, 100, "Clearing previous data...")

	// Clear previous data when forcing a resync
	if forceResync {
		// Delete items via order IDs (TiktokEscrowItem has no Month/Year fields)
		if err := s.tenantDB.WithContext(ctx).
			Where("escrow_order_id IN (?)",
				s.tenantDB.WithContext(ctx).Model(&models.TiktokEscrowOrder{}).
					Select("id").
					Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year),
			).
			Delete(&models.TiktokEscrowItem{}).Error; err != nil {
			return fmt.Errorf("failed to clear escrow items during force resync: %w", err)
		}
		if err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.TiktokEscrowOrder{}).Error; err != nil {
			return fmt.Errorf("failed to clear escrow orders during force resync: %w", err)
		}
		if err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.TiktokEscrowSync{}).Error; err != nil {
			return fmt.Errorf("failed to clear escrow sync record during force resync: %w", err)
		}
	}

	onProgress(60, 100, "Processing escrow data...")

	// TODO: Actual sync logic — fetch from TikTok SDK, save to DB
	// For MVP: mark synced in TiktokEscrowSync table

	sync := models.TiktokEscrowSync{
		ID:           uuid.New().String(),
		TenantID:     s.tenantID,
		Month:        month,
		Year:         year,
		TotalOrders:  0,
		FailedOrders: 0,
		SyncedAt:     time.Now(),
	}

	if err := s.tenantDB.WithContext(ctx).Create(&sync).Error; err != nil {
		return fmt.Errorf("failed to save sync record: %w", err)
	}

	onProgress(100, 100, "Sync completed")
	log.Info().
		Str("tenant_id", s.tenantID).
		Int("month", month).
		Int("year", year).
		Msg("TikTok escrow sync completed")
	return nil
}
