package image

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/omni/backend/internal/models"

	"gorm.io/gorm"
)

// AddRef increments reference count
func (m *manager) AddRef(ctx context.Context, imageID uint) error {
	if imageID == 0 {
		return fmt.Errorf("invalid image_id")
	}

	result := m.db.WithContext(ctx).Model(&models.Image{}).
		Where("id = ?", imageID).
		UpdateColumn("ref_count", gorm.Expr("ref_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("failed to increment ref_count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("image not found: %d", imageID)
	}

	return nil
}

// RemoveRef decrements reference count
func (m *manager) RemoveRef(ctx context.Context, imageID uint) error {
	if imageID == 0 {
		return fmt.Errorf("invalid image_id")
	}

	result := m.db.WithContext(ctx).Model(&models.Image{}).
		Where("id = ?", imageID).
		UpdateColumn("ref_count", gorm.Expr("ref_count - 1"))

	if result.Error != nil {
		return fmt.Errorf("failed to decrement ref_count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("image not found: %d", imageID)
	}

	return nil
}

// CleanupOrphans deletes images with ref_count <= 0
func (m *manager) CleanupOrphans(ctx context.Context, tenantID string) (int, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}

	var orphans []models.Image
	masterProductImageSubquery := m.db.WithContext(ctx).Model(&models.MasterProductImage{}).Select("image_id")
	shopeeProductImageSubquery := m.db.WithContext(ctx).Model(&models.ShopeeProductImage{}).Select("image_id")
	tiktokProductImageSubquery := m.db.WithContext(ctx).Model(&models.TiktokProductImage{}).Select("image_id")
	lazadaProductImageSubquery := m.db.WithContext(ctx).Model(&models.LazadaProductImage{}).Select("image_id")

	if err := m.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where(
			`ref_count <= 0 OR (
				id NOT IN (?) AND
				id NOT IN (?) AND
				id NOT IN (?) AND
				id NOT IN (?)
			)`,
			masterProductImageSubquery,
			shopeeProductImageSubquery,
			tiktokProductImageSubquery,
			lazadaProductImageSubquery,
		).
		Find(&orphans).Error; err != nil {
		return 0, fmt.Errorf("failed to find orphan images: %w", err)
	}

	if len(orphans) == 0 {
		return 0, nil
	}

	// Delete files and DB records in transaction
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, img := range orphans {
			// Delete physical files
			basePath := filepath.Join(m.uploadPath, img.LocalPath)
			if err := os.RemoveAll(basePath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete files for image %d: %w", img.ID, err)
			}

			// Delete DB record
			if err := tx.Delete(&img).Error; err != nil {
				return fmt.Errorf("failed to delete image record %d: %w", img.ID, err)
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return len(orphans), nil
}
