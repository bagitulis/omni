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
		if err := tx.Where("tenant_id = ?", s.tenantID).Delete(&models.ShopeeSku{}).Error; err != nil {
			return fmt.Errorf("clear shopee skus: %w", err)
		}

		if err := tx.Where("tenant_id = ?", s.tenantID).Delete(&models.ShopeeProduct{}).Error; err != nil {
			return fmt.Errorf("clear shopee products: %w", err)
		}

		return nil
	})
}
