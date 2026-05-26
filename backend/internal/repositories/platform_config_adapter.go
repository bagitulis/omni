package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// PlatformConfigAdapter provides structured-like access to tenant platform config
// backed by the canonical key-value TenantPlatformConfigRepository.
//
// CONTAINMENT PURPOSE:
// This is the SINGLE canonical compatibility adapter that routes all active
// OAuth/status/refresh/TikTok paths through the key-value (TenantPlatformConfigRepository)
// pattern. It prevents the structured PlatformConfigRepository from being used
// on active paths where it would conflict with key-value data.
//
// The adapter reads from the same platform_configs table used by TenantPlatformConfigRepository
// (key-value rows: id, platform, config_key, config_value) and assembles the data
// into the PlatformConfig struct for backward compatibility with callers that
// previously used the structured PlatformConfigRepository.
//
// NOTE: This adapter operates on TENANT-SCOPED databases (schema-isolated).
// It does NOT use or need a tenant_id column — tenant isolation is provided
// by the database schema.
type PlatformConfigAdapter struct {
	kvRepo *TenantPlatformConfigRepository
}

// NewPlatformConfigAdapter creates a new adapter backed by the key-value repository.
// The db parameter MUST be a tenant-scoped database connection (not system DB).
func NewPlatformConfigAdapter(db *gorm.DB) *PlatformConfigAdapter {
	return &PlatformConfigAdapter{
		kvRepo: NewTenantPlatformConfigRepository(db),
	}
}

// FindByTenantAndPlatform reads key-value config for a platform and assembles
// a PlatformConfig struct. Returns nil if the platform has no key-value data.
//
// The tenantID parameter is accepted for interface compatibility with legacy
// callers but is NOT used in queries — tenant isolation is schema-based.
func (a *PlatformConfigAdapter) FindByTenantAndPlatform(ctx context.Context, tenantID, platform string) (*PlatformConfig, error) {
	configMap, err := a.kvRepo.GetAllConfigByPlatform(ctx, platform)
	if err != nil {
		return nil, fmt.Errorf("platform config adapter: get all config: %w", err)
	}

	// No data or no access token means platform is not configured
	accessToken := configMap["accessToken"]
	if accessToken == "" {
		return nil, nil
	}

	shopID := configMap["shopId"]
	shopName := configMap["shopName"]

	// Parse shop ID as int64
	var shopIDInt int64
	if shopID != "" {
		shopIDInt, _ = strconv.ParseInt(shopID, 10, 64)
	}

	// Parse expiry — stored as milliseconds in tokenExpiry key
	expiresAt := parseExpiry(configMap["tokenExpiry"])

	// Parse refresh token expiry
	_ = parseExpiry(configMap["refreshTokenExpiry"]) // used for future compatibility

	// Region — check both "region" and "country" keys
	region := configMap["region"]
	if region == "" {
		region = configMap["country"]
	}

	// IsActive: platform is active if it has a valid (non-expired) token
	// Default to true unless token is expired
	isActive := true
	if expiresAt > 0 && expiresAt < time.Now().UnixMilli() {
		isActive = false
	}

	return &PlatformConfig{
		Platform:     platform,
		ShopID:       shopID,
		ShopIDInt:    shopIDInt,
		ShopName:     shopName,
		AccessToken:  accessToken,
		RefreshToken: configMap["refreshToken"],
		ExpiresAt:    expiresAt,
		Region:       region,
		IsActive:     isActive,
	}, nil
}

// FindByTenant returns PlatformConfig entries for all platforms that have
// key-value data. Returns empty slice (not nil) when no platforms configured.
//
// The tenantID parameter is accepted for interface compatibility but is NOT
// used in queries — tenant isolation is schema-based.
func (a *PlatformConfigAdapter) FindByTenant(ctx context.Context, tenantID string) ([]PlatformConfig, error) {
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}
	var result []PlatformConfig

	for _, platform := range platforms {
		cfg, err := a.FindByTenantAndPlatform(ctx, tenantID, platform)
		if err != nil {
			// Log and continue — one platform failure shouldn't block others
			continue
		}
		if cfg != nil {
			result = append(result, *cfg)
		}
	}

	// Return empty slice, not nil, for consistent JSON serialization
	if result == nil {
		result = make([]PlatformConfig, 0)
	}

	return result, nil
}

// DeleteByPlatform removes all key-value config entries for a platform.
// Unlike the structured PlatformConfigRepository.Delete(), this does NOT
// require a tenant_id column — it deletes by platform only (schema-isolated).
func (a *PlatformConfigAdapter) DeleteByPlatform(ctx context.Context, platform string) error {
	return a.kvRepo.db.WithContext(ctx).
		Where("platform = ?", platform).
		Delete(&TenantPlatformConfig{}).Error
}

// HasConfig checks if a platform has any key-value config data.
// Returns true if the platform has at least one config entry.
func (a *PlatformConfigAdapter) HasConfig(ctx context.Context, platform string) (bool, error) {
	var count int64
	err := a.kvRepo.db.WithContext(ctx).
		Model(&TenantPlatformConfig{}).
		Where("platform = ?", platform).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("platform config adapter: count: %w", err)
	}
	return count > 0, nil
}
