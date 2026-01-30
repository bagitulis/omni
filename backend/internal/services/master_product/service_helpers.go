// Package master_product provides Master Product management services
package master_product

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
)

// ================== Helper Methods ==================

func (s *Service) validateCreateInput(input CreateInput) error {
	if input.Title == "" {
		return ErrTitleRequired
	}
	if len(input.Title) > models.MasterProductMaxTitleLength {
		return ErrTitleTooLong
	}
	if len(input.Description) > models.MasterProductMaxDescriptionLength {
		return ErrDescriptionTooLong
	}
	if len(input.Images) > models.MasterProductMaxImages {
		return ErrTooManyImages
	}
	if len(input.SKUs) > models.MasterProductMaxSKUs {
		return ErrTooManySKUs
	}

	// Validate each SKU
	for _, sku := range input.SKUs {
		if sku.SellerSku == "" {
			return ErrSellerSkuRequired
		}
	}

	return nil
}

func (s *Service) createSkuForProduct(ctx context.Context, tenantID string, productID uint, input CreateSkuInput) error {
	var variantData models.JSONMap
	if input.VariantData != nil {
		variantData = models.JSONMap(input.VariantData)
	}

	sku := &models.MasterProductSku{
		TenantID:        tenantID,
		MasterProductID: productID,
		SellerSku:       input.SellerSku,
		VariantName:     input.VariantName,
		VariantData:     variantData,
		Price:           input.Price,
		Stock:           input.Stock,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	return s.repo.CreateSku(ctx, sku)
}

func (s *Service) findSkuByID(ctx context.Context, skuID uint) (*models.MasterProductSku, error) {
	var sku models.MasterProductSku
	err := s.db.WithContext(ctx).Where("id = ?", skuID).First(&sku).Error
	if err != nil {
		return nil, err
	}
	return &sku, nil
}
