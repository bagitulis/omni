package repositories

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
	"gorm.io/gorm"
)

// PlatformCredentialsRepository handles platform credentials from key-value storage
// NOTE: This uses schema-level multi-tenancy - each tenant has their own schema
// No tenant_id column in the table, tenant is determined by the database connection
type PlatformCredentialsRepository struct {
	db *gorm.DB
}

// NewPlatformCredentialsRepository creates a new repository
func NewPlatformCredentialsRepository(db *gorm.DB) *PlatformCredentialsRepository {
	return &PlatformCredentialsRepository{db: db}
}

// getEncryptionService creates encryption service from environment
func getEncryptionService() *utils.EncryptionService {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		return nil
	}
	svc, err := utils.NewEncryptionService(key)
	if err != nil {
		log.Printf("[PlatformCredentials] Warning: failed to create encryption service: %v", err)
		return nil
	}
	return svc
}

// GetConfigValue gets a single config value by platform and key
func (r *PlatformCredentialsRepository) GetConfigValue(ctx context.Context, platform, configKey string) (string, error) {
	var config models.PlatformConfigKV
	err := r.db.WithContext(ctx).
		Where("platform = ? AND config_key = ?", platform, configKey).
		First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	
	// Decrypt if encrypted
	if config.IsEncrypted && config.ConfigValue != "" {
		if encSvc := getEncryptionService(); encSvc != nil {
			decrypted, err := encSvc.Decrypt(config.ConfigValue)
			if err != nil {
				log.Printf("[PlatformCredentials] Warning: failed to decrypt %s.%s: %v", platform, configKey, err)
				// Return encrypted value as fallback (for debugging)
				return config.ConfigValue, nil
			}
			return decrypted, nil
		}
	}
	
	return config.ConfigValue, nil
}

// GetAllConfigForPlatform gets all config entries for a platform as a map
// NOTE: Encrypted values (accessToken, refreshToken) will be decrypted
func (r *PlatformCredentialsRepository) GetAllConfigForPlatform(ctx context.Context, platform string) (map[string]string, error) {
	var configs []models.PlatformConfigKV
	err := r.db.WithContext(ctx).
		Where("platform = ?", platform).
		Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// Get encryption service (may be nil if ENCRYPTION_KEY not set)
	encSvc := getEncryptionService()

	result := make(map[string]string)
	for _, c := range configs {
		value := c.ConfigValue
		
		// Decrypt if encrypted
		if c.IsEncrypted && value != "" && encSvc != nil {
			decrypted, err := encSvc.Decrypt(value)
			if err != nil {
				log.Printf("[PlatformCredentials] Warning: failed to decrypt %s.%s: %v", platform, c.ConfigKey, err)
				// Keep encrypted value as fallback
			} else {
				value = decrypted
			}
		}
		
		result[c.ConfigKey] = value
	}
	return result, nil
}

// GetShopeeCredentials gets all Shopee credentials for the current tenant
func (r *PlatformCredentialsRepository) GetShopeeCredentials(ctx context.Context) (*models.TenantPlatformCredentials, error) {
	configMap, err := r.GetAllConfigForPlatform(ctx, "shopee")
	if err != nil {
		return nil, err
	}

	creds := &models.TenantPlatformCredentials{
		Platform:     "shopee",
		ShopID:       configMap["shopId"],
		ShopName:     configMap["shopName"],
		AccessToken:  configMap["accessToken"],
		RefreshToken: configMap["refreshToken"],
		TokenExpiry:  configMap["tokenExpiry"],
		Region:       configMap["region"],
	}

	// Parse ShopIDInt
	if creds.ShopID != "" {
		if shopIDInt, err := strconv.ParseInt(creds.ShopID, 10, 64); err == nil {
			creds.ShopIDInt = shopIDInt
		}
	}

	return creds, nil
}

// GetLazadaCredentials gets all Lazada credentials for the current tenant
func (r *PlatformCredentialsRepository) GetLazadaCredentials(ctx context.Context) (*models.TenantPlatformCredentials, error) {
	configMap, err := r.GetAllConfigForPlatform(ctx, "lazada")
	if err != nil {
		return nil, err
	}

	// Try both 'region' and 'country' keys for region (backward compatibility)
	region := configMap["region"]
	if region == "" {
		region = configMap["country"]
	}

	creds := &models.TenantPlatformCredentials{
		Platform:     "lazada",
		ShopID:       configMap["shopId"],
		ShopName:     configMap["shopName"],
		AccessToken:  configMap["accessToken"],
		RefreshToken: configMap["refreshToken"],
		TokenExpiry:  configMap["tokenExpiry"],
		AppKey:       configMap["appKey"],
		AppSecret:    configMap["appSecret"],
		Code:         configMap["code"],
		Region:       region,
	}

	// Parse ExpiresAt
	if expiresStr, ok := configMap["expiresAt"]; ok && expiresStr != "" {
		if expiresAt, err := strconv.ParseInt(expiresStr, 10, 64); err == nil {
			creds.ExpiresAt = expiresAt
		}
	}

	return creds, nil
}

// GetTiktokCredentials gets all TikTok credentials for the current tenant
func (r *PlatformCredentialsRepository) GetTiktokCredentials(ctx context.Context) (*models.TenantPlatformCredentials, error) {
	configMap, err := r.GetAllConfigForPlatform(ctx, "tiktok")
	if err != nil {
		return nil, err
	}

	creds := &models.TenantPlatformCredentials{
		Platform:     "tiktok",
		ShopID:       configMap["shopId"],
		ShopName:     configMap["shopName"],
		ShopCipher:   configMap["shopCipher"], // TikTok API requires shop_cipher, not shop_id
		AccessToken:  configMap["accessToken"],
		RefreshToken: configMap["refreshToken"],
		AppKey:       configMap["appKey"],
		AppSecret:    configMap["appSecret"],
		Region:       configMap["region"],
	}

	// Parse ShopIDInt (for TikTok, shopId might be numeric string)
	if creds.ShopID != "" {
		if shopIDInt, err := strconv.ParseInt(creds.ShopID, 10, 64); err == nil {
			creds.ShopIDInt = shopIDInt
		}
	}

	// Parse ExpiresAt
	if expiresStr, ok := configMap["expiresAt"]; ok && expiresStr != "" {
		if expiresAt, err := strconv.ParseInt(expiresStr, 10, 64); err == nil {
			creds.ExpiresAt = expiresAt
		}
	}

	return creds, nil
}

// SetConfigValue sets a config value for a platform
func (r *PlatformCredentialsRepository) SetConfigValue(ctx context.Context, platform, configKey, configValue string, isEncrypted bool) error {
	var existing models.PlatformConfigKV
	err := r.db.WithContext(ctx).
		Where("platform = ? AND config_key = ?", platform, configKey).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		newConfig := models.PlatformConfigKV{
			ID:          generateID(),
			Platform:    platform,
			ConfigKey:   configKey,
			ConfigValue: configValue,
			DataType:    "string",
			IsEncrypted: isEncrypted,
		}
		return r.db.WithContext(ctx).Create(&newConfig).Error
	} else if err != nil {
		return err
	}

	// Update existing
	return r.db.WithContext(ctx).Model(&existing).
		Update("config_value", configValue).Error
}

// generateID generates a simple unique ID
func generateID() string {
	return "cfg_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
