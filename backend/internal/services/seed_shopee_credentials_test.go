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

func TestSeedShopeeAppCredentials_ReadsFromDB(t *testing.T) {
	db := seedTestDB(t)
	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	// Pre-seed DB with Shopee credentials (as-if from migration or admin UI)
	err := repo.UpsertAppConfig(ctx, &models.CredentialAppConfig{
		TenantID:   "test_tenant",
		Platform:   models.PlatformShopee,
		PartnerID:  2011782,
		PartnerKey: "db-partner-key",
		Configured: true,
		CreatedBy:  "migration",
		UpdatedBy:  "migration",
	})
	require.NoError(t, err)

	// Function should succeed — credentials exist in DB
	err = SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err)

	// Verify DB values are unchanged (function only reads, does not write)
	cfg, err := repo.GetAppConfig(ctx, "test_tenant", models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, int64(2011782), cfg.PartnerID)
	assert.Equal(t, "db-partner-key", cfg.PartnerKey)
}

func TestSeedShopeeAppCredentials_ReadsFromDB_NotEnv(t *testing.T) {
	db := seedTestDB(t)
	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	// Pre-seed DB with live credentials (as-if from migration)
	err := repo.UpsertAppConfig(ctx, &models.CredentialAppConfig{
		TenantID:   "test_tenant",
		Platform:   models.PlatformShopee,
		PartnerID:  2011782,
		PartnerKey: "db-partner-key",
		Configured: true,
		CreatedBy:  "migration",
		UpdatedBy:  "migration",
	})
	require.NoError(t, err)

	// Call the function — should read from DB only
	err = SeedShopeeAppCredentials(db, "test_tenant")
	require.NoError(t, err)

	// Verify DB values are intact — function should not overwrite
	cfg, err := repo.GetAppConfig(ctx, "test_tenant", models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, int64(2011782), cfg.PartnerID, "should read partner_id from DB, not env")
	assert.Equal(t, "db-partner-key", cfg.PartnerKey, "should read partner_key from DB, not env")
}

func TestSeedShopeeAppCredentials_MissingDB_ReturnsError(t *testing.T) {
	db := seedTestDB(t)
	// Intentionally leave DB empty — no pre-seeded record

	// Call the function — no DB record → should error
	err := SeedShopeeAppCredentials(db, "test_tenant")
	require.Error(t, err, "should error when no DB record exists")
	assert.Contains(t, err.Error(), "no Shopee app config found")
}

func TestSeedShopeeAppCredentials_DifferentTenant_ReturnsError(t *testing.T) {
	db := seedTestDB(t)
	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	// Seed credentials for tenant A
	err := repo.UpsertAppConfig(ctx, &models.CredentialAppConfig{
		TenantID:   "tenant_a",
		Platform:   models.PlatformShopee,
		PartnerID:  1111111,
		PartnerKey: "key-a",
		Configured: true,
		CreatedBy:  "migration",
		UpdatedBy:  "migration",
	})
	require.NoError(t, err)

	// Query for tenant B — should error (no config)
	err = SeedShopeeAppCredentials(db, "tenant_b")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no Shopee app config found")
}
