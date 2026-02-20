package repositories

import (
	"context"
	"errors"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// MasterProductRepository handles Master Product data access
type MasterProductRepository struct {
	db *gorm.DB
}

// NewMasterProductRepository creates a new repository instance
func NewMasterProductRepository(db *gorm.DB) *MasterProductRepository {
	return &MasterProductRepository{db: db}
}

// ================== Product CRUD ==================

// Create inserts a new master product
func (r *MasterProductRepository) Create(ctx context.Context, product *models.MasterProduct) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// FindByID returns a master product by ID with SKUs preloaded
func (r *MasterProductRepository) FindByID(ctx context.Context, id uint) (*models.MasterProduct, error) {
	var product models.MasterProduct
	err := r.db.WithContext(ctx).
		Preload("SKUs").
		Preload("SKUs.PlatformLinks").
		Where("id = ?", id).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindAll returns paginated master products for a tenant
func (r *MasterProductRepository) FindAll(ctx context.Context, tenantID string, page, pageSize int) ([]models.MasterProduct, int64, error) {
	var products []models.MasterProduct
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Preload("SKUs").
		Preload("SKUs.PlatformLinks").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// Update modifies an existing master product
func (r *MasterProductRepository) Update(ctx context.Context, product *models.MasterProduct) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete removes a master product (hard delete)
func (r *MasterProductRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.MasterProduct{}, id).Error
}

// FindByTenantAndID returns a product ensuring tenant isolation
func (r *MasterProductRepository) FindByTenantAndID(ctx context.Context, tenantID string, id uint) (*models.MasterProduct, error) {
	var product models.MasterProduct
	err := r.db.WithContext(ctx).
		Preload("SKUs").
		Preload("SKUs.PlatformLinks").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// ================== SKU Operations ==================

// FindBySku finds a SKU by seller_sku within a tenant
func (r *MasterProductRepository) FindBySku(ctx context.Context, tenantID, sellerSku string) (*models.MasterProductSku, error) {
	var sku models.MasterProductSku
	err := r.db.WithContext(ctx).
		Preload("PlatformLinks").
		Where("tenant_id = ? AND seller_sku = ?", tenantID, sellerSku).
		First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}

// FindByExactTitle finds a master product by exact normalized title within a tenant.
// normalizedTitle must already be lowercased and trimmed by the caller.
// Returns (nil, nil) if no matching product is found.
func (r *MasterProductRepository) FindByExactTitle(ctx context.Context, tenantID, normalizedTitle string) (*models.MasterProduct, error) {
	var product models.MasterProduct
	result := r.db.WithContext(ctx).
		Where("tenant_id = ? AND regexp_replace(LOWER(TRIM(title)), '\\s+', ' ', 'g') = ?", tenantID, normalizedTitle).
		First(&product)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &product, nil
}

// CreateSku inserts a new SKU
func (r *MasterProductRepository) CreateSku(ctx context.Context, sku *models.MasterProductSku) error {
	return r.db.WithContext(ctx).Create(sku).Error
}

// UpdateSku modifies an existing SKU
func (r *MasterProductRepository) UpdateSku(ctx context.Context, sku *models.MasterProductSku) error {
	return r.db.WithContext(ctx).Save(sku).Error
}

// DeleteSku removes a SKU
func (r *MasterProductRepository) DeleteSku(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.MasterProductSku{}, id).Error
}

// FindSkusByProductID returns all SKUs for a master product
func (r *MasterProductRepository) FindSkusByProductID(ctx context.Context, masterProductID uint) ([]models.MasterProductSku, error) {
	var skus []models.MasterProductSku
	err := r.db.WithContext(ctx).
		Preload("PlatformLinks").
		Where("master_product_id = ?", masterProductID).
		Order("created_at ASC").
		Find(&skus).Error
	return skus, err
}

// ================== Query Helpers ==================

// CountByTenant returns total count of master products for a tenant
func (r *MasterProductRepository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

// FindByStatus returns products by status for a tenant
func (r *MasterProductRepository) FindByStatus(ctx context.Context, tenantID, status string, page, pageSize int) ([]models.MasterProduct, int64, error) {
	var products []models.MasterProduct
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ? AND status = ?", tenantID, status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("SKUs").
		Preload("SKUs.PlatformLinks").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// SearchByTitle searches products by title (case-insensitive)
func (r *MasterProductRepository) SearchByTitle(ctx context.Context, tenantID, search string, page, pageSize int) ([]models.MasterProduct, int64, error) {
	var products []models.MasterProduct
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ? AND LOWER(title) LIKE LOWER(?)", tenantID, "%"+search+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("SKUs").
		Preload("SKUs.PlatformLinks").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// FindLinked returns paginated products that are linked on marketplace platforms.
func (r *MasterProductRepository) FindLinked(
	ctx context.Context,
	tenantID string,
	page, pageSize int,
	status, search, platform string,
) ([]models.MasterProduct, int64, error) {
	var products []models.MasterProduct
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("LOWER(title) LIKE LOWER(?)", "%"+search+"%")
	}

	linkedStatuses := []string{models.SyncStatusSynced, models.SyncStatusOutdated, models.SyncStatusPending, models.SyncStatusError}
	linkedSubQuery := r.db.WithContext(ctx).
		Model(&models.MasterProductPlatformLink{}).
		Select("1").
		Where("master_product_platform_links.master_product_id = master_products.id").
		Where("master_product_platform_links.sync_status IN ?", linkedStatuses)

	if platform != "" {
		linkedSubQuery = linkedSubQuery.Where("master_product_platform_links.platform = ?", platform)
	}

	query = query.Where("EXISTS (?)", linkedSubQuery)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("SKUs").
		Preload("SKUs.PlatformLinks").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// FindForImageBackfill returns master products for image backfill.
func (r *MasterProductRepository) FindForImageBackfill(ctx context.Context, tenantID string, limit int, force bool) ([]models.MasterProduct, error) {
	var products []models.MasterProduct

	query := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Select("id", "tenant_id", "images").
		Where("tenant_id = ?", tenantID)

	if !force {
		query = query.Where("images IS NULL OR jsonb_array_length(images) = 0")
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("id ASC").Find(&products).Error
	return products, err
}

// FindImagesByID returns the images array for a master product.
func (r *MasterProductRepository) FindImagesByID(ctx context.Context, tenantID string, id uint) (models.JSONArray, error) {
	var result struct {
		Images models.JSONArray `gorm:"column:images"`
	}

	err := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Select("images").
		Where("id = ? AND tenant_id = ?", id, tenantID).
		First(&result).Error

	if err != nil {
		return nil, err
	}

	return result.Images, nil
}

// UpdateImages updates the images array for a master product.
func (r *MasterProductRepository) UpdateImages(ctx context.Context, tenantID string, id uint, images models.JSONArray) error {
	return r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("id = ? AND tenant_id = ?", id, tenantID).
		Update("images", images).Error
}
