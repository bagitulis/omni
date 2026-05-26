package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
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

func setupCredentialLifecycleTest(t *testing.T) (*CredentialLifecycleService, *repositories.CredentialRepository, context.Context) {
	t.Helper()
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	t.Setenv("ENCRYPTION_KEY", key)
	db := testutils.SetupTestPostgresWithModels(t, &models.CredentialConnection{}, &models.CredentialAuditEvent{})
	require.NoError(t, db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_credential_connections_active
		 ON credential_connections (tenant_id, platform, store_identifier)
		 WHERE disabled_at IS NULL`).Error)
	repo := repositories.NewCredentialRepository(db)
	return NewCredentialLifecycleService(db), repo, context.Background()
}

func createLifecycleConnection(t *testing.T, repo *repositories.CredentialRepository, ctx context.Context, tenantID, platform, storeID string) {
	t.Helper()
	nowMs := time.Now().Add(time.Hour).UnixMilli()
	require.NoError(t, repo.CreateConnection(ctx, &models.CredentialConnection{
		ID:              uuid.New().String(),
		TenantID:        tenantID,
		Platform:        platform,
		StoreIdentifier: storeID,
		StoreName:       "Test Store",
		Status:          "connected",
		AccessToken:     "original-access-token",
		RefreshToken:    "original-refresh-token",
		TokenExpiry:     nowMs,
		RefreshExpiry:   time.Now().Add(24 * time.Hour).UnixMilli(),
		Version:         1,
		CreatedBy:       "setup",
		UpdatedBy:       "setup",
	}))
}

func TestCredentialLifecycleServiceConcurrentRefreshPersistsOneFinalVersion(t *testing.T) {
	svc, repo, ctx := setupCredentialLifecycleTest(t)
	createLifecycleConnection(t, repo, ctx, "tenant-refresh", models.PlatformShopee, "shop-123")
	createLifecycleConnection(t, repo, ctx, "tenant-refresh", models.PlatformShopee, "shop-456")

	var calls int32
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.RefreshConnection(ctx, "tenant-refresh", models.PlatformShopee, "shop-123", "service", "service", func(conn *models.CredentialConnection) (*CredentialRefreshOutcome, error) {
				call := atomic.AddInt32(&calls, 1)
				time.Sleep(25 * time.Millisecond)
				assert.Equal(t, "shop-123", conn.StoreIdentifier)
				return &CredentialRefreshOutcome{
					AccessToken:    "new-access-token-" + string(rune('0'+call)),
					RefreshToken:   "new-refresh-token-" + string(rune('0'+call)),
					TokenExpiry:    time.Now().Add(2 * time.Hour).UnixMilli(),
					RefreshExpiry:  time.Now().Add(24 * time.Hour).UnixMilli(),
					Status:         "connected",
					Code:           "refresh_success",
				}, nil
			})
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	for err := range results {
		require.NoError(t, err)
	}

	conn, err := repo.GetConnection(ctx, "tenant-refresh", models.PlatformShopee, "shop-123")
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.Equal(t, 3, conn.Version)
	assert.True(t, strings.HasPrefix(conn.AccessToken, "new-access-token-"))
	assert.Equal(t, "connected", conn.Status)
	other, err := repo.GetConnection(ctx, "tenant-refresh", models.PlatformShopee, "shop-456")
	require.NoError(t, err)
	require.NotNil(t, other)
	assert.Equal(t, "original-access-token", other.AccessToken)
	assert.Equal(t, 1, other.Version)
	assert.Equal(t, "connected", other.Status)
	events, err := repo.ListAuditEvents(ctx, "tenant-refresh", models.PlatformShopee, 10, 0)
	require.NoError(t, err)
	require.Len(t, events, 2)
	for _, event := range events {
		assert.Equal(t, "refresh", event.EventType)
		assert.Equal(t, "success", event.Status)
		assert.Equal(t, "refresh_success", event.Code)
		assert.Equal(t, "***123", event.Metadata["store_identifier_mask"])
		assertAuditContainsNoSecrets(t, event)
	}
}

func TestCredentialLifecycleServiceRefreshDuringDisconnectDoesNotReenableCredential(t *testing.T) {
	_, repo, ctx := setupCredentialLifecycleTest(t)
	createLifecycleConnection(t, repo, ctx, "tenant-disconnect", models.PlatformTiktok, "store-777")

	conn, err := repo.GetConnection(ctx, "tenant-disconnect", models.PlatformTiktok, "store-777")
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.NoError(t, repo.DisableConnectionWithVersion(ctx, conn, conn.Version, "disconnect_requested"))
	conn.AccessToken = "stale-access-token"
	conn.RefreshToken = "stale-refresh-token"
	conn.TokenExpiry = time.Now().Add(2 * time.Hour).UnixMilli()
	conn.RefreshExpiry = time.Now().Add(24 * time.Hour).UnixMilli()
	err = repo.UpdateConnectionTokensWithVersion(ctx, conn, 1, "connected", "refresh_success", nil)
	require.Error(t, err)

	conn, err = repo.GetConnection(ctx, "tenant-disconnect", models.PlatformTiktok, "store-777")
	require.NoError(t, err)
	assert.Nil(t, conn)
}

func TestCredentialLifecycleServiceRefreshFailurePreservesTokenAndAudits(t *testing.T) {
	svc, repo, ctx := setupCredentialLifecycleTest(t)
	createLifecycleConnection(t, repo, ctx, "tenant-failure", models.PlatformLazada, "seller-321")

	_, err := svc.RefreshConnection(ctx, "tenant-failure", models.PlatformLazada, "seller-321", "service", "service", func(conn *models.CredentialConnection) (*CredentialRefreshOutcome, error) {
		return nil, errors.New("provider rejected secret-token-value")
	})
	require.NoError(t, err)
	conn, err := repo.GetConnection(ctx, "tenant-failure", models.PlatformLazada, "seller-321")
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.Equal(t, "original-access-token", conn.AccessToken)
	assert.Equal(t, 2, conn.Version)
	assert.Equal(t, "refresh_failed", conn.Status)

	events, err := repo.ListAuditEvents(ctx, "tenant-failure", models.PlatformLazada, 10, 0)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "refresh_failed", events[0].EventType)
	assert.Equal(t, "failed", events[0].Status)
	assert.Equal(t, "refresh_failed", events[0].Code)
	assert.Equal(t, "***321", events[0].Metadata["store_identifier_mask"])
	assertAuditContainsNoSecrets(t, events[0])
}

func assertAuditContainsNoSecrets(t *testing.T, event models.CredentialAuditEvent) {
	t.Helper()
	data, err := json.Marshal(event)
	require.NoError(t, err)
	serialized := string(data)
	for _, forbidden := range []string{"access-token", "refresh-token", "secret-token-value", "original-access-token", "original-refresh-token", "stale-access-token"} {
		assert.NotContains(t, serialized, forbidden)
	}
}
