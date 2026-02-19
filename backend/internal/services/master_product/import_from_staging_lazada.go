// Package master_product provides import service for Master Product
package master_product

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
)

// ImportFromLazadaStaging imports all Lazada staging products into Master Product for a tenant.
// It iterates LazadaProduct rows, finds-or-creates a MasterProduct per product, then maps
// every LazadaSku into a MasterProductSku with a corresponding MasterProductPlatformLink.
func (s *StagingImportService) ImportFromLazadaStaging(ctx context.Context, tenantID string) (*StagingImportResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	result := &StagingImportResult{}

	// 1. Fetch all Lazada products for the tenant.
	var products []models.LazadaProduct
	if err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("fetch lazada products: %w", err)
	}

	for _, p := range products {
		// 2a. Find or create master product by name.
		masterProduct, created, err := s.findOrCreateMasterProduct(ctx, tenantID, p.Name)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("lazada product %s master product: %w", p.ItemID, err).Error())
			result.ProductsSkipped++
			continue
		}

		// 2b. Track created vs matched.
		if created {
			result.ProductsCreated++
		} else {
			result.ProductsMatched++
		}

		// 2c. Fetch all SKUs for this product (join via ItemID string).
		var skus []models.LazadaSku
		if err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND item_id = ?", tenantID, p.ItemID).
			Find(&skus).Error; err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("lazada product %s fetch skus: %w", p.ItemID, err).Error())
			continue
		}

		// 2e. No SKUs — create one default SKU with an empty variant name.
		if len(skus) == 0 {
			sellerSku := fmt.Sprintf("lazada_%s", p.ItemID)

			defaultSku := &models.MasterProductSku{
				TenantID:        tenantID,
				MasterProductID: masterProduct.ID,
				SellerSku:       sellerSku,
				VariantName:     "",
			}

			upsertedSku, created, err := s.upsertMasterSku(ctx, defaultSku)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s default sku: %w", p.ItemID, err).Error())
				continue
			}
			if created {
				result.SkusCreated++
			}

			link := &models.MasterProductPlatformLink{
				MasterProductID:   masterProduct.ID,
				MasterSkuID:       &upsertedSku.ID,
				Platform:          "lazada",
				PlatformProductID: p.ItemID,
				PlatformItemID:    p.ItemID,
				PlatformSkuID:     "",
				SyncStatus:        "synced",
			}
			if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s default sku link: %w", p.ItemID, err).Error())
				continue
			}
			result.LinksCreated++
			continue
		}

		// 2d. Process each SKU.
		for _, sku := range skus {
			// i. Variant name: prefer Name, fallback to VariantName.
			variantName := sku.Name
			if variantName == "" {
				variantName = sku.VariantName
			}

			// ii. Seller SKU: prefer SellerSku, fallback to "lazada_{ItemID}".
			sellerSku := sku.SellerSku
			if sellerSku == "" {
				sellerSku = fmt.Sprintf("lazada_%s", p.ItemID)
			}

			// iii. Create master SKU.
			masterSku := &models.MasterProductSku{
				TenantID:        tenantID,
				MasterProductID: masterProduct.ID,
				SellerSku:       sellerSku,
				VariantName:     variantName,
				Price:           sku.Price,
				Stock:           sku.Quantity,
			}

			// iv. Persist SKU; on error append and continue.
			upsertedSku, created, err := s.upsertMasterSku(ctx, masterSku)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s sku: %w", p.ItemID, err).Error())
				continue
			}
			if created {
				result.SkusCreated++
			}

			// vi. Upsert platform link.
			link := &models.MasterProductPlatformLink{
				MasterProductID:   masterProduct.ID,
				MasterSkuID:       &upsertedSku.ID,
				Platform:          "lazada",
				PlatformProductID: p.ItemID,
				PlatformItemID:    p.ItemID,
				PlatformSkuID:     sku.SkuID,
				SyncStatus:        "synced",
			}
			// vii. On error append and continue.
			if err := s.repo.UpsertPlatformLink(ctx, link); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("lazada product %s sku link: %w", p.ItemID, err).Error())
				continue
			}
			// viii.
			result.LinksCreated++
		}
	}

	return result, nil
}
