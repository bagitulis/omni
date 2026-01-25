package orders

import (
	"context"

	"gorm.io/gorm"
)

// LockedOrderItem represents a locked order item
type LockedOrderItem struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	TenantID      string `gorm:"index;not null" json:"tenantId"`
	SKU           string `gorm:"not null" json:"sku"`
	ProductName   string `gorm:"not null" json:"productName"`
	VariationName string `json:"variationName,omitempty"`
	Qty           int    `gorm:"not null" json:"qty"`
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

	// Set tenant ID and create
	for i := range items {
		items[i].TenantID = tenantID
		items[i].ID = 0 // Reset ID for new records
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
