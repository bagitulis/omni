package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/omni/backend/internal/utils"
)

// TokenManager handles platform token management
// Architecture matches Node.js:
// - Global credentials (partnerId, partnerKey, appKey, appSecret) from system schema
// - Tenant-specific config (shopId, accessToken, refreshToken, tokenExpiry) from tenant schema PlatformConfig table
type TokenManager struct {
	globalConfigRepo *repositories.GlobalConfigRepository // Global credentials (system DB)
	encryption       *utils.EncryptionService
	httpClient       *http.Client
	basePath         string // For getting tenant DB
	refreshGroup     singleflight.Group // Deduplicates concurrent refresh calls per platform-tenant
}

// NewTokenManager creates a new token manager
func NewTokenManager(globalConfigRepo *repositories.GlobalConfigRepository, encryption *utils.EncryptionService, basePath string) *TokenManager {
	return &TokenManager{
		globalConfigRepo: globalConfigRepo,
		encryption:       encryption,
		basePath:         basePath,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// refreshKey builds a singleflight key for a platform-tenant combination
func refreshKey(tenantID, platform string) string {
	return tenantID + ":" + platform
}

// getCredentialRepo gets a CredentialRepository for a specific tenant
// This reads from credential_connections table (canonical store)
func (m *TokenManager) getCredentialRepo(tenantID string) (*repositories.CredentialRepository, error) {
	db, err := config.GetTenantDBWithContext(tenantID, m.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant DB for %s: %w", tenantID, err)
	}
	return repositories.NewCredentialRepository(db), nil
}

// getConnectionTokenInfo reads from credential_connections and converts to TokenInfo.
// Returns nil,nil when no active connection exists for the given platform.
//
// Bug A companion fix: the old implementation called
// `GetConnection(ctx, tenantID, platform, "")` with an empty store_identifier,
// which fails validation ("store_identifier is required") the moment the
// service is actually invoked. Now uses ListConnections + first-active
// selection to mirror how the credential API and the sync layer resolve a
// tenant's default connection.
func (m *TokenManager) getConnectionTokenInfo(ctx context.Context, tenantID, platform string) (*TokenInfo, error) {
	credRepo, err := m.getCredentialRepo(tenantID)
	if err != nil {
		return nil, err
	}

	conns, err := credRepo.ListConnections(ctx, tenantID, platform)
	if err != nil {
		return nil, err
	}
	var conn *models.CredentialConnection
	for i := range conns {
		if conns[i].DisabledAt == nil && conns[i].AccessToken != "" {
			conn = &conns[i]
			break
		}
	}
	if conn == nil {
		return &TokenInfo{
			Platform: platform,
			IsValid:  false,
		}, nil
	}

	expiresAt := time.UnixMilli(conn.TokenExpiry)
	refreshExpiresAt := time.UnixMilli(conn.RefreshExpiry)

	needsRefresh := time.Now().After(expiresAt.Add(-1 * time.Hour))
	refreshExpired := time.Now().After(refreshExpiresAt)

	if refreshExpired {
		needsRefresh = false
	}

	var shopID int64
	if conn.StoreIdentifier != "" {
		shopID, _ = strconv.ParseInt(conn.StoreIdentifier, 10, 64)
	}

	return &TokenInfo{
		Platform:            platform,
		AccessToken:         conn.AccessToken,
		RefreshToken:        conn.RefreshToken,
		ExpiresAt:           expiresAt,
		RefreshTokenExpires: refreshExpiresAt,
		ShopID:              shopID,
		ShopCipher:          conn.ShopCipher,
		IsValid:             time.Now().Before(expiresAt),
		NeedsRefresh:        needsRefresh,
		RefreshExpired:      refreshExpired,
	}, nil
}

// TokenInfo represents token information
type TokenInfo struct {
	Platform            string
	AccessToken         string
	RefreshToken        string
	ExpiresAt           time.Time
	RefreshTokenExpires time.Time
	ShopID              int64
	ShopCipher          string // TikTok shop cipher
	IsValid             bool
	NeedsRefresh        bool
	RefreshExpired      bool // True when refresh token expiry < now
}

// GetTokenStatus gets token status for a tenant/platform
// Reads from credential_connections table (canonical store)
func (m *TokenManager) GetTokenStatus(ctx context.Context, tenantID, platform string) (*TokenInfo, error) {
	tokenInfo, err := m.getConnectionTokenInfo(ctx, tenantID, platform)
	if err != nil {
		return nil, err
	}
	if tokenInfo == nil || tokenInfo.AccessToken == "" {
		return &TokenInfo{
			Platform: platform,
			IsValid:  false,
		}, nil
	}
	return tokenInfo, nil
}

// RefreshShopeeToken refreshes Shopee access token
// Gets global credentials (partnerId, partnerKey) from system schema
// Gets tenant-specific config (shopId, refreshToken) from tenant schema
// Uses singleflight to deduplicate concurrent refresh attempts for the same tenant.
func (m *TokenManager) RefreshShopeeToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	key := refreshKey(tenantID, models.PlatformShopee)

	result, err, shared := m.refreshGroup.Do(key, func() (interface{}, error) {
		return m.doRefreshShopeeToken(ctx, tenantID)
	})
	if shared {
		log.Info().Str("tenant_id", tenantID).Str("platform", models.PlatformShopee).
			Msg("[TOKEN REFRESH] Deduplicated concurrent refresh request via singleflight")
	}
	if err != nil {
		return nil, err
	}
	tokenInfo, ok := result.(*TokenInfo)
	if !ok {
		return nil, fmt.Errorf("unexpected token refresh result type: %T", result)
	}
	return tokenInfo, nil
}

func (m *TokenManager) doRefreshShopeeToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	// Get GLOBAL credentials from system schema
	globalCreds, err := m.globalConfigRepo.GetShopeeCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global shopee credentials: %w", err)
	}
	if globalCreds.PartnerID == 0 || globalCreds.PartnerKey == "" {
		return nil, errors.New("shopee global credentials (partnerId/partnerKey) not configured in system schema")
	}

	// Get TENANT-specific tokens from credential_connections
	tokenInfo, err := m.getConnectionTokenInfo(ctx, tenantID, models.PlatformShopee)
	if err != nil {
		return nil, fmt.Errorf("failed to get shopee token info: %w", err)
	}
	if tokenInfo == nil || tokenInfo.RefreshToken == "" {
		return nil, errors.New("shopee refresh token not found in credential connections")
	}
	if tokenInfo.ShopID == 0 {
		return nil, errors.New("shopee shopId not found in credential connections")
	}

	// Build refresh request using global credentials + tenant's refresh token
	shopeeService := oauth.NewShopeeOAuthService(globalCreds.PartnerID, globalCreds.PartnerKey, "", false)
	refreshURL, reqBody := shopeeService.BuildRefreshTokenRequest(tokenInfo.RefreshToken, tokenInfo.ShopID)

	return m.executeShopeeTokenRefresh(ctx, tenantID, refreshURL, reqBody)
}

