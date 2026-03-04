package wholesale

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WholesaleTier is an alias for models.WholesaleTier to avoid cross-package duplication (DRY)
type WholesaleTier = models.WholesaleTier

// ShopeeWholesaleService handles Shopee wholesale operations
type ShopeeWholesaleService struct {
	db        *gorm.DB
	tenantID  string
	shopeeAPI *shopee.ProductAPI
}

// NewShopeeWholesaleService creates a new Shopee wholesale service
func NewShopeeWholesaleService(db *gorm.DB, tenantID string, shopeeAPI *shopee.ProductAPI) *ShopeeWholesaleService {
	return &ShopeeWholesaleService{
		db:        db,
		tenantID:  tenantID,
		shopeeAPI: shopeeAPI,
	}
}

// DeleteWholesaleTiers deletes wholesale tiers for an item
func (s *ShopeeWholesaleService) DeleteWholesaleTiers(ctx context.Context, itemID int64) error {
	if s.shopeeAPI == nil {
		return fmt.Errorf("Shopee API client not configured")
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Msg("Deleting wholesale tiers")

	// Call Shopee API to delete wholesale (empty array)
	err := s.shopeeAPI.UpdateItemWholesale(ctx, itemID, []shopee.WholesaleTier{})
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Msg("Failed to delete wholesale tiers")
		return err
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Msg("Wholesale tiers deleted successfully")

	return nil
}

// UpdateWholesaleTiers updates wholesale tiers for an item
func (s *ShopeeWholesaleService) UpdateWholesaleTiers(ctx context.Context, itemID int64, tiers []WholesaleTier) error {
	if s.shopeeAPI == nil {
		return fmt.Errorf("Shopee API client not configured")
	}

	// Convert to Shopee API format
	shopeeTiers := make([]shopee.WholesaleTier, len(tiers))
	for i, t := range tiers {
		shopeeTiers[i] = shopee.WholesaleTier{
			MinCount:  t.MinCount,
			MaxCount:  t.MaxCount,
			UnitPrice: t.UnitPrice,
		}
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Int("tier_count", len(tiers)).
		Msg("Updating wholesale tiers")

	err := s.shopeeAPI.UpdateItemWholesale(ctx, itemID, shopeeTiers)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Msg("Failed to update wholesale tiers")
		return err
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Int("tier_count", len(tiers)).
		Msg("Wholesale tiers updated successfully")

	return nil
}

// GetWholesaleTiers gets wholesale tiers for an item
func (s *ShopeeWholesaleService) GetWholesaleTiers(ctx context.Context, itemID int64) ([]WholesaleTier, error) {
	if s.shopeeAPI == nil {
		return nil, fmt.Errorf("Shopee API client not configured")
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Msg("Getting wholesale tiers")

	// Call Shopee API to get item info
	shopeeTiers, err := s.shopeeAPI.GetItemWholesale(ctx, itemID)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Msg("Failed to get wholesale tiers")
		return nil, err
	}

	// Convert from Shopee API format
	tiers := make([]WholesaleTier, len(shopeeTiers))
	for i, t := range shopeeTiers {
		tiers[i] = WholesaleTier{
			MinCount:  t.MinCount,
			MaxCount:  t.MaxCount,
			UnitPrice: t.UnitPrice,
		}
	}

	return tiers, nil
}

// BatchDeleteResult represents batch delete result
type BatchDeleteResult struct {
	TotalSKUs   int                     `json:"total_skus"`
	UniqueItems int                     `json:"unique_items"`
	Processed   int                     `json:"processed"`
	Failed      int                     `json:"failed"`
	Skipped     int                     `json:"skipped"`
	Success     bool                    `json:"success"`
	Results     []SingleWholesaleResult `json:"results"`
}

// SingleWholesaleResult represents single operation result
type SingleWholesaleResult struct {
	ItemID  int64  `json:"item_id,omitempty"`
	SKU     string `json:"sku,omitempty"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// BatchDeleteBySkus deletes wholesale for multiple SKUs (with deduplication)
func (s *ShopeeWholesaleService) BatchDeleteBySkus(ctx context.Context, skus []string) (*BatchDeleteResult, error) {
	result := &BatchDeleteResult{
		TotalSKUs: len(skus),
		Results:   make([]SingleWholesaleResult, 0),
	}

	// Map SKUs to item_ids (deduplicate)
	itemMap := make(map[int64][]string) // item_id -> []sku

	for _, sku := range skus {
		var product models.ShopeeSku
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND seller_sku = ?", s.tenantID, sku).
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

// BatchUpdateResult represents batch update result
type BatchUpdateResult struct {
	TotalSKUs   int                     `json:"total_skus"`
	UniqueItems int                     `json:"unique_items"`
	Processed   int                     `json:"processed"`
	Failed      int                     `json:"failed"`
	Skipped     int                     `json:"skipped"`
	Success     bool                    `json:"success"`
	Results     []SingleWholesaleResult `json:"results"`
}

// BatchUpdateBySkus updates wholesale for multiple SKUs
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
		var product models.ShopeeSku
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND seller_sku = ?", s.tenantID, sku).
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
	var product models.ShopeeSku
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND seller_sku = ?", s.tenantID, sku).
		First(&product).Error

	if err == gorm.ErrRecordNotFound {
		return 0, fmt.Errorf("SKU not found: %s", sku)
	}
	if err != nil {
		return 0, err
	}

	return product.ItemID, nil
}
