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
		if err := tx.Where("tenant_id = ?", s.tenantID).Delete(&models.TiktokSku{}).Error; err != nil {
			return fmt.Errorf("clear tiktok skus: %w", err)
		}

		if err := tx.Where("tenant_id = ?", s.tenantID).Delete(&models.TiktokProduct{}).Error; err != nil {
			return fmt.Errorf("clear tiktok products: %w", err)
		}

		return nil
	})
}
