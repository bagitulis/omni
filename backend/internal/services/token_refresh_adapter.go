package services

import (
	"context"
	"fmt"
)

// TokenRefreshAdapter adapts TokenManager to platform.TokenRefreshService interface
type TokenRefreshAdapter struct {
	tokenManager *TokenManager
}

// NewTokenRefreshAdapter creates a new adapter
func NewTokenRefreshAdapter(tm *TokenManager) *TokenRefreshAdapter {
	return &TokenRefreshAdapter{tokenManager: tm}
}

// RefreshShopeeToken refreshes Shopee token and returns accessToken, refreshToken
func (a *TokenRefreshAdapter) RefreshShopeeToken(ctx context.Context, tenantID string) (string, string, error) {
	tokenInfo, err := a.tokenManager.RefreshShopeeToken(ctx, tenantID)
	if err != nil {
		return "", "", fmt.Errorf("refresh shopee token: %w", err)
	}
	if tokenInfo == nil {
		return "", "", fmt.Errorf("refresh returned nil token info")
	}
	return tokenInfo.AccessToken, tokenInfo.RefreshToken, nil
}

// RefreshLazadaToken refreshes Lazada token and returns accessToken, refreshToken
func (a *TokenRefreshAdapter) RefreshLazadaToken(ctx context.Context, tenantID string) (string, string, error) {
	tokenInfo, err := a.tokenManager.RefreshLazadaToken(ctx, tenantID)
	if err != nil {
		return "", "", fmt.Errorf("refresh lazada token: %w", err)
	}
	if tokenInfo == nil {
		return "", "", fmt.Errorf("refresh returned nil token info")
	}
	return tokenInfo.AccessToken, tokenInfo.RefreshToken, nil
}

// RefreshTiktokToken refreshes TikTok token and returns accessToken, refreshToken
func (a *TokenRefreshAdapter) RefreshTiktokToken(ctx context.Context, tenantID string) (string, string, error) {
	tokenInfo, err := a.tokenManager.RefreshTiktokToken(ctx, tenantID)
	if err != nil {
		return "", "", fmt.Errorf("refresh tiktok token: %w", err)
	}
	if tokenInfo == nil {
		return "", "", fmt.Errorf("refresh returned nil token info")
	}
	return tokenInfo.AccessToken, tokenInfo.RefreshToken, nil
}