// RefreshLazadaToken refreshes Lazada access token
// Uses singleflight to deduplicate concurrent refresh attempts for the same tenant.
func (m *TokenManager) RefreshLazadaToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	key := refreshKey(tenantID, models.PlatformLazada)

	result, err, shared := m.refreshGroup.Do(key, func() (interface{}, error) {
		return m.doRefreshLazadaToken(ctx, tenantID)
	})
	if shared {
		log.Info().Str("tenant_id", tenantID).Str("platform", models.PlatformLazada).
			Msg("[TOKEN REFRESH] Deduplicated concurrent refresh request via singleflight")
	}
	if err != nil {
		return nil, err
	}
	tokenInfo, ok := result.(*TokenInfo)
	if !ok {
		return nil, fmt.Errorf("unexpected token refresh result type: %T", result)
	}
	return tokenInfo, nil
}

func (m *TokenManager) doRefreshLazadaToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	// Get GLOBAL credentials from system schema
	globalCreds, err := m.globalConfigRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global lazada credentials: %w", err)
	}
	if globalCreds.AppKey == "" || globalCreds.AppSecret == "" {
		return nil, errors.New("lazada global credentials (appKey/appSecret) not configured in system schema")
	}

	// Get TENANT-specific tokens from credential_connections
	tokenInfo, err := m.getConnectionTokenInfo(ctx, tenantID, models.PlatformLazada)
	if err != nil {
		return nil, fmt.Errorf("failed to get lazada token info: %w", err)
	}
	if tokenInfo == nil || tokenInfo.RefreshToken == "" {
		return nil, errors.New("lazada refresh token not found in credential connections")
	}

	// Build refresh request
	lazadaService := oauth.NewLazadaOAuthService(globalCreds.AppKey, globalCreds.AppSecret, "", false)
	url, params := lazadaService.BuildRefreshTokenRequest(tokenInfo.RefreshToken)

	return m.executeLazadaTokenRefresh(ctx, tenantID, url, params)
}

// RefreshTiktokToken refreshes TikTok access token
// Uses singleflight to deduplicate concurrent refresh attempts for the same tenant.
func (m *TokenManager) RefreshTiktokToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	key := refreshKey(tenantID, models.PlatformTiktok)

	result, err, shared := m.refreshGroup.Do(key, func() (interface{}, error) {
		return m.doRefreshTiktokToken(ctx, tenantID)
	})
	if shared {
		log.Info().Str("tenant_id", tenantID).Str("platform", models.PlatformTiktok).
			Msg("[TOKEN REFRESH] Deduplicated concurrent refresh request via singleflight")
	}
	if err != nil {
		return nil, err
	}
	tokenInfo, ok := result.(*TokenInfo)
	if !ok {
		return nil, fmt.Errorf("unexpected token refresh result type: %T", result)
	}
	return tokenInfo, nil
}

func (m *TokenManager) doRefreshTiktokToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	// Get GLOBAL credentials from system schema
	globalCreds, err := m.globalConfigRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global tiktok credentials: %w", err)
	}
	if globalCreds.AppKey == "" || globalCreds.AppSecret == "" {
		return nil, errors.New("tiktok global credentials (appKey/appSecret) not configured in system schema")
	}

	// Get TENANT-specific tokens from credential_connections
	tokenInfo, err := m.getConnectionTokenInfo(ctx, tenantID, models.PlatformTiktok)
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktok token info: %w", err)
	}
	if tokenInfo == nil || tokenInfo.RefreshToken == "" {
		return nil, errors.New("tiktok refresh token not found in credential connections")
	}

	// Build refresh request
	tiktokService := oauth.NewTiktokOAuthService(globalCreds.AppKey, globalCreds.AppSecret, "", false)
	url, params := tiktokService.BuildRefreshTokenRequest(tokenInfo.RefreshToken)

	return m.executeTiktokTokenRefresh(ctx, tenantID, url, params)
}
