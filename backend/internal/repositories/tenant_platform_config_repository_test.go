package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantPlatformConfigRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &TenantPlatformConfig{})
	repo := NewTenantPlatformConfigRepository(db)
	ctx := context.Background()

	// Helper to create config
	createConfig := func(t *testing.T, platform, key, value string) {
		config := TenantPlatformConfig{
			ID:          platform + "_" + key + "_" + time.Now().String(),
			Platform:    platform,
			ConfigKey:   key,
			ConfigValue: value,
			DataType:    "string",
			IsEncrypted: false,
		}
		err := db.WithContext(ctx).Create(&config).Error
		require.NoError(t, err)
	}

	t.Run("GetConfig", func(t *testing.T) {
		createConfig(t, "shopee", "getConfigKey", "getConfigValue")

		value, err := repo.GetConfig(ctx, "shopee", "getConfigKey")
		assert.NoError(t, err)
		assert.Equal(t, "getConfigValue", value)
	})

	t.Run("GetConfig_NotFound", func(t *testing.T) {
		value, err := repo.GetConfig(ctx, "nonexistent", "nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, value)
	})

	t.Run("SetConfig_Create", func(t *testing.T) {
		err := repo.SetConfig(ctx, "newplatform", "setKey", "setValue", false)
		assert.NoError(t, err)

		value, err := repo.GetConfig(ctx, "newplatform", "setKey")
		assert.NoError(t, err)
		assert.Equal(t, "setValue", value)
	})

	t.Run("SetConfig_Update", func(t *testing.T) {
		createConfig(t, "updateplat", "updateKey", "originalValue")

		err := repo.SetConfig(ctx, "updateplat", "updateKey", "newValue", false)
		assert.NoError(t, err)

		value, err := repo.GetConfig(ctx, "updateplat", "updateKey")
		assert.NoError(t, err)
		assert.Equal(t, "newValue", value)
	})

	t.Run("GetAllConfigByPlatform", func(t *testing.T) {
		createConfig(t, "allconfig", "key1", "value1")
		createConfig(t, "allconfig", "key2", "value2")
		createConfig(t, "allconfig", "key3", "value3")

		configs, err := repo.GetAllConfigByPlatform(ctx, "allconfig")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(configs), 3)
		assert.Equal(t, "value1", configs["key1"])
		assert.Equal(t, "value2", configs["key2"])
		assert.Equal(t, "value3", configs["key3"])
	})

	t.Run("GetTokenInfo", func(t *testing.T) {
		// Create token-related configs
		createConfig(t, "tokenplat", "shopId", "123456")
		createConfig(t, "tokenplat", "accessToken", "token-123")
		createConfig(t, "tokenplat", "refreshToken", "refresh-456")
		createConfig(t, "tokenplat", "tokenExpiry", "1704067200000") // Milliseconds
		createConfig(t, "tokenplat", "shopCipher", "cipher-789")
		createConfig(t, "tokenplat", "region", "ID")

		info, err := repo.GetTokenInfo(ctx, "tokenplat")
		assert.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, int64(123456), info.ShopID)
		assert.Equal(t, "token-123", info.AccessToken)
		assert.Equal(t, "refresh-456", info.RefreshToken)
		assert.Equal(t, "cipher-789", info.ShopCipherOfSeller)
		assert.Equal(t, "ID", info.Region)
		assert.Equal(t, int64(1704067200000), info.TokenExpiry)
	})

	t.Run("UpdateTokens", func(t *testing.T) {
		err := repo.UpdateTokens(ctx, "newtokenplat", "new-access-token", "new-refresh-token", 3600, 86400)
		assert.NoError(t, err)

		info, err := repo.GetTokenInfo(ctx, "newtokenplat")
		assert.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "new-access-token", info.AccessToken)
		assert.Equal(t, "new-refresh-token", info.RefreshToken)
		assert.Greater(t, info.TokenExpiry, int64(0))
		assert.Greater(t, info.RefreshTokenExpiry, int64(0))
	})

	t.Run("parseExpiry", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected int64
		}{
			{
				name:     "parses milliseconds",
				input:    "1704067200000",
				expected: 1704067200000,
			},
			{
				name:     "parses ISO string",
				input:    "2024-01-01T00:00:00Z",
				expected: 1704067200000,
			},
			{
				name:     "returns 0 for invalid",
				input:    "invalid",
				expected: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := parseExpiry(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}
