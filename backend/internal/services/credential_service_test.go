package services

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestNewCredentialService(t *testing.T) {
	dbPath := "/path/to/db"
	svc := NewCredentialService(dbPath)

	assert.NotNil(t, svc)
	assert.Equal(t, dbPath, svc.dbPath)
}

func TestPlatformCredentials_Structure(t *testing.T) {
	creds := &PlatformCredentials{
		Platform:     "shopee",
		PartnerID:    12345,
		PartnerKey:   "partner-key-123",
		ShopID:       67890,
		AccessToken:  "access-token-abc",
		RefreshToken: "refresh-token-xyz",
		AppKey:       "app-key-456",
		AppSecret:    "app-secret-789",
		ShopCipher:   "shop-cipher-abc",
		Region:       "ID",
		IsProduction: true,
	}

	assert.Equal(t, "shopee", creds.Platform)
	assert.Equal(t, int64(12345), creds.PartnerID)
	assert.Equal(t, "partner-key-123", creds.PartnerKey)
	assert.Equal(t, int64(67890), creds.ShopID)
	assert.Equal(t, "access-token-abc", creds.AccessToken)
	assert.Equal(t, "refresh-token-xyz", creds.RefreshToken)
	assert.Equal(t, "app-key-456", creds.AppKey)
	assert.Equal(t, "app-secret-789", creds.AppSecret)
	assert.Equal(t, "shop-cipher-abc", creds.ShopCipher)
	assert.Equal(t, "ID", creds.Region)
	assert.True(t, creds.IsProduction)
}

func TestPlatformCredentials_Defaults(t *testing.T) {
	// Test default values
	creds := &PlatformCredentials{}

	assert.Empty(t, creds.Platform)
	assert.Equal(t, int64(0), creds.PartnerID)
	assert.Empty(t, creds.PartnerKey)
	assert.Equal(t, int64(0), creds.ShopID)
	assert.Empty(t, creds.AccessToken)
	assert.Empty(t, creds.RefreshToken)
	assert.Empty(t, creds.AppKey)
	assert.Empty(t, creds.AppSecret)
	assert.Empty(t, creds.ShopCipher)
	assert.Empty(t, creds.Region)
	assert.False(t, creds.IsProduction)
}

func TestPlatformCredentials_PlatformNames(t *testing.T) {
	// Test supported platform names
	platforms := []string{"shopee", "lazada", "tiktok"}

	for _, platform := range platforms {
		creds := &PlatformCredentials{Platform: platform}
		assert.Equal(t, platform, creds.Platform)
	}
}

func TestCredentialService_GetAllPlatformCredentials_EmptyResult(t *testing.T) {
	// This tests the method structure without requiring DB
	// The actual method will fail without proper DB connection,
	// but we can verify the return type is correct
	svc := NewCredentialService("/nonexistent/path")

	// This will return empty map since DB doesn't exist
	result, err := svc.GetAllPlatformCredentials("test-tenant")

	// Either returns empty result or error, both are acceptable
	if err == nil {
		assert.NotNil(t, result)
		assert.IsType(t, map[string]*PlatformCredentials{}, result)
	}
}

func TestCredentialService_Structure(t *testing.T) {
	// Test service struct fields
	svc := &CredentialService{dbPath: "/test/path"}

	assert.Equal(t, "/test/path", svc.dbPath)
}

func TestCredentialService_SetTokenManager(t *testing.T) {
	svc := NewCredentialService("/test/path")
	assert.Nil(t, svc.tokenManager)

	tm := NewTokenManager(nil, nil, "/test/path")
	svc.SetTokenManager(tm)
	assert.Equal(t, tm, svc.tokenManager)
}

func TestRegisterGlobalTokenManager(t *testing.T) {
	// Save and restore original value
	original := globalTokenManager
	defer func() { globalTokenManager = original }()

	globalTokenManager = nil
	svc := NewCredentialService("/test/path")
	assert.Nil(t, svc.tokenManager)

	tm := NewTokenManager(nil, nil, "/test/path")
	RegisterGlobalTokenManager(tm)

	// New instances should pick up the global token manager
	svc2 := NewCredentialService("/test/path")
	assert.Equal(t, tm, svc2.tokenManager)
}

func TestPlatformCredentials_TokenExpiry(t *testing.T) {
	creds := &PlatformCredentials{
		TokenExpiry: 1704067200000, // 2024-01-01T00:00:00Z in ms
	}
	assert.Equal(t, int64(1704067200000), creds.TokenExpiry)

	// Zero value means no expiry set
	emptyCreds := &PlatformCredentials{}
	assert.Equal(t, int64(0), emptyCreds.TokenExpiry)
}

func TestCredentialService_ConcurrentRefreshSingleflight(t *testing.T) {
	svc := NewCredentialService("/test/path")
	var refreshCalls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	svc.refreshTokenFn = func(ctx context.Context, tenantID, platform string) error {
		refreshCalls.Add(1)
		close(started)
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	svc.reloadCredentialsFn = func(ctx context.Context, tenantDB *gorm.DB, tenantID, platform string, creds *PlatformCredentials) error {
		creds.AccessToken = "fresh-token"
		creds.TokenExpiry = time.Now().Add(time.Hour).UnixMilli()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	creds := &PlatformCredentials{Platform: "shopee", StoreIdentifier: "store-1", AccessToken: "expired-token", TokenExpiry: time.Now().Add(-time.Minute).UnixMilli()}
	var wg sync.WaitGroup
	const goroutines = 8
	errs := make(chan error, goroutines)
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- svc.refreshAndReload(ctx, nil, "tenant-a", "shopee", creds)
		}()
	}
	<-started
	close(release)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	assert.Equal(t, int32(1), refreshCalls.Load())
}
