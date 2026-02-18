package master_product

import (
	"context"
	"errors"

	imageService "github.com/omni/backend/internal/services/image"
	"gorm.io/gorm"
)

// RefreshProductImagesResult summarizes single-product image refresh output.
type RefreshProductImagesResult struct {
	MasterProductID    uint     `json:"master_product_id"`
	PreviousImageCount int      `json:"previous_image_count"`
	ImageCount         int      `json:"image_count"`
	Updated            bool     `json:"updated"`
	Forced             bool     `json:"forced"`
	Images             []string `json:"images"`
}

// RefreshProductImages refreshes a master product image list using linked platform local_images.
func (s *Service) RefreshProductImages(ctx context.Context, tenantID string, productID uint, force bool) (*RefreshProductImagesResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	product, err := s.repo.FindByTenantAndID(ctx, tenantID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	previousCount := countImages(product.Images)

	if force || previousCount == 0 {
		aggregator := NewImageAggregator(s.db)
		if err := aggregator.AggregateImagesForProduct(ctx, productID); err != nil {
			return nil, err
		}
	}

	images, err := s.repo.FindImagesByID(ctx, tenantID, productID)
	if err != nil {
		return nil, err
	}

	cacheUpdated, err := s.ensureLocalMasterProductImages(
		ctx,
		tenantID,
		productID,
		images,
		imageService.NewProductImageCacheService(),
		make(map[string]string),
	)
	if err != nil {
		return nil, err
	}

	if cacheUpdated {
		images, err = s.repo.FindImagesByID(ctx, tenantID, productID)
		if err != nil {
			return nil, err
		}
	}

	imageList := jsonArrayToStrings(images)
	currentCount := countImages(images)

	return &RefreshProductImagesResult{
		MasterProductID:    productID,
		PreviousImageCount: previousCount,
		ImageCount:         currentCount,
		Updated:            cacheUpdated || currentCount != previousCount,
		Forced:             force,
		Images:             imageList,
	}, nil
}
