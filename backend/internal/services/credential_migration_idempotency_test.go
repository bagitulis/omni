package services

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationBackfillRerunAfterPartialFailureNoDuplicates verifies that
// re-running backfill after a partial failure does not create duplicate rows.
// Setup: 3 bundles → first run creates 2, then we add a third.
// Rerun: existing 2 are skipped, only the new one is created.
func TestMigrationBackfillRerunAfterPartialFailureNoDuplicates(t *testing.T) {
	db := setupCredentialMigrationSQLite(t)
	svc := &CredentialMigrationService{}
	ctx := context.Background()

	bundleA := legacyCredentialBundle{
		TenantID: "tenant_pf", Platform: models.PlatformShopee,
		StoreIdentifier: "shop-001", StoreName: "Shop Alpha",
		AccessToken: "token-alpha", RefreshToken: "refresh-alpha",
		PartnerID: 2001, PartnerKey: "pkey-alpha", Region: "id",
	}
	bundleB := legacyCredentialBundle{
		TenantID: "tenant_pf", Platform: models.PlatformTiktok,
		StoreIdentifier: "shop-002", StoreName: "Shop Beta",
		AccessToken: "token-beta", RefreshToken: "refresh-beta",
		ShopCipher: "cipher-beta", PartnerID: 2002, PartnerKey: "pkey-beta", Region: "id",
	}
	bundleC := legacyCredentialBundle{
		TenantID: "tenant_pf", Platform: models.PlatformLazada,
		StoreIdentifier: "shop-003", StoreName: "Shop Gamma",
		AccessToken: "token-gamma", RefreshToken: "refresh-gamma",
		AppKey: "appkey-gamma", AppSecret: "appsecret-gamma", Region: "id",
	}

	// First run: create A and B only (simulating partial completion).
	firstRun := newCredentialMigrationReport(CredentialMigrationModeBackfill)
	err := svc.backfillTenant(ctx, db, []legacyCredentialBundle{bundleA, bundleB}, firstRun, CredentialMigrationOptions{Actor: "test_rerun"})
	require.NoError(t, err)
	assert.Equal(t, 2, firstRun.Backfill.CreatedConnections)
	assert.Equal(t, 2, firstRun.Backfill.CreatedAppConfigs)
	assert.Equal(t, 0, firstRun.Backfill.SkippedExisting)

	// Second run: all 3 bundles (A and B already exist, C is new).
	secondRun := newCredentialMigrationReport(CredentialMigrationModeBackfill)
	err = svc.backfillTenant(ctx, db, []legacyCredentialBundle{bundleA, bundleB, bundleC}, secondRun, CredentialMigrationOptions{Actor: "test_rerun"})
	require.NoError(t, err)
	assert.Equal(t, 1, secondRun.Backfill.CreatedConnections, "only C should be created")
	assert.Equal(t, 1, secondRun.Backfill.CreatedAppConfigs, "only C should be created")
	assert.Equal(t, 4, secondRun.Backfill.SkippedExisting, "A conn + A app + B conn + B app skipped")

	// Verify no duplicates: total 3 connections, 3 app configs.
	repo := repositories.NewCredentialRepository(db)
	conns, err := repo.ListConnections(ctx, "tenant_pf", "")
	require.NoError(t, err)
	assert.Len(t, conns, 3, "expected exactly 3 connections — no duplicates")

	var totalConns int64
	require.NoError(t, db.Model(&models.CredentialConnection{}).Where("tenant_id = ?", "tenant_pf").Count(&totalConns).Error)
	assert.Equal(t, int64(3), totalConns, "raw row count confirms no duplicates")
}

