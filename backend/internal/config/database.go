package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/gorm"
)

// ErrMissingTenantID is returned when tenant ID is empty
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
var ErrMissingTenantID = errors.New("tenant_id is required - no default tenant allowed")

// DBDriver represents the database driver type
type DBDriver string

const (
	DriverPostgres DBDriver = "postgres"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Driver       DBDriver
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLife  time.Duration
	Schema       string
}

// Database connection pools
var (
	tenantDBs      = make(map[string]*gorm.DB)
	systemDB       *gorm.DB
	mu             sync.RWMutex
	globalDriver   DBDriver = DriverPostgres
	globalPGConfig *PostgresConfig
	migrationMode  bool
)

// SetDatabaseDriver sets the global database driver
func SetDatabaseDriver(driver DBDriver, pgConfig *PostgresConfig) {
	mu.Lock()
	defer mu.Unlock()
	globalDriver = driver
	globalPGConfig = pgConfig
	log.Printf("Database driver set to: %s", driver)
}

// GetDatabaseDriver returns the current database driver
func GetDatabaseDriver() DBDriver {
	mu.RLock()
	defer mu.RUnlock()
	return globalDriver
}

// GetTenantDB returns database connection for a specific tenant
// Uses dbPath from tenants.json to locate the correct database file
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func GetTenantDB(tenantID string, basePath string) (*gorm.DB, error) {
	// HARD GUARD: Reject empty tenant ID immediately
	if tenantID == "" {
		return nil, ErrMissingTenantID
	}

	mu.RLock()
	if db, exists := tenantDBs[tenantID]; exists {
		mu.RUnlock()
		return db, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	// Double-check after acquiring write lock
	if db, exists := tenantDBs[tenantID]; exists {
		return db, nil
	}

	var db *gorm.DB
	var err error

	// PostgreSQL only
	db, err = openPostgresForTenant(tenantID)

	if err != nil {
		return nil, fmt.Errorf("failed to open tenant database: %w", err)
	}

	tenantDBs[tenantID] = db
	log.Printf("Connected to tenant database: %s (driver: %s)", tenantID, globalDriver)
	return db, nil
}

// GetTenantDBWithContext returns a tenant DB with proper schema context for PostgreSQL
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func GetTenantDBWithContext(tenantID string, basePath string) (*gorm.DB, error) {
	// HARD GUARD: Reject empty tenant ID immediately
	if tenantID == "" {
		return nil, ErrMissingTenantID
	}

	db, err := GetTenantDB(tenantID, basePath)
	if err != nil {
		return nil, err
	}

	if globalDriver == DriverPostgres {
		if err := SetTenantSchema(db, tenantID); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// SetTenantSchema sets PostgreSQL search_path for a tenant
// Directly sets search_path without calling set_tenant_context function
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func SetTenantSchema(db *gorm.DB, tenantID string) error {
	// HARD GUARD: Reject empty tenant ID
	if tenantID == "" {
		return ErrMissingTenantID
	}

	if globalDriver != DriverPostgres {
		return nil
	}

	// Direct SET search_path - more reliable than function call
	schemaName := fmt.Sprintf("tenant_%s", tenantID)
	result := db.Session(&gorm.Session{}).Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName))
	if result.Error != nil {
		return fmt.Errorf("failed to set schema: %w", result.Error)
	}

	return nil
}

// GetSystemDB returns the system/global database connection
func GetSystemDB(basePath string) (*gorm.DB, error) {
	if systemDB != nil {
		return systemDB, nil
	}

	mu.Lock()
	defer mu.Unlock()

	if systemDB != nil {
		return systemDB, nil
	}

	var db *gorm.DB
	var err error

	// PostgreSQL only
	db, err = openPostgresForSystem()

	if err != nil {
		return nil, fmt.Errorf("failed to open system database: %w", err)
	}

	systemDB = db
	log.Printf("Connected to system database (driver: %s)", globalDriver)
	return db, nil
}

// CloseAllDBs closes all database connections
func CloseAllDBs() {
	mu.Lock()
	defer mu.Unlock()

	for tenantID, db := range tenantDBs {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		log.Printf("Closed tenant database: %s", tenantID)
	}
	tenantDBs = make(map[string]*gorm.DB)

	if systemDB != nil {
		if sqlDB, err := systemDB.DB(); err == nil {
			sqlDB.Close()
		}
		log.Println("Closed system database")
		systemDB = nil
	}
}

// ResetTenantConnection closes and removes a specific tenant connection
func ResetTenantConnection(tenantID string) {
	mu.Lock()
	defer mu.Unlock()

	if db, exists := tenantDBs[tenantID]; exists {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		delete(tenantDBs, tenantID)
		log.Printf("Reset tenant connection: %s", tenantID)
	}
}

// GetTenantDBByID returns database connection for tenant using default base path
// Convenience wrapper around GetTenantDB for use without explicit basePath
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func GetTenantDBByID(tenantID string) (*gorm.DB, error) {
	// HARD GUARD: Reject empty tenant ID immediately
	if tenantID == "" {
		return nil, ErrMissingTenantID
	}

	basePath := os.Getenv("CONFIG_PATH")
	if basePath == "" {
		basePath = "/app/config"
	}
	return GetTenantDBWithContext(tenantID, basePath)
}

// CheckDatabaseHealth verifies database connectivity
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func CheckDatabaseHealth(tenantID string, basePath string) error {
	// HARD GUARD: Reject empty tenant ID
	if tenantID == "" {
		return ErrMissingTenantID
	}

	db, err := GetTenantDB(tenantID, basePath)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

// GetConnectionStats returns database connection statistics
// AGENTS.MD: TIDAK ADA DEFAULT TENANT - harus error jika kosong
func GetConnectionStats(tenantID string) (map[string]interface{}, error) {
	// HARD GUARD: Reject empty tenant ID
	if tenantID == "" {
		return nil, ErrMissingTenantID
	}

	mu.RLock()
	db, exists := tenantDBs[tenantID]
	mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no connection for tenant: %s", tenantID)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"driver":              string(globalDriver),
		"tenant_id":           tenantID,
		"max_open":            stats.MaxOpenConnections,
		"open":                stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}, nil
}

// SetMigrationMode enables/disables migration mode
func SetMigrationMode(enabled bool) {
	mu.Lock()
	defer mu.Unlock()
	migrationMode = enabled
	if enabled {
		log.Println("⚠️ Migration mode ENABLED")
	} else {
		log.Println("✅ Migration mode DISABLED")
	}
}

// IsMigrationMode returns true if running in migration mode
func IsMigrationMode() bool {
	mu.RLock()
	defer mu.RUnlock()
	return migrationMode
}
