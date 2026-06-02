package services

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

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

// TestBackfillParity verifies row counts and key field integrity after backfill.
// Compares legacy platform_configs against canonical tables:
//   - 3 app_configs + 3 connections = 6 canonical rows
//   - All canonical rows have valid non-null platform + identifier fields
//   - Encrypted values preserved (decrypted canonical == original plaintext)
func TestBackfillParity(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "parity_test"

	// Seed app-level credentials for all 3 platforms
	seedBackfillAppData(t, db, models.PlatformShopee, 2001, "shopee-pk-encrypted", "", "")
	seedBackfillAppData(t, db, models.PlatformLazada, 0, "", "lazada-appkey-enc", "lazada-appsecret-enc")
	seedBackfillAppData(t, db, models.PlatformTiktok, 0, "", "tiktok-appkey-enc", "tiktok-appsecret-enc")

	// Seed connection-level data for all 3 platforms (including TikTok)
	seedBackfillConnData(t, db, models.PlatformShopee, 11223, "Shopee Parity Store", "shopee-at-enc", "shopee-rt-enc", "id")
	seedBackfillConnData(t, db, models.PlatformLazada, 44556, "Lazada Parity Store", "lazada-at-enc", "lazada-rt-enc", "sg")
	seedBackfillConnData(t, db, models.PlatformTiktok, 77889, "TikTok Parity Store", "tiktok-at-enc", "tiktok-rt-enc", "")

	// Run backfill
	err := BackfillAppConfigs(ctx, db, tenantID)
	require.NoError(t, err)
	err = BackfillConnections(ctx, db, tenantID)
	require.NoError(t, err)

	credRepo := repositories.NewCredentialRepository(db)

	// --- Row count parity ---
	var appCfgs []models.CredentialAppConfig
	require.NoError(t, db.WithContext(ctx).Find(&appCfgs).Error)
	assert.Len(t, appCfgs, 3, "Expected 3 canonical app_config rows")

	conns, err := credRepo.ListConnections(ctx, tenantID, "")
	require.NoError(t, err)
	assert.Len(t, conns, 3, "Expected 3 canonical connection rows")

	totalCanonical := len(appCfgs) + len(conns)
	assert.Equal(t, 6, totalCanonical, "3 app_configs + 3 connections = 6 canonical rows")

	// --- Key field validation ---
	for _, cfg := range appCfgs {
		assert.NotEmpty(t, cfg.Platform, "App config platform must not be empty")
		assert.NotEmpty(t, cfg.TenantID, "App config tenant_id must not be empty")
	}
	for _, conn := range conns {
		assert.NotEmpty(t, conn.Platform, "Connection platform must not be empty")
		assert.NotEmpty(t, conn.TenantID, "Connection tenant_id must not be empty")
		assert.NotEmpty(t, conn.StoreIdentifier, "Connection store_identifier must not be empty")
	}

	// --- Encrypted value preservation ---
	// Verify app config secrets preserved via decryption round-trip
	shopeeApp, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, shopeeApp)
	assert.Equal(t, int64(2001), shopeeApp.PartnerID, "Shopee partner_id preserved")
	assert.Equal(t, "shopee-pk-encrypted", shopeeApp.PartnerKey, "Shopee partner_key decrypted matches original")

	lazadaApp, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformLazada)
	require.NoError(t, err)
	require.NotNil(t, lazadaApp)
	assert.Equal(t, "lazada-appkey-enc", lazadaApp.AppKey, "Lazada app_key decrypted matches original")
	assert.Equal(t, "lazada-appsecret-enc", lazadaApp.AppSecret, "Lazada app_secret decrypted matches original")

	tiktokApp, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformTiktok)
	require.NoError(t, err)
	require.NotNil(t, tiktokApp)
	assert.Equal(t, "tiktok-appkey-enc", tiktokApp.AppKey, "TikTok app_key decrypted matches original")
	assert.Equal(t, "tiktok-appsecret-enc", tiktokApp.AppSecret, "TikTok app_secret decrypted matches original")

	// Verify connection secrets preserved via decryption round-trip
	shopeeConn, err := credRepo.GetConnection(ctx, tenantID, models.PlatformShopee, "11223")
	require.NoError(t, err)
	require.NotNil(t, shopeeConn)
	assert.Equal(t, "shopee-at-enc", shopeeConn.AccessToken, "Shopee access_token decrypted matches original")
	assert.Equal(t, "shopee-rt-enc", shopeeConn.RefreshToken, "Shopee refresh_token decrypted matches original")
	assert.Equal(t, "Shopee Parity Store", shopeeConn.StoreName, "Shopee store_name preserved")
	assert.Equal(t, "id", shopeeConn.Region, "Shopee region preserved")

	lazadaConn, err := credRepo.GetConnection(ctx, tenantID, models.PlatformLazada, "44556")
	require.NoError(t, err)
	require.NotNil(t, lazadaConn)
	assert.Equal(t, "lazada-at-enc", lazadaConn.AccessToken, "Lazada access_token decrypted matches original")
	assert.Equal(t, "lazada-rt-enc", lazadaConn.RefreshToken, "Lazada refresh_token decrypted matches original")
	assert.Equal(t, "sg", lazadaConn.Region, "Lazada region preserved")

	tiktokConn, err := credRepo.GetConnection(ctx, tenantID, models.PlatformTiktok, "77889")
	require.NoError(t, err)
	require.NotNil(t, tiktokConn)
	assert.Equal(t, "tiktok-at-enc", tiktokConn.AccessToken, "TikTok access_token decrypted matches original")
	assert.Equal(t, "tiktok-rt-enc", tiktokConn.RefreshToken, "TikTok refresh_token decrypted matches original")
	assert.Equal(t, "TikTok Parity Store", tiktokConn.StoreName, "TikTok store_name preserved")
}

