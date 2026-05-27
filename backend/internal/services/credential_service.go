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
	"gorm.io/gorm"
)

// CredentialService handles platform credential retrieval
// Architecture matches Node.js:
// - Global credentials (partnerId, partnerKey, appKey, appSecret) from system schema
// - Tenant-specific config (shopId, accessToken, refreshToken) from tenant.db
type CredentialService struct {
	dbPath       string
	tokenManager *TokenManager // Optional: enables auto-refresh on expired tokens
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
	Platform     string
	PartnerID    int64
	PartnerKey   string
	ShopID       int64
	AccessToken  string
	RefreshToken string
	AppKey       string // For Lazada/TikTok
	AppSecret    string
	ShopCipher   string // For TikTok
	Region       string // For Lazada
	IsProduction bool
	TokenExpiry  int64 // milliseconds since epoch
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

	// Get system database for global config
	systemDB, err := config.GetSystemDB(s.dbPath)
	if err != nil {
		return nil, fmt.Errorf("get system DB: %w", err)
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
		return creds, nil
	}
	if !IsCredentialLegacyFallbackEnabled() {
		return creds, nil
	}
	_ = repositories.NewCredentialRepository(tenantDB).CreateAuditEvent(ctx, &models.CredentialAuditEvent{
		TenantID:  tenantID,
		Platform:  platform,
		EventType: "credential_legacy_fallback_used",
		Status:    "success",
		Code:      "canonical_missing",
		Actor:     "credential_service",
		ActorRole: "service",
		Metadata:  models.JSONMap{"legacy_read_only": true, "canonical_first": true},
	})

	legacyCreds := &PlatformCredentials{Platform: platform, IsProduction: true}
	if err := s.loadGlobalCredentials(systemDB, platform, legacyCreds); err != nil {
		return nil, err
	}
	if err := s.loadTenantCredentials(ctx, tenantDB, platform, legacyCreds); err != nil {
		log.Warn().Err(err).Str("platform", platform).Str("tenant", tenantID).Msg("legacy credential fallback read failed")
	}
	fillMissingPlatformCredentials(creds, legacyCreds)

	// Check token expiry and auto-refresh if needed
	if s.tokenManager != nil && creds.AccessToken != "" && creds.TokenExpiry > 0 {
		bufferMs := int64(5 * 60 * 1000) // 5 minutes
		if time.Now().UnixMilli()+bufferMs >= creds.TokenExpiry {
			log.Info().Str("platform", platform).Str("tenant", tenantID).
				Msg("[CredentialService] Token expired or expiring soon, auto-refreshing")
			if err := s.refreshAndReload(ctx, tenantDB, tenantID, platform, creds); err != nil {
				log.Warn().Err(err).Str("platform", platform).
					Msg("[CredentialService] Auto-refresh failed, returning stale credentials")
			}
		}
	}

	return creds, nil
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
		if conn.Region != "" {
			creds.Region = conn.Region
		}
		state.StoreConfigured = true
		break
	}
	return state, nil
}

func fillMissingPlatformCredentials(target, fallback *PlatformCredentials) {
	if target.PartnerID == 0 {
		target.PartnerID = fallback.PartnerID
	}
	if target.PartnerKey == "" {
		target.PartnerKey = fallback.PartnerKey
	}
	if target.ShopID == 0 {
		target.ShopID = fallback.ShopID
	}
	if target.AccessToken == "" {
		target.AccessToken = fallback.AccessToken
	}
	if target.RefreshToken == "" {
		target.RefreshToken = fallback.RefreshToken
	}
	if target.AppKey == "" {
		target.AppKey = fallback.AppKey
	}
	if target.AppSecret == "" {
		target.AppSecret = fallback.AppSecret
	}
	if target.ShopCipher == "" {
		target.ShopCipher = fallback.ShopCipher
	}
	if target.Region == "" {
		target.Region = fallback.Region
	}
	if target.TokenExpiry == 0 {
		target.TokenExpiry = fallback.TokenExpiry
	}
	target.IsProduction = target.IsProduction || fallback.IsProduction
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

	// Reload tenant credentials so creds has the fresh token
	if err := s.loadTenantCredentials(ctx, tenantDB, platform, creds); err != nil {
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
