package config

import (
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// MigrateSystemDatabase runs migrations for system database
func MigrateSystemDatabase(db *gorm.DB) error {
	log.Info().Msg("Running system database migrations...")

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
		&models.RefreshSession{},
	}

	for _, model := range systemModels {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", model, err)
		}
		log.Info().Msgf("  ✅ Migrated: %T", model)
	}

	// Add deactivated_at column to system.tenants if it doesn't exist
	if GetDatabaseDriver() == DriverPostgres {
		if err := db.Session(&gorm.Session{}).Exec(`ALTER TABLE system.tenants ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMP NULL`).Error; err != nil {
			log.Warn().Err(err).Msg("Failed to add deactivated_at column to system.tenants (may already exist)")
		}
	}

	// NOTE: Zombie column cleanup is intentionally RETIRED (2026-05-28).
	// GORM AutoMigrate is additive-only by design — it adds columns but never drops them.
	// Columns removed from model structs are retained in the database to avoid:
	//   - Destructive data loss from premature DROP COLUMN
	//   - GORM schema parsing edge cases with embedded structs/associations
	//   - Unproven column ownership across 60+ model types
	// Destructive cleanup requires a dedicated migration with disposable DB verification.
	// See migration_cleanup.go for the retired implementation (kept as reference).
	log.Info().Msg("System database migrations completed")
	return nil
}

