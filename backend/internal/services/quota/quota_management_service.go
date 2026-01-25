package quota

import (
	"context"
	"time"

	"github.com/omni/backend/internal/utils"
	"gorm.io/gorm"
)

// QuotaType represents the type of quota
type QuotaType string

const (
	QuotaAPICall     QuotaType = "api_call"
	QuotaOrderSync   QuotaType = "order_sync"
	QuotaProductSync QuotaType = "product_sync"
	QuotaShipment    QuotaType = "shipment"
)

// QuotaConfig represents quota configuration
type QuotaConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"index;not null" json:"tenantId"`
	Platform    string    `gorm:"index;not null" json:"platform"`
	QuotaType   QuotaType `gorm:"index;not null" json:"quotaType"`
	DailyLimit  int       `json:"dailyLimit"`
	HourlyLimit int       `json:"hourlyLimit"`
	MinuteLimit int       `json:"minuteLimit"`
	IsEnabled   bool      `gorm:"default:true" json:"isEnabled"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (QuotaConfig) TableName() string {
	return "quota_configs"
}

// QuotaUsage represents quota usage tracking
type QuotaUsage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  string    `gorm:"index;not null" json:"tenantId"`
	Platform  string    `gorm:"index;not null" json:"platform"`
	QuotaType QuotaType `gorm:"index;not null" json:"quotaType"`
	Period    string    `gorm:"index;not null" json:"period"` // YYYY-MM-DD or YYYY-MM-DD-HH
	Usage     int       `json:"usage"`
	Limit     int       `json:"limit"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (QuotaUsage) TableName() string {
	return "quota_usage"
}

// QuotaStatus represents current quota status
type QuotaStatus struct {
	QuotaType       QuotaType `json:"quotaType"`
	DailyUsage      int       `json:"dailyUsage"`
	DailyLimit      int       `json:"dailyLimit"`
	DailyRemaining  int       `json:"dailyRemaining"`
	HourlyUsage     int       `json:"hourlyUsage"`
	HourlyLimit     int       `json:"hourlyLimit"`
	HourlyRemaining int       `json:"hourlyRemaining"`
	IsExhausted     bool      `json:"isExhausted"`
}

// QuotaManagementService handles quota management
type QuotaManagementService struct {
	db    *gorm.DB
	cache map[string]*QuotaUsage
}

// NewQuotaManagementService creates a new quota management service
func NewQuotaManagementService(db *gorm.DB) *QuotaManagementService {
	return &QuotaManagementService{
		db:    db,
		cache: make(map[string]*QuotaUsage),
	}
}

// GetQuotaConfig retrieves quota configuration
func (s *QuotaManagementService) GetQuotaConfig(
	ctx context.Context,
	tenantID string,
	platform string,
	quotaType QuotaType,
) (*QuotaConfig, error) {
	var config QuotaConfig

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND quota_type = ?",
			tenantID, platform, quotaType).
		First(&config).Error

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveQuotaConfig saves quota configuration
func (s *QuotaManagementService) SaveQuotaConfig(
	ctx context.Context,
	config *QuotaConfig,
) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND quota_type = ?",
			config.TenantID, config.Platform, config.QuotaType).
		Assign(config).
		FirstOrCreate(&QuotaConfig{}).Error
}

// IncrementUsage increments quota usage
func (s *QuotaManagementService) IncrementUsage(
	ctx context.Context,
	tenantID string,
	platform string,
	quotaType QuotaType,
	count int,
) error {
	now := utils.NowWIB()
	dailyPeriod := now.Format("2006-01-02")
	hourlyPeriod := now.Format("2006-01-02-15")

	// Update daily usage
	if err := s.incrementPeriodUsage(ctx, tenantID, platform, quotaType, dailyPeriod, count); err != nil {
		return err
	}

	// Update hourly usage
	return s.incrementPeriodUsage(ctx, tenantID, platform, quotaType, hourlyPeriod, count)
}

func (s *QuotaManagementService) incrementPeriodUsage(
	ctx context.Context,
	tenantID string,
	platform string,
	quotaType QuotaType,
	period string,
	count int,
) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var usage QuotaUsage
		err := tx.Where("tenant_id = ? AND platform = ? AND quota_type = ? AND period = ?",
			tenantID, platform, quotaType, period).
			First(&usage).Error

		if err == gorm.ErrRecordNotFound {
			usage = QuotaUsage{
				TenantID:  tenantID,
				Platform:  platform,
				QuotaType: quotaType,
				Period:    period,
				Usage:     count,
			}
			return tx.Create(&usage).Error
		}

		if err != nil {
			return err
		}

		return tx.Model(&usage).Update("usage", gorm.Expr("usage + ?", count)).Error
	})
}

// CheckQuota checks if quota is available
func (s *QuotaManagementService) CheckQuota(
	ctx context.Context,
	tenantID string,
	platform string,
	quotaType QuotaType,
) (bool, error) {
	config, err := s.GetQuotaConfig(ctx, tenantID, platform, quotaType)
	if err != nil {
		// No config means unlimited
		return true, nil
	}

	if !config.IsEnabled {
		return true, nil
	}

	status, err := s.GetQuotaStatus(ctx, tenantID, platform, quotaType)
	if err != nil {
		return true, nil
	}

	return !status.IsExhausted, nil
}

// GetQuotaStatus returns current quota status
func (s *QuotaManagementService) GetQuotaStatus(
	ctx context.Context,
	tenantID string,
	platform string,
	quotaType QuotaType,
) (*QuotaStatus, error) {
	config, err := s.GetQuotaConfig(ctx, tenantID, platform, quotaType)
	if err != nil {
		return nil, err
	}

	status := &QuotaStatus{
		QuotaType:   quotaType,
		DailyLimit:  config.DailyLimit,
		HourlyLimit: config.HourlyLimit,
	}

	now := utils.NowWIB()
	dailyPeriod := now.Format("2006-01-02")
	hourlyPeriod := now.Format("2006-01-02-15")

	// Get daily usage
	var dailyUsage QuotaUsage
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND quota_type = ? AND period = ?",
			tenantID, platform, quotaType, dailyPeriod).
		First(&dailyUsage).Error; err == nil {
		status.DailyUsage = dailyUsage.Usage
	}
	status.DailyRemaining = config.DailyLimit - status.DailyUsage
	if status.DailyRemaining < 0 {
		status.DailyRemaining = 0
	}

	// Get hourly usage
	var hourlyUsage QuotaUsage
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND quota_type = ? AND period = ?",
			tenantID, platform, quotaType, hourlyPeriod).
		First(&hourlyUsage).Error; err == nil {
		status.HourlyUsage = hourlyUsage.Usage
	}
	status.HourlyRemaining = config.HourlyLimit - status.HourlyUsage
	if status.HourlyRemaining < 0 {
		status.HourlyRemaining = 0
	}

	// Check if exhausted
	status.IsExhausted = (config.DailyLimit > 0 && status.DailyUsage >= config.DailyLimit) ||
		(config.HourlyLimit > 0 && status.HourlyUsage >= config.HourlyLimit)

	return status, nil
}

// GetAllQuotaStatuses returns all quota statuses for a tenant/platform
func (s *QuotaManagementService) GetAllQuotaStatuses(
	ctx context.Context,
	tenantID string,
	platform string,
) ([]*QuotaStatus, error) {
	var statuses []*QuotaStatus

	quotaTypes := []QuotaType{QuotaAPICall, QuotaOrderSync, QuotaProductSync, QuotaShipment}
	for _, qt := range quotaTypes {
		status, err := s.GetQuotaStatus(ctx, tenantID, platform, qt)
		if err == nil {
			statuses = append(statuses, status)
		}
	}

	return statuses, nil
}

// CleanupOldUsage removes old usage records
func (s *QuotaManagementService) CleanupOldUsage(
	ctx context.Context,
	olderThan time.Duration,
) (int64, error) {
	cutoff := time.Now().Add(-olderThan).Format("2006-01-02")

	result := s.db.WithContext(ctx).
		Where("period < ?", cutoff).
		Delete(&QuotaUsage{})

	return result.RowsAffected, result.Error
}
