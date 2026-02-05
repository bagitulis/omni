package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTenantTable(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  string
		tableName string
		expected  string
	}{
		{
			name:      "basic tenant table",
			tenantID:  "tenant123",
			tableName: "users",
			expected:  "tenant_tenant123.users",
		},
		{
			name:      "orders table",
			tenantID:  "acme-corp",
			tableName: "orders",
			expected:  "tenant_acme-corp.orders",
		},
		{
			name:      "products table",
			tenantID:  "shop-001",
			tableName: "products",
			expected:  "tenant_shop-001.products",
		},
		{
			name:      "empty tenant id",
			tenantID:  "",
			tableName: "customers",
			expected:  "tenant_.customers",
		},
		{
			name:      "special characters in tenant",
			tenantID:  "tenant_with_underscore",
			tableName: "transactions",
			expected:  "tenant_tenant_with_underscore.transactions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TenantTable(tt.tenantID, tt.tableName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTenantSchema(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		expected string
	}{
		{
			name:     "basic tenant schema",
			tenantID: "tenant123",
			expected: "tenant_tenant123",
		},
		{
			name:     "tenant with dashes",
			tenantID: "acme-corp-001",
			expected: "tenant_acme-corp-001",
		},
		{
			name:     "tenant with numbers",
			tenantID: "shop2024",
			expected: "tenant_shop2024",
		},
		{
			name:     "empty tenant",
			tenantID: "",
			expected: "tenant_",
		},
		{
			name:     "tenant with underscores",
			tenantID: "my_company_name",
			expected: "tenant_my_company_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TenantSchema(tt.tenantID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSystemTable(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
		expected  string
	}{
		{
			name:      "users table",
			tableName: "users",
			expected:  "system.users",
		},
		{
			name:      "tenants table",
			tableName: "tenants",
			expected:  "system.tenants",
		},
		{
			name:      "audit_log table",
			tableName: "audit_log",
			expected:  "system.audit_log",
		},
		{
			name:      "settings table",
			tableName: "settings",
			expected:  "system.settings",
		},
		{
			name:      "empty table name",
			tableName: "",
			expected:  "system.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SystemTable(tt.tableName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
