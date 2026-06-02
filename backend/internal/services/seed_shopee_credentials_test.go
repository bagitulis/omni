package services

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// seedTestDB creates a test DB with all credential models auto-migrated.
func seedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	t.Setenv("ENCRYPTION_KEY", key)
	return testutils.SetupTestPostgresWithModels(t,
		&models.CredentialConnection{},
		&models.CredentialAppConfig{},
		&models.OAuthConnectionAttempt{},
		&models.CredentialAuditEvent{},
	)
}

// clearShopeeEnv unsets all SHOPEE_* env vars to ensure test isolation.
func clearShopeeEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SHOPEE_ENV", "")
	t.Setenv("SHOPEE_PARTNER_ID", "")
	t.Setenv("SHOPEE_PARTNER_KEY", "")
	t.Setenv("SHOPEE_LIVE_PARTNER_ID", "")
	t.Setenv("SHOPEE_LIVE_PARTNER_KEY", "")
}

func TestSeedShopeeAppCredentials_LiveEnv(t *testing.T) {
	clearShopeeEnv(t)
	t.Setenv("SHOPEE_ENV", "live")
	t.Setenv("SHOPEE_LIVE_PARTNER_ID", "2011782")
	t.Setenv("SHOPEE_LIVE_PARTNER_KEY", "livekey")

	db := seedTestDB(t)
	err := SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err)

	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAppConfig(ctx, "test_tenant", models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, int64(2011782), cfg.PartnerID)
	assert.Equal(t, "livekey", cfg.PartnerKey)
	assert.True(t, cfg.Configured)
	assert.Equal(t, "system-seed", cfg.CreatedBy)
}

func TestSeedShopeeAppCredentials_TestEnv(t *testing.T) {
	clearShopeeEnv(t)

	t.Setenv("SHOPEE_ENV", "test")
	t.Setenv("SHOPEE_PARTNER_ID", "1187586")
	t.Setenv("SHOPEE_PARTNER_KEY", "testkey")

	db := seedTestDB(t)

	err := SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err)

	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAppConfig(ctx, "test_tenant", models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, int64(1187586), cfg.PartnerID)
	assert.Equal(t, "testkey", cfg.PartnerKey)
}

func TestSeedShopeeAppCredentials_DefaultsToTestEnv(t *testing.T) {
	clearShopeeEnv(t)

	// SHOPEE_ENV not set → defaults to "test"
	t.Setenv("SHOPEE_PARTNER_ID", "9999")
	t.Setenv("SHOPEE_PARTNER_KEY", "defaultkey")

	db := seedTestDB(t)

	err := SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err)

	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAppConfig(ctx, "test_tenant", models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, int64(9999), cfg.PartnerID)
}

func TestSeedShopeeAppCredentials_IdempotentReseed(t *testing.T) {
	clearShopeeEnv(t)

	t.Setenv("SHOPEE_ENV", "test")
	t.Setenv("SHOPEE_PARTNER_ID", "1187586")
	t.Setenv("SHOPEE_PARTNER_KEY", "testkey")

	db := seedTestDB(t)

	// Seed twice with same values
	err := SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err)

	err = SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err)

	// Verify only 1 row exists
	var count int64
	db.Model(&models.CredentialAppConfig{}).
		Where("tenant_id = ? AND platform = ?", "test_tenant", models.PlatformShopee).
		Count(&count)
	assert.Equal(t, int64(1), count, "idempotent re-seed should produce exactly 1 row")
}

func TestSeedShopeeAppCredentials_MissingEnvVarsReturnsNil(t *testing.T) {
	clearShopeeEnv(t)

	// All SHOPEE_* vars unset (cleared above)
	db := seedTestDB(t)

	err := SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err, "missing env vars should return nil, not error")

	// Verify no row was created
	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAppConfig(ctx, "test_tenant", models.PlatformShopee)
	require.NoError(t, err)
	assert.Nil(t, cfg, "no row should be created when env vars are missing")
}

func TestSeedShopeeAppCredentials_InvalidPartnerIDReturnsNil(t *testing.T) {
	clearShopeeEnv(t)

	t.Setenv("SHOPEE_ENV", "test")
	t.Setenv("SHOPEE_PARTNER_ID", "not_a_number")
	t.Setenv("SHOPEE_PARTNER_KEY", "testkey")

	db := seedTestDB(t)

	err := SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err, "invalid partner_id should return nil, not error")

	// Verify no row was created
	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAppConfig(ctx, "test_tenant", models.PlatformShopee)
	require.NoError(t, err)
	assert.Nil(t, cfg, "no row should be created when partner_id is invalid")
}
