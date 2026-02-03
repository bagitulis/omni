package shopee

import (
	"context"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

// OrderService handles Shopee order business logic
type OrderService struct {
	repo *repositories.ShopeeOrderRepository
}

// NewOrderService creates a new OrderService
func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{
		repo: repositories.NewShopeeOrderRepository(db),
	}
}

// GetOrders retrieves orders with pagination
func (s *OrderService) GetOrders(ctx context.Context, page, pageSize int) ([]models.ShopeeOrder, int64, error) {
	return s.repo.FindAll(ctx, page, pageSize)
}

// GetOrderBySN retrieves a single order by OrderSN
func (s *OrderService) GetOrderBySN(ctx context.Context, orderSN string) (*models.ShopeeOrder, error) {
	return s.repo.FindByOrderSN(ctx, orderSN)
}
