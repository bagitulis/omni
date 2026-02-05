package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.OAuthState{}, &models.OAuthLog{})
	repo := NewOAuthRepository(db)
	ctx := context.Background()

	t.Run("CreateState", func(t *testing.T) {
		state, err := repo.CreateState(ctx, "tenant-oauth-1", "shopee", "https://example.com/callback")
		assert.NoError(t, err)
		require.NotNil(t, state)
		assert.NotEmpty(t, state.ID)
		assert.NotEmpty(t, state.State)
		assert.Equal(t, "tenant-oauth-1", state.TenantID)
		assert.Equal(t, "shopee", state.Platform)
		assert.True(t, state.ExpiresAt.After(time.Now()))
	})

	t.Run("FindStateByState", func(t *testing.T) {
		state, err := repo.CreateState(ctx, "tenant-oauth-2", "lazada", "https://example.com/callback")
		require.NoError(t, err)

		// Find existing
		found, err := repo.FindStateByState(ctx, state.State)
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, state.ID, found.ID)
		assert.Equal(t, "lazada", found.Platform)

		// Not found
		notFound, err := repo.FindStateByState(ctx, "nonexistent-state")
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("DeleteState", func(t *testing.T) {
		state, err := repo.CreateState(ctx, "tenant-oauth-3", "tiktok", "https://example.com/callback")
		require.NoError(t, err)

		err = repo.DeleteState(ctx, state.ID)
		assert.NoError(t, err)

		// Verify deleted
		found, err := repo.FindStateByState(ctx, state.State)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("DeleteExpiredStates", func(t *testing.T) {
		// Create an expired state directly
		expiredState := &models.OAuthState{
			ID:          "expired-state-id",
			TenantID:    "tenant-expired",
			Platform:    "shopee",
			State:       "expired-state-value",
			RedirectURL: "https://example.com",
			ExpiresAt:   time.Now().Add(-1 * time.Hour), // Already expired
			CreatedAt:   time.Now().Add(-2 * time.Hour),
		}
		err := db.WithContext(ctx).Create(expiredState).Error
		require.NoError(t, err)

		// Delete expired states
		err = repo.DeleteExpiredStates(ctx)
		assert.NoError(t, err)

		// Verify deleted
		found, err := repo.FindStateByState(ctx, "expired-state-value")
		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("CreateLog", func(t *testing.T) {
		log := &models.OAuthLog{
			TenantID:  "tenant-log-1",
			Platform:  "shopee",
			EventType: models.OAuthEventCallback,
			ShopID:    "shop-123",
			Code:      "auth-code-123",
			State:     "state-value",
			Status:    models.OAuthStatusReceived,
		}

		err := repo.CreateLog(ctx, log)
		assert.NoError(t, err)
		assert.NotEmpty(t, log.ID)
		assert.False(t, log.CreatedAt.IsZero())
	})

	t.Run("UpdateLogStatus", func(t *testing.T) {
		log := &models.OAuthLog{
			TenantID:  "tenant-log-2",
			Platform:  "lazada",
			EventType: models.OAuthEventTokenExchange,
			Status:    models.OAuthStatusReceived,
		}
		err := repo.CreateLog(ctx, log)
		require.NoError(t, err)

		// Update to success
		err = repo.UpdateLogStatus(ctx, log.ID, models.OAuthStatusSuccess, "")
		assert.NoError(t, err)

		// Verify
		var updated models.OAuthLog
		db.Where("id = ?", log.ID).First(&updated)
		assert.Equal(t, models.OAuthStatusSuccess, updated.Status)
		assert.NotNil(t, updated.ProcessedAt)

		// Update to failed with error
		log2 := &models.OAuthLog{
			TenantID:  "tenant-log-3",
			Platform:  "tiktok",
			EventType: models.OAuthEventError,
			Status:    models.OAuthStatusReceived,
		}
		err = repo.CreateLog(ctx, log2)
		require.NoError(t, err)

		err = repo.UpdateLogStatus(ctx, log2.ID, models.OAuthStatusFailed, "Token exchange failed")
		assert.NoError(t, err)

		var failed models.OAuthLog
		db.Where("id = ?", log2.ID).First(&failed)
		assert.Equal(t, models.OAuthStatusFailed, failed.Status)
		assert.Equal(t, "Token exchange failed", failed.ErrorMsg)
	})

	t.Run("FindLogsByTenant", func(t *testing.T) {
		tenantID := "tenant-find-logs-1"
		for i := 0; i < 3; i++ {
			log := &models.OAuthLog{
				TenantID:  tenantID,
				Platform:  "shopee",
				EventType: models.OAuthEventCallback,
				Status:    models.OAuthStatusSuccess,
			}
			err := repo.CreateLog(ctx, log)
			require.NoError(t, err)
		}

		logs, err := repo.FindLogsByTenant(ctx, tenantID, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 3)
	})

	t.Run("FindLogsByPlatform", func(t *testing.T) {
		tenantID := "tenant-find-platform-1"
		log := &models.OAuthLog{
			TenantID:  tenantID,
			Platform:  "lazada",
			EventType: models.OAuthEventTokenRefresh,
			Status:    models.OAuthStatusSuccess,
		}
		err := repo.CreateLog(ctx, log)
		require.NoError(t, err)

		logs, err := repo.FindLogsByPlatform(ctx, tenantID, "lazada", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 1)
		for _, l := range logs {
			assert.Equal(t, "lazada", l.Platform)
		}
	})
}
