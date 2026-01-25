package repositories

import (
	"context"

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

	return orders, total, err
}

// FindByOrderID returns order by TikTok order ID
func (r *TiktokOrderRepository) FindByOrderID(ctx context.Context, orderID string) (*models.TiktokOrder, error) {
	var order models.TiktokOrder
	err := r.db.WithContext(ctx).Where("order_sn = ?", orderID).First(&order).Error
	if err != nil {
		return nil, err
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
