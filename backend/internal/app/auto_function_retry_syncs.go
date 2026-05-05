package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"os"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/inventory"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// retryFailedSyncsHandler retries failed marketplace sync operations from the last 24 hours
// It queries marketplace_sync_history for failed entries with retry_count < 3,
// groups by SKU+platform, and retries using the appropriate orchestrator.
func retryFailedSyncsHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Info().Msgf("[AutoFunction] Running 'Retry Failed Syncs' for tenant: %s", tenantID)

	// Get tenant database
	db, err := getTenantDB(tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant DB: %w", err)
	}

	// Query failed syncs from last 24 hours with retry_count < 3
	var failedSyncs []models.MarketplaceSyncHistory
	cutoffTime := time.Now().Add(-24 * time.Hour)

	if err := db.WithContext(ctx).
		Where("tenant_id = ? AND status = ? AND retry_count < ? AND created_at > ?",
			tenantID, "failed", 3, cutoffTime).
		Order("created_at DESC").
		Find(&failedSyncs).Error; err != nil {
		return "", fmt.Errorf("failed to query failed syncs: %w", err)
	}

	if len(failedSyncs) == 0 {
		return "No failed syncs to retry", nil
	}

	log.Info().Msgf("[AutoFunction] Found %d failed syncs to retry", len(failedSyncs))

	// Filter out credential-related errors (skip retry)
	var retryable []models.MarketplaceSyncHistory
	skippedCredErrors := 0

	for _, sync := range failedSyncs {
		if sync.ErrorMessage != nil {
			errMsg := strings.ToLower(*sync.ErrorMessage)
			if strings.Contains(errMsg, "credential") || strings.Contains(errMsg, "access_token") {
				log.Info().Msgf("[AutoFunction] Skipping retry for %s/%s (credential error)", sync.SKU, sync.Platform)
				skippedCredErrors++
				continue
			}
		}
		retryable = append(retryable, sync)
	}

	if len(retryable) == 0 {
		return fmt.Sprintf("No retryable syncs (all %d have credential errors)", skippedCredErrors), nil
	}

	log.Info().Msgf("[AutoFunction] Retrying %d syncs (skipped %d credential errors)", len(retryable), skippedCredErrors)

	// Initialize credential service with dbPath
	basePath := os.Getenv("DATA_PATH")
	if basePath == "" {
		basePath = "./data"
	}
	credService := services.NewCredentialService(basePath)

	// Group by SKU+platform+operation to avoid duplicate retries
	type retryKey struct {
		SKU       string
		Platform  string
		Operation string
	}
	retryMap := make(map[retryKey]models.MarketplaceSyncHistory)

	for _, sync := range retryable {
		key := retryKey{SKU: sync.SKU, Platform: sync.Platform, Operation: sync.Operation}
		// Keep the most recent entry for each SKU+platform+operation
		if existing, exists := retryMap[key]; !exists || sync.CreatedAt.After(existing.CreatedAt) {
			retryMap[key] = sync
		}
	}

	log.Info().Msgf("[AutoFunction] Grouped into %d unique retry operations", len(retryMap))

	// Retry each unique operation
	successCount := 0
	failCount := 0

	for _, sync := range retryMap {
		retryErr := retrySync(ctx, db, tenantID, credService, &sync)
		if retryErr != nil {
			log.Error().Err(retryErr).Msgf("[AutoFunction] Retry failed for %s/%s/%s", sync.SKU, sync.Platform, sync.Operation)
			failCount++
			// Update retry_count on failure
			updateRetryCount(ctx, db, sync.ID)
		} else {
			log.Info().Msgf("[AutoFunction] Retry succeeded for %s/%s/%s", sync.SKU, sync.Platform, sync.Operation)
			successCount++
		}
	}

	resultMsg := fmt.Sprintf("Retried %d items: %d success, %d failed (skipped %d credential errors)",
		len(retryMap), successCount, failCount, skippedCredErrors)
	log.Info().Msgf("[AutoFunction] %s for tenant: %s", resultMsg, tenantID)

	return resultMsg, nil
}

