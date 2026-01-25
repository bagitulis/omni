package utils

import (
	"fmt"
)

// TenantTable returns the fully qualified table name for a tenant schema
// Format: tenant_{tenantID}.{tableName}
func TenantTable(tenantID, tableName string) string {
	return fmt.Sprintf("tenant_%s.%s", tenantID, tableName)
}

// TenantSchema returns the schema name for a tenant
// Format: tenant_{tenantID}
func TenantSchema(tenantID string) string {
	return fmt.Sprintf("tenant_%s", tenantID)
}

// SystemTable returns the fully qualified table name for the system schema
// Format: system.{tableName}
func SystemTable(tableName string) string {
	return fmt.Sprintf("system.%s", tableName)
}
