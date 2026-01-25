package repositories

import (
	"context"
	"os"
	"strconv"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
	"gorm.io/gorm"
)

// GlobalConfigRepository handles GlobalConfig data access
type GlobalConfigRepository struct {
	db         *gorm.DB
	encryption *utils.EncryptionService
}

// NewGlobalConfigRepository creates a new repository
func NewGlobalConfigRepository(db *gorm.DB) *GlobalConfigRepository {
	encKey := os.Getenv("ENCRYPTION_KEY")
	var enc *utils.EncryptionService
	if encKey != "" {
		enc, _ = utils.NewEncryptionService(encKey)
	}

	return &GlobalConfigRepository{db: db, encryption: enc}
}

// GetByPlatformAndKey retrieves a config by platform and key
func (r *GlobalConfigRepository) GetByPlatformAndKey(ctx context.Context, platform, key string) (*models.GlobalConfig, error) {
	var config models.GlobalConfig
	err := r.db.WithContext(ctx).
		Where("platform = ? AND config_key = ?", platform, key).
		First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// decryptValue decrypts value if encrypted
func (r *GlobalConfigRepository) decryptValue(cfg *models.GlobalConfig) string {
	if cfg == nil {
		return ""
	}
	if cfg.IsEncrypted && r.encryption != nil {
		decrypted, err := r.encryption.Decrypt(cfg.ConfigValue)
		if err == nil {
			return decrypted
		}
	}
	return cfg.ConfigValue
}

// GetShopeeCredentials returns Shopee GLOBAL API credentials (from system.db)
// NOTE: shopId and tokens are stored in tenant.db, NOT here!
func (r *GlobalConfigRepository) GetShopeeCredentials(ctx context.Context) (*models.PlatformCredentials, error) {
	creds := &models.PlatformCredentials{}

	// GLOBAL credentials only (stored in system.db)
	if cfg, err := r.GetByPlatformAndKey(ctx, "shopee", "partnerId"); err == nil {
		creds.PartnerID, _ = strconv.ParseInt(r.decryptValue(cfg), 10, 64)
	}
	if cfg, err := r.GetByPlatformAndKey(ctx, "shopee", "partnerKey"); err == nil {
		creds.PartnerKey = r.decryptValue(cfg)
	}
	// pushPartnerKey for webhook verification
	if cfg, err := r.GetByPlatformAndKey(ctx, "shopee", "pushPartnerKey"); err == nil {
		// Store in Region field for now (or add new field)
		_ = r.decryptValue(cfg)
	}

	return creds, nil
}

// GetLazadaCredentials returns Lazada GLOBAL API credentials (from system.db)
// NOTE: tokens are stored in tenant.db, NOT here!
func (r *GlobalConfigRepository) GetLazadaCredentials(ctx context.Context) (*models.PlatformCredentials, error) {
	creds := &models.PlatformCredentials{}

	// GLOBAL credentials only
	if cfg, err := r.GetByPlatformAndKey(ctx, "lazada", "appKey"); err == nil {
		creds.AppKey = r.decryptValue(cfg)
	}
	if cfg, err := r.GetByPlatformAndKey(ctx, "lazada", "appSecret"); err == nil {
		creds.AppSecret = r.decryptValue(cfg)
	}

	return creds, nil
}

// GetTiktokCredentials returns TikTok GLOBAL API credentials (from system.db)
// NOTE: tokens are stored in tenant.db, NOT here!
func (r *GlobalConfigRepository) GetTiktokCredentials(ctx context.Context) (*models.PlatformCredentials, error) {
	creds := &models.PlatformCredentials{}

	// GLOBAL credentials only
	if cfg, err := r.GetByPlatformAndKey(ctx, "tiktok", "appKey"); err == nil {
		creds.AppKey = r.decryptValue(cfg)
	}
	if cfg, err := r.GetByPlatformAndKey(ctx, "tiktok", "appSecret"); err == nil {
		creds.AppSecret = r.decryptValue(cfg)
	}

	return creds, nil
}
