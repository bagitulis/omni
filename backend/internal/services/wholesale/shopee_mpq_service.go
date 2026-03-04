package wholesale

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ShopeeMpqService handles Shopee MPQ operations
type ShopeeMpqService struct {
	db           *gorm.DB
	tenantID     string
	shopeeAPI    *shopee.ProductAPI
	wholesaleSvc *ShopeeWholesaleService
}

// NewShopeeMpqService creates a new Shopee MPQ service
func NewShopeeMpqService(db *gorm.DB, tenantID string, shopeeAPI *shopee.ProductAPI) *ShopeeMpqService {
	return &ShopeeMpqService{
		db:           db,
		tenantID:     tenantID,
		shopeeAPI:    shopeeAPI,
		wholesaleSvc: NewShopeeWholesaleService(db, tenantID, shopeeAPI),
	}
}

// MpqResult represents MPQ operation result
type MpqResult struct {
	ItemID  int64  `json:"item_id"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SetMpq sets minimum purchase quantity for an item
func (s *ShopeeMpqService) SetMpq(ctx context.Context, itemID int64, mpq int) error {
	if s.shopeeAPI == nil {
		return fmt.Errorf("Shopee API client not configured")
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Int("mpq", mpq).
		Msg("Setting MPQ")

	err := s.shopeeAPI.UpdateItemMPQ(ctx, itemID, mpq)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Msg("Failed to set MPQ")
		return err
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Int("mpq", mpq).
		Msg("MPQ set successfully")

	return nil
}

// UpdatePrice updates price for an item/model
func (s *ShopeeMpqService) UpdatePrice(ctx context.Context, itemID int64, modelID *int64, price float64) error {
	if s.shopeeAPI == nil {
		return fmt.Errorf("Shopee API client not configured")
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Float64("price", price).
		Msg("Updating price")

	err := s.shopeeAPI.UpdatePrice(ctx, itemID, modelID, price)
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Msg("Failed to update price")
		return err
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Float64("price", price).
		Msg("Price updated successfully")

	return nil
}

// SetMpqMode performs complete MPQ mode setup:
// 1. Delete wholesale tiers
// 2. Update price (if provided)
// 3. Set MPQ
func (s *ShopeeMpqService) SetMpqMode(ctx context.Context, itemID int64, mpq int, price float64, modelID *int64) *MpqResult {
	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Int("mpq", mpq).
		Float64("price", price).
		Msg("Setting MPQ mode (delete wholesale + update price + set MPQ)")

	// Step 1: Delete wholesale tiers (wholesale and MPQ cannot coexist)
	if err := s.wholesaleSvc.DeleteWholesaleTiers(ctx, itemID); err != nil {
		log.Warn().
			Err(err).
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Msg("Warning: Failed to delete wholesale tiers (may not exist)")
		// Don't fail - wholesale may not exist
	}

	// Step 2: Update price (if provided and > 0)
	if price > 0 {
		if err := s.UpdatePrice(ctx, itemID, modelID, price); err != nil {
			return &MpqResult{
				ItemID:  itemID,
				Success: false,
				Error:   fmt.Sprintf("Failed to update price: %v", err),
			}
		}
		log.Info().
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Float64("price", price).
			Msg("Price updated successfully")
	}

	// Step 3: Set MPQ
	if err := s.SetMpq(ctx, itemID, mpq); err != nil {
		return &MpqResult{
			ItemID:  itemID,
			Success: false,
			Error:   fmt.Sprintf("Failed to set MPQ: %v", err),
		}
	}

	return &MpqResult{
		ItemID:  itemID,
		Success: true,
		Message: fmt.Sprintf("MPQ mode set: MPQ=%d, Price=%.2f", mpq, price),
	}
}

// BatchSetMpqBySkusResult represents batch MPQ result
type BatchSetMpqBySkusResult struct {
	TotalSKUs   int         `json:"total_skus"`
	UniqueItems int         `json:"unique_items"`
	Processed   int         `json:"processed"`
	Failed      int         `json:"failed"`
	Skipped     []string    `json:"skipped"`
	Success     bool        `json:"success"`
	Results     []MpqResult `json:"results"`
}

// BatchSetMpqBySkus sets MPQ for multiple SKUs
// Flow: Delete wholesale → Update price → Set MPQ for each item
func (s *ShopeeMpqService) BatchSetMpqBySkus(
	ctx context.Context,
	skuPriceMap map[string]float64,
	mpq int,
) (*BatchSetMpqBySkusResult, error) {
	result := &BatchSetMpqBySkusResult{
		TotalSKUs: len(skuPriceMap),
		Results:   make([]MpqResult, 0),
		Skipped:   make([]string, 0),
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int("sku_count", len(skuPriceMap)).
		Int("mpq", mpq).
		Msg("Batch set MPQ by SKUs")

	// Map SKUs to item_ids with prices
	itemMap := make(map[int64]struct {
		skus    []string
		price   float64
		modelID *int64
	})

	for sku, price := range skuPriceMap {
		var skuModel models.ShopeeSku
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND seller_sku = ?", s.tenantID, sku).
			First(&skuModel).Error

		if err == gorm.ErrRecordNotFound {
			result.Skipped = append(result.Skipped, sku)
			log.Warn().
				Str("tenant_id", s.tenantID).
				Str("sku", sku).
				Msg("SKU not found, skipping")
			continue
		}
		if err != nil {
			result.Failed++
			result.Skipped = append(result.Skipped, sku)
			log.Error().
				Err(err).
				Str("tenant_id", s.tenantID).
				Str("sku", sku).
				Msg("Database error, skipping")
			continue
		}

		if skuModel.ItemID != 0 {
			if existing, exists := itemMap[skuModel.ItemID]; exists {
				existing.skus = append(existing.skus, sku)
				itemMap[skuModel.ItemID] = existing
			} else {
				itemMap[skuModel.ItemID] = struct {
					skus    []string
					price   float64
					modelID *int64
				}{
					skus:    []string{sku},
					price:   price,
					modelID: skuModel.ModelID,
				}
			}
		}
	}

	result.UniqueItems = len(itemMap)

	// Process each unique item
	for itemID, data := range itemMap {
		mpqResult := s.SetMpqMode(ctx, itemID, mpq, data.price, data.modelID)
		mpqResult.Message = fmt.Sprintf("(SKUs: %v) - %s",
			data.skus, mpqResult.Message)

		result.Results = append(result.Results, *mpqResult)

		if mpqResult.Success {
			result.Processed++
		} else {
			result.Failed++
		}
	}

	result.Success = result.Failed == 0

	log.Info().
		Str("tenant_id", s.tenantID).
		Int("processed", result.Processed).
		Int("failed", result.Failed).
		Int("skipped", len(result.Skipped)).
		Msg("Batch MPQ completed")

	return result, nil
}
