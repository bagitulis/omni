package products

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ProductDetail represents detailed product information
type ProductDetail struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	TenantID         string    `gorm:"index;not null" json:"tenantId"`
	Platform         string    `gorm:"index;not null" json:"platform"`
	ItemID           string    `gorm:"index;not null" json:"itemId"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	LongDescription  string    `json:"longDescription,omitempty"`
	Price            float64   `json:"price"`
	OriginalPrice    float64   `json:"originalPrice,omitempty"`
	Currency         string    `json:"currency"`
	Stock            int       `json:"stock"`
	ReservedStock    int       `json:"reservedStock"`
	Sales            int       `json:"sales"`
	Views            int       `json:"views"`
	Likes            int       `json:"likes"`
	Rating           float64   `json:"rating"`
	RatingCount      int       `json:"ratingCount"`
	CategoryID       string    `json:"categoryId"`
	CategoryName     string    `json:"categoryName"`
	Brand            string    `json:"brand,omitempty"`
	Weight           float64   `json:"weight"`
	WeightUnit       string    `json:"weightUnit"`
	Dimensions       string    `json:"dimensions,omitempty"`
	Condition        string    `json:"condition"`
	PreOrder         bool      `json:"preOrder"`
	DaysToShip       int       `json:"daysToShip"`
	Status           string    `json:"status"`
	Images           string    `json:"images"` // JSON array
	Videos           string    `json:"videos,omitempty"` // JSON array
	Attributes       string    `json:"attributes,omitempty"` // JSON
	LogisticsInfo    string    `json:"logisticsInfo,omitempty"` // JSON
	WholesaleInfo    string    `json:"wholesaleInfo,omitempty"` // JSON
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	LastSyncAt       time.Time `json:"lastSyncAt"`
}

// TableName returns the table name for GORM
func (ProductDetail) TableName() string {
	return "product_details"
}

// ProductDetailService handles product detail operations
type ProductDetailService struct {
	db *gorm.DB
}

// NewProductDetailService creates a new product detail service
func NewProductDetailService(db *gorm.DB) *ProductDetailService {
	return &ProductDetailService{db: db}
}

// GetProductDetail retrieves product detail by item ID
func (s *ProductDetailService) GetProductDetail(
	ctx context.Context,
	tenantID string,
	platform string,
	itemID string,
) (*ProductDetail, error) {
	var detail ProductDetail

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND item_id = ?",
			tenantID, platform, itemID).
		First(&detail).Error

	if err != nil {
		return nil, err
	}

	return &detail, nil
}

// SaveProductDetail saves or updates product detail
func (s *ProductDetailService) SaveProductDetail(
	ctx context.Context,
	tenantID string,
	detail *ProductDetail,
) error {
	detail.TenantID = tenantID
	detail.LastSyncAt = time.Now()

	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND item_id = ?",
			detail.TenantID, detail.Platform, detail.ItemID).
		Assign(detail).
		FirstOrCreate(&ProductDetail{}).Error
}

// GetProductDetailsBatch retrieves multiple product details
func (s *ProductDetailService) GetProductDetailsBatch(
	ctx context.Context,
	tenantID string,
	platform string,
	itemIDs []string,
) ([]ProductDetail, error) {
	var details []ProductDetail

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND item_id IN ?",
			tenantID, platform, itemIDs).
		Find(&details).Error

	return details, err
}

// GetStaleProducts returns products that need re-sync
func (s *ProductDetailService) GetStaleProducts(
	ctx context.Context,
	tenantID string,
	platform string,
	maxAge time.Duration,
) ([]ProductDetail, error) {
	var details []ProductDetail
	cutoff := time.Now().Add(-maxAge)

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND last_sync_at < ?",
			tenantID, platform, cutoff).
		Find(&details).Error

	return details, err
}
