//go:build tools
// +build tools

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load config
	cfg := config.Load()

	// Set up database driver (PostgreSQL only)
	pgConfig := &config.PostgresConfig{
		Host:     cfg.PGHost,
		Port:     cfg.PGPort,
		User:     cfg.PGUser,
		Password: cfg.PGPassword,
		DBName:   cfg.PGDatabase,
		SSLMode:  cfg.PGSSLMode,
	}
	config.SetDatabaseDriver(config.DriverPostgres, pgConfig)

	// Migrate system DB first
	systemDB, err := config.GetSystemDB(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to system database: %v", err)
	}
	if err := config.MigrateSystemDatabase(systemDB); err != nil {
		log.Fatalf("Failed to migrate system database: %v", err)
	}

	// Seed users for each tenant
	tenants := []string{"yumna_bertigamart", "tika_nusseyba"}

	for _, tenantID := range tenants {
		if err := seedTenantUser(tenantID, cfg.DatabasePath); err != nil {
			log.Printf("Failed to seed user for tenant %s: %v", tenantID, err)
		} else {
			log.Printf("✅ Successfully seeded user for tenant %s", tenantID)
		}
	}

	log.Println("Seed completed!")
}

func seedTenantUser(tenantID string, basePath string) error {
	// Migrate tenant schema first
	db, err := config.GetTenantDB(tenantID, basePath)
	if err != nil {
		return fmt.Errorf("failed to get tenant db: %w", err)
	}

	if err := config.MigrateTenantDatabase(db, tenantID); err != nil {
		return fmt.Errorf("failed to migrate tenant db: %w", err)
	}

	// Get DB with proper schema context
	db, err = config.GetTenantDBWithContext(tenantID, basePath)
	if err != nil {
		return fmt.Errorf("failed to get tenant db with context: %w", err)
	}

	// Check if user already exists
	var existingUser models.User
	result := db.Where("username = ?", "yumna").First(&existingUser)
	if result.Error == nil {
		log.Printf("User 'yumna' already exists in tenant %s", tenantID)
		return nil
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("DevOwner@2024"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:                  uuid.New().String(),
		Username:            "yumna",
		Email:               fmt.Sprintf("yumna@%s.local", tenantID),
		Password:            string(hashedPassword),
		Role:                models.RoleOwner,
		FailedLoginAttempts: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("   Username: %s\n", user.Username)
	fmt.Printf("   Email: %s\n", user.Email)
	fmt.Printf("   Role: %s\n", user.Role)

	return nil
}
