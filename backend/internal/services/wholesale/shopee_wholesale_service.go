package wholesale

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// WholesaleTier represents a single wholesale tier
type WholesaleTier struct {
	MinCount  int     `json:"min_count"`
	MaxCount  int     `json:"max_count"`
	UnitPrice float64 `json:"unit_price"`
}

// ShopeeWholesaleService handles Shopee wholesale operations
type ShopeeWholesaleService struct {
	db       *gorm.DB
	tenantID string
}

// NewShopeeWholesaleService creates a new Shopee wholesale service
func NewShopeeWholesaleService(db *gorm.DB, tenantID string) *ShopeeWholesaleService {
	return &ShopeeWholesaleService{
		db:       db,
		tenantID: tenantID,
	}
}

// DeleteWholesaleTiers deletes wholesale tiers for an item
func (s *ShopeeWholesaleService) DeleteWholesaleTiers(ctx context.Context, itemID int64) error {
	// TODO: Call Shopee API to delete wholesale tiers
	// This requires ShopeeAPIClient integration
	return fmt.Errorf("not implemented: Shopee API integration required")
}

// UpdateWholesaleTiers updates wholesale tiers for an item
func (s *ShopeeWholesaleService) UpdateWholesaleTiers(ctx context.Context, itemID int64, tiers []WholesaleTier) error {
	// TODO: Call Shopee API to update wholesale tiers
	// This requires ShopeeAPIClient integration
	return fmt.Errorf("not implemented: Shopee API integration required")
}

// GetWholesaleTiers gets wholesale tiers for an item
func (s *ShopeeWholesaleService) GetWholesaleTiers(ctx context.Context, itemID int64) ([]WholesaleTier, error) {
	// TODO: Call Shopee API to get wholesale tiers
	// This requires ShopeeAPIClient integration
	return nil, fmt.Errorf("not implemented: Shopee API integration required")
}

// BatchDeleteBySkus deletes wholesale for multiple SKUs (with deduplication)
type BatchDeleteResult struct {
	TotalSKUs   int                     `json:"total_skus"`
	UniqueItems int                     `json:"unique_items"`
	Processed   int                     `json:"processed"`
	Failed      int                     `json:"failed"`
	Skipped     int                     `json:"skipped"`
	Success     bool                    `json:"success"`
	Results     []SingleWholesaleResult `json:"results"`
}

type SingleWholesaleResult struct {
	ItemID  int64  `json:"item_id,omitempty"`
	SKU     string `json:"sku,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

func (s *ShopeeWholesaleService) BatchDeleteBySkus(ctx context.Context, skus []string) (*BatchDeleteResult, error) {
	result := &BatchDeleteResult{
		TotalSKUs: len(skus),
		Results:   make([]SingleWholesaleResult, 0),
	}

	// Map SKUs to item_ids (deduplicate)
	itemMap := make(map[int64][]string) // item_id -> []sku

	for _, sku := range skus {
		var product models.ShopeeProduct
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
			First(&product).Error

		if err == gorm.ErrRecordNotFound {
			result.Skipped++
			result.Results = append(result.Results, SingleWholesaleResult{
				SKU:     sku,
				Success: false,
				Error:   "SKU not found",
			})
			continue
		}
		if err != nil {
			result.Failed++
			result.Results = append(result.Results, SingleWholesaleResult{
				SKU:     sku,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		if product.ItemID != 0 {
			itemMap[product.ItemID] = append(itemMap[product.ItemID], sku)
		}
	}

	result.UniqueItems = len(itemMap)

	// Delete wholesale for each unique item
	for itemID, itemSKUs := range itemMap {
		err := s.DeleteWholesaleTiers(ctx, itemID)
		if err != nil {
			result.Failed++
			result.Results = append(result.Results, SingleWholesaleResult{
				ItemID:  itemID,
				SKU:     itemSKUs[0], // Representative SKU
				Success: false,
				Error:   err.Error(),
			})
		} else {
			result.Processed++
			result.Results = append(result.Results, SingleWholesaleResult{
				ItemID:  itemID,
				SKU:     itemSKUs[0],
				Success: true,
				Message: fmt.Sprintf("Deleted for %d SKUs", len(itemSKUs)),
			})
		}
	}

	result.Success = result.Failed == 0
	return result, nil
}

// BatchUpdateBySkus updates wholesale for multiple SKUs
type BatchUpdateResult struct {
	TotalSKUs   int                     `json:"total_skus"`
	UniqueItems int                     `json:"unique_items"`
	Processed   int                     `json:"processed"`
	Failed      int                     `json:"failed"`
	Skipped     int                     `json:"skipped"`
	Success     bool                    `json:"success"`
	Results     []SingleWholesaleResult `json:"results"`
}

func (s *ShopeeWholesaleService) BatchUpdateBySkus(
	ctx context.Context,
	skuPriceMap map[string]float64,
	calculateTiers func(float64) []WholesaleTier,
) (*BatchUpdateResult, error) {
	result := &BatchUpdateResult{
		TotalSKUs: len(skuPriceMap),
		Results:   make([]SingleWholesaleResult, 0),
	}

	// Map SKUs to item_ids with prices
	itemMap := make(map[int64]struct {
		skus  []string
		price float64
	})

	for sku, price := range skuPriceMap {
		var product models.ShopeeProduct
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
			First(&product).Error

		if err == gorm.ErrRecordNotFound {
			result.Skipped++
			continue
		}
		if err != nil {
			result.Failed++
			continue
		}

		if product.ItemID != 0 {
			if existing, exists := itemMap[product.ItemID]; exists {
				existing.skus = append(existing.skus, sku)
				itemMap[product.ItemID] = existing
			} else {
				itemMap[product.ItemID] = struct {
					skus  []string
					price float64
				}{
					skus:  []string{sku},
					price: price,
				}
			}
		}
	}

	result.UniqueItems = len(itemMap)

	// Update wholesale for each unique item
	for itemID, data := range itemMap {
		tiers := calculateTiers(data.price)
		err := s.UpdateWholesaleTiers(ctx, itemID, tiers)
		if err != nil {
			result.Failed++
			result.Results = append(result.Results, SingleWholesaleResult{
				ItemID:  itemID,
				SKU:     data.skus[0],
				Success: false,
				Error:   err.Error(),
			})
		} else {
			result.Processed++
			result.Results = append(result.Results, SingleWholesaleResult{
				ItemID:  itemID,
				SKU:     data.skus[0],
				Success: true,
				Message: fmt.Sprintf("Updated for %d SKUs", len(data.skus)),
			})
		}
	}

	result.Success = result.Failed == 0
	return result, nil
}

// LookupItemIDBySKU finds item_id by SKU
func (s *ShopeeWholesaleService) LookupItemIDBySKU(ctx context.Context, sku string) (int64, error) {
	var product models.ShopeeProduct
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND sku = ?", s.tenantID, sku).
		First(&product).Error

	if err == gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("SKU not found: %s", sku)
	}
	if err != nil {
		return 0, err
	}

	return product.ItemID, nil
}
