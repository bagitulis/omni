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

// ModelPriceInfo holds price info for a single model/variant within an item
type ModelPriceInfo struct {
	SKU     string
	ModelID *int64
	Price   float64
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

// updateAllModelPrices updates prices for multiple models of an item
func (s *ShopeeMpqService) updateAllModelPrices(ctx context.Context, itemID int64, modelPrices []ModelPriceInfo) error {
	if s.shopeeAPI == nil {
		return fmt.Errorf("Shopee API client not configured")
	}

	for _, m := range modelPrices {
		log.Info().
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Str("sku", m.SKU).
			Float64("price", m.Price).
			Msg("Updating model price")

		if err := s.shopeeAPI.UpdatePrice(ctx, itemID, m.ModelID, m.Price); err != nil {
			return fmt.Errorf("failed to update price for SKU %s: %w", m.SKU, err)
		}

		log.Info().
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Str("sku", m.SKU).
			Float64("price", m.Price).
			Msg("Price updated successfully")
	}

	return nil
}

// SetMpqMode performs complete MPQ mode setup for an item with ALL its models:
// 1. Delete wholesale tiers
// 2. Update price for EACH model/variant
// 3. Set MPQ
func (s *ShopeeMpqService) SetMpqMode(ctx context.Context, itemID int64, mpq int, allModels []ModelPriceInfo) *MpqResult {
	skuNames := make([]string, len(allModels))
	for i, m := range allModels {
		skuNames[i] = m.SKU
	}

	log.Info().
		Str("tenant_id", s.tenantID).
		Int64("item_id", itemID).
		Int("mpq", mpq).
		Int("model_count", len(allModels)).
		Strs("skus", skuNames).
		Msg("Setting MPQ mode (delete wholesale + update prices + set MPQ)")

	// Step 1: Delete wholesale tiers (wholesale and MPQ cannot coexist)
	if err := s.wholesaleSvc.DeleteWholesaleTiers(ctx, itemID); err != nil {
		log.Warn().
			Err(err).
			Str("tenant_id", s.tenantID).
			Int64("item_id", itemID).
			Msg("Warning: Failed to delete wholesale tiers (may not exist)")
	}

	// Step 2: Update price for ALL models of this item
	if len(allModels) > 0 {
		if err := s.updateAllModelPrices(ctx, itemID, allModels); err != nil {
			return &MpqResult{
				ItemID:  itemID,
				Success: false,
				Error:   fmt.Sprintf("Failed to update prices: %v", err),
			}
		}
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
		Message: fmt.Sprintf("MPQ=%d set for %d models", mpq, len(allModels)),
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
// Flow: Group SKUs by item_id → Delete wholesale → Update ALL model prices → Set MPQ
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

	// Map SKUs to item_ids — store ALL models per item (not just first)
	itemMap := make(map[int64][]ModelPriceInfo)

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
			continue
		}

		if skuModel.ItemID != 0 {
			itemMap[skuModel.ItemID] = append(itemMap[skuModel.ItemID], ModelPriceInfo{
				SKU:     sku,
				ModelID: skuModel.ModelID,
				Price:   price,
			})
		}
	}

	result.UniqueItems = len(itemMap)

	// Process each unique item with ALL its models
	for itemID, allModels := range itemMap {
		mpqResult := s.SetMpqMode(ctx, itemID, mpq, allModels)
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
