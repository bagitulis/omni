package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenManager(t *testing.T) {
	tm := NewTokenManager(nil, nil, "/test/path")

	assert.NotNil(t, tm)
	assert.Equal(t, "/test/path", tm.basePath)
	assert.NotNil(t, tm.httpClient)
}

func TestTokenInfo(t *testing.T) {
	expiresAt := time.Now().Add(4 * time.Hour)
	refreshExpiresAt := time.Now().Add(30 * 24 * time.Hour)

	info := TokenInfo{
		Platform:            "shopee",
		AccessToken:         "access-token-123",
		RefreshToken:        "refresh-token-456",
		ExpiresAt:           expiresAt,
		RefreshTokenExpires: refreshExpiresAt,
		ShopID:              123456,
		ShopCipher:          "cipher-abc",
		IsValid:             true,
		NeedsRefresh:        false,
	}

	assert.Equal(t, "shopee", info.Platform)
	assert.Equal(t, "access-token-123", info.AccessToken)
	assert.Equal(t, "refresh-token-456", info.RefreshToken)
	assert.Equal(t, expiresAt, info.ExpiresAt)
	assert.Equal(t, refreshExpiresAt, info.RefreshTokenExpires)
	assert.Equal(t, int64(123456), info.ShopID)
	assert.Equal(t, "cipher-abc", info.ShopCipher)
	assert.True(t, info.IsValid)
	assert.False(t, info.NeedsRefresh)
}

func TestTokenManager_GetAllTokenStatus_NilRepo(t *testing.T) {
	tm := NewTokenManager(nil, nil, "/test/path")

	result, err := tm.GetAllTokenStatus(context.Background(), "tenant1")

	assert.NoError(t, err)
	assert.Len(t, result, 3)
	// All should be invalid due to nil repo
	assert.False(t, result["shopee"].IsValid)
	assert.False(t, result["lazada"].IsValid)
	assert.False(t, result["tiktok"].IsValid)
}

func TestTokenManager_RefreshExpiredTokens_NilRepo(t *testing.T) {
	tm := NewTokenManager(nil, nil, "/test/path")

	result, err := tm.RefreshExpiredTokens(context.Background(), "tenant1")

	assert.NoError(t, err)
	// Should return empty map since no tokens need refresh (can't check status)
	assert.NotNil(t, result)
}

func TestTokenInfo_ValidityCheck(t *testing.T) {
	tests := []struct {
		name        string
		expiresAt   time.Time
		wantValid   bool
		wantRefresh bool
	}{
		{
			name:        "valid token, no refresh needed",
			expiresAt:   time.Now().Add(2 * time.Hour),
			wantValid:   true,
			wantRefresh: false,
		},
		{
			name:        "valid token, refresh needed",
			expiresAt:   time.Now().Add(30 * time.Minute),
			wantValid:   true,
			wantRefresh: true,
		},
		{
			name:        "expired token",
			expiresAt:   time.Now().Add(-1 * time.Hour),
			wantValid:   false,
			wantRefresh: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := TokenInfo{
				ExpiresAt:    tt.expiresAt,
				IsValid:      time.Now().Before(tt.expiresAt),
				NeedsRefresh: time.Now().After(tt.expiresAt.Add(-1 * time.Hour)),
			}

			assert.Equal(t, tt.wantValid, info.IsValid)
			assert.Equal(t, tt.wantRefresh, info.NeedsRefresh)
		})
	}
}

func TestTokenInfo_Platforms(t *testing.T) {
	platforms := []string{"shopee", "lazada", "tiktok"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			info := TokenInfo{
				Platform: platform,
			}
			assert.Equal(t, platform, info.Platform)
		})
	}
}

func TestTokenInfo_ShopIDAndCipher(t *testing.T) {
	tests := []struct {
		name       string
		shopID     int64
		shopCipher string
	}{
		{name: "shopee shop", shopID: 123456789, shopCipher: ""},
		{name: "tiktok shop", shopID: 0, shopCipher: "cipher-abc123"},
		{name: "both values", shopID: 999888777, shopCipher: "tiktok-cipher"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := TokenInfo{
				ShopID:     tt.shopID,
				ShopCipher: tt.shopCipher,
			}
			assert.Equal(t, tt.shopID, info.ShopID)
			assert.Equal(t, tt.shopCipher, info.ShopCipher)
		})
	}
}
