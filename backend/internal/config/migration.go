package config

import (
	"fmt"
	"log"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// MigrateSystemDatabase runs migrations for system database
func MigrateSystemDatabase(db *gorm.DB) error {
	log.Println("🔄 Running system database migrations...")

	// Create system schema for PostgreSQL
	if GetDatabaseDriver() == DriverPostgres {
		// Use Session() to prevent query condition accumulation
		if err := db.Session(&gorm.Session{}).Exec("CREATE SCHEMA IF NOT EXISTS system").Error; err != nil {
			return fmt.Errorf("failed to create system schema: %w", err)
		}
		if err := db.Session(&gorm.Session{}).Exec("SET search_path TO system, public").Error; err != nil {
			return fmt.Errorf("failed to set search_path: %w", err)
		}
	}

	// Migrate system models
	systemModels := []interface{}{
		&models.User{},
		&models.AuditLog{},
		&models.GlobalConfig{},
	}

	for _, model := range systemModels {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
		log.Printf("  ✅ Migrated: %T", model)
	}

	log.Println("✅ System database migrations completed")
	return nil
}

// MigrateTenantDatabase runs migrations for a tenant database
func MigrateTenantDatabase(db *gorm.DB, tenantID string) error {
	log.Printf("🔄 Running tenant database migrations for: %s", tenantID)

	// Create tenant schema for PostgreSQL
	if GetDatabaseDriver() == DriverPostgres {
		schemaName := fmt.Sprintf("tenant_%s", tenantID)
		// Use Session() to prevent query condition accumulation
		if err := db.Session(&gorm.Session{}).Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error; err != nil {
			return fmt.Errorf("failed to create tenant schema: %w", err)
		}
		if err := db.Session(&gorm.Session{}).Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName)).Error; err != nil {
			return fmt.Errorf("failed to set search_path: %w", err)
		}
	}

	// Migrate tenant-specific models
	tenantModels := []interface{}{
		// OAuth
		&models.OAuthState{},
		&models.OAuthLog{},

		// Platform Config
		&models.PlatformConfig{},

		// Webhooks
		&models.WebhookLog{},
		&models.WebhookOrderEvent{},
		&models.WebhookProductEvent{},
		&models.WebhookReturnEvent{},
		&models.WebhookMarketingEvent{},
		&models.WebhookShopeeEvent{},
		&models.WebhookWebchatEvent{},
		&models.WebhookFBSEvent{},

		// Analytics
		&models.AnalyticsSettings{},
		&models.ShopeeEscrowSync{},
		&models.ShopeeEscrowOrder{},
		&models.ShopeeEscrowItem{},
		&models.TiktokEscrowSync{},
		&models.TiktokEscrowOrder{},
		&models.TiktokEscrowItem{},

		// Shopee
		&models.ShopeeOrder{},
		&models.ShopeeOrderItem{},
		&models.ShopeeProduct{},
		&models.ShopeeSku{},

		// Lazada
		&models.LazadaOrder{},
		&models.LazadaOrderItem{},
		&models.LazadaProduct{},
		&models.LazadaSku{},

		// TikTok
		&models.TiktokOrder{},
		&models.TiktokOrderItem{},
		&models.TiktokProduct{},
		&models.TiktokSku{},

		// Inventory
		&models.InventorySettings{},
		&models.InventoryRecord{},
		&models.InventorySyncHistory{},
		&models.SheetSnapshot{},
		&models.InventorySkuPlatformStatus{},

		// Settings
		&models.GoogleSheetsSettings{},
		&models.FilterPreference{},
		&models.Spreadsheet{},
		&models.RouteConfig{},
		&models.WholesaleSettings{},

		// Unified Products
		&models.Product{},
		&models.ProductSKU{},

		// Image Gallery
		&models.Image{},

		// Orders
		&models.OrderTodayItem{},
		&models.LockedOrder{},

		// Ads
		&models.ShopeeAdsUploadBatch{},
		&models.ShopeeAdsProductData{},
		&models.TiktokAdsUploadBatch{},
		&models.TiktokAdsCreativeData{},
		&models.TiktokAdsProductSummary{},
		&models.TiktokAdsMLPrediction{},

		// Jobs
		&models.Job{},
		&models.JobHistory{},
		&models.AutoFunctionConfig{},
		&models.AutoFunctionHistory{},
		&models.RouteExecutionConfig{},
	}

	for _, model := range tenantModels {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("  ⚠️  Warning migrating %T: %v", model, err)
			// Continue with other models
			continue
		}
		log.Printf("  ✅ Migrated: %T", model)
	}

	log.Printf("✅ Tenant database migrations completed for: %s", tenantID)
	return nil
}

// RunAllMigrations runs migrations for system and all known tenants
func RunAllMigrations(basePath string, tenantIDs []string) error {
	// Migrate system database
	systemDB, err := GetSystemDB(basePath)
	if err != nil {
		return fmt.Errorf("failed to get system DB: %w", err)
	}

	if err := MigrateSystemDatabase(systemDB); err != nil {
		return fmt.Errorf("failed to migrate system DB: %w", err)
	}

	// Migrate each tenant database
	for _, tenantID := range tenantIDs {
		tenantDB, err := GetTenantDB(tenantID, basePath)
		if err != nil {
			log.Printf("⚠️  Warning: failed to get tenant DB %s: %v", tenantID, err)
			continue
		}

		if err := MigrateTenantDatabase(tenantDB, tenantID); err != nil {
			log.Printf("⚠️  Warning: failed to migrate tenant DB %s: %v", tenantID, err)
			continue
		}
	}

	return nil
}
