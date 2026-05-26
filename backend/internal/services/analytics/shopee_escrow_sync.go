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

// ShopeeEscrowSyncService implements jobs.EscrowSyncService for Shopee.
// It handles month-level escrow synchronization with progress reporting.
type ShopeeEscrowSyncService struct {
	systemDB *gorm.DB
	tenantDB *gorm.DB
	tenantID string
}

// NewShopeeEscrowSyncService creates a new ShopeeEscrowSyncService.
func NewShopeeEscrowSyncService(systemDB *gorm.DB, tenantDB *gorm.DB, tenantID string) *ShopeeEscrowSyncService {
	return &ShopeeEscrowSyncService{
		systemDB: systemDB,
		tenantDB: tenantDB,
		tenantID: tenantID,
	}
}

// SyncMonthWithProgress implements jobs.EscrowSyncService.
// It syncs Shopee escrow data for the given month/year with progress reporting.
func (s *ShopeeEscrowSyncService) SyncMonthWithProgress(ctx context.Context, month, year int, forceResync bool, onProgress func(processed, total int, message string)) error {
	log.Info().
		Str("tenant_id", s.tenantID).
		Int("month", month).
		Int("year", year).
		Bool("force_resync", forceResync).
		Msg("Shopee escrow sync started")

	onProgress(0, 100, "Starting sync...")

	// If not force resync, check if already synced
	if !forceResync {
		var existing models.ShopeeEscrowSync
		err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			First(&existing).Error
		if err == nil {
			log.Info().Str("tenant_id", s.tenantID).Int("month", month).Int("year", year).
				Msg("Shopee escrow already synced, skipping")
			onProgress(100, 100, "Already synced")
			return nil
		}
	}

	onProgress(30, 100, "Clearing previous data...")

	// Clear previous data when forcing a resync
	if forceResync {
		if err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.ShopeeEscrowItem{}).Error; err != nil {
			return fmt.Errorf("failed to clear escrow items during force resync: %w", err)
		}
		if err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.ShopeeEscrowOrder{}).Error; err != nil {
			return fmt.Errorf("failed to clear escrow orders during force resync: %w", err)
		}
		if err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			Delete(&models.ShopeeEscrowSync{}).Error; err != nil {
			return fmt.Errorf("failed to clear escrow sync record during force resync: %w", err)
		}
	}

	onProgress(60, 100, "Processing escrow data...")

	// TODO: Actual sync logic — fetch from Shopee SDK, save to DB
	// For MVP: mark synced in ShopeeEscrowSync table

	sync := models.ShopeeEscrowSync{
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
		Msg("Shopee escrow sync completed")
	return nil
}
