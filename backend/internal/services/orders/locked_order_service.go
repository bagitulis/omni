package orders

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LockedOrderItem represents a locked order item
// NOTE: JSON tags use snake_case for frontend compatibility
// NOTE: ID is VARCHAR(255) in PostgreSQL, uses UUID format
type LockedOrderItem struct {
	ID            string `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID      string `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	SKU           string `gorm:"column:sku;not null" json:"sku"`
	ProductName   string `gorm:"column:product_name;not null" json:"product_name"`
	VariationName string `gorm:"column:variation_name" json:"variation_name,omitempty"`
	Qty           int    `gorm:"column:qty;not null" json:"qty"`
}

// TableName returns the table name for GORM
func (LockedOrderItem) TableName() string {
	return "locked_orders"
}

// LockedOrderService handles locked order operations
type LockedOrderService struct {
	db *gorm.DB
}

// NewLockedOrderService creates a new locked order service
func NewLockedOrderService(db *gorm.DB) *LockedOrderService {
	return &LockedOrderService{db: db}
}

// SaveLockedOrders saves locked orders for a tenant (replaces existing)
func (s *LockedOrderService) SaveLockedOrders(
	ctx context.Context,
	tenantID string,
	items []LockedOrderItem,
) (int, error) {
	// Delete existing locked orders for this tenant
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Delete(&LockedOrderItem{}).Error; err != nil {
		return 0, err
	}

	if len(items) == 0 {
		return 0, nil
	}

	// Set tenant ID and generate UUID for each item
	for i := range items {
		items[i].ID = uuid.New().String()
		items[i].TenantID = tenantID
	}

	if err := s.db.WithContext(ctx).Create(&items).Error; err != nil {
		return 0, err
	}

	return len(items), nil
}

// GetLockedOrders retrieves locked orders for a tenant
func (s *LockedOrderService) GetLockedOrders(
	ctx context.Context,
	tenantID string,
) ([]LockedOrderItem, error) {
	var orders []LockedOrderItem

	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("qty DESC").
		Find(&orders).Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

// GetTotalQty returns the total quantity of locked orders
func (s *LockedOrderService) GetTotalQty(
	ctx context.Context,
	tenantID string,
) (int64, error) {
	var total int64

	err := s.db.WithContext(ctx).
		Model(&LockedOrderItem{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(qty), 0)").
		Scan(&total).Error

	return total, err
}

// ClearLockedOrders clears all locked orders for a tenant
func (s *LockedOrderService) ClearLockedOrders(
	ctx context.Context,
	tenantID string,
) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Delete(&LockedOrderItem{}).Error
}
