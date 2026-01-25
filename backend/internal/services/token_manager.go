package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

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

// getTenantConfigRepo gets a TenantPlatformConfigRepository for a specific tenant
// This reads from tenant schema's PlatformConfig table (key-value format)
func (m *TokenManager) getTenantConfigRepo(tenantID string) (*repositories.TenantPlatformConfigRepository, error) {
	db, err := config.GetTenantDB(tenantID, m.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant DB for %s: %w", tenantID, err)
	}
	return repositories.NewTenantPlatformConfigRepository(db), nil
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
}

// GetTokenStatus gets token status for a tenant/platform
// Reads from tenant schema PlatformConfig table (key-value format like Node.js)
func (m *TokenManager) GetTokenStatus(ctx context.Context, tenantID, platform string) (*TokenInfo, error) {
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return nil, err
	}

	// Get token info from tenant database (key-value format)
	tokenInfo, err := tenantRepo.GetTokenInfo(ctx, platform)
	if err != nil {
		return nil, err
	}
	if tokenInfo == nil || tokenInfo.AccessToken == "" {
		return &TokenInfo{
			Platform: platform,
			IsValid:  false,
		}, nil
	}

	// Convert milliseconds to time
	expiresAt := time.UnixMilli(tokenInfo.TokenExpiry)
	refreshExpiresAt := time.UnixMilli(tokenInfo.RefreshTokenExpiry)

	// Token needs refresh if within 1 hour of expiry
	needsRefresh := time.Now().After(expiresAt.Add(-1 * time.Hour))

	return &TokenInfo{
		Platform:            platform,
		AccessToken:         tokenInfo.AccessToken,
		RefreshToken:        tokenInfo.RefreshToken,
		ExpiresAt:           expiresAt,
		RefreshTokenExpires: refreshExpiresAt,
		ShopID:              tokenInfo.ShopID,
		ShopCipher:          tokenInfo.ShopCipherOfSeller,
		IsValid:             time.Now().Before(expiresAt),
		NeedsRefresh:        needsRefresh,
	}, nil
}

// RefreshShopeeToken refreshes Shopee access token
// Gets global credentials (partnerId, partnerKey) from system schema
// Gets tenant-specific config (shopId, refreshToken) from tenant schema
func (m *TokenManager) RefreshShopeeToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	// Get GLOBAL credentials from system schema
	globalCreds, err := m.globalConfigRepo.GetShopeeCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global shopee credentials: %w", err)
	}
	if globalCreds.PartnerID == 0 || globalCreds.PartnerKey == "" {
		return nil, errors.New("shopee global credentials (partnerId/partnerKey) not configured in system schema")
	}

	// Get TENANT-specific config from tenant schema
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return nil, err
	}

	tokenInfo, err := tenantRepo.GetTokenInfo(ctx, models.PlatformShopee)
	if err != nil {
		return nil, err
	}
	if tokenInfo == nil || tokenInfo.RefreshToken == "" {
		return nil, errors.New("shopee refresh token not found in tenant config")
	}
	if tokenInfo.ShopID == 0 {
		return nil, errors.New("shopee shopId not found in tenant config")
	}

	// Build refresh request using global credentials + tenant's refresh token
	shopeeService := oauth.NewShopeeOAuthService(globalCreds.PartnerID, globalCreds.PartnerKey, "", false)
	refreshURL, reqBody := shopeeService.BuildRefreshTokenRequest(tokenInfo.RefreshToken, tokenInfo.ShopID)

	return m.executeShopeeTokenRefresh(ctx, tenantID, refreshURL, reqBody)
}

// RefreshLazadaToken refreshes Lazada access token
func (m *TokenManager) RefreshLazadaToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	// Get GLOBAL credentials from system schema
	globalCreds, err := m.globalConfigRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global lazada credentials: %w", err)
	}
	if globalCreds.AppKey == "" || globalCreds.AppSecret == "" {
		return nil, errors.New("lazada global credentials (appKey/appSecret) not configured in system schema")
	}

	// Get TENANT-specific config from tenant schema
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return nil, err
	}

	tokenInfo, err := tenantRepo.GetTokenInfo(ctx, models.PlatformLazada)
	if err != nil {
		return nil, err
	}
	if tokenInfo == nil || tokenInfo.RefreshToken == "" {
		return nil, errors.New("lazada refresh token not found in tenant config")
	}

	// Build refresh request
	lazadaService := oauth.NewLazadaOAuthService(globalCreds.AppKey, globalCreds.AppSecret, "", false)
	url, params := lazadaService.BuildRefreshTokenRequest(tokenInfo.RefreshToken)

	return m.executeLazadaTokenRefresh(ctx, tenantID, url, params)
}

// RefreshTiktokToken refreshes TikTok access token
func (m *TokenManager) RefreshTiktokToken(ctx context.Context, tenantID string) (*TokenInfo, error) {
	// Get GLOBAL credentials from system schema
	globalCreds, err := m.globalConfigRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get global tiktok credentials: %w", err)
	}
	if globalCreds.AppKey == "" || globalCreds.AppSecret == "" {
		return nil, errors.New("tiktok global credentials (appKey/appSecret) not configured in system schema")
	}

	// Get TENANT-specific config from tenant schema
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return nil, err
	}

	tokenInfo, err := tenantRepo.GetTokenInfo(ctx, models.PlatformTiktok)
	if err != nil {
		return nil, err
	}
	if tokenInfo == nil || tokenInfo.RefreshToken == "" {
		return nil, errors.New("tiktok refresh token not found in tenant config")
	}

	// Build refresh request
	tiktokService := oauth.NewTiktokOAuthService(globalCreds.AppKey, globalCreds.AppSecret, "", false)
	url, params := tiktokService.BuildRefreshTokenRequest(tokenInfo.RefreshToken)

	return m.executeTiktokTokenRefresh(ctx, tenantID, url, params)
}
