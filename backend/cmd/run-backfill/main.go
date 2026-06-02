package main

import (
	"context"
	"fmt"
	"os"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/services"
)

func main() {
	cfg := config.Load()
	pgConfig := &config.PostgresConfig{
		Host:     cfg.PGHost,
		Port:     cfg.PGPort,
		User:     cfg.PGUser,
		Password: cfg.PGPassword,
		DBName:   cfg.PGDatabase,
		SSLMode:  cfg.PGSSLMode,
	}
	config.SetDatabaseDriver(config.DriverPostgres, pgConfig)
	basePath := os.Getenv("CONFIG_PATH")
	if basePath == "" {
		basePath = "/app/config"
	}

	ctx := context.Background()
	tenantIDs := []string{"yumna_bertigamart", "tika_nusseyba"}

	for _, tenantID := range tenantIDs {
		fmt.Printf("=== Backfilling tenant: %s ===\n", tenantID)
		db, err := config.GetTenantDB(tenantID, basePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get DB for %s: %v\n", tenantID, err)
			continue
		}

		// Run BackfillAppConfigs
		err = services.BackfillAppConfigs(ctx, db, tenantID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "BackfillAppConfigs failed for %s: %v\n", tenantID, err)
		} else {
			fmt.Printf("BackfillAppConfigs: OK\n")
		}

		// Run BackfillConnections
		err = services.BackfillConnections(ctx, db, tenantID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "BackfillConnections failed for %s: %v\n", tenantID, err)
		} else {
			fmt.Printf("BackfillConnections: OK\n")
		}
	}

	fmt.Println("=== Backfill complete ===")
}