// retrySync retries a single sync operation using the appropriate orchestrator
func retrySync(ctx context.Context, db *gorm.DB, tenantID string, credService *services.CredentialService, sync *models.MarketplaceSyncHistory) error {
	// Parse request_data to extract parameters
	if sync.RequestData == nil {
		return fmt.Errorf("missing request_data")
	}

	var requestData map[string]interface{}
	if err := json.Unmarshal([]byte(*sync.RequestData), &requestData); err != nil {
		return fmt.Errorf("failed to parse request_data: %w", err)
	}

	switch sync.Operation {
	case "stock_update":
		return retryStockUpdate(ctx, db, tenantID, credService, sync, requestData)
	case "price_update":
		return retryPriceUpdate(ctx, db, tenantID, credService, sync, requestData)
	default:
		return fmt.Errorf("unsupported operation: %s", sync.Operation)
	}
}

// retryStockUpdate retries a stock update operation
func retryStockUpdate(ctx context.Context, db *gorm.DB, tenantID string, credService *services.CredentialService, sync *models.MarketplaceSyncHistory, requestData map[string]interface{}) error {
	sku := sync.SKU
	platform := sync.Platform

	// Extract stock value from request_data
	stockFloat, ok := requestData["stock"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid stock value in request_data")
	}
	stock := int(stockFloat)

	// Create orchestrator
	orchestrator := inventory.NewStockUpdateOrchestrator(db, tenantID, credService)

	// Retry stock update for single platform
	result, err := orchestrator.UpdateStock(ctx, sku, stock, []string{platform})
	if err != nil {
		return fmt.Errorf("stock update failed: %w", err)
	}

	// Check platform result
	platformResult, exists := result.Platforms[platform]
	if !exists || !platformResult.Success {
		errMsg := "unknown error"
		if exists && platformResult.Error != "" {
			errMsg = platformResult.Error
		}
		return fmt.Errorf("platform update failed: %s", errMsg)
	}

	return nil
}

// retryPriceUpdate retries a price update operation
func retryPriceUpdate(ctx context.Context, db *gorm.DB, tenantID string, credService *services.CredentialService, sync *models.MarketplaceSyncHistory, requestData map[string]interface{}) error {
	sku := sync.SKU
	platform := sync.Platform

	// Extract price value from request_data
	price, ok := requestData["price"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid price value in request_data")
	}

	// Create orchestrator
	orchestrator := inventory.NewPriceUpdateOrchestrator(db, tenantID, credService)

	// Retry price update for single platform
	result, err := orchestrator.UpdatePrice(ctx, sku, price, []string{platform})
	if err != nil {
		return fmt.Errorf("price update failed: %w", err)
	}

	// Check platform result
	platformResult, exists := result.Platforms[platform]
	if !exists || !platformResult.Success {
		errMsg := "unknown error"
		if exists && platformResult.Error != "" {
			errMsg = platformResult.Error
		}
		return fmt.Errorf("platform update failed: %s", errMsg)
	}

	return nil
}

// updateRetryCount increments the retry_count for a sync history entry
func updateRetryCount(ctx context.Context, db *gorm.DB, syncID string) {
	if err := db.WithContext(ctx).
		Model(&models.MarketplaceSyncHistory{}).
		Where("id = ?", syncID).
		UpdateColumn("retry_count", gorm.Expr("retry_count + 1")).Error; err != nil {
		log.Error().Err(err).Msgf("[AutoFunction] Failed to update retry_count for sync ID: %s", syncID)
	}
}

// getTenantDB is a helper to get tenant database (reuse pattern from other auto functions)
// getTenantDB is a helper to get tenant database
func getTenantDB(tenantID string) (*gorm.DB, error) {
	basePath := os.Getenv("DATA_PATH")
	if basePath == "" {
		basePath = "./data"
	}
	return config.GetTenantDBWithContext(tenantID, basePath)
}
