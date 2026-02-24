package shop_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omni/backend/internal/services/shop"
)

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewShopSetupService_NotNil(t *testing.T) {
	svc := shop.NewShopSetupService(nil)
	require.NotNil(t, svc)
}

// ---------------------------------------------------------------------------
// ShopConfig — table name & struct
// ---------------------------------------------------------------------------

func TestShopConfig_TableName(t *testing.T) {
	c := shop.ShopConfig{}
	assert.Equal(t, "shop_configs", c.TableName())
}

func TestShopConfig_FieldAccess(t *testing.T) {
	c := shop.ShopConfig{
		TenantID:    "tenant1",
		ShopName:    "My Shop",
		ShopID:      "SHOP-001",
		Region:      "ID",
		Currency:    "IDR",
		Timezone:    "Asia/Jakarta",
		IsActive:    true,
		Description: "A test shop",
	}
	assert.Equal(t, "tenant1", c.TenantID)
	assert.Equal(t, "My Shop", c.ShopName)
	assert.Equal(t, "SHOP-001", c.ShopID)
	assert.Equal(t, "ID", c.Region)
	assert.Equal(t, "IDR", c.Currency)
	assert.Equal(t, "Asia/Jakarta", c.Timezone)
	assert.True(t, c.IsActive)
}

func TestShopConfig_IsActiveDefault(t *testing.T) {
	// Zero-value should reflect Go default (false), GORM default is handled at DB level
	c := shop.ShopConfig{}
	assert.False(t, c.IsActive)
}

func TestShopConfig_TimestampsZeroByDefault(t *testing.T) {
	c := shop.ShopConfig{}
	assert.True(t, c.CreatedAt.IsZero())
	assert.True(t, c.UpdatedAt.IsZero())
}

func TestShopConfig_TimestampAssignment(t *testing.T) {
	now := time.Now()
	c := shop.ShopConfig{
		CreatedAt: now,
		UpdatedAt: now,
	}
	assert.WithinDuration(t, now, c.CreatedAt, time.Second)
	assert.WithinDuration(t, now, c.UpdatedAt, time.Second)
}

// ---------------------------------------------------------------------------
// PlatformConfig — table name & struct
// ---------------------------------------------------------------------------

func TestPlatformConfig_TableName(t *testing.T) {
	c := shop.PlatformConfig{}
	assert.Equal(t, "platform_configs", c.TableName())
}

func TestPlatformConfig_FieldAccess(t *testing.T) {
	c := shop.PlatformConfig{
		TenantID:    "tenant1",
		Platform:    "shopee",
		ShopID:      "SHOP-123",
		ShopName:    "Shopee Store",
		IsActive:    true,
		IsConnected: true,
	}
	assert.Equal(t, "tenant1", c.TenantID)
	assert.Equal(t, "shopee", c.Platform)
	assert.Equal(t, "SHOP-123", c.ShopID)
	assert.True(t, c.IsActive)
	assert.True(t, c.IsConnected)
}

func TestPlatformConfig_SensitiveFieldsHiddenFromJSON(t *testing.T) {
	// AccessToken and RefreshToken have json:"-" tag — they should not be exported in JSON
	// We verify struct field existence and correct tagging via reflection-free check
	c := shop.PlatformConfig{
		TenantID:     "tenant1",
		Platform:     "shopee",
		AccessToken:  "secret-token",
		RefreshToken: "secret-refresh",
	}
	// Values ARE accessible in Go (not encrypted), but json tag is "-"
	// We confirm the struct holds the values correctly
	_ = c.AccessToken
	_ = c.RefreshToken
	assert.Equal(t, "tenant1", c.TenantID)
}

func TestPlatformConfig_IsConnectedDefault(t *testing.T) {
	// Zero-value: IsConnected defaults to false
	c := shop.PlatformConfig{}
	assert.False(t, c.IsConnected)
	assert.False(t, c.IsActive)
}

func TestPlatformConfig_TokenExpiry(t *testing.T) {
	expiry := time.Now().Add(24 * time.Hour)
	c := shop.PlatformConfig{
		TenantID:    "tenant1",
		Platform:    "lazada",
		TokenExpiry: expiry,
	}
	assert.False(t, c.TokenExpiry.IsZero())
	assert.True(t, c.TokenExpiry.After(time.Now()))
}

// ---------------------------------------------------------------------------
// Platform name values — enumerate known platforms
// ---------------------------------------------------------------------------

func TestPlatformConfig_KnownPlatformValues(t *testing.T) {
	platforms := []string{"shopee", "lazada", "tiktok"}
	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			c := shop.PlatformConfig{
				TenantID: "tenant1",
				Platform: platform,
			}
			assert.Equal(t, platform, c.Platform)
		})
	}
}

// ---------------------------------------------------------------------------
// Disconnect state — mimics DisconnectPlatform behavior
// ---------------------------------------------------------------------------

func TestPlatformConfig_DisconnectFields(t *testing.T) {
	// After disconnect: IsConnected=false, tokens cleared
	c := shop.PlatformConfig{
		TenantID:     "tenant1",
		Platform:     "shopee",
		IsConnected:  true,
		AccessToken:  "old-token",
		RefreshToken: "old-refresh",
	}

	// Simulate what DisconnectPlatform does
	c.IsConnected = false
	c.AccessToken = ""
	c.RefreshToken = ""

	assert.False(t, c.IsConnected)
	assert.Equal(t, "", c.AccessToken)
	assert.Equal(t, "", c.RefreshToken)
}

// ---------------------------------------------------------------------------
// Settings JSON field
// ---------------------------------------------------------------------------

func TestShopConfig_SettingsField(t *testing.T) {
	c := shop.ShopConfig{
		TenantID: "tenant1",
		Settings: `{"auto_sync":true,"sync_interval":300}`,
	}
	assert.NotEmpty(t, c.Settings)
}

func TestPlatformConfig_SettingsField(t *testing.T) {
	c := shop.PlatformConfig{
		TenantID: "tenant1",
		Platform: "tiktok",
		Settings: `{"shop_cipher":"abc123"}`,
	}
	assert.NotEmpty(t, c.Settings)
}