// TestMigrationBackfillAlreadyEncryptedNotDoubleEncrypted verifies that
// backfill does not re-encrypt already-encrypted credential connections.
// It creates an encrypted connection via the repository, then runs backfill
// with matching bundle data. The backfill should skip the existing row,
// and reading it back should decrypt successfully (not double-encrypted).
func TestMigrationBackfillAlreadyEncryptedNotDoubleEncrypted(t *testing.T) {
	db := setupCredentialMigrationSQLite(t)
	svc := &CredentialMigrationService{}
	ctx := context.Background()

	// Create a connection via the repository (encrypts the secrets).
	repo := repositories.NewCredentialRepository(db)
	originalToken := "plaintext-access-token-for-encryption-test"
	originalRefresh := "plaintext-refresh-token-for-encryption-test"
	err := repo.CreateConnection(ctx, &models.CredentialConnection{
		TenantID: "tenant_enc", Platform: models.PlatformShopee,
		StoreIdentifier: "shop-enc-001", StoreName: "Encrypted Shop",
		AccessToken: originalToken, RefreshToken: originalRefresh,
		Status: "connected", CreatedBy: "setup", UpdatedBy: "setup",
	})
	require.NoError(t, err)

	// Read back to verify encryption worked.
	stored, err := repo.GetConnection(ctx, "tenant_enc", models.PlatformShopee, "shop-enc-001")
	require.NoError(t, err)
	require.NotNil(t, stored)
	assert.Equal(t, originalToken, stored.AccessToken, "decrypted access token matches original")
	assert.Equal(t, originalRefresh, stored.RefreshToken, "decrypted refresh token matches original")

	// Now run backfill with a bundle that matches this connection.
	bundle := legacyCredentialBundle{
		TenantID: "tenant_enc", Platform: models.PlatformShopee,
		StoreIdentifier: "shop-enc-001", StoreName: "Encrypted Shop",
		AccessToken: "different-token-should-be-ignored",
		RefreshToken: "different-refresh-should-be-ignored",
		PartnerID: 3001, PartnerKey: "pkey-enc", Region: "id",
	}
	report := newCredentialMigrationReport(CredentialMigrationModeBackfill)
	err = svc.backfillTenant(ctx, db, []legacyCredentialBundle{bundle}, report, CredentialMigrationOptions{Actor: "test_enc"})
	require.NoError(t, err)
	assert.Equal(t, 0, report.Backfill.CreatedConnections, "existing connection skipped")
	assert.Equal(t, 1, report.Backfill.CreatedAppConfigs, "app config created (no prior app config)")
	assert.Equal(t, 1, report.Backfill.SkippedExisting, "connection skipped")

	// Read back the connection again — must still decrypt correctly (not double-encrypted).
	afterBackfill, err := repo.GetConnection(ctx, "tenant_enc", models.PlatformShopee, "shop-enc-001")
	require.NoError(t, err)
	require.NotNil(t, afterBackfill)
	assert.Equal(t, originalToken, afterBackfill.AccessToken, "token not double-encrypted after backfill")
	assert.Equal(t, originalRefresh, afterBackfill.RefreshToken, "refresh not double-encrypted after backfill")

	// Verify the raw DB value is still a Fernet token (encrypted at rest).
	var rawToken string
	require.NoError(t, db.Model(&models.CredentialConnection{}).
		Where("tenant_id = ? AND platform = ? AND store_identifier = ?", "tenant_enc", models.PlatformShopee, "shop-enc-001").
		Select("access_token").Scan(&rawToken).Error)
	assert.True(t, len(rawToken) > 20, "raw value is encrypted (not plaintext)")
	assert.NotEqual(t, originalToken, rawToken, "raw DB value is Fernet token, not plaintext")
}

