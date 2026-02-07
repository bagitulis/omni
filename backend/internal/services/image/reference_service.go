package image

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ReferenceService manages image reference counting
type ReferenceService struct {
	db *gorm.DB
}

// NewReferenceService creates a new reference service
func NewReferenceService(db *gorm.DB) *ReferenceService {
	return &ReferenceService{
		db: db,
	}
}

// IncrementRef atomically increments ref_count for an image
// Thread-safe using database transaction
func (s *ReferenceService) IncrementRef(ctx context.Context, imageID int64) error {
	// Atomic increment
	err := s.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("id = ?", imageID).
		UpdateColumn("ref_count", gorm.Expr("ref_count + ?", 1)).
		Error

	if err != nil {
		return fmt.Errorf("failed to increment ref count for image %d: %w", imageID, err)
	}

	return nil
}

// DecrementRef atomically decrements ref_count for an image
// Thread-safe using database transaction
func (s *ReferenceService) DecrementRef(ctx context.Context, imageID int64) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Atomic decrement
		if err := tx.Model(&models.Image{}).
			Where("id = ?", imageID).
			UpdateColumn("ref_count", gorm.Expr("ref_count - ?", 1)).
			Error; err != nil {
			return fmt.Errorf("failed to decrement ref count for image %d: %w", imageID, err)
		}

		// Check current ref_count
		var img models.Image
		if err := tx.Select("ref_count").First(&img, imageID).Error; err != nil {
			return fmt.Errorf("failed to get ref count for image %d: %w", imageID, err)
		}

		// If ref_count <= 0, delete the image record
		if img.RefCount <= 0 {
			if err := tx.Delete(&models.Image{}, imageID).Error; err != nil {
				return fmt.Errorf("failed to delete image %d: %w", imageID, err)
			}
		}

		return nil
	})
}

// GetRefCount returns current reference count for an image
func (s *ReferenceService) GetRefCount(ctx context.Context, imageID int64) (int, error) {
	var img models.Image
	err := s.db.WithContext(ctx).
		Select("ref_count").
		First(&img, imageID).
		Error

	if err != nil {
		return 0, fmt.Errorf("failed to get ref count for image %d: %w", imageID, err)
	}

	return img.RefCount, nil
}

// BulkDecrement decrements ref_count for multiple images
// Used when deleting a product with multiple images
func (s *ReferenceService) BulkDecrement(ctx context.Context, imageIDs []int64) error {
	if len(imageIDs) == 0 {
		return nil
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, id := range imageIDs {
			// Atomic decrement
			if err := tx.Model(&models.Image{}).
				Where("id = ?", id).
				UpdateColumn("ref_count", gorm.Expr("ref_count - ?", 1)).
				Error; err != nil {
				return fmt.Errorf("failed to decrement ref count for image %d: %w", id, err)
			}

			// Check current ref_count
			var img models.Image
			if err := tx.Select("ref_count").First(&img, id).Error; err != nil {
				return fmt.Errorf("failed to get ref count for image %d: %w", id, err)
			}

			// If ref_count <= 0, soft delete
			if img.RefCount <= 0 {
				now := time.Now()
				if err := tx.Model(&models.Image{}).
					Where("id = ?", id).
					UpdateColumn("deleted_at", now).
					Error; err != nil {
					return fmt.Errorf("failed to soft delete image %d: %w", id, err)
				}
			}
		}
		return nil
	})
}
