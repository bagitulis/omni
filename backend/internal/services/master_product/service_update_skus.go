package master_product

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/omni/backend/internal/models"
)

// UpdateSkuInput represents SKU payload for product update.
type UpdateSkuInput struct {
	ID          *uint                  `json:"id,omitempty"`
	SellerSku   string                 `json:"seller_sku"`
	VariantName string                 `json:"variant_name,omitempty"`
	VariantData map[string]interface{} `json:"variant_data,omitempty"`
	Price       float64                `json:"price"`
	Stock       int                    `json:"stock"`
}

func (s *Service) replaceProductSKUs(ctx context.Context, tenantID string, productID uint, inputs []UpdateSkuInput) error {
	if len(inputs) > models.MasterProductMaxSKUs {
		return ErrTooManySKUs
	}

	existingSKUs, err := s.repo.FindSkusByProductID(ctx, productID)
	if err != nil {
		return err
	}

	existingByID := make(map[uint]*models.MasterProductSku, len(existingSKUs))
	for i := range existingSKUs {
		sku := &existingSKUs[i]
		existingByID[sku.ID] = sku
	}

	keepIDs := make(map[uint]bool, len(inputs))
	now := time.Now()

	for _, input := range inputs {
		sellerSKU := strings.TrimSpace(input.SellerSku)
		if sellerSKU == "" {
			return ErrSellerSkuRequired
		}

		variantData := models.JSONMap{}
		if input.VariantData != nil {
			variantData = models.JSONMap(input.VariantData)
		}

		if input.ID != nil {
			existing, ok := existingByID[*input.ID]
			if !ok || existing.TenantID != tenantID || existing.MasterProductID != productID {
				return ErrSkuNotFound
			}

			existing.SellerSku = sellerSKU
			existing.VariantName = input.VariantName
			existing.VariantData = variantData
			existing.Price = input.Price
			existing.Stock = input.Stock
			existing.UpdatedAt = now

			if err := s.repo.UpdateSku(ctx, existing); err != nil {
				return fmt.Errorf("failed to update sku %d: %w", existing.ID, err)
			}

			keepIDs[existing.ID] = true
			continue
		}

		newSKU := &models.MasterProductSku{
			TenantID:        tenantID,
			MasterProductID: productID,
			SellerSku:       sellerSKU,
			VariantName:     input.VariantName,
			VariantData:     variantData,
			Price:           input.Price,
			Stock:           input.Stock,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		if err := s.repo.CreateSku(ctx, newSKU); err != nil {
			return fmt.Errorf("failed to create sku %s: %w", sellerSKU, err)
		}

		keepIDs[newSKU.ID] = true
	}

	links, err := s.repo.FindPlatformLinks(ctx, productID)
	if err != nil {
		return err
	}

	for _, existing := range existingSKUs {
		if keepIDs[existing.ID] {
			continue
		}

		for _, link := range links {
			if link.MasterSkuID != nil && *link.MasterSkuID == existing.ID {
				if err := s.repo.DeletePlatformLink(ctx, link.ID); err != nil {
					return fmt.Errorf("failed to delete platform link %d: %w", link.ID, err)
				}
			}
		}

		if err := s.repo.DeleteSku(ctx, existing.ID); err != nil {
			return fmt.Errorf("failed to delete sku %d: %w", existing.ID, err)
		}
	}

	return nil
}
