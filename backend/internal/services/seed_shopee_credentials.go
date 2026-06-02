package services

import (
	"context"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

// SeedShopeeAppCredentials seeds Shopee app credentials from environment variables.
// It is idempotent: existing rows are updated only if values differ.
// Missing env vars produce a warning and return nil (no crash).
func SeedShopeeAppCredentials(db *gorm.DB, tenantID string) error {
	env := os.Getenv("SHOPEE_ENV")
	if env == "" {
		env = "test"
	}

	var partnerIDStr, partnerKey string

	switch env {
	case "live":
		partnerIDStr = os.Getenv("SHOPEE_LIVE_PARTNER_ID")
		partnerKey = os.Getenv("SHOPEE_LIVE_PARTNER_KEY")
		if partnerIDStr == "" || partnerKey == "" {
			log.Warn().Str("reason", "SHOPEE_LIVE_PARTNER_ID or SHOPEE_LIVE_PARTNER_KEY not set").
				Msg("Shopee credentials not configured")
			return nil
		}
	default:
		partnerIDStr = os.Getenv("SHOPEE_PARTNER_ID")
		partnerKey = os.Getenv("SHOPEE_PARTNER_KEY")
		if partnerIDStr == "" || partnerKey == "" {
			log.Warn().Str("reason", "SHOPEE_PARTNER_ID or SHOPEE_PARTNER_KEY not set").
				Msg("Shopee credentials not configured")
			return nil
		}
	}

	partnerID, err := strconv.ParseInt(partnerIDStr, 10, 64)
	if err != nil {
		log.Warn().Str("partner_id", partnerIDStr).Err(err).
			Msg("Invalid Shopee partner_id, skipping seed")
		return nil
	}

	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = &models.CredentialAppConfig{
			TenantID: tenantID,
			Platform: models.PlatformShopee,
		}
	}

	cfg.PartnerID = partnerID
	cfg.PartnerKey = partnerKey
	cfg.Configured = true
	cfg.CreatedBy = "system-seed"
	cfg.UpdatedBy = "system-seed"

	if err := repo.UpsertAppConfig(ctx, cfg); err != nil {
		return err
	}

	log.Info().Int64("partner_id", partnerID).Str("env", env).
		Msg("Seeded Shopee app credentials")

	return nil
}
