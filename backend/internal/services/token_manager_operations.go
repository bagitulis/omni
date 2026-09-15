package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/services/sync"
	"github.com/omni/backend/internal/utils"
)

// getFirstActiveConnection looks up the first non-disabled credential
// connection for a tenant/platform. Kept private to token_manager because
// several ops (save, GetShopID) need the same lookup and calling
// GetConnection(..., "") fails scope validation.
func (m *TokenManager) getFirstActiveConnection(ctx context.Context, tenantID, platformName string) (*models.CredentialConnection, error) {
	credRepo, err := m.getCredentialRepo(tenantID)
	if err != nil {
		return nil, err
	}
	conns, err := credRepo.ListConnections(ctx, tenantID, platformName)
	if err != nil {
		return nil, err
	}
	for i := range conns {
		if conns[i].DisabledAt == nil {
			return &conns[i], nil
		}
	}
	return nil, nil
}

// saveNewTokens saves refreshed tokens to credential_connections (canonical store)
// IMPORTANT: Also invalidates cached platform clients so they will reload with new tokens
func (m *TokenManager) saveNewTokens(ctx context.Context, tenantID, platformName, accessToken, refreshToken string, expiresIn, refreshExpiresIn int64) (*TokenInfo, error) {
	credRepo, err := m.getCredentialRepo(tenantID)
	if err != nil {
		return nil, err
	}

	// Get existing connection to obtain version and current shop_cipher.
	// Bug A companion: use first-active lookup instead of the invalid
	// GetConnection(..., "") that fails scope validation.
	conn, err := m.getFirstActiveConnection(ctx, tenantID, platformName)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection for token save: %w", err)
	}
	if conn == nil {
		return nil, fmt.Errorf("no credential connection found for %s/%s", tenantID, platformName)
	}

	// Build encryption service (same key as CredentialRepository)
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		return nil, fmt.Errorf("ENCRYPTION_KEY not configured")
	}
	enc, err := utils.NewEncryptionService(encKey)
	if err != nil {
		return nil, fmt.Errorf("invalid ENCRYPTION_KEY: %w", err)
	}

	// Encrypt tokens before storing
	encryptedAccess, err := enc.Encrypt(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}
	encryptedRefresh, err := enc.Encrypt(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	// Calculate absolute expiry timestamps in milliseconds
	now := time.Now()
	accessExpiryMs := now.UnixMilli() + (expiresIn * 1000)
	refreshExpiryMs := now.UnixMilli() + (refreshExpiresIn * 1000)

	// Update connection with new tokens using optimistic locking
	updateConn := &models.CredentialConnection{
		TenantID:        tenantID,
		Platform:        platformName,
		StoreIdentifier: conn.StoreIdentifier,
		AccessToken:     encryptedAccess,
		RefreshToken:    encryptedRefresh,
		ShopCipher:      conn.ShopCipher,
		TokenExpiry:     accessExpiryMs,
		RefreshExpiry:   refreshExpiryMs,
		Status:          "connected",
	}
	err = credRepo.UpdateConnectionTokensWithVersion(ctx, updateConn, conn.Version, "connected", "refresh_success", &now)
	if err != nil {
		return nil, fmt.Errorf("failed to save tokens: %w", err)
	}

	// CRITICAL: Invalidate cached platform clients so they reload with new tokens
	platform.InvalidateTenantPlatformService(tenantID)
	sync.InvalidateTenantInstance(tenantID)

	return &TokenInfo{
		Platform:            platformName,
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		ExpiresAt:           now.Add(time.Duration(expiresIn) * time.Second),
		RefreshTokenExpires: now.Add(time.Duration(refreshExpiresIn) * time.Second),
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

// GetShopID returns the shop ID for a platform from credential_connections
func (m *TokenManager) GetShopID(ctx context.Context, tenantID, platformName string) (int64, error) {
	conn, err := m.getFirstActiveConnection(ctx, tenantID, platformName)
	if err != nil {
		return 0, err
	}
	if conn == nil {
		return 0, fmt.Errorf("no connection found for platform %s", platformName)
	}

	shopID, err := strconv.ParseInt(conn.StoreIdentifier, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid shop ID for platform %s: %s", platformName, conn.StoreIdentifier)
	}
	return shopID, nil
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
