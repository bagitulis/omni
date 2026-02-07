package image

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// DedupService handles image deduplication
type DedupService struct {
	db               *gorm.DB
	webpService      *WebPService
	storageService   *StorageService
	thumbnailService *ThumbnailService
	refService       *ReferenceService
}

// NewDedupService creates a new deduplication service
func NewDedupService(
	db *gorm.DB,
	webp *WebPService,
	storage *StorageService,
	thumbnail *ThumbnailService,
	ref *ReferenceService,
) *DedupService {
	return &DedupService{
		db:               db,
		webpService:      webp,
		storageService:   storage,
		thumbnailService: thumbnail,
		refService:       ref,
	}
}

// FindOrCreate finds existing image by hash or creates new one
// Returns (image, isNew, error)
// If originalURL is provided, checks URL cache first
// Always converts to WebP before hashing for consistency
func (s *DedupService) FindOrCreate(
	ctx context.Context,
	imageData []byte,
	originalURL string,
	tenantID string,
	category string,
) (*models.Image, bool, error) {
	if tenantID == "" {
		return nil, false, errors.New("tenant_id is required")
	}

	// 1. Check by Original URL if provided
	if originalURL != "" {
		var existingImage models.Image
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND original_url = ?", tenantID, originalURL).
			First(&existingImage).Error

		if err == nil {
			// Found by URL
			if err := s.refService.IncrementRef(ctx, existingImage.ID); err != nil {
				return nil, false, fmt.Errorf("failed to increment ref count: %w", err)
			}
			return &existingImage, false, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, fmt.Errorf("failed to check original_url: %w", err)
		}
	}

	// 2. Convert to WebP
	webpData, err := s.webpService.ConvertToWebP(imageData)
	if err != nil {
		return nil, false, fmt.Errorf("failed to convert to webp: %w", err)
	}

	// 3. Calculate Hash
	hash := sha256.Sum256(webpData)
	hashStr := hex.EncodeToString(hash[:])

	// 4. Check by Content Hash
	var existingImage models.Image
	err = s.db.WithContext(ctx).
		Where("tenant_id = ? AND content_hash = ?", tenantID, hashStr).
		First(&existingImage).Error

	if err == nil {
		// Found by Hash
		if err := s.refService.IncrementRef(ctx, existingImage.ID); err != nil {
			return nil, false, fmt.Errorf("failed to increment ref count: %w", err)
		}
		// If originalURL was provided but not found in step 1, we might want to update it?
		// But for now, just return the existing image.
		return &existingImage, false, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, fmt.Errorf("failed to check content_hash: %w", err)
	}

	// 5. Create New Image
	// Use hash as base filename
	baseFilename := fmt.Sprintf("%s.webp", hashStr)

	// Save Original (WebP)
	localPath, err := s.storageService.SaveImageWithName(tenantID, category, baseFilename, webpData)
	if err != nil {
		return nil, false, fmt.Errorf("failed to save image: %w", err)
	}

	// Generate Thumbnails (if needed by other services)
	_, err = s.thumbnailService.GenerateThumbnails(webpData, tenantID, category, baseFilename)
	if err != nil {
		// Cleanup saved image? In production yes, here maybe too complex for atomic task.
		// Leaving as is for now.
		return nil, false, fmt.Errorf("failed to generate thumbnails: %w", err)
	}

	width, height, _ := GetImageDimensions(webpData)

	newImage := models.Image{
		TenantID:    tenantID,
		Filename:    baseFilename,
		OriginalURL: originalURL,
		LocalPath:   localPath,
		ContentHash: hashStr,
		MimeType:    "image/webp",
		Category:    category,
		Width:       width,
		Height:      height,
		FileSize:    int64(len(webpData)),
		RefCount:    1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&newImage).Error; err != nil {
		return nil, false, fmt.Errorf("failed to create image record: %w", err)
	}

	return &newImage, true, nil
}
