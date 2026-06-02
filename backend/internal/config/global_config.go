package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"sync"

	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
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
		log.Info().Msg("GlobalConfigService initialized with existing system DB connection")
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
		log.Warn().Msg("GlobalConfigService created without DB - will create connection on first use")
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

	log.Warn().Msg("GlobalConfigService.db is nil, attempting to get system DB")

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
			return "", fmt.Errorf("failed to decrypt config %s.%s: %w", platform, key, err)
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

// decrypt decrypts value using Fernet encryption.
// Requires ENCRYPTION_KEY to be set. Legacy base64 fallback is only available
// when LEGACY_CRYPTO_FALLBACK=true is explicitly set in environment.
func decrypt(encryptedValue string) (string, error) {
	encKey := os.Getenv("ENCRYPTION_KEY")
	legacyFallback := os.Getenv("LEGACY_CRYPTO_FALLBACK") == "true"

	if encKey == "" {
		if !legacyFallback {
			return "", fmt.Errorf("ENCRYPTION_KEY not set; set LEGACY_CRYPTO_FALLBACK=true to allow base64 fallback")
		}
		// Legacy fallback: base64 decode only when explicitly enabled
		decoded, err := base64.StdEncoding.DecodeString(encryptedValue)
		if err != nil {
			return "", fmt.Errorf("ENCRYPTION_KEY not set and value is not valid base64: %w", err)
		}
		log.Warn().Msg("ENCRYPTION_KEY not set, decoded base64 value (legacy fallback)")
		return string(decoded), nil
	}

	encService, err := utils.NewEncryptionService(encKey)
	if err != nil {
		return "", fmt.Errorf("failed to init encryption service: %w", err)
	}

	decrypted, err := encService.Decrypt(encryptedValue)
	if err != nil {
		if !legacyFallback {
			return "", fmt.Errorf("decrypt failed: %w", err)
		}
		// Legacy fallback: try base64 decode for pre-Fernet values
		decoded, b64Err := base64.StdEncoding.DecodeString(encryptedValue)
		if b64Err != nil {
			return "", fmt.Errorf("decrypt failed: %w (also not valid base64: %v)", err, b64Err)
		}
		log.Warn().Msg("Fernet decrypt failed, fell back to base64 decode (legacy value)")
		return string(decoded), nil
	}
	return decrypted, nil
}

// encryptValue encrypts a value using Fernet encryption.
// Requires ENCRYPTION_KEY to be set — refuses to silently fall back to base64.
func encryptValue(value string) (string, error) {
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		return "", fmt.Errorf("ENCRYPTION_KEY environment variable is not set; cannot encrypt sensitive value")
	}

	encService, err := utils.NewEncryptionService(encKey)
	if err != nil {
		return "", fmt.Errorf("failed to init encryption service: %w", err)
	}

	return encService.Encrypt(value)
}
