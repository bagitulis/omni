package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// CredentialService handles platform credential retrieval
// Architecture matches Node.js:
// - Global credentials (partnerId, partnerKey, appKey, appSecret) from system schema
// - Tenant-specific config (shopId, accessToken, refreshToken) from tenant.db
type CredentialService struct {
	dbPath              string
	tokenManager        *TokenManager // Optional: enables auto-refresh on expired tokens
	refreshGroup        singleflight.Group
	refreshTokenFn      func(context.Context, string, string) error
	reloadCredentialsFn func(context.Context, *gorm.DB, string, string, *PlatformCredentials) error
}

// globalTokenManager is set once at app startup so all CredentialService instances
// can auto-refresh expired tokens without needing explicit wiring.
var globalTokenManager *TokenManager

// RegisterGlobalTokenManager registers a TokenManager for automatic token refresh.
// Called once during app initialization.
func RegisterGlobalTokenManager(tm *TokenManager) {
	globalTokenManager = tm
}

// PlatformCredentials holds credentials for a platform
type PlatformCredentials struct {
	Platform        string
	PartnerID       int64
	PartnerKey      string
	ShopID          int64
	AccessToken     string
	RefreshToken    string
	AppKey          string // For Lazada/TikTok
	AppSecret       string
	ShopCipher      string // For TikTok
	Region          string // For Lazada
	IsProduction    bool
	TokenExpiry     int64 // milliseconds since epoch
	StoreIdentifier string
}

// NewCredentialService creates a new credential service
func NewCredentialService(dbPath string) *CredentialService {
	return &CredentialService{
		dbPath:       dbPath,
		tokenManager: globalTokenManager,
	}
}

// SetTokenManager enables automatic token refresh when credentials are expired
func (s *CredentialService) SetTokenManager(tm *TokenManager) {
	s.tokenManager = tm
}

// GetPlatformCredentials retrieves credentials for a platform
// Combines global credentials (system schema) with tenant config (tenant schema)
func (s *CredentialService) GetPlatformCredentials(tenantID, platform string) (*PlatformCredentials, error) {
	ctx := context.Background()

	// Get tenant database
	tenantDB, err := config.GetTenantDB(tenantID, s.dbPath)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}

	creds := &PlatformCredentials{
		Platform:     platform,
		IsProduction: true, // Default to production
	}

	canonicalState, err := s.loadCanonicalCredentials(ctx, tenantDB, tenantID, platform, creds)
	if err != nil {
		return nil, err
	}
	if canonicalState.AppConfigured && canonicalState.StoreConfigured {
		if err := s.refreshIfExpiring(ctx, tenantDB, tenantID, platform, creds); err != nil {
			log.Warn().Err(err).Str("platform", platform).
				Msg("[CredentialService] Auto-refresh failed, returning current credentials")
		}
		return creds, nil
	}
	if !canonicalState.AppConfigured {
		return nil, fmt.Errorf("canonical app credentials not configured for %s/%s", tenantID, platform)
	}
	if !canonicalState.StoreConfigured {
		return nil, fmt.Errorf("canonical store connection not configured for %s/%s", tenantID, platform)
	}
	return nil, fmt.Errorf("canonical credentials incomplete for %s/%s", tenantID, platform)

}

func (s *CredentialService) refreshIfExpiring(ctx context.Context, tenantDB *gorm.DB, tenantID, platform string, creds *PlatformCredentials) error {
	if s.tokenManager != nil && creds.AccessToken != "" && creds.TokenExpiry > 0 {
		bufferMs := int64(5 * 60 * 1000) // 5 minutes
		if time.Now().UnixMilli()+bufferMs >= creds.TokenExpiry {
			log.Info().Str("platform", platform).Str("tenant", tenantID).
				Msg("[CredentialService] Token expired or expiring soon, auto-refreshing")
			return s.refreshAndReload(ctx, tenantDB, tenantID, platform, creds)
		}
	}
	return nil
}

type canonicalCredentialLoadState struct {
	AppConfigured   bool
	StoreConfigured bool
}

func (s *CredentialService) loadCanonicalCredentials(ctx context.Context, db *gorm.DB, tenantID, platform string, creds *PlatformCredentials) (canonicalCredentialLoadState, error) {
	repo := repositories.NewCredentialRepository(db)
	appConfig, err := repo.GetAppConfig(ctx, tenantID, platform)
	if err != nil {
		return canonicalCredentialLoadState{}, err
	}
	connections, err := repo.ListConnections(ctx, tenantID, platform)
	if err != nil {
		return canonicalCredentialLoadState{}, err
	}
	state := canonicalCredentialLoadState{}
	if appConfig != nil && appConfig.Configured {
		creds.PartnerID = appConfig.PartnerID
		creds.PartnerKey = appConfig.PartnerKey
		creds.AppKey = appConfig.AppKey
		creds.AppSecret = appConfig.AppSecret
		creds.Region = appConfig.Region
		state.AppConfigured = true
	}
	for _, conn := range connections {
		if conn.DisabledAt != nil {
			continue
		}
		creds.ShopID, _ = strconv.ParseInt(conn.StoreIdentifier, 10, 64)
		creds.AccessToken = conn.AccessToken
		creds.RefreshToken = conn.RefreshToken
		creds.ShopCipher = conn.ShopCipher
		creds.TokenExpiry = conn.TokenExpiry
		creds.StoreIdentifier = conn.StoreIdentifier
		if conn.Region != "" {
			creds.Region = conn.Region
		}
		state.StoreConfigured = true
		break
	}
	return state, nil
}

