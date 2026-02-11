package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// TiktokProductRepository handles TikTok product data access
type TiktokProductRepository struct {
	db *gorm.DB
}

// NewTiktokProductRepository creates a new repository instance
func NewTiktokProductRepository(db *gorm.DB) *TiktokProductRepository {
	return &TiktokProductRepository{db: db}
}

// FindAll returns paginated products
func (r *TiktokProductRepository) FindAll(ctx context.Context, page, pageSize int) ([]models.TiktokProduct, int64, error) {
	var products []models.TiktokProduct
	var total int64

	r.db.Model(&models.TiktokProduct{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// FindByProductID returns product by TikTok product ID
func (r *TiktokProductRepository) FindByProductID(ctx context.Context, productID string) (*models.TiktokProduct, error) {
	var product models.TiktokProduct
	err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Upsert creates or updates product
func (r *TiktokProductRepository) Upsert(ctx context.Context, product *models.TiktokProduct) error {
	return r.db.WithContext(ctx).
		Where("product_id = ?", product.ProductID).
		Assign(*product).
		FirstOrCreate(product).Error
}

// Search searches products by query
func (r *TiktokProductRepository) Search(ctx context.Context, query string, page, pageSize int) ([]models.TiktokProduct, int64, error) {
	var products []models.TiktokProduct
	var total int64

	searchQuery := "%" + query + "%"

	r.db.Model(&models.TiktokProduct{}).
		Where("name ILIKE ? OR description ILIKE ? OR product_id ILIKE ?", searchQuery, searchQuery, searchQuery).
		Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Where("name ILIKE ? OR description ILIKE ? OR product_id ILIKE ?", searchQuery, searchQuery, searchQuery).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}
