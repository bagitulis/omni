package services

import (
	"context"
	"strconv"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// =============================================================================
// Stub Functions — RED Phase (connections backfill not yet implemented)
// =============================================================================

// BackfillConnections backfills credential_connections from platform_configs.
// RED phase: stub does nothing. Tests expecting row creation will fail.
func BackfillConnections(ctx context.Context, db *gorm.DB, tenantID string) error {
	return nil
}

// =============================================================================
// Helpers
// =============================================================================

func backfillTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	t.Setenv("ENCRYPTION_KEY", key)
	return testutils.SetupTestPostgresWithModels(t,
		&models.CredentialConnection{},
		&models.CredentialAppConfig{},
		&repositories.TenantPlatformConfig{},
	)
}

// seedBackfillAppData populates platform_configs with app-level config for a platform.
func seedBackfillAppData(t *testing.T, db *gorm.DB, platform string, partnerID int64, partnerKey, appKey, appSecret string) {
	t.Helper()
	repo := repositories.NewTenantPlatformConfigRepository(db)
	ctx := context.Background()
	if partnerID > 0 {
		require.NoError(t, repo.SetConfig(ctx, platform, "partnerId", strconv.FormatInt(partnerID, 10), false))
	}
	if partnerKey != "" {
		require.NoError(t, repo.SetConfig(ctx, platform, "partnerKey", partnerKey, false))
	}
	if appKey != "" {
		require.NoError(t, repo.SetConfig(ctx, platform, "appKey", appKey, false))
	}
	if appSecret != "" {
		require.NoError(t, repo.SetConfig(ctx, platform, "appSecret", appSecret, true))
	}
}

// seedBackfillConnData populates platform_configs with connection-level config for a platform.
func seedBackfillConnData(t *testing.T, db *gorm.DB, platform string, shopID int64, shopName, accessToken, refreshToken, region string) {
	t.Helper()
	repo := repositories.NewTenantPlatformConfigRepository(db)
	ctx := context.Background()
	require.NoError(t, repo.SetConfig(ctx, platform, "shopId", strconv.FormatInt(shopID, 10), false))
	require.NoError(t, repo.SetConfig(ctx, platform, "shopName", shopName, false))
	require.NoError(t, repo.SetConfig(ctx, platform, "accessToken", accessToken, true))
	require.NoError(t, repo.SetConfig(ctx, platform, "refreshToken", refreshToken, true))
	if region != "" {
		require.NoError(t, repo.SetConfig(ctx, platform, "region", region, false))
	}
}

// =============================================================================
// Tests — RED Phase (all tests must fail)
// =============================================================================

// TestBackfillAppConfigs_EmptyCanonical seeds platform_configs with app-level
// config for 3 platforms and verifies BackfillAppConfigs populates the
// credential_app_configs table.
func TestBackfillAppConfigs_EmptyCanonical(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "backfill_app_test"

	// Seed app-level configs for all 3 platforms
	seedBackfillAppData(t, db, models.PlatformShopee, 1001, "shopee-partner-key", "", "")
	seedBackfillAppData(t, db, models.PlatformLazada, 0, "", "lazada-app-key", "lazada-app-secret")
	seedBackfillAppData(t, db, models.PlatformTiktok, 0, "", "tiktok-app-key", "tiktok-app-secret")

	// Call backfill — RED phase: stub returns nil, does nothing
	err := BackfillAppConfigs(ctx, db, tenantID)
	require.NoError(t, err)

	// Assert canonical rows were created
	credRepo := repositories.NewCredentialRepository(db)

	shopeeCfg, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, shopeeCfg, "Shopee app config should exist — RED: stub did not create it")
	assert.Equal(t, int64(1001), shopeeCfg.PartnerID)
	assert.True(t, shopeeCfg.Configured)

	lazadaCfg, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformLazada)
	require.NoError(t, err)
	require.NotNil(t, lazadaCfg, "Lazada app config should exist — RED: stub did not create it")
	assert.True(t, lazadaCfg.Configured)

	tiktokCfg, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformTiktok)
	require.NoError(t, err)
	require.NotNil(t, tiktokCfg, "TikTok app config should exist — RED: stub did not create it")
	assert.True(t, tiktokCfg.Configured)
}

// TestBackfillConnections_EmptyCanonical seeds platform_configs with connection
// data and verifies BackfillConnections populates credential_connections.
func TestBackfillConnections_EmptyCanonical(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "backfill_conn_test"

	// Seed connection configs for 2 platforms
	seedBackfillConnData(t, db, models.PlatformShopee, 99887, "Shopee Store", "shopee-access-token", "shopee-refresh-token", "id")
	seedBackfillConnData(t, db, models.PlatformLazada, 55667, "Lazada Store", "lazada-access-token", "lazada-refresh-token", "sg")

	// Call backfill — RED phase: stub returns nil, does nothing
	err := BackfillConnections(ctx, db, tenantID)
	require.NoError(t, err)

	// Assert canonical rows were created
	credRepo := repositories.NewCredentialRepository(db)

	shopeeConn, err := credRepo.GetConnection(ctx, tenantID, models.PlatformShopee, "99887")
	require.NoError(t, err)
	require.NotNil(t, shopeeConn, "Shopee connection should exist — RED: stub did not create it")
	assert.Equal(t, "connected", shopeeConn.Status)
	assert.Equal(t, "Shopee Store", shopeeConn.StoreName)

	lazadaConn, err := credRepo.GetConnection(ctx, tenantID, models.PlatformLazada, "55667")
	require.NoError(t, err)
	require.NotNil(t, lazadaConn, "Lazada connection should exist — RED: stub did not create it")
	assert.Equal(t, "connected", lazadaConn.Status)
	assert.Equal(t, "sg", lazadaConn.Region)
}

