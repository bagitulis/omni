// =============================================================================
// 🔄 SQLite to PostgreSQL Data Migration Tool
// =============================================================================
// Usage:
//   go run cmd/migrate/main.go --tenant yumna_bertigamart
//   go run cmd/migrate/main.go --tenant yumna_bertigamart --verify
//   go run cmd/migrate/main.go --tenant yumna_bertigamart --dry-run
// =============================================================================

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	config := parseFlags()

	migrator, err := NewMigrator(config)
	if err != nil {
		log.Fatalf("❌ Failed to create migrator: %v", err)
	}

	if err := migrator.Connect(); err != nil {
		log.Fatalf("❌ Connection failed: %v", err)
	}
	defer migrator.Close()

	schema := fmt.Sprintf("tenant_%s", config.TenantID)
	if err := migrator.SetPostgresSchema(schema); err != nil {
		log.Fatalf("❌ Schema setup failed: %v", err)
	}

	if config.VerifyOnly {
		if err := migrator.VerifyMigration(); err != nil {
			os.Exit(1)
		}
		return
	}

	runMigration(migrator, config)
}

func parseFlags() MigrationConfig {
	sqlitePath := flag.String("sqlite", "", "Path to SQLite database file")
	postgresDSN := flag.String("postgres", "", "PostgreSQL connection string")
	tenantID := flag.String("tenant", "", "Tenant ID to migrate")
	batchSize := flag.Int("batch", 1000, "Batch size for migration")
	verifyOnly := flag.Bool("verify", false, "Only verify migration without migrating")
	dryRun := flag.Bool("dry-run", false, "Simulate migration without writing")
	continueOnError := flag.Bool("continue-on-error", false, "Continue migration on errors")
	basePath := flag.String("base-path", "backend/config/databases", "Base path for SQLite databases")

	flag.Parse()

	if *tenantID == "" {
		printUsage()
		os.Exit(1)
	}

	if *sqlitePath == "" {
		*sqlitePath = filepath.Join(*basePath, fmt.Sprintf("%s.db", *tenantID))
	}

	if *postgresDSN == "" {
		*postgresDSN = buildPostgresDSN()
	}

	return MigrationConfig{
		SQLitePath:      *sqlitePath,
		PostgresDSN:     *postgresDSN,
		TenantID:        *tenantID,
		BatchSize:       *batchSize,
		VerifyOnly:      *verifyOnly,
		DryRun:          *dryRun,
		ContinueOnError: *continueOnError,
	}
}

func printUsage() {
	fmt.Println("Usage: migrate --tenant <tenant_id> [options]")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	fmt.Println("\nExamples:")
	fmt.Println("  migrate --tenant yumna_bertigamart")
	fmt.Println("  migrate --tenant yumna_bertigamart --postgres 'host=localhost port=5432 user=omni password=secret dbname=omni_main sslmode=disable'")
	fmt.Println("  migrate --tenant yumna_bertigamart --verify")
}

func buildPostgresDSN() string {
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "omni")
	pgPass := getEnv("POSTGRES_PASSWORD", "omni_secure_2026")
	pgDB := getEnv("POSTGRES_DB", "omni_main")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		pgHost, pgPort, pgUser, pgPass, pgDB)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runMigration(migrator *Migrator, config MigrationConfig) {
	log.Printf("🚀 Starting migration for tenant: %s", config.TenantID)
	if config.DryRun {
		log.Println("⚠️  DRY RUN MODE - No data will be written")
	}

	if err := migrator.MigrateAllTables(); err != nil {
		log.Printf("❌ Migration failed: %v", err)
	}

	migrator.PrintReport()

	if !config.DryRun {
		log.Println("\n🔍 Running post-migration verification...")
		if err := migrator.VerifyMigration(); err != nil {
			log.Printf("⚠️  Verification issues found: %v", err)
		}
	}

	log.Println("\n✨ Migration process completed!")
}
