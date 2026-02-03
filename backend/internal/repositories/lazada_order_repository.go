package repositories

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// LazadaOrderRepository handles Lazada order data access
type LazadaOrderRepository struct {
	db *gorm.DB
}

// NewLazadaOrderRepository creates a new repository instance
func NewLazadaOrderRepository(db *gorm.DB) *LazadaOrderRepository {
	return &LazadaOrderRepository{db: db}
}

// FindAll returns paginated orders
func (r *LazadaOrderRepository) FindAll(ctx context.Context, page, pageSize int) ([]models.LazadaOrder, int64, error) {
	var orders []models.LazadaOrder
	var total int64

	r.db.Model(&models.LazadaOrder{}).Count(&total)

	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error

	// Calculate countdown
	now := time.Now().Unix()
	for i := range orders {
		if orders[i].ShipByDate != nil {
			diff := *orders[i].ShipByDate - now
			orders[i].Countdown = &diff
		}
	}

	return orders, total, err
}

// FindByOrderID returns order by Lazada order ID
func (r *LazadaOrderRepository) FindByOrderID(ctx context.Context, orderID string) (*models.LazadaOrder, error) {
	var order models.LazadaOrder
	err := r.db.WithContext(ctx).Where("order_sn = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
	}

	// Calculate countdown
	if order.ShipByDate != nil {
		now := time.Now().Unix()
		diff := *order.ShipByDate - now
		order.Countdown = &diff
	}

	return &order, nil
}

// Upsert creates or updates order
func (r *LazadaOrderRepository) Upsert(ctx context.Context, order *models.LazadaOrder) error {
	return r.db.WithContext(ctx).
		Where("order_sn = ?", order.OrderSN).
		Assign(*order).
		FirstOrCreate(order).Error
}