// MigrateTenantDatabase runs migrations for a tenant database
func MigrateTenantDatabase(db *gorm.DB, tenantID string) error {
	log.Info().Msgf("🔄 Running tenant database migrations for: %s", tenantID)

	// Defense-in-depth: validate tenantID format before using in SQL
	// Even though middleware validates, this prevents injection if called from other paths
	if !isValidTenantID(tenantID) {
		return fmt.Errorf("invalid tenant ID format: %s", tenantID)
	}

	// Create tenant schema for PostgreSQL
	if GetDatabaseDriver() == DriverPostgres {
		schemaName := fmt.Sprintf("tenant_%s", tenantID)
		// Use Session() to prevent query condition accumulation
		if err := db.Session(&gorm.Session{}).Exec("CREATE SCHEMA IF NOT EXISTS " + quoteIdentifier(schemaName)).Error; err != nil {
			return fmt.Errorf("failed to create tenant schema: %w", err)
		}
		if err := db.Session(&gorm.Session{}).Exec("SET search_path TO " + quoteIdentifier(schemaName) + ", public").Error; err != nil {
			return fmt.Errorf("failed to set search_path: %w", err)
		}
	}

	// Required models — migration failure aborts startup
	requiredModels := []interface{}{
		// Auth (needed for multi-tenant login and refresh token)
		&models.User{},
		&models.RefreshSession{},

		// OAuth
		&models.OAuthState{},
		&models.OAuthLog{},

		// Platform Config
		&models.PlatformConfig{},

		// Orders
		&models.OrderTodayItem{},
		&models.LockedOrder{},

		// Jobs
		&models.Job{},
		&models.JobHistory{},
		&models.AutoFunctionConfig{},
		&models.AutoFunctionHistory{},
		&models.RouteExecutionConfig{},

		// Notifications (Facebook-style persistent) + per-user reads (V2)
		&models.Notification{},
		&models.NotificationRead{},
		&models.NotificationSettings{},

		// Credentials (canonical tenant-scoped credential storage)
		&models.CredentialConnection{},
		&models.CredentialAppConfig{},
		&models.OAuthConnectionAttempt{},
		&models.CredentialAuditEvent{},
	}

	// Optional models — migration failure logs warning and continues
	optionalModels := []interface{}{
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

		// Shopee Bookings
		&models.ShopeeBooking{},
		&models.ShopeeBookingItem{},

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
		&models.MarketplaceSyncHistory{},
		&models.SheetSnapshot{},
		&models.InventorySkuPlatformStatus{},

		// Settings
		&models.GoogleSheetsSettings{},
		&models.FilterPreference{},
		&models.Spreadsheet{},
		&models.RouteConfig{},
		&models.WholesaleSettings{},
		&models.GeneralSettings{},

		// Unified Products (Master Product System)
		&models.MasterProduct{},
		&models.MasterProductSku{},
		&models.MasterProductPlatformLink{},

		// Legacy Products (deprecated, kept for backward compat)
		&models.Product{},
		&models.ProductSKU{},

		// Image Gallery
		&models.Image{},

		// Product-Image Join Tables
		&models.MasterProductImage{},
		&models.ShopeeProductImage{},
		&models.TiktokProductImage{},
		&models.LazadaProductImage{},

		// Extensions (browser automation — paired Chrome installs, pairing
		// codes, and Shopee scrape results). Tenant-scoped: these live inside
		// tenant_{id} and carry no tenant_id column (schema isolation).
		&models.Extension{},
		&models.PairingCode{},
		&models.ScrapedProduct{},
	}

	// Migrate required models — failure aborts startup
	for _, model := range requiredModels {
		if err := dropLegacyUniqueConstraints(db, model); err != nil {
			log.Info().Msgf("  ⚠️  Warning normalizing unique constraints for %T: %v", model, err)
		}

		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate required model %T: %w", model, err)
		}
		log.Info().Msgf("  ✅ Migrated: %T", model)
	}

	// Migrate optional models — failure logs warning and continues
	for _, model := range optionalModels {
		if err := dropLegacyUniqueConstraints(db, model); err != nil {
			log.Info().Msgf("  ⚠️  Warning normalizing unique constraints for %T: %v", model, err)
		}

		if err := db.AutoMigrate(model); err != nil {
			log.Warn().Err(err).Msgf("Failed to migrate optional model %T, skipping", model)
			continue
		}
		log.Info().Msgf("  ✅ Migrated: %T", model)
	}

	// Post-migration: Ensure expression-based unique index for platform links.
	// GORM AutoMigrate cannot create this index (it uses COALESCE + partial WHERE).
	// This is idempotent — IF NOT EXISTS prevents re-creation.
	if GetDatabaseDriver() == DriverPostgres {
		linkTable := (&models.MasterProductPlatformLink{}).TableName()
		idxSQL := fmt.Sprintf(
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_links_unique
			 ON %s (platform, platform_product_id, COALESCE(platform_sku_id, ''))
			 WHERE (platform_product_id IS NOT NULL)`,
			linkTable,
		)
		if err := db.Session(&gorm.Session{}).Exec(idxSQL).Error; err != nil {
			log.Info().Msgf("  ⚠️  Warning creating platform links unique index: %v", err)
		} else {
			log.Info().Msg("  ✅ Ensured idx_platform_links_unique index")
		}

		// Credential connections: partial unique index for active (non-disabled) rows
		connTable := (&models.CredentialConnection{}).TableName()
		activeIdx := fmt.Sprintf(
			`CREATE UNIQUE INDEX IF NOT EXISTS idx_credential_connections_active
			 ON %s (tenant_id, platform, store_identifier)
			 WHERE disabled_at IS NULL`,
			connTable,
		)
		if err := db.Session(&gorm.Session{}).Exec(activeIdx).Error; err != nil {
			log.Info().Msgf("  ⚠️  Warning creating credential active unique index: %v", err)
		} else {
			log.Info().Msg("  ✅ Ensured idx_credential_connections_active index")
		}

		// Notifications V2: partial unique index on dedup_key so UpsertDedup
		// can UPSERT race-safely. Predicate is intentionally time-free
		// (Postgres forbids now() in index predicates); the "active window"
		// scoping is enforced at the app layer via ExpireStaleDedupKeys.
		notifDedupIdx := `CREATE UNIQUE INDEX IF NOT EXISTS ix_notif_dedup_active
			 ON notifications (dedup_key)
			 WHERE dedup_key IS NOT NULL`
		if err := db.Session(&gorm.Session{}).Exec(notifDedupIdx).Error; err != nil {
			log.Info().Msgf("  ⚠️  Warning creating notifications dedup index: %v", err)
		} else {
			log.Info().Msg("  ✅ Ensured ix_notif_dedup_active index")
		}
	}

	// NOTE: Zombie column cleanup intentionally RETIRED — see MigrateSystemDatabase comments.
	// GORM AutoMigrate is additive-only; removed model fields are retained in the database.

	log.Info().Msgf("✅ Tenant database migrations completed for: %s", tenantID)
	return nil
}

type uniqueConstraintInfo struct {
	ConstraintName string `gorm:"column:constraint_name"`
	ColumnName     string `gorm:"column:column_name"`
}

func dropLegacyUniqueConstraints(db *gorm.DB, model interface{}) error {
	if GetDatabaseDriver() != DriverPostgres {
		return nil
	}

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return fmt.Errorf("failed to parse model schema %T: %w", model, err)
	}

	tableName := stmt.Schema.Table
	if tableName == "" {
		return nil
	}

	var constraints []uniqueConstraintInfo
	err := db.Raw(`
		SELECT tc.constraint_name, kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.table_schema = kcu.table_schema
		 AND tc.table_name = kcu.table_name
		 AND tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = current_schema()
		  AND tc.table_name = ?
		  AND tc.constraint_type = 'UNIQUE'
		  AND (tc.constraint_name LIKE '%_key' OR tc.constraint_name LIKE 'uni_%')
		  AND (
			SELECT COUNT(*)
			FROM information_schema.key_column_usage k2
			WHERE k2.table_schema = tc.table_schema
			  AND k2.table_name = tc.table_name
			  AND k2.constraint_name = tc.constraint_name
		  ) = 1
	`, tableName).Scan(&constraints).Error
	if err != nil {
		return fmt.Errorf("failed to inspect unique constraints for %s: %w", tableName, err)
	}

	for _, constraint := range constraints {
		field := stmt.Schema.LookUpField(constraint.ColumnName)
		if field != nil && hasStandaloneUniqueTag(strings.ToLower(field.Tag.Get("gorm"))) {
			continue
		}

		dropSQL := fmt.Sprintf(
			"ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s",
			quoteIdentifier(tableName),
			quoteIdentifier(constraint.ConstraintName),
		)
		if err := db.Session(&gorm.Session{}).Exec(dropSQL).Error; err != nil {
			return fmt.Errorf("failed to drop legacy constraint %s on %s: %w", constraint.ConstraintName, tableName, err)
		}
	}

	return nil
}

func hasStandaloneUniqueTag(gormTag string) bool {
	if gormTag == "" {
		return false
	}

	for _, token := range strings.Split(gormTag, ";") {
		part := strings.TrimSpace(strings.ToLower(token))
		if part == "" || strings.HasPrefix(part, "uniqueindex") {
			continue
		}

		if part == "unique" || strings.HasPrefix(part, "unique:") {
			return true
		}
	}

	return false
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
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
			log.Info().Msgf("⚠️  Warning: failed to get tenant DB %s: %v", tenantID, err)
			continue
		}

		if err := MigrateTenantDatabase(tenantDB, tenantID); err != nil {
			log.Info().Msgf("⚠️  Warning: failed to migrate tenant DB %s: %v", tenantID, err)
			continue
		}
	}

	return nil
}
