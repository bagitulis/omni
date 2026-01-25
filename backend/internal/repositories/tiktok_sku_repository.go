package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TiktokSkuRepository handles TikTok SKU database operations
type TiktokSkuRepository struct {
	db *gorm.DB
}

// NewTiktokSkuRepository creates a new TikTok SKU repository
func NewTiktokSkuRepository(db *gorm.DB) *TiktokSkuRepository {
	return &TiktokSkuRepository{db: db}
}

// Upsert creates or updates a TikTok SKU
func (r *TiktokSkuRepository) Upsert(ctx context.Context, sku *models.TiktokSku) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sku_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"product_id", "seller_sku", "price", "quantity", "variant_name", "variant_data", "updated_at"}),
	}).Create(sku).Error
}

// FindBySkuID finds a SKU by SKU ID
func (r *TiktokSkuRepository) FindBySkuID(ctx context.Context, skuID string) (*models.TiktokSku, error) {
	var sku models.TiktokSku
	err := r.db.WithContext(ctx).Where("sku_id = ?", skuID).First(&sku).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &sku, err
}

// FindByProductID finds all SKUs for a product
func (r *TiktokSkuRepository) FindByProductID(ctx context.Context, productID uint) ([]models.TiktokSku, error) {
	var skus []models.TiktokSku
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).Find(&skus).Error
	return skus, err
}

// DeleteByProductID deletes all SKUs for a product
func (r *TiktokSkuRepository) DeleteByProductID(ctx context.Context, productID uint) error {
	return r.db.WithContext(ctx).Where("product_id = ?", productID).Delete(&models.TiktokSku{}).Error
}

// Count returns total SKU count
func (r *TiktokSkuRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.TiktokSku{}).Count(&count).Error
	return count, err
}
