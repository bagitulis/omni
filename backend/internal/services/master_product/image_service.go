package master_product

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/image"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ImageManager handles image processing for master products
type ImageManager struct {
	db           *gorm.DB
	dedupService *image.DedupService
	refService   *image.ReferenceService
}

// NewImageManager creates a new ImageManager
func NewImageManager(db *gorm.DB, dedup *image.DedupService, ref *image.ReferenceService) *ImageManager {
	return &ImageManager{
		db:           db,
		dedupService: dedup,
		refService:   ref,
	}
}

// ProcessAndLinkImages downloads, deduplicates, and links images to a product
// This is a high-level helper calling ProcessImages and CreateJoinTableEntries/UpdateJoinTableEntries
func (m *ImageManager) ProcessAndLinkImages(ctx context.Context, tenantID string, productID uint, imageURLs []string) error {
	if len(imageURLs) == 0 {
		return nil
	}

	images, err := m.ProcessImages(ctx, tenantID, imageURLs)
	if err != nil {
		return err
	}

	return m.UpdateJoinTableEntries(ctx, productID, images)
}

// ProcessImages downloads and deduplicates images from URLs
func (m *ImageManager) ProcessImages(ctx context.Context, tenantID string, imageURLs []string) ([]models.Image, error) {
	var processedImages []models.Image

	for _, url := range imageURLs {
		if url == "" {
			continue
		}

		// Download image (uses 30s timeout + 2 attempts with retry; see downloadImage)
		imageData, err := m.downloadImage(ctx, url)
		if err != nil {
			log.Warn().Err(err).Str("url", url).Msg("Failed to download image, skipping")
			continue
		}

		// Dedup/Create
		// Using "gallery" as default category for master product images
		img, _, err := m.dedupService.FindOrCreate(ctx, imageData, url, tenantID, "gallery")
		if err != nil {
			log.Error().Err(err).Str("url", url).Msg("Failed to process image")
			continue
		}

		processedImages = append(processedImages, *img)
	}

	return processedImages, nil
}

// CreateJoinTableEntries creates new entries in MasterProductImage table
func (m *ImageManager) CreateJoinTableEntries(ctx context.Context, productID uint, images []models.Image) error {
	if len(images) == 0 {
		return nil
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, img := range images {
			entry := models.MasterProductImage{
				ProductID: productID,
				ImageID:   uint(img.ID),
				SortOrder: i,
				Role:      "gallery", // Default role
			}

			// We use FirstOrCreate to avoid duplicates if something went wrong or partially succeeded
			if err := tx.Where(models.MasterProductImage{ProductID: productID, ImageID: uint(img.ID)}).
				Attrs(models.MasterProductImage{SortOrder: i, Role: "gallery"}).
				FirstOrCreate(&entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateJoinTableEntries updates the join table for a product
// It removes old entries (decrementing refs) and adds new ones
// Strategy:
// 1. Get all current linked ImageIDs
// 2. BulkDecrement(currentImageIDs) - release old refs
// 3. Delete all MasterProductImage rows for this product
// 4. Create new MasterProductImage rows for newImages
// Note: newImages already have RefCount incremented by ProcessImages/FindOrCreate
func (m *ImageManager) UpdateJoinTableEntries(ctx context.Context, productID uint, newImages []models.Image) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Get existing entries
		var existingEntries []models.MasterProductImage
		if err := tx.Where("product_id = ?", productID).Find(&existingEntries).Error; err != nil {
			return err
		}

		var existingImageIDs []uint
		for _, entry := range existingEntries {
			existingImageIDs = append(existingImageIDs, entry.ImageID)
		}

		// 2. Decrement existing refs and delete links
		if len(existingImageIDs) > 0 {
			if err := m.refService.BulkDecrement(ctx, existingImageIDs); err != nil {
				return err
			}

			// 3. Delete existing links
			if err := tx.Where("product_id = ?", productID).Delete(&models.MasterProductImage{}).Error; err != nil {
				return err
			}
		}

		// 4. Create new links
		for i, img := range newImages {
			entry := models.MasterProductImage{
				ProductID: productID,
				ImageID:   uint(img.ID),
				SortOrder: i,
				Role:      "gallery",
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteJoinTableEntries removes all join table entries for a product and updates ref counts
func (m *ImageManager) DeleteJoinTableEntries(ctx context.Context, productID uint) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Get existing entries to know what to decrement
		var existingEntries []models.MasterProductImage
		if err := tx.Where("product_id = ?", productID).Find(&existingEntries).Error; err != nil {
			return err
		}

		if len(existingEntries) == 0 {
			return nil
		}

		var imageIDs []uint
		for _, entry := range existingEntries {
			imageIDs = append(imageIDs, entry.ImageID)
		}

		// 2. Delete entries
		if err := tx.Where("product_id = ?", productID).Delete(&models.MasterProductImage{}).Error; err != nil {
			return err
		}

		// 3. Decrement refs
		return m.refService.BulkDecrement(ctx, imageIDs)
	})
}

func (m *ImageManager) downloadImage(ctx context.Context, url string) ([]byte, error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(2 * time.Second):
			}
		}

		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			log.Warn().Err(err).Str("url", url).Int("attempt", attempt+1).Msg("Image download failed, retrying")
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
			log.Warn().Int("status", resp.StatusCode).Str("url", url).Int("attempt", attempt+1).Msg("Image download non-OK status, retrying")
			continue
		}

		return io.ReadAll(resp.Body)
	}

	return nil, lastErr
}
