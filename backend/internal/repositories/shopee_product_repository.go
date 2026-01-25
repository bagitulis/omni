package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ShopeeProductRepository handles Shopee product data access
type ShopeeProductRepository struct {
	db *gorm.DB
}

// NewShopeeProductRepository creates a new repository instance
func NewShopeeProductRepository(db *gorm.DB) *ShopeeProductRepository {
	return &ShopeeProductRepository{db: db}
}

// FindAll returns paginated products
func (r *ShopeeProductRepository) FindAll(ctx context.Context, page, pageSize int) ([]models.ShopeeProduct, int64, error) {
	var products []models.ShopeeProduct
	var total int64

	r.db.Model(&models.ShopeeProduct{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// FindByItemID returns product by Shopee item ID
func (r *ShopeeProductRepository) FindByItemID(ctx context.Context, itemID int64) (*models.ShopeeProduct, error) {
	var product models.ShopeeProduct
	err := r.db.WithContext(ctx).Where("item_id = ?", itemID).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Create inserts a new product
func (r *ShopeeProductRepository) Create(ctx context.Context, product *models.ShopeeProduct) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// Update modifies an existing product
func (r *ShopeeProductRepository) Update(ctx context.Context, product *models.ShopeeProduct) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Upsert creates or updates product by ItemID
func (r *ShopeeProductRepository) Upsert(ctx context.Context, product *models.ShopeeProduct) error {
	return r.db.WithContext(ctx).
		Where("item_id = ?", product.ItemID).
		Assign(*product).
		FirstOrCreate(product).Error
}

// FindBySKU finds product by SKU from ShopeeSku table
func (r *ShopeeProductRepository) FindBySKU(ctx context.Context, sku string) (*models.ShopeeProduct, error) {
	var skuRecord models.ShopeeSku
	if err := r.db.WithContext(ctx).Where("model_sku = ?", sku).First(&skuRecord).Error; err != nil {
		return nil, err
	}
	return r.FindByItemID(ctx, skuRecord.ItemID)
}
