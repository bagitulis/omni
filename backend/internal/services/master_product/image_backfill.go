package master_product

import (
	"context"
	"strings"

	"github.com/omni/backend/internal/models"
	imageService "github.com/omni/backend/internal/services/image"
	"github.com/rs/zerolog/log"
)

// BackfillImagesResult summarizes the backfill outcome.
type BackfillImagesResult struct {
	Processed int `json:"processed"`
	Updated   int `json:"updated"`
	Skipped   int `json:"skipped"`
	Errors    int `json:"errors"`
}

// BackfillImages aggregates platform local_images into master_products.images.
// When force is false, only products with empty images are updated.
func (s *Service) BackfillImages(ctx context.Context, tenantID string, limit int, force bool) (*BackfillImagesResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	products, err := s.repo.FindForImageBackfill(ctx, tenantID, limit, force)
	if err != nil {
		return nil, err
	}

	result := &BackfillImagesResult{}
	if len(products) == 0 {
		return result, nil
	}

	aggregator := NewImageAggregator(s.db)
	cacheService := imageService.NewProductImageCacheService()
	urlCache := make(map[string]string)
	for _, product := range products {
		result.Processed++

		beforeCount := countImages(product.Images)
		if force || beforeCount == 0 {
			if err := aggregator.AggregateImagesForProduct(ctx, product.ID); err != nil {
				result.Errors++
				log.Error().
					Err(err).
					Str("tenant_id", tenantID).
					Uint("master_product_id", product.ID).
					Msg("Failed to aggregate master product images")
				continue
			}
		}

		images, err := s.repo.FindImagesByID(ctx, tenantID, product.ID)
		if err != nil {
			result.Errors++
			log.Error().
				Err(err).
				Str("tenant_id", tenantID).
				Uint("master_product_id", product.ID).
				Msg("Failed to read master product images")
			continue
		}

		updated, err := s.ensureLocalMasterProductImages(
			ctx,
			tenantID,
			product.ID,
			images,
			cacheService,
			urlCache,
		)
		if err != nil {
			result.Errors++
			log.Error().
				Err(err).
				Str("tenant_id", tenantID).
				Uint("master_product_id", product.ID).
				Msg("Failed to cache master product images")
			continue
		}

		afterCount := countImages(images)
		if updated || (afterCount > 0 && afterCount != beforeCount) || (afterCount > 0 && beforeCount == 0) {
			result.Updated++
		} else {
			result.Skipped++
		}
	}

	return result, nil
}

func countImages(images models.JSONArray) int {
	if len(images) == 0 {
		return 0
	}
	count := 0
	for _, entry := range images {
		if value, ok := entry.(string); ok && value != "" {
			count++
		}
	}
	return count
}

func (s *Service) ensureLocalMasterProductImages(
	ctx context.Context,
	tenantID string,
	productID uint,
	images models.JSONArray,
	cacheService *imageService.ProductImageCacheService,
	urlCache map[string]string,
) (bool, error) {
	imageList := jsonArrayToStrings(images)
	if len(imageList) == 0 {
		return false, nil
	}

	updated := false
	newImages := make([]string, 0, len(imageList))
	for _, img := range imageList {
		if len(newImages) >= models.MasterProductMaxImages {
			break
		}
		if img == "" {
			continue
		}
		if isLocalImageURL(img) {
			newImages = appendUniqueString(newImages, img)
			continue
		}
		if cached, ok := urlCache[img]; ok {
			newImages = appendUniqueString(newImages, cached)
			updated = true
			continue
		}

		localPath, err := cacheService.CacheRemoteImage(
			ctx,
			tenantID,
			img,
			"asset",
			nil,
		)
		if err != nil || localPath == "" {
			newImages = appendUniqueString(newImages, img)
			continue
		}

		urlCache[img] = localPath
		newImages = appendUniqueString(newImages, localPath)
		updated = true
	}

	if !updated {
		return false, nil
	}

	jsonImages := make(models.JSONArray, len(newImages))
	for i, img := range newImages {
		jsonImages[i] = img
	}

	if err := s.repo.UpdateImages(ctx, tenantID, productID, jsonImages); err != nil {
		return false, err
	}

	return true, nil
}

func isLocalImageURL(url string) bool {
	return strings.HasPrefix(url, "/uploads/") || strings.Contains(url, "/uploads/")
}

func appendUniqueString(list []string, value string) []string {
	for _, entry := range list {
		if entry == value {
			return list
		}
	}
	return append(list, value)
}
