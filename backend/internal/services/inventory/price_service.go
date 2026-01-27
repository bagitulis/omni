package inventory

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// PriceService handles price update operations
type PriceService struct {
	db       *gorm.DB
	tenantID string
}

// NewPriceService creates a new price service
func NewPriceService(db *gorm.DB, tenantID string) *PriceService {
	return &PriceService{db: db, tenantID: tenantID}
}

// PriceUpdateItem represents a single price update request
// Matches Node.js backend format: { sku, price, platforms?: string[] }
type PriceUpdateItem struct {
	SKU       string   `json:"sku" binding:"required"`
	Price     float64  `json:"price" binding:"required"`
	Platforms []string `json:"platforms,omitempty"` // Optional: array of platforms
}

// PriceUpdateResult represents the result of a price update
type PriceUpdateResult struct {
	SKU      string  `json:"sku"`
	OldPrice float64 `json:"old_price"`
	NewPrice float64 `json:"new_price"`
	Platform string  `json:"platform,omitempty"`
	Success  bool    `json:"success"`
	Error    string  `json:"error,omitempty"`
}

// BatchPriceResult represents the result of batch price update
type BatchPriceResult struct {
	Total   int                 `json:"total"`
	Success int                 `json:"success"`
	Failed  int                 `json:"failed"`
	Results []PriceUpdateResult `json:"results"`
}

// UpdatePrice updates price for a single SKU using JSONB data
func (s *PriceService) UpdatePrice(ctx context.Context, sku string, price float64, platform string) (*PriceUpdateResult, error) {
	var record models.InventoryRecord

	// Find by key_value (SKU) - supports flexible key column
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		First(&record).Error

	if err == gorm.ErrRecordNotFound {
		return &PriceUpdateResult{
			SKU:     sku,
			Success: false,
			Error:   "SKU not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// Get old price from JSONB data using helper
	oldPrice := GetPrice(record)

	// Update price in JSONB data - try common price column names
	data := GetDataMap(record)
	priceUpdated := false
	for _, key := range []string{"HARGA", "Harga", "harga", "Price", "price", "PRICE"} {
		if _, exists := data[key]; exists {
			SetDataValue(&record, key, price)
			priceUpdated = true
			break
		}
	}
	if !priceUpdated {
		// Default to HARGA if no price column found
		SetDataValue(&record, "HARGA", price)
	}

	if err := s.db.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, err
	}

	// TODO: If platform is specified, sync to that platform via API

	return &PriceUpdateResult{
		SKU:      sku,
		OldPrice: oldPrice,
		NewPrice: price,
		Platform: platform,
		Success:  true,
	}, nil
}

// UpdatePriceBatch updates prices for multiple SKUs
func (s *PriceService) UpdatePriceBatch(ctx context.Context, items []PriceUpdateItem) (*BatchPriceResult, error) {
	result := &BatchPriceResult{
		Total:   len(items),
		Results: make([]PriceUpdateResult, 0, len(items)),
	}

	for _, item := range items {
		// Use first platform if specified, otherwise empty string (all platforms)
		platform := ""
		if len(item.Platforms) > 0 {
			platform = item.Platforms[0]
		}

		updateResult, err := s.UpdatePrice(ctx, item.SKU, item.Price, platform)
		if err != nil {
			result.Results = append(result.Results, PriceUpdateResult{
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
