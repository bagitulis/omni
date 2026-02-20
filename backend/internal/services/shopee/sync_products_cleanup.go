package shopee

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

func (s *ProductSyncService) clearShopeeProductCache(ctx context.Context) error {
	if s.tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.clearShopeeProductCacheWithDB(ctx, tx)
	})
}

func (s *ProductSyncService) clearShopeeProductCacheWithDB(ctx context.Context, db *gorm.DB) error {
	if s.tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	productIDs := db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Select("id").
		Where("tenant_id = ?", s.tenantID)

	if err := db.WithContext(ctx).
		Where("shopee_product_id IN (?)", productIDs).
		Delete(&models.ShopeeProductImage{}).Error; err != nil {
		return fmt.Errorf("clear shopee product images: %w", err)
	}

	if err := db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Delete(&models.ShopeeSku{}).Error; err != nil {
		return fmt.Errorf("clear shopee skus: %w", err)
	}

	if err := db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Delete(&models.ShopeeProduct{}).Error; err != nil {
		return fmt.Errorf("clear shopee products: %w", err)
	}

	return nil
}
