package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewTokenManager verifies the constructor builds a valid struct.
func TestNewTokenManager_WithBasePath(t *testing.T) {
	mgr := NewTokenManager(nil, nil, "/test/basepath")

	assert.NotNil(t, mgr)
	assert.Equal(t, "/test/basepath", mgr.basePath)
	assert.Nil(t, mgr.globalConfigRepo)
	assert.Nil(t, mgr.encryption)
	// httpClient is always initialised by the constructor
	assert.NotNil(t, mgr.httpClient)
}

// TestNewTokenManager_EmptyBasePath verifies construction with empty basePath.
func TestNewTokenManager_WithEmptyBasePath(t *testing.T) {
	mgr := NewTokenManager(nil, nil, "")

	assert.NotNil(t, mgr)
	assert.Equal(t, "", mgr.basePath)
}

// TestTokenManager_GetAllTokenStatus_NoBasePath verifies that GetAllTokenStatus returns
// an error-free map (with IsValid=false entries) when the basePath is invalid.
// It must NOT panic even with nil repos.
func TestTokenManager_GetAllTokenStatus_NoBasePath(t *testing.T) {
	mgr := NewTokenManager(nil, nil, "")

	// getTenantConfigRepo will fail because basePath="" and config.GetTenantDBWithContext
	// cannot connect, so each platform entry will be marked invalid.
	result, err := mgr.GetAllTokenStatus(nil, "test-tenant")

	// GetAllTokenStatus silently logs errors per platform and always returns the map
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// All three platforms should be present with IsValid=false
	for _, platform := range []string{"shopee", "lazada", "tiktok"} {
		info, ok := result[platform]
		assert.True(t, ok, "platform %s missing from result", platform)
		if ok {
			assert.False(t, info.IsValid, "expected IsValid=false for platform %s", platform)
		}
	}
}

// TestTokenInfo_Fields verifies TokenInfo struct fields are accessible.
func TestTokenInfo_Fields(t *testing.T) {
	info := &TokenInfo{
		Platform:     "shopee",
		AccessToken:  "at-123",
		RefreshToken: "rt-456",
		ShopID:       987654,
		IsValid:      true,
		NeedsRefresh: false,
	}

	assert.Equal(t, "shopee", info.Platform)
	assert.Equal(t, "at-123", info.AccessToken)
	assert.Equal(t, "rt-456", info.RefreshToken)
	assert.Equal(t, int64(987654), info.ShopID)
	assert.True(t, info.IsValid)
	assert.False(t, info.NeedsRefresh)
}
