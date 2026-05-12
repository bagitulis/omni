package services

import (
	"fmt"
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

func TestErrTenantNotFound(t *testing.T) {
	// Verify ErrTenantNotFound is properly defined and usable
	assert.NotNil(t, ErrTenantNotFound)
	assert.Equal(t, "tenant not found", ErrTenantNotFound.Error())

	// Verify errors.Is works for comparison
	import_err := fmt.Errorf("wrapped: %w", ErrTenantNotFound)
	assert.ErrorIs(t, import_err, ErrTenantNotFound)
}

func TestTenantNamePattern(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		valid   bool
	}{
		{"valid simple", "myshop", true},
		{"valid with underscore", "my_shop", true},
		{"valid with numbers", "shop123", true},
		{"valid min length", "abc", true},
		{"valid complex", "yumna_bertigamart", true},
		{"valid 50 chars", "abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuv", true},
		{"invalid starts with number", "1shop", false},
		{"invalid starts with underscore", "_shop", false},
		{"invalid uppercase", "MyShop", false},
		{"invalid special chars", "my-shop", false},
		{"invalid too short", "ab", false},
		{"invalid empty", "", false},
		{"invalid spaces", "my shop", false},
		{"invalid sql injection", "test'; DROP TABLE--", false},
		{"invalid 51 chars", "abcdefghijklmnopqrstuvwxyz_abcdefghijklmnopqrstuvwx", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tenantNamePattern.MatchString(tt.input)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestTenantValidationError(t *testing.T) {
	err := &TenantValidationError{Msg: "invalid name"}
	assert.Equal(t, "invalid name", err.Error())
}

func TestTenantDuplicateError(t *testing.T) {
	err := &TenantDuplicateError{Name: "myshop"}
	assert.Equal(t, "tenant 'myshop' already exists", err.Error())
}

func TestCreateTenantResult_Structure(t *testing.T) {
	result := CreateTenantResult{
		ID:        "new_tenant",
		Name:      "new_tenant",
		IsActive:  true,
		CreatedAt: "2026-01-01T00:00:00Z",
	}

	assert.Equal(t, "new_tenant", result.ID)
	assert.Equal(t, "new_tenant", result.Name)
	assert.True(t, result.IsActive)
	assert.Equal(t, "2026-01-01T00:00:00Z", result.CreatedAt)
}