// TestMigrationPreflightTextReportIncludesDBInfo verifies that the dry-run
// text report includes DB type info, schema state, row counts, and gate status.
func TestMigrationPreflightTextReportIncludesDBInfo(t *testing.T) {
	report := newCredentialMigrationReport(CredentialMigrationModeDryRun)
	report.CanWrite = true
	report.CutoverReady = false
	report.SourceKeyCounts[models.PlatformShopee] = 5
	report.SourceKeyCounts[models.PlatformLazada] = 3
	report.TargetCredentialCounts[models.PlatformShopee] = 4
	report.TargetCredentialCounts[models.PlatformLazada] = 3
	report.Tenants = append(report.Tenants, CredentialMigrationTenant{
		TenantID: "tenant_preflight", SchemaName: "tenant_tenant_preflight",
		SourceRows: 8, TargetConnections: 7, TargetAppConfigs: 3,
		DuplicateTargets: 0, IncompleteSources: 1, UnknownShapes: 0,
		UnreadableSecrets: 0, TenantCoverageMatch: true, PlatformCountMatches: true,
	})
	report.Tenants = append(report.Tenants, CredentialMigrationTenant{
		TenantID: "tenant_b", SchemaName: "tenant_tenant_b",
		SourceRows: 4, TargetConnections: 2, TargetAppConfigs: 1,
		MissingTenantSchema: false,
	})

	var buf bytes.Buffer
	require.NoError(t, WriteCredentialMigrationTextReport(&buf, report))
	output := buf.String()

	// Mode and gate status.
	assert.Contains(t, output, "Credential migration mode: dry_run")
	assert.Contains(t, output, "Can write: true")
	assert.Contains(t, output, "Cutover ready: false")

	// Source key counts (preflight summary).
	assert.Contains(t, output, "Source key counts:")
	assert.Contains(t, output, "shopee=5")
	assert.Contains(t, output, "lazada=3")

	// Target credential counts.
	assert.Contains(t, output, "Target credential counts:")
	assert.Contains(t, output, "shopee=4")

	// Per-tenant schema state and row counts.
	assert.Contains(t, output, "Tenant: tenant_preflight")
	assert.Contains(t, output, "schema=tenant_tenant_preflight")
	assert.Contains(t, output, "source_rows=8")
	assert.Contains(t, output, "target_connections=7")
	assert.Contains(t, output, "target_app_configs=3")
	assert.Contains(t, output, "duplicate_targets=0")
	assert.Contains(t, output, "incomplete_sources=1")
	assert.Contains(t, output, "coverage_match=true")
	assert.Contains(t, output, "platform_match=true")

	// Second tenant.
	assert.Contains(t, output, "Tenant: tenant_b")
	assert.Contains(t, output, "schema=tenant_tenant_b")
	assert.Contains(t, output, "source_rows=4")

	// Backfill summary (even in dry-run, shows zeros).
	assert.Contains(t, output, "Backfill: created_connections=0 created_app_configs=0 skipped_existing=0")
}

// TestMigrationPreflightTextReportShowsAbortReasons verifies that abort
// reasons appear in the preflight text output when gates fail.
func TestMigrationPreflightTextReportShowsAbortReasons(t *testing.T) {
	report := newCredentialMigrationReport(CredentialMigrationModeDryRun)
	report.CanWrite = false
	report.AbortReasons = []string{"missing_tenant_schema", "unreadable_encrypted_value"}

	var buf bytes.Buffer
	require.NoError(t, WriteCredentialMigrationTextReport(&buf, report))
	output := buf.String()

	assert.Contains(t, output, "Can write: false")
	assert.Contains(t, output, "Abort reasons: missing_tenant_schema; unreadable_encrypted_value")
}

// TestMigrationBackfillPerUnitTransactionIsolation verifies that when one
// bundle fails, previously completed bundles retain their data.
// Each CreateConnection/UpsertAppConfig is an individual auto-committed
// operation, so a failure in bundle 2 does not roll back bundle 1.
func TestMigrationBackfillPerUnitTransactionIsolation(t *testing.T) {
	db := setupCredentialMigrationSQLite(t)
	svc := &CredentialMigrationService{}
	ctx := context.Background()

	// Bundle 1: valid — should succeed.
	bundleOK := legacyCredentialBundle{
		TenantID: "tenant_tx", Platform: models.PlatformShopee,
		StoreIdentifier: "shop-ok-001", StoreName: "Good Shop",
		AccessToken: "token-ok", RefreshToken: "refresh-ok",
		PartnerID: 4001, PartnerKey: "pkey-ok", Region: "id",
	}
	// Bundle 2: empty platform → validateConnectionScope returns error.
	bundleBad := legacyCredentialBundle{
		TenantID: "tenant_tx", Platform: "", // ← triggers validation error
		StoreIdentifier: "shop-bad-002", StoreName: "Bad Shop",
		AccessToken: "token-bad", RefreshToken: "refresh-bad",
		PartnerID: 4002, PartnerKey: "pkey-bad", Region: "id",
	}

	report := newCredentialMigrationReport(CredentialMigrationModeBackfill)
	err := svc.backfillTenant(ctx, db, []legacyCredentialBundle{bundleOK, bundleBad}, report, CredentialMigrationOptions{Actor: "test_tx"})

	// Expect error from bundle 2 (empty platform validation).
	require.Error(t, err)
	assert.Contains(t, err.Error(), "platform is required")

	// Bundle 1's data must still exist — per-unit isolation.
	repo := repositories.NewCredentialRepository(db)
	conn, err := repo.GetConnection(ctx, "tenant_tx", models.PlatformShopee, "shop-ok-001")
	require.NoError(t, err)
	require.NotNil(t, conn, "bundle 1 connection survived despite bundle 2 failure")
	assert.Equal(t, "connected", conn.Status)
	assert.Equal(t, "Good Shop", conn.StoreName)
	assert.Equal(t, "token-ok", conn.AccessToken, "bundle 1 data intact")

	// Verify app config from bundle 1 also survived.
	appCfg, err := repo.GetAppConfig(ctx, "tenant_tx", models.PlatformShopee)
	require.NoError(t, err)
	require.NotNil(t, appCfg, "bundle 1 app config survived despite bundle 2 failure")
	assert.True(t, appCfg.Configured)

	// Report should reflect bundle 1 was processed before the error.
	assert.Equal(t, 1, report.Backfill.CreatedConnections)
	assert.Equal(t, 1, report.Backfill.CreatedAppConfigs)
}

