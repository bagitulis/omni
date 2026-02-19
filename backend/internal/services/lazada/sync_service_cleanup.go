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

		if err := tx.Where("tenant_id = ?", s.tenantID).Delete(&models.LazadaSku{}).Error; err != nil {
			return fmt.Errorf("clear lazada skus: %w", err)
		}

		if err := tx.Where("tenant_id = ?", s.tenantID).Delete(&models.LazadaProduct{}).Error; err != nil {
			return fmt.Errorf("clear lazada products: %w", err)
		}

		return nil
	})
}
