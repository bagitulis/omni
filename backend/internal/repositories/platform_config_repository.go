package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// WARNING: LEGACY READ-ONLY repository.
//
// PlatformConfigRepository provides STRUCTURED-COLUMN access to the platform_configs table.
// It queries by tenant_id column — but the active OAuth paths write KEY-VALUE rows
// (platform + config_key) that do NOT have a tenant_id column.
//
// This means:
//   - FindByTenantAndPlatform() returns nil for key-value data
//   - All write methods (Create, Update, Upsert, UpdateTokens, Delete)
//     affect ONLY structured rows that have matching tenant_id
//
// ACTIVE PATHS MUST USE PlatformConfigAdapter instead, which reads from the
// key-value TenantPlatformConfigRepository and assembles PlatformConfig structs.
//
// This repository is kept for reference and backward compatibility only.
// DO NOT use it for new code on tenant-scoped databases.

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
	IsActive     bool `gorm:"default:true"`
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

// FindByTenantAndPlatform finds config by tenant and platform.
// NOTE: On tenant schemas with key-value data, this returns nil because
// key-value rows don't have a tenant_id column.
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

// Create creates a new platform config.
// LEGACY-GUARD: Refuses to create if key-value rows exist for this platform.
func (r *PlatformConfigRepository) Create(ctx context.Context, config *PlatformConfig) error {
	if err := r.checkNoKeyValueConflict(ctx, config.Platform); err != nil {
		return err
	}
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(config).Error
}

// Update updates platform config.
// LEGACY-GUARD: Refuses to update if key-value rows exist for this platform.
func (r *PlatformConfigRepository) Update(ctx context.Context, config *PlatformConfig) error {
	if err := r.checkNoKeyValueConflict(ctx, config.Platform); err != nil {
		return err
	}
	config.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(config).Error
}

// UpdateTokens updates tokens for a platform.
// LEGACY-GUARD: Refuses to update if key-value rows exist for this platform.
func (r *PlatformConfigRepository) UpdateTokens(ctx context.Context, tenantID, platform, accessToken, refreshToken string, expiresAt int64) error {
	if err := r.checkNoKeyValueConflict(ctx, platform); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&PlatformConfig{}).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
			"updated_at":    time.Now(),
		}).Error
}

// Upsert creates or updates platform config.
// LEGACY-GUARD: Refuses to upsert if key-value rows exist for this platform.
func (r *PlatformConfigRepository) Upsert(ctx context.Context, config *PlatformConfig) error {
	if err := r.checkNoKeyValueConflict(ctx, config.Platform); err != nil {
		return err
	}
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

// Delete deletes platform config.
// LEGACY-GUARD: Refuses to delete if key-value rows exist for this platform.
func (r *PlatformConfigRepository) Delete(ctx context.Context, tenantID, platform string) error {
	if err := r.checkNoKeyValueConflict(ctx, platform); err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Delete(&PlatformConfig{}).Error
}

// SetActiveStatus sets active status.
// LEGACY-GUARD: Refuses to set active status if key-value rows exist for this platform.
func (r *PlatformConfigRepository) SetActiveStatus(ctx context.Context, tenantID, platform string, isActive bool) error {
	if err := r.checkNoKeyValueConflict(ctx, platform); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&PlatformConfig{}).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Updates(map[string]interface{}{
			"is_active":  isActive,
			"updated_at": time.Now(),
		}).Error
}

// checkNoKeyValueConflict verifies that no key-value TenantPlatformConfig rows
// exist for the given platform. This prevents the structured repository from
// accidentally creating conflicting rows alongside key-value data.
// Returns nil if no conflict (safe to proceed) or if the table check fails.
func (r *PlatformConfigRepository) checkNoKeyValueConflict(ctx context.Context, platform string) error {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TenantPlatformConfig{}).
		Where("platform = ? AND config_key = ?", platform, "accessToken").
		Count(&count).Error
	if err != nil {
		// Table or column may not exist (structured-only schema), allow write
		return nil
	}
	if count > 0 {
		return errors.New("cannot write via PlatformConfigRepository: platform '" + platform + "' has " +
			strconv.FormatInt(count, 10) + " key-value row(s); use PlatformConfigAdapter instead")
	}
	return nil
}
