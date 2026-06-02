package services

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// syncTestDB creates a test DB with platform_configs and credential_connections.
func syncTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	t.Setenv("ENCRYPTION_KEY", key)
	return testutils.SetupTestPostgresWithModels(t,
		&models.CredentialConnection{},
		&repositories.TenantPlatformConfig{},
	)
}

// seedPlatformConfig inserts token config entries into platform_configs for a platform.
func seedPlatformConfig(t *testing.T, db *gorm.DB, platform string, shopID int64, accessToken string, tokenExpiryMs, refreshExpiryMs int64) {
	t.Helper()
	repo := repositories.NewTenantPlatformConfigRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.SetConfig(ctx, platform, "shopId", strconv.FormatInt(shopID, 10), false))
	require.NoError(t, repo.SetConfig(ctx, platform, "accessToken", accessToken, true))
	require.NoError(t, repo.SetConfig(ctx, platform, "refreshToken", "test-refresh-token", true))
	require.NoError(t, repo.SetConfig(ctx, platform, "tokenExpiry", strconv.FormatInt(tokenExpiryMs, 10), false))
	require.NoError(t, repo.SetConfig(ctx, platform, "refreshTokenExpiry", strconv.FormatInt(refreshExpiryMs, 10), false))
}

// seedCredentialConnection creates a credential_connection row.
func seedCredentialConnection(t *testing.T, db *gorm.DB, tenantID, platform, storeIdentifier string, tokenExpirySec, refreshExpirySec int64) {
	t.Helper()
	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	conn := &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		Platform:        platform,
		StoreIdentifier: storeIdentifier,
		StoreName:       "Test Store",
		Status:          "connected",
		AccessToken:     "old-access-token",
		RefreshToken:    "old-refresh-token",
		TokenExpiry:     tokenExpirySec,
		RefreshExpiry:   refreshExpirySec,
		Version:         1,
		CreatedBy:       "test",
		UpdatedBy:       "test",
	}
	require.NoError(t, repo.CreateConnection(ctx, conn))
}

func TestSyncCredentialTokens_NewerToken(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	tenantID := "sync_test_tenant"
	platform := models.PlatformShopee
	var shopID int64 = 99887

	// platform_configs has a token expiring in 20 minutes (ms)
	nowMs := time.Now().UnixMilli()
	platformExpiryMs := nowMs + (20 * 60 * 1000)
	platformRefreshExpiryMs := nowMs + (30 * 24 * 60 * 60 * 1000)

	seedPlatformConfig(t, db, platform, shopID, "new-access-token", platformExpiryMs, platformRefreshExpiryMs)

	// credential_connection has an OLDER token (expiring 10 min ago, in seconds)
	oldExpirySec := time.Now().Add(-10 * time.Minute).Unix()
	oldRefreshExpirySec := time.Now().Add(30 * 24 * time.Hour).Unix()
	storeIdentifier := strconv.FormatInt(shopID, 10)
	seedCredentialConnection(t, db, tenantID, platform, storeIdentifier, oldExpirySec, oldRefreshExpirySec)

	synced, err := SyncCredentialTokens(ctx, db, tenantID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, synced, 1, "at least 1 connection should be synced")

	// Verify the credential_connection was updated
	repo := repositories.NewCredentialRepository(db)
	conn, err := repo.GetConnection(ctx, tenantID, platform, storeIdentifier)
	require.NoError(t, err)
	require.NotNil(t, conn)

	assert.Equal(t, "new-access-token", conn.AccessToken)
	assert.Equal(t, "auto_sync", conn.UpdatedBy)
	// Token expiry should be updated to platformExpiryMs / 1000
	expectedExpirySec := platformExpiryMs / 1000
	assert.Equal(t, expectedExpirySec, conn.TokenExpiry)
}

func TestSyncCredentialTokens_SameToken(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	tenantID := "sync_test_tenant"
	platform := models.PlatformShopee
	var shopID int64 = 55667

	// Both platform_configs and credential_connection have the same expiry
	nowSec := time.Now().Unix()
	platformExpiryMs := nowSec * 1000 // Same time in ms
	refreshExpiryMs := (nowSec + 86400) * 1000

	seedPlatformConfig(t, db, platform, shopID, "same-access-token", platformExpiryMs, refreshExpiryMs)

	storeIdentifier := strconv.FormatInt(shopID, 10)
	seedCredentialConnection(t, db, tenantID, platform, storeIdentifier, nowSec, nowSec+86400)

	synced, err := SyncCredentialTokens(ctx, db, tenantID)
	require.NoError(t, err)
	assert.Equal(t, 0, synced, "no connections should be synced when expiry is the same")
}

func TestSyncCredentialTokens_NoConnection(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	tenantID := "sync_test_tenant"
	platform := models.PlatformLazada
	var shopID int64 = 11223

	// platform_configs has a token, but no matching credential_connection
	nowMs := time.Now().UnixMilli()
	platformExpiryMs := nowMs + (20 * 60 * 1000)
	platformRefreshExpiryMs := nowMs + (30 * 24 * 60 * 60 * 1000)

	seedPlatformConfig(t, db, platform, shopID, "new-access-token", platformExpiryMs, platformRefreshExpiryMs)
	// No seedCredentialConnection call — no matching connection exists

	synced, err := SyncCredentialTokens(ctx, db, tenantID)
	require.NoError(t, err)
	assert.Equal(t, 0, synced, "no connections should be synced when no matching connection exists")
}

// TestSyncCredentialTokens_VerifyPlatforms tests that all three platforms are processed.
func TestSyncCredentialTokens_VerifyPlatforms(t *testing.T) {
	assert.Equal(t, []string{
		models.PlatformShopee,
		models.PlatformLazada,
		models.PlatformTiktok,
	}, syncPlatforms, "syncPlatforms must include all three platforms")

	// Verify the sync iterates all platforms by checking the variable directly
	platformSet := make(map[string]bool)
	for _, p := range syncPlatforms {
		platformSet[p] = true
	}
	assert.True(t, platformSet["shopee"])
	assert.True(t, platformSet["lazada"])
	assert.True(t, platformSet["tiktok"])
}

// TestSyncPlatformTokens_SkipNoToken tests that sync skips when platform_configs has no token.
func TestSyncPlatformTokens_SkipNoToken(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	tenantID := "sync_test_tenant"

	// No platform_configs at all for any platform
	synced, err := SyncCredentialTokens(ctx, db, tenantID)
	require.NoError(t, err)
	assert.Equal(t, 0, synced, "should sync 0 when no platform_configs exist")
}
