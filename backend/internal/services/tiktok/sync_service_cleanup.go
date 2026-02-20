package tiktok

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

func (s *SyncService) clearTiktokProductCache(ctx context.Context) error {
	if s.tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.clearTiktokProductCacheWithDB(ctx, tx)
	})
}

func (s *SyncService) clearTiktokProductCacheWithDB(ctx context.Context, db *gorm.DB) error {
	if s.tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	productIDs := db.WithContext(ctx).
		Model(&models.TiktokProduct{}).
		Select("id").
		Where("tenant_id = ?", s.tenantID)

	if err := db.WithContext(ctx).
		Where("tiktok_product_id IN (?)", productIDs).
		Delete(&models.TiktokProductImage{}).Error; err != nil {
		return fmt.Errorf("clear tiktok product images: %w", err)
	}

	if err := db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Delete(&models.TiktokSku{}).Error; err != nil {
		return fmt.Errorf("clear tiktok skus: %w", err)
	}

	if err := db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Delete(&models.TiktokProduct{}).Error; err != nil {
		return fmt.Errorf("clear tiktok products: %w", err)
	}

	return nil
}
