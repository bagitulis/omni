package repositories

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// OrderRepository defines generic order data access interface
type OrderRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]interface{}, int64, error)
	FindByID(ctx context.Context, id string) (interface{}, error)
	Create(ctx context.Context, order interface{}) error
	Update(ctx context.Context, order interface{}) error
}

// ProductRepository defines generic product data access interface
type ProductRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]interface{}, int64, error)
	FindByID(ctx context.Context, id string) (interface{}, error)
	Create(ctx context.Context, product interface{}) error
	Update(ctx context.Context, product interface{}) error
}

// ShopeeOrderRepository handles Shopee order data access
type ShopeeOrderRepository struct {
	db *gorm.DB
}

// NewShopeeOrderRepository creates a new repository instance
func NewShopeeOrderRepository(db *gorm.DB) *ShopeeOrderRepository {
	return &ShopeeOrderRepository{db: db}
}

// FindAll returns paginated orders
func (r *ShopeeOrderRepository) FindAll(ctx context.Context, page, pageSize int) ([]models.ShopeeOrder, int64, error) {
	var orders []models.ShopeeOrder
	var total int64

	r.db.Model(&models.ShopeeOrder{}).Count(&total)

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
			val := *orders[i].ShipByDate
			// Only calculate if ship by date is in future
			if val > now {
				diff := val - now
				orders[i].Countdown = &diff
			} else {
				// If past due, countdown is 0 or negative
				diff := val - now
				orders[i].Countdown = &diff
			}
		}
	}

	return orders, total, err
}

// FindByOrderSN returns order by order serial number
func (r *ShopeeOrderRepository) FindByOrderSN(ctx context.Context, orderSN string) (*models.ShopeeOrder, error) {
	var order models.ShopeeOrder
	err := r.db.WithContext(ctx).Where("order_sn = ?", orderSN).First(&order).Error
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

// Create inserts a new order
func (r *ShopeeOrderRepository) Create(ctx context.Context, order *models.ShopeeOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// Update modifies an existing order
func (r *ShopeeOrderRepository) Update(ctx context.Context, order *models.ShopeeOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// Upsert creates or updates order by OrderSN
func (r *ShopeeOrderRepository) Upsert(ctx context.Context, order *models.ShopeeOrder) error {
	return r.db.WithContext(ctx).
		Where("order_sn = ?", order.OrderSN).
		Assign(*order).
		FirstOrCreate(order).Error
}
