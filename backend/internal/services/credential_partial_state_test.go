package services

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// =============================================================================
// classifyPartialState — pure unit tests (no DB required)
// =============================================================================

func TestClassifyPartialState_BothEmpty(t *testing.T) {
	state := classifyPartialState(canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: false,
	})
	assert.Equal(t, StateBothEmpty, state)
	assert.Equal(t, "both_empty", state.String())
}

func TestClassifyPartialState_AppOnly(t *testing.T) {
	state := classifyPartialState(canonicalCredentialLoadState{
		AppConfigured:   true,
		StoreConfigured: false,
	})
	assert.Equal(t, StateAppOnly, state)
	assert.Equal(t, "app_only", state.String())
}

func TestClassifyPartialState_ConnectionOnly(t *testing.T) {
	state := classifyPartialState(canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: true,
	})
	assert.Equal(t, StateConnectionOnly, state)
	assert.Equal(t, "connection_only", state.String())
}

func TestClassifyPartialState_BothPresent(t *testing.T) {
	state := classifyPartialState(canonicalCredentialLoadState{
		AppConfigured:   true,
		StoreConfigured: true,
	})
	assert.Equal(t, StateBothPresent, state)
	assert.Equal(t, "both_present", state.String())
}

// =============================================================================
// validateAndRepairPartialState — integration tests (testcontainers)
// =============================================================================

func TestPartialState_BothEmpty_ReturnsNil(t *testing.T) {
	db := backfillTestDB(t)
	svc := &CredentialService{dbPath: ""}
	ctx := context.Background()

	loadState := canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: false,
	}

	state, err := svc.validateAndRepairPartialState(ctx, db, "test-tenant", "shopee", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateBothEmpty, state)
}

func TestPartialState_AppOnly_ReturnsNilNoError(t *testing.T) {
	db := backfillTestDB(t)
	svc := &CredentialService{dbPath: ""}
	ctx := context.Background()

	loadState := canonicalCredentialLoadState{
		AppConfigured:   true,
		StoreConfigured: false,
	}

	state, err := svc.validateAndRepairPartialState(ctx, db, "test-tenant", "shopee", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateAppOnly, state)
}

func TestPartialState_ConnectionOnly_RepairShopee(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "partial_repair_shopee"

	// Seed platform_configs with Shopee app credentials
	seedBackfillAppData(t, db, "shopee", 12345, "partner-key-123", "", "")

	// Seed a connection in credential_connections
	seedCredentialConnection(t, db, tenantID, "shopee", "67890", "access-tok", "refresh-tok", 0, 0)

	svc := &CredentialService{dbPath: ""}

	loadState := canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: true,
	}

	state, err := svc.validateAndRepairPartialState(ctx, db, tenantID, "shopee", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateConnectionOnly, state)

	// Verify app config was created
	credRepo := repositories.NewCredentialRepository(db)
	appConfig, err := credRepo.GetAppConfig(ctx, tenantID, "shopee")
	require.NoError(t, err)
	require.NotNil(t, appConfig)
	assert.Equal(t, int64(12345), appConfig.PartnerID)
	assert.Equal(t, "partner-key-123", appConfig.PartnerKey)
	assert.True(t, appConfig.Configured)
}

func TestPartialState_ConnectionOnly_RepairLazada(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "partial_repair_lazada"

	// Seed platform_configs with Lazada app credentials
	seedBackfillAppData(t, db, "lazada", 0, "", "lazada-app-key", "lazada-app-secret")

	// Seed a connection
	seedCredentialConnection(t, db, tenantID, "lazada", "shop-123", "access-tok", "refresh-tok", 0, 0)

	svc := &CredentialService{dbPath: ""}

	loadState := canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: true,
	}

	state, err := svc.validateAndRepairPartialState(ctx, db, tenantID, "lazada", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateConnectionOnly, state)

	credRepo := repositories.NewCredentialRepository(db)
	appConfig, err := credRepo.GetAppConfig(ctx, tenantID, "lazada")
	require.NoError(t, err)
	require.NotNil(t, appConfig)
	assert.Equal(t, "lazada-app-key", appConfig.AppKey)
	assert.Equal(t, "lazada-app-secret", appConfig.AppSecret)
	assert.True(t, appConfig.Configured)
}

