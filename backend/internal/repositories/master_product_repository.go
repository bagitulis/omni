package repositories

import (
	"context"
	"time"

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

	// Count total for this tenant
	if err := r.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results with SKUs
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Preload("SKUs").
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

// ================== Platform Link Operations ==================

// CreatePlatformLink creates a new platform link
func (r *MasterProductRepository) CreatePlatformLink(ctx context.Context, link *models.MasterProductPlatformLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

// FindPlatformLinks returns all platform links for a master product
func (r *MasterProductRepository) FindPlatformLinks(ctx context.Context, masterProductID uint) ([]models.MasterProductPlatformLink, error) {
	var links []models.MasterProductPlatformLink
	err := r.db.WithContext(ctx).
		Where("master_product_id = ?", masterProductID).
		Order("platform ASC").
		Find(&links).Error
	return links, err
}

// FindPlatformLinkByPlatform finds a link for a specific platform
func (r *MasterProductRepository) FindPlatformLinkByPlatform(ctx context.Context, masterProductID uint, platform string) (*models.MasterProductPlatformLink, error) {
	var link models.MasterProductPlatformLink
	err := r.db.WithContext(ctx).
		Where("master_product_id = ? AND platform = ?", masterProductID, platform).
		First(&link).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// UpdateLinkStatus updates the sync status of a platform link
func (r *MasterProductRepository) UpdateLinkStatus(ctx context.Context, linkID uint, status string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.MasterProductPlatformLink{}).
		Where("id = ?", linkID).
		Updates(map[string]interface{}{
			"sync_status":    status,
			"last_synced_at": &now,
			"updated_at":     now,
		}).Error
}

// DeletePlatformLink removes a platform link
func (r *MasterProductRepository) DeletePlatformLink(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.MasterProductPlatformLink{}, id).Error
}

// UpsertPlatformLink creates or updates a platform link
func (r *MasterProductRepository) UpsertPlatformLink(ctx context.Context, link *models.MasterProductPlatformLink) error {
	return r.db.WithContext(ctx).
		Where("master_product_id = ? AND platform = ? AND COALESCE(master_sku_id, 0) = COALESCE(?, 0)",
			link.MasterProductID, link.Platform, link.MasterSkuID).
		Assign(*link).
		FirstOrCreate(link).Error
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
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}

// FindPlatformLinksByItemID finds platform links by platform item ID
// Used to check if a platform product is already imported
func (r *MasterProductRepository) FindPlatformLinksByItemID(ctx context.Context, tenantID, platform, platformItemID string) ([]models.MasterProductPlatformLink, error) {
	var links []models.MasterProductPlatformLink

	// Join with master_products to filter by tenant
	err := r.db.WithContext(ctx).
		Joins("JOIN master_products ON master_products.id = master_product_platform_links.master_product_id").
		Where("master_products.tenant_id = ? AND master_product_platform_links.platform = ? AND master_product_platform_links.platform_item_id = ?",
			tenantID, platform, platformItemID).
		Find(&links).Error

	return links, err
}
