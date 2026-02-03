package shopee

import (
	"context"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

// ProductService handles Shopee product business logic
type ProductService struct {
	repo *repositories.ShopeeProductRepository
}

// NewProductService creates a new ProductService
func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{
		repo: repositories.NewShopeeProductRepository(db),
	}
}

// GetProductByItemID retrieves a single product by ItemID
func (s *ProductService) GetProductByItemID(ctx context.Context, itemID int64) (*models.ShopeeProduct, error) {
	return s.repo.FindByItemID(ctx, itemID)
}
