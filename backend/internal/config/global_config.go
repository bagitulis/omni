package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/omni/backend/internal/utils"
	"gorm.io/gorm"
)

// GlobalConfig represents a global configuration entry stored in system database
type GlobalConfig struct {
	ID          string `gorm:"primaryKey;column:id;type:uuid"`
	Platform    string `gorm:"column:platform;uniqueIndex:idx_platform_key"`
	ConfigKey   string `gorm:"column:config_key;uniqueIndex:idx_platform_key"`
	ConfigValue string `gorm:"column:config_value"`
	IsEncrypted bool   `gorm:"column:is_encrypted"`
	Description string `gorm:"column:description"`
}

// TableName specifies the table name for GORM
func (GlobalConfig) TableName() string {
	return "global_config"
}

// ShopeeCredentials holds Shopee platform credentials
type ShopeeCredentials struct {
	PartnerID      string
	PartnerKey     string
	PushPartnerKey string
}

// TiktokCredentials holds TikTok platform credentials
type TiktokCredentials struct {
	AppKey    string
	AppSecret string
}

// LazadaCredentials holds Lazada platform credentials
type LazadaCredentials struct {
	AppKey    string
	AppSecret string
}

// GlobalConfigService manages global configuration
type GlobalConfigService struct {
	db       *gorm.DB
	basePath string
	mu       sync.RWMutex
}

var (
	globalConfigInstance *GlobalConfigService
	globalConfigOnce     sync.Once
)

// InitGlobalConfigService initializes the singleton with an existing DB connection
// This should be called once during app startup with the system DB
func InitGlobalConfigService(systemDB *gorm.DB) *GlobalConfigService {
	globalConfigOnce.Do(func() {
		basePath := os.Getenv("CONFIG_PATH")
		if basePath == "" {
			basePath = "/app/config"
		}
		globalConfigInstance = &GlobalConfigService{
			db:       systemDB, // Use provided DB instead of creating new connection
			basePath: basePath,
		}
		log.Printf("GlobalConfigService initialized with existing system DB connection")
	})
	return globalConfigInstance
}

// GetGlobalConfigService returns the singleton instance of GlobalConfigService
func GetGlobalConfigService() *GlobalConfigService {
	globalConfigOnce.Do(func() {
		basePath := os.Getenv("CONFIG_PATH")
		if basePath == "" {
			basePath = "/app/config"
		}
		globalConfigInstance = &GlobalConfigService{
			basePath: basePath,
		}
		log.Printf("Warning: GlobalConfigService created without DB - will create connection on first use")
	})
	return globalConfigInstance
}

// getDB returns database connection for system database
// Uses the injected DB connection directly - GORM handles connection pooling and reconnection
func (s *GlobalConfigService) getDB() (*gorm.DB, error) {
	s.mu.RLock()
	db := s.db
	s.mu.RUnlock()

	// If we have an existing connection, use it directly
	// GORM's connection pool handles reconnection automatically
	if db != nil {
		return db, nil
	}

	// Only create new connection if none exists (fallback for legacy code paths)
	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.db != nil {
		return s.db, nil
	}

	log.Printf("Warning: GlobalConfigService.db is nil, attempting to get system DB...")

	// Get existing system DB from global pool
	db, err := GetSystemDB(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get system database: %w", err)
	}

	// For PostgreSQL, ensure we're using the system schema
	if GetDatabaseDriver() == DriverPostgres {
		result := db.Session(&gorm.Session{}).Exec("SET search_path TO system, public")
		if result.Error != nil {
			return nil, fmt.Errorf("failed to set schema: %w", result.Error)
		}
	}

	s.db = db
	return db, nil
}

// GetConfig retrieves a configuration value by platform and key
func (s *GlobalConfigService) GetConfig(platform, key string) (string, error) {
	db, err := s.getDB()
	if err != nil {
		return "", err
	}

	// IMPORTANT: Create a new session to avoid query condition accumulation
	// Without Session(), GORM may accumulate WHERE conditions between calls
	var config GlobalConfig
	result := db.Session(&gorm.Session{}).Where("platform = ? AND config_key = ?", platform, key).First(&config)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", result.Error
	}

	if config.IsEncrypted {
		decrypted, err := decrypt(config.ConfigValue)
		if err != nil {
			log.Printf("Warning: failed to decrypt config %s.%s: %v", platform, key, err)
			return config.ConfigValue, nil
		}
		return decrypted, nil
	}

	return config.ConfigValue, nil
}

