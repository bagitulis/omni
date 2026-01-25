// Package price provides price management services
package price

import (
	"context"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	inventorySvc "github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// PriceService handles price operations
type PriceService struct {
	db          *gorm.DB
	tenantID    string
	dbPath      string
	credService *services.CredentialService
}

// NewPriceService creates a new price service
func NewPriceService(db *gorm.DB, tenantID string) *PriceService {
	return &PriceService{db: db, tenantID: tenantID}
}

// NewPriceServiceWithCreds creates a price service with credential support
func NewPriceServiceWithCreds(db *gorm.DB, tenantID, dbPath string) *PriceService {
	return &PriceService{
		db:          db,
		tenantID:    tenantID,
		dbPath:      dbPath,
		credService: services.NewCredentialService(dbPath),
	}
}

// PriceDTO represents price data for API
type PriceDTO struct {
	SKU            string          `json:"sku"`
	ProductName    string          `json:"product_name"`
	CurrentPrice   float64         `json:"current_price"`
	Cost           float64         `json:"cost"`
	Margin         float64         `json:"margin"`
	PlatformPrices []PlatformPrice `json:"platform_prices"`
}

// PlatformPrice represents price per platform
type PlatformPrice struct {
	Platform       string  `json:"platform"`
	PlatformItemID string  `json:"platform_item_id"`
	Price          float64 `json:"price"`
	OriginalPrice  float64 `json:"original_price,omitempty"`
	Currency       string  `json:"currency"`
	LastSyncedAt   string  `json:"last_synced_at,omitempty"`
}

// PriceUpdateRequest represents request to update price
type PriceUpdateRequest struct {
	SKU       string   `json:"sku" binding:"required"`
	Price     float64  `json:"price" binding:"required"`
	Platforms []string `json:"platforms"` // empty = all platforms
}

// BulkPriceUpdateRequest represents bulk price update request
type BulkPriceUpdateRequest struct {
	Updates []PriceUpdateRequest `json:"updates" binding:"required"`
}

// PriceUpdateResult represents result of price update
type PriceUpdateResult struct {
	SKU             string                `json:"sku"`
	Success         bool                  `json:"success"`
	Message         string                `json:"message,omitempty"`
	PlatformResults []PlatformPriceResult `json:"platform_results,omitempty"`
}

// PlatformPriceResult represents result per platform
type PlatformPriceResult struct {
	Platform string  `json:"platform"`
	Success  bool    `json:"success"`
	Message  string  `json:"message,omitempty"`
	OldPrice float64 `json:"old_price,omitempty"`
	NewPrice float64 `json:"new_price,omitempty"`
}

// BulkPriceUpdateResult represents bulk update result
type BulkPriceUpdateResult struct {
	TotalRequested int                 `json:"total_requested"`
	TotalSuccess   int                 `json:"total_success"`
	TotalFailed    int                 `json:"total_failed"`
	Results        []PriceUpdateResult `json:"results"`
}

// GetPrice retrieves price for a SKU
func (s *PriceService) GetPrice(ctx context.Context, sku string) (*PriceDTO, error) {
	var record models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", s.tenantID, sku).
		First(&record).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Extract values from JSONB
	skuVal := inventorySvc.GetSKU(record)
	productName := inventorySvc.GetProductName(record)
	price := inventorySvc.GetPrice(record)
	cost := inventorySvc.GetDataFloat(record, "Cost")

	platformPrices, _ := s.getPlatformPrices(ctx, skuVal)
	margin := 0.0
	if cost > 0 {
		margin = ((price - cost) / cost) * 100
	}

	return &PriceDTO{
		SKU:            skuVal,
		ProductName:    productName,
		CurrentPrice:   price,
		Cost:           cost,
		Margin:         margin,
		PlatformPrices: platformPrices,
	}, nil
}

// GetAllPrices retrieves all prices
func (s *PriceService) GetAllPrices(ctx context.Context, limit, offset int) ([]PriceDTO, int64, error) {
	var total int64
	if err := s.db.WithContext(ctx).Model(&models.InventoryRecord{}).
		Where("tenant_id = ?", s.tenantID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit == 0 {
		limit = 50
	}

	var records []models.InventoryRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", s.tenantID).
		Limit(limit).Offset(offset).Order("key_value ASC").
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	prices := make([]PriceDTO, len(records))
	for i, r := range records {
		skuVal := inventorySvc.GetSKU(r)
		productName := inventorySvc.GetProductName(r)
		price := inventorySvc.GetPrice(r)
		cost := inventorySvc.GetDataFloat(r, "Cost")

		platformPrices, _ := s.getPlatformPrices(ctx, skuVal)
		margin := 0.0
		if cost > 0 {
			margin = ((price - cost) / cost) * 100
		}
		prices[i] = PriceDTO{
			SKU:            skuVal,
			ProductName:    productName,
			CurrentPrice:   price,
			Cost:           cost,
			Margin:         margin,
			PlatformPrices: platformPrices,
		}
	}

	return prices, total, nil
}
