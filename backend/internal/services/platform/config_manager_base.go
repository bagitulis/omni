package platform

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/utils"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

// TokenRefreshFunc is a callback function for refreshing tokens
// Returns new accessToken, refreshToken, and error
type TokenRefreshFunc func(ctx context.Context, tenantID string) (string, string, error)

// BaseConfigManager provides common functionality for platform configs
type BaseConfigManager struct {
	platform       PlatformType
	tenantID       string
	accessToken    string
	refreshToken   string
	configs        map[string]string
	logger         *logger.Logger
	mu             sync.RWMutex
	tokenRefresher TokenRefreshFunc // Optional token refresh callback
}

// NewBaseConfigManager creates a base config manager
func NewBaseConfigManager(platform PlatformType, tenantID string) *BaseConfigManager {
	return &BaseConfigManager{
		platform: platform,
		tenantID: tenantID,
		configs:  make(map[string]string),
		logger:   logger.Named(string(platform) + "Config"),
	}
}

// SetTokenRefresher sets the callback function for token refresh
func (m *BaseConfigManager) SetTokenRefresher(refresher TokenRefreshFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokenRefresher = refresher
}

// GetPlatform returns the platform type
func (m *BaseConfigManager) GetPlatform() PlatformType {
	return m.platform
}

// GetTenantID returns the tenant ID
func (m *BaseConfigManager) GetTenantID() string {
	return m.tenantID
}

// GetAccessToken returns the access token
func (m *BaseConfigManager) GetAccessToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accessToken
}

// GetRefreshToken returns the refresh token
func (m *BaseConfigManager) GetRefreshToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.refreshToken
}

// GetTokenExpiry returns the token expiry time in milliseconds
func (m *BaseConfigManager) GetTokenExpiry() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if expiry, ok := m.configs["tokenExpiry"]; ok && expiry != "" {
		if exp, err := strconv.ParseInt(expiry, 10, 64); err == nil {
			return exp
		}
	}
	return 0
}

// IsTokenExpired checks if the access token is expired
// Returns true if token is expired or will expire within 5 minutes
func (m *BaseConfigManager) IsTokenExpired() bool {
	expiry := m.GetTokenExpiry()
	if expiry == 0 {
		return true // No expiry set, consider expired
	}

	// Token expires if current time + 5 minute buffer >= expiry time
	now := time.Now().UnixMilli()
	buffer := int64(5 * 60 * 1000) // 5 minutes in milliseconds
	return now+buffer >= expiry
}

// EnsureValidToken checks if token is expired and refreshes if needed
// Returns error if refresh fails or no refresher is set
func (m *BaseConfigManager) EnsureValidToken(ctx context.Context) error {
	if !m.IsTokenExpired() {
		return nil // Token is still valid
	}

	m.mu.RLock()
	refresher := m.tokenRefresher
	m.mu.RUnlock()

	if refresher == nil {
		return fmt.Errorf("token expired and no refresher configured for %s", m.platform)
	}

	m.logger.WithTenantID(m.tenantID).Info("Token expired, attempting refresh...")

	newAccessToken, newRefreshToken, err := refresher(ctx, m.tenantID)
	if err != nil {
		m.logger.WithTenantID(m.tenantID).Error("Token refresh failed: " + err.Error())
		return fmt.Errorf("token refresh failed: %w", err)
	}

	// Update tokens in memory
	m.mu.Lock()
	m.accessToken = newAccessToken
	m.refreshToken = newRefreshToken
	m.mu.Unlock()

	m.logger.WithTenantID(m.tenantID).Info("Token refreshed successfully")

	// Reload config from DB to get updated expiry
	return m.LoadConfigFromDB(ctx)
}

// SetTokens updates access and refresh tokens
func (m *BaseConfigManager) SetTokens(accessToken, refreshToken string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accessToken = accessToken
	m.refreshToken = refreshToken
	return nil
}

// GetConfig returns a config value
func (m *BaseConfigManager) GetConfig(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.configs[key]; ok {
		return val, true
	}
	return "", false
}

// SetConfig stores a config value
func (m *BaseConfigManager) SetConfig(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[key] = value
	return nil
}

// LoadConfigFromDB loads platform config from tenant database
// Note: platform_configs table doesn't have tenant_id - it's tenant-isolated by schema
// IMPORTANT: This decrypts encrypted values (accessToken, refreshToken)
func (m *BaseConfigManager) LoadConfigFromDB(ctx context.Context) error {
	db, err := config.GetTenantDBByID(m.tenantID)
	if err != nil {
		return fmt.Errorf("get tenant db: %w", err)
	}

	var configs []struct {
		ConfigKey   string `gorm:"column:config_key"`
		ConfigValue string `gorm:"column:config_value"`
		IsEncrypted bool   `gorm:"column:is_encrypted"`
	}

	tableName := "platform_configs"
	result := db.WithContext(ctx).
		Table(tableName).
		Where("platform = ?", string(m.platform)).
		Find(&configs)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		// Table might not exist - return without error
		m.logger.WithTenantID(m.tenantID).Debug("Config table not found or empty")
		return nil
	}

	// Get encryption service for decrypting tokens
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	var encryptionSvc *utils.EncryptionService
	if encryptionKey != "" {
		encryptionSvc, _ = utils.NewEncryptionService(encryptionKey)
	}

	m.mu.Lock()
	for _, cfg := range configs {
		value := cfg.ConfigValue

		// Decrypt if encrypted and encryption service is available
		if cfg.IsEncrypted && encryptionSvc != nil {
			decrypted, err := encryptionSvc.Decrypt(cfg.ConfigValue)
			if err == nil {
				value = decrypted
			} else {
				m.logger.WithTenantID(m.tenantID).Warn("Failed to decrypt " + cfg.ConfigKey + ": " + err.Error())
			}
		}

		m.configs[cfg.ConfigKey] = value
		if cfg.ConfigKey == "accessToken" {
			m.accessToken = value
		}
		if cfg.ConfigKey == "refreshToken" {
			m.refreshToken = value
		}
	}
	m.mu.Unlock()

	return nil
}

// SaveConfigToDB saves config to tenant database
// Note: platform_configs uses unique constraint on (platform, config_key)
func (m *BaseConfigManager) SaveConfigToDB(ctx context.Context, key, value string) error {
	db, err := config.GetTenantDBByID(m.tenantID)
	if err != nil {
		return fmt.Errorf("get tenant db: %w", err)
	}

	tableName := "platform_configs"
	result := db.WithContext(ctx).Exec(`
		INSERT INTO `+tableName+` (id, platform, config_key, config_value, updated_at)
		VALUES (gen_random_uuid()::text, ?, ?, ?, NOW())
		ON CONFLICT (platform, config_key) 
		DO UPDATE SET config_value = ?, updated_at = NOW()
	`, string(m.platform), key, value, value)

	return result.Error
}
