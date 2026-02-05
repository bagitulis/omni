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

func TestRefreshSessionRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.RefreshSession{})
	repo := NewRefreshSessionRepository(db)
	ctx := context.Background()

	// Helper to create test session
	createSession := func(t *testing.T, userID, tenantID string) *models.RefreshSession {
		token := "test-token-" + time.Now().String()
		session := &models.RefreshSession{
			UserID:    userID,
			TenantID:  tenantID,
			TokenHash: HashToken(token),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IPAddress: "127.0.0.1",
			UserAgent: "Test Agent",
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)
		return session
	}

	t.Run("HashToken", func(t *testing.T) {
		token1 := "test-token-1"
		token2 := "test-token-2"

		hash1 := HashToken(token1)
		hash2 := HashToken(token2)

		assert.NotEmpty(t, hash1)
		assert.NotEmpty(t, hash2)
		assert.NotEqual(t, hash1, hash2)
		assert.Equal(t, hash1, HashToken(token1)) // Deterministic
	})

	t.Run("Create", func(t *testing.T) {
		session := &models.RefreshSession{
			UserID:    "user-create-1",
			TenantID:  "tenant-1",
			TokenHash: HashToken("create-test-token"),
			ExpiresAt: time.Now().Add(24 * time.Hour),
			IPAddress: "192.168.1.1",
			UserAgent: "Mozilla/5.0",
		}

		err := repo.Create(ctx, session)
		assert.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.LastUsedAt.IsZero())
	})

	t.Run("FindByTokenHash", func(t *testing.T) {
		token := "find-test-token-" + time.Now().String()
		tokenHash := HashToken(token)

		session := &models.RefreshSession{
			UserID:    "user-find-1",
			TenantID:  "tenant-find-1",
			TokenHash: tokenHash,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		err := repo.Create(ctx, session)
		require.NoError(t, err)

		// Find existing
		found, err := repo.FindByTokenHash(ctx, tokenHash)
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, session.ID, found.ID)

		// Not found
		notFound, err := repo.FindByTokenHash(ctx, HashToken("nonexistent-token"))
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("FindByUserID", func(t *testing.T) {
		userID := "user-finduserid-1"
		createSession(t, userID, "tenant-1")
		createSession(t, userID, "tenant-1")

		sessions, err := repo.FindByUserID(ctx, userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 2)
	})

	t.Run("Rotate", func(t *testing.T) {
		oldToken := "old-token-" + time.Now().String()
		oldHash := HashToken(oldToken)
		newToken := "new-token-" + time.Now().String()
		newHash := HashToken(newToken)

		// Create old session
		oldSession := &models.RefreshSession{
			UserID:    "user-rotate-1",
			TenantID:  "tenant-rotate-1",
			TokenHash: oldHash,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		err := repo.Create(ctx, oldSession)
		require.NoError(t, err)

		// Rotate
		newSession := &models.RefreshSession{
			UserID:    "user-rotate-1",
			TenantID:  "tenant-rotate-1",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		err = repo.Rotate(ctx, oldHash, newHash, newSession)
		assert.NoError(t, err)

		// Verify old session is marked as replaced
		oldFound, _ := repo.FindByTokenHash(ctx, oldHash)
		require.NotNil(t, oldFound)
		assert.NotNil(t, oldFound.ReplacedByHash)
		assert.Equal(t, newHash, *oldFound.ReplacedByHash)

		// Verify new session exists
		newFound, _ := repo.FindByTokenHash(ctx, newHash)
		require.NotNil(t, newFound)
	})

	t.Run("RevokeByTokenHash", func(t *testing.T) {
		session := createSession(t, "user-revoke-1", "tenant-revoke-1")

		err := repo.RevokeByTokenHash(ctx, session.TokenHash)
		assert.NoError(t, err)

		// Verify revoked
		found, _ := repo.FindByTokenHash(ctx, session.TokenHash)
		require.NotNil(t, found)
		assert.NotNil(t, found.RevokedAt)
		assert.True(t, found.IsRevoked())
	})

	t.Run("RevokeAllForUser", func(t *testing.T) {
		userID := "user-revokeall-1"
		createSession(t, userID, "tenant-1")
		createSession(t, userID, "tenant-1")

		err := repo.RevokeAllForUser(ctx, userID)
		assert.NoError(t, err)

		// Verify all sessions are revoked
		sessions, _ := repo.FindByUserID(ctx, userID)
		for _, s := range sessions {
			assert.NotNil(t, s.RevokedAt)
		}
	})

	t.Run("RevokeSessionChain", func(t *testing.T) {
		userID := "user-chain-1"
		createSession(t, userID, "tenant-chain-1")

		err := repo.RevokeSessionChain(ctx, userID)
		assert.NoError(t, err)

		// Verify sessions are revoked
		sessions, _ := repo.FindByUserID(ctx, userID)
		for _, s := range sessions {
			assert.NotNil(t, s.RevokedAt)
		}
	})

	t.Run("UpdateLastUsed", func(t *testing.T) {
		session := createSession(t, "user-lastused-1", "tenant-lastused-1")
		originalLastUsed := session.LastUsedAt

		time.Sleep(10 * time.Millisecond)

		err := repo.UpdateLastUsed(ctx, session.TokenHash)
		assert.NoError(t, err)

		found, _ := repo.FindByTokenHash(ctx, session.TokenHash)
		assert.True(t, found.LastUsedAt.After(originalLastUsed) || found.LastUsedAt.Equal(originalLastUsed))
	})

	t.Run("CleanupExpired", func(t *testing.T) {
		// Create expired session
		expiredSession := &models.RefreshSession{
			ID:        "expired-session-id",
			UserID:    "user-cleanup-1",
			TenantID:  "tenant-cleanup-1",
			TokenHash: HashToken("expired-cleanup-token"),
			ExpiresAt: time.Now().Add(-48 * time.Hour), // Expired 2 days ago
			CreatedAt: time.Now().Add(-72 * time.Hour),
		}
		err := db.WithContext(ctx).Create(expiredSession).Error
		require.NoError(t, err)

		// Cleanup with 1 day retention
		deleted, err := repo.CleanupExpired(ctx, 1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))

		// Verify deleted
		found, _ := repo.FindByTokenHash(ctx, HashToken("expired-cleanup-token"))
		assert.Nil(t, found)
	})

	t.Run("GetActiveSessions", func(t *testing.T) {
		userID := "user-active-1"
		createSession(t, userID, "tenant-active-1")

		sessions, err := repo.GetActiveSessions(ctx, userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(sessions), 1)
		for _, s := range sessions {
			assert.Nil(t, s.RevokedAt)
			assert.True(t, s.ExpiresAt.After(time.Now()))
		}
	})

	t.Run("CountActiveSessions", func(t *testing.T) {
		userID := "user-count-1"
		createSession(t, userID, "tenant-count-1")
		createSession(t, userID, "tenant-count-1")

		count, err := repo.CountActiveSessions(ctx, userID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(2))
	})
}
