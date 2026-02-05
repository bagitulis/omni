package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalConfigRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.GlobalConfig{})
	repo := NewGlobalConfigRepository(db)
	ctx := context.Background()

	// Helper to create config
	createConfig := func(t *testing.T, platform, key, value string) {
		config := &models.GlobalConfig{
			ID:          platform + "_" + key,
			Platform:    platform,
			ConfigKey:   key,
			ConfigValue: value,
			IsEncrypted: false,
		}
		err := db.WithContext(ctx).Create(config).Error
		require.NoError(t, err)
	}

	t.Run("GetByPlatformAndKey", func(t *testing.T) {
		createConfig(t, "shopee", "testKey", "testValue")

		config, err := repo.GetByPlatformAndKey(ctx, "shopee", "testKey")
		assert.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, "testValue", config.ConfigValue)
	})

	t.Run("GetByPlatformAndKey_NotFound", func(t *testing.T) {
		_, err := repo.GetByPlatformAndKey(ctx, "nonexistent", "nonexistent")
		assert.Error(t, err)
	})

	t.Run("GetShopeeCredentials", func(t *testing.T) {
		createConfig(t, "shopee", "partnerId", "12345")
		createConfig(t, "shopee", "partnerKey", "secret-key")

		creds, err := repo.GetShopeeCredentials(ctx)
		assert.NoError(t, err)
		require.NotNil(t, creds)
		assert.Equal(t, int64(12345), creds.PartnerID)
		assert.Equal(t, "secret-key", creds.PartnerKey)
	})

	t.Run("GetLazadaCredentials", func(t *testing.T) {
		createConfig(t, "lazada", "appKey", "lazada-app-key")
		createConfig(t, "lazada", "appSecret", "lazada-app-secret")

		creds, err := repo.GetLazadaCredentials(ctx)
		assert.NoError(t, err)
		require.NotNil(t, creds)
		assert.Equal(t, "lazada-app-key", creds.AppKey)
		assert.Equal(t, "lazada-app-secret", creds.AppSecret)
	})

	t.Run("GetTiktokCredentials", func(t *testing.T) {
		createConfig(t, "tiktok", "appKey", "tiktok-app-key")
		createConfig(t, "tiktok", "appSecret", "tiktok-app-secret")

		creds, err := repo.GetTiktokCredentials(ctx)
		assert.NoError(t, err)
		require.NotNil(t, creds)
		assert.Equal(t, "tiktok-app-key", creds.AppKey)
		assert.Equal(t, "tiktok-app-secret", creds.AppSecret)
	})
}
