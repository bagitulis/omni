package lazada

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

func (s *SyncService) clearLazadaProductCache(ctx context.Context) error {
	if s.tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.clearLazadaProductCacheWithDB(ctx, tx)
	})
}

func (s *SyncService) clearLazadaProductCacheWithDB(ctx context.Context, db *gorm.DB) error {
	if s.tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	productIDs := db.WithContext(ctx).
		Model(&models.LazadaProduct{}).
		Select("id").
		Where("tenant_id = ?", s.tenantID)

	if err := db.WithContext(ctx).
		Where("lazada_product_id IN (?)", productIDs).
		Delete(&models.LazadaProductImage{}).Error; err != nil {
		return fmt.Errorf("clear lazada product images: %w", err)
	}

	if err := db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Delete(&models.LazadaSku{}).Error; err != nil {
		return fmt.Errorf("clear lazada skus: %w", err)
	}

	if err := db.WithContext(ctx).Where("tenant_id = ?", s.tenantID).Delete(&models.LazadaProduct{}).Error; err != nil {
		return fmt.Errorf("clear lazada products: %w", err)
	}

	return nil
}
