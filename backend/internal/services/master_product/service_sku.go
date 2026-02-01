// Package master_product provides Master Product management services
package master_product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ================== SKU Operations ==================

// AddSku adds a new SKU to a product
func (s *Service) AddSku(ctx context.Context, tenantID string, productID uint, input CreateSkuInput) (*models.MasterProductSku, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	// Verify product exists
	_, err := s.repo.FindByTenantAndID(ctx, tenantID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// Check SKU count
	existingSkus, _ := s.repo.FindSkusByProductID(ctx, productID)
	if len(existingSkus) >= models.MasterProductMaxSKUs {
		return nil, ErrTooManySKUs
	}

	// Validate SKU input
	if input.SellerSku == "" {
		return nil, ErrSellerSkuRequired
	}

	if err := s.createSkuForProduct(ctx, tenantID, productID, input); err != nil {
		return nil, err
	}

	// Find the created SKU
	return s.repo.FindBySku(ctx, tenantID, input.SellerSku)
}

// UpdateSku updates an existing SKU
func (s *Service) UpdateSku(ctx context.Context, tenantID string, skuID uint, input CreateSkuInput) (*models.MasterProductSku, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	// Find existing SKU by fetching all and filtering
	sku, err := s.findSkuByID(ctx, skuID)
	if err != nil {
		return nil, ErrSkuNotFound
	}

	// Verify tenant
	if sku.TenantID != tenantID {
		return nil, ErrSkuNotFound
	}

	// Update fields
	if input.SellerSku != "" {
		sku.SellerSku = input.SellerSku
	}
	sku.VariantName = input.VariantName
	if input.VariantData != nil {
		sku.VariantData = models.JSONMap(input.VariantData)
	}
	sku.Price = input.Price
	sku.Stock = input.Stock
	sku.UpdatedAt = time.Now()

	if err := s.repo.UpdateSku(ctx, sku); err != nil {
		return nil, err
	}

	return sku, nil
}

// DeleteSku removes a SKU from a product
func (s *Service) DeleteSku(ctx context.Context, tenantID string, skuID uint) error {
	if tenantID == "" {
		return ErrTenantIDRequired
	}

	sku, err := s.findSkuByID(ctx, skuID)
	if err != nil {
		return ErrSkuNotFound
	}

	if sku.TenantID != tenantID {
		return ErrSkuNotFound
	}

	// Delete platform links for this SKU first
	links, _ := s.repo.FindPlatformLinks(ctx, sku.MasterProductID)
	for _, link := range links {
		if link.MasterSkuID != nil && *link.MasterSkuID == skuID {
			_ = s.repo.DeletePlatformLink(ctx, link.ID)
		}
	}

	return s.repo.DeleteSku(ctx, skuID)
}

// GetProductWithSKUs retrieves a product with all its SKUs
func (s *Service) GetProductWithSKUs(ctx context.Context, tenantID string, id uint) (*models.MasterProduct, error) {
	return s.GetByID(ctx, tenantID, id)
}

// BatchUpdateSkuInput represents input for batch updating multiple SKUs
type BatchUpdateSkuInput struct {
	SkuIDs []uint   `json:"sku_ids"`
	Price  *float64 `json:"price,omitempty"`
	Stock  *int     `json:"stock,omitempty"`
}

// BatchUpdateSkuResult represents the result of a batch update operation
type BatchUpdateSkuResult struct {
	Updated int                       `json:"updated"`
	Failed  int                       `json:"failed"`
	SKUs    []models.MasterProductSku `json:"skus"`
	Errors  []string                  `json:"errors,omitempty"`
}

// BatchUpdateSkus updates multiple SKUs at once with the same price/stock values
func (s *Service) BatchUpdateSkus(ctx context.Context, tenantID string, input BatchUpdateSkuInput) (*BatchUpdateSkuResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	if len(input.SkuIDs) == 0 {
		return nil, errors.New("no SKU IDs provided")
	}

	if input.Price == nil && input.Stock == nil {
		return nil, errors.New("at least one of price or stock must be provided")
	}

	result := &BatchUpdateSkuResult{
		Updated: 0,
		Failed:  0,
		SKUs:    make([]models.MasterProductSku, 0),
		Errors:  make([]string, 0),
	}

	for _, skuID := range input.SkuIDs {
		sku, err := s.findSkuByID(ctx, skuID)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("SKU ID %d not found", skuID))
			continue
		}

		// Verify tenant
		if sku.TenantID != tenantID {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("SKU ID %d access denied", skuID))
			continue
		}

		// Update fields
		if input.Price != nil {
			sku.Price = *input.Price
		}
		if input.Stock != nil {
			sku.Stock = *input.Stock
		}
		sku.UpdatedAt = time.Now()

		if err := s.repo.UpdateSku(ctx, sku); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to update SKU ID %d", skuID))
			continue
		}

		result.Updated++
		result.SKUs = append(result.SKUs, *sku)
	}

	return result, nil
}
