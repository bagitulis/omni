package handlers

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	"github.com/stretchr/testify/assert"
)

// TestFormatTokenStatusText_NilStatus tests with nil status
func TestFormatTokenStatusText_NilStatus(t *testing.T) {
	result := FormatTokenStatusText("shopee", nil)
	assert.Contains(t, result, "Shopee")
	assert.Contains(t, result, "Not configured")
}

// TestFormatTokenStatusText_ValidStatus tests with valid token
func TestFormatTokenStatusText_ValidStatus(t *testing.T) {
	status := &services.TokenInfo{
		IsValid:             true,
		ExpiresAt:           time.Now().Add(24 * time.Hour),
		RefreshTokenExpires: time.Now().Add(7 * 24 * time.Hour),
	}

	result := FormatTokenStatusText("shopee", status)
	assert.Contains(t, result, "Shopee")
	assert.Contains(t, result, "Valid")
	assert.Contains(t, result, "Expires in")
}

// TestFormatTokenStatusText_ExpiredStatus tests with expired token
func TestFormatTokenStatusText_ExpiredStatus(t *testing.T) {
	status := &services.TokenInfo{
		IsValid:             false,
		ExpiresAt:           time.Now().Add(-24 * time.Hour),
		RefreshTokenExpires: time.Now().Add(-7 * 24 * time.Hour),
	}

	result := FormatTokenStatusText("lazada", status)
	assert.Contains(t, result, "Lazada")
	assert.Contains(t, result, "Expired")
}

// TestFormatDurationHuman tests duration formatting
func TestFormatDurationHuman(t *testing.T) {
	tests := []struct {
		days     int
		hours    int
		mins     int
		expected string
	}{
		{0, 1, 30, "1h30m0s"},
		{1, 2, 15, "26h15m0s"},
		{0, 0, 45, "45m0s"},
	}

	for _, tt := range tests {
		result := FormatDurationHuman(tt.days, tt.hours, tt.mins)
		assert.Equal(t, tt.expected, result)
	}
}

// TestGetPlatformTitle tests platform title resolution
func TestGetPlatformTitle(t *testing.T) {
	assert.Equal(t, "Shopee", getPlatformTitle(models.PlatformShopee))
	assert.Equal(t, "Lazada", getPlatformTitle(models.PlatformLazada))
	assert.Equal(t, "TikTok", getPlatformTitle(models.PlatformTiktok))
	assert.Equal(t, "unknown", getPlatformTitle("unknown"))
}

// TestFormatISOTimestamp tests ISO timestamp formatting
func TestFormatISOTimestamp(t *testing.T) {
	testTime := time.Date(2025, 1, 15, 10, 30, 45, 123000000, time.UTC)
	result := FormatISOTimestamp(testTime)
	assert.Equal(t, "2025-01-15T10:30:45.123Z", result)
}

// TestValidatePlatform tests platform validation
func TestValidatePlatform(t *testing.T) {
	assert.True(t, ValidatePlatform(models.PlatformShopee))
	assert.True(t, ValidatePlatform(models.PlatformLazada))
	assert.True(t, ValidatePlatform(models.PlatformTiktok))
	assert.False(t, ValidatePlatform("invalid"))
	assert.False(t, ValidatePlatform(""))
}

// TestGetAllPlatforms tests getting all platforms
func TestGetAllPlatforms(t *testing.T) {
	platforms := GetAllPlatforms()
	assert.Len(t, platforms, 3)
	assert.Contains(t, platforms, models.PlatformShopee)
	assert.Contains(t, platforms, models.PlatformLazada)
	assert.Contains(t, platforms, models.PlatformTiktok)
}