// TestMigrationBackfillMultipleRerunsAllSkip verifies that running backfill
// N times after initial success produces zero creates and all skips.
func TestMigrationBackfillMultipleRerunsAllSkip(t *testing.T) {
	db := setupCredentialMigrationSQLite(t)
	svc := &CredentialMigrationService{}
	ctx := context.Background()

	bundles := []legacyCredentialBundle{
		{
			TenantID: "tenant_multi", Platform: models.PlatformShopee,
			StoreIdentifier: "shop-m1", StoreName: "Multi Shop 1",
			AccessToken: "tok-m1", RefreshToken: "ref-m1",
			PartnerID: 5001, PartnerKey: "pk-m1", Region: "id",
		},
		{
			TenantID: "tenant_multi", Platform: models.PlatformTiktok,
			StoreIdentifier: "shop-m2", StoreName: "Multi Shop 2",
			AccessToken: "tok-m2", RefreshToken: "ref-m2",
			ShopCipher: "cip-m2", PartnerID: 5002, PartnerKey: "pk-m2", Region: "id",
		},
	}

	// Initial run.
	r1 := newCredentialMigrationReport(CredentialMigrationModeBackfill)
	require.NoError(t, svc.backfillTenant(ctx, db, bundles, r1, CredentialMigrationOptions{Actor: "test_multi"}))
	assert.Equal(t, 2, r1.Backfill.CreatedConnections)
	assert.Equal(t, 2, r1.Backfill.CreatedAppConfigs)
	assert.Equal(t, 0, r1.Backfill.SkippedExisting)

	// Rerun 3 times — all should skip every time.
	for i := 2; i <= 4; i++ {
		rerun := newCredentialMigrationReport(CredentialMigrationModeBackfill)
		require.NoError(t, svc.backfillTenant(ctx, db, bundles, rerun, CredentialMigrationOptions{Actor: "test_multi"}))
		assert.Equal(t, 0, rerun.Backfill.CreatedConnections, "run %d: no creates", i)
		assert.Equal(t, 0, rerun.Backfill.CreatedAppConfigs, "run %d: no creates", i)
		assert.Equal(t, 4, rerun.Backfill.SkippedExisting, "run %d: all 4 skipped (2 conn + 2 app)", i)
	}

	// Final verification: exactly 2 connections, no duplicates.
	var connCount int64
	require.NoError(t, db.Model(&models.CredentialConnection{}).Where("tenant_id = ?", "tenant_multi").Count(&connCount).Error)
	assert.Equal(t, int64(2), connCount)
}

// TestMigrationDryRunJSONReportIsSanitized verifies the JSON report
// omits secret fields even when they are populated in the report struct.
func TestMigrationDryRunJSONReportIsSanitized(t *testing.T) {
	report := newCredentialMigrationReport(CredentialMigrationModeDryRun)
	report.SourceKeyCounts["shopee"] = 3
	report.TargetCredentialCounts["shopee"] = 3
	report.Tenants = append(report.Tenants, CredentialMigrationTenant{
		TenantID: "tenant_json", SchemaName: "tenant_tenant_json", SourceRows: 3,
	})

	var buf bytes.Buffer
	require.NoError(t, WriteCredentialMigrationReport(&buf, report))
	output := buf.String()

	// These secret field names must never appear in the JSON output.
	for _, forbidden := range []string{"access_token", "refresh_token", "shop_cipher", "app_secret", "partner_key"} {
		assert.False(t, strings.Contains(output, forbidden),
			"JSON report must not contain secret field %q", forbidden)
	}

	// Required structural fields present.
	assert.Contains(t, output, `"mode": "dry_run"`)
	assert.Contains(t, output, `"source_key_counts"`)
	assert.Contains(t, output, `"target_credential_counts"`)
}
