package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/omni/backend/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		log.Fatal().Msg("ENCRYPTION_KEY environment variable is required")
	}

	basePath := os.Getenv("BASE_PATH")
	if basePath == "" {
		basePath = "/app/data"
	}

	tenantService := services.NewTenantService(basePath)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Info().Msg("[Preflight] Starting encrypted value validation...")

	report, err := services.ValidateEncryptedValuesPreflight(ctx, tenantService, encKey)
	if err != nil {
		log.Fatal().Err(err).Msg("[Preflight] Validation failed")
	}

	// Print report to stdout
	fmt.Println(services.FormatPreflightReport(report))

	// Save report to evidence file
	evidenceDir := ".sisyphus/evidence"
	if err := os.MkdirAll(evidenceDir, 0755); err != nil {
		log.Warn().Err(err).Msg("Failed to create evidence directory")
	}

	reportPath := evidenceDir + "/task-1-6-preflight-validation.txt"
	if err := services.SavePreflightReport(report, reportPath); err != nil {
		log.Error().Err(err).Str("path", reportPath).Msg("Failed to save report")
		os.Exit(1)
	}

	log.Info().Str("path", reportPath).Msg("[Preflight] Report saved")

	if len(report.BlockedTenants) > 0 {
		log.Error().
			Strs("blocked_tenants", report.BlockedTenants).
			Msg("[Preflight] MIGRATION BLOCKED — fix decrypt failures before proceeding")
		os.Exit(1)
	}

	if report.FernetKeyChanged {
		log.Error().Msg("[Preflight] CRITICAL: Fernet key change detected — ALL encrypted values unreadable")
		os.Exit(2)
	}

	log.Info().Msg("[Preflight] All encrypted values validated successfully")
}
