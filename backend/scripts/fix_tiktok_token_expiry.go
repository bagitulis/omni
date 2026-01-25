//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TenantPlatformConfig matches the database table structure
type TenantPlatformConfig struct {
	ID          string    `gorm:"primaryKey"`
	Platform    string    `gorm:"index;not null"`
	ConfigKey   string    `gorm:"not null"`
	ConfigValue string    `gorm:"not null"`
	DataType    string    `gorm:"default:string"`
	IsEncrypted bool      `gorm:"default:true"`
	Metadata    *string
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}

func (TenantPlatformConfig) TableName() string {
	return "platform_configs"
}

func main() {
	// Load environment variables
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	// Get database connection info
	pgHost := getEnv("PG_HOST", "localhost")
	pgPort := getEnv("PG_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "omni")
	pgPassword := getEnv("POSTGRES_PASSWORD", "omni_secure_2026")
	pgDatabase := getEnv("POSTGRES_DB", "omni_main")

	// List of tenant schemas to fix
	tenantSchemas := []string{
		"tenant_yumna_bertigamart",
		"tenant_tika_nusseyba",
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPassword, pgDatabase)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()

	for _, schema := range tenantSchemas {
		fmt.Printf("\n========================================\n")
		fmt.Printf("Processing schema: %s\n", schema)
		fmt.Printf("========================================\n")

		// Set search path to tenant schema
		if err := db.Exec(fmt.Sprintf("SET search_path TO %s, public", schema)).Error; err != nil {
			log.Printf("Failed to set search_path for %s: %v", schema, err)
			continue
		}

		// Get current TikTok token expiry values
		var configs []TenantPlatformConfig
		if err := db.WithContext(ctx).Where("platform = ?", "tiktok").Find(&configs).Error; err != nil {
			log.Printf("Failed to read configs for %s: %v", schema, err)
			continue
		}

		fmt.Printf("Found %d TikTok config entries\n", len(configs))

		for _, cfg := range configs {
			if cfg.ConfigKey == "tokenExpiry" || cfg.ConfigKey == "refreshTokenExpiry" {
				currentValue, err := strconv.ParseInt(cfg.ConfigValue, 10, 64)
				if err != nil {
					fmt.Printf("  %s: invalid value '%s'\n", cfg.ConfigKey, cfg.ConfigValue)
					continue
				}

				currentDate := time.UnixMilli(currentValue)
				now := time.Now()

				fmt.Printf("\n  %s:\n", cfg.ConfigKey)
				fmt.Printf("    Current value (ms): %d\n", currentValue)
				fmt.Printf("    Parsed date: %s\n", currentDate.Format(time.RFC3339))

				// Check if the date is unreasonable (more than 10 years in the future)
				maxReasonable := now.Add(10 * 365 * 24 * time.Hour)
				if currentDate.After(maxReasonable) {
					fmt.Printf("    ⚠️  UNREASONABLE! Date is more than 10 years in the future\n")
					
					// Calculate new reasonable value
					var newExpiryMs int64
					if cfg.ConfigKey == "tokenExpiry" {
						// Access token: 7 days from now
						newExpiryMs = now.Add(7 * 24 * time.Hour).UnixMilli()
						fmt.Printf("    🔧 Fixing to 7 days from now\n")
					} else {
						// Refresh token: 90 days from now
						newExpiryMs = now.Add(90 * 24 * time.Hour).UnixMilli()
						fmt.Printf("    🔧 Fixing to 90 days from now\n")
					}

					newDate := time.UnixMilli(newExpiryMs)
					fmt.Printf("    New value (ms): %d\n", newExpiryMs)
					fmt.Printf("    New date: %s\n", newDate.Format(time.RFC3339))

					// Update the value
					if err := db.WithContext(ctx).Model(&cfg).Update("config_value", strconv.FormatInt(newExpiryMs, 10)).Error; err != nil {
						fmt.Printf("    ❌ Failed to update: %v\n", err)
					} else {
						fmt.Printf("    ✅ Updated successfully\n")
					}
				} else if currentDate.Before(now) {
					fmt.Printf("    ⚠️  EXPIRED! Date is in the past\n")
					fmt.Printf("    ℹ️  Token needs refresh via API\n")
				} else {
					daysRemaining := currentDate.Sub(now).Hours() / 24
					fmt.Printf("    ✅ OK - expires in %.1f days\n", daysRemaining)
				}
			}
		}
	}

	fmt.Printf("\n========================================\n")
	fmt.Printf("Done!\n")
	fmt.Printf("========================================\n")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
