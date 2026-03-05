package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ShopeeSkuRepository handles Shopee SKU data access
type ShopeeSkuRepository struct {
	db *gorm.DB
}

// NewShopeeSkuRepository creates a new repository instance
func NewShopeeSkuRepository(db *gorm.DB) *ShopeeSkuRepository {
	return &ShopeeSkuRepository{db: db}
}

// FindByModelID finds SKU by model ID
func (r *ShopeeSkuRepository) FindByModelID(ctx context.Context, modelID int64) (*models.ShopeeSku, error) {
	var sku models.ShopeeSku
	err := r.db.WithContext(ctx).Where("model_id = ?", modelID).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

// FindByItemID finds all SKUs for an item
func (r *ShopeeSkuRepository) FindByItemID(ctx context.Context, itemID int64) ([]models.ShopeeSku, error) {
	var skus []models.ShopeeSku
	err := r.db.WithContext(ctx).Where("item_id = ?", itemID).Find(&skus).Error
	return skus, err
}

// Create inserts a new SKU
func (r *ShopeeSkuRepository) Create(ctx context.Context, sku *models.ShopeeSku) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

// Update modifies an existing SKU
func (r *ShopeeSkuRepository) Update(ctx context.Context, sku *models.ShopeeSku) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

// Upsert creates or updates SKU by ModelID (or ItemID for non-variant products).
// Uses explicit map-based updates so zero-value fields (e.g. Quantity=0) are written.
func (r *ShopeeSkuRepository) Upsert(ctx context.Context, sku *models.ShopeeSku) error {
	var existing models.ShopeeSku
	var err error

	if sku.ModelID != nil && *sku.ModelID > 0 {
		err = r.db.WithContext(ctx).Where("model_id = ?", sku.ModelID).First(&existing).Error
	} else {
		err = r.db.WithContext(ctx).Where("item_id = ? AND model_id IS NULL", sku.ItemID).First(&existing).Error
	}

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(sku).Error
	}
	if err != nil {
		return err
	}

	// Map-based update ensures zero values (Quantity=0, Price=0) are written
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"tenant_id":    sku.TenantID,
		"product_id":   sku.ProductID,
		"item_id":      sku.ItemID,
		"model_id":     sku.ModelID,
		"seller_sku":   sku.SellerSku,
		"variant_name": sku.VariantName,
		"price":        sku.Price,
		"quantity":     sku.Quantity,
	}).Error
}

// DeleteByItemID removes all SKUs for an item
func (r *ShopeeSkuRepository) DeleteByItemID(ctx context.Context, itemID int64) error {
	return r.db.WithContext(ctx).Where("item_id = ?", itemID).Delete(&models.ShopeeSku{}).Error
}
