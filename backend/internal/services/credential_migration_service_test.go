package services

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCredentialMigrationReportIsSanitized(t *testing.T) {
	report := newCredentialMigrationReport(CredentialMigrationModeDryRun)
	report.SourceKeyCounts["shopee"] = 2
	report.TargetCredentialCounts["shopee"] = 1
	report.Tenants = append(report.Tenants, CredentialMigrationTenant{TenantID: "tenant_a", SchemaName: "tenant_tenant_a", SourceRows: 2})

	var buf bytes.Buffer
	require.NoError(t, WriteCredentialMigrationReport(&buf, report))
	output := buf.String()
	assert.Contains(t, output, "source_key_counts")
	assert.NotContains(t, output, "access_token")
	assert.NotContains(t, output, "refresh_token")
	assert.NotContains(t, output, "app_secret")
	assert.NotContains(t, output, "partner_key")

	var decoded CredentialMigrationReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, 2, decoded.SourceKeyCounts["shopee"])
}

func TestCredentialMigrationBackfillTenantIsIdempotent(t *testing.T) {
	db := setupCredentialMigrationSQLite(t)
	svc := &CredentialMigrationService{}
	bundles := []legacyCredentialBundle{{
		TenantID:        "tenant_a",
		Platform:        models.PlatformShopee,
		StoreIdentifier: "12345",
		StoreName:       "Test Store",
		AccessToken:     "test-access-token",
		RefreshToken:    "test-refresh-token",
		PartnerID:       1001,
		PartnerKey:      "test-partner-key",
		Region:          "id",
	}}

	first := newCredentialMigrationReport(CredentialMigrationModeBackfill)
	require.NoError(t, svc.backfillTenant(context.Background(), db, bundles, first, CredentialMigrationOptions{Actor: "test"}))
	assert.Equal(t, 1, first.Backfill.CreatedConnections)
	assert.Equal(t, 1, first.Backfill.CreatedAppConfigs)
	assert.Equal(t, 0, first.Backfill.SkippedExisting)

	second := newCredentialMigrationReport(CredentialMigrationModeBackfill)
	require.NoError(t, svc.backfillTenant(context.Background(), db, bundles, second, CredentialMigrationOptions{Actor: "test"}))
	assert.Equal(t, 0, second.Backfill.CreatedConnections)
	assert.Equal(t, 0, second.Backfill.CreatedAppConfigs)
	assert.Equal(t, 2, second.Backfill.SkippedExisting)

	repo := repositories.NewCredentialRepository(db)
	conn, err := repo.GetConnection(context.Background(), "tenant_a", models.PlatformShopee, "12345")
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.Equal(t, "connected", conn.Status)
	assert.Equal(t, "Test Store", conn.StoreName)
}

func TestCredentialMigrationKeyValueRowsBecomeSingleBundle(t *testing.T) {
	rows := []inventoryRowData{
		{"tenant_id": "tenant_a", "platform": models.PlatformShopee, "config_key": "accessToken", "config_value": "test-access-token"},
		{"tenant_id": "tenant_a", "platform": models.PlatformShopee, "config_key": "refreshToken", "config_value": "test-refresh-token"},
		{"tenant_id": "tenant_a", "platform": models.PlatformShopee, "config_key": "shopId", "config_value": "98765"},
		{"tenant_id": "tenant_a", "platform": models.PlatformShopee, "config_key": "shopName", "config_value": "Test Shop"},
		{"tenant_id": "tenant_a", "platform": models.PlatformShopee, "config_key": "partnerId", "config_value": "1001"},
		{"tenant_id": "tenant_a", "platform": models.PlatformShopee, "config_key": "partnerKey", "config_value": "test-partner-key"},
	}

	bundles, reasons := buildLegacyBundles("tenant_a", rows)
	require.Empty(t, reasons)
	require.Len(t, bundles, 1)
	assert.Equal(t, "98765", bundles[0].StoreIdentifier)
	assert.Equal(t, "Test Shop", bundles[0].StoreName)
	assert.Equal(t, "test-access-token", bundles[0].AccessToken)
	assert.Equal(t, "test-refresh-token", bundles[0].RefreshToken)
	assert.Equal(t, int64(1001), bundles[0].PartnerID)
	assert.Equal(t, "test-partner-key", bundles[0].PartnerKey)
	assert.NotEqual(t, models.PlatformShopee+"_legacy", bundles[0].StoreIdentifier)
}

func TestCredentialMigrationAbortGatesDetectBadSources(t *testing.T) {
	rows := []inventoryRowData{
		{"platform": "shopee", "config_key": "accessToken", "config_value": "not-fernet", "is_encrypted": true},
		{"platform": "lazada"},
	}
	_, reasons := buildLegacyBundles("tenant_a", rows)
	assert.Contains(t, reasons, "unreadable_encrypted_value")
	assert.Contains(t, reasons, "unknown_row_shape")
}

func TestCredentialLegacyFallbackDefaultAndOverride(t *testing.T) {
	t.Setenv(CredentialLegacyFallbackEnv, "")
	assert.False(t, IsCredentialLegacyFallbackEnabled())
	t.Setenv(CredentialLegacyFallbackEnv, "true")
	assert.True(t, IsCredentialLegacyFallbackEnabled())
	report := newCredentialMigrationReport(CredentialMigrationModeRollback)
	assert.True(t, report.Fallback.CanonicalFirst)
	assert.True(t, report.Fallback.LegacyReadOnly)
	assert.Equal(t, "credential_legacy_fallback_mode_reported", report.Fallback.AuditEvent)
}

func setupCredentialMigrationSQLite(t *testing.T) *gorm.DB {
	t.Helper()
	t.Setenv("ENCRYPTION_KEY", "test-key-32-chars-long-for-aes-256!!")
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.CredentialConnection{}, &models.CredentialAppConfig{}, &models.CredentialAuditEvent{}))
	return db
}
