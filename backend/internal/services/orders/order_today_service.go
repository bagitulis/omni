package orders

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

var orderTodayLogger = logger.Named("OrderTodayService")

// OrderTodayItem represents an order item for "Order Today" feature
// This is for API responses - maps to order_today_items table
// NOTE: JSON uses camelCase to match Node.js backend for frontend compatibility
type OrderTodayItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;index;not null" json:"tenantId"`
	Platform      string    `gorm:"column:platform;index;not null" json:"platform"`
	OrderSN       string    `gorm:"column:order_sn;index;not null" json:"orderSn"`
	TrackingNo    string    `gorm:"column:tracking_no" json:"trackingNo"`
	Courier       string    `gorm:"column:courier" json:"courier"`
	SellerSku     string    `gorm:"column:seller_sku" json:"sellerSku"`
	ProductName   string    `gorm:"column:product_name" json:"productName"`
	VariationName string    `gorm:"column:variation_name" json:"variationName"`
	Quantity      int       `gorm:"column:quantity;default:1" json:"quantity"`
	SyncedAt      time.Time `gorm:"column:synced_at;index" json:"syncedAt"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName returns the table name for GORM
func (OrderTodayItem) TableName() string {
	return models.GetTableName("OrderTodayItem")
}

// OrderTodayService handles "Order Today" operations
type OrderTodayService struct {
	db *gorm.DB
}

// NewOrderTodayService creates a new order today service
func NewOrderTodayService(db *gorm.DB) *OrderTodayService {
	return &OrderTodayService{db: db}
}

// OrderTodayData is the input data for saving order today items
type OrderTodayData struct {
	Platform      string
	OrderSN       string
	TrackingNo    string
	Courier       string
	SellerSku     string
	ProductName   string
	VariationName string
	Quantity      int
}

// SaveOrderTodayItems saves order today items (replaces existing for tenant)
// Deduplicates by platform+order_sn+seller_sku before saving
func (s *OrderTodayService) SaveOrderTodayItems(
	ctx context.Context,
	tenantID string,
	items []OrderTodayData,
) (int, error) {
	// Delete existing items for this tenant first
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Delete(&OrderTodayItem{}).Error; err != nil {
		return 0, err
	}

	if len(items) == 0 {
		return 0, nil
	}

	now := time.Now()

	// Deduplicate items by platform+order_sn+seller_sku
	// Use a map to aggregate quantities for duplicates
	seen := make(map[string]*OrderTodayItem)
	for _, item := range items {
		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}

		// Create unique key
		key := item.Platform + "|" + item.OrderSN + "|" + item.SellerSku

		if existing, ok := seen[key]; ok {
			// Aggregate quantity for duplicates
			existing.Quantity += qty
		} else {
			seen[key] = &OrderTodayItem{
				TenantID:      tenantID,
				Platform:      item.Platform,
				OrderSN:       item.OrderSN,
				TrackingNo:    item.TrackingNo,
				Courier:       item.Courier,
				SellerSku:     item.SellerSku,
				ProductName:   item.ProductName,
				VariationName: item.VariationName,
				Quantity:      qty,
				SyncedAt:      now,
				CreatedAt:     now,
				UpdatedAt:     now,
			}
		}
	}

	// Convert map to slice
	dbItems := make([]OrderTodayItem, 0, len(seen))
	for _, item := range seen {
		dbItems = append(dbItems, *item)
	}

	// Insert in batches to avoid issues with large datasets
	batchSize := 100
	for i := 0; i < len(dbItems); i += batchSize {
		end := i + batchSize
		if end > len(dbItems) {
			end = len(dbItems)
		}
		batch := dbItems[i:end]

		if err := s.db.WithContext(ctx).Create(&batch).Error; err != nil {
			orderTodayLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
				"batch": i / batchSize,
				"error": err.Error(),
			}).Warn("Failed to save batch, continuing...")
			// Continue with next batch instead of failing completely
		}
	}

	orderTodayLogger.WithTenantID(tenantID).WithFields(map[string]interface{}{
		"original_count":     len(items),
		"deduplicated_count": len(dbItems),
	}).Info("Saved order today items")

	return len(dbItems), nil
}

// GetOrderTodayItems retrieves order today items for a tenant
func (s *OrderTodayService) GetOrderTodayItems(
	ctx context.Context,
	tenantID string,
) ([]OrderTodayItem, error) {
	var items []OrderTodayItem

	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("platform, order_sn").
		Find(&items).Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

// GetOrderTodayCount returns the count of order today items
func (s *OrderTodayService) GetOrderTodayCount(
	ctx context.Context,
	tenantID string,
) (int64, error) {
	var count int64

	err := s.db.WithContext(ctx).
		Model(&OrderTodayItem{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error

	return count, err
}

// GetPlatformCounts returns counts by platform
func (s *OrderTodayService) GetPlatformCounts(
	ctx context.Context,
	tenantID string,
) (map[string]int64, error) {
	type PlatformCount struct {
		Platform string
		Count    int64
	}

	var results []PlatformCount
	err := s.db.WithContext(ctx).
		Model(&OrderTodayItem{}).
		Select("platform, COUNT(DISTINCT order_sn) as count").
		Where("tenant_id = ?", tenantID).
		Group("platform").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Platform] = r.Count
	}

	return counts, nil
}

// ClearOrderTodayItems clears all order today items for a tenant
func (s *OrderTodayService) ClearOrderTodayItems(
	ctx context.Context,
	tenantID string,
) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Delete(&OrderTodayItem{}).Error
}
