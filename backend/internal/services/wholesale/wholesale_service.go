package wholesale

import (
	"context"
	"fmt"
	"math"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// PlatformWholesaleAPI defines interface for platform wholesale operations
type PlatformWholesaleAPI interface {
	SetWholesaleTiers(ctx context.Context, itemID int64, tiers []models.WholesaleTier) error
	GetItemBysku(ctx context.Context, sku string) (int64, error)
}

// WholesaleService handles wholesale operations
type WholesaleService struct {
	db       *gorm.DB
	tenantID string
}

// NewWholesaleService creates a new wholesale service
func NewWholesaleService(db *gorm.DB, tenantID string) *WholesaleService {
	return &WholesaleService{db: db, tenantID: tenantID}
}

// DefaultSettings returns the default wholesale settings
// Matches Node.js defaults (wholesaleSettingsService.ts DEFAULT_SETTINGS)
func DefaultSettings(tenantID string) *models.WholesaleSettings {
	return &models.WholesaleSettings{
		TenantID:      tenantID,
		Platform:      "shopee",
		AdminFee:      1500,
		MinOrder1:     2,
		MaxOrder1:     3,
		MaxOrderTier3: 1000,
		IsActive:      true,
	}
}

// GetSettings retrieves wholesale settings for tenant
func (s *WholesaleService) GetSettings(ctx context.Context) (*models.WholesaleSettings, error) {
	var settings models.WholesaleSettings
	err := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		return DefaultSettings(s.tenantID), nil
	}
	return &settings, err
}

// UpdateSettings updates wholesale settings
func (s *WholesaleService) UpdateSettings(ctx context.Context, settings *models.WholesaleSettings) error {
	settings.TenantID = s.tenantID
	if settings.Platform == "" {
		settings.Platform = "shopee"
	}
	return s.db.WithContext(ctx).Save(settings).Error
}

// CalculateTiers calculates wholesale tiers for a price
func (s *WholesaleService) CalculateTiers(ctx context.Context, basePrice float64) (*models.WholesaleCalculateResponse, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	tiers := s.CalculateTiersFromSettings(basePrice, settings)

	return &models.WholesaleCalculateResponse{
		BasePrice: basePrice,
		AdminFee:  settings.AdminFee,
		Tiers: func() []models.WholesaleTier {
			result := make([]models.WholesaleTier, len(tiers))
			for i, t := range tiers {
				result[i] = models.WholesaleTier{
					MinCount:  t.MinCount,
					MaxCount:  t.MaxCount,
					UnitPrice: t.UnitPrice,
				}
			}
			return result
		}(),
	}, nil
}

// ApplyToProducts applies wholesale to specific products
func (s *WholesaleService) ApplyToProducts(ctx context.Context, api PlatformWholesaleAPI, _ string, skus []string) []models.WholesaleApplyResult {
	results := make([]models.WholesaleApplyResult, len(skus))

	settings, err := s.GetSettings(ctx)
	if err != nil {
		for i, sku := range skus {
			results[i] = models.WholesaleApplyResult{SKU: sku, Success: false, Error: err.Error()}
		}
		return results
	}

	for i, sku := range skus {
		result := s.applyToSingleProduct(ctx, api, sku, settings)
		results[i] = result
	}

	return results
}

func (s *WholesaleService) applyToSingleProduct(ctx context.Context, api PlatformWholesaleAPI, sku string, settings *models.WholesaleSettings) models.WholesaleApplyResult {
	// Get item ID from SKU
	itemID, err := api.GetItemBysku(ctx, sku)
	if err != nil {
		return models.WholesaleApplyResult{SKU: sku, Success: false, Error: fmt.Sprintf("item not found: %v", err)}
	}

	// Lookup actual price from database (ShopeeSku table)
	basePrice, err := s.lookupSkuPrice(ctx, sku)
	if err != nil || basePrice <= 0 {
		return models.WholesaleApplyResult{SKU: sku, Success: false, Error: fmt.Sprintf("price not found for SKU %s: %v", sku, err)}
	}

	// Calculate tiers using admin fee formula with real price
	tiers := s.CalculateTiersFromSettings(basePrice, settings)
	apiTiers := make([]models.WholesaleTier, len(tiers))
	for i, t := range tiers {
		apiTiers[i] = models.WholesaleTier{
			MinCount:  t.MinCount,
			MaxCount:  t.MaxCount,
			UnitPrice: t.UnitPrice,
		}
	}

	// Apply to platform
	if err := api.SetWholesaleTiers(ctx, itemID, apiTiers); err != nil {
		return models.WholesaleApplyResult{SKU: sku, Success: false, Error: err.Error()}
	}

	return models.WholesaleApplyResult{SKU: sku, Success: true}
}

// lookupSkuPrice finds the price of a SKU from the shopee_skus table
func (s *WholesaleService) lookupSkuPrice(ctx context.Context, sku string) (float64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not configured")
	}
	var skuRecord models.ShopeeSku
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND seller_sku = ?", s.tenantID, sku).
		First(&skuRecord).Error
	if err != nil {
		return 0, err
	}
	return skuRecord.Price, nil
}

// calculateAdminFeePrice calculates tier price using admin fee redistribution formula
// Formula: (BasePrice - AdminFee) + (AdminFee / MinQty)
// Matches Node.js: Math.round(basePrice - adminFee + adminFee / minQty)
func calculateAdminFeePrice(basePrice float64, adminFee int, minQty int) float64 {
	if minQty <= 0 {
		return basePrice
	}
	return math.Round(basePrice - float64(adminFee) + float64(adminFee)/float64(minQty))
}

// CalculateTiersFromSettings calculates wholesale tiers based on base price and settings
// Tier boundaries use cascade logic matching Node.js:
// - Tier 1: minOrder1 → maxOrder1 (editable)
// - Tier 2: maxOrder1+1 → maxOrder1+2 (auto, range=2)
// - Tier 3: maxOrder1+3 → maxOrderTier3 (auto, up to max)
func (s *WholesaleService) CalculateTiersFromSettings(basePrice float64, settings *models.WholesaleSettings) []WholesaleTier {
	min1 := settings.MinOrder1
	max1 := settings.MaxOrder1
	min2 := max1 + 1
	max2 := min2 + 1
	min3 := max2 + 1
	max3 := settings.MaxOrderTier3

	return []WholesaleTier{
		{
			MinCount:  min1,
			MaxCount:  max1,
			UnitPrice: calculateAdminFeePrice(basePrice, settings.AdminFee, min1),
		},
		{
			MinCount:  min2,
			MaxCount:  max2,
			UnitPrice: calculateAdminFeePrice(basePrice, settings.AdminFee, min2),
		},
		{
			MinCount:  min3,
			MaxCount:  max3,
			UnitPrice: calculateAdminFeePrice(basePrice, settings.AdminFee, min3),
		},
	}
}
