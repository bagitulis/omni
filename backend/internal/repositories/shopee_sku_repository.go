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

// Upsert creates or updates SKU by ModelID (or ItemID for non-variant products)
func (r *ShopeeSkuRepository) Upsert(ctx context.Context, sku *models.ShopeeSku) error {
	if sku.ModelID != nil && *sku.ModelID > 0 {
		// Product with variants - use ModelID as key
		return r.db.WithContext(ctx).
			Where("model_id = ?", sku.ModelID).
			Assign(*sku).
			FirstOrCreate(sku).Error
	}
	// Non-variant product - use ItemID as key
	return r.db.WithContext(ctx).
		Where("item_id = ? AND model_id IS NULL", sku.ItemID).
		Assign(*sku).
		FirstOrCreate(sku).Error
}

// DeleteByItemID removes all SKUs for an item
func (r *ShopeeSkuRepository) DeleteByItemID(ctx context.Context, itemID int64) error {
	return r.db.WithContext(ctx).Where("item_id = ?", itemID).Delete(&models.ShopeeSku{}).Error
}
