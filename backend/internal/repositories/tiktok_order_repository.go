package repositories

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// TiktokOrderRepository handles TikTok order data access
type TiktokOrderRepository struct {
	db *gorm.DB
}

// NewTiktokOrderRepository creates a new repository instance
func NewTiktokOrderRepository(db *gorm.DB) *TiktokOrderRepository {
	return &TiktokOrderRepository{db: db}
}

// FindAll returns paginated orders
func (r *TiktokOrderRepository) FindAll(ctx context.Context, page, pageSize int) ([]models.TiktokOrder, int64, error) {
	var orders []models.TiktokOrder
	var total int64

	r.db.Model(&models.TiktokOrder{}).Count(&total)

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

// FindByOrderID returns order by TikTok order ID
func (r *TiktokOrderRepository) FindByOrderID(ctx context.Context, orderID string) (*models.TiktokOrder, error) {
	var order models.TiktokOrder
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
func (r *TiktokOrderRepository) Upsert(ctx context.Context, order *models.TiktokOrder) error {
	return r.db.WithContext(ctx).
		Where("order_sn = ?", order.OrderSN).
		Assign(*order).
		FirstOrCreate(order).Error
}