// TestBackfillCorruptedRequired verifies that BackfillAppConfigs fails fast when
// a required credential field contains a corrupted Fernet token that cannot be
// decrypted. The tenant/platform must be skipped — no garbage inserted.
func TestBackfillCorruptedRequired(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "corrupted_test"

	repo := repositories.NewTenantPlatformConfigRepository(db)

	// Seed Shopee partnerId as plaintext (non-credential, always passes)
	require.NoError(t, repo.SetConfig(ctx, models.PlatformShopee, "partnerId", "1001", false))

	// Insert a corrupted Fernet token for partnerKey directly via GORM.
	// It has the "gAAAAA" prefix so IsEncrypted() returns true, but the
	// payload is garbage so Decrypt() fails.
	corruptedToken := "gAAAAA" + "corrupted_fernet_payload_that_will_never_decrypt"
	require.NoError(t, db.WithContext(ctx).Create(&repositories.TenantPlatformConfig{
		ID:          fmt.Sprintf("%s_%s_%d", models.PlatformShopee, "partnerKey", time.Now().UnixNano()),
		Platform:    models.PlatformShopee,
		ConfigKey:   "partnerKey",
		ConfigValue: corruptedToken,
		IsEncrypted: true,
	}).Error)

	// Seed valid Lazada data (should NOT be processed since Shopee fails first)
	seedBackfillAppData(t, db, models.PlatformLazada, 0, "", "lazada-app-key", "lazada-app-secret")

	// Call backfill — should FAIL FAST due to corrupted Shopee partnerKey
	err := BackfillAppConfigs(ctx, db, tenantID)
	assert.Error(t, err, "BackfillAppConfigs should fail on corrupted required credential")
	assert.Contains(t, err.Error(), "corrupted", "Error should mention corruption")
	assert.Contains(t, err.Error(), models.PlatformShopee, "Error should mention the platform")

	// Verify NO app configs were created (fail-fast means no partial inserts)
	credRepo := repositories.NewCredentialRepository(db)
	shopeeCfg, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
	require.NoError(t, err)
	assert.Nil(t, shopeeCfg, "Shopee app config should NOT exist — corrupted data must not be inserted")

	lazadaCfg, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformLazada)
	require.NoError(t, err)
	assert.Nil(t, lazadaCfg, "Lazada app config should NOT exist — backfill failed before processing it")
}

// TestBackfillCorruptedConnectionRequired verifies that BackfillConnections
// skips a connection when accessToken or refreshToken contains a corrupted
// Fernet token.
func TestBackfillCorruptedConnectionRequired(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "corrupted_conn_test"

	repo := repositories.NewTenantPlatformConfigRepository(db)

	// Seed non-credential fields as plaintext
	require.NoError(t, repo.SetConfig(ctx, models.PlatformShopee, "shopId", "99887", false))
	require.NoError(t, repo.SetConfig(ctx, models.PlatformShopee, "shopName", "Test Shop", false))
	require.NoError(t, repo.SetConfig(ctx, models.PlatformShopee, "region", "id", false))

	// Insert corrupted Fernet token for accessToken
	corruptedToken := "gAAAAA" + "corrupted_fernet_payload"
	require.NoError(t, db.WithContext(ctx).Create(&repositories.TenantPlatformConfig{
		ID:          fmt.Sprintf("%s_%s_%d", models.PlatformShopee, "accessToken", time.Now().UnixNano()),
		Platform:    models.PlatformShopee,
		ConfigKey:   "accessToken",
		ConfigValue: corruptedToken,
		IsEncrypted: true,
	}).Error)

	// Insert valid refreshToken (doesn't matter — accessToken is already corrupted)
	require.NoError(t, repo.SetConfig(ctx, models.PlatformShopee, "refreshToken", "valid-refresh", true))

	// Call backfill — should skip Shopee connection, return nil (other platforms OK)
	err := BackfillConnections(ctx, db, tenantID)
	require.NoError(t, err, "BackfillConnections should not return error — just skips corrupted connections")

	// Verify no Shopee connection was created
	credRepo := repositories.NewCredentialRepository(db)
	shopeeConn, err := credRepo.GetConnection(ctx, tenantID, models.PlatformShopee, "99887")
	require.NoError(t, err)
	assert.Nil(t, shopeeConn, "Shopee connection should NOT exist — corrupted token must prevent insertion")
}

