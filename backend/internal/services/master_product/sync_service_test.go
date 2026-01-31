// Package master_product provides sync service tests
package master_product

import (
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestSyncService_SyncToPlatform_UnsupportedPlatform tests unsupported platform error
func TestSyncService_SyncToPlatform_UnsupportedPlatform(t *testing.T) {
	// Test that unsupported platform returns correct error
	// This is a unit test that doesn't need DB
	service := NewSyncService(nil, "")

	result, err := service.SyncToPlatform(nil, "test_tenant", 1, "invalid_platform")

	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedPlatform, err)
	assert.Nil(t, result)
}

// TestSyncService_Errors tests error constants are defined correctly
func TestSyncService_Errors(t *testing.T) {
	// Verify error messages are defined
	assert.NotNil(t, ErrUnsupportedPlatform)
	assert.NotNil(t, ErrMasterProductNotFound)
	assert.NotNil(t, ErrNoSkusToSync)
	assert.NotNil(t, ErrSyncFailed)
	assert.NotNil(t, ErrPlatformNotConfigured)

	// Check error messages are meaningful
	assert.Contains(t, ErrUnsupportedPlatform.Error(), "unsupported")
	assert.Contains(t, ErrMasterProductNotFound.Error(), "not found")
	assert.Contains(t, ErrNoSkusToSync.Error(), "no SKUs")
}

// TestSyncResult_Structure tests SyncResult struct
func TestSyncResult_Structure(t *testing.T) {
	result := &SyncResult{
		MasterProductID: 1,
		TargetPlatform:  models.PlatformShopee,
		Status:          "success",
		SkusSynced:      5,
	}

	assert.Equal(t, uint(1), result.MasterProductID)
	assert.Equal(t, models.PlatformShopee, result.TargetPlatform)
	assert.Equal(t, "success", result.Status)
	assert.Equal(t, 5, result.SkusSynced)
}

// TestNewSyncService tests service creation
func TestNewSyncService(t *testing.T) {
	service := NewSyncService(nil, "/test/path")

	assert.NotNil(t, service)
	assert.Equal(t, "/test/path", service.basePath)
}

// TestSyncService_PlatformConstants tests platform constants
func TestSyncService_PlatformConstants(t *testing.T) {
	// Verify platform constants are valid
	validPlatforms := []string{
		models.PlatformShopee,
		models.PlatformTiktok,
		models.PlatformLazada,
	}

	for _, platform := range validPlatforms {
		assert.NotEmpty(t, platform, "Platform constant should not be empty")
	}
}
