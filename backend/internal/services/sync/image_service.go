package sync

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils/logger"
	"github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

var imageLogger = logger.Named("ImageService")

const (
	// ShopeeImageBatchSize is max items per Shopee API request
	ShopeeImageBatchSize = 50
)

// ImageService handles product image fetching and caching
type ImageService struct {
	shopeeClient *shopee.Client
	db           *gorm.DB
	tenantID     string
}

// NewImageService creates a new image service
func NewImageService(shopeeClient *shopee.Client, db *gorm.DB, tenantID string) *ImageService {
	return &ImageService{
		shopeeClient: shopeeClient,
		db:           db,
		tenantID:     tenantID,
	}
}

// GetProductImages fetches images for given item IDs
// Returns map of item_id -> image_url
// Strategy: Check ShopeeProduct cache first, fetch from API if not found
func (s *ImageService) GetProductImages(ctx context.Context, itemIDs []int64) (map[int64]string, error) {
	if len(itemIDs) == 0 {
		return make(map[int64]string), nil
	}

	result := make(map[int64]string, len(itemIDs))

	// Step 1: Check cache in ShopeeProduct table
	cachedImages, missingIDs := s.getCachedImages(ctx, itemIDs)
	for itemID, imageURL := range cachedImages {
		result[itemID] = imageURL
	}

	if len(missingIDs) == 0 {
		imageLogger.WithTenantID(s.tenantID).WithFields(map[string]interface{}{
			"total_requested": len(itemIDs),
			"cache_hits":      len(cachedImages),
		}).Info("All images found in cache")
		return result, nil
	}

	// Step 2: Fetch missing images from Shopee API
	fetchedImages, err := s.fetchImagesFromAPI(ctx, missingIDs)
	if err != nil {
		imageLogger.WithTenantID(s.tenantID).Error("Failed to fetch images from API: " + err.Error())
		// Return partial result with cached images
		return result, err
	}

	// Merge fetched images into result
	for itemID, imageURL := range fetchedImages {
		result[itemID] = imageURL
	}

	imageLogger.WithTenantID(s.tenantID).WithFields(map[string]interface{}{
		"total_requested": len(itemIDs),
		"cache_hits":      len(cachedImages),
		"api_fetched":     len(fetchedImages),
	}).Info("Product images retrieved")

	return result, nil
}

// getCachedImages retrieves cached images from ShopeeProduct table
// Returns map of found images and slice of missing item IDs
func (s *ImageService) getCachedImages(ctx context.Context, itemIDs []int64) (map[int64]string, []int64) {
	cached := make(map[int64]string)
	var missing []int64

	// Query products with images
	var products []models.ShopeeProduct
	err := s.db.WithContext(ctx).
		Select("item_id, image").
		Where("tenant_id = ? AND item_id IN ? AND image != ''", s.tenantID, itemIDs).
		Find(&products).Error

	if err != nil {
		imageLogger.WithTenantID(s.tenantID).Warn("Failed to query cached images: " + err.Error())
		// Return all as missing
		return cached, itemIDs
	}

	// Build cached map
	foundIDs := make(map[int64]bool)
	for _, p := range products {
		if p.Image != "" {
			cached[p.ItemID] = p.Image
			foundIDs[p.ItemID] = true
		}
	}

	// Build missing list
	for _, id := range itemIDs {
		if !foundIDs[id] {
			missing = append(missing, id)
		}
	}

	return cached, missing
}

// fetchImagesFromAPI fetches images from Shopee API for given item IDs
// Also updates ShopeeProduct cache for future lookups
func (s *ImageService) fetchImagesFromAPI(ctx context.Context, itemIDs []int64) (map[int64]string, error) {
	if s.shopeeClient == nil {
		return nil, fmt.Errorf("shopee client not configured")
	}

	result := make(map[int64]string)

	// Process in batches of 50 (Shopee API limit)
	for i := 0; i < len(itemIDs); i += ShopeeImageBatchSize {
		end := i + ShopeeImageBatchSize
		if end > len(itemIDs) {
			end = len(itemIDs)
		}
		batch := itemIDs[i:end]

		batchResult, err := s.fetchBatchImages(ctx, batch)
		if err != nil {
			imageLogger.WithTenantID(s.tenantID).Warn("Batch image fetch failed: " + err.Error())
			continue // Continue with other batches
		}

		for itemID, imageURL := range batchResult {
			result[itemID] = imageURL
		}
	}

	return result, nil
}

// fetchBatchImages fetches images for a single batch from API
func (s *ImageService) fetchBatchImages(ctx context.Context, itemIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)

	resp, err := s.shopeeClient.GetProductDetailWithImages(ctx, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("shopee API error: %w", err)
	}

	if resp == nil || resp.Response.ItemList == nil {
		return result, nil
	}

	// Extract images and update cache
	for _, item := range resp.Response.ItemList {
		imageURL := extractFirstImageURL(item)
		if imageURL != "" {
			result[item.ItemID] = imageURL
			// Update cache in background
			s.updateProductImageCache(ctx, item.ItemID, imageURL)
		}
	}

	return result, nil
}

// extractFirstImageURL extracts the first image URL from product detail
func extractFirstImageURL(item shopee.ProductDetailWithImages) string {
	if len(item.Image.ImageURLList) > 0 {
		return item.Image.ImageURLList[0]
	}
	return ""
}

// updateProductImageCache updates the image field in ShopeeProduct table
func (s *ImageService) updateProductImageCache(ctx context.Context, itemID int64, imageURL string) {
	err := s.db.WithContext(ctx).
		Model(&models.ShopeeProduct{}).
		Where("tenant_id = ? AND item_id = ?", s.tenantID, itemID).
		Update("image", imageURL).Error

	if err != nil {
		imageLogger.WithTenantID(s.tenantID).WithFields(map[string]interface{}{
			"item_id": itemID,
		}).Warn("Failed to update image cache: " + err.Error())
	}
}

// GetProductImagesAsync fetches images asynchronously and returns results via channel
// Useful for non-blocking image loading
func (s *ImageService) GetProductImagesAsync(ctx context.Context, itemIDs []int64) <-chan ImageResult {
	resultChan := make(chan ImageResult, 1)

	go func() {
		defer close(resultChan)
		images, err := s.GetProductImages(ctx, itemIDs)
		resultChan <- ImageResult{
			Images: images,
			Error:  err,
		}
	}()

	return resultChan
}

// ImageResult represents the result of async image fetch
type ImageResult struct {
	Images map[int64]string
	Error  error
}
