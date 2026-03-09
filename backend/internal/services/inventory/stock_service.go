package inventory

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// StockService handles stock update operations
type StockService struct {
	db       *gorm.DB
	tenantID string
}

// NewStockService creates a new stock service
func NewStockService(db *gorm.DB, tenantID string) *StockService {
	return &StockService{db: db, tenantID: tenantID}
}

// StockUpdateItem represents a single stock update request
type StockUpdateItem struct {
	SKU       string   `json:"sku"`
	Quantity  int      `json:"quantity"`
	Stock     *int     `json:"stock,omitempty"`
	Platform  string   `json:"platform"`
	Platforms []string `json:"platforms,omitempty"`
}

// NormalizeStockUpdateItems expands mixed payload formats into service-ready items.
// Supported input formats:
// - {sku, quantity, platform}
// - {sku, stock, platforms:[...]}
func NormalizeStockUpdateItems(items []StockUpdateItem) []StockUpdateItem {
	normalized := make([]StockUpdateItem, 0, len(items))

	for _, item := range items {
		quantity := item.Quantity
		if item.Stock != nil {
			quantity = *item.Stock
		}

		if len(item.Platforms) > 0 {
			for _, platform := range item.Platforms {
				normalized = append(normalized, StockUpdateItem{
					SKU:      item.SKU,
					Quantity: quantity,
					Platform: platform,
				})
			}
			continue
		}

		normalized = append(normalized, StockUpdateItem{
			SKU:      item.SKU,
			Quantity: quantity,
			Platform: item.Platform,
		})
	}

	return normalized
}

// StockUpdateResult represents the result of a stock update
type StockUpdateResult struct {
	SKU         string `json:"sku"`
	OldQuantity int    `json:"old_quantity"`
	NewQuantity int    `json:"new_quantity"`
	Platform    string `json:"platform,omitempty"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// BatchUpdateResult represents the result of batch stock update
type BatchUpdateResult struct {
	Total   int                 `json:"total"`
	Success int                 `json:"success"`
	Failed  int                 `json:"failed"`
	Results []StockUpdateResult `json:"results"`
}

// PlatformIdLookup represents a platform ID lookup result
type PlatformIdLookup struct {
	SKU       string `json:"sku"`
	ProductID string `json:"product_id,omitempty"`
	ItemID    string `json:"item_id,omitempty"`
	ModelID   string `json:"model_id,omitempty"`
	Found     bool   `json:"found"`
}

// UpdateStock updates stock for a single SKU using JSONB data
func (s *StockService) UpdateStock(ctx context.Context, sku string, quantity int, platform string) (*StockUpdateResult, error) {
	var record models.InventoryRecord

	// Find by key_value (SKU)
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		First(&record).Error

	if err == gorm.ErrRecordNotFound {
		return &StockUpdateResult{
			SKU:     sku,
			Success: false,
			Error:   "SKU not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// Get old quantity from JSONB data using helper
	oldQty := GetQuantity(record)

	// Update quantity in JSONB data - try common stock column names
	data := GetDataMap(record)
	stockUpdated := false
	for _, key := range []string{"Stock", "stock", "STOCK", "Quantity", "quantity", "QTY", "qty", "Stok", "stok"} {
		if _, exists := data[key]; exists {
			SetDataValue(&record, key, quantity)
			stockUpdated = true
			break
		}
	}
	if !stockUpdated {
		// Default to Stock if no stock column found
		SetDataValue(&record, "Stock", quantity)
	}

	if err := s.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, err
	}

	// Note: Platform sync is handled by StockUpdateOrchestrator (stock_orchestrator.go),
	// not at this service level. This method only updates the local inventory record.

	return &StockUpdateResult{
		SKU:         sku,
		OldQuantity: oldQty,
		NewQuantity: quantity,
		Platform:    platform,
		Success:     true,
	}, nil
}

// UpdateStockBatch updates stock for multiple SKUs
func (s *StockService) UpdateStockBatch(ctx context.Context, items []StockUpdateItem) (*BatchUpdateResult, error) {
	result := &BatchUpdateResult{
		Total:   len(items),
		Results: make([]StockUpdateResult, 0, len(items)),
	}

	for _, item := range items {
		updateResult, err := s.UpdateStock(ctx, item.SKU, item.Quantity, item.Platform)
		if err != nil {
			result.Results = append(result.Results, StockUpdateResult{
				SKU:     item.SKU,
				Success: false,
				Error:   err.Error(),
			})
			result.Failed++
			continue
		}

		result.Results = append(result.Results, *updateResult)
		if updateResult.Success {
			result.Success++
		} else {
			result.Failed++
		}
	}

	return result, nil
}

// LookupPlatformIds looks up platform IDs for given SKUs
func (s *StockService) LookupPlatformIds(ctx context.Context, skus []string, platform string) ([]PlatformIdLookup, error) {
	results := make([]PlatformIdLookup, 0, len(skus))

	for _, sku := range skus {
		lookup := PlatformIdLookup{SKU: sku, Found: false}

		switch platform {
		case "shopee":
			var record models.InventorySkuPlatformStatus
			err := s.db.WithContext(ctx).
				Where("tenant_id = ? AND sku = ? AND platform = ?", s.tenantID, sku, "shopee").
				First(&record).Error
			if err == nil && record.PlatformProductID != "" {
				lookup.ProductID = record.PlatformProductID
				lookup.ItemID = fmt.Sprintf("%v", record.PlatformItemID)
				lookup.Found = true
			}

		case "lazada":
			var record models.InventorySkuPlatformStatus
			err := s.db.WithContext(ctx).
				Where("tenant_id = ? AND sku = ? AND platform = ?", s.tenantID, sku, "lazada").
				First(&record).Error
			if err == nil && record.PlatformProductID != "" {
				lookup.ProductID = record.PlatformProductID
				lookup.ItemID = fmt.Sprintf("%v", record.PlatformItemID)
				lookup.Found = true
			}

		case "tiktok":
			var record models.InventorySkuPlatformStatus
			err := s.db.WithContext(ctx).
				Where("tenant_id = ? AND sku = ? AND platform = ?", s.tenantID, sku, "tiktok").
				First(&record).Error
			if err == nil && record.PlatformProductID != "" {
				lookup.ProductID = record.PlatformProductID
				lookup.Found = true
			}
		}

		results = append(results, lookup)
	}

	return results, nil
}
