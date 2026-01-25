package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// LazadaProductRepository handles Lazada product data access
type LazadaProductRepository struct {
	db *gorm.DB
}

// NewLazadaProductRepository creates a new repository instance
func NewLazadaProductRepository(db *gorm.DB) *LazadaProductRepository {
	return &LazadaProductRepository{db: db}
}

// FindAll returns paginated products
func (r *LazadaProductRepository) FindAll(ctx context.Context, page, pageSize int) ([]models.LazadaProduct, int64, error) {
	var products []models.LazadaProduct
	var total int64

	r.db.Model(&models.LazadaProduct{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// FindByItemID returns product by item ID
func (r *LazadaProductRepository) FindByItemID(ctx context.Context, itemID string) (*models.LazadaProduct, error) {
	var product models.LazadaProduct
	err := r.db.WithContext(ctx).Where("item_id = ?", itemID).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// Upsert creates or updates product
func (r *LazadaProductRepository) Upsert(ctx context.Context, product *models.LazadaProduct) error {
	return r.db.WithContext(ctx).
		Where("item_id = ?", product.ItemID).
		Assign(*product).
		FirstOrCreate(product).Error
}

// UpsertSku creates or updates a SKU
func (r *LazadaProductRepository) UpsertSku(ctx context.Context, sku *models.LazadaSku) error {
	return r.db.WithContext(ctx).
		Where("sku_id = ?", sku.SkuID).
		Assign(*sku).
		FirstOrCreate(sku).Error
}

// FindSkusByItemID returns all SKUs for a product
func (r *LazadaProductRepository) FindSkusByItemID(ctx context.Context, itemID string) ([]models.LazadaSku, error) {
	var skus []models.LazadaSku
	err := r.db.WithContext(ctx).Where("item_id = ?", itemID).Find(&skus).Error
	return skus, err
}

// GetProductsWithSkus returns products with their SKUs for frontend display
func (r *LazadaProductRepository) GetProductsWithSkus(ctx context.Context, page, pageSize int) ([]map[string]interface{}, int64, error) {
	var total int64
	var results []map[string]interface{}

	// Count total products
	r.db.Model(&models.LazadaProduct{}).Count(&total)

	// Query products with LEFT JOIN to skus
	offset := (page - 1) * pageSize
	rows, err := r.db.WithContext(ctx).
		Table("lazada_products p").
		Select(`
			p.item_id,
			COALESCE(s.sku_id, '') as sku_id,
			COALESCE(s.name, p.name) as sku_name,
			p.name as product_name,
			COALESCE(s.variant_name, '') as variant_name,
			COALESCE(s.price, p.price) as price,
			COALESCE(s.quantity, p.quantity) as quantity,
			p.status,
			COALESCE(s.updated_at, p.updated_at) as updated_at
		`).
		Joins("LEFT JOIN lazada_skus s ON p.item_id = s.item_id AND p.tenant_id = s.tenant_id").
		Order("p.created_at DESC, s.created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Rows()

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var itemID, skuID, skuName, productName, variantName, status string
		var price float64
		var quantity int
		var updatedAt interface{}

		if err := rows.Scan(&itemID, &skuID, &skuName, &productName, &variantName, &price, &quantity, &status, &updatedAt); err != nil {
			continue
		}

		results = append(results, map[string]interface{}{
			"item_id":      itemID,
			"sku_id":       skuID,
			"sku_name":     skuName,
			"product_name": productName,
			"variant_name": variantName,
			"price":        price,
			"quantity":     quantity,
			"status":       status,
			"updated_at":   updatedAt,
		})
	}

	return results, total, nil
}