// TestBackfillDuplicateKeyResolution verifies that when platform_configs has
// duplicate rows for the same platform+config_key, the backfill resolves them:
//   - Duplicate keys with conflicting non-empty values → latest updated_at wins
//   - Duplicate keys with one empty and one non-empty → use non-empty
func TestBackfillDuplicateKeyResolution(t *testing.T) {
	db := backfillTestDB(t)
	repo := repositories.NewTenantPlatformConfigRepository(db)
	ctx := context.Background()
	tenantID := "dup_key_test"

	// --- Test 1: Duplicate keys with conflicting values → latest wins ---
	t1 := time.Now().Add(-1 * time.Hour)
	t2 := time.Now()

	// Insert older row: partnerKey = "old-key"
	require.NoError(t, db.WithContext(ctx).Create(&repositories.TenantPlatformConfig{
		ID:          fmt.Sprintf("%s_%s_%d", models.PlatformShopee, "partnerKey", t1.UnixNano()),
		Platform:    models.PlatformShopee,
		ConfigKey:   "partnerKey",
		ConfigValue: "old-key",
		IsEncrypted: false,
		UpdatedAt:   t1,
	}).Error)

	// Insert newer row: partnerKey = "new-key"
	require.NoError(t, db.WithContext(ctx).Create(&repositories.TenantPlatformConfig{
		ID:          fmt.Sprintf("%s_%s_%d", models.PlatformShopee, "partnerKey", t2.UnixNano()),
		Platform:    models.PlatformShopee,
		ConfigKey:   "partnerKey",
		ConfigValue: "new-key",
		IsEncrypted: false,
		UpdatedAt:   t2,
	}).Error)

	// Seed required partnerId via repo (single row, no conflict)
	require.NoError(t, repo.SetConfig(ctx, models.PlatformShopee, "partnerId", "1001", false))

	// Backfill should succeed — duplicate partnerKey resolved to latest value
	err := BackfillAppConfigs(ctx, db, tenantID)
	require.NoError(t, err)

	// Verify canonical app config uses "new-key" (latest updated_at)
	credRepo := repositories.NewCredentialRepository(db)
	shopeeCfg, err := credRepo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, shopeeCfg, "Shopee app config should exist")
	assert.Equal(t, "new-key", shopeeCfg.PartnerKey, "Duplicate key should resolve to latest updated_at value")
	assert.Equal(t, int64(1001), shopeeCfg.PartnerID)

	// --- Test 2: Duplicate keys with one empty and one non-empty → use non-empty ---
	tenantID2 := "dup_key_empty_test"
	t3 := time.Now().Add(-2 * time.Hour)
	t4 := time.Now().Add(-1 * time.Hour)

	// Insert older row: appKey = ""
	require.NoError(t, db.WithContext(ctx).Create(&repositories.TenantPlatformConfig{
		ID:          fmt.Sprintf("%s_%s_%d", models.PlatformLazada, "appKey", t3.UnixNano()),
		Platform:    models.PlatformLazada,
		ConfigKey:   "appKey",
		ConfigValue: "",
		IsEncrypted: false,
		UpdatedAt:   t3,
	}).Error)

	// Insert newer row: appKey = "lazada-app"
	require.NoError(t, db.WithContext(ctx).Create(&repositories.TenantPlatformConfig{
		ID:          fmt.Sprintf("%s_%s_%d", models.PlatformLazada, "appKey", t4.UnixNano()),
		Platform:    models.PlatformLazada,
		ConfigKey:   "appKey",
		ConfigValue: "lazada-app",
		IsEncrypted: false,
		UpdatedAt:   t4,
	}).Error)

	// Seed appSecret via repo (single row, no conflict)
	require.NoError(t, repo.SetConfig(ctx, models.PlatformLazada, "appSecret", "lazada-secret", false))

	err = BackfillAppConfigs(ctx, db, tenantID2)
	require.NoError(t, err)

	lazadaCfg, err := credRepo.GetAppConfig(ctx, tenantID2, models.PlatformLazada)
	require.NoError(t, err)
	require.NotNil(t, lazadaCfg, "Lazada app config should exist")
	assert.Equal(t, "lazada-app", lazadaCfg.AppKey, "Empty+non-empty duplicate should use non-empty value")
	assert.Equal(t, "lazada-secret", lazadaCfg.AppSecret)
}

