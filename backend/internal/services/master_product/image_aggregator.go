// Package master_product provides image aggregation services
package master_product

import (
	"context"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ImageAggregator handles image aggregation from platform products
type ImageAggregator struct {
	db *gorm.DB
}

// NewImageAggregator creates a new image aggregator
func NewImageAggregator(db *gorm.DB) *ImageAggregator {
	return &ImageAggregator{db: db}
}

// AggregateImagesForProduct collects local_images from linked platform products
// Priority: Shopee → TikTok → Lazada (first non-empty wins)
func (a *ImageAggregator) AggregateImagesForProduct(ctx context.Context, masterProductID uint) error {
	// 1. Get all platform links for this master product
	var links []models.MasterProductPlatformLink
	if err := a.db.WithContext(ctx).
		Where("master_product_id = ?", masterProductID).
		Find(&links).Error; err != nil {
		return err
	}

	if len(links) == 0 {
		return nil // No links, nothing to aggregate
	}

	// 2. Collect images with priority order
	var images []string
	platforms := []string{models.PlatformShopee, models.PlatformTiktok, models.PlatformLazada}

	for _, platform := range platforms {
		if len(images) >= models.MasterProductMaxImages {
			break // Already have enough images
		}

		for _, link := range links {
			if link.Platform != platform {
				continue
			}

			platformImages := a.getImagesFromPlatformProduct(ctx, platform, link.PlatformItemID)
			for _, img := range platformImages {
				if len(images) >= models.MasterProductMaxImages {
					break
				}
				// Avoid duplicates
				if !containsImage(images, img) {
					images = append(images, img)
				}
			}
		}
	}

	if len(images) == 0 {
		return nil // No images found
	}

	// 3. Update master product images
	jsonImages := make(models.JSONArray, len(images))
	for i, img := range images {
		jsonImages[i] = img
	}

	err := a.db.WithContext(ctx).
		Model(&models.MasterProduct{}).
		Where("id = ?", masterProductID).
		Update("images", jsonImages).Error

	if err != nil {
		log.Error().Err(err).Uint("master_product_id", masterProductID).Msg("Failed to update master product images")
		return err
	}

	log.Info().
		Uint("master_product_id", masterProductID).
		Int("image_count", len(images)).
		Msg("Master product images aggregated from platform products")

	return nil
}

// getImagesFromPlatformProduct retrieves local_images from platform-specific product table
func (a *ImageAggregator) getImagesFromPlatformProduct(ctx context.Context, platform, itemID string) []string {
	if itemID == "" {
		return nil
	}

	switch platform {
	case models.PlatformShopee:
		return a.getShopeeProductImages(ctx, itemID)
	case models.PlatformTiktok:
		return a.getTiktokProductImages(ctx, itemID)
	case models.PlatformLazada:
		return a.getLazadaProductImages(ctx, itemID)
	}
	return nil
}

func (a *ImageAggregator) getShopeeProductImages(ctx context.Context, itemID string) []string {
	var product struct {
		LocalImages models.JSONArray `gorm:"column:local_images"`
		Image       string           `gorm:"column:image"`
	}

	err := a.db.WithContext(ctx).
		Table(models.GetTableName("ShopeeProduct")).
		Select("local_images", "image").
		Where("item_id = ?", itemID).
		First(&product).Error

	if err != nil {
		return nil
	}

	if product.LocalImages != nil {
		images := jsonArrayToStrings(product.LocalImages)
		if len(images) > 0 {
			return images
		}
	}

	if product.Image != "" {
		return []string{product.Image}
	}

	return nil
}

func (a *ImageAggregator) getTiktokProductImages(ctx context.Context, productID string) []string {
	var product struct {
		LocalImages models.JSONArray `gorm:"column:local_images"`
		Image       string           `gorm:"column:image"`
	}

	err := a.db.WithContext(ctx).
		Table(models.GetTableName("TiktokProduct")).
		Select("local_images", "image").
		Where("product_id = ?", productID).
		First(&product).Error

	if err != nil {
		return nil
	}

	if product.LocalImages != nil {
		images := jsonArrayToStrings(product.LocalImages)
		if len(images) > 0 {
			return images
		}
	}

	if product.Image != "" {
		return []string{product.Image}
	}

	return nil
}

func (a *ImageAggregator) getLazadaProductImages(ctx context.Context, itemID string) []string {
	var product struct {
		LocalImages models.JSONArray `gorm:"column:local_images"`
		Image       string           `gorm:"column:image"`
	}

	err := a.db.WithContext(ctx).
		Table(models.GetTableName("LazadaProduct")).
		Select("local_images", "image").
		Where("item_id = ?", itemID).
		First(&product).Error

	if err != nil {
		return nil
	}

	if product.LocalImages != nil {
		images := jsonArrayToStrings(product.LocalImages)
		if len(images) > 0 {
			return images
		}
	}

	if product.Image != "" {
		return []string{product.Image}
	}

	return nil
}

// Helper functions
func jsonArrayToStrings(arr models.JSONArray) []string {
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok && s != "" {
			result = append(result, s)
		}
	}
	return result
}

func containsImage(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
