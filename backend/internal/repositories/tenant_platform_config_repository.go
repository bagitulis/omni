package repositories

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
	"gorm.io/gorm"
)

// TenantPlatformConfig represents key-value config in tenant database
// Matches Node.js Prisma schema for PlatformConfig table
type TenantPlatformConfig struct {
	ID          string    `gorm:"primaryKey"`
	Platform    string    `gorm:"index;not null"`
	ConfigKey   string    `gorm:"not null"`
	ConfigValue string    `gorm:"not null"`
	DataType    string    `gorm:"default:string"`
	IsEncrypted bool      `gorm:"default:true"`
	Metadata    *string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

// TableName returns the table name - uses dynamic naming for PostgreSQL compatibility
func (TenantPlatformConfig) TableName() string {
	return models.GetTableName("PlatformConfig")
}

// TenantPlatformConfigRepository handles tenant-specific platform config
// This reads from tenant.db (NOT system.db!)
type TenantPlatformConfigRepository struct {
	db         *gorm.DB
	encryption *utils.EncryptionService
}

// NewTenantPlatformConfigRepository creates a new repository
func NewTenantPlatformConfigRepository(db *gorm.DB) *TenantPlatformConfigRepository {
	encKey := os.Getenv("ENCRYPTION_KEY")
	var enc *utils.EncryptionService
	if encKey != "" {
		enc, _ = utils.NewEncryptionService(encKey)
	}

	return &TenantPlatformConfigRepository{db: db, encryption: enc}
}

// GetConfig retrieves a single config value by platform and key
func (r *TenantPlatformConfigRepository) GetConfig(ctx context.Context, platform, configKey string) (string, error) {
	var cfg TenantPlatformConfig
	err := r.db.WithContext(ctx).
		Where("platform = ? AND config_key = ?", platform, configKey).
		First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return r.decryptValue(&cfg), nil
}

// SetConfig saves a config value (upsert)
func (r *TenantPlatformConfigRepository) SetConfig(ctx context.Context, platform, configKey, value string, encrypt bool) error {
	storedValue := value
	if encrypt && r.encryption != nil && value != "" {
		encrypted, err := r.encryption.Encrypt(value)
		if err == nil {
			storedValue = encrypted
		}
	}

	// Check if exists
	var existing TenantPlatformConfig
	err := r.db.WithContext(ctx).
		Where("platform = ? AND config_key = ?", platform, configKey).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		cfg := TenantPlatformConfig{
			ID:          fmt.Sprintf("%s_%s_%d", platform, configKey, time.Now().UnixNano()),
			Platform:    platform,
			ConfigKey:   configKey,
			ConfigValue: storedValue,
			IsEncrypted: encrypt,
		}
		return r.db.WithContext(ctx).Create(&cfg).Error
	} else if err != nil {
		return err
	}

	// Update existing
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"config_value": storedValue,
		"is_encrypted": encrypt,
		"updated_at":   time.Now(),
	}).Error
}

// decryptValue decrypts if needed
func (r *TenantPlatformConfigRepository) decryptValue(cfg *TenantPlatformConfig) string {
	if cfg == nil || cfg.ConfigValue == "" {
		return ""
	}
	if cfg.IsEncrypted && r.encryption != nil {
		decrypted, err := r.encryption.Decrypt(cfg.ConfigValue)
		if err == nil {
			return decrypted
		}
		// Return original if decryption fails
	}
	return cfg.ConfigValue
}

// GetAllConfigByPlatform gets all configs for a platform
func (r *TenantPlatformConfigRepository) GetAllConfigByPlatform(ctx context.Context, platform string) (map[string]string, error) {
	var configs []TenantPlatformConfig
	err := r.db.WithContext(ctx).
		Where("platform = ?", platform).
		Find(&configs).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.ConfigKey] = r.decryptValue(&cfg)
	}
	return result, nil
}

// TenantTokenInfo represents token information from tenant database
type TenantTokenInfo struct {
	Platform            string
	ShopID              int64
	AccessToken         string
	RefreshToken        string
	TokenExpiry         int64  // milliseconds since epoch
	RefreshTokenExpiry  int64  // milliseconds since epoch
	ShopCipherOfSeller  string // TikTok cipher
	Region              string // Lazada region
}

// GetTokenInfo retrieves all token-related config for a platform
func (r *TenantPlatformConfigRepository) GetTokenInfo(ctx context.Context, platform string) (*TenantTokenInfo, error) {
	configs, err := r.GetAllConfigByPlatform(ctx, platform)
	if err != nil {
		return nil, err
	}

	info := &TenantTokenInfo{Platform: platform}

	// Parse shopId
	if val, ok := configs["shopId"]; ok {
		info.ShopID, _ = strconv.ParseInt(val, 10, 64)
	}
	// TikTok uses shopCipher (not shopCipherOfSeller)
	if val, ok := configs["shopCipher"]; ok {
		info.ShopCipherOfSeller = val
	}
	// Also check legacy key name
	if val, ok := configs["shopCipherOfSeller"]; ok && info.ShopCipherOfSeller == "" {
		info.ShopCipherOfSeller = val
	}

	// Tokens
	if val, ok := configs["accessToken"]; ok {
		info.AccessToken = val
	}
	if val, ok := configs["refreshToken"]; ok {
		info.RefreshToken = val
	}

	// Token expiry (stored as milliseconds or ISO string)
	if val, ok := configs["tokenExpiry"]; ok {
		info.TokenExpiry = parseExpiry(val)
	}
	if val, ok := configs["refreshTokenExpiry"]; ok {
		info.RefreshTokenExpiry = parseExpiry(val)
	}

	// Lazada region - check both "region" and "country" keys
	if val, ok := configs["region"]; ok && val != "" {
		info.Region = val
	} else if val, ok := configs["country"]; ok && val != "" {
		info.Region = val
	}

	return info, nil
}

// UpdateTokens saves new tokens after refresh
func (r *TenantPlatformConfigRepository) UpdateTokens(ctx context.Context, platform, accessToken, refreshToken string, expiresInSeconds, refreshExpiresInSeconds int64) error {
	// Always encrypt tokens
	if err := r.SetConfig(ctx, platform, "accessToken", accessToken, true); err != nil {
		return fmt.Errorf("save access token: %w", err)
	}
	if err := r.SetConfig(ctx, platform, "refreshToken", refreshToken, true); err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}

	// Store expiry as milliseconds
	tokenExpiryMs := time.Now().UnixMilli() + (expiresInSeconds * 1000)
	if err := r.SetConfig(ctx, platform, "tokenExpiry", strconv.FormatInt(tokenExpiryMs, 10), false); err != nil {
		return fmt.Errorf("save token expiry: %w", err)
	}

	if refreshExpiresInSeconds > 0 {
		refreshExpiryMs := time.Now().UnixMilli() + (refreshExpiresInSeconds * 1000)
		if err := r.SetConfig(ctx, platform, "refreshTokenExpiry", strconv.FormatInt(refreshExpiryMs, 10), false); err != nil {
			return fmt.Errorf("save refresh token expiry: %w", err)
		}
	}

	return nil
}

// parseExpiry parses expiry from either milliseconds or ISO string
func parseExpiry(val string) int64 {
	// Try parsing as number first (milliseconds)
	if ms, err := strconv.ParseInt(val, 10, 64); err == nil {
		return ms
	}
	// Try parsing as ISO string
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t.UnixMilli()
	}
	return 0
}
