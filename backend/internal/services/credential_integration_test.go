package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCredentialIntegrationTest creates a test PostgreSQL with credential tables
// and returns a CredentialRepository for direct assertion.
func setupCredentialIntegrationTest(t *testing.T) (*repositories.CredentialRepository, context.Context) {
	t.Helper()
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	t.Setenv("ENCRYPTION_KEY", key)

	db := testutils.SetupTestPostgresWithModels(t,
		&models.CredentialAppConfig{},
		&models.CredentialConnection{},
	)
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_credential_connections_active
		ON credential_connections (tenant_id, platform, store_identifier)
		WHERE disabled_at IS NULL`).Error)

	repo := repositories.NewCredentialRepository(db)
	return repo, context.Background()
}

// insertShopeeAppConfig inserts a Shopee app config with partnerId=2011782.
func insertShopeeAppConfig(t *testing.T, repo *repositories.CredentialRepository, ctx context.Context, tenantID string) {
	t.Helper()
	require.NoError(t, repo.CreateAppConfig(ctx, &models.CredentialAppConfig{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Platform:   models.PlatformShopee,
		PartnerID:  2011782,
		PartnerKey: "shopee-partner-key-secret",
		Configured: true,
		CreatedBy:  "test",
		UpdatedBy:  "test",
	}))
}

// insertLazadaAppConfig inserts a Lazada app config with appKey/appSecret.
func insertLazadaAppConfig(t *testing.T, repo *repositories.CredentialRepository, ctx context.Context, tenantID string) {
	t.Helper()
	require.NoError(t, repo.CreateAppConfig(ctx, &models.CredentialAppConfig{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Platform:   models.PlatformLazada,
		AppKey:     "lazada-app-key-123",
		AppSecret:  "lazada-app-secret-456",
		Configured: true,
		Region:     "id",
		CreatedBy:  "test",
		UpdatedBy:  "test",
	}))
}

// insertTikTokAppConfig inserts a TikTok app config with appKey/appSecret.
func insertTikTokAppConfig(t *testing.T, repo *repositories.CredentialRepository, ctx context.Context, tenantID string) {
	t.Helper()
	require.NoError(t, repo.CreateAppConfig(ctx, &models.CredentialAppConfig{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Platform:   models.PlatformTiktok,
		AppKey:     "tiktok-app-key-789",
		AppSecret:  "tiktok-app-secret-012",
		Configured: true,
		CreatedBy:  "test",
		UpdatedBy:  "test",
	}))
}

// insertShopeeConnection inserts a Shopee store connection with access token.
func insertShopeeConnection(t *testing.T, repo *repositories.CredentialRepository, ctx context.Context, tenantID, storeID string) {
	t.Helper()
	require.NoError(t, repo.CreateConnection(ctx, &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		Platform:        models.PlatformShopee,
		StoreIdentifier: storeID,
		StoreName:       "Test Shopee Store",
		Status:          "connected",
		AccessToken:     "shopee-access-token-abc",
		RefreshToken:    "shopee-refresh-token-def",
		TokenExpiry:     time.Now().Add(4 * time.Hour).UnixMilli(),
		RefreshExpiry:   time.Now().Add(30 * 24 * time.Hour).UnixMilli(),
		Version:         1,
		CreatedBy:       "test",
		UpdatedBy:       "test",
	}))
}

// insertLazadaConnection inserts a Lazada store connection with access token.
func insertLazadaConnection(t *testing.T, repo *repositories.CredentialRepository, ctx context.Context, tenantID, storeID string) {
	t.Helper()
	require.NoError(t, repo.CreateConnection(ctx, &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		Platform:        models.PlatformLazada,
		StoreIdentifier: storeID,
		StoreName:       "Test Lazada Store",
		Status:          "connected",
		Region:          "id",
		AccessToken:     "lazada-access-token-ghi",
		RefreshToken:    "lazada-refresh-token-jkl",
		TokenExpiry:     time.Now().Add(2 * time.Hour).UnixMilli(),
		RefreshExpiry:   time.Now().Add(30 * 24 * time.Hour).UnixMilli(),
		Version:         1,
		CreatedBy:       "test",
		UpdatedBy:       "test",
	}))
}

// insertTikTokConnection inserts a TikTok store connection with access token and shopCipher.
func insertTikTokConnection(t *testing.T, repo *repositories.CredentialRepository, ctx context.Context, tenantID, storeID string) {
	t.Helper()
	require.NoError(t, repo.CreateConnection(ctx, &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		Platform:        models.PlatformTiktok,
		StoreIdentifier: storeID,
		StoreName:       "Test TikTok Store",
		Status:          "connected",
		AccessToken:     "tiktok-access-token-mno",
		RefreshToken:    "tiktok-refresh-token-pqr",
		ShopCipher:      "tiktok-shop-cipher-stu",
		TokenExpiry:     time.Now().Add(1 * time.Hour).UnixMilli(),
		RefreshExpiry:   time.Now().Add(30 * 24 * time.Hour).UnixMilli(),
		Version:         1,
		CreatedBy:       "test",
		UpdatedBy:       "test",
	}))
}

// TestIntegrationCredentialChain verifies the full credential resolution chain
// for all 3 platforms using canonical tables (credential_app_configs + credential_connections).
func TestIntegrationCredentialChain(t *testing.T) {
	t.Run("Shopee happy path", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantID := "tenant-shopee-happy"

		// Insert app config + connection
		insertShopeeAppConfig(t, repo, ctx, tenantID)
		insertShopeeConnection(t, repo, ctx, tenantID, "123456789")

		// Verify app config resolves correctly
		appCfg, err := repo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		require.NotNil(t, appCfg, "app config should exist")
		assert.Equal(t, int64(2011782), appCfg.PartnerID, "partnerId must be 2011782")
		assert.Equal(t, "shopee-partner-key-secret", appCfg.PartnerKey)
		assert.True(t, appCfg.Configured)

		// Verify connection resolves correctly
		conns, err := repo.ListConnections(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		require.Len(t, conns, 1, "should have exactly 1 active connection")
		assert.Equal(t, "shopee-access-token-abc", conns[0].AccessToken)
		assert.Equal(t, "shopee-refresh-token-def", conns[0].RefreshToken)
		assert.Equal(t, "123456789", conns[0].StoreIdentifier)
		assert.Equal(t, "connected", conns[0].Status)
	})

	t.Run("Lazada happy path", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantID := "tenant-lazada-happy"

		// Insert app config + connection
		insertLazadaAppConfig(t, repo, ctx, tenantID)
		insertLazadaConnection(t, repo, ctx, tenantID, "lazada-store-001")

		// Verify app config resolves correctly
		appCfg, err := repo.GetAppConfig(ctx, tenantID, models.PlatformLazada)
		require.NoError(t, err)
		require.NotNil(t, appCfg, "app config should exist")
		assert.Equal(t, "lazada-app-key-123", appCfg.AppKey)
		assert.Equal(t, "lazada-app-secret-456", appCfg.AppSecret)
		assert.Equal(t, "id", appCfg.Region)
		assert.True(t, appCfg.Configured)

		// Verify connection resolves correctly
		conns, err := repo.ListConnections(ctx, tenantID, models.PlatformLazada)
		require.NoError(t, err)
		require.Len(t, conns, 1, "should have exactly 1 active connection")
		assert.Equal(t, "lazada-access-token-ghi", conns[0].AccessToken)
		assert.Equal(t, "lazada-refresh-token-jkl", conns[0].RefreshToken)
		assert.Equal(t, "lazada-store-001", conns[0].StoreIdentifier)
		assert.Equal(t, "id", conns[0].Region)
	})

	t.Run("TikTok happy path", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantID := "tenant-tiktok-happy"

		// Insert app config + connection (TikTok has shopCipher)
		insertTikTokAppConfig(t, repo, ctx, tenantID)
		insertTikTokConnection(t, repo, ctx, tenantID, "tiktok-store-999")

		// Verify app config resolves correctly
		appCfg, err := repo.GetAppConfig(ctx, tenantID, models.PlatformTiktok)
		require.NoError(t, err)
		require.NotNil(t, appCfg, "app config should exist")
		assert.Equal(t, "tiktok-app-key-789", appCfg.AppKey)
		assert.Equal(t, "tiktok-app-secret-012", appCfg.AppSecret)
		assert.True(t, appCfg.Configured)

		// Verify connection resolves correctly (including shopCipher)
		conns, err := repo.ListConnections(ctx, tenantID, models.PlatformTiktok)
		require.NoError(t, err)
		require.Len(t, conns, 1, "should have exactly 1 active connection")
		assert.Equal(t, "tiktok-access-token-mno", conns[0].AccessToken)
		assert.Equal(t, "tiktok-refresh-token-pqr", conns[0].RefreshToken)
		assert.Equal(t, "tiktok-shop-cipher-stu", conns[0].ShopCipher, "TikTok must have shopCipher")
		assert.Equal(t, "tiktok-store-999", conns[0].StoreIdentifier)
	})

	t.Run("missing app_config returns nil", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantID := "tenant-no-app-config"

		// Insert only connection, no app config
		insertShopeeConnection(t, repo, ctx, tenantID, "orphan-store")

		// GetAppConfig should return nil, nil (not found)
		appCfg, err := repo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err, "GetAppConfig should not error on missing record")
		assert.Nil(t, appCfg, "app config should be nil when not inserted")

		// Connection still exists independently
		conns, err := repo.ListConnections(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		assert.Len(t, conns, 1, "connection should exist even without app config")
	})

	t.Run("missing connection returns empty list", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantID := "tenant-no-connection"

		// Insert only app config, no connection
		insertShopeeAppConfig(t, repo, ctx, tenantID)

		// GetAppConfig should succeed
		appCfg, err := repo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		require.NotNil(t, appCfg, "app config should exist")
		assert.Equal(t, int64(2011782), appCfg.PartnerID)

		// ListConnections should return empty (no connection inserted)
		conns, err := repo.ListConnections(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		assert.Empty(t, conns, "connections should be empty when none inserted")
	})

	t.Run("disabled connection is excluded", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantID := "tenant-disabled-conn"

		insertShopeeAppConfig(t, repo, ctx, tenantID)
		insertShopeeConnection(t, repo, ctx, tenantID, "disabled-store")

		// Verify connection exists first
		conns, err := repo.ListConnections(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		require.Len(t, conns, 1)

		// Disable the connection
		conn, err := repo.GetConnection(ctx, tenantID, models.PlatformShopee, "disabled-store")
		require.NoError(t, err)
		require.NotNil(t, conn)
		require.NoError(t, repo.DisableConnectionWithVersion(ctx, conn, conn.Version, "test_disable"))

		// ListConnections should now exclude disabled
		conns, err = repo.ListConnections(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		assert.Empty(t, conns, "disabled connections should be excluded from ListConnections")
	})

	t.Run("all 3 platforms for same tenant", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantID := "tenant-multi-platform"

		// Insert all 3 platforms
		insertShopeeAppConfig(t, repo, ctx, tenantID)
		insertShopeeConnection(t, repo, ctx, tenantID, "shopee-shop-1")
		insertLazadaAppConfig(t, repo, ctx, tenantID)
		insertLazadaConnection(t, repo, ctx, tenantID, "lazada-shop-1")
		insertTikTokAppConfig(t, repo, ctx, tenantID)
		insertTikTokConnection(t, repo, ctx, tenantID, "tiktok-shop-1")

		// Verify each platform resolves independently
		shopeeApp, err := repo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		require.NotNil(t, shopeeApp)
		assert.Equal(t, int64(2011782), shopeeApp.PartnerID)

		lazadaApp, err := repo.GetAppConfig(ctx, tenantID, models.PlatformLazada)
		require.NoError(t, err)
		require.NotNil(t, lazadaApp)
		assert.Equal(t, "lazada-app-key-123", lazadaApp.AppKey)

		tiktokApp, err := repo.GetAppConfig(ctx, tenantID, models.PlatformTiktok)
		require.NoError(t, err)
		require.NotNil(t, tiktokApp)
		assert.Equal(t, "tiktok-app-key-789", tiktokApp.AppKey)

		// Verify connections are platform-scoped
		shopeeConns, err := repo.ListConnections(ctx, tenantID, models.PlatformShopee)
		require.NoError(t, err)
		require.Len(t, shopeeConns, 1)
		assert.Equal(t, "shopee-access-token-abc", shopeeConns[0].AccessToken)

		lazadaConns, err := repo.ListConnections(ctx, tenantID, models.PlatformLazada)
		require.NoError(t, err)
		require.Len(t, lazadaConns, 1)
		assert.Equal(t, "lazada-access-token-ghi", lazadaConns[0].AccessToken)

		tiktokConns, err := repo.ListConnections(ctx, tenantID, models.PlatformTiktok)
		require.NoError(t, err)
		require.Len(t, tiktokConns, 1)
		assert.Equal(t, "tiktok-shop-cipher-stu", tiktokConns[0].ShopCipher)
	})

	t.Run("tenant isolation", func(t *testing.T) {
		repo, ctx := setupCredentialIntegrationTest(t)
		tenantA := "tenant-alpha"
		tenantB := "tenant-beta"

		// Insert Shopee config for tenant A only
		insertShopeeAppConfig(t, repo, ctx, tenantA)
		insertShopeeConnection(t, repo, ctx, tenantA, "alpha-shop")

		// Tenant B should see nothing
		appCfg, err := repo.GetAppConfig(ctx, tenantB, models.PlatformShopee)
		require.NoError(t, err)
		assert.Nil(t, appCfg, "tenant B should not see tenant A's app config")

		conns, err := repo.ListConnections(ctx, tenantB, models.PlatformShopee)
		require.NoError(t, err)
		assert.Empty(t, conns, "tenant B should not see tenant A's connections")

		// Tenant A should see its own data
		appCfg, err = repo.GetAppConfig(ctx, tenantA, models.PlatformShopee)
		require.NoError(t, err)
		require.NotNil(t, appCfg)
		assert.Equal(t, int64(2011782), appCfg.PartnerID)
	})
}
