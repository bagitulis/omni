package app

import (
	"context"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// runStartupSync runs a product sync for all tenants after server startup.
// Waits 30 seconds to allow other services to initialize, then triggers
// sync_products_inventory for each tenant that hasn't synced recently.
// This ensures data is fresh after server downtime (server not 24/7).
func runStartupSync(ctx context.Context, systemDB *gorm.DB, _ interface{}, basePath string) {
	// Wait for services to fully initialize (cancellable)
	select {
	case <-time.After(30 * time.Second):
	case <-ctx.Done():
		log.Info().Msg("[StartupSync] Cancelled during initialization wait")
		return
	}

	// Clear stale is_running flags from previous crash
	clearStaleRunningFlags(systemDB, basePath)
	log.Info().Msg("🚀 [StartupSync] Starting post-boot product sync for all tenants...")

	// Get all tenant IDs
	var schemas []string
	err := systemDB.Raw(`
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'tenant_%'
		ORDER BY schema_name
	`).Scan(&schemas).Error
	if err != nil {
		log.Error().Err(err).Msg("[StartupSync] Failed to get tenant schemas")
		return
	}

	synced := 0
	skipped := 0

	for _, schema := range schemas {
		// Check for shutdown before processing next tenant
		select {
		case <-ctx.Done():
			log.Info().Msg("[StartupSync] Shutdown signal received, stopping sync")
			return
		default:
		}
		if len(schema) <= 7 {
			continue
		}
		tenantID := schema[7:] // Remove "tenant_" prefix

		// Check if sync ran recently (within last 60 minutes)
		tenantDB, err := config.GetTenantDBWithContext(tenantID, basePath)
		if err != nil {
			log.Warn().Str("tenant_id", tenantID).Err(err).Msg("[StartupSync] Failed to get tenant DB")
			continue
		}

		if wasSyncedRecently(tenantDB, tenantID, 60*time.Minute) {
			skipped++
			continue
		}

		// Trigger sync for this tenant
		syncCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		cfg := &models.AutoFunctionConfig{
			Name:            "sync_products_inventory",
			IntervalMinutes: 60,
		}

		result, syncErr := syncProductsHandler(syncCtx, tenantID, cfg)
		cancel()

		if syncErr != nil {
			log.Warn().Str("tenant_id", tenantID).Err(syncErr).Msg("[StartupSync] Sync failed")
		} else {
			log.Info().Str("tenant_id", tenantID).Str("result", result).Msg("[StartupSync] Sync completed")
			synced++
		}

		// Small delay between tenants to avoid overwhelming platform APIs (cancellable)
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			log.Info().Msg("[StartupSync] Shutdown signal received during inter-tenant delay")
			return
		}
	}

	log.Info().
		Int("synced", synced).
		Int("skipped", skipped).
		Int("total_tenants", len(schemas)).
		Msg("🏁 [StartupSync] Post-boot sync completed")
}

// wasSyncedRecently checks if sync_products_inventory ran within the given duration.
func wasSyncedRecently(db *gorm.DB, tenantID string, within time.Duration) bool {
	var lastExec time.Time
	err := db.Model(&models.AutoFunctionConfig{}).
		Where("name = ? AND last_executed IS NOT NULL", "sync_products_inventory").
		Select("last_executed").
		Scan(&lastExec).Error

	if err != nil || lastExec.IsZero() {
		return false
	}

	return time.Since(lastExec) < within
}

// clearStaleRunningFlags resets is_running flags that were left stale from a server crash.
// If server died while an auto-function was running, is_running=true would be stuck forever.
// We clear the flag but KEEP progress_data so the next execution can resume.
func clearStaleRunningFlags(systemDB *gorm.DB, basePath string) {
	var schemas []string
	err := systemDB.Raw(`
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name LIKE 'tenant_%'
	`).Scan(&schemas).Error
	if err != nil {
		log.Warn().Err(err).Msg("[StartupSync] Failed to get schemas for stale flag cleanup")
		return
	}

	cleaned := 0
	for _, schema := range schemas {
		if len(schema) <= 7 {
			continue
		}
		tenantID := schema[7:]
		tenantDB, err := config.GetTenantDBWithContext(tenantID, basePath)
		if err != nil {
			continue
		}

		// Clear is_running but keep progress_data for resume
		result := tenantDB.Model(&models.AutoFunctionConfig{}).
			Where("is_running = ?", true).
			Updates(map[string]interface{}{
				"is_running":    false,
				"run_started_at": nil,
			})
		if result.RowsAffected > 0 {
			cleaned += int(result.RowsAffected)
			log.Info().Str("tenant_id", tenantID).Int64("count", result.RowsAffected).
				Msg("[StartupSync] Cleared stale is_running flags")
		}
	}

	if cleaned > 0 {
		log.Info().Int("total_cleaned", cleaned).Msg("[StartupSync] Stale running flags cleared")
	}
}
