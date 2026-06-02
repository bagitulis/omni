package services

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/services/sync"
)

// saveNewTokens saves refreshed tokens to tenant database using key-value format
// This matches Node.js behavior: saves to PlatformConfig table with configKey/configValue
// IMPORTANT: Also invalidates cached platform clients so they will reload with new tokens
func (m *TokenManager) saveNewTokens(ctx context.Context, tenantID, platformName, accessToken, refreshToken string, expiresIn, refreshExpiresIn int64) (*TokenInfo, error) {
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return nil, err
	}

	// Save tokens using the key-value format (matches Node.js)
	err = tenantRepo.UpdateTokens(ctx, platformName, accessToken, refreshToken, expiresIn, refreshExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("failed to save tokens: %w", err)
	}

	// CRITICAL: Invalidate cached platform clients so they reload with new tokens
	// Without this, the old token would still be used until server restart
	platform.InvalidateTenantPlatformService(tenantID)
	sync.InvalidateTenantInstance(tenantID)

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	refreshExpiresAt := time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)

	return &TokenInfo{
		Platform:            platformName,
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		ExpiresAt:           expiresAt,
		RefreshTokenExpires: refreshExpiresAt,
		IsValid:             true,
		NeedsRefresh:        false,
		RefreshExpired:      false,
	}, nil
}

// GetAllTokenStatus returns token status for all platforms for a tenant
func (m *TokenManager) GetAllTokenStatus(ctx context.Context, tenantID string) (map[string]*TokenInfo, error) {
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}
	result := make(map[string]*TokenInfo)

	for _, p := range platforms {
		info, err := m.GetTokenStatus(ctx, tenantID, p)
		if err != nil {
			// Log but continue with other platforms
			result[p] = &TokenInfo{Platform: p, IsValid: false}
			continue
		}
		result[p] = info
	}

	return result, nil
}

// RefreshExpiredTokens refreshes all expired tokens for a tenant
func (m *TokenManager) RefreshExpiredTokens(ctx context.Context, tenantID string) (map[string]bool, error) {
	results := make(map[string]bool)

	// Check and refresh Shopee
	if status, _ := m.GetTokenStatus(ctx, tenantID, models.PlatformShopee); status != nil && status.NeedsRefresh {
		_, err := m.RefreshShopeeToken(ctx, tenantID)
		results[models.PlatformShopee] = err == nil
	}

	// Check and refresh Lazada
	if status, _ := m.GetTokenStatus(ctx, tenantID, models.PlatformLazada); status != nil && status.NeedsRefresh {
		_, err := m.RefreshLazadaToken(ctx, tenantID)
		results[models.PlatformLazada] = err == nil
	}

	// Check and refresh TikTok
	if status, _ := m.GetTokenStatus(ctx, tenantID, models.PlatformTiktok); status != nil && status.NeedsRefresh {
		_, err := m.RefreshTiktokToken(ctx, tenantID)
		results[models.PlatformTiktok] = err == nil
	}

	return results, nil
}

// GetShopID returns the shop ID for a platform from tenant config
func (m *TokenManager) GetShopID(ctx context.Context, tenantID, platformName string) (int64, error) {
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return 0, err
	}

	tokenInfo, err := tenantRepo.GetTokenInfo(ctx, platformName)
	if err != nil {
		return 0, err
	}
	if tokenInfo == nil {
		return 0, fmt.Errorf("no config found for platform %s", platformName)
	}

	return tokenInfo.ShopID, nil
}

// GetAccessToken returns decrypted access token for a platform
func (m *TokenManager) GetAccessToken(ctx context.Context, tenantID, platformName string) (string, error) {
	status, err := m.GetTokenStatus(ctx, tenantID, platformName)
	if err != nil {
		return "", err
	}
	if !status.IsValid {
		return "", fmt.Errorf("token is invalid or expired for platform %s", platformName)
	}
	return status.AccessToken, nil
}
