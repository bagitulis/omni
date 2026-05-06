package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var ErrInventoryStockSKUNotFound = errors.New("SKU not found in inventory")

// StockBatchInventoryItem represents batch stock payload sourced from inventory records.
type StockBatchInventoryItem struct {
	SKU       string
	Stock     *int
	Platform  string
	Platforms []string
}

// UpdateStockFromInventory reads inventory stock, optionally updates local stock, and syncs to platforms.
func (o *StockUpdateOrchestrator) UpdateStockFromInventory(
	ctx context.Context,
	sku string,
	requestedStock *int,
	platforms []string,
) (*StockUpdateOrchestratorResult, int, error) {
	var record models.InventoryRecord
	err := o.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", o.tenantID, sku).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, ErrInventoryStockSKUNotFound
		}
		return nil, 0, err
	}

	stockValue, err := applyRequestedStock(&record, requestedStock)
	if err != nil {
		return nil, 0, err
	}

	if requestedStock != nil {
		tx := o.db.WithContext(ctx).Begin()
		if tx.Error != nil {
			return nil, 0, fmt.Errorf("failed to begin transaction: %w", tx.Error)
		}
		if err := tx.Save(&record).Error; err != nil {
			tx.Rollback()
			return nil, 0, err
		}
		tx.Commit()
	}

	result, err := o.UpdateStock(ctx, sku, stockValue, platforms)
	if err != nil {
		return nil, 0, err
	}

	recordInventoryStockSyncHistory(ctx, o.db, o.tenantID, sku, stockValue, result)
	return result, stockValue, nil
}

// UpdateStockBatchFromInventory processes batch stock updates from inventory records and returns per-item results.
func (o *StockUpdateOrchestrator) UpdateStockBatchFromInventory(
	ctx context.Context,
	items []StockBatchInventoryItem,
	requestPlatforms []string,
	requestPlatform string,
) []interface{} {
	if len(items) == 0 {
		return nil
	}

	// Build work items
	workItems := make([]BatchWorkItem, len(items))
	for i, item := range items {
		platforms := resolveStockPlatforms(item.Platforms, item.Platform, requestPlatforms, requestPlatform)
		workItems[i] = BatchWorkItem{
			Index:     i,
			SKU:       item.SKU,
			Platforms: platforms,
		}
	}

	// Worker function for each item
	worker := func(workerCtx context.Context, workItem BatchWorkItem) interface{} {
		item := items[workItem.Index]
		result, _, err := o.UpdateStockFromInventory(workerCtx, item.SKU, item.Stock, workItem.Platforms)
		if err != nil {
			// Fallback: if no inventory record but stock value is provided
			if errors.Is(err, ErrInventoryStockSKUNotFound) && item.Stock != nil {
				directResult, directErr := o.UpdateStock(workerCtx, item.SKU, *item.Stock, workItem.Platforms)
				if directErr != nil {
					return map[string]interface{}{
						"sku":     item.SKU,
						"success": false,
						"error":   directErr.Error(),
					}
				}
				return directResult
			}

			errMessage := err.Error()
			if errors.Is(err, ErrInventoryStockSKUNotFound) {
				errMessage = "SKU not found in inventory"
			}

			return map[string]interface{}{
				"sku":     item.SKU,
				"success": false,
				"error":   errMessage,
			}
		}

		return result
	}

	// Execute in parallel
	config := DefaultParallelBatchConfig()
	return RunParallelBatch(ctx, workItems, config, worker)
}

func detectStockField(data map[string]interface{}) string {
	for _, key := range []string{"Stock", "stock", "STOCK", "Quantity", "quantity", "QTY", "qty", "Stok", "stok"} {
		if _, exists := data[key]; exists {
			return key
		}
	}

	return "Stock"
}

func applyRequestedStock(record *models.InventoryRecord, requestedStock *int) (int, error) {
	currentStock := GetQuantity(*record)
	if requestedStock == nil {
		return currentStock, nil
	}
	if *requestedStock < 0 {
		return 0, fmt.Errorf("stock must be non-negative")
	}

	data := GetDataMap(*record)
	stockField := detectStockField(data)
	if err := SetDataValue(record, stockField, *requestedStock); err != nil {
		return 0, fmt.Errorf("failed to apply stock to inventory record: %w", err)
	}

	return *requestedStock, nil
}

func resolveStockPlatforms(itemPlatforms []string, itemPlatform string, reqPlatforms []string, reqPlatform string) []string {
	if len(itemPlatforms) > 0 {
		return itemPlatforms
	}
	if itemPlatform != "" {
		return []string{itemPlatform}
	}
	if len(reqPlatforms) > 0 {
		return reqPlatforms
	}
	if reqPlatform != "" {
		return []string{reqPlatform}
	}

	return nil
}

func toPtr(s string) *string {
	return &s
}

func recordInventoryStockSyncHistory(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
	sku string,
	stock int,
	result *StockUpdateOrchestratorResult,
) {
	if result == nil {
		return
	}

	repo := repositories.NewMarketplaceSyncHistoryRepo(db)

	for platform, platformResult := range result.Platforms {
		status := "failed"
		if platformResult.Success {
			status = "success"
		}

		requestDataBytes, _ := json.Marshal(map[string]interface{}{
			"sku":      sku,
			"platform": platform,
			"stock":    stock,
		})
		responseDataBytes, _ := json.Marshal(platformResult)

		var errorMessage *string
		if platformResult.Error != "" {
			errorMessage = toPtr(platformResult.Error)
		}

		entry := &models.MarketplaceSyncHistory{
			TenantID:     tenantID,
			SKU:          sku,
			Platform:     platform,
			Operation:    "stock_update",
			Status:       status,
			RequestData:  toPtr(string(requestDataBytes)),
			ResponseData: toPtr(string(responseDataBytes)),
			ErrorMessage: errorMessage,
		}

		if err := repo.Create(ctx, entry); err != nil {
			log.Error().
				Err(err).
				Str("tenant_id", tenantID).
				Str("sku", sku).
				Str("platform", platform).
				Msg("Failed to record stock sync history")
		}
	}
}
