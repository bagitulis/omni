package repositories

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCredentialTest(t *testing.T) (*CredentialRepository, context.Context) {
	t.Helper()
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	t.Setenv("ENCRYPTION_KEY", key)

	modelsList := []any{
		&models.CredentialConnection{},
		&models.CredentialAppConfig{},
		&models.OAuthConnectionAttempt{},
		&models.CredentialAuditEvent{},
	}
	db := testutils.SetupTestPostgresWithModels(t, modelsList...)

	// Create the partial unique index for active credential connections
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_credential_connections_active
		 ON credential_connections (tenant_id, platform, store_identifier)
		 WHERE disabled_at IS NULL`)

	repo := NewCredentialRepository(db)
	ctx := context.Background()
	return repo, ctx
}

func createTestConnection(t *testing.T, repo *CredentialRepository, ctx context.Context, tenantID, platform, storeID string) *models.CredentialConnection {
	t.Helper()
	conn := &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		Platform:        platform,
		StoreIdentifier: storeID,
		StoreName:       "Test Store",
		Status:          "connected",
		Region:          "id",
		AccessToken:     "test-access-token-" + storeID,
		RefreshToken:    "test-refresh-token-" + storeID,
		TokenExpiry:     1704067200000,
		CreatedBy:       "test-user",
		UpdatedBy:       "test-user",
	}
	err := repo.CreateConnection(ctx, conn)
	require.NoError(t, err)
	return conn
}

//------------------------------------------------------------------------------
// Cross-tenant isolation tests
//------------------------------------------------------------------------------

func TestCredentialRepository_CrossTenantIsolation(t *testing.T) {
	repo, ctx := setupCredentialTest(t)

	// Create connections for tenant A and tenant B
	createTestConnection(t, repo, ctx, "tenant-a", "shopee", "store-001")
	createTestConnection(t, repo, ctx, "tenant-b", "shopee", "store-001")

	t.Run("cannot_read_cross_tenant", func(t *testing.T) {
		// Tenant A should NOT see tenant B's connection
		conn, err := repo.GetConnection(ctx, "tenant-a", "shopee", "store-001")
		assert.NoError(t, err)
		require.NotNil(t, conn)
		assert.Equal(t, "tenant-a", conn.TenantID)

		// Wrong tenant — should return nil
		conn, err = repo.GetConnection(ctx, "tenant-b", "shopee", "store-002")
		assert.NoError(t, err)
		assert.Nil(t, conn)

		// Wrong platform — should return nil
		conn, err = repo.GetConnection(ctx, "tenant-a", "lazada", "store-001")
		assert.NoError(t, err)
		assert.Nil(t, conn)
	})

	t.Run("cannot_update_status_cross_tenant", func(t *testing.T) {
		// Tenant A tries to update tenant B's connection status — should fail
		err := repo.UpdateConnectionStatus(ctx, "tenant-a", "shopee", "store-999", "disconnected")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("cannot_update_disabled_connection", func(t *testing.T) {
		now := time.Now()
		disabled := &models.CredentialConnection{
			ID:              uuid.New().String(),
			TenantID:        "tenant-a",
			Platform:        "lazada",
			StoreIdentifier: "disabled-store",
			Status:          "disconnected",
			DisabledAt:      &now,
		}
		require.NoError(t, repo.CreateConnection(ctx, disabled))

		found, err := repo.GetConnection(ctx, "tenant-a", "lazada", "disabled-store")
		assert.NoError(t, err)
		assert.Nil(t, found)

		err = repo.UpdateConnectionStatus(ctx, "tenant-a", "lazada", "disabled-store", "connected")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

//------------------------------------------------------------------------------
// Missing tenant tests
//------------------------------------------------------------------------------

func TestCredentialRepository_MissingTenant(t *testing.T) {
	repo, ctx := setupCredentialTest(t)

	t.Run("get_connection_empty_tenant", func(t *testing.T) {
		_, err := repo.GetConnection(ctx, "", "shopee", "store-001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant_id is required")
	})

	t.Run("get_connection_empty_platform", func(t *testing.T) {
		_, err := repo.GetConnection(ctx, "tenant-a", "", "store-001")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "platform is required")
	})

	t.Run("get_app_config_empty_platform", func(t *testing.T) {
		_, err := repo.GetAppConfig(ctx, "tenant-a", "")
		require.EqualError(t, err, "platform is required")
	})

	t.Run("create_connection_empty_tenant", func(t *testing.T) {
		conn := &models.CredentialConnection{
			Platform:        "shopee",
			StoreIdentifier: "store-001",
		}
		err := repo.CreateConnection(ctx, conn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant_id is required")
	})

	t.Run("create_connection_empty_platform", func(t *testing.T) {
		conn := &models.CredentialConnection{
			TenantID:        "tenant-a",
			StoreIdentifier: "store-001",
		}
		err := repo.CreateConnection(ctx, conn)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "platform is required")
	})

	t.Run("upsert_app_config_empty_platform", func(t *testing.T) {
		err := repo.UpsertAppConfig(ctx, &models.CredentialAppConfig{TenantID: "tenant-a"})
		require.EqualError(t, err, "platform is required")
	})

	t.Run("list_connections_empty_tenant", func(t *testing.T) {
		_, err := repo.ListConnections(ctx, "", "shopee")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tenant_id is required")
	})

	t.Run("get_connection_empty_store_identifier", func(t *testing.T) {
		_, err := repo.GetConnection(ctx, "tenant-a", "shopee", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "store_identifier is required")
	})

	t.Run("update_connection_empty_platform", func(t *testing.T) {
		err := repo.UpdateConnection(ctx, &models.CredentialConnection{TenantID: "tenant-a", StoreIdentifier: "store-1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "platform is required")
	})

	t.Run("update_status_empty_platform", func(t *testing.T) {
		err := repo.UpdateConnectionStatus(ctx, "tenant-a", "", "store-1", "connected")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "platform is required")
	})
}

//------------------------------------------------------------------------------
// Duplicate store identifier tests
//------------------------------------------------------------------------------

func TestCredentialRepository_DuplicateStore(t *testing.T) {
	repo, ctx := setupCredentialTest(t)

	// Create first connection
	createTestConnection(t, repo, ctx, "tenant-d", "shopee", "dup-store")

	t.Run("duplicate_active_connection_fails", func(t *testing.T) {
		// Try to create another active connection with same scope — should fail
		dup := &models.CredentialConnection{
			ID:              uuid.New().String(),
			TenantID:        "tenant-d",
			Platform:        "shopee",
			StoreIdentifier: "dup-store",
			Status:          "connected",
			AccessToken:     "dup-token",
		}
		err := repo.CreateConnection(ctx, dup)
		assert.Error(t, err)
	})

	t.Run("duplicate_disabled_allowed", func(t *testing.T) {
		// A disabled record (disabled_at != NULL) with same scope should be allowed
		now := time.Now()
		disabled := &models.CredentialConnection{
			ID:              uuid.New().String(),
			TenantID:        "tenant-d",
			Platform:        "shopee",
			StoreIdentifier: "dup-store",
			Status:          "disconnected",
			AccessToken:     "old-token",
			DisabledAt:      &now,
		}
		err := repo.CreateConnection(ctx, disabled)
		assert.NoError(t, err)
	})
}

//------------------------------------------------------------------------------
// Incomplete record tests
//------------------------------------------------------------------------------

func TestCredentialRepository_IncompleteRecord(t *testing.T) {
	repo, ctx := setupCredentialTest(t)

	t.Run("create_minimal_connection", func(t *testing.T) {
		conn := &models.CredentialConnection{
			ID:              uuid.New().String(),
			TenantID:        "tenant-e",
			Platform:        "lazada",
			StoreIdentifier: "minimal-store",
		}
		err := repo.CreateConnection(ctx, conn)
		assert.NoError(t, err)

		// Read it back
		found, err := repo.GetConnection(ctx, "tenant-e", "lazada", "minimal-store")
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "disconnected", found.Status) // default
		assert.Empty(t, found.AccessToken)
		assert.Empty(t, found.RefreshToken)
	})
}

func TestCredentialRepository_UpdateConnectionIgnoresDisabledRows(t *testing.T) {
	repo, ctx := setupCredentialTest(t)
	now := time.Now()
	disabled := &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        "tenant-disabled",
		Platform:        "shopee",
		StoreIdentifier: "disabled-store",
		Status:          "disconnected",
		AccessToken:     "old-disabled-token",
		DisabledAt:      &now,
	}
	require.NoError(t, repo.CreateConnection(ctx, disabled))

	disabled.AccessToken = "new-disabled-token"
	disabled.Status = "connected"
	err := repo.UpdateConnection(ctx, disabled)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

//------------------------------------------------------------------------------
// Redaction tests
//------------------------------------------------------------------------------

func TestCredentialRepository_Redaction(t *testing.T) {
	repo, ctx := setupCredentialTest(t)

	conn := createTestConnection(t, repo, ctx, "tenant-f", "shopee", "redact-store")

	t.Run("json_marshal_excludes_secrets", func(t *testing.T) {
		data, err := json.Marshal(conn)
		require.NoError(t, err)
		jsonStr := string(data)

		// Secret field names must NOT appear in JSON
		assert.False(t, strings.Contains(jsonStr, "access_token"), "JSON must not contain access_token")
		assert.False(t, strings.Contains(jsonStr, "refresh_token"), "JSON must not contain refresh_token")
		assert.False(t, strings.Contains(jsonStr, "shop_cipher"), "JSON must not contain shop_cipher")

		// Safe fields MUST appear
		assert.True(t, strings.Contains(jsonStr, "store_identifier"), "JSON must contain store_identifier")
		assert.True(t, strings.Contains(jsonStr, "status"), "JSON must contain status")
		assert.True(t, strings.Contains(jsonStr, "platform"), "JSON must contain platform")
	})

	t.Run("masked_response_no_secrets", func(t *testing.T) {
		masked := conn.ToMaskedResponse()
		data, err := json.Marshal(masked)
		require.NoError(t, err)
		jsonStr := string(data)

		assert.True(t, strings.Contains(jsonStr, "store_identifier_mask"), "Masked response must contain store_identifier_mask")
		assert.True(t, strings.Contains(jsonStr, "***"), "Masked store identifier must have *** prefix")
		assert.True(t, strings.Contains(masked.StoreIdentifierMask, "***"), "StoreIdentifierMask must start with ***")
		assert.True(t, strings.HasSuffix(masked.StoreIdentifierMask, "ore"), "StoreIdentifierMask must end with last 3 chars")
	})
}

func TestCredentialRepository_AppConfigRedactionAndScope(t *testing.T) {
	repo, ctx := setupCredentialTest(t)

	cfg := &models.CredentialAppConfig{
		TenantID:   "tenant-app",
		Platform:   "lazada",
		Region:     "id",
		AppKey:     "raw-app-key",
		AppSecret:  "raw-app-secret",
		Configured: true,
		CreatedBy:  "developer",
		UpdatedBy:  "developer",
	}
	require.NoError(t, repo.UpsertAppConfig(ctx, cfg))

	found, err := repo.GetAppConfig(ctx, "tenant-app", "lazada")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "raw-app-key", found.AppKey)
	assert.Equal(t, "raw-app-secret", found.AppSecret)

	masked := found.ToMaskedResponse()
	data, err := json.Marshal(masked)
	require.NoError(t, err)
	jsonStr := string(data)
	assert.NotContains(t, jsonStr, "raw-app-key")
	assert.NotContains(t, jsonStr, "raw-app-secret")
	assert.Contains(t, jsonStr, "configured")

	otherTenant, err := repo.GetAppConfig(ctx, "tenant-other", "lazada")
	require.NoError(t, err)
	assert.Nil(t, otherTenant)
}

func TestCredentialRepository_MissingCanonicalTablesAreReadSafe(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t)
	repo := NewCredentialRepository(db)
	ctx := context.Background()

	app, err := repo.GetAppConfig(ctx, "tenant-missing", models.PlatformShopee)
	require.NoError(t, err)
	assert.Nil(t, app)

	connections, err := repo.ListConnections(ctx, "tenant-missing", models.PlatformShopee)
	require.NoError(t, err)
	assert.Empty(t, connections)

	conn, err := repo.GetConnection(ctx, "tenant-missing", models.PlatformShopee, "12345")
	require.NoError(t, err)
	assert.Nil(t, conn)

	events, err := repo.ListAuditEvents(ctx, "tenant-missing", models.PlatformShopee, 10, 0)
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestCredentialRepository_OAuthAttemptValidation(t *testing.T) {
	repo, ctx := setupCredentialTest(t)
	baseAttempt := &models.OAuthConnectionAttempt{TenantID: "tenant-oauth", Platform: "tiktok", AttemptID: "attempt-1"}

	t.Run("create_attempt_empty_platform", func(t *testing.T) {
		err := repo.CreateAttempt(ctx, &models.OAuthConnectionAttempt{TenantID: "tenant-oauth", AttemptID: "attempt-1"})
		require.EqualError(t, err, "platform is required")
	})

	t.Run("create_attempt_empty_attempt_id", func(t *testing.T) {
		err := repo.CreateAttempt(ctx, &models.OAuthConnectionAttempt{TenantID: "tenant-oauth", Platform: "tiktok"})
		require.EqualError(t, err, "attempt_id is required")
	})

	require.NoError(t, repo.CreateAttempt(ctx, baseAttempt))

	t.Run("get_attempt_empty_platform", func(t *testing.T) {
		_, err := repo.GetAttempt(ctx, "tenant-oauth", "", "attempt-1")
		require.EqualError(t, err, "platform is required")
	})

	t.Run("get_attempt_empty_attempt_id", func(t *testing.T) {
		_, err := repo.GetAttempt(ctx, "tenant-oauth", "tiktok", "")
		require.EqualError(t, err, "attempt_id is required")
	})

	t.Run("complete_attempt_empty_platform", func(t *testing.T) {
		err := repo.CompleteAttempt(ctx, "tenant-oauth", "", "attempt-1", "completed")
		require.EqualError(t, err, "platform is required")
	})

	t.Run("complete_attempt_empty_attempt_id", func(t *testing.T) {
		err := repo.CompleteAttempt(ctx, "tenant-oauth", "tiktok", "", "completed")
		require.EqualError(t, err, "attempt_id is required")
	})
}

func TestCredentialRepository_OAuthAttemptScope(t *testing.T) {
	repo, ctx := setupCredentialTest(t)
	attemptID := "attempt-" + uuid.New().String()
	attempt := &models.OAuthConnectionAttempt{
		TenantID:    "tenant-oauth",
		Platform:    "tiktok",
		AttemptID:   attemptID,
		SignedState: "signed-state-value",
		ExpiresAt:   time.Now().Add(10 * time.Minute),
		CreatedBy:   "tenant-admin",
	}

	require.NoError(t, repo.CreateAttempt(ctx, attempt))
	found, err := repo.GetAttempt(ctx, "tenant-oauth", "tiktok", attemptID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "pending", found.Status)

	wrongTenant, err := repo.GetAttempt(ctx, "tenant-other", "tiktok", attemptID)
	require.NoError(t, err)
	assert.Nil(t, wrongTenant)

	err = repo.CompleteAttempt(ctx, "tenant-other", "tiktok", attemptID, "completed")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	require.NoError(t, repo.CompleteAttempt(ctx, "tenant-oauth", "tiktok", attemptID, "completed"))
	completed, err := repo.GetAttempt(ctx, "tenant-oauth", "tiktok", attemptID)
	require.NoError(t, err)
	require.NotNil(t, completed)
	assert.Equal(t, "completed", completed.Status)
	assert.NotNil(t, completed.CompletedAt)

	data, err := json.Marshal(completed)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "signed-state-value")
}

func TestCredentialRepository_CreateAuditEvent(t *testing.T) {
	repo, ctx := setupCredentialTest(t)
	event := &models.CredentialAuditEvent{
		TenantID:        "tenant-audit",
		Platform:        "shopee",
		StoreIdentifier: "store-123",
		EventType:       "connect",
		Status:          "success",
		Code:            "connected",
		Actor:           "developer",
		ActorRole:       "developer",
		Metadata:        models.JSONMap{"store_identifier_mask": "***123"},
	}

	require.NoError(t, repo.CreateAuditEvent(ctx, event))
	events, err := repo.ListAuditEvents(ctx, "tenant-audit", "shopee", 10, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "connect", events[0].EventType)
	assert.Equal(t, "***123", events[0].Metadata["store_identifier_mask"])
}

//------------------------------------------------------------------------------
// Encryption round-trip tests
//------------------------------------------------------------------------------

func TestCredentialRepository_EncryptionRoundTrip(t *testing.T) {
	repo, ctx := setupCredentialTest(t)

	conn := &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        "tenant-g",
		Platform:        "tiktok",
		StoreIdentifier: "encrypt-store",
		AccessToken:     "raw-access-token-123",
		RefreshToken:    "raw-refresh-token-456",
		ShopCipher:      "raw-shop-cipher-789",
		Status:          "connected",
	}

	err := repo.CreateConnection(ctx, conn)
	require.NoError(t, err)

	// Read back and verify decryption
	found, err := repo.GetConnection(ctx, "tenant-g", "tiktok", "encrypt-store")
	require.NoError(t, err)
	require.NotNil(t, found)

	assert.Equal(t, "raw-access-token-123", found.AccessToken)
	assert.Equal(t, "raw-refresh-token-456", found.RefreshToken)
	assert.Equal(t, "raw-shop-cipher-789", found.ShopCipher)

	var stored models.CredentialConnection
	require.NoError(t, repo.db.WithContext(ctx).Where("tenant_id = ? AND platform = ? AND store_identifier = ?", "tenant-g", "tiktok", "encrypt-store").First(&stored).Error)
	assert.NotEqual(t, "raw-access-token-123", stored.AccessToken)
	assert.NotEqual(t, "raw-refresh-token-456", stored.RefreshToken)
	assert.NotEqual(t, "raw-shop-cipher-789", stored.ShopCipher)
	assert.True(t, utils.IsEncrypted(stored.AccessToken))
	assert.True(t, utils.IsEncrypted(stored.RefreshToken))
	assert.True(t, utils.IsEncrypted(stored.ShopCipher))
}

func TestCredentialConnectionMasking(t *testing.T) {
	conn := &models.CredentialConnection{StoreIdentifier: "store-12345"}
	assert.Equal(t, "***345", conn.StoreIdentifierMask())
	masked := conn.ToMaskedResponse()
	assert.Equal(t, "***345", masked.StoreIdentifierMask)
}