// TestBackfill_Idempotent verifies that running backfill twice does not
// duplicate rows in canonical tables.
func TestBackfill_Idempotent(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "backfill_idempotent_test"

	// Seed data for 3 platforms
	seedBackfillAppData(t, db, models.PlatformShopee, 1001, "shopee-partner-key", "", "")
	seedBackfillAppData(t, db, models.PlatformLazada, 0, "", "lazada-app-key", "lazada-app-secret")
	seedBackfillAppData(t, db, models.PlatformTiktok, 0, "", "tiktok-app-key", "tiktok-app-secret")
	seedBackfillConnData(t, db, models.PlatformShopee, 99887, "Shopee Store", "shopee-access-token", "shopee-refresh-token", "id")
	seedBackfillConnData(t, db, models.PlatformLazada, 55667, "Lazada Store", "lazada-access-token", "lazada-refresh-token", "sg")

	// First run — RED: stub does nothing, no rows created
	err := BackfillAppConfigs(ctx, db, tenantID)
	require.NoError(t, err)
	err = BackfillConnections(ctx, db, tenantID)
	require.NoError(t, err)

	// Count rows after first run
	credRepo := repositories.NewCredentialRepository(db)
	conns, err := credRepo.ListConnections(ctx, tenantID, "")
	require.NoError(t, err)
	firstConnCount := len(conns)

	// Second run
	err = BackfillAppConfigs(ctx, db, tenantID)
	require.NoError(t, err)
	err = BackfillConnections(ctx, db, tenantID)
	require.NoError(t, err)

	// Count rows after second run — should not increase
	conns2, err := credRepo.ListConnections(ctx, tenantID, "")
	require.NoError(t, err)
	secondConnCount := len(conns2)

	// RED phase: both counts are 0, so the equality assertion passes (0 == 0).
	// We assert that rows WERE created to trigger failure.
	assert.Greater(t, firstConnCount, 0, "First run should create connections — RED: stub created none")
	assert.Equal(t, firstConnCount, secondConnCount, "Second run should not add duplicate connections")
}

// TestBackfill_EmptyTenant verifies backfill handles tenants with zero
// platform_configs rows gracefully (tika_nusseyba scenario).
func TestBackfill_EmptyTenant(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "tika_nusseyba"

	// No platform_configs seeded — simulating tenant with empty platform_configs

	// Backfill should handle gracefully (no panic, no error, no rows)
	err := BackfillAppConfigs(ctx, db, tenantID)
	require.NoError(t, err)

	err = BackfillConnections(ctx, db, tenantID)
	require.NoError(t, err)

	// Verify no canonical rows were created
	credRepo := repositories.NewCredentialRepository(db)

	var appCfgs []models.CredentialAppConfig
	require.NoError(t, db.WithContext(ctx).Find(&appCfgs).Error)
	assert.Len(t, appCfgs, 0, "No app configs should exist for empty tenant")

	conns, err := credRepo.ListConnections(ctx, tenantID, "")
	require.NoError(t, err)
	assert.Len(t, conns, 0, "No connections should exist for empty tenant")
}

// TestBackfill_MissingRequiredKeys verifies backfill returns an error when
// platform_configs rows are missing required fields (e.g., partnerId for
// Shopee app config).
func TestBackfill_MissingRequiredKeys(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "backfill_missing_keys_test"

	// Seed shopee with connection data but NO partnerId/partnerKey/appKey
	seedBackfillConnData(t, db, models.PlatformShopee, 99887, "Shopee Store", "shopee-access-token", "shopee-refresh-token", "id")

	// BackfillConnections should succeed — connection data is present
	err := BackfillConnections(ctx, db, tenantID)
	require.NoError(t, err)

	// BackfillAppConfigs should FAIL — missing partnerId and credentials
	err = BackfillAppConfigs(ctx, db, tenantID)
	// RED phase: stub returns nil, so assert.Error FAILS
	assert.Error(t, err, "BackfillAppConfigs should error when partnerId is missing — RED: stub returned nil")
}

// TestClassifyRows_CredentialVsNonCredential verifies ClassifyRows correctly
// separates credential rows (accessToken, refreshToken, partnerId) from
// non-credential rows (shopName, shopId, region).
func TestClassifyRows_CredentialVsNonCredential(t *testing.T) {
	rows := []inventoryRowData{
		{"config_key": "accessToken", "config_value": "tok-123"},
		{"config_key": "shopName", "config_value": "My Shop"},
		{"config_key": "refreshToken", "config_value": "ref-456"},
		{"config_key": "shopId", "config_value": "99887"},
		{"config_key": "partnerId", "config_value": "1001"},
		{"config_key": "region", "config_value": "id"},
	}

	credRows, nonCredRows := ClassifyRows(rows)

	// RED phase: stub returns nil,nil — both Len assertions fail
	assert.Len(t, credRows, 3, "Expected 3 credential rows — RED: stub returned nil")
	assert.Len(t, nonCredRows, 3, "Expected 3 non-credential rows — RED: stub returned nil")

	// Verify classification correctness (only reached if stub returns data)
	if len(credRows) > 0 {
		for _, row := range credRows {
			key := asString(row["config_key"])
			assert.Contains(t, []string{"accessToken", "refreshToken", "partnerId"}, key,
				"Credential row should have a credential key")
		}
	}
	if len(nonCredRows) > 0 {
		for _, row := range nonCredRows {
			key := asString(row["config_key"])
			assert.Contains(t, []string{"shopName", "shopId", "region"}, key,
				"Non-credential row should have a non-credential key")
		}
	}
}
