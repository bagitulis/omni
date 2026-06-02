package services

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

// SeedShopeeAppCredentials validates that Shopee app credentials exist in the
// canonical credential_app_configs table. It reads from the database only and
// returns an error if no configuration is found for the given tenant.
func SeedShopeeAppCredentials(db *gorm.DB, tenantID string) error {
	repo := repositories.NewCredentialRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAppConfig(ctx, tenantID, models.PlatformShopee)
	if err != nil {
		return fmt.Errorf("failed to get Shopee app config for tenant %s: %w", tenantID, err)
	}

	if cfg == nil {
		return fmt.Errorf("no Shopee app config found for tenant %s", tenantID)
	}

	return nil
}
