package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/services"
)

func main() {
	modeFlag := flag.String("mode", "dry-run", "migration mode: dry-run, backfill, cutover, rollback")
	jsonFlag := flag.Bool("json", false, "emit sanitized JSON report")
	writeFlag := flag.Bool("write", false, "allow write modes after preflight gates pass")
	flag.Parse()

	mode, err := parseMode(*modeFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
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

	report, err := services.NewCredentialMigrationService(basePath).Run(context.Background(), services.CredentialMigrationOptions{
		Mode:       mode,
		BasePath:   basePath,
		Actor:      "credential_migration_command",
		ActorRole:  "service",
		AllowWrite: *writeFlag,
	})
	if *jsonFlag {
		_ = services.WriteCredentialMigrationReport(os.Stdout, report)
	} else {
		_ = services.WriteCredentialMigrationTextReport(os.Stdout, report)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func parseMode(value string) (services.CredentialMigrationMode, error) {
	switch strings.ToLower(strings.ReplaceAll(value, "-", "_")) {
	case "dry_run":
		return services.CredentialMigrationModeDryRun, nil
	case "backfill":
		return services.CredentialMigrationModeBackfill, nil
	case "cutover":
		return services.CredentialMigrationModeCutover, nil
	case "rollback":
		return services.CredentialMigrationModeRollback, nil
	default:
		return "", fmt.Errorf("unsupported migration mode: %s", value)
	}
}
