// Package master_product provides import service for Master Product
package master_product

import (
	"errors"

	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

// Import errors
var (
	ErrShopeeProductNotFound = errors.New("shopee product not found")
	ErrNoSellerSku           = errors.New("product has no seller_sku - cannot import")
	ErrProductAlreadyExists  = errors.New("master product already exists for this shopee item")
	ErrShopeeNotConfigured   = errors.New("shopee credentials not configured")
)

// ImportService handles importing products from platforms to Master Product
type ImportService struct {
	db       *gorm.DB
	repo     *repositories.MasterProductRepository
	basePath string
	imgMgr   *ImageManager // optional: downloads and links images during import
}

// NewImportService creates a new import service
func NewImportService(db *gorm.DB, basePath string) *ImportService {
	return &ImportService{
		db:       db,
		repo:     repositories.NewMasterProductRepository(db),
		basePath: basePath,
		imgMgr:   newDefaultImageManager(db),
	}
}
