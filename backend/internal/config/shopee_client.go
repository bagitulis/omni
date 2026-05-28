package config

import (
	"context"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

// GetShopeeClient returns a configured Shopee client for the given tenant.
// All tenant marketplace credentials are loaded from canonical credential tables.
func GetShopeeClient(tenantID, basePath string) (*shopee.Client, error) {
	if tenantID == "" {
		return nil, ErrMissingTenantID
	}

	// Get tenant database
	tenantDB, err := GetTenantDB(tenantID, basePath)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}

	client, err := loadShopeeCanonicalCredentials(context.Background(), tenantDB, tenantID)
	if err != nil {
		return nil, fmt.Errorf("load canonical shopee credentials: %w", err)
	}
	return client, nil
}

// loadShopeeCanonicalCredentials loads Shopee credentials from canonical tenant tables
func loadShopeeCanonicalCredentials(ctx context.Context, tenantDB *gorm.DB, tenantID string) (*shopee.Client, error) {
	repo := repositories.NewCredentialRepository(tenantDB)

	// Get app config (partner_id, partner_key)
	appConfig, err := repo.GetAppConfig(ctx, tenantID, "shopee")
	if err != nil || appConfig == nil {
		return nil, fmt.Errorf("canonical app config not available")
	}

	// Check partner ID is configured
	partnerID := appConfig.PartnerID
	if partnerID == 0 {
		return nil, fmt.Errorf("canonical partner_id not configured")
	}

	// Get first active store connection
	conns, err := repo.ListConnections(ctx, tenantID, "shopee")
	if err != nil || len(conns) == 0 {
		return nil, fmt.Errorf("canonical store connection not available")
	}

	// Create client with per-tenant canonical credentials
	shopID, _ := strconv.ParseInt(conns[0].StoreIdentifier, 10, 64)
	client := shopee.NewClient(partnerID, appConfig.PartnerKey, true)
	client.SetShopCredentials(shopID, conns[0].AccessToken)
	return client, nil
}
