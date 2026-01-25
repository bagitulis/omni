package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresConfig holds PostgreSQL specific configuration
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	Schema   string
}

// BuildDSN builds PostgreSQL connection string
func (c *PostgresConfig) BuildDSN() string {
	// Base DSN with connect timeout for reliability
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=10",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)

	if c.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s,public", c.Schema)
	}
	return dsn
}

// openPostgresForTenant opens PostgreSQL connection for a tenant
func openPostgresForTenant(tenantID string) (*gorm.DB, error) {
	if globalPGConfig == nil {
		return nil, fmt.Errorf("PostgreSQL config not set")
	}

	config := *globalPGConfig
	config.Schema = fmt.Sprintf("tenant_%s", tenantID)

	return openPostgresDatabase(&config)
}

// openPostgresForSystem opens PostgreSQL connection for system schema
func openPostgresForSystem() (*gorm.DB, error) {
	if globalPGConfig == nil {
		return nil, fmt.Errorf("PostgreSQL config not set")
	}

	config := *globalPGConfig
	config.Schema = "system"

	return openPostgresDatabase(&config)
}

// openPostgresDatabase opens a PostgreSQL database connection
func openPostgresDatabase(config *PostgresConfig) (*gorm.DB, error) {
	dsn := config.BuildDSN()

	// Set log level based on environment
	// Production: only warnings and errors (reduces log spam from frequent polling)
	// Development: info level for debugging
	logLevel := logger.Warn
	if os.Getenv("GO_ENV") != "production" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// Disable prepared statement cache to avoid stale connection issues
		PrepareStmt: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// PostgreSQL connection pool settings - optimized for container environment
	// Keep connections alive longer to avoid DNS resolution issues
	maxOpen := 25
	maxIdle := 10                   // Increased idle connections
	connMaxLife := 30 * time.Minute // Shorter lifetime to refresh connections
	connMaxIdle := 10 * time.Minute // Max time a connection can be idle

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(connMaxLife)
	sqlDB.SetConnMaxIdleTime(connMaxIdle)

	log.Printf("PostgreSQL connection established (schema: %s, pool: %d/%d)",
		config.Schema, maxOpen, maxIdle)

	return db, nil
}
