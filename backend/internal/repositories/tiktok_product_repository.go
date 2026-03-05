package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// Upsert creates or updates product.
// Uses explicit map-based updates so zero-value fields (e.g. Quantity=0) are written.
func (r *TiktokProductRepository) Upsert(ctx context.Context, product *models.TiktokProduct) error {
	var existing models.TiktokProduct
	err := r.db.WithContext(ctx).Where("product_id = ?", product.ProductID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(product).Error
	}
	if err != nil {
		return err
	}

	// Map-based update ensures zero values (Quantity=0, Price=0) are written
	return r.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"tenant_id":   product.TenantID,
		"name":        product.Name,
		"description": product.Description,
		"status":      product.Status,
		"price":       product.Price,
		"quantity":    product.Quantity,
		"image":       product.Image,
	}).Error
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

// UpsertSku creates or updates a TikTok SKU
func (r *TiktokProductRepository) UpsertSku(ctx context.Context, sku *models.TiktokSku) error {
	updates := map[string]interface{}{
		"seller_sku":   gorm.Expr("COALESCE(NULLIF(EXCLUDED.seller_sku, ''), tiktok_skus.seller_sku)"),
		"variant_name": gorm.Expr("COALESCE(NULLIF(EXCLUDED.variant_name, ''), tiktok_skus.variant_name)"),
		"variant_data": gorm.Expr("COALESCE(EXCLUDED.variant_data, tiktok_skus.variant_data)"),
		"price":        gorm.Expr("EXCLUDED.price"),
		"quantity":     gorm.Expr("EXCLUDED.quantity"),
		"updated_at":   gorm.Expr("CURRENT_TIMESTAMP"),
		"product_id":   gorm.Expr("EXCLUDED.product_id"),
		"tenant_id":    gorm.Expr("COALESCE(NULLIF(EXCLUDED.tenant_id, ''), tiktok_skus.tenant_id)"),
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "sku_id"}},
			DoUpdates: clause.Assignments(updates),
		}).
		Create(sku).Error
}
