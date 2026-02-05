package repositories

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformCredentialsRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.PlatformConfigKV{})
	repo := NewPlatformCredentialsRepository(db)
	ctx := context.Background()

	// Helper to create config
	createConfigKV := func(t *testing.T, platform, key, value string, encrypted bool) {
		config := &models.PlatformConfigKV{
			ID:          platform + "_" + key,
			Platform:    platform,
			ConfigKey:   key,
			ConfigValue: value,
			DataType:    "string",
			IsEncrypted: encrypted,
		}
		err := db.WithContext(ctx).Create(config).Error
		require.NoError(t, err)
	}

	t.Run("GetConfigValue", func(t *testing.T) {
		createConfigKV(t, "shopee", "testKey1", "testValue1", false)

		value, err := repo.GetConfigValue(ctx, "shopee", "testKey1")
		assert.NoError(t, err)
		assert.Equal(t, "testValue1", value)
	})

	t.Run("GetConfigValue_NotFound", func(t *testing.T) {
		value, err := repo.GetConfigValue(ctx, "nonexistent", "nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, value)
	})

	t.Run("GetAllConfigForPlatform", func(t *testing.T) {
		createConfigKV(t, "lazada", "key1", "value1", false)
		createConfigKV(t, "lazada", "key2", "value2", false)

		configs, err := repo.GetAllConfigForPlatform(ctx, "lazada")
		assert.NoError(t, err)
		assert.Contains(t, configs, "key1")
		assert.Contains(t, configs, "key2")
		assert.Equal(t, "value1", configs["key1"])
		assert.Equal(t, "value2", configs["key2"])
	})

	t.Run("GetShopeeCredentials", func(t *testing.T) {
		createConfigKV(t, "shopee", "shopId", "12345", false)
		createConfigKV(t, "shopee", "shopName", "Test Shopee Shop", false)
		createConfigKV(t, "shopee", "accessToken", "shopee-access-token", false)
		createConfigKV(t, "shopee", "refreshToken", "shopee-refresh-token", false)
		createConfigKV(t, "shopee", "region", "ID", false)

		creds, err := repo.GetShopeeCredentials(ctx)
		assert.NoError(t, err)
		require.NotNil(t, creds)
		assert.Equal(t, "shopee", creds.Platform)
		assert.Equal(t, "12345", creds.ShopID)
		assert.Equal(t, int64(12345), creds.ShopIDInt)
		assert.Equal(t, "Test Shopee Shop", creds.ShopName)
		assert.Equal(t, "shopee-access-token", creds.AccessToken)
		assert.Equal(t, "ID", creds.Region)
	})

	t.Run("GetLazadaCredentials", func(t *testing.T) {
		createConfigKV(t, "lazada", "shopId", "laz-shop-1", false)
		createConfigKV(t, "lazada", "accessToken", "lazada-access-token", false)
		createConfigKV(t, "lazada", "appKey", "lazada-app-key", false)
		createConfigKV(t, "lazada", "appSecret", "lazada-app-secret", false)
		createConfigKV(t, "lazada", "country", "ID", false) // legacy key

		creds, err := repo.GetLazadaCredentials(ctx)
		assert.NoError(t, err)
		require.NotNil(t, creds)
		assert.Equal(t, "lazada", creds.Platform)
		assert.Equal(t, "lazada-access-token", creds.AccessToken)
		assert.Equal(t, "lazada-app-key", creds.AppKey)
		assert.Equal(t, "ID", creds.Region) // Should pick up 'country' key
	})

	t.Run("GetTiktokCredentials", func(t *testing.T) {
		createConfigKV(t, "tiktok", "shopId", "tt-shop-1", false)
		createConfigKV(t, "tiktok", "shopCipher", "shop-cipher-123", false)
		createConfigKV(t, "tiktok", "accessToken", "tiktok-access-token", false)
		createConfigKV(t, "tiktok", "appKey", "tiktok-app-key", false)
		createConfigKV(t, "tiktok", "appSecret", "tiktok-app-secret", false)
		createConfigKV(t, "tiktok", "region", "ID", false)

		creds, err := repo.GetTiktokCredentials(ctx)
		assert.NoError(t, err)
		require.NotNil(t, creds)
		assert.Equal(t, "tiktok", creds.Platform)
		assert.Equal(t, "shop-cipher-123", creds.ShopCipher)
		assert.Equal(t, "tiktok-access-token", creds.AccessToken)
		assert.Equal(t, "tiktok-app-key", creds.AppKey)
	})

	t.Run("SetConfigValue_Create", func(t *testing.T) {
		err := repo.SetConfigValue(ctx, "newplatform", "newKey", "newValue", false)
		assert.NoError(t, err)

		value, err := repo.GetConfigValue(ctx, "newplatform", "newKey")
		assert.NoError(t, err)
		assert.Equal(t, "newValue", value)
	})

	t.Run("SetConfigValue_Update", func(t *testing.T) {
		createConfigKV(t, "updateplatform", "updateKey", "originalValue", false)

		err := repo.SetConfigValue(ctx, "updateplatform", "updateKey", "updatedValue", false)
		assert.NoError(t, err)

		value, err := repo.GetConfigValue(ctx, "updateplatform", "updateKey")
		assert.NoError(t, err)
		assert.Equal(t, "updatedValue", value)
	})
}