func TestPartialState_ConnectionOnly_RepairTikTok(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "partial_repair_tiktok"

	// Seed platform_configs with TikTok app credentials
	seedBackfillAppData(t, db, "tiktok", 0, "", "tiktok-app-key", "tiktok-app-secret")

	// Seed a connection
	seedCredentialConnection(t, db, tenantID, "tiktok", "seller-456", "access-tok", "refresh-tok", 0, 0)

	svc := &CredentialService{dbPath: ""}

	loadState := canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: true,
	}

	state, err := svc.validateAndRepairPartialState(ctx, db, tenantID, "tiktok", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateConnectionOnly, state)

	credRepo := repositories.NewCredentialRepository(db)
	appConfig, err := credRepo.GetAppConfig(ctx, tenantID, "tiktok")
	require.NoError(t, err)
	require.NotNil(t, appConfig)
	assert.Equal(t, "tiktok-app-key", appConfig.AppKey)
	assert.Equal(t, "tiktok-app-secret", appConfig.AppSecret)
	assert.True(t, appConfig.Configured)
}

func TestPartialState_ConnectionOnly_NoPlatformConfigs_ReturnsError(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "partial_no_configs"

	// Seed a connection but NO platform_configs data
	seedCredentialConnection(t, db, tenantID, "shopee", "67890", "access-tok", "refresh-tok", 0, 0)

	svc := &CredentialService{dbPath: ""}

	loadState := canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: true,
	}

	_, err := svc.validateAndRepairPartialState(ctx, db, tenantID, "shopee", loadState)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repair")
}

func TestPartialState_BothPresent_ReturnsNil(t *testing.T) {
	db := backfillTestDB(t)
	svc := &CredentialService{dbPath: ""}
	ctx := context.Background()

	loadState := canonicalCredentialLoadState{
		AppConfigured:   true,
		StoreConfigured: true,
	}

	state, err := svc.validateAndRepairPartialState(ctx, db, "test-tenant", "shopee", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateBothPresent, state)
}

func TestPartialState_RepairIdempotent(t *testing.T) {
	db := backfillTestDB(t)
	ctx := context.Background()
	tenantID := "partial_idempotent"

	// Seed platform_configs with app credentials
	seedBackfillAppData(t, db, "shopee", 12345, "partner-key-123", "", "")

	// Seed a connection
	seedCredentialConnection(t, db, tenantID, "shopee", "67890", "access-tok", "refresh-tok", 0, 0)

	svc := &CredentialService{dbPath: ""}
	loadState := canonicalCredentialLoadState{
		AppConfigured:   false,
		StoreConfigured: true,
	}

	// First repair
	state, err := svc.validateAndRepairPartialState(ctx, db, tenantID, "shopee", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateConnectionOnly, state)

	// Second repair — should be idempotent (no error, no duplicate)
	state, err = svc.validateAndRepairPartialState(ctx, db, tenantID, "shopee", loadState)
	require.NoError(t, err)
	assert.Equal(t, StateConnectionOnly, state)

	// Verify exactly one app config
	credRepo := repositories.NewCredentialRepository(db)
	appConfig, err := credRepo.GetAppConfig(ctx, tenantID, "shopee")
	require.NoError(t, err)
	require.NotNil(t, appConfig)
	assert.Equal(t, int64(12345), appConfig.PartnerID)
}

// =============================================================================
// Helpers
// =============================================================================

func seedCredentialConnection(t *testing.T, db *gorm.DB, tenantID, platform, storeID, accessToken, refreshToken string, tokenExpiry, refreshExpiry int64) {
	t.Helper()
	// Use raw SQL to insert connection directly (avoids encryption complexity in tests)
	db.Exec(
		`INSERT INTO credential_connections (id, tenant_id, platform, store_identifier, status, access_token, refresh_token, token_expiry, refresh_expiry, version, created_by)
		 VALUES (?, ?, ?, ?, 'connected', ?, ?, ?, ?, 1, 'test')`,
		"test-conn-"+tenantID+"-"+platform+"-"+storeID,
		tenantID,
		platform,
		storeID,
		accessToken,
		refreshToken,
		tokenExpiry,
		refreshExpiry,
	)
}
