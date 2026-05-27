package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// TiktokAnalyticsService implements handlers.TiktokAnalyticsService.
// It handles settings management, sync orchestration, and related operations.
type TiktokAnalyticsService struct {
	systemDB     *gorm.DB
	tenantDB     *gorm.DB
	tenantID     string
	queueManager *jobs.QueueManager
}

// NewTiktokAnalyticsService creates a new TiktokAnalyticsService.
func NewTiktokAnalyticsService(systemDB *gorm.DB, tenantDB *gorm.DB, tenantID string) *TiktokAnalyticsService {
	return &TiktokAnalyticsService{
		systemDB:     systemDB,
		tenantDB:     tenantDB,
		tenantID:     tenantID,
		queueManager: jobs.NewQueueManager(tenantDB, tenantID),
	}
}

// GetSettings retrieves analytics settings for the tenant, platform 'tiktok'.
// Returns default values when no settings are found.
func (s *TiktokAnalyticsService) GetSettings(ctx context.Context, tenantID string) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, "tiktok").
		First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.AnalyticsSettingsDTO{
				PriceColumn:       "HARGA",
				FormulaDeduction:  1500,
				FormulaMultiplier: 0.86,
			}, nil
		}
		return nil, fmt.Errorf("failed to get analytics settings: %w", err)
	}

	return &dto.AnalyticsSettingsDTO{
		PriceColumn:       settings.PriceColumn,
		FormulaDeduction:  settings.FormulaDeduction,
		FormulaMultiplier: settings.FormulaMultiplier,
	}, nil
}

// SaveSettings upserts analytics settings for the tenant, platform 'tiktok'.
func (s *TiktokAnalyticsService) SaveSettings(ctx context.Context, tenantID string, input *dto.AnalyticsSettingsDTO) (*dto.AnalyticsSettingsDTO, error) {
	var settings models.AnalyticsSettings
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, "tiktok").
		First(&settings).Error

	if err == gorm.ErrRecordNotFound {
		settings = models.AnalyticsSettings{
			ID:       uuid.New().String(),
			TenantID: tenantID,
			Platform: "tiktok",
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to find analytics settings: %w", err)
	}

	settings.PriceColumn = input.PriceColumn
	settings.FormulaDeduction = input.FormulaDeduction
	settings.FormulaMultiplier = input.FormulaMultiplier
	settings.UpdatedAt = time.Now()

	if err := s.tenantDB.WithContext(ctx).Save(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to save analytics settings: %w", err)
	}

	return &dto.AnalyticsSettingsDTO{
		PriceColumn:       settings.PriceColumn,
		FormulaDeduction:  settings.FormulaDeduction,
		FormulaMultiplier: settings.FormulaMultiplier,
	}, nil
}

// GetSyncStatus returns the escrow sync status for the given month/year.
func (s *TiktokAnalyticsService) GetSyncStatus(ctx context.Context, tenantID string, month, year int) (*dto.SyncStatusDTO, error) {
	var sync models.TiktokEscrowSync
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		First(&sync).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &dto.SyncStatusDTO{
				Synced: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get sync status: %w", err)
	}

	syncedAt := sync.SyncedAt
	return &dto.SyncStatusDTO{
		Synced:       true,
		TotalOrders:  sync.TotalOrders,
		FailedOrders: sync.FailedOrders,
		SyncedAt:     &syncedAt,
	}, nil
}

// SyncEscrow creates an escrow sync job for the given month/year and returns the job ID.
func (s *TiktokAnalyticsService) SyncEscrow(ctx context.Context, tenantID string, month, year int, forceResync bool) (string, error) {
	data := models.EscrowSyncJobData{
		TenantID:    tenantID,
		Platform:    "tiktok",
		Month:       month,
		Year:        year,
		ForceResync: forceResync,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job data: %w", err)
	}

	req := models.CreateJobRequest{
		ID:       uuid.New().String(),
		Type:     models.JobTypeTiktokEscrowSync,
		Data:     string(dataJSON),
		Priority: "normal",
	}

	job, err := s.queueManager.AddJob(req)
	if err != nil {
		return "", fmt.Errorf("failed to create sync job: %w", err)
	}

	return job.ID, nil
}

// DeleteSyncData deletes all escrow sync data for the given month/year.
// Deletes items first (via order ID subquery since TiktokEscrowItem has no Month/Year),
// then orders, then the sync record.
func (s *TiktokAnalyticsService) DeleteSyncData(ctx context.Context, tenantID string, month, year int) error {
	// Delete items via order IDs (TiktokEscrowItem has no Month/Year fields)
	if err := s.tenantDB.WithContext(ctx).
		Where("escrow_order_id IN (?)",
			s.tenantDB.WithContext(ctx).Model(&models.TiktokEscrowOrder{}).
				Select("id").
				Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year),
		).
		Delete(&models.TiktokEscrowItem{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow items: %w", err)
	}

	// Delete orders
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Delete(&models.TiktokEscrowOrder{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow orders: %w", err)
	}

	// Delete sync record
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Delete(&models.TiktokEscrowSync{}).Error; err != nil {
		return fmt.Errorf("failed to delete escrow sync record: %w", err)
	}

	return nil
}

