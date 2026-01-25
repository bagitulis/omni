package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// AnalyticsRepository handles analytics data access
type AnalyticsRepository struct {
	db *gorm.DB
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *gorm.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetSettings gets analytics settings for a tenant
func (r *AnalyticsRepository) GetSettings(ctx context.Context, tenantID, platform string) (*models.AnalyticsSettings, error) {
	var settings models.AnalyticsSettings
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

// CreateOrUpdateSettings creates or updates analytics settings
func (r *AnalyticsRepository) CreateOrUpdateSettings(ctx context.Context, settings *models.AnalyticsSettings) error {
	existing, err := r.GetSettings(ctx, settings.TenantID, settings.Platform)
	if err != nil {
		return err
	}
	if existing == nil {
		settings.ID = uuid.New().String()
		settings.CreatedAt = time.Now()
		settings.UpdatedAt = time.Now()
		return r.db.WithContext(ctx).Create(settings).Error
	}
	settings.ID = existing.ID
	settings.CreatedAt = existing.CreatedAt
	settings.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(settings).Error
}

// GetShopeeEscrowSync gets Shopee escrow sync status
func (r *AnalyticsRepository) GetShopeeEscrowSync(ctx context.Context, tenantID string, month, year int) (*models.ShopeeEscrowSync, error) {
	var sync models.ShopeeEscrowSync
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		First(&sync).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sync, nil
}

// CreateOrUpdateShopeeEscrowSync creates or updates Shopee escrow sync
func (r *AnalyticsRepository) CreateOrUpdateShopeeEscrowSync(ctx context.Context, tenantID string, month, year, totalOrders int) error {
	existing, _ := r.GetShopeeEscrowSync(ctx, tenantID, month, year)
	if existing == nil {
		sync := &models.ShopeeEscrowSync{
			ID:          uuid.New().String(),
			TenantID:    tenantID,
			Month:       month,
			Year:        year,
			TotalOrders: totalOrders,
			SyncedAt:    time.Now(),
		}
		return r.db.WithContext(ctx).Create(sync).Error
	}
	return r.db.WithContext(ctx).Model(&models.ShopeeEscrowSync{}).
		Where("id = ?", existing.ID).
		Updates(map[string]interface{}{
			"total_orders": totalOrders,
			"synced_at":    time.Now(),
		}).Error
}

// GetTiktokEscrowSync gets TikTok escrow sync status
func (r *AnalyticsRepository) GetTiktokEscrowSync(ctx context.Context, tenantID string, month, year int) (*models.TiktokEscrowSync, error) {
	var sync models.TiktokEscrowSync
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		First(&sync).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &sync, nil
}

// GetOrderCountByPlatform gets order count by platform for a period
func (r *AnalyticsRepository) GetOrderCountByPlatform(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]int, error) {
	result := make(map[string]int)

	// Shopee orders
	var shopeeCount int64
	r.db.WithContext(ctx).Model(&models.ShopeeOrder{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&shopeeCount)
	result["shopee"] = int(shopeeCount)

	// Add Lazada and TikTok counts similarly
	// For now, return shopee count
	return result, nil
}

// GetTotalSalesByPlatform gets total sales by platform for a period
func (r *AnalyticsRepository) GetTotalSalesByPlatform(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]float64, error) {
	result := make(map[string]float64)

	// Shopee sales
	var shopeeTotal float64
	r.db.WithContext(ctx).Model(&models.ShopeeOrder{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COALESCE(SUM(total_amount), 0)").
		Scan(&shopeeTotal)
	result["shopee"] = shopeeTotal

	return result, nil
}
