package config

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

// GetShopeeClient returns a configured Shopee client for the given tenant
// Loads REAL credentials from database with ENV fallback - NO hardcoded test values
func GetShopeeClient(tenantID, basePath string) (*shopee.Client, error) {
	if tenantID == "" {
		return nil, ErrMissingTenantID
	}

	// Get tenant database
	tenantDB, err := GetTenantDB(tenantID, basePath)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}

	// Get system database for global config
	systemDB, err := GetSystemDB(basePath)
	if err != nil {
		return nil, fmt.Errorf("get system DB: %w", err)
	}

	// Try canonical credentials first (per-tenant credential_app_configs + credential_connections)
	if creds, err := loadShopeeCanonicalCredentials(context.Background(), tenantDB, tenantID); err == nil {
		return creds, nil
	}
	// Load global credentials (partner_id, partner_key) with ENV fallback
	globalCreds, err := loadShopeeGlobalCredentials(systemDB)
	if err != nil {
		return nil, fmt.Errorf("load global credentials: %w", err)
	}

	// Load tenant credentials (shop_id, access_token)
	tenantCreds, err := loadShopeeTenantCredentials(context.Background(), tenantDB)
	if err != nil {
		return nil, fmt.Errorf("load tenant credentials: %w", err)
	}

	// Create client with REAL credentials from database
	client := shopee.NewClient(globalCreds.PartnerID, globalCreds.PartnerKey, true)
	client.SetShopCredentials(tenantCreds.ShopID, tenantCreds.AccessToken)

	return client, nil
}

// loadShopeeCanonicalCredentials loads Shopee credentials from canonical tenant tables
// Returns (client, nil) on success, (nil, error) to trigger legacy fallback
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

// loadShopeeGlobalCredentials loads partner_id and partner_key from system DB with ENV fallback
func loadShopeeGlobalCredentials(db *gorm.DB) (*ShopeeGlobalCreds, error) {
	var configs []models.GlobalConfig
	if err := db.Where("platform = ?", "shopee").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("query global config: %w", err)
	}

	creds := &ShopeeGlobalCreds{}
	for _, cfg := range configs {
		switch cfg.ConfigKey {
		case "partner_id", "partnerId":
			if pid, err := strconv.ParseInt(cfg.ConfigValue, 10, 64); err == nil {
				creds.PartnerID = pid
			}
		case "partner_key", "partnerKey":
			creds.PartnerKey = cfg.ConfigValue
		}
	}

	// Fallback to ENV if not in database (matching Node backend behavior)
	if creds.PartnerID == 0 {
		if envPartnerID := os.Getenv("SHOPEE_PARTNER_ID"); envPartnerID != "" {
			if pid, err := strconv.ParseInt(envPartnerID, 10, 64); err == nil {
				creds.PartnerID = pid
			}
		}
	}
	if creds.PartnerKey == "" {
		if envPartnerKey := os.Getenv("SHOPEE_PARTNER_KEY"); envPartnerKey != "" {
			creds.PartnerKey = envPartnerKey
		}
	}

	if creds.PartnerID == 0 || creds.PartnerKey == "" {
		return nil, fmt.Errorf("Shopee credentials not configured (checked database and ENV vars)")
	}

	return creds, nil
}

// loadShopeeTenantCredentials loads shop_id and access_token from tenant DB
func loadShopeeTenantCredentials(ctx context.Context, db *gorm.DB) (*ShopeeTenantCredentials, error) {
	repo := repositories.NewTenantPlatformConfigRepository(db)
	tokenInfo, err := repo.GetTokenInfo(ctx, "shopee")
	if err != nil {
		return nil, fmt.Errorf("get token info: %w", err)
	}
	if tokenInfo == nil {
		return nil, fmt.Errorf("Shopee tenant credentials not found in database")
	}

	return &ShopeeTenantCredentials{
		ShopID:      tokenInfo.ShopID,
		AccessToken: tokenInfo.AccessToken,
	}, nil
}

// ShopeeGlobalCreds holds global Shopee credentials (partner level)
type ShopeeGlobalCreds struct {
	PartnerID  int64
	PartnerKey string
}

// ShopeeTenantCredentials holds shop-level credentials
type ShopeeTenantCredentials struct {
	ShopID      int64
	AccessToken string
}
