package products

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ProductStatus represents product status
type ProductStatus string

const (
	StatusNormal  ProductStatus = "NORMAL"
	StatusDeleted ProductStatus = "DELETED"
	StatusBanned  ProductStatus = "BANNED"
)

// Product represents a master product - uses models.Product for table naming
type Product struct {
	ID          string        `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID    string        `gorm:"index;not null;type:varchar(100)" json:"tenant_id"`
	Platform    string        `gorm:"index;not null;type:varchar(20)" json:"platform"`
	ItemID      string        `gorm:"index;not null;type:varchar(100)" json:"item_id"`
	SKU         string        `gorm:"index;type:varchar(100)" json:"sku"`
	Name        string        `gorm:"type:varchar(500)" json:"name"`
	Description string        `gorm:"type:text" json:"description,omitempty"`
	Status      ProductStatus `gorm:"type:varchar(20);default:NORMAL" json:"status"`
	Price       float64       `json:"price"`
	Stock       int           `json:"stock"`
	ImageURL    string        `gorm:"type:varchar(1000)" json:"image_url,omitempty"`
	CategoryID  string        `gorm:"type:varchar(100)" json:"category_id,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	SKUs        []ProductSKU  `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
}

// TableName returns the table name for GORM
func (Product) TableName() string {
	return models.GetTableName("Product")
}

// ProductSKU represents a product SKU/variation
type ProductSKU struct {
	ID            string  `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ProductID     string  `gorm:"index;not null;type:varchar(36)" json:"product_id"`
	TenantID      string  `gorm:"index;not null;type:varchar(100)" json:"tenant_id"`
	ModelID       string  `gorm:"index;type:varchar(100)" json:"model_id"`
	SKU           string  `gorm:"index;not null;type:varchar(100)" json:"sku"`
	Name          string  `gorm:"type:varchar(500)" json:"name"`
	VariationName string  `gorm:"type:varchar(500)" json:"variation_name,omitempty"`
	Price         float64 `json:"price"`
	Stock         int     `json:"stock"`
}

// TableName returns the table name for GORM
func (ProductSKU) TableName() string {
	return models.GetTableName("ProductSKU")
}

// ProductStats represents product statistics
type ProductStats struct {
	TotalProducts int                   `json:"total_products"`
	TotalSKUs     int                   `json:"total_skus"`
	TotalStock    int                   `json:"total_stock"`
	TotalValue    float64               `json:"total_value"`
	ByStatus      map[ProductStatus]int `json:"by_status"`
	ByPlatform    map[string]int        `json:"by_platform"`
}

// ProductQueryParams represents query parameters
type ProductQueryParams struct {
	Platform string        `json:"platform,omitempty"`
	Status   ProductStatus `json:"status,omitempty"`
	SKU      string        `json:"sku,omitempty"`
	Name     string        `json:"name,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	Offset   int           `json:"offset,omitempty"`
}

// ProductMasterService handles product master operations
type ProductMasterService struct {
	db *gorm.DB
}

// NewProductMasterService creates a new product master service
func NewProductMasterService(db *gorm.DB) *ProductMasterService {
	return &ProductMasterService{db: db}
}

// GetProducts retrieves products with optional filtering
func (s *ProductMasterService) GetProducts(
	ctx context.Context,
	tenantID string,
	params ProductQueryParams,
) ([]Product, int64, error) {
	var products []Product
	var total int64

	query := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Preload("SKUs")

	if params.Platform != "" {
		query = query.Where("platform = ?", params.Platform)
	}
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.SKU != "" {
		query = query.Where("sku LIKE ?", "%"+params.SKU+"%")
	}
	if params.Name != "" {
		query = query.Where("name LIKE ?", "%"+params.Name+"%")
	}

	// Get total count
	if err := query.Model(&Product{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	} else {
		query = query.Limit(100)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// GetProductStats returns product statistics
func (s *ProductMasterService) GetProductStats(
	ctx context.Context,
	tenantID string,
) (*ProductStats, error) {
	stats := &ProductStats{
		ByStatus:   make(map[ProductStatus]int),
		ByPlatform: make(map[string]int),
	}

	// Total products
	var productCount int64
	if err := s.db.WithContext(ctx).
		Model(&Product{}).
		Where("tenant_id = ?", tenantID).
		Count(&productCount).Error; err != nil {
		return nil, err
	}
	stats.TotalProducts = int(productCount)

	// Total SKUs
	var skuCount int64
	if err := s.db.WithContext(ctx).
		Model(&ProductSKU{}).
		Where("tenant_id = ?", tenantID).
		Count(&skuCount).Error; err != nil {
		return nil, err
	}
	stats.TotalSKUs = int(skuCount)

	// Total stock and value
	var stockSum struct {
		Stock int
		Value float64
	}
	s.db.WithContext(ctx).
		Model(&Product{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(stock), 0) as stock, COALESCE(SUM(price * stock), 0) as value").
		Scan(&stockSum)
	stats.TotalStock = stockSum.Stock
	stats.TotalValue = stockSum.Value

	// By status
	type statusCount struct {
		Status ProductStatus
		Count  int
	}
	var statusCounts []statusCount
	s.db.WithContext(ctx).
		Model(&Product{}).
		Where("tenant_id = ?", tenantID).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// By platform
	type platformCount struct {
		Platform string
		Count    int
	}
	var platformCounts []platformCount
	s.db.WithContext(ctx).
		Model(&Product{}).
		Where("tenant_id = ?", tenantID).
		Select("platform, COUNT(*) as count").
		Group("platform").
		Scan(&platformCounts)
	for _, pc := range platformCounts {
		stats.ByPlatform[pc.Platform] = pc.Count
	}

	return stats, nil
}

// SaveProducts saves or updates products
func (s *ProductMasterService) SaveProducts(
	ctx context.Context,
	tenantID string,
	products []Product,
) error {
	for i := range products {
		products[i].TenantID = tenantID
	}

	// Use upsert for each product
	for _, p := range products {
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND platform = ? AND item_id = ?",
				p.TenantID, p.Platform, p.ItemID).
			Assign(p).
			FirstOrCreate(&Product{}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// GetProductByItemID retrieves a product by item ID
func (s *ProductMasterService) GetProductByItemID(
	ctx context.Context,
	tenantID string,
	platform string,
	itemID string,
) (*Product, error) {
	var product Product

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND item_id = ?",
			tenantID, platform, itemID).
		Preload("SKUs").
		First(&product).Error

	if err != nil {
		return nil, err
	}

	return &product, nil
}
