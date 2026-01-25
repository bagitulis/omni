package handlers

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"gorm.io/gorm"
)

// DBGetter is a function type that returns a tenant DB connection
type DBGetter func(tenantID string) (*gorm.DB, error)

// GetTenantDB returns the tenant database with proper schema context
// This should be called in each handler method to get the correct DB
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func GetTenantDB(c *gin.Context) (*gorm.DB, error) {
	tenantID := c.GetString("tenantID")
	// HARD GUARD: Reject empty tenant ID immediately
	if tenantID == "" {
		return nil, config.ErrMissingTenantID
	}

	basePath := os.Getenv("DATABASE_PATH")
	if basePath == "" {
		basePath = "./data"
	}

	db, err := config.GetTenantDBWithContext(tenantID, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant DB: %w", err)
	}

	return db, nil
}

// GetTenantDBFromContext returns the tenant DB with schema already set
// This is the preferred method to use in handlers
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func GetTenantDBFromContext(c *gin.Context, fallbackDB *gorm.DB) (*gorm.DB, error) {
	tenantID := c.GetString("tenantID")
	// HARD GUARD: Reject empty tenant ID immediately
	if tenantID == "" {
		return nil, config.ErrMissingTenantID
	}

	// Check if using PostgreSQL
	if config.GetDatabaseDriver() == config.DriverPostgres {
		basePath := os.Getenv("DATABASE_PATH")
		if basePath == "" {
			basePath = "./data"
		}
		return config.GetTenantDBWithContext(tenantID, basePath)
	}

	// For SQLite, use the fallback DB (which is already the correct tenant DB)
	return fallbackDB, nil
}

// SetSchemaForTenant sets the PostgreSQL schema for the given tenant
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func SetSchemaForTenant(db *gorm.DB, tenantID string) error {
	// HARD GUARD: Reject empty tenant ID
	if tenantID == "" {
		return config.ErrMissingTenantID
	}

	if config.GetDatabaseDriver() != config.DriverPostgres {
		return nil
	}

	if err := config.SetTenantSchema(db, tenantID); err != nil {
		return fmt.Errorf("failed to set schema: %w", err)
	}

	return nil
}
