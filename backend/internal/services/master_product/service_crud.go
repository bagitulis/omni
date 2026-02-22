package master_product

import (
	"context"
	"errors"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Update updates a master product
func (s *Service) Update(ctx context.Context, tenantID string, id uint, input UpdateInput) (*models.MasterProduct, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	// Fetch existing product
	product, err := s.repo.FindByTenantAndID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if input.Title != nil {
		if len(*input.Title) == 0 {
			return nil, ErrTitleRequired
		}
		if len(*input.Title) > models.MasterProductMaxTitleLength {
			return nil, ErrTitleTooLong
		}
		product.Title = *input.Title
	}

	if input.Description != nil {
		if len(*input.Description) > models.MasterProductMaxDescriptionLength {
			return nil, ErrDescriptionTooLong
		}
		product.Description = *input.Description
	}

	if input.Images != nil {
		if len(input.Images) > models.MasterProductMaxImages {
			return nil, ErrTooManyImages
		}
		images := make(models.JSONArray, len(input.Images))
		for i, img := range input.Images {
			images[i] = img
		}
		product.Images = images
	}

	if input.Status != nil {
		product.Status = *input.Status
	}

	if input.SKUs != nil {
		if err := s.replaceProductSKUs(ctx, tenantID, id, input.SKUs); err != nil {
			return nil, err
		}
	}

	product.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, product); err != nil {
		log.Error().Err(err).Uint("product_id", id).Msg("Failed to update master product")
		return nil, err
	}

	// Update join table entries (dual-write)
	if s.imageManager != nil && input.Images != nil {
		if err := s.imageManager.ProcessAndLinkImages(ctx, tenantID, id, input.Images); err != nil {
			log.Warn().Err(err).Uint("product_id", id).Msg("Failed to update image join entries (non-fatal)")
		}
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("product_id", id).
		Msg("Master product updated")

	return s.repo.FindByID(ctx, product.ID)
}

// Delete deletes a master product and its SKUs
func (s *Service) Delete(ctx context.Context, tenantID string, id uint) error {
	if tenantID == "" {
		return ErrTenantIDRequired
	}

	// Verify product exists and belongs to tenant
	_, err := s.repo.FindByTenantAndID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	// Delete image join entries and update ref counts
	if s.imageManager != nil {
		if err := s.imageManager.DeleteJoinTableEntries(ctx, id); err != nil {
			log.Warn().Err(err).Uint("product_id", id).Msg("Failed to delete image join entries (non-fatal)")
		}
	}

	// Delete associated SKUs first
	skus, _ := s.repo.FindSkusByProductID(ctx, id)
	for _, sku := range skus {
		// Delete platform links for this SKU
		links, _ := s.repo.FindPlatformLinks(ctx, id)
		for _, link := range links {
			if link.MasterSkuID != nil && *link.MasterSkuID == sku.ID {
				_ = s.repo.DeletePlatformLink(ctx, link.ID)
			}
		}
		_ = s.repo.DeleteSku(ctx, sku.ID)
	}

	// Delete product-level platform links
	links, _ := s.repo.FindPlatformLinks(ctx, id)
	for _, link := range links {
		_ = s.repo.DeletePlatformLink(ctx, link.ID)
	}

	// Delete product
	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Uint("product_id", id).Msg("Failed to delete master product")
		return err
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("product_id", id).
		Msg("Master product deleted")

	return nil
}
