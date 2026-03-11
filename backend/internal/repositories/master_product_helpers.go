package repositories

import (
	"context"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// paginatedQuery applies pagination, preloading, and ordering to a base query,
// returns the items and total count.
func (r *MasterProductRepository) paginatedQuery(
	ctx context.Context,
	baseQuery *gorm.DB,
	page, pageSize int,
) ([]models.MasterProduct, int64, error) {
	var products []models.MasterProduct
	var total int64

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := baseQuery.
		Preload("SKUs").
		Preload("SKUs.PlatformLinks").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// inventoryExistsSubQuery builds a sub-query that checks whether any of a
// product's SKUs exist in inventory_records. Caller wraps with EXISTS / NOT EXISTS.
func (r *MasterProductRepository) inventoryExistsSubQuery(ctx context.Context, tenantID string) *gorm.DB {
	inventoryTableName := models.GetTableName("InventoryRecord")
	skuTableName := models.GetTableName("MasterProductSku")

	return r.db.WithContext(ctx).
		Table(skuTableName+" AS mps").
		Select("1").
		Where("mps.master_product_id = master_products.id").
		Where("EXISTS (SELECT 1 FROM "+inventoryTableName+" AS ir WHERE LOWER(ir.key_value) = LOWER(mps.seller_sku) AND ir.tenant_id = ?)", tenantID)
}