// SetConfig stores a configuration value
func (s *GlobalConfigService) SetConfig(platform, key, value string, encrypt bool) error {
	db, err := s.getDB()
	if err != nil {
		return err
	}

	finalValue := value
	if encrypt {
		encrypted, err := encryptValue(value)
		if err != nil {
			return fmt.Errorf("failed to encrypt value: %w", err)
		}
		finalValue = encrypted
	}

	config := GlobalConfig{
		Platform:    platform,
		ConfigKey:   key,
		ConfigValue: finalValue,
		IsEncrypted: encrypt,
	}

	// IMPORTANT: Create a new session to avoid query condition accumulation
	result := db.Session(&gorm.Session{}).Where("platform = ? AND config_key = ?", platform, key).
		Assign(GlobalConfig{ConfigValue: finalValue, IsEncrypted: encrypt}).
		FirstOrCreate(&config)

	return result.Error
}

// GetShopeeCredentials returns Shopee platform credentials
func (s *GlobalConfigService) GetShopeeCredentials() (*ShopeeCredentials, error) {
	partnerId, _ := s.GetConfig("shopee", "partnerId")
	if partnerId == "" {
		partnerId = os.Getenv("SHOPEE_PARTNER_ID")
	}

	partnerKey, _ := s.GetConfig("shopee", "partnerKey")
	if partnerKey == "" {
		partnerKey = os.Getenv("SHOPEE_PARTNER_KEY")
	}

	pushPartnerKey, _ := s.GetConfig("shopee", "pushPartnerKey")
	if pushPartnerKey == "" {
		pushPartnerKey = os.Getenv("SHOPEE_PUSH_PARTNER_KEY")
	}

	return &ShopeeCredentials{
		PartnerID:      partnerId,
		PartnerKey:     partnerKey,
		PushPartnerKey: pushPartnerKey,
	}, nil
}

// GetTiktokCredentials returns TikTok platform credentials
func (s *GlobalConfigService) GetTiktokCredentials() (*TiktokCredentials, error) {
	appKey, _ := s.GetConfig("tiktok", "appKey")
	if appKey == "" {
		appKey = os.Getenv("TIKTOK_APP_KEY")
	}

	appSecret, _ := s.GetConfig("tiktok", "appSecret")
	if appSecret == "" {
		appSecret = os.Getenv("TIKTOK_APP_SECRET")
	}

	return &TiktokCredentials{
		AppKey:    appKey,
		AppSecret: appSecret,
	}, nil
}

// GetLazadaCredentials returns Lazada platform credentials
func (s *GlobalConfigService) GetLazadaCredentials() (*LazadaCredentials, error) {
	appKey, _ := s.GetConfig("lazada", "appKey")
	if appKey == "" {
		appKey = os.Getenv("LAZADA_APP_KEY")
	}

	appSecret, _ := s.GetConfig("lazada", "appSecret")
	if appSecret == "" {
		appSecret = os.Getenv("LAZADA_APP_SECRET")
	}

	return &LazadaCredentials{
		AppKey:    appKey,
		AppSecret: appSecret,
	}, nil
}

// HasShopeeCredentials checks if Shopee credentials are configured
func (s *GlobalConfigService) HasShopeeCredentials() bool {
	creds, err := s.GetShopeeCredentials()
	if err != nil {
		return false
	}
	return creds.PartnerID != "" && creds.PartnerKey != ""
}

// HasTiktokCredentials checks if TikTok credentials are configured
func (s *GlobalConfigService) HasTiktokCredentials() bool {
	creds, err := s.GetTiktokCredentials()
	if err != nil {
		return false
	}
	return creds.AppKey != "" && creds.AppSecret != ""
}

// HasLazadaCredentials checks if Lazada credentials are configured
func (s *GlobalConfigService) HasLazadaCredentials() bool {
	creds, err := s.GetLazadaCredentials()
	if err != nil {
		return false
	}
	return creds.AppKey != "" && creds.AppSecret != ""
}

// decrypt decrypts value using Fernet encryption
func decrypt(encryptedValue string) (string, error) {
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		// No encryption key configured — try base64 decode as fallback
		decoded, err := base64.StdEncoding.DecodeString(encryptedValue)
		if err != nil {
			return encryptedValue, nil
		}
		return string(decoded), nil
	}

	encService, err := utils.NewEncryptionService(encKey)
	if err != nil {
		log.Printf("Warning: failed to init encryption service: %v, returning raw value", err)
		return encryptedValue, nil
	}

	decrypted, err := encService.Decrypt(encryptedValue)
	if err != nil {
		// May be legacy base64-only value — try base64 decode
		decoded, b64Err := base64.StdEncoding.DecodeString(encryptedValue)
		if b64Err != nil {
			return encryptedValue, nil
		}
		return string(decoded), nil
	}
	return decrypted, nil
}

// encryptValue encrypts a value using Fernet encryption
func encryptValue(value string) (string, error) {
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		// No encryption key configured — fall back to base64
		log.Printf("Warning: ENCRYPTION_KEY not set, using base64 encoding (not secure)")
		return base64.StdEncoding.EncodeToString([]byte(value)), nil
	}

	encService, err := utils.NewEncryptionService(encKey)
	if err != nil {
		return "", fmt.Errorf("failed to init encryption service: %w", err)
	}

	return encService.Encrypt(value)
}
