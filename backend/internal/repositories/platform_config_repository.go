package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// PlatformConfig represents platform configuration
type PlatformConfig struct {
	ID           string `gorm:"primaryKey"`
	TenantID     string `gorm:"index"`
	Platform     string `gorm:"index"`
	ShopID       string
	ShopIDInt    int64
	ShopName     string
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	Region       string
	IsActive     bool      `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TableName specifies table name - uses dynamic naming for PostgreSQL compatibility
func (PlatformConfig) TableName() string {
	return models.GetTableName("PlatformConfig")
}

// PlatformConfigRepository handles platform config data access
type PlatformConfigRepository struct {
	db *gorm.DB
}

// NewPlatformConfigRepository creates a new platform config repository
func NewPlatformConfigRepository(db *gorm.DB) *PlatformConfigRepository {
	return &PlatformConfigRepository{db: db}
}

// FindByTenantAndPlatform finds config by tenant and platform
func (r *PlatformConfigRepository) FindByTenantAndPlatform(ctx context.Context, tenantID, platform string) (*PlatformConfig, error) {
	var config PlatformConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}

// FindByTenant finds all configs for a tenant
func (r *PlatformConfigRepository) FindByTenant(ctx context.Context, tenantID string) ([]PlatformConfig, error) {
	var configs []PlatformConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&configs).Error
	return configs, err
}

// Create creates a new platform config
func (r *PlatformConfigRepository) Create(ctx context.Context, config *PlatformConfig) error {
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(config).Error
}

// Update updates platform config
func (r *PlatformConfigRepository) Update(ctx context.Context, config *PlatformConfig) error {
	config.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(config).Error
}

// UpdateTokens updates tokens for a platform
func (r *PlatformConfigRepository) UpdateTokens(ctx context.Context, tenantID, platform, accessToken, refreshToken string, expiresAt int64) error {
	return r.db.WithContext(ctx).Model(&PlatformConfig{}).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"updated_at":    time.Now(),
		}).Error
}

// Upsert creates or updates platform config
func (r *PlatformConfigRepository) Upsert(ctx context.Context, config *PlatformConfig) error {
	existing, err := r.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
	if err != nil {
		return err
	}
	if existing == nil {
		return r.Create(ctx, config)
	}
	config.ID = existing.ID
	config.CreatedAt = existing.CreatedAt
	return r.Update(ctx, config)
}

// Delete deletes platform config
func (r *PlatformConfigRepository) Delete(ctx context.Context, tenantID, platform string) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Delete(&PlatformConfig{}).Error
}

// SetActiveStatus sets active status
func (r *PlatformConfigRepository) SetActiveStatus(ctx context.Context, tenantID, platform string, isActive bool) error {
	return r.db.WithContext(ctx).Model(&PlatformConfig{}).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Updates(map[string]interface{}{
			"is_active":  isActive,
			"updated_at": time.Now(),
		}).Error
}
