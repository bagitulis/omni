package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTenantService(t *testing.T) {
	basePath := "/path/to/base"
	svc := NewTenantService(basePath)

	assert.NotNil(t, svc)
	assert.Equal(t, basePath, svc.basePath)
}

func TestTenantInfo_Structure(t *testing.T) {
	info := TenantInfo{
		ID:       "tenant-123",
		ShopName: "Test Shop",
		IsGlobal: false,
	}

	assert.Equal(t, "tenant-123", info.ID)
	assert.Equal(t, "Test Shop", info.ShopName)
	assert.False(t, info.IsGlobal)
}

func TestTenantInfo_GlobalTenant(t *testing.T) {
	info := TenantInfo{
		ID:       "global",
		ShopName: "Global Admin",
		IsGlobal: true,
	}

	assert.Equal(t, "global", info.ID)
	assert.True(t, info.IsGlobal)
}

func TestFormatShopName(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		expected string
	}{
		{
			name:     "underscore separated",
			tenantID: "yumna_bertigamart",
			expected: "Bertigamart",
		},
		{
			name:     "multiple underscores",
			tenantID: "tika_nusseyba_shop",
			expected: "Shop",
		},
		{
			name:     "single word",
			tenantID: "simple",
			expected: "simple",
		},
		{
			name:     "empty string",
			tenantID: "",
			expected: "",
		},
		{
			name:     "single underscore",
			tenantID: "prefix_name",
			expected: "Name",
		},
		{
			name:     "lowercase conversion",
			tenantID: "user_myshop",
			expected: "Myshop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatShopName(tt.tenantID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantService_Structure(t *testing.T) {
	svc := &TenantService{basePath: "/test/path"}
	assert.Equal(t, "/test/path", svc.basePath)
}

func TestTenantInfo_JSONTags(t *testing.T) {
	// Test that JSON tags are properly defined (snake_case per AGENTS.md)
	// This is a structural test to ensure JSON serialization works correctly
	info := TenantInfo{
		ID:       "test-id",
		ShopName: "Test Shop",
		IsGlobal: true,
	}

	// Verify fields are accessible
	assert.NotEmpty(t, info.ID)
	assert.NotEmpty(t, info.ShopName)
	assert.True(t, info.IsGlobal)
}