func (s *CredentialService) loadGlobalCredentials(db *gorm.DB, platform string, creds *PlatformCredentials) error {
	var configs []models.GlobalConfig
	if err := db.Where("platform = ?", platform).Find(&configs).Error; err != nil {
		return fmt.Errorf("query global config: %w", err)
	}

	for _, cfg := range configs {
		// Match Node.js keys (camelCase)
		switch cfg.ConfigKey {
		case "partnerId":
			fmt.Sscanf(cfg.ConfigValue, "%d", &creds.PartnerID)
		case "partnerKey":
			creds.PartnerKey = cfg.ConfigValue
		case "appKey":
			creds.AppKey = cfg.ConfigValue
		case "appSecret":
			creds.AppSecret = cfg.ConfigValue
		case "isProduction":
			creds.IsProduction = cfg.ConfigValue == "true"
		}
	}

	return nil
}

// loadTenantCredentials loads credentials from tenant.db PlatformConfig table (key-value format)
func (s *CredentialService) loadTenantCredentials(ctx context.Context, db *gorm.DB, platform string, creds *PlatformCredentials) error {
	// Use the new TenantPlatformConfigRepository
	tenantRepo := repositories.NewTenantPlatformConfigRepository(db)

	tokenInfo, err := tenantRepo.GetTokenInfo(ctx, platform)
	if err != nil {
		return err
	}
	if tokenInfo == nil {
		return nil // No tenant config yet
	}

	creds.ShopID = tokenInfo.ShopID
	creds.AccessToken = tokenInfo.AccessToken
	creds.RefreshToken = tokenInfo.RefreshToken
	creds.ShopCipher = tokenInfo.ShopCipherOfSeller
	creds.Region = tokenInfo.Region
	creds.TokenExpiry = tokenInfo.TokenExpiry

	return nil
}

// refreshAndReload refreshes the token for a platform and reloads credentials
func (s *CredentialService) refreshAndReload(ctx context.Context, tenantDB *gorm.DB, tenantID, platform string, creds *PlatformCredentials) error {
	key := credentialRefreshKey(tenantID, platform, creds.StoreIdentifier)
	_, err, shared := s.refreshGroup.Do(key, func() (any, error) {
		return nil, s.doRefreshAndReload(ctx, tenantDB, tenantID, platform, creds)
	})
	if shared {
		log.Info().Str("tenant_id", tenantID).Str("platform", platform).Str("store_identifier", creds.StoreIdentifier).
			Msg("[CredentialService] Deduplicated concurrent credential refresh")
	}
	return err
}

func credentialRefreshKey(tenantID, platform, storeIdentifier string) string {
	return tenantID + ":" + platform + ":" + storeIdentifier
}

func (s *CredentialService) doRefreshAndReload(ctx context.Context, tenantDB *gorm.DB, tenantID, platform string, creds *PlatformCredentials) error {
	if s.refreshTokenFn != nil {
		if err := s.refreshTokenFn(ctx, tenantID, platform); err != nil {
			return fmt.Errorf("refresh %s token: %w", platform, err)
		}
		if s.reloadCredentialsFn != nil {
			return s.reloadCredentialsFn(ctx, tenantDB, tenantID, platform, creds)
		}
		return nil
	}
	var err error
	switch platform {
	case models.PlatformShopee:
		_, err = s.tokenManager.RefreshShopeeToken(ctx, tenantID)
	case models.PlatformLazada:
		_, err = s.tokenManager.RefreshLazadaToken(ctx, tenantID)
	case models.PlatformTiktok:
		_, err = s.tokenManager.RefreshTiktokToken(ctx, tenantID)
	default:
		return fmt.Errorf("unsupported platform for token refresh: %s", platform)
	}
	if err != nil {
		return fmt.Errorf("refresh %s token: %w", platform, err)
	}

	// Reload canonical credentials so creds has the fresh token.
	if _, err := s.loadCanonicalCredentials(ctx, tenantDB, tenantID, platform, creds); err != nil {
		return fmt.Errorf("reload credentials after refresh: %w", err)
	}

	log.Info().Str("platform", platform).Str("tenant", tenantID).
		Msg("[CredentialService] Token refreshed successfully")
	return nil
}

// GetAllPlatformCredentials gets credentials for all platforms for a tenant
func (s *CredentialService) GetAllPlatformCredentials(tenantID string) (map[string]*PlatformCredentials, error) {
	platforms := []string{"shopee", "lazada", "tiktok"}
	result := make(map[string]*PlatformCredentials)

	for _, platform := range platforms {
		creds, err := s.GetPlatformCredentials(tenantID, platform)
		if err != nil {
			continue // Skip if no credentials for this platform
		}
		result[platform] = creds
	}

	return result, nil
}
