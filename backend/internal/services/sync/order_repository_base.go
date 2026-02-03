package sync

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// GormOrderRepository implements OrderRepository using GORM
// SRP: Coordinates platform-specific repositories
type GormOrderRepository struct {
	tenantID string
}

// NewGormOrderRepository creates a new order repository
func NewGormOrderRepository(tenantID string) *GormOrderRepository {
	return &GormOrderRepository{
		tenantID: tenantID,
	}
}

// getDB returns tenant database connection
func (r *GormOrderRepository) getDB() (*gorm.DB, error) {
	return config.GetTenantDBByID(r.tenantID)
}

// SaveOrders saves orders to the database (delegates to platform-specific methods)
func (r *GormOrderRepository) SaveOrders(ctx context.Context, platform PlatformType, orders []Order) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	switch platform {
	case PlatformShopee:
		return r.saveShopeeOrders(ctx, db, orders)
	case PlatformLazada:
		return r.saveLazadaOrders(ctx, db, orders)
	case PlatformTiktok:
		return r.saveTiktokOrders(ctx, db, orders)
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}
}

// GetOrdersByStatus gets orders by status (delegates to platform-specific methods)
func (r *GormOrderRepository) GetOrdersByStatus(ctx context.Context, platform PlatformType, status string, limit int) ([]Order, error) {
	db, err := r.getDB()
	if err != nil {
		return nil, err
	}

	switch platform {
	case PlatformShopee:
		return r.getShopeeOrdersByStatus(ctx, db, status, limit)
	case PlatformLazada:
		return r.getLazadaOrdersByStatus(ctx, db, status, limit)
	case PlatformTiktok:
		return r.getTiktokOrdersByStatus(ctx, db, status, limit)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// ClearOrdersByStatus deletes orders and their items by status
// IMPORTANT: Also clears order items to prevent orphan data
func (r *GormOrderRepository) ClearOrdersByStatus(ctx context.Context, platform PlatformType, status string) error {
	db, err := r.getDB()
	if err != nil {
		return err
	}

	switch platform {
	case PlatformShopee:
		return r.clearShopeeOrdersByStatus(ctx, db, status)
	case PlatformLazada:
		return r.clearLazadaOrdersByStatus(ctx, db, status)
	case PlatformTiktok:
		return r.clearTiktokOrdersByStatus(ctx, db, status)
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}
}

// clearShopeeOrdersByStatus clears Shopee orders and their items
func (r *GormOrderRepository) clearShopeeOrdersByStatus(ctx context.Context, db *gorm.DB, status string) error {
	// First, get order SNs to delete their items
	var orderSNs []string
	if err := db.WithContext(ctx).
		Model(&models.ShopeeOrder{}).
		Where("order_status = ?", status).
		Pluck("order_sn", &orderSNs).Error; err != nil {
		return err
	}

	// Delete order items first (if any orders exist)
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).
			Where("order_sn IN ?", orderSNs).
			Delete(&models.ShopeeOrderItem{}).Error; err != nil {
			return err
		}
	}

	// Then delete orders
	return db.WithContext(ctx).Where("order_status = ?", status).Delete(&models.ShopeeOrder{}).Error
}

// clearLazadaOrdersByStatus clears Lazada orders and their items
func (r *GormOrderRepository) clearLazadaOrdersByStatus(ctx context.Context, db *gorm.DB, status string) error {
	// First, get order SNs to delete their items
	var orderSNs []string
	if err := db.WithContext(ctx).
		Model(&models.LazadaOrder{}).
		Where("order_status = ?", status).
		Pluck("order_sn", &orderSNs).Error; err != nil {
		return err
	}

	// Delete order items first (if any orders exist)
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).
			Where("order_sn IN ?", orderSNs).
			Delete(&models.LazadaOrderItem{}).Error; err != nil {
			return err
		}
	}

	// Then delete orders
	return db.WithContext(ctx).Where("order_status = ?", status).Delete(&models.LazadaOrder{}).Error
}

// clearTiktokOrdersByStatus clears TikTok orders and their items
func (r *GormOrderRepository) clearTiktokOrdersByStatus(ctx context.Context, db *gorm.DB, status string) error {
	// First, get order SNs to delete their items
	var orderSNs []string
	if err := db.WithContext(ctx).
		Model(&models.TiktokOrder{}).
		Where("order_status = ?", status).
		Pluck("order_sn", &orderSNs).Error; err != nil {
		return err
	}

	// Delete order items first (if any orders exist)
	if len(orderSNs) > 0 {
		if err := db.WithContext(ctx).
			Where("order_sn IN ?", orderSNs).
			Delete(&models.TiktokOrderItem{}).Error; err != nil {
			return err
		}
	}

	// Then delete orders
	return db.WithContext(ctx).Where("order_status = ?", status).Delete(&models.TiktokOrder{}).Error
}

// GetOrdersCount returns count of orders by status
func (r *GormOrderRepository) GetOrdersCount(ctx context.Context, platform PlatformType, status string) (int64, error) {
	db, err := r.getDB()
	if err != nil {
		return 0, err
	}

	var count int64
	var result *gorm.DB

	switch platform {
	case PlatformShopee:
		query := db.WithContext(ctx).Model(&models.ShopeeOrder{})
		if status != "" {
			query = query.Where("order_status = ?", status)
		}
		result = query.Count(&count)
	case PlatformLazada:
		query := db.WithContext(ctx).Model(&models.LazadaOrder{})
		if status != "" {
			query = query.Where("order_status = ?", status)
		}
		result = query.Count(&count)
	case PlatformTiktok:
		query := db.WithContext(ctx).Model(&models.TiktokOrder{})
		if status != "" {
			query = query.Where("order_status = ?", status)
		}
		result = query.Count(&count)
	default:
		return 0, fmt.Errorf("unsupported platform: %s", platform)
	}

	return count, result.Error
}

func (r *GormOrderRepository) getMasterProductImagesByItemIDs(
	ctx context.Context,
	db *gorm.DB,
	platform string,
	itemIDs []string,
) map[string]string {
	result := make(map[string]string)
	if len(itemIDs) == 0 {
		return result
	}

	linkTable := models.GetTableName("MasterProductPlatformLink")
	productTable := models.GetTableName("MasterProduct")

	type masterProductImageRow struct {
		PlatformItemID string           `gorm:"column:platform_item_id"`
		Images         models.JSONArray `gorm:"column:images"`
	}

	var rows []masterProductImageRow
	err := db.WithContext(ctx).
		Table(linkTable+" links").
		Select("links.platform_item_id, products.images").
		Joins("JOIN "+productTable+" products ON products.id = links.master_product_id").
		Where("links.platform = ?", platform).
		Where("links.platform_item_id IN ?", itemIDs).
		Find(&rows).Error
	if err != nil {
		return result
	}

	for _, row := range rows {
		image := firstImageFromJSONArray(row.Images)
		if image != "" {
			result[row.PlatformItemID] = image
		}
	}

	return result
}

func firstImageFromJSONArray(images models.JSONArray) string {
	for _, entry := range images {
		if value, ok := entry.(string); ok && value != "" {
			return value
		}
	}
	return ""
}
