package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// GetAppConfig retrieves the app config for a tenant/platform.
// Returns nil, nil if not found or the table does not exist.
func (r *CredentialRepository) GetAppConfig(ctx context.Context, tenantID, platform string) (*models.CredentialAppConfig, error) {
	if err := validateTenantPlatformScope(tenantID, platform); err != nil {
		return nil, err
	}
	var cfg models.CredentialAppConfig
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		First(&cfg).Error
	if err != nil {
		if isMissingRelationError(err) {
			return nil, nil
		}
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get app config: %w", err)
	}
	if err := r.decryptAppConfigSecrets(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// UpsertAppConfig encrypts secrets and creates or updates an app config.
func (r *CredentialRepository) UpsertAppConfig(ctx context.Context, cfg *models.CredentialAppConfig) error {
	if err := validateTenantPlatformScope(cfg.TenantID, cfg.Platform); err != nil {
		return err
	}
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	if cfg.Region == "" {
		cfg.Region = "id"
	}
	if err := r.encryptAppConfigSecrets(cfg); err != nil {
		return err
	}
	var existing models.CredentialAppConfig
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND COALESCE(store_identifier, '') = COALESCE(?, '')", cfg.TenantID, cfg.Platform, cfg.StoreIdentifier).
		First(&existing)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("upsert app config lookup: %w", result.Error)
	}
	if result.Error == gorm.ErrRecordNotFound {
		if err := r.db.WithContext(ctx).Create(cfg).Error; err != nil {
			return fmt.Errorf("create app config: %w", err)
		}
		return nil
	}
	cfg.ID = existing.ID
	if err := r.db.WithContext(ctx).
		Model(&existing).
		Where("id = ?", existing.ID).
		Updates(cfg).Error; err != nil {
		return fmt.Errorf("upsert app config: %w", err)
	}
	return nil
}

// CreateAppConfig encrypts secrets and inserts a new app config.
// Returns an error if the record already exists (use UpsertAppConfig for idempotent writes).
func (r *CredentialRepository) CreateAppConfig(ctx context.Context, cfg *models.CredentialAppConfig) error {
	if err := validateTenantPlatformScope(cfg.TenantID, cfg.Platform); err != nil {
		return err
	}
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}
	if cfg.Region == "" {
		cfg.Region = "id"
	}
	if err := r.encryptAppConfigSecrets(cfg); err != nil {
		return err
	}
	if err := r.db.WithContext(ctx).Create(cfg).Error; err != nil {
		return fmt.Errorf("create app config: %w", err)
	}
	return nil
}
