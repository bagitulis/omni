package wholesale

import (
	"context"
	"fmt"

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

// GetSettings retrieves wholesale settings for tenant
func (s *WholesaleService) GetSettings(ctx context.Context) (*models.WholesaleSettings, error) {
	var settings models.WholesaleSettings
	err := s.db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		// Return default settings
		return &models.WholesaleSettings{
			TenantID:  s.tenantID,
			MinQty1:   5,
			Discount1: 5.0,
			MinQty2:   10,
			Discount2: 10.0,
			MinQty3:   20,
			Discount3: 15.0,
			IsActive:  true,
		}, nil
	}
	return &settings, err
}

// UpdateSettings updates wholesale settings
func (s *WholesaleService) UpdateSettings(ctx context.Context, settings *models.WholesaleSettings) error {
	settings.TenantID = s.tenantID
	return s.db.WithContext(ctx).Save(settings).Error
}

// CalculateTiers calculates wholesale tiers for a price
func (s *WholesaleService) CalculateTiers(ctx context.Context, originalPrice float64) (*models.WholesaleCalculateResponse, error) {
	settings, err := s.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	tiers := []models.WholesaleTier{
		{
			MinQty:   settings.MinQty1,
			MaxQty:   settings.MinQty2 - 1,
			Discount: settings.Discount1,
			Price:    s.applyDiscount(originalPrice, settings.Discount1),
		},
		{
			MinQty:   settings.MinQty2,
			MaxQty:   settings.MinQty3 - 1,
			Discount: settings.Discount2,
			Price:    s.applyDiscount(originalPrice, settings.Discount2),
		},
		{
			MinQty:   settings.MinQty3,
			MaxQty:   0, // Unlimited
			Discount: settings.Discount3,
			Price:    s.applyDiscount(originalPrice, settings.Discount3),
		},
	}

	return &models.WholesaleCalculateResponse{
		OriginalPrice: originalPrice,
		Tiers:         tiers,
	}, nil
}

// ApplyToProducts applies wholesale to specific products
func (s *WholesaleService) ApplyToProducts(ctx context.Context, api PlatformWholesaleAPI, platform string, skus []string) []models.WholesaleApplyResult {
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

	// Get current price to calculate tier prices
	// For now, we'll use discount percentages only
	tiers := []models.WholesaleTier{
		{MinQty: settings.MinQty1, Discount: settings.Discount1},
		{MinQty: settings.MinQty2, Discount: settings.Discount2},
		{MinQty: settings.MinQty3, Discount: settings.Discount3},
	}

	// Apply to platform
	if err := api.SetWholesaleTiers(ctx, itemID, tiers); err != nil {
		return models.WholesaleApplyResult{SKU: sku, Success: false, Error: err.Error()}
	}

	return models.WholesaleApplyResult{SKU: sku, Success: true}
}

func (s *WholesaleService) applyDiscount(price, discountPercent float64) float64 {
	return price * (1 - discountPercent/100)
}
