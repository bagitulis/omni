package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformConfigRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &PlatformConfig{})
	repo := NewPlatformConfigRepository(db)
	ctx := context.Background()

	// Helper to create test config
	createConfig := func(t *testing.T, tenantID, platform string) *PlatformConfig {
		config := &PlatformConfig{
			ID:           uuid.New().String(),
			TenantID:     tenantID,
			Platform:     platform,
			ShopID:       "shop-123",
			ShopIDInt:    123,
			ShopName:     "Test Shop",
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-123",
			ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
			Region:       "ID",
			IsActive:     true,
		}
		err := repo.Create(ctx, config)
		require.NoError(t, err)
		return config
	}

	t.Run("Create", func(t *testing.T) {
		config := &PlatformConfig{
			ID:           uuid.New().String(),
			TenantID:     "tenant-create-1",
			Platform:     "shopee",
			ShopID:       "shop-456",
			AccessToken:  "token-456",
			RefreshToken: "refresh-456",
			IsActive:     true,
		}

		err := repo.Create(ctx, config)
		assert.NoError(t, err)
		assert.False(t, config.CreatedAt.IsZero())
		assert.False(t, config.UpdatedAt.IsZero())
	})

	t.Run("FindByTenantAndPlatform", func(t *testing.T) {
		config := createConfig(t, "tenant-find-1", "shopee")

		// Find existing
		found, err := repo.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, config.ID, found.ID)
		assert.Equal(t, config.ShopID, found.ShopID)

		// Not found
		notFound, err := repo.FindByTenantAndPlatform(ctx, "nonexistent-tenant", "shopee")
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("FindByTenant", func(t *testing.T) {
		tenantID := "tenant-findall-1"
		createConfig(t, tenantID, "shopee")
		createConfig(t, tenantID, "lazada")

		configs, err := repo.FindByTenant(ctx, tenantID)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(configs), 2)
	})

	t.Run("Update", func(t *testing.T) {
		config := createConfig(t, "tenant-update-1", "shopee")
		originalUpdatedAt := config.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		config.ShopName = "Updated Shop Name"
		err := repo.Update(ctx, config)
		assert.NoError(t, err)
		assert.True(t, config.UpdatedAt.After(originalUpdatedAt))

		// Verify
		found, _ := repo.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
		assert.Equal(t, "Updated Shop Name", found.ShopName)
	})

	t.Run("UpdateTokens", func(t *testing.T) {
		config := createConfig(t, "tenant-tokens-1", "lazada")

		newExpiry := time.Now().Add(48 * time.Hour).Unix()
		err := repo.UpdateTokens(ctx, config.TenantID, config.Platform, "new-access-token", "new-refresh-token", newExpiry)
		assert.NoError(t, err)

		// Verify
		found, _ := repo.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
		assert.Equal(t, "new-access-token", found.AccessToken)
		assert.Equal(t, "new-refresh-token", found.RefreshToken)
		assert.Equal(t, newExpiry, found.ExpiresAt)
	})

	t.Run("Upsert", func(t *testing.T) {
		// Create via upsert
		config := &PlatformConfig{
			ID:           uuid.New().String(),
			TenantID:     "tenant-upsert-1",
			Platform:     "tiktok",
			ShopID:       "shop-upsert",
			AccessToken:  "upsert-token",
			RefreshToken: "upsert-refresh",
			IsActive:     true,
		}
		err := repo.Upsert(ctx, config)
		assert.NoError(t, err)

		// Update via upsert
		config.ShopName = "Upserted Shop"
		err = repo.Upsert(ctx, config)
		assert.NoError(t, err)

		// Verify
		found, _ := repo.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
		assert.Equal(t, "Upserted Shop", found.ShopName)
	})

	t.Run("Delete", func(t *testing.T) {
		config := createConfig(t, "tenant-delete-1", "shopee")

		err := repo.Delete(ctx, config.TenantID, config.Platform)
		assert.NoError(t, err)

		// Verify deleted
		found, _ := repo.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
		assert.Nil(t, found)
	})

	t.Run("SetActiveStatus", func(t *testing.T) {
		config := createConfig(t, "tenant-status-1", "lazada")
		assert.True(t, config.IsActive)

		// Deactivate
		err := repo.SetActiveStatus(ctx, config.TenantID, config.Platform, false)
		assert.NoError(t, err)

		found, _ := repo.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
		assert.False(t, found.IsActive)

		// Reactivate
		err = repo.SetActiveStatus(ctx, config.TenantID, config.Platform, true)
		assert.NoError(t, err)

		found, _ = repo.FindByTenantAndPlatform(ctx, config.TenantID, config.Platform)
		assert.True(t, found.IsActive)
	})
}
