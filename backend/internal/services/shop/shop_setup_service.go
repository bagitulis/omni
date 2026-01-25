package shop

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ShopConfig represents shop configuration
type ShopConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     string    `gorm:"uniqueIndex;not null" json:"tenantId"`
	ShopName     string    `json:"shopName"`
	ShopID       string    `json:"shopId,omitempty"`
	LogoURL      string    `json:"logoUrl,omitempty"`
	Description  string    `json:"description,omitempty"`
	Region       string    `json:"region"`
	Currency     string    `json:"currency"`
	Timezone     string    `json:"timezone"`
	IsActive     bool      `gorm:"default:true" json:"isActive"`
	Settings     string    `json:"settings,omitempty"` // JSON
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (ShopConfig) TableName() string {
	return "shop_configs"
}

// PlatformConfig represents platform-specific configuration
type PlatformConfig struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     string    `gorm:"index;not null" json:"tenantId"`
	Platform     string    `gorm:"index;not null" json:"platform"` // shopee, lazada, tiktok
	ShopID       string    `json:"shopId"`
	ShopName     string    `json:"shopName,omitempty"`
	IsActive     bool      `gorm:"default:false" json:"isActive"`
	IsConnected  bool      `gorm:"default:false" json:"isConnected"`
	AccessToken  string    `json:"-"` // Hidden in JSON
	RefreshToken string    `json:"-"` // Hidden in JSON
	TokenExpiry  time.Time `json:"tokenExpiry,omitempty"`
	Settings     string    `json:"settings,omitempty"` // JSON
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (PlatformConfig) TableName() string {
	return "platform_configs"
}

// ShopSetupService handles shop setup operations
type ShopSetupService struct {
	db *gorm.DB
}

// NewShopSetupService creates a new shop setup service
func NewShopSetupService(db *gorm.DB) *ShopSetupService {
	return &ShopSetupService{db: db}
}

// GetShopConfig retrieves shop configuration
func (s *ShopSetupService) GetShopConfig(
	ctx context.Context,
	tenantID string,
) (*ShopConfig, error) {
	var config ShopConfig

	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		First(&config).Error

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveShopConfig saves or updates shop configuration
func (s *ShopSetupService) SaveShopConfig(
	ctx context.Context,
	config *ShopConfig,
) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ?", config.TenantID).
		Assign(config).
		FirstOrCreate(&ShopConfig{}).Error
}

// GetPlatformConfigs retrieves all platform configurations
func (s *ShopSetupService) GetPlatformConfigs(
	ctx context.Context,
	tenantID string,
) ([]PlatformConfig, error) {
	var configs []PlatformConfig

	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&configs).Error

	return configs, err
}

// GetPlatformConfig retrieves specific platform configuration
func (s *ShopSetupService) GetPlatformConfig(
	ctx context.Context,
	tenantID string,
	platform string,
) (*PlatformConfig, error) {
	var config PlatformConfig

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		First(&config).Error

	if err != nil {
		return nil, err
	}

	return &config, nil
}

// SavePlatformConfig saves or updates platform configuration
func (s *ShopSetupService) SavePlatformConfig(
	ctx context.Context,
	config *PlatformConfig,
) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", config.TenantID, config.Platform).
		Assign(config).
		FirstOrCreate(&PlatformConfig{}).Error
}

// GetActiveConnections returns number of active platform connections
func (s *ShopSetupService) GetActiveConnections(
	ctx context.Context,
	tenantID string,
) (int64, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&PlatformConfig{}).
		Where("tenant_id = ? AND is_connected = ?", tenantID, true).
		Count(&count).Error

	return count, err
}

// DisconnectPlatform disconnects a platform
func (s *ShopSetupService) DisconnectPlatform(
	ctx context.Context,
	tenantID string,
	platform string,
) error {
	return s.db.WithContext(ctx).
		Model(&PlatformConfig{}).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Updates(map[string]interface{}{
			"is_connected":  false,
			"access_token":  "",
			"refresh_token": "",
		}).Error
}
