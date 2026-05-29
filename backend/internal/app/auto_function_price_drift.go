package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
)

// priceDriftDetectionHandler detects SKUs with price drift (inventory updated after last successful sync)
func priceDriftDetectionHandler(ctx context.Context, tenantID string, cfg *models.AutoFunctionConfig) (string, error) {
	log.Info().Msgf("[AutoFunction] Running 'Price Drift Detection' for tenant: %s", tenantID)

	basePath := os.Getenv("UPLOAD_PATH")
	if basePath == "" {
		basePath = "./data"
	}

	db, err := config.GetTenantDBWithContext(tenantID, basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant DB: %w", err)
	}

	// Query: Find inventory records with their latest successful price_update sync
	// Using raw SQL for efficiency (LEFT JOIN LATERAL for per-SKU latest sync)
	type QueryResult struct {
		SKU                string     `gorm:"column:sku"`
		InventoryUpdatedAt time.Time  `gorm:"column:inventory_updated_at"`
		LastSyncAt         *time.Time `gorm:"column:last_sync_at"`
		LastSyncPlatform   *string    `gorm:"column:last_sync_platform"`
	}

	var results []QueryResult
	query := `
		SELECT 
			ir.key_value as sku,
			ir.updated_at as inventory_updated_at,
			msh.created_at as last_sync_at,
			msh.platform as last_sync_platform
		FROM inventory_records ir
		LEFT JOIN LATERAL (
			SELECT created_at, platform
			FROM marketplace_sync_histories
			WHERE sku = ir.key_value 
				AND tenant_id = ir.tenant_id
				AND operation = 'price_update' 
				AND status = 'success'
			ORDER BY created_at DESC 
			LIMIT 1
		) msh ON true
		WHERE ir.tenant_id = ?
			AND (msh.created_at IS NULL OR ir.updated_at > msh.created_at)
	`

	if err := db.WithContext(ctx).Raw(query, tenantID).Scan(&results).Error; err != nil {
		return "", fmt.Errorf("failed to query price drift: %w", err)
	}

	totalChecked := len(results)
	if totalChecked == 0 {
		return "No price drift detected (checked 0 SKUs)", nil
	}

	// Group drifted SKUs by platform
	driftedSKUs := make([]string, 0, len(results))
	for _, result := range results {
		if result.LastSyncPlatform != nil {
			driftedSKUs = append(driftedSKUs, fmt.Sprintf("%s (%s)", result.SKU, *result.LastSyncPlatform))
		} else {
			driftedSKUs = append(driftedSKUs, fmt.Sprintf("%s (never synced)", result.SKU))
		}
	}

	// Limit SKU list in message to first 10
	skuList := strings.Join(driftedSKUs[:min(10, len(driftedSKUs))], ", ")
	if len(driftedSKUs) > 10 {
		skuList += "..."
	}

	resultMsg := fmt.Sprintf("Found %d SKUs with price drift: [%s] (checked %d total)", len(driftedSKUs), skuList, totalChecked)
	log.Info().Msgf("[AutoFunction] %s for tenant: %s", resultMsg, tenantID)
	return resultMsg, nil
}
