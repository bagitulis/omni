package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
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
	// IMPORTANT:
	// - Avoid wiping existing non-empty string fields (name/image/etc) when upstream sends empty.
	// - Use DB-level upsert to prevent race conditions.
	// Lazada products table has a unique constraint on (tenant_id, item_id).
	updates := map[string]interface{}{
		"description":  gorm.Expr("COALESCE(NULLIF(EXCLUDED.description, ''), lazada_products.description)"),
		"brand":        gorm.Expr("COALESCE(NULLIF(EXCLUDED.brand, ''), lazada_products.brand)"),
		"status":       gorm.Expr("COALESCE(NULLIF(EXCLUDED.status, ''), lazada_products.status)"),
		"price":        gorm.Expr("EXCLUDED.price"),
		"quantity":     gorm.Expr("EXCLUDED.quantity"),
		"image":        gorm.Expr("COALESCE(NULLIF(EXCLUDED.image, ''), lazada_products.image)"),
		"name":         gorm.Expr("COALESCE(NULLIF(EXCLUDED.name, ''), lazada_products.name)"),
		"local_images": gorm.Expr("COALESCE(EXCLUDED.local_images, lazada_products.local_images)"),
		// Cross-database timestamp (SQLite + Postgres).
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
	}

	// Use WithContext(ctx) and surface SQL in debug-level logs only.
	return r.db.WithContext(ctx).
		Session(&gorm.Session{Logger: r.db.Logger.LogMode(logger.Silent)}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "item_id"}},
			DoUpdates: clause.Assignments(updates),
		}).
		Create(product).Error
}

// UpsertSku creates or updates a SKU
func (r *LazadaProductRepository) UpsertSku(ctx context.Context, sku *models.LazadaSku) error {
	// SKU table is unique on sku_id.
	// Preserve non-empty string fields when upstream sends empty.
	updates := map[string]interface{}{
		"seller_sku":    gorm.Expr("COALESCE(NULLIF(EXCLUDED.seller_sku, ''), lazada_skus.seller_sku)"),
		"shop_sku":      gorm.Expr("COALESCE(NULLIF(EXCLUDED.shop_sku, ''), lazada_skus.shop_sku)"),
		"name":          gorm.Expr("COALESCE(NULLIF(EXCLUDED.name, ''), lazada_skus.name)"),
		"variant_name":  gorm.Expr("COALESCE(NULLIF(EXCLUDED.variant_name, ''), lazada_skus.variant_name)"),
		"variant_data":  gorm.Expr("COALESCE(EXCLUDED.variant_data, lazada_skus.variant_data)"),
		"price":         gorm.Expr("EXCLUDED.price"),
		"special_price": gorm.Expr("EXCLUDED.special_price"),
		"quantity":      gorm.Expr("EXCLUDED.quantity"),
		"available":     gorm.Expr("EXCLUDED.available"),
		// Cross-database timestamp (SQLite + Postgres).
		"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		"item_id":    gorm.Expr("COALESCE(NULLIF(EXCLUDED.item_id, ''), lazada_skus.item_id)"),
		"tenant_id":  gorm.Expr("COALESCE(NULLIF(EXCLUDED.tenant_id, ''), lazada_skus.tenant_id)"),
		"product_id": gorm.Expr("EXCLUDED.product_id"),
	}

	return r.db.WithContext(ctx).
		Session(&gorm.Session{Logger: r.db.Logger.LogMode(logger.Silent)}).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "sku_id"}},
			DoUpdates: clause.Assignments(updates),
		}).
		Create(sku).Error
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
