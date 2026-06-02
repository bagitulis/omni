package services

import (
	"context"
	"fmt"
	"os"

	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// BackfillAppConfigs reads platform_configs for the given tenant, identifies
// platforms with app-level credentials, and inserts corresponding
// credential_app_configs rows. Idempotent: skips platforms that already have
// an app config. Returns error if required app credential keys are missing.
func BackfillAppConfigs(ctx context.Context, db *gorm.DB, tenantID string) error {
	if tenantID == "" {
		return nil
	}

	credRows, _, err := readAndClassify(ctx, db)
	if err != nil {
		return err
	}
	if len(credRows) == 0 {
		return nil
	}

	credRepo := repositories.NewCredentialRepository(db)
	platformRows := groupByPlatform(credRows)

	for platform, pRows := range platformRows {
		requiredKeys, ok := appCredentialKeys[platform]
		if !ok {
			log.Warn().Str("platform", platform).Msg("Unknown platform, skipping app config backfill")
			continue
		}

		existing, err := credRepo.GetAppConfig(ctx, tenantID, platform)
		if err != nil {
			return fmt.Errorf("check existing app config for %s: %w", platform, err)
		}
		if existing != nil {
			log.Debug().Str("platform", platform).Msg("App config already exists, skipping")
			continue
		}

		configMap := buildConfigMap(pRows)

		// Validate encrypted credential fields can be decrypted
		valReport := validateCredentialConfigMap(configMap, requiredKeys)
		if hasBlockingFailures(valReport) {
			emitBackfillValidationReport(valReport, tenantID, "app_config_"+platform)
			return fmt.Errorf("corrupted encrypted credentials for platform %s — %d required field(s) failed to decrypt", platform, countBlockingFailures(valReport))
		}

		for _, requiredKey := range requiredKeys {
			if configMap[requiredKey] == "" {
				return fmt.Errorf("missing required app credential key %q for platform %s", requiredKey, platform)
			}
		}

		cfg := buildAppConfig(tenantID, platform, configMap)
		if err := credRepo.UpsertAppConfig(ctx, cfg); err != nil {
			return fmt.Errorf("upsert app config for %s: %w", platform, err)
		}

		log.Info().
			Str("tenant_id", tenantID).
			Str("platform", platform).
			Msg("Backfilled app config")
	}

	return nil
}

// BackfillConnections reads platform_configs for the given tenant and creates
// credential_connections rows for each platform-shop pair with tokens present.
// Idempotent: skips platforms that already have a connection.
// Platforms without a shopId are skipped. Missing tokens trigger a warning log.
func BackfillConnections(ctx context.Context, db *gorm.DB, tenantID string) error {
	if tenantID == "" {
		return nil
	}

	_, allRows, err := readAndClassify(ctx, db)
	if err != nil {
		return err
	}
	if len(allRows) == 0 {
		return nil
	}

	credRepo := repositories.NewCredentialRepository(db)
	platformRows := groupByPlatform(allRows)

	for platform, pRows := range platformRows {
		configMap := buildConfigMap(pRows)

		shopID, ok := configMap["shopId"]
		if !ok || shopID == "" {
			continue
		}

		if configMap["accessToken"] == "" && configMap["refreshToken"] == "" {
			log.Warn().
				Str("tenant_id", tenantID).
				Str("platform", platform).
				Str("shop_id", shopID).
				Msg("Skipping connection backfill — no tokens found")
			continue
		}

		// Validate encrypted credential fields can be decrypted
		connRequiredKeys := []string{"accessToken", "refreshToken"}
		valReport := validateCredentialConfigMap(configMap, connRequiredKeys)
		if hasBlockingFailures(valReport) {
			emitBackfillValidationReport(valReport, tenantID, "connection_"+platform+"_"+shopID)
			log.Error().
				Str("tenant_id", tenantID).
				Str("platform", platform).
				Str("shop_id", shopID).
				Msg("Skipping connection — required credential fields failed to decrypt")
			continue
		}
		// Optional field: shopCipher — silently clear if corrupted
		if platform == "tiktok" {
			if v := configMap["shopCipher"]; v != "" && utils.IsEncrypted(v) {
				if enc, encErr := utils.NewEncryptionService(os.Getenv("ENCRYPTION_KEY")); encErr == nil {
					if _, decErr := enc.Decrypt(v); decErr != nil {
						log.Warn().Str("tenant_id", tenantID).Str("platform", platform).Msg("[Backfill] shopCipher decrypt failed — inserting NULL")
						configMap["shopCipher"] = ""
					}
				}
			}
		}

		existing, err := credRepo.GetConnection(ctx, tenantID, platform, shopID)
		if err != nil {
			return fmt.Errorf("check existing connection for %s/%s: %w", platform, shopID, err)
		}
		if existing != nil {
			continue
		}

		conn := buildConnectionFromConfig(tenantID, platform, configMap)

		if err := credRepo.CreateConnection(ctx, conn); err != nil {
			return fmt.Errorf("create connection for %s/%s: %w", platform, shopID, err)
		}

		log.Info().
			Str("tenant_id", tenantID).
			Str("platform", platform).
			Str("shop_id", shopID).
			Msg("Backfilled connection")
	}

	return nil
}
